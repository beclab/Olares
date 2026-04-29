package gpu

import (
	"context"
	"net/http"
	"strings"
	"testing"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestRunGet_HappyPath: a single flat detail Item, no fan-out. The
// envelope kind must be the per-GPU detail kind so JSON consumers
// can demux from a list-of-details fan-in.
func TestRunGet_HappyPath(t *testing.T) {
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"uuid":"GPU-A","type":"NVIDIA-A","health":true,"shareMode":"0","nodeName":"node-1"}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunGet(context.Background(), c, cf, "GPU-A") })
	if !strings.Contains(out, `"GPU-A"`) {
		t.Errorf("missing GPU-A in:\n%s", out)
	}
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.detail"`) {
		t.Errorf("missing detail kind in:\n%s", out)
	}
}

// TestRunGet_NotFound404 pins the 404 branch (UUID typo or HAMI
// not installed) — Empty=true / EmptyReason="no_vgpu_integration".
func TestRunGet_NotFound404(t *testing.T) {
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunGet(context.Background(), c, cf, "GPU-Bogus") })
	if !strings.Contains(out, `"empty_reason":"no_vgpu_integration"`) {
		t.Errorf("missing no_vgpu_integration; got:\n%s", out)
	}
}
