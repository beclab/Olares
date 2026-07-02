package apiserver

import (
	"reflect"
	"sync"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// FakeDispatcher — direct behavioural tests. Every HTTP handler
// unit test relies on Calls() returning what Trigger observed, so
// exercise the primitive directly.
// -----------------------------------------------------------------

func TestFakeDispatcher_RecordsCallsInOrder(t *testing.T) {
	var d FakeDispatcher

	d.Trigger("a", domain.TriggerOp{Op: appv1.InstallOp})
	d.Trigger("b", domain.TriggerOp{Op: appv1.CancelOp})
	d.Trigger("a", domain.Tick{})

	want := []DispatchCall{
		{Name: "a", Action: domain.TriggerOp{Op: appv1.InstallOp}},
		{Name: "b", Action: domain.TriggerOp{Op: appv1.CancelOp}},
		{Name: "a", Action: domain.Tick{}},
	}
	if got := d.Calls(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Calls mismatch:\n got %#v\nwant %#v", got, want)
	}
}

func TestFakeDispatcher_CallsReturnsCopy(t *testing.T) {
	var d FakeDispatcher
	d.Trigger("a", domain.TriggerOp{Op: appv1.InstallOp})

	snap := d.Calls()
	snap[0].Name = "MUTATED"

	if got := d.Calls()[0].Name; got != "a" {
		t.Fatalf("Calls() must return a copy; got name=%q, want %q", got, "a")
	}
}

func TestFakeDispatcher_Reset(t *testing.T) {
	var d FakeDispatcher
	d.Trigger("a", domain.TriggerOp{Op: appv1.InstallOp})
	d.Trigger("b", domain.TriggerOp{Op: appv1.CancelOp})

	d.Reset()

	if got := d.Calls(); len(got) != 0 {
		t.Fatalf("Reset must clear calls, got %d", len(got))
	}

	d.Trigger("c", domain.TriggerOp{Op: appv1.UninstallOp})
	if got := d.Calls(); len(got) != 1 || got[0].Name != "c" {
		t.Fatalf("post-Reset Trigger must record fresh, got %#v", got)
	}
}

func TestFakeDispatcher_ConcurrentTrigger(t *testing.T) {
	var d FakeDispatcher
	const total = 200
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			d.Trigger("app", domain.TriggerOp{Op: appv1.InstallOp})
		}()
	}
	wg.Wait()
	if got := len(d.Calls()); got != total {
		t.Fatalf("concurrent Trigger dropped calls: got %d, want %d", got, total)
	}
}

// -----------------------------------------------------------------
// RunnerDispatcher — wire-up tests. RecordingRequeuer records the
// wake-up without running Reconcile.
// -----------------------------------------------------------------

func TestRunnerDispatcher_TriggerRequeues(t *testing.T) {
	req := &runner.RecordingRequeuer{}
	d := &RunnerDispatcher{requeuer: req}

	d.Trigger("myapp", domain.TriggerOp{Op: appv1.InstallOp})

	if got, want := req.Calls(), []string{"myapp"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requeuer calls mismatch:\n got %#v\nwant %#v", got, want)
	}
}

func TestNewRunnerDispatcher_PanicsOnNilRunner(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil runner")
		}
	}()
	_ = NewRunnerDispatcher(nil, runner.NoopRequeuer{})
}

func TestNewRunnerDispatcher_PanicsOnNilRequeuer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil requeuer")
		}
	}()
	// We can't easily build a real *runner.Runner here (it needs
	// full ports), so we panic-check the requeuer branch by using
	// a non-nil sentinel — the runner nil check runs first, so
	// substituting any non-nil *runner.Runner would work. Since we
	// specifically want to hit the second nil check, construct a
	// minimally-typed non-nil Runner via a zero-value pointer.
	r := &runner.Runner{}
	_ = NewRunnerDispatcher(r, nil)
}

// -----------------------------------------------------------------
// Compile-time contract: FakeDispatcher and RunnerDispatcher both
// satisfy OpDispatcher. If either drifts (e.g. Trigger renamed),
// the whole apiserver-facing consumer set stops compiling.
// -----------------------------------------------------------------

var (
	_ OpDispatcher = (*FakeDispatcher)(nil)
	_ OpDispatcher = (*RunnerDispatcher)(nil)
)
