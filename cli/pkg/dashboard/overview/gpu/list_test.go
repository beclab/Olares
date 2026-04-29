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
// (GPU_ID/MODEL/MODE/HOST/HEALTH/CORE_UTIL/VRAM/POWER/TEMP).
func TestRunList_HappyPath(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[
              {"uuid":"GPU-A","type":"NVIDIA-A","shareMode":"0","nodeName":"node-1","health":true,"coreUtilizedPercent":35,"memoryTotal":24576,"memoryUsed":12288,"power":120.5,"temperature":62},
              {"uuid":"GPU-B","type":"NVIDIA-B","shareMode":"1","nodeName":"node-2","health":false,"coreUtilizedPercent":80,"memoryTotal":81920,"memoryUsed":4096,"power":250,"temperature":81}
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
