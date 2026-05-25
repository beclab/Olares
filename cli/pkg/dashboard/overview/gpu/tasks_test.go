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

// TestRunTasks_TableModeSurfacesTransportError pins Bug 2's
// RunTasks half. Same contract as the RunList counterpart: an
// unclassifiable 4xx (here 400) must surface to cobra in table
// mode instead of being silently absorbed into an empty table.
// The bug let the user see a header row + a "-" placeholder line
// with no indication that the upstream call failed.
func TestRunTasks_TableModeSurfacesTransportError(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad filter"}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputTable

	_, runErr := captureStdoutAndErr(t, func() error { return RunTasks(context.Background(), c, cf) })
	if runErr == nil {
		t.Fatal("RunTasks returned nil; want a non-nil error surfacing Meta.Error in table mode")
	}
	if !strings.Contains(runErr.Error(), "400") && !strings.Contains(runErr.Error(), "bad filter") {
		t.Errorf("RunTasks err = %q, want it to mention HTTP 400 or 'bad filter' so the user has a diagnostic", runErr)
	}
}

// TestRunTasks_JSONModeKeepsTransportErrorOnEnvelope — RunTasks
// JSON path must keep emitting the envelope with meta.error
// populated. See the RunList equivalent for the rationale (agents
// shouldn't have to parse stderr to detect a failed iteration).
func TestRunTasks_JSONModeKeepsTransportErrorOnEnvelope(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad filter"}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunTasks(context.Background(), c, cf) })
	if !strings.Contains(out, `"error":`) {
		t.Errorf("expected meta.error in envelope; got:\n%s", out)
	}
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.tasks"`) {
		t.Errorf("expected envelope kind to still emit alongside meta.error; got:\n%s", out)
	}
}

// TestRunTaskByRef_ResolvesPodUIDFromList — happy path for the
// `gpu tasks <ref>` shorthand. With a single-name match in the
// task list the resolver MUST forward to RunTaskDetail with the
// auto-resolved pod-uid + sharemode, producing a
// `dashboard.overview.gpu.task.detail.full` envelope. The user
// should never have to copy-paste pod-uid from kubectl.
func TestRunTaskByRef_ResolvesPodUIDFromList(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"comfyai","status":"running","appName":"comfyai-7d4f4915-7d-zwmgs","namespace":"ns-1","podUid":"pod-uid-A","nodeName":"node-1","deviceShareModes":["2"],"devicesCoreUtilizedPercent":[0],"devicesMemUtilized":[256]}
            ]}`))
		},
		taskGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"name":"comfyai","status":"running","appName":"comfyai-7d4f4915-7d-zwmgs","namespace":"ns-1","podUid":"pod-uid-A","deviceShareModes":["2"]}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
		rangeVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunTaskByRef(context.Background(), c, cf, "comfyai") })
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.task.detail.full"`) {
		t.Errorf("expected full task-detail envelope; got:\n%s", out)
	}
	if !strings.Contains(out, `"comfyai"`) {
		t.Errorf("expected resolved task name in detail body; got:\n%s", out)
	}
}

// TestRunTaskByRef_ResolvesByPodUID — pinned regression for the
// "user copy-pasted POD_UID column instead of TASK column" path
// (the bug that prompted RunTaskByRef in the first place). The
// resolver MUST treat the arg as a pod-uid and produce the same
// detail envelope the name-arg path does. Without this the
// listing-then-drill-down workflow breaks for any user whose eye
// lands on the rightmost column first.
func TestRunTaskByRef_ResolvesByPodUID(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"comfyai","status":"running","appName":"comfyai-7d4f4915-7d-zwmgs","namespace":"ns-1","podUid":"d2a8ea32-8e56-49f9-b876-21b2fa0c5a83","nodeName":"node-1","deviceShareModes":["2"]}
            ]}`))
		},
		taskGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"name":"comfyai","status":"running","appName":"comfyai-7d4f4915-7d-zwmgs","namespace":"ns-1","podUid":"d2a8ea32-8e56-49f9-b876-21b2fa0c5a83","deviceShareModes":["2"]}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
		rangeVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error {
		return RunTaskByRef(context.Background(), c, cf, "d2a8ea32-8e56-49f9-b876-21b2fa0c5a83")
	})
	if !strings.Contains(out, `"kind":"dashboard.overview.gpu.task.detail.full"`) {
		t.Errorf("pod-uid ref must resolve to the same detail envelope as a name ref; got:\n%s", out)
	}
	if !strings.Contains(out, `"comfyai"`) {
		t.Errorf("expected pod-uid lookup to recover the row's name (comfyai); got:\n%s", out)
	}
}

// TestRunTaskByRef_NotFoundEmitsNoGPUDetected — when the ref
// matches neither a name nor a pod-uid the resolver MUST emit a
// standard `no_gpu_detected` envelope (not error out) and, in
// table mode, suggest the actual candidates so the user can
// recover without re-running the listing command.
func TestRunTaskByRef_NotFoundEmitsNoGPUDetected(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	out := captureStdout(t, func() error { return RunTaskByRef(context.Background(), c, cf, "ghost") })
	if !strings.Contains(out, `"empty_reason":"no_gpu_detected"`) {
		t.Errorf("expected empty_reason=no_gpu_detected; got:\n%s", out)
	}
}

// TestRunTaskByRef_NotFoundTableShowsHint — table-mode counterpart
// of the above: when the ref doesn't match anything but the list
// has rows, the user should see a "(no task matches …)" line plus
// a hint listing real (name, pod-uid) pairs they can copy.
func TestRunTaskByRef_NotFoundTableShowsHint(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"comfyai","podUid":"pod-A","deviceShareModes":["2"]},
              {"name":"trainer","podUid":"pod-B","deviceShareModes":["0"]}
            ]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputTable

	out := captureStdout(t, func() error { return RunTaskByRef(context.Background(), c, cf, "ghost") })
	for _, want := range []string{
		`(no task matches "ghost"`,
		"hint: try one of:",
		"comfyai (pod-A)",
		"trainer (pod-B)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table-mode not-found output missing %q in:\n%s", want, out)
		}
	}
}

// TestRunTaskByRef_AmbiguousErrorsWithCandidates — two pods sharing
// the same task name MUST surface a typed error listing the
// candidate pod-uids. Pod-uids are themselves valid refs so the
// hint says `gpu tasks <pod-uid>` (the disambiguating re-run) —
// not the legacy `gpu task-detail <name> <pod-uid>` two-arg form.
func TestRunTaskByRef_AmbiguousErrorsWithCandidates(t *testing.T) {
	srv := gpuStubMux{
		taskList: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"items":[
              {"name":"trainer","podUid":"pod-A","deviceShareModes":["0"]},
              {"name":"trainer","podUid":"pod-B","deviceShareModes":["0"]}
            ]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	err := RunTaskByRef(context.Background(), c, cf, "trainer")
	if err == nil {
		t.Fatal("expected ambiguity error; got nil")
	}
	msg := err.Error()
	for _, want := range []string{`"trainer"`, "pod-A", "pod-B", "gpu tasks <pod-uid>"} {
		if !strings.Contains(msg, want) {
			t.Errorf("ambiguity error %q missing %q", msg, want)
		}
	}
}
