package ctrladapter

import (
	"context"
	"reflect"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/ctrlrequeue"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// stubManager satisfies ctrl.Manager at compile-time by embedding
// the interface (all methods delegate to a nil interface value and
// would panic if called). The Setup nil-arg tests below rely on
// short-circuiting BEFORE any manager method is invoked, so a
// method call from stubManager would be a test bug — not a
// production concern. Kept package-local so no other test suite
// accidentally reuses it.
type stubManager struct{ ctrl.Manager }

// newTestRunner builds a Runner wired to every fake port. Mirrors
// the newHarness pattern in runner_test.go but is intentionally
// duplicated (rather than imported) — runner.newHarness is
// package-private, and cross-package harness reuse would couple
// two test suites that should evolve independently.
func newTestRunner(t *testing.T) (*runner.Runner, *fake.AMStore, *domain.TransitionTable) {
	t.Helper()
	table := domain.NewTransitionTable()
	engine := domain.NewEngine(table)
	amStore := fake.NewAMStore()
	r, err := runner.New(runner.Config{
		Engine:      engine,
		AMStore:     amStore,
		WorkloadOps: &fake.WorkloadOps{},
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       fake.NewClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	t.Cleanup(r.Stop)
	return r, amStore, table
}

// seedAM installs an AM in state `state` and its paired Application.
func seedAM(t *testing.T, amStore *fake.AMStore, name, state string) {
	t.Helper()
	amStore.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName: name, AppOwner: "alice",
			AppNamespace: name + "-alice",
			Type:         appv1.App,
		},
		Status: appv1.ApplicationManagerStatus{
			State: appv1.ApplicationManagerState(state),
		},
	})
	amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

// -----------------------------------------------------------------
// outcomeToResult — pure mapping. Table-driven.
// -----------------------------------------------------------------

func TestOutcomeToResult(t *testing.T) {
	cases := []struct {
		name string
		in   runner.ReconcileOutcome
		want reconcile.Result
	}{
		{
			name: "zero → zero",
			in:   runner.ReconcileOutcome{},
			want: reconcile.Result{},
		},
		{
			name: "RequeueAfter propagates",
			in:   runner.ReconcileOutcome{RequeueAfter: 5 * time.Second},
			want: reconcile.Result{RequeueAfter: 5 * time.Second},
		},
		{
			name: "Requeue=true propagates (conflict path)",
			in:   runner.ReconcileOutcome{Requeue: true},
			want: reconcile.Result{Requeue: true},
		},
		{
			name: "both fields propagate",
			in:   runner.ReconcileOutcome{Requeue: true, RequeueAfter: 30 * time.Second},
			want: reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := outcomeToResult(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("outcomeToResult(%+v) = %+v, want %+v", c.in, got, c.want)
			}
		})
	}
}

// -----------------------------------------------------------------
// Reconciler — behaviour end-to-end against a real Runner + fakes.
// -----------------------------------------------------------------

// TestReconciler_NilRunner — the sanity check main.go depends on:
// a mis-wired Reconciler returns an error rather than panicking,
// so controller-runtime's backoff surfaces the misconfiguration in
// logs instead of the process dying.
func TestReconciler_NilRunner(t *testing.T) {
	rec := &Reconciler{}
	_, err := rec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "any"},
	})
	if err == nil {
		t.Fatal("Reconcile(nil Runner) err = nil, want non-nil")
	}
}

// TestReconciler_MissingAM — the "AM was deleted between the
// watch event and Reconcile" case must be a benign no-op (delegated
// to Runner.Reconcile which explicitly returns zero outcome on
// NotFound). Pins that the adapter doesn't paper over the delegate
// with a stray Requeue.
func TestReconciler_MissingAM(t *testing.T) {
	r, _, _ := newTestRunner(t)
	rec := &Reconciler{Runner: r}

	got, err := rec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "ghost"},
	})
	if err != nil {
		t.Fatalf("Reconcile(missing AM) err = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, reconcile.Result{}) {
		t.Errorf("Reconcile(missing AM) result = %+v, want zero", got)
	}
}

// TestReconciler_NoTransition — an AM in a state with no registered
// Tick handler yields ErrNoTransition inside the Runner, which
// Reconcile classifies as benign (zero Result, nil error). Pins
// that the adapter surfaces that classification unchanged.
func TestReconciler_NoTransition(t *testing.T) {
	r, amStore, _ := newTestRunner(t)
	seedAM(t, amStore, "app1", "Running")
	rec := &Reconciler{Runner: r}

	got, err := rec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "app1"},
	})
	if err != nil {
		t.Fatalf("Reconcile(no-transition) err = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, reconcile.Result{}) {
		t.Errorf("result = %+v, want zero", got)
	}
}

// TestReconciler_MapsRequeueAfter — the load-bearing end-to-end.
// A handler emits ScheduleTick(500ms); the Runner surfaces that as
// ReconcileOutcome.RequeueAfter=500ms; the Reconciler must map it
// to reconcile.Result.RequeueAfter with the identical value.
func TestReconciler_MapsRequeueAfter(t *testing.T) {
	r, amStore, table := newTestRunner(t)
	table.AddReachable(appv1.Pending, appv1.Pending) // self-loop OK
	table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Next: appv1.Pending,
				Effects: []domain.Effect{
					domain.ScheduleTick{After: 500 * time.Millisecond},
				},
			}
		},
	})
	seedAM(t, amStore, "app1", string(appv1.Pending))
	rec := &Reconciler{Runner: r}

	got, err := rec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "app1"},
	})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if got.RequeueAfter != 500*time.Millisecond {
		t.Errorf("Result.RequeueAfter = %v, want 500ms", got.RequeueAfter)
	}
	if got.Requeue {
		t.Errorf("Result.Requeue = true; ScheduleTick should not set Requeue")
	}
}

// TestReconciler_PropagatesRunnerError — a Runner error that is
// NOT a benign classification must reach the caller so controller-
// runtime can rate-limit + log it. We synthesize this by seeding
// NO AM and NO Application, then invoking Reconcile with a name
// that also doesn't exist — which is actually the NotFound path
// (benign). To hit a real error we need to break the AMStore itself;
// we skip that scenario here rather than adding a test-only
// error-injection port (would leak test concerns into ports/).
// The outcomeToResult table above already covers the mapping side
// of the error semantics (nil error → zero Result).
//
// Left as a comment rather than an empty test so a future reader
// knows the shape was considered.

// -----------------------------------------------------------------
// Setup — argument validation only. End-to-end wiring is exercised
// in envtest / integration; unit tests here just prove the nil
// arguments fail loudly instead of panicking or silently no-oping.
// -----------------------------------------------------------------

func TestSetup_RejectsNilArgs(t *testing.T) {
	// We do not spin up a real manager; nil-arg branches short-
	// circuit before any manager method is called.
	r, _, _ := newTestRunner(t)
	req := ctrlrequeue.New()

	cases := []struct {
		name string
		fn   func() error
	}{
		{"nil manager", func() error { return Setup(nil, r, req) }},
		{"nil runner", func() error { return Setup(stubManager{}, nil, req) }},
		{"nil requeuer", func() error { return Setup(stubManager{}, r, nil) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err == nil {
				t.Errorf("Setup(%s) err = nil, want non-nil", c.name)
			}
		})
	}
}
