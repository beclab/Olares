package dashboard

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// Capability gates (overview fan / overview gpu)
// ----------------------------------------------------------------------------
//
// Both subtrees mirror the SPA's hard gates from
// `Overview2/ClusterResource.vue` (line 232-238 + 278-293):
//
//   Fan: only Olares One hardware
//     → device_name == "Olares One" via /user-service/api/system/status
//   GPU: admin AND any node carries a gpu.bytetrade.io/* GPU label
//     → cluster role check + label scan on /kapis/.../nodes
//
// The CLI replicates these gates so agents see a structured empty
// envelope (with EmptyReason / Note / DeviceName) instead of a
// "silently zero" payload from the BFF.

// GPUVendor names the exporter a card's numbers come from. The three
// vendors publish unrelated metric families, so nothing downstream can stay
// vendor-blind.
type GPUVendor string

const (
	VendorNVIDIA GPUVendor = "nvidia"
	VendorIntel  GPUVendor = "intel"
	VendorAMD    GPUVendor = "amd"
)

// vendorByModeLabel maps the per-mode node label to the vendor whose
// exporter it turns on. Mirrors VENDOR_BY_MODE_LABEL in the SPA's
// utils/gpuVendor.ts; the two lists have to agree or the CLI and the page
// disagree about which cards exist.
//
// The integrated-GPU modes (`gpu.bytetrade.io/intel`, `gpu.bytetrade.io/amd`)
// are left out on purpose: xpumd drops `hw_gpu_type=integrated` and the AMD
// exporter is only installed for discrete cards, so an iGPU-only node has no
// series to render at all.
var vendorByModeLabel = []struct {
	Label  string
	Vendor GPUVendor
}{
	{"gpu.bytetrade.io/nvidia", VendorNVIDIA},
	{"gpu.bytetrade.io/nvidia-gb10", VendorNVIDIA},
	{"gpu.bytetrade.io/intel-gpu", VendorIntel},
	{"gpu.bytetrade.io/amd-gpu", VendorAMD},
}

// VendorsOfNodeLabels returns every vendor a single node's labels turn on.
// A node can carry a discrete AMD card next to an Intel one, so this reports
// all matches rather than stopping at the first.
//
// `cuda-supported` counts as NVIDIA alongside the per-mode labels. It is the
// older marker, written since before the per-mode labels existed, and a node
// labelled by an Olares from that era carries nothing else — without it such a
// node drops out of the inventory and the CLI reports "no GPU" on a machine
// that plainly has one.
//
// The SPA dropped this fallback and we deliberately do not follow it there:
// the SPA ships with the system, so a build new enough to have dropped it only
// ever runs against a cluster new enough to carry the per-mode labels. This
// binary is installed separately and talks to whatever cluster the profile
// points at, so it has to keep reading the old marker.
func VendorsOfNodeLabels(labels map[string]string) []GPUVendor {
	var out []GPUVendor
	seen := map[GPUVendor]bool{}
	add := func(v GPUVendor) {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	for _, m := range vendorByModeLabel {
		if labels[m.Label] == "true" {
			add(m.Vendor)
		}
	}
	if labels["gpu.bytetrade.io/cuda-supported"] == "true" {
		add(VendorNVIDIA)
	}
	return out
}

// gpuNodeMu / gpuNodeCache cache the node label scan for the duration of the
// *Client. We attach the cache to a per-Client map keyed by the client
// pointer so tests with multiple fixtures don't share state, but in
// production each invocation has a single Client so it's a free win.
var (
	gpuNodeMu    sync.Mutex
	gpuNodeCache = map[*Client]gpuNodeResult{}
)

// ResetGPUNodeCache forgets the cached label scan for `c`. Test-only escape
// hatch — production code never calls this; tests use it between fixtures to
// guarantee each scenario re-issues the upstream label scan.
func ResetGPUNodeCache(c *Client) {
	gpuNodeMu.Lock()
	delete(gpuNodeCache, c)
	gpuNodeMu.Unlock()
}

type gpuNodeResult struct {
	// nodesByVendor records which nodes turned each vendor on, so a detail
	// view can scope its query to the one node holding the card instead of
	// asking every node in the cluster.
	nodesByVendor map[GPUVendor][]string
	err           error
	done          bool
}

// GPUInventory is what a single node-label scan can tell us: which vendors
// are present and, for each, the nodes carrying them.
type GPUInventory struct {
	Vendors       []GPUVendor
	NodesByVendor map[GPUVendor][]string
}

// Has reports whether the cluster carries any card from `v`.
func (inv GPUInventory) Has(v GPUVendor) bool {
	for _, got := range inv.Vendors {
		if got == v {
			return true
		}
	}
	return false
}

// DetectGPUVendors scans node labels once and reports every GPU vendor the
// cluster carries, in the fixed nvidia/intel/amd order so output is stable
// across invocations. Cached per-Client, so the second call inside the same
// CLI invocation is free; the label-only fast path keeps payloads small even
// on large clusters since we just need to know which modes are enabled.
func DetectGPUVendors(ctx context.Context, c *Client) (GPUInventory, error) {
	r, err := scanGPUNodes(ctx, c)
	if err != nil {
		return GPUInventory{}, err
	}
	inv := GPUInventory{NodesByVendor: r.nodesByVendor}
	for _, m := range vendorByModeLabel {
		if len(r.nodesByVendor[m.Vendor]) > 0 && !inv.Has(m.Vendor) {
			inv.Vendors = append(inv.Vendors, m.Vendor)
		}
	}
	return inv, nil
}

// scanGPUNodes performs the one label scan every GPU gate reads from, and
// caches it per-Client. A node's labels are the only thing that says which
// cards it carries, so this one payload answers every vendor question the
// subtree asks.
func scanGPUNodes(ctx context.Context, c *Client) (gpuNodeResult, error) {
	gpuNodeMu.Lock()
	if r, ok := gpuNodeCache[c]; ok && r.done {
		gpuNodeMu.Unlock()
		return r, r.err
	}
	gpuNodeMu.Unlock()

	var raw struct {
		Items []struct {
			Metadata struct {
				Name   string            `json:"name"`
				Labels map[string]string `json:"labels"`
			} `json:"metadata"`
		} `json:"items"`
	}
	q := url.Values{"sortBy": []string{"createTime"}}
	err := c.DoJSON(ctx, http.MethodGet, "/kapis/resources.kubesphere.io/v1alpha3/nodes", q, nil, &raw)
	res := gpuNodeResult{nodesByVendor: map[GPUVendor][]string{}, err: err, done: true}
	if err == nil {
		for _, it := range raw.Items {
			for _, v := range VendorsOfNodeLabels(it.Metadata.Labels) {
				res.nodesByVendor[v] = append(res.nodesByVendor[v], it.Metadata.Name)
			}
		}
	}
	gpuNodeMu.Lock()
	gpuNodeCache[c] = res
	gpuNodeMu.Unlock()
	return res, err
}

// metaTimeAt returns `now` projected into cf.Timezone (defaulting to
// time.Local when cf or its Timezone is nil). Centralised here so the
// gate / fetcher helpers don't each re-derive the projection.
func metaTimeAt(cf *CommonFlags, now time.Time) time.Time {
	if cf == nil || cf.Timezone == nil {
		return now
	}
	return now.In(cf.Timezone.Time())
}

// userOf returns cf.User defensively (cf may be nil in tests).
func userOf(cf *CommonFlags) string {
	if cf == nil {
		return ""
	}
	return cf.User
}

// stderrOr resolves a writable stderr target, falling back to
// os.Stderr when the caller hasn't supplied one. Convenience for the
// gate helpers that scribble human hints alongside the envelope.
func stderrOr(w io.Writer) io.Writer {
	if w == nil {
		return os.Stderr
	}
	return w
}

// GateOlaresOne returns (gatedEnvelope, true) when the active device
// is not Olares One; the caller should emit `gatedEnvelope` and skip
// any data fetch. The hint message is also written to `stderr` (when
// non-nil and `cf.Output != OutputJSON`) so humans see why the table
// is empty.
//
// On error from EnsureSystemStatus we let the caller proceed (gated=false,
// nil envelope) — the downstream BFF call will surface the real error
// itself rather than masking it with a confused "not Olares One" hint.
func GateOlaresOne(ctx context.Context, c *Client, cf *CommonFlags, kind string, now time.Time, stderr io.Writer) (Envelope, bool) {
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
		Meta: NewMeta(metaTimeAt(cf, now), c.OlaresID(), userOf(cf)),
	}
	env.Meta.Empty = true
	env.Meta.EmptyReason = "not_olares_one"
	env.Meta.Note = "Fan / cooling integration is only available on Olares One devices"
	env.Meta.DeviceName = dev
	if cf != nil && cf.Output != OutputJSON {
		fmt.Fprintf(stderrOr(stderr), "fan is only available on Olares One devices (current: %s)\n", dev)
	}
	return env, true
}

// GPUAdvisory is the soft-gate companion to GateOlaresOne. The SPA's
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
//   - non-admin profile → "gpu_sidebar_hidden_non_admin"
//   - no GPU-mode node  → "gpu_sidebar_hidden_no_gpu_node"
//
// Both are advisory-only; the caller continues to fetch and renders
// data when an exporter returns it. Returns (note, "") when no advisory
// applies, or (note, reason) — both empty when EnsureUser /
// DetectGPUVendors fail (we fall silent rather than misleading agents).
func GPUAdvisory(ctx context.Context, c *Client, cf *CommonFlags, stderr io.Writer) (note, reason string) {
	u, err := c.EnsureUser(ctx)
	if err != nil || u == nil {
		return "", ""
	}
	if !u.IsAdmin() {
		if cf != nil && cf.Output != OutputJSON {
			fmt.Fprintf(stderrOr(stderr),
				"(advisory) GPU sidebar entry is hidden for non-admin profiles in the SPA; current user (%s) is %s\n",
				u.Name, DisplayRole(u.GlobalRole))
		}
		return "GPU sidebar entry is hidden for non-admin profiles in the SPA; HAMI was queried directly", "gpu_sidebar_hidden_non_admin"
	}
	// The card is hidden when NO vendor is present, not when NVIDIA
	// specifically is absent. Gating on `cuda-supported` alone made every
	// Intel and AMD machine report "no GPU" while the page listed its cards.
	inv, err := DetectGPUVendors(ctx, c)
	if err != nil {
		return "", ""
	}
	if len(inv.Vendors) == 0 {
		if cf != nil && cf.Output != OutputJSON {
			fmt.Fprintln(stderrOr(stderr),
				"(advisory) no node carries a gpu.bytetrade.io/* GPU label; SPA hides the GPU card")
		}
		return "no node carries a gpu.bytetrade.io/* GPU label; SPA hides the GPU card", "gpu_sidebar_hidden_no_gpu_node"
	}
	return "", ""
}

// VgpuUnavailableFromError converts a HAMI-side error into the
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
func VgpuUnavailableFromError(c *Client, cf *CommonFlags, err error, kind string, now time.Time, stderr io.Writer) (Envelope, bool) {
	he, ok := IsHTTPError(err)
	if !ok || he.Status < 500 || he.Status >= 600 {
		return Envelope{}, false
	}
	msg := ExtractHAMIMessage(he.Body)
	env := Envelope{
		Kind: kind,
		Meta: NewMeta(metaTimeAt(cf, now), c.OlaresID(), userOf(cf)),
	}
	env.Meta.Empty = true
	env.Meta.EmptyReason = "vgpu_unavailable"
	env.Meta.Note = "HAMI vGPU controller responded with " + http.StatusText(he.Status) + "; the integration is installed but unhealthy"
	env.Meta.HTTPStatus = he.Status
	if msg != "" {
		env.Meta.Error = msg
	}
	if cf != nil && cf.Output != OutputJSON {
		w := stderrOr(stderr)
		if msg != "" {
			fmt.Fprintf(w,
				"gpu data temporarily unavailable: HAMI returned HTTP %d (%s)\n",
				he.Status, msg)
		} else {
			fmt.Fprintf(w,
				"gpu data temporarily unavailable: HAMI returned HTTP %d\n",
				he.Status)
		}
	}
	return env, true
}

// ExtractHAMIMessage tries to surface the `message` field from a HAMI
// JSON-shaped body (`{"code": <int>, "message": "..."}`); falls back
// to the trimmed body itself capped at 256 bytes. Caller pre-strips
// the body via the *HTTPError struct.
func ExtractHAMIMessage(body string) string {
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

// DisplayRole pretty-prints an empty / unknown role string for the
// stderr hint so humans see "(unset)" rather than two consecutive
// spaces.
func DisplayRole(r string) string {
	if strings.TrimSpace(r) == "" {
		return "(unset)"
	}
	return r
}
