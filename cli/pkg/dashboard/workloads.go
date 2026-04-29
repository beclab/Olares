package dashboard

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// ----------------------------------------------------------------------------
// FetchWorkloadsMetrics — overview ranking + applications (workload-grain)
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

// WorkloadAggregate is the merged shape returned by FetchWorkloadsMetrics.
// Each row carries the application's logical identity (name / title /
// namespace / system flag / owner) plus the four metric values plus the
// running-pod count for the row. State / Title / Icon ride along so the
// CLI can present the same per-app card the SPA does without a second
// round-trip.
type WorkloadAggregate struct {
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

// WorkloadRequest is what callers feed into FetchWorkloadsMetrics:
// a flattened list of registered apps + the active user's namespace.
type WorkloadRequest struct {
	Apps          []WorkloadApp
	UserNamespace string
	Sort          string // "asc" | "desc"
	SortBy        string // "cpu" (default) | "memory" | "net_in" | "net_out"
}

// WorkloadApp is one row of the SPA's app inventory (post `entrances`
// filter). Mirrors `AppListItem` from
// `controlPanelCommon/network/network.ts:280` for the fields the workload
// merge actually reads.
type WorkloadApp struct {
	Name       string
	Title      string
	Icon       string
	Namespace  string
	Deployment string
	OwnerKind  string
	State      string
	IsSystem   bool
}

// SystemFrontendDeployment mirrors the SPA's `SYSTEM_FRONTEND_DEPLOYMENT`
// constant (Applications2/config.ts:25). Multiple "entrance" apps share
// the same physical deployment; the merge step has to clone the metric to
// each app entry so the cards line up with what users see in the UI.
const SystemFrontendDeployment = "system-frontend-deployment"

// PodDeploymentName mirrors `podDeploymentName(item)` in
// Applications2/config.ts:277 — strip the trailing two `-` segments
// (replicaset hash + pod index) from `metric.pod` to recover the
// deployment name.
func PodDeploymentName(pod string) string {
	if pod == "" {
		return ""
	}
	parts := strings.Split(pod, "-")
	if len(parts) <= 2 {
		return pod
	}
	return strings.Join(parts[:len(parts)-2], "-")
}

// FetchWorkloadsMetrics fans out two BFF calls (pods inside the user's
// namespace + per-namespace summary), merges them into a per-app
// aggregate, and orders the result.
func FetchWorkloadsMetrics(ctx context.Context, c *Client, cf *CommonFlags, req WorkloadRequest, w MonitoringWindow, now time.Time) ([]WorkloadAggregate, error) {
	if req.Sort == "" {
		req.Sort = "desc"
	}
	if req.SortBy == "" {
		req.SortBy = "cpu"
	}
	systemApps := make([]WorkloadApp, 0, len(req.Apps))
	customApps := make([]WorkloadApp, 0, len(req.Apps))
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
		filter := BuildSystemDeploymentFilter(systemApps)
		q := MonitoringQuery(cf, podMetrics, w, now, false)
		q.Set("resources_filter", filter+"$")
		path := fmt.Sprintf("/kapis/monitoring.kubesphere.io/v1alpha3/namespaces/%s/pods", url.PathEscape(req.UserNamespace))
		data, err := DoMonitoring(ctx, c, path, q)
		podsCh <- podsResp{data: data, err: err}
	}()

	// Namespaces endpoint — for non-system (custom) apps each in its own
	// namespace.
	go func() {
		if len(customApps) == 0 {
			nsCh <- nsResp{data: nil}
			return
		}
		filter := BuildCustomNamespaceFilter(customApps)
		q := MonitoringQuery(cf, nsMetrics, w, now, false)
		q.Set("resources_filter", filter)
		data, err := DoMonitoring(ctx, c, "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces", q)
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

	out := MergeWorkloadMetrics(req.Apps, pods.data, ns.data)
	SortWorkloadAggregates(out, req.SortBy, req.Sort)
	return out, nil
}

// BuildSystemDeploymentFilter turns systemApps into the SPA's per-deployment
// pipe-joined regex with `.*` suffix per deployment, mirroring
// `resources_filter_system` in Applications2/config.ts.
func BuildSystemDeploymentFilter(systemApps []WorkloadApp) string {
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

// BuildCustomNamespaceFilter mirrors `resources_filter_custom` — pipe-joined
// list of namespaces, no anchors.
func BuildCustomNamespaceFilter(customApps []WorkloadApp) string {
	parts := make([]string, 0, len(customApps))
	for _, a := range customApps {
		if a.Namespace != "" {
			parts = append(parts, a.Namespace)
		}
	}
	return strings.Join(parts, "|")
}

// MergeWorkloadMetrics walks the apps list once, looking up each app's row
// in either the pod result (system apps, keyed by deployment via
// PodDeploymentName(metric.pod)) or the namespace result (custom apps,
// keyed by metric.namespace). Last-sample values are pulled per-metric.
//
// The system-frontend specialcase mirrors `getTabOptions` in
// Applications2/config.ts:347 — multiple entrance apps share the same
// deployment, and each entrance gets a clone of the same metric.
func MergeWorkloadMetrics(apps []WorkloadApp, podData, nsData map[string]format.MonitoringResult) []WorkloadAggregate {
	podCPU := AggregateByDeployment(podData, "pod_cpu_usage")
	podMem := AggregateByDeployment(podData, "pod_memory_usage_wo_cache")
	podNetIn := AggregateByDeployment(podData, "pod_net_bytes_received")
	podNetOut := AggregateByDeployment(podData, "pod_net_bytes_transmitted")
	podCount := PodCountByDeployment(podData)

	nsCPU := AggregateByNamespace(nsData, "namespace_cpu_usage")
	nsMem := AggregateByNamespace(nsData, "namespace_memory_usage_wo_cache")
	nsNetIn := AggregateByNamespace(nsData, "namespace_net_bytes_received")
	nsNetOut := AggregateByNamespace(nsData, "namespace_net_bytes_transmitted")
	nsPodCount := AggregateByNamespace(nsData, "namespace_pod_count")

	rows := make([]WorkloadAggregate, 0, len(apps))
	for _, a := range apps {
		row := WorkloadAggregate{
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

// AggregateByDeployment groups the per-pod result rows by deployment name
// (derived via PodDeploymentName(metric.pod)) and sums the last sample of
// each pod under that deployment. Matches how the SPA ends up with one row
// per deployment after the lodash chain in `getTabOptions`.
func AggregateByDeployment(data map[string]format.MonitoringResult, metric string) map[string]float64 {
	out := map[string]float64{}
	mr, ok := data[metric]
	if !ok {
		return out
	}
	for _, r := range mr.Data.Result {
		dep := PodDeploymentName(r.Metric["pod"])
		if dep == "" {
			continue
		}
		v, ok := LastValueOfRow(r.Value, r.Values)
		if !ok {
			continue
		}
		out[dep] += v
	}
	return out
}

// PodCountByDeployment counts result rows per deployment from the
// pod_cpu_usage metric (one row per pod). Mirrors
// `podCountByDeploymentFromPodMetrics` in Applications2/config.ts:280.
func PodCountByDeployment(data map[string]format.MonitoringResult) map[string]int {
	out := map[string]int{}
	mr, ok := data["pod_cpu_usage"]
	if !ok {
		return out
	}
	for _, r := range mr.Data.Result {
		dep := PodDeploymentName(r.Metric["pod"])
		if dep == "" {
			continue
		}
		out[dep]++
	}
	return out
}

// AggregateByNamespace returns the last-sample of `metric` keyed by
// `metric.namespace`. Each namespace appears at most once per metric in
// the BFF response, so no summation is needed (unlike the per-pod path).
func AggregateByNamespace(data map[string]format.MonitoringResult, metric string) map[string]float64 {
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
		v, ok := LastValueOfRow(r.Value, r.Values)
		if !ok {
			continue
		}
		out[ns] = v
	}
	return out
}

// LastValueOfRow extracts the numeric last sample from either Values
// (matrix range) or Value (instant). Returns (v, true) on success.
func LastValueOfRow(value []interface{}, values [][]interface{}) (float64, bool) {
	if len(values) > 0 {
		row := values[len(values)-1]
		if len(row) >= 2 {
			return ScalarFloat(row[1])
		}
	}
	if len(value) >= 2 {
		return ScalarFloat(value[1])
	}
	return 0, false
}

// ScalarFloat turns a JSON-decoded scalar into a float64. Returns
// (0, false) for nil / empty / unparsable input.
func ScalarFloat(v interface{}) (float64, bool) {
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

// SortWorkloadAggregates orders rows by `sortBy` (cpu|memory|net_in|net_out)
// in `dir` (asc|desc). Ties break on Title (mirrors the SPA's
// `orderBy(total, ['value','title'], [sort,'asc'])` in formatResult).
func SortWorkloadAggregates(rows []WorkloadAggregate, sortBy, dir string) {
	pick := func(r WorkloadAggregate) float64 {
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
