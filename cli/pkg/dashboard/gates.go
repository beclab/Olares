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
//   GPU: admin AND any node has gpu.bytetrade.io/cuda-supported=true
//     → cluster role check + label scan on /kapis/.../nodes
//
// The CLI replicates these gates so agents see a structured empty
// envelope (with EmptyReason / Note / DeviceName) instead of a
// "silently zero" payload from the BFF.

// cudaNodeMu / cudaNodeCache cache the result of HasCUDANode for the
// duration of the *Client. We attach the cache to a per-Client map keyed
// by the client pointer so tests with multiple fixtures don't share state,
// but in production each invocation has a single Client so it's a free win.
var (
	cudaNodeMu    sync.Mutex
	cudaNodeCache = map[*Client]cudaNodeResult{}
)

// ResetCUDANodeCache forgets the cached HasCUDANode result for `c`.
// Test-only escape hatch — production code never calls this; tests use
// it between fixtures to guarantee each scenario re-issues the upstream
// label scan.
func ResetCUDANodeCache(c *Client) {
	cudaNodeMu.Lock()
	delete(cudaNodeCache, c)
	cudaNodeMu.Unlock()
}

type cudaNodeResult struct {
	present bool
	err     error
	done    bool
}

// HasCUDANode reports whether the cluster has at least one node with
// label `gpu.bytetrade.io/cuda-supported=true`. Mirrors the SPA's
// `checkGpu` (Overview2/ClusterResource.vue:278-293) which iterates
// the nodes list and looks for the cuda-supported label.
//
// Cached per-Client; the second call inside the same CLI invocation is
// free. The label-only fast path keeps payloads small even on large
// clusters since we just need a presence check.
func HasCUDANode(ctx context.Context, c *Client) (bool, error) {
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
//   - non-admin profile  → "gpu_sidebar_hidden_non_admin"
//   - no CUDA-capable node → "gpu_sidebar_hidden_no_cuda_node"
//
// Both are advisory-only; the caller continues to fetch and renders
// data when HAMI returns it. Returns (note, "") when no advisory
// applies, or (note, reason) — both empty when EnsureUser /
// HasCUDANode fail (we fall silent rather than misleading agents).
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
	present, err := HasCUDANode(ctx, c)
	if err != nil {
		return "", ""
	}
	if !present {
		if cf != nil && cf.Output != OutputJSON {
			fmt.Fprintln(stderrOr(stderr),
				"(advisory) no node carries gpu.bytetrade.io/cuda-supported=true; SPA hides the GPU card. HAMI was queried directly")
		}
		return "no node carries gpu.bytetrade.io/cuda-supported=true; SPA hides the GPU card. HAMI was queried directly", "gpu_sidebar_hidden_no_cuda_node"
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
