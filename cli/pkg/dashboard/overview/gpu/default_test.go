package gpu

import (
	"context"
	"net/http"
	"strings"
	"testing"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestRunDefault_HappyPath: bare `dashboard overview gpu` emits a
// `dashboard.overview.gpu` parent envelope whose `sections` map
// has both `graphics` (kind .list) and `tasks` (kind .tasks)
// populated. Mirrors the SPA's GPU overview page that loads both
// tabs eagerly. Item ordering inside each section is governed by
// the leaf builders (already covered by RunList/RunTasks tests).
func TestRunDefault_HappyPath(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"list":[
              {"uuid":"GPU-A","type":"NVIDIA-A","shareMode":"0","nodeName":"node-1","health":true,"coreUtilizedPercent":35,"memoryTotal":24576,"memoryUsed":12288,"memoryUtilizedPercent":50,"power":120.5,"temperature":62}
            ]}`))
		},
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

	out := captureStdout(t, func() error { return RunDefault(context.Background(), c, cf) })
	for _, want := range []string{
		`"kind":"dashboard.overview.gpu"`,
		`"graphics":`,
		`"tasks":`,
		`"kind":"dashboard.overview.gpu.list"`,
		`"kind":"dashboard.overview.gpu.tasks"`,
		`"GPU-A"`,
		`"task-A"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// TestRunDefault_PartialFailureKeepsOtherSection: a 5xx on
// /v1/gpus must NOT abort the tasks section. The graphics section
// surfaces empty_reason=vgpu_unavailable + http_status=500, while
// tasks still emits its items. This is the canonical
// sections-envelope partial-failure invariant — same shape as
// `dashboard overview disk` (one disk's partition fetch failing
// must not blank the other disks' tables).
func TestRunDefault_PartialFailureKeepsOtherSection(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"hami down"}`))
		},
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"task-A","status":"running","appName":"pod-1","namespace":"ns-1","podUid":"uid-1","deviceShareModes":["0"]}
            ]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunDefault(context.Background(), c, cf) })
	if !strings.Contains(out, `"empty_reason":"vgpu_unavailable"`) {
		t.Errorf("expected graphics section vgpu_unavailable; got:\n%s", out)
	}
	if !strings.Contains(out, `"task-A"`) {
		t.Errorf("expected tasks section to still render task-A despite graphics 500; got:\n%s", out)
	}
}

// TestRunDefault_HAMINotInstalled404: 404 on both endpoints folds
// into per-section empty_reason=no_vgpu_integration. The parent
// envelope itself stays non-empty (it always carries fetched_at +
// the two section keys); agents distinguish the case via
// `sections.<key>.meta.empty_reason`. Pinned because earlier
// drafts collapsed both 404s into a parent-level empty.
func TestRunDefault_HAMINotInstalled404(t *testing.T) {
	srv := gpuStubMux{
		graphicsList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunDefault(context.Background(), c, cf) })
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu"`) {
		t.Errorf("missing parent kind in:\n%s", out)
	}
	// Both sections should carry no_vgpu_integration.
	if c := strings.Count(out, `"empty_reason":"no_vgpu_integration"`); c != 2 {
		t.Errorf("expected 2 no_vgpu_integration markers (graphics+tasks); got %d in:\n%s", c, out)
	}
}
