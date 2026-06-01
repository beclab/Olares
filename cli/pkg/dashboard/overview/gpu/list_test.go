package gpu

import (
	"context"
	"net/http"
	"strings"
	"testing"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestRunList_HappyPath pins the table-rendered GPU list shape
// against an admin profile + CUDA node + 2-device HAMI fixture.
// The Display columns must match the SPA's column order
// (GPU_ID/MODEL/MODE/HOST/HEALTH/CORE_UTIL/VRAM/VRAM_USAGE/POWER/TEMP).
func TestRunList_HappyPath(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[
              {"uuid":"GPU-A","type":"NVIDIA-A","shareMode":"0","nodeName":"node-1","health":true,"coreUtilizedPercent":35,"memoryTotal":24576,"memoryUsed":12288,"memoryUtilizedPercent":50,"power":120.5,"temperature":62},
              {"uuid":"GPU-B","type":"NVIDIA-B","shareMode":"1","nodeName":"node-2","health":false,"coreUtilizedPercent":80,"memoryTotal":81920,"memoryUsed":4096,"memoryUtilizedPercent":5,"power":250,"temperature":81}
            ]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunList(context.Background(), c, cf) })
	if !strings.Contains(out, `"GPU-A"`) || !strings.Contains(out, `"GPU-B"`) {
		t.Errorf("expected both GPU IDs in JSON; got:\n%s", out)
	}
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.list"`) {
		t.Errorf("missing kind=dashboard.overview.gpu.list in:\n%s", out)
	}
	// SPA "VRAM usage rate" column — the regression net for the
	// GPUsTable.vue:159-166 column that earlier CLI revisions did
	// not surface.
	if !strings.Contains(out, `"vram_usage":"50%"`) || !strings.Contains(out, `"vram_usage":"5%"`) {
		t.Errorf("expected vram_usage column for both GPUs (50%% / 5%%); got:\n%s", out)
	}
}

// TestRunList_NoVgpuIntegration404 pins the "HAMI not installed"
// branch: a 404 from /v1/gpus must surface as Empty=true +
// EmptyReason="no_vgpu_integration", not as a transport error.
func TestRunList_NoVgpuIntegration404(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunList(context.Background(), c, cf) })
	if !strings.Contains(out, `"empty_reason":"no_vgpu_integration"`) {
		t.Errorf("missing no_vgpu_integration; got:\n%s", out)
	}
}

// TestRunList_TableModeSurfacesTransportError pins Bug 2: when
// BuildListEnvelope hits a 4xx HAMI response that's neither a 404
// (no_vgpu_integration) nor a 5xx (vgpu_unavailable) — e.g. a 400
// from a misconfigured deployment, an auth 401/403, or a generic
// transport breakage — the envelope captures the error on
// `Meta.Error` for JSON consumers, but table mode used to silently
// render an empty table with no diagnostic.
//
// Contract: RunList in table mode MUST return that error to cobra
// so it surfaces on stderr instead of being swallowed. JSON mode
// keeps emitting the envelope unchanged (consumers parse
// meta.error directly). The chosen status here (400) is the
// canonical "unclassifiable" case: a 4xx that doesn't match any
// of the soft-empty branches.
func TestRunList_TableModeSurfacesTransportError(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad filter"}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	// Explicit table mode — fixtureFlags already defaults here, but
	// the assertion is about TABLE not JSON, so we pin it locally so
	// a future fixture flip doesn't silently mute this test.
	cf.Output = pkgdashboard.OutputTable

	_, runErr := captureStdoutAndErr(t, func() error { return RunList(context.Background(), c, cf) })
	if runErr == nil {
		t.Fatal("RunList returned nil; want a non-nil error surfacing Meta.Error in table mode")
	}
	if !strings.Contains(runErr.Error(), "400") && !strings.Contains(runErr.Error(), "bad filter") {
		t.Errorf("RunList err = %q, want it to mention either the HTTP 400 status or the HAMI message 'bad filter' so the user has a diagnostic", runErr)
	}
}

// TestRunList_JSONModeKeepsTransportErrorOnEnvelope pins the
// other half of Bug 2's contract: JSON mode MUST NOT regress to
// "return err and emit nothing". Agents drive --output json and
// rely on `meta.error` being present on a non-zero-length envelope
// — converting that to a cobra-side error would skip the envelope
// emit entirely and force agents to parse stderr.
func TestRunList_JSONModeKeepsTransportErrorOnEnvelope(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad filter"}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunList(context.Background(), c, cf) })
	if !strings.Contains(out, `"error":`) {
		t.Errorf("expected meta.error in envelope; got:\n%s", out)
	}
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.list"`) {
		t.Errorf("expected envelope kind to still emit alongside meta.error; got:\n%s", out)
	}
}

// TestRunList_NoGPUDetected pins the "HAMI up, zero devices"
// branch.
func TestRunList_NoGPUDetected(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunList(context.Background(), c, cf) })
	if !strings.Contains(out, `"empty_reason":"no_gpu_detected"`) {
		t.Errorf("missing no_gpu_detected; got:\n%s", out)
	}
}
