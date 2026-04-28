package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/format"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// ----------------------------------------------------------------------------
// Active client construction (shared across the whole tree)
// ----------------------------------------------------------------------------

// buildDashboardClient resolves the active profile, hands the factory's
// authenticated *http.Client to a fresh *Client, and returns it. Each leaf
// command calls this from its RunE; the factory memoises the http.Client so
// repeated calls inside a single invocation are cheap.
func buildDashboardClient(ctx context.Context, f *cmdutil.Factory) (*Client, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	if rp == nil {
		return nil, errors.New("no profile selected; run `olares-cli profile use <olaresId>` or pass --profile <olaresId>")
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(hc, rp), nil
}

// requireProfileURL is a defensive check for callers that hand-build URLs
// (logging, error messages). Returns the OlaresID for context or "" when
// there is no profile.
func profileOlaresID(rp *credential.ResolvedProfile) string {
	if rp == nil {
		return ""
	}
	return rp.OlaresID
}

// ----------------------------------------------------------------------------
// Monitoring helpers — used by overview cpu/memory/disk/pods/physical/ranking
// ----------------------------------------------------------------------------

// monitoringWindow encapsulates the time-window flags (--since / --start /
// --end) and the SPA's default fallback. The SPA's
// controlPanelCommon/containers/Monitoring/config.ts uses step=600s,
// times=20 for the cluster headline and step=60s for detail pages; we mirror
// the cluster default and let leaves override.
type monitoringWindow struct {
	Step       time.Duration // upstream "step" param (Prometheus-flavoured)
	Times      int           // upstream "times" param (sample count)
	DefaultDur time.Duration // fallback when --since / --start are unset
}

// defaultClusterWindow is the headline cluster window. 600s × 20 = 200min,
// matching the SPA's `getParams({})` defaults for the overview page.
func defaultClusterWindow() monitoringWindow {
	return monitoringWindow{
		Step:       600 * time.Second,
		Times:      20,
		DefaultDur: 200 * time.Minute,
	}
}

// defaultDetailWindow is the per-detail-page sliding window. 60s × 50 = 50min,
// matching the SPA's CPU/Memory/Pods detail page `timeRangeDefault`.
func defaultDetailWindow() monitoringWindow {
	return monitoringWindow{
		Step:       60 * time.Second,
		Times:      50,
		DefaultDur: 50 * time.Minute,
	}
}

// monitoringQuery composes the URL query for a /cluster or /nodes monitoring
// fetch. We honour --since / --start / --end via CommonFlags; everything
// else (step, times, instant) defaults to the SPA contract.
func monitoringQuery(metricsFilter []string, w monitoringWindow, now time.Time, instant bool) url.Values {
	v := url.Values{}
	v.Set("metrics_filter", strings.Join(metricsFilter, "|")+"$")
	if instant {
		// "step=0s" is the SPA's `step: '0s'` shorthand for "give me the
		// instantaneous reading" (used by Disk detail). The BFF treats
		// it as a value-only query and skips the values[] array.
		v.Set("step", "0s")
		return v
	}
	if common.HasAbsoluteWindow() {
		v.Set("start", fmt.Sprintf("%d", common.Start.Unix()))
		v.Set("end", fmt.Sprintf("%d", common.End.Unix()))
	} else {
		end := now
		dur := common.Since
		if dur == 0 {
			dur = w.DefaultDur
		}
		start := end.Add(-dur)
		v.Set("start", fmt.Sprintf("%d", start.Unix()))
		v.Set("end", fmt.Sprintf("%d", end.Unix()))
	}
	v.Set("step", fmt.Sprintf("%ds", int(w.Step.Seconds())))
	v.Set("times", fmt.Sprintf("%d", w.Times))
	return v
}

// fetchClusterMetrics issues GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster
// with metrics_filter and returns the per-metric raw payload (one entry per
// `metric_name`). Used by overview physical / overview ranking sections.
//
// Wire shape (from the SPA's getClusterMonitoring):
//
//	GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster?metrics_filter=...$&start=...&end=...&step=600s&times=20
//
// Response: { results: [ { metric_name, data: { result: [ { metric, values, value } ] } } ] }
func fetchClusterMetrics(ctx context.Context, c *Client, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	q := monitoringQuery(metrics, w, now, instant)
	return doMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/cluster", q)
}

// fetchNodeMetrics issues GET /kapis/.../v1alpha3/nodes — used by every
// per-node detail page (CPU / Memory / Disk / Pods / Fan-via-monitoring,
// when applicable). Response shape is identical to fetchClusterMetrics.
func fetchNodeMetrics(ctx context.Context, c *Client, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	q := monitoringQuery(metrics, w, now, instant)
	return doMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q)
}

// fetchUserMetric issues GET /kapis/.../v1alpha3/users/<username> — used by
// overview user. Same monitoring-response shape as cluster / nodes.
func fetchUserMetric(ctx context.Context, c *Client, username string, metrics []string, w monitoringWindow, now time.Time, instant bool) (map[string]format.MonitoringResult, error) {
	if username == "" {
		return nil, errors.New("fetchUserMetric: username is required")
	}
	q := monitoringQuery(metrics, w, now, instant)
	path := fmt.Sprintf("/kapis/monitoring.kubesphere.io/v1alpha3/users/%s", url.PathEscape(username))
	return doMonitoring(ctx, c, path, q)
}

// doMonitoring shells out to the BFF, decodes the standard
// `{ results: [...] }` envelope, and returns a metric_name → MonitoringResult
// map ready for format.GetLastMonitoringData.
func doMonitoring(ctx context.Context, c *Client, path string, q url.Values) (map[string]format.MonitoringResult, error) {
	var raw struct {
		Results []struct {
			MetricName string                 `json:"metric_name"`
			Data       map[string]interface{} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, path, q, nil, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]format.MonitoringResult, len(raw.Results))
	for _, r := range raw.Results {
		b, err := encodeJSONMap(map[string]interface{}{"data": r.Data})
		if err != nil {
			continue
		}
		var mr format.MonitoringResult
		if err := decodeBytesMap(b, &mr); err != nil {
			continue
		}
		out[r.MetricName] = mr
	}
	return out, nil
}

// fetchNodesList returns the sorted set of node names. Used by overview
// physical (to label cluster-level rows with hostnames).
func fetchNodesList(ctx context.Context, c *Client) ([]string, error) {
	var raw struct {
		Items []struct {
			Metadata struct {
				Name string `json:"name"`
			} `json:"metadata"`
		} `json:"items"`
	}
	q := url.Values{"sortBy": []string{"createTime"}}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/resources.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(raw.Items))
	for _, it := range raw.Items {
		out = append(out, it.Metadata.Name)
	}
	sort.Strings(out)
	return out, nil
}

// ----------------------------------------------------------------------------
// Capability gates (overview fan / overview gpu)
// ----------------------------------------------------------------------------
//
// Both subtrees mirror the SPA's hard gates from
// `Overview2/ClusterResource.vue` (line 232-238 + 278-293):
//
//   Fan: only Olares One hardware
//     → device_name == "Olares One" via /user-service/api/system/status
//   GPU: admin AND any node has gpu.bytetrade.io/cuda-supported=true
//     → cluster role check + label scan on /kapis/.../nodes
//
// The CLI replicates these gates so agents see a structured empty
// envelope (with EmptyReason / Note / DeviceName) instead of a
// "silently zero" payload from the BFF.

// cudaNodeOnce / cudaNodePresent / cudaNodeErr cache the result of
// hasCUDANode for the duration of the *Client. We attach the cache to a
// per-Client map keyed by the client pointer so tests with multiple
// fixtures don't share state, but in production each invocation has a
// single Client so it's a free win.
var (
	cudaNodeMu       sync.Mutex
	cudaNodeCache    = map[*Client]cudaNodeResult{}
)

type cudaNodeResult struct {
	present bool
	err     error
	done    bool
}

// hasCUDANode reports whether the cluster has at least one node with
// label `gpu.bytetrade.io/cuda-supported=true`. Mirrors the SPA's
// `checkGpu` (Overview2/ClusterResource.vue:278-293) which iterates
// the nodes list and looks for the cuda-supported label.
//
// Cached per-Client; the second call inside the same CLI invocation is
// free. The label-only fast path keeps payloads small even on large
// clusters since we just need a presence check.
func hasCUDANode(ctx context.Context, c *Client) (bool, error) {
	cudaNodeMu.Lock()
	if r, ok := cudaNodeCache[c]; ok && r.done {
		cudaNodeMu.Unlock()
		return r.present, r.err
	}
	cudaNodeMu.Unlock()

	var raw struct {
		Items []struct {
			Metadata struct {
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"items"`
	}
	q := url.Values{"sortBy": []string{"createTime"}}
	err := c.DoJSON(ctx, http.MethodGet, "/kapis/resources.kubesphere.io/v1alpha3/nodes", q, nil, &raw)
	present := false
	if err == nil {
		for _, it := range raw.Items {
			if it.Metadata.Labels["gpu.bytetrade.io/cuda-supported"] == "true" {
				present = true
				break
			}
		}
	}
	cudaNodeMu.Lock()
	cudaNodeCache[c] = cudaNodeResult{present: present, err: err, done: true}
	cudaNodeMu.Unlock()
	return present, err
}

// gateOlaresOne returns (gatedEnvelope, true) when the active device
// is not Olares One; the caller should emit `gatedEnvelope` and skip
// any data fetch. The hint message is also written to stderr in
// non-JSON output modes so humans see why the table is empty.
//
// On error from EnsureSystemStatus we let the caller proceed (gated=false,
// nil envelope) — the downstream BFF call will surface the real error
// itself rather than masking it with a confused "not Olares One" hint.
func gateOlaresOne(ctx context.Context, c *Client, kind string, now time.Time) (Envelope, bool) {
	st, err := c.EnsureSystemStatus(ctx)
	if err != nil || st == nil {
		return Envelope{}, false
	}
	if st.IsOlaresOne() {
		return Envelope{}, false
	}
	dev := st.DeviceName
	if dev == "" {
		dev = "unknown"
	}
	env := Envelope{
		Kind: kind,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	env.Meta.Empty = true
	env.Meta.EmptyReason = "not_olares_one"
	env.Meta.Note = "Fan / cooling integration is only available on Olares One devices"
	env.Meta.DeviceName = dev
	if common.Output != OutputJSON {
		fmt.Fprintf(os.Stderr,
			"fan is only available on Olares One devices (current: %s)\n", dev)
	}
	return env, true
}

// gpuAdvisory is the soft-gate companion to gateOlaresOne. The SPA's
// GPU detail pages (`Overview2/GPU/IndexPage.vue`) carry NO admin or
// CUDA gate themselves — the only hard gate in the SPA is at the
// sidebar card (Overview2/ClusterResource.vue:232+278-293) which just
// hides the entry. Anyone landing on the URL directly hits HAMI without
// pre-checks.
//
// To match that behaviour the CLI no longer blocks data fetches; it
// only emits a one-line stderr advisory and tags the envelope
// `meta.note` with the reason the SPA would have hidden the card. Two
// soft signals:
//
//   - non-admin profile  → "gpu_sidebar_hidden_non_admin"
//   - no CUDA-capable node → "gpu_sidebar_hidden_no_cuda_node"
//
// Both are advisory-only; the caller continues to fetch and renders
// data when HAMI returns it. Returns (note, "") when no advisory
// applies, or (note, reason) — both empty when EnsureUser /
// hasCUDANode fail (we fall silent rather than misleading agents).
func gpuAdvisory(ctx context.Context, c *Client) (note, reason string) {
	u, err := c.EnsureUser(ctx)
	if err != nil || u == nil {
		return "", ""
	}
	if !u.IsAdmin() {
		if common.Output != OutputJSON {
			fmt.Fprintf(os.Stderr,
				"(advisory) GPU sidebar entry is hidden for non-admin profiles in the SPA; current user (%s) is %s\n",
				u.Name, displayRole(u.GlobalRole))
		}
		return "GPU sidebar entry is hidden for non-admin profiles in the SPA; HAMI was queried directly", "gpu_sidebar_hidden_non_admin"
	}
	present, err := hasCUDANode(ctx, c)
	if err != nil {
		return "", ""
	}
	if !present {
		if common.Output != OutputJSON {
			fmt.Fprintln(os.Stderr,
				"(advisory) no node carries gpu.bytetrade.io/cuda-supported=true; SPA hides the GPU card. HAMI was queried directly")
		}
		return "no node carries gpu.bytetrade.io/cuda-supported=true; SPA hides the GPU card. HAMI was queried directly", "gpu_sidebar_hidden_no_cuda_node"
	}
	return "", ""
}

// vgpuUnavailableFromError converts a HAMI-side error into the
// (empty=true, empty_reason=vgpu_unavailable) envelope when the
// upstream came back with a 5xx. The caller is responsible for the
// 404 branch (no_vgpu_integration) which keeps existing semantics.
//
// `err` is the result of one of the fetch* helpers; `kind` / `now` /
// `c` provide envelope context. Returns (env, true) when the error
// matches the 5xx HAMI-down pattern; (zero, false) otherwise so the
// caller can re-raise.
//
// We extract a short body message (capped at 256 bytes) and stash it
// in `meta.error` so agents can drill in without parsing free-form
// strings. Stderr in non-JSON mode prints a single advisory line.
func vgpuUnavailableFromError(c *Client, err error, kind string, now time.Time) (Envelope, bool) {
	he, ok := IsHTTPError(err)
	if !ok || he.Status < 500 || he.Status >= 600 {
		return Envelope{}, false
	}
	msg := extractHAMIMessage(he.Body)
	env := Envelope{
		Kind: kind,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	env.Meta.Empty = true
	env.Meta.EmptyReason = "vgpu_unavailable"
	env.Meta.Note = "HAMI vGPU controller responded with " + http.StatusText(he.Status) + "; the integration is installed but unhealthy"
	env.Meta.HTTPStatus = he.Status
	if msg != "" {
		env.Meta.Error = msg
	}
	if common.Output != OutputJSON {
		if msg != "" {
			fmt.Fprintf(os.Stderr,
				"gpu data temporarily unavailable: HAMI returned HTTP %d (%s)\n",
				he.Status, msg)
		} else {
			fmt.Fprintf(os.Stderr,
				"gpu data temporarily unavailable: HAMI returned HTTP %d\n",
				he.Status)
		}
	}
	return env, true
}

// extractHAMIMessage tries to surface the `message` field from a HAMI
// JSON-shaped body (`{"code": <int>, "message": "..."}`); falls back
// to the trimmed body itself capped at 256 bytes. Caller pre-strips
// the body via the *HTTPError struct.
func extractHAMIMessage(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}
	if strings.HasPrefix(body, "{") {
		var probe struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := jsonUnmarshal([]byte(body), &probe); err == nil && probe.Message != "" {
			return probe.Message
		}
	}
	if len(body) > 256 {
		body = body[:256]
	}
	return body
}

// displayRole pretty-prints an empty / unknown role string for the
// stderr hint so humans see "(unset)" rather than two consecutive
// spaces.
func displayRole(r string) string {
	if strings.TrimSpace(r) == "" {
		return "(unset)"
	}
	return r
}

// ----------------------------------------------------------------------------
// fetchWorkloadsMetrics — overview ranking + applications (workload-grain)
// ----------------------------------------------------------------------------
//
// Mirrors the SPA's `fetchWorkloadsMetrics(apps, namespace, sort)` —
// see apps/packages/app/src/apps/dashboard/pages/Applications2/config.ts.
//
// The SPA fans out two parallel calls:
//
//	1. GET /kapis/monitoring.kubesphere.io/v1alpha3/namespaces/<userNs>/pods
//	   filtered to pod_cpu_usage|pod_memory_usage_wo_cache|pod_net_bytes_*
//	   and resources_filter = system-app deployment regex.
//	2. GET /kapis/monitoring.kubesphere.io/v1alpha3/namespaces
//	   filtered to namespace_cpu_usage|namespace_memory_usage_wo_cache|
//	   namespace_net_bytes_* and resources_filter = pipe-joined custom
//	   namespaces.
//
// It then merges the two responses into a single `{cpu_usage, memory_usage,
// net_transmitted, net_received}` map indexed by application name.
//
// We deliberately preserve the SPA's dual-fetch + merge so the CLI's
// `overview ranking` first-row matches `applications` first-row byte-for-byte.

// workloadAggregate is the merged shape returned by fetchWorkloadsMetrics.
// Each row carries the application's logical identity (name / title /
// namespace / system flag / owner) plus the four metric values plus the
// running-pod count for the row. State / Title / Icon ride along so the
// CLI can present the same per-app card the SPA does without a second
// round-trip.
type workloadAggregate struct {
	Name       string
	Title      string
	Icon       string
	Namespace  string
	Deployment string
	OwnerKind  string
	State      string
	IsSystem   bool

	PodCount int

	CPU    float64 // pod_cpu_usage / namespace_cpu_usage (last sample)
	Memory float64 // pod_memory_usage_wo_cache / namespace_memory_usage_wo_cache
	NetIn  float64 // pod_net_bytes_received / namespace_net_bytes_received
	NetOut float64 // pod_net_bytes_transmitted / namespace_net_bytes_transmitted
}

// workloadRequest is what callers feed into fetchWorkloadsMetrics:
// a flattened list of registered apps + the active user's namespace.
type workloadRequest struct {
	Apps          []workloadApp
	UserNamespace string
	Sort          string // "asc" | "desc"
	SortBy        string // "cpu" (default) | "memory" | "net_in" | "net_out"
}

// workloadApp is one row of the SPA's app inventory (post `entrances`
// filter). Mirrors `AppListItem` from
// `controlPanelCommon/network/network.ts:280` for the fields the workload
// merge actually reads.
type workloadApp struct {
	Name       string
	Title      string
	Icon       string
	Namespace  string
	Deployment string
	OwnerKind  string
	State      string
	IsSystem   bool
}

// systemFrontendDeployment mirrors the SPA's `SYSTEM_FRONTEND_DEPLOYMENT`
// constant (Applications2/config.ts:25). Multiple "entrance" apps share
// the same physical deployment; the merge step has to clone the metric to
// each app entry so the cards line up with what users see in the UI.
const systemFrontendDeployment = "system-frontend-deployment"

// podDeploymentName mirrors `podDeploymentName(item)` in
// Applications2/config.ts:277 — strip the trailing two `-` segments
// (replicaset hash + pod index) from `metric.pod` to recover the
// deployment name.
func podDeploymentName(pod string) string {
	if pod == "" {
		return ""
	}
	parts := strings.Split(pod, "-")
	if len(parts) <= 2 {
		return pod
	}
	return strings.Join(parts[:len(parts)-2], "-")
}

func fetchWorkloadsMetrics(ctx context.Context, c *Client, req workloadRequest, w monitoringWindow, now time.Time) ([]workloadAggregate, error) {
	if req.Sort == "" {
		req.Sort = "desc"
	}
	if req.SortBy == "" {
		req.SortBy = "cpu"
	}
	systemApps := make([]workloadApp, 0, len(req.Apps))
	customApps := make([]workloadApp, 0, len(req.Apps))
	for _, a := range req.Apps {
		if a.IsSystem {
			systemApps = append(systemApps, a)
		} else {
			customApps = append(customApps, a)
		}
	}

	podMetrics := []string{
		"pod_cpu_usage",
		"pod_memory_usage_wo_cache",
		"pod_net_bytes_transmitted",
		"pod_net_bytes_received",
	}
	// SPA also ships `namespace_pod_count` so we can populate the per-row
	// pod count without a second roundtrip — see
	// Applications2/config.ts:33 (NamespaceMetricTypes.pod_count).
	nsMetrics := []string{
		"namespace_cpu_usage",
		"namespace_memory_usage_wo_cache",
		"namespace_net_bytes_transmitted",
		"namespace_net_bytes_received",
		"namespace_pod_count",
	}

	type podsResp struct {
		data map[string]format.MonitoringResult
		err  error
	}
	type nsResp struct {
		data map[string]format.MonitoringResult
		err  error
	}
	podsCh := make(chan podsResp, 1)
	nsCh := make(chan nsResp, 1)

	// Pods endpoint — only meaningful for system apps (which are listed
	// by deployment regex inside a single namespace, the active user's
	// `user-space-<username>` ns).
	go func() {
		if len(systemApps) == 0 || req.UserNamespace == "" {
			podsCh <- podsResp{data: nil}
			return
		}
		filter := buildSystemDeploymentFilter(systemApps)
		q := monitoringQuery(podMetrics, w, now, false)
		q.Set("resources_filter", filter+"$")
		path := fmt.Sprintf("/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/%s/pods", url.PathEscape(req.UserNamespace))
		data, err := doMonitoring(ctx, c, path, q)
		podsCh <- podsResp{data: data, err: err}
	}()

	// Namespaces endpoint — for non-system (custom) apps each in its own
	// namespace.
	go func() {
		if len(customApps) == 0 {
			nsCh <- nsResp{data: nil}
			return
		}
		filter := buildCustomNamespaceFilter(customApps)
		q := monitoringQuery(nsMetrics, w, now, false)
		q.Set("resources_filter", filter)
		data, err := doMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces", q)
		nsCh <- nsResp{data: data, err: err}
	}()

	pods := <-podsCh
	ns := <-nsCh
	if pods.err != nil {
		return nil, pods.err
	}
	if ns.err != nil {
		return nil, ns.err
	}

	out := mergeWorkloadMetrics(req.Apps, pods.data, ns.data)
	sortWorkloadAggregates(out, req.SortBy, req.Sort)
	return out, nil
}

// buildSystemDeploymentFilter turns systemApps into the SPA's per-deployment
// pipe-joined regex with `.*` suffix per deployment, mirroring
// `resources_filter_system` in Applications2/config.ts.
func buildSystemDeploymentFilter(systemApps []workloadApp) string {
	var b strings.Builder
	for i, a := range systemApps {
		b.WriteString(a.Deployment)
		b.WriteString(".*")
		if i != len(systemApps)-1 {
			b.WriteString("|")
		}
	}
	return b.String()
}

// buildCustomNamespaceFilter mirrors `resources_filter_custom` — pipe-joined
// list of namespaces, no anchors.
func buildCustomNamespaceFilter(customApps []workloadApp) string {
	parts := make([]string, 0, len(customApps))
	for _, a := range customApps {
		if a.Namespace != "" {
			parts = append(parts, a.Namespace)
		}
	}
	return strings.Join(parts, "|")
}

// mergeWorkloadMetrics walks the apps list once, looking up each app's row
// in either the pod result (system apps, keyed by deployment via
// podDeploymentName(metric.pod)) or the namespace result (custom apps,
// keyed by metric.namespace). Last-sample values are pulled per-metric.
//
// The system-frontend specialcase mirrors `getTabOptions` in
// Applications2/config.ts:347 — multiple entrance apps share the same
// deployment, and each entrance gets a clone of the same metric.
func mergeWorkloadMetrics(apps []workloadApp, podData, nsData map[string]format.MonitoringResult) []workloadAggregate {
	// Aggregate per-deployment last samples up front (sum across pods of
	// the same deployment). Mirrors how `getTabOptions` ends up with one
	// row per deployment after the lodash chain finishes.
	podCPU := aggregateByDeployment(podData, "pod_cpu_usage")
	podMem := aggregateByDeployment(podData, "pod_memory_usage_wo_cache")
	podNetIn := aggregateByDeployment(podData, "pod_net_bytes_received")
	podNetOut := aggregateByDeployment(podData, "pod_net_bytes_transmitted")
	podCount := podCountByDeployment(podData)

	nsCPU := aggregateByNamespace(nsData, "namespace_cpu_usage")
	nsMem := aggregateByNamespace(nsData, "namespace_memory_usage_wo_cache")
	nsNetIn := aggregateByNamespace(nsData, "namespace_net_bytes_received")
	nsNetOut := aggregateByNamespace(nsData, "namespace_net_bytes_transmitted")
	nsPodCount := aggregateByNamespace(nsData, "namespace_pod_count")

	rows := make([]workloadAggregate, 0, len(apps))
	for _, a := range apps {
		row := workloadAggregate{
			Name:       a.Name,
			Title:      a.Title,
			Icon:       a.Icon,
			Namespace:  a.Namespace,
			Deployment: a.Deployment,
			OwnerKind:  a.OwnerKind,
			State:      a.State,
			IsSystem:   a.IsSystem,
		}
		if a.IsSystem {
			// System-frontend specialcase: many entrance apps map to
			// the same physical deployment, so the metric lookup
			// keys on the *deployment*, not the app name.
			key := a.Deployment
			row.CPU = podCPU[key]
			row.Memory = podMem[key]
			row.NetIn = podNetIn[key]
			row.NetOut = podNetOut[key]
			row.PodCount = podCount[key]
		} else {
			row.CPU = nsCPU[a.Namespace]
			row.Memory = nsMem[a.Namespace]
			row.NetIn = nsNetIn[a.Namespace]
			row.NetOut = nsNetOut[a.Namespace]
			row.PodCount = int(nsPodCount[a.Namespace])
		}
		rows = append(rows, row)
	}
	return rows
}

// aggregateByDeployment groups the per-pod result rows by deployment name
// (derived via podDeploymentName(metric.pod)) and sums the last sample of
// each pod under that deployment. Matches how the SPA ends up with one row
// per deployment after the lodash chain in `getTabOptions`.
func aggregateByDeployment(data map[string]format.MonitoringResult, metric string) map[string]float64 {
	out := map[string]float64{}
	mr, ok := data[metric]
	if !ok {
		return out
	}
	for _, r := range mr.Data.Result {
		dep := podDeploymentName(r.Metric["pod"])
		if dep == "" {
			continue
		}
		v, ok := lastValueOfRow(r.Value, r.Values)
		if !ok {
			continue
		}
		out[dep] += v
	}
	return out
}

// podCountByDeployment counts result rows per deployment from the
// pod_cpu_usage metric (one row per pod). Mirrors
// `podCountByDeploymentFromPodMetrics` in Applications2/config.ts:280.
func podCountByDeployment(data map[string]format.MonitoringResult) map[string]int {
	out := map[string]int{}
	mr, ok := data["pod_cpu_usage"]
	if !ok {
		return out
	}
	for _, r := range mr.Data.Result {
		dep := podDeploymentName(r.Metric["pod"])
		if dep == "" {
			continue
		}
		out[dep]++
	}
	return out
}

// aggregateByNamespace returns the last-sample of `metric` keyed by
// `metric.namespace`. Each namespace appears at most once per metric in
// the BFF response, so no summation is needed (unlike the per-pod path).
func aggregateByNamespace(data map[string]format.MonitoringResult, metric string) map[string]float64 {
	out := map[string]float64{}
	mr, ok := data[metric]
	if !ok {
		return out
	}
	for _, r := range mr.Data.Result {
		ns := r.Metric["namespace"]
		if ns == "" {
			continue
		}
		v, ok := lastValueOfRow(r.Value, r.Values)
		if !ok {
			continue
		}
		out[ns] = v
	}
	return out
}

// lastValueOfRow extracts the numeric last sample from either Values
// (matrix range) or Value (instant). Returns (v, true) on success.
func lastValueOfRow(value []interface{}, values [][]interface{}) (float64, bool) {
	if len(values) > 0 {
		row := values[len(values)-1]
		if len(row) >= 2 {
			return scalarFloat(row[1])
		}
	}
	if len(value) >= 2 {
		return scalarFloat(value[1])
	}
	return 0, false
}

func scalarFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case nil:
		return 0, false
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		if x == "" {
			return 0, false
		}
		var f float64
		if _, err := fmt.Sscanf(x, "%g", &f); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// sortWorkloadAggregates orders rows by `sortBy` (cpu|memory|net_in|net_out)
// in `dir` (asc|desc). Ties break on Title (mirrors the SPA's
// `orderBy(total, ['value','title'], [sort,'asc'])` in formatResult).
func sortWorkloadAggregates(rows []workloadAggregate, sortBy, dir string) {
	pick := func(r workloadAggregate) float64 {
		switch sortBy {
		case "memory":
			return r.Memory
		case "net_in":
			return r.NetIn
		case "net_out":
			return r.NetOut
		default:
			return r.CPU
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		vi := pick(rows[i])
		vj := pick(rows[j])
		if vi != vj {
			if dir == "asc" {
				return vi < vj
			}
			return vi > vj
		}
		ti := rows[i].Title
		if ti == "" {
			ti = rows[i].Name
		}
		tj := rows[j].Title
		if tj == "" {
			tj = rows[j].Name
		}
		return ti < tj
	})
}

// ----------------------------------------------------------------------------
// fetchSystemIFS — overview network (capi /system/ifs)
// ----------------------------------------------------------------------------

// SystemIFSItem mirrors the dashboard SystemIFSItem type: the union of
// fields the SPA's overview network page reads.
type SystemIFSItem struct {
	Iface             string `json:"iface"`
	IsHostIp          bool   `json:"isHostIp,omitempty"`
	Hostname          string `json:"hostname,omitempty"`
	Method            string `json:"method,omitempty"`
	MTU               any    `json:"mtu,omitempty"`
	IP                string `json:"ip,omitempty"`
	IPv4Mask          string `json:"ipv4Mask,omitempty"`
	IPv4Gateway       string `json:"ipv4Gateway,omitempty"`
	IPv4DNS           string `json:"ipv4DNS,omitempty"`
	IPv6Address       string `json:"ipv6Address,omitempty"`
	IPv6Gateway       string `json:"ipv6Gateway,omitempty"`
	IPv6DNS           string `json:"ipv6DNS,omitempty"`
	InternetConnected bool   `json:"internetConnected,omitempty"`
	IPv6Connectivity  bool   `json:"ipv6Connectivity,omitempty"`
	TxRate            any    `json:"txRate,omitempty"`
	RxRate            any    `json:"rxRate,omitempty"`
}

// fetchSystemIFS queries `/capi/system/ifs?testConnectivity=...`. The SPA's
// initial fetch passes testConnectivity=true so the server probes outgoing
// connectivity per-iface.
func fetchSystemIFS(ctx context.Context, c *Client, testConnectivity bool) ([]SystemIFSItem, error) {
	q := url.Values{}
	if testConnectivity {
		q.Set("testConnectivity", "true")
	} else {
		q.Set("testConnectivity", "false")
	}
	var raw []SystemIFSItem
	if err := c.DoJSON(ctx, http.MethodGet, "/capi/system/ifs", q, nil, &raw); err != nil {
		return nil, err
	}
	// SPA sorts isHostIp first.
	sort.SliceStable(raw, func(i, j int) bool {
		if raw[i].IsHostIp != raw[j].IsHostIp {
			return raw[i].IsHostIp
		}
		return false
	})
	return raw, nil
}

// ----------------------------------------------------------------------------
// fetchSystemFan — overview fan live (capi user-service)
// ----------------------------------------------------------------------------

// SystemFanData mirrors the SPA's getSystemFan response payload
// (.data field of /user-service/api/mdns/olares-one/cpu-gpu).
type SystemFanData struct {
	GPUFanSpeed    float64 `json:"gpu_fan_speed"`
	GPUTemperature float64 `json:"gpu_temperature"`
	CPUFanSpeed    float64 `json:"cpu_fan_speed"`
	CPUTemperature float64 `json:"cpu_temperature"`
}

func fetchSystemFan(ctx context.Context, c *Client) (*SystemFanData, error) {
	var raw struct {
		Data SystemFanData `json:"data"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/user-service/api/mdns/olares-one/cpu-gpu", nil, nil, &raw); err != nil {
		return nil, err
	}
	return &raw.Data, nil
}

// ----------------------------------------------------------------------------
// fetchAppsList — myapps_v2 (the SPA's appList store source)
// ----------------------------------------------------------------------------

// rawAppListItem mirrors the subset of `AppListItem`
// (controlPanelCommon/network/network.ts:280) the workload merge consumes.
// We tolerate unknown extra fields — the BFF freely adds new ones, and
// nothing in this struct has tag `json:"-"` to swallow them.
type rawAppListItem struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Title      string                   `json:"title"`
	Icon       string                   `json:"icon"`
	Namespace  string                   `json:"namespace"`
	Deployment string                   `json:"deployment"`
	OwnerKind  string                   `json:"ownerKind"`
	State      string                   `json:"state"`
	Entrances  []map[string]interface{} `json:"entrances"`
}

// ----------------------------------------------------------------------------
// lsblk tree (overview disk partitions)
// ----------------------------------------------------------------------------
//
// lsblkRow is the canonical shape extracted from one
// `node_disk_lsblk_info.data.result[].metric` entry. Mirrors the
// `LsblkMetricRow` type in `Overview2/Disk/config.ts:13`.

type lsblkRow struct {
	Name         string
	Node         string
	Pkname       string
	Size         string
	Fstype       string
	Mountpoint   string
	Fsused       string
	FsusePercent string
}

// lsblkFlatRow is one rendered row in the partitions table. `Depth` is 0
// for the root and increments per nesting level; `TreePrefix` is the
// ASCII-art prefix to prepend to `Name` for human display. `Parent`
// carries the resolved parent name so agents can rebuild the tree from
// the flat list without re-reading pkname / prefix logic.
type lsblkFlatRow struct {
	Row        lsblkRow
	Parent     string
	Depth      int
	TreePrefix string
}

// hasPknameLabels mirrors `Overview2/Disk/config.ts:267` —
// trim/non-empty pkname on at least one row turns the resolver onto the
// label-aware path; otherwise we fall back to prefix matching.
func hasPknameLabels(rows []lsblkRow) bool {
	for _, r := range rows {
		if strings.TrimSpace(r.Pkname) != "" {
			return true
		}
	}
	return false
}

// collectSubtreeByPkname BFS-walks the pkname graph from `rootName`,
// returning the rows in their original order. Mirrors
// `collectSubtreeByPkname` in Overview2/Disk/config.ts:273.
//
// When the BFS hits an empty `seen` set (e.g. root itself absent from
// the rows) the SPA recursively gathers descendants directly — we
// replicate that fallback so empty rooted views still produce sane data.
func collectSubtreeByPkname(allRows []lsblkRow, rootName string) []lsblkRow {
	byName := map[string]bool{}
	for _, r := range allRows {
		byName[r.Name] = true
	}
	seen := map[string]bool{}
	queue := []string{}
	if byName[rootName] {
		seen[rootName] = true
		queue = append(queue, rootName)
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, r := range allRows {
			pk := strings.TrimSpace(r.Pkname)
			if pk == n && !seen[r.Name] {
				seen[r.Name] = true
				queue = append(queue, r.Name)
			}
		}
	}
	if len(seen) == 0 {
		var addDesc func(parent string)
		addDesc = func(parent string) {
			for _, r := range allRows {
				pk := strings.TrimSpace(r.Pkname)
				if pk == parent && !seen[r.Name] {
					seen[r.Name] = true
					addDesc(r.Name)
				}
			}
		}
		addDesc(rootName)
	}
	out := make([]lsblkRow, 0, len(allRows))
	for _, r := range allRows {
		if seen[r.Name] {
			out = append(out, r)
		}
	}
	return out
}

// resolveParent picks the parent for a row, mirroring
// `Overview2/Disk/config.ts:313`:
//
//  1. root row has no parent
//  2. trimmed pkname wins if it points at a row in the set
//  3. otherwise pick the longest other-name prefix of `r.Name`
//  4. last-ditch fallback: the root itself
func resolveParent(r lsblkRow, rootName string, nameSet map[string]bool) string {
	if r.Name == rootName {
		return ""
	}
	pk := strings.TrimSpace(r.Pkname)
	if pk != "" && nameSet[pk] {
		return pk
	}
	bestPrefix := ""
	for n := range nameSet {
		if n == "" || n == r.Name {
			continue
		}
		if strings.HasPrefix(r.Name, n) && len(n) > len(bestPrefix) {
			bestPrefix = n
		}
	}
	if bestPrefix != "" {
		return bestPrefix
	}
	if nameSet[rootName] {
		return rootName
	}
	return ""
}

// buildLsblkTreePrefix mirrors the SPA's `buildLsblkTreePrefix`
// (Overview2/Disk/config.ts:332). `lastStack[i]==true` means the
// ancestor at depth `i` is the last sibling at that level — we draw
// "    " under it so the trunk doesn't dribble down past a finished
// sibling.
func buildLsblkTreePrefix(depth int, lastStack []bool) string {
	if depth == 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < depth-1; i++ {
		if i < len(lastStack) && lastStack[i] {
			b.WriteString("    ")
		} else {
			b.WriteString("│   ")
		}
	}
	if depth-1 < len(lastStack) && lastStack[depth-1] {
		b.WriteString("└── ")
	} else {
		b.WriteString("├── ")
	}
	return b.String()
}

// flattenLsblkHierarchy walks the rows pre-order, decorating each row
// with depth + tree prefix + resolved parent. Mirrors
// `Overview2/Disk/config.ts:342`. When `rootName` isn't in the row set
// we degrade to a flat list (no tree), matching the SPA.
func flattenLsblkHierarchy(rows []lsblkRow, rootName string) []lsblkFlatRow {
	nameSet := map[string]bool{}
	byName := map[string]lsblkRow{}
	for _, r := range rows {
		nameSet[r.Name] = true
		byName[r.Name] = r
	}
	if !nameSet[rootName] {
		sorted := append([]lsblkRow(nil), rows...)
		sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
		out := make([]lsblkFlatRow, len(sorted))
		for i, r := range sorted {
			out[i] = lsblkFlatRow{Row: r, Depth: 0, TreePrefix: ""}
		}
		return out
	}
	children := map[string][]lsblkRow{}
	parents := map[string]string{}
	for _, r := range rows {
		if r.Name == rootName {
			continue
		}
		p := resolveParent(r, rootName, nameSet)
		if p == "" || !nameSet[p] {
			p = rootName
		}
		parents[r.Name] = p
		children[p] = append(children[p], r)
	}
	for k := range children {
		c := children[k]
		sort.SliceStable(c, func(i, j int) bool { return c[i].Name < c[j].Name })
		children[k] = c
	}

	var out []lsblkFlatRow
	var walk func(rname string, depth int, lastStack []bool)
	walk = func(rname string, depth int, lastStack []bool) {
		r, ok := byName[rname]
		if !ok {
			return
		}
		prefix := ""
		if depth > 0 {
			prefix = buildLsblkTreePrefix(depth, lastStack)
		}
		out = append(out, lsblkFlatRow{
			Row:        r,
			Parent:     parents[rname],
			Depth:      depth,
			TreePrefix: prefix,
		})
		ch := children[rname]
		for idx, c := range ch {
			isLast := idx == len(ch)-1
			next := append(append([]bool(nil), lastStack...), isLast)
			walk(c.Name, depth+1, next)
		}
	}
	walk(rootName, 0, nil)
	return out
}

// fetchAppsList queries `/user-service/api/myapps_v2` and returns the
// SPA's `appsWithNamespace` selector (entries with at least one entrance).
// Empty entrance list ⇒ filtered out, mirroring `appsWithNamespace` in
// stores/AppList.ts.
func fetchAppsList(ctx context.Context, c *Client) ([]rawAppListItem, error) {
	var raw struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    []rawAppListItem `json:"data"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/user-service/api/myapps_v2", nil, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]rawAppListItem, 0, len(raw.Data))
	for _, it := range raw.Data {
		if len(it.Entrances) == 0 {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

// fanCurveTable is the hardcoded fan-curve specification — 1:1 with the
// SPA's `apps/packages/app/src/apps/dashboard/pages/Overview2/Fan/config.ts`
// `tableData` constant. NEVER drift from upstream without updating both
// sides; the iteration red-line in SKILL.md pins this.
var fanCurveTable = []fanCurveRow{
	{Step: 1, CPUFanRPM: 0, GPUFanRPM: 0, CPUTempRange: "0 - 54", GPUTempRange: "0 - 48"},
	{Step: 2, CPUFanRPM: 1100, GPUFanRPM: 1300, CPUTempRange: "47 - 64", GPUTempRange: "39 - 58"},
	{Step: 3, CPUFanRPM: 1300, GPUFanRPM: 1500, CPUTempRange: "54 - 71", GPUTempRange: "48 - 65"},
	{Step: 4, CPUFanRPM: 1500, GPUFanRPM: 1700, CPUTempRange: "64 - 74", GPUTempRange: "58 - 68"},
	{Step: 5, CPUFanRPM: 1800, GPUFanRPM: 2000, CPUTempRange: "71 - 77", GPUTempRange: "65 - 71"},
	{Step: 6, CPUFanRPM: 2100, GPUFanRPM: 2300, CPUTempRange: "74 - 80", GPUTempRange: "68 - 74"},
	{Step: 7, CPUFanRPM: 2300, GPUFanRPM: 2500, CPUTempRange: "77 - 83", GPUTempRange: "71 - 77"},
	{Step: 8, CPUFanRPM: 2300, GPUFanRPM: 2700, CPUTempRange: "80 - 86", GPUTempRange: "75 - 80"},
	{Step: 9, CPUFanRPM: 2700, GPUFanRPM: 2900, CPUTempRange: "83 - 88", GPUTempRange: "77 - 83"},
	{Step: 10, CPUFanRPM: 2900, GPUFanRPM: 3100, CPUTempRange: "86 - 96", GPUTempRange: "80 - 86"},
}

const (
	// fanSpeedMaxCPU / fanSpeedMaxGPU mirror the same constants in
	// Fan/config.ts. Used by overview fan live to expose the "max RPM"
	// column alongside the live RPM reading.
	fanSpeedMaxCPU = 2900
	fanSpeedMaxGPU = 3100
)

type fanCurveRow struct {
	Step         int
	CPUFanRPM    int
	GPUFanRPM    int
	CPUTempRange string
	GPUTempRange string
}

// ----------------------------------------------------------------------------
// fetchGraphicsList / fetchTaskList / fetchGraphicsDetail / fetchTaskDetail
// ----------------------------------------------------------------------------

// graphicsListBody mirrors the SPA's GraphicsListParams. The fields are
// emitted UNCONDITIONALLY (no `omitempty`) — HAMI's WebUI rejects a body
// missing the `filters` key with a 500 "unknown request error" because
// downstream code dereferences the (would-be) Filters struct without a
// nil guard. The SPA always sends `"filters": {}` (see
// `Overview2/GPU/GPUsTable.vue:195-201`); we match that wire shape.
//
// History: an earlier revision used `omitempty` on both fields. With a
// nil-input filter map, encoding/json emits `{"pageRequest":{...}}` —
// HAMI then panics, the gin recovery middleware returns a generic 5xx,
// and `olares-cli dashboard overview gpu` lights up `vgpu_unavailable`
// while the SPA in the same browser tab continues to render data.
// `TestGraphicsListBody_AlwaysIncludesFiltersKey` is the regression net
// for this.
type graphicsListBody struct {
	Filters     map[string]string `json:"filters"`
	PageRequest map[string]string `json:"pageRequest"`
}

func fetchGraphicsList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	if filters == nil {
		filters = map[string]string{}
	}
	body := graphicsListBody{
		Filters: filters,
		PageRequest: map[string]string{
			"sort":      "DESC",
			"sortField": "id",
		},
	}
	// HAMI returns the list at the TOP LEVEL: `{"list": [...]}` — there
	// is no `data` envelope around it. The SPA's `GraphicsListResponse`
	// type confirms this (`{ list: Graphics[] }`). Wrapping in a `data`
	// struct here used to silently produce "0 GPUs" even on machines
	// where the SPA shows devices.
	var raw struct {
		List []map[string]any `json:"list"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/gpus", nil, body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/gpus", ErrorKind: "http_4xx"}
	}
	if status >= 400 {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/gpus", Body: string(payload), ErrorKind: classifyKind(status)}
	}
	if err := decodeBytesMap(payload, &raw); err != nil {
		return nil, err
	}
	return raw.List, nil
}

func fetchTaskList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	if filters == nil {
		filters = map[string]string{}
	}
	body := graphicsListBody{
		Filters: filters,
		PageRequest: map[string]string{
			"sort":      "DESC",
			"sortField": "id",
		},
	}
	// HAMI returns `{"items": [...]}` at the top level (matches
	// `TaskListResponse`). No `data` envelope.
	var raw struct {
		Items []map[string]any `json:"items"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/containers", nil, body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/containers", ErrorKind: "http_4xx"}
	}
	if status >= 400 {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/containers", Body: string(payload), ErrorKind: classifyKind(status)}
	}
	if err := decodeBytesMap(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Items, nil
}

// fetchGraphicsDetail returns HAMI's `/v1/gpu` payload directly — the
// SPA's `GraphicsDetailsResponse` is a flat object, no `data` envelope.
func fetchGraphicsDetail(ctx context.Context, c *Client, uuid string) (map[string]any, error) {
	q := url.Values{"uuid": []string{uuid}}
	var raw map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/hami/api/vgpu/v1/gpu", q, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// fetchTaskDetail returns HAMI's `/v1/container` payload directly —
// like the GPU detail, the response is a flat object.
func fetchTaskDetail(ctx context.Context, c *Client, name, podUID, sharemode string) (map[string]any, error) {
	q := url.Values{"name": []string{name}, "podUid": []string{podUID}}
	if sharemode != "" {
		q.Set("sharemode", sharemode)
	}
	var raw map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/hami/api/vgpu/v1/container", q, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ----------------------------------------------------------------------------
// HAMI monitor query endpoints (instant-vector / range-vector)
// ----------------------------------------------------------------------------
//
// IMPORTANT — wire-shape gotcha:
//
// The "list/tasks/detail" HAMI endpoints return their payload at the TOP
// LEVEL (no `data` envelope; see fetchGraphicsList et al. above). The
// **monitor query endpoints** are the exception — they DO wrap the result
// in a single-level `data` field, matching the SPA's
// `InstantVectorResponse { data: InstantVector[] }` and
// `RangeVectorResponse { data: RangeVector[] }` types in
// src/apps/dashboard/types/gpu.ts. Do NOT "normalise" the wrapper away;
// `TestFetchInstantVector_ParsesDataEnvelope` /
// `TestFetchRangeVector_ParsesDataEnvelope` enforce this contract.

type instantVectorBody struct {
	Query string `json:"query"`
}

// instantVectorSample mirrors HAMI's `data[i]` row. Value is a number on
// the wire but float64 covers HAMI's full range (it caps at 1e308).
type instantVectorSample struct {
	Metric    map[string]string `json:"metric"`
	Value     float64           `json:"value"`
	Timestamp string            `json:"timestamp"`
}

// fetchInstantVector posts `query` to /hami/api/vgpu/v1/monitor/query/instant-vector
// and returns the decoded `data` array. HAMI returns one element per
// matching series; most CLI gauges just read `data[0]`, but a query
// can theoretically expand into >1 series so we hand back the slice.
func fetchInstantVector(ctx context.Context, c *Client, query string) ([]instantVectorSample, error) {
	body := instantVectorBody{Query: query}
	var raw struct {
		Data []instantVectorSample `json:"data"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/monitor/query/instant-vector", nil, body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, &HTTPError{
			Status:    status,
			URL:       c.BaseURL() + "/hami/api/vgpu/v1/monitor/query/instant-vector",
			Body:      string(payload),
			ErrorKind: classifyKind(status),
		}
	}
	if err := decodeBytesMap(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// rangeVectorRange mirrors the SPA's `RangeVectorParams.range` object.
type rangeVectorRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step"`
}

type rangeVectorBody struct {
	Query string           `json:"query"`
	Range rangeVectorRange `json:"range"`
}

// rangeVectorPoint is one (timestamp, value) pair inside a series. Both
// fields are strings on the wire (per the SPA's `RangeVector.values`
// type definition); the CLI parses them lazily on render.
type rangeVectorPoint struct {
	Value     any    `json:"value"`
	Timestamp string `json:"timestamp"`
}

type rangeVectorSeries struct {
	Metric map[string]string  `json:"metric"`
	Values []rangeVectorPoint `json:"values"`
}

// fetchRangeVector posts a range query (start/end/step) to HAMI's
// /v1/monitor/query/range-vector. SPA's `getStepWithTimeRange` builds
// `step` (a string like "30m"); ISO-formatted start/end are computed by
// the caller via `gpuTrendTimestampISO`.
func fetchRangeVector(ctx context.Context, c *Client, query, start, end, step string) ([]rangeVectorSeries, error) {
	body := rangeVectorBody{
		Query: query,
		Range: rangeVectorRange{Start: start, End: end, Step: step},
	}
	var raw struct {
		Data []rangeVectorSeries `json:"data"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/monitor/query/range-vector", nil, body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, &HTTPError{
			Status:    status,
			URL:       c.BaseURL() + "/hami/api/vgpu/v1/monitor/query/range-vector",
			Body:      string(payload),
			ErrorKind: classifyKind(status),
		}
	}
	if err := decodeBytesMap(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// gpuTrendStep is a 1:1 port of `timeRangeFormate(diff_s, 16)` from
// packages/app/src/apps/controlPanelCommon/containers/Monitoring/utils.js.
// Algorithm:
//
//  1. Convert (end-start) to whole minutes (rounded down).
//  2. If the minutes count matches one of the SPA's preset windows
//     (10/20/30, 60, 120, 180, 300, 480, 720, 1440, 4320, 10080), use
//     the matching preset step (1m / 1m / 1m / 10m / 20m / 10m / 10m /
//     30m / 30m / 60m / 60m / 60m).
//  3. Otherwise compute `floor(minutes/16)m` (same as `getStep(value, 16)`),
//     then enforce a [1m..60m] range (the SPA bumps a 0m result to a 10x
//     coarser window and caps anything >60m at 60m).
//
// We accept time.Time start/end (instead of duration) so the test cases
// exactly match the SPA call sites in GPUsDetails.vue / TasksDetails.vue.
func gpuTrendStep(start, end time.Time) string {
	totalMinutes := int(end.Sub(start) / time.Minute)
	if totalMinutes <= 0 {
		// Defensive — empty / inverted ranges: pick the smallest sane
		// step so the caller doesn't divide by zero downstream.
		return "1m"
	}
	if step, ok := gpuStepPreset(totalMinutes); ok {
		return step
	}
	step := totalMinutes / 16
	if step < 1 {
		// SPA fallback: bump to a 10-bucket window when 16-bucket
		// rounds to 0m. (See the `if (stepNum < 1) { times = 10 }`
		// branch in utils.js.)
		step = totalMinutes / 10
		if step < 1 {
			step = 1
		}
	}
	if step > 60 {
		step = 60
	}
	return fmt.Sprintf("%dm", step)
}

// gpuStepPreset reproduces the `timeReflection` table in utils.js.
// Returns the preset step + true when minutes match a known window.
func gpuStepPreset(minutes int) (string, bool) {
	switch minutes {
	case 10, 20, 30:
		return "1m", true
	case 60:
		return "10m", true
	case 120:
		return "20m", true
	case 180, 300:
		return "10m", true
	case 480, 720:
		return "30m", true
	case 1440, 4320, 10080:
		return "60m", true
	default:
		return "", false
	}
}

// gpuTrendTimestampISO formats a time.Time the way the SPA's
// `timeParse(date)` does for monitor queries: `YYYY-MM-DD HH:mm:ss` in
// the caller's timezone (no offset suffix). HAMI's WebUI accepts
// either Unix-seconds or this human-readable form; the SPA exclusively
// sends the latter, so we match.
func gpuTrendTimestampISO(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ----------------------------------------------------------------------------
// Misc helpers
// ----------------------------------------------------------------------------

// classifyKind maps an HTTP status code to the error_kind enum surfaced via
// Meta.ErrorKind / HTTPError.ErrorKind. Centralised so leaves don't drift.
func classifyKind(status int) string {
	switch {
	case status >= 500:
		return "http_5xx"
	case status >= 400:
		return "http_4xx"
	default:
		return ""
	}
}

// renderTemperature picks the right unit suffix for ConvertTemperature.
// Used by overview cpu / overview fan live.
func renderTemperature(celsius float64, target format.TempUnit) string {
	v := format.ConvertTemperature(celsius, target)
	suffix := "°C"
	switch target {
	case format.TempF:
		suffix = "°F"
	case format.TempK:
		suffix = "K"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", v), "0"), ".") + suffix
}

// percentString formats a 0..1 ratio as "N.NN%" (SPA style — utilisation
// metrics are percent of unit interval).
func percentString(ratio float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", ratio*100), "0"), ".") + "%"
}

// percentDirect formats a value already expressed as a percent (e.g. HAMI
// returns `coreUtilizedPercent: 25.5`, NOT 0.255). The SPA renders these
// with `round(val, 2) + '%'`; we match that and trim trailing zeros for
// readability ("25%" instead of "25.00%").
func percentDirect(pct float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", pct), "0"), ".") + "%"
}

// gpuModeLabel translates HAMI's `shareMode` string into the SPA-rendered
// label. SPA mapping (constant/index.ts:VRAMModeLabel):
//
//	"0" → "App exclusive"
//	"1" → "Memory slicing"
//	"2" → "Time slicing"
//
// Real fixtures sometimes return "3" (observed on `olarestest005` —
// HAMI's WebUI silently falls back to showing the raw value). To avoid
// surfacing an empty cell in the CLI table we pass unknown values
// through unchanged, prefixed with `mode=` so a human can tell that we
// preserved the wire byte instead of mistranslating.
func gpuModeLabel(raw any) string {
	s := fmt.Sprintf("%v", raw)
	switch s {
	case "0":
		return "App exclusive"
	case "1":
		return "Memory slicing"
	case "2":
		return "Time slicing"
	case "":
		return "-"
	default:
		return "mode=" + s
	}
}

// gpuHealthLabel turns HAMI's boolean `health` into a human-readable
// status. The SPA leaves it as "true"/"false"; we surface the friendlier
// "healthy"/"unhealthy" pair (raw envelope still carries the original
// bool for agents that prefer the wire shape).
func gpuHealthLabel(raw any) string {
	if b, ok := raw.(bool); ok {
		if b {
			return "healthy"
		}
		return "unhealthy"
	}
	return fmt.Sprintf("%v", raw)
}

// firstAnyInArray returns the first element of a slice-shaped value
// (e.g. `[]any` or `[]string`) decoded from JSON. HAMI returns
// per-device fields like `devicesCoreUtilizedPercent` as arrays — the
// SPA uses `val[0]` because tasks observed in the wild only ever bind
// a single device. We mirror that decision here, while preserving the
// full slice in `Raw` for multi-GPU consumers down the line.
func firstAnyInArray(v any) any {
	switch x := v.(type) {
	case []any:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	case []string:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	case []float64:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	default:
		return nil
	}
}

// gpuVRAMHuman formats a MiB count (HAMI's `memoryTotal` / `memoryUsed`
// units) as a SPA-style "1.5GiB"-shaped string. Mirrors the SPA's
// `getDiskSize(val * 1024 * 1024)` call; treats 0 as "-" so the table
// doesn't show a misleading "0B" for honest "no allocation" cases.
func gpuVRAMHuman(mibVal any) string {
	mib := toFloat(mibVal)
	if mib <= 0 {
		return "-"
	}
	bytes := mib * 1024.0 * 1024.0
	return format.GetDiskSize(strconv.FormatFloat(bytes, 'f', -1, 64))
}

// ----------------------------------------------------------------------------
// JSON encode/decode shims (kept private so leaves don't import encoding/json
// directly — easier to refactor later)
// ----------------------------------------------------------------------------

func encodeJSONMap(v interface{}) ([]byte, error)   { return jsonMarshal(v) }
func decodeBytesMap(b []byte, v interface{}) error  { return jsonUnmarshal(b, v) }

// jsonMarshal / jsonUnmarshal are tiny shims so this file doesn't pull in
// "encoding/json" at the top level. The actual implementation lives in
// json_shim.go to keep helpers.go tidy.

// User helpers ---------------------------------------------------------------

// resolveTargetUser returns the user to operate on for `--user`-aware
// commands (overview user, etc.).
//
//   - explicit (positional or --user) is honoured if the active profile is
//     admin; non-admins targeting a third party get ErrUserScope.
//   - empty falls back to the active profile's identity.
func resolveTargetUser(ctx context.Context, c *Client, requested string) (*UserDetail, error) {
	u, err := c.EnsureUser(ctx)
	if err != nil {
		return nil, err
	}
	if requested == "" || requested == u.Name {
		return u, nil
	}
	if !u.IsAdmin() {
		return nil, fmt.Errorf("--user %q requires platform-admin; %s does not have that role", requested, u.Name)
	}
	// Admin path — return a synthetic UserDetail so callers can use
	// .Name without re-fetching IAM; downstream helpers only use Name.
	return &UserDetail{Name: requested, GlobalRole: "<admin-target>"}, nil
}
