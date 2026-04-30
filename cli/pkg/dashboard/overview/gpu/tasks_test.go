package gpu

import (
	"context"
	"net/http"
	"strings"
	"testing"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestRunTasks_HappyPath pins per-task display fields. Note that
// `deviceShareModes` and util/mem arrays must be reduced via
// firstAnyInArray (the SPA's task list shows only index 0 of the
// allocated-device fan-out).
func TestRunTasks_HappyPath(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"task-A","status":"running","appName":"pod-1","namespace":"ns-1","podUid":"uid-1","nodeName":"node-1","deviceShareModes":["0"],"devicesCoreUtilizedPercent":[42],"devicesMemUtilized":[1024]}
            ]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunTasks(context.Background(), c, cf) })
	if !strings.Contains(out, `"task-A"`) {
		t.Errorf("missing task-A in:\n%s", out)
	}
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.tasks"`) {
		t.Errorf("missing tasks kind in:\n%s", out)
	}
}

// TestRunTasks_NoVgpuIntegration404 mirrors RunList's 404 branch.
func TestRunTasks_NoVgpuIntegration404(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunTasks(context.Background(), c, cf) })
	if !strings.Contains(out, `"empty_reason":"no_vgpu_integration"`) {
		t.Errorf("missing no_vgpu_integration; got:\n%s", out)
	}
}
