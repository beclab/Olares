package runner_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// harness bundles a Runner + every fake port + a table builder. It is
// the "reference environment" for skeleton Runner tests; once real
// handlers land they'll build the same harness and register their
// own transitions via `h.table.Add(...)`.
type harness struct {
	t           *testing.T
	amStore     *fake.AMStore
	workloadOps *fake.WorkloadOps
	validator   *fake.InstallValidator
	imagePuller *fake.ImagePuller
	recorder    *fake.EventRecorder
	clock       *fake.Clock
	requeuer    *runner.RecordingRequeuer

	table  *domain.TransitionTable
	engine *domain.Engine
	runner *runner.Runner
}

// newHarness wires up an empty table + all fake ports and returns the
// harness ready for per-test transitions to be registered.
func newHarness(t *testing.T) *harness {
	t.Helper()
	table := domain.NewTransitionTable()
	engine := domain.NewEngine(table)

	h := &harness{
		t:           t,
		amStore:     fake.NewAMStore(),
		workloadOps: &fake.WorkloadOps{},
		validator:   &fake.InstallValidator{},
		imagePuller: fake.NewImagePuller(),
		recorder:    &fake.EventRecorder{},
		clock:       fake.NewClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)),
		requeuer:    &runner.RecordingRequeuer{},
		table:       table,
		engine:      engine,
	}

	r, err := runner.New(runner.Config{
		Engine:      engine,
		AMStore:     h.amStore,
		WorkloadOps: h.workloadOps,
		Validator:   h.validator,
		ImagePuller: h.imagePuller,
		Recorder:    h.recorder,
		Clock:       h.clock,
		Requeuer:    h.requeuer,
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	h.runner = r
	t.Cleanup(r.Stop)
	return h
}

// seedAM installs an AM in the fake AMStore and returns its captured
// RV. Every test starts from a seeded AM so buildWorld succeeds.
func (h *harness) seedAM(name, state string) string {
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Type:         appv1.App,
		},
		Status: appv1.ApplicationManagerStatus{
			State: appv1.ApplicationManagerState(state),
		},
	}
	// Also seed the paired Application so AppExists is true by
	// default; tests that need the missing-Application branch can
	// override.
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestReconcile_NoAM pins that a missing AM is a benign no-op — the
// AM was deleted between the watch event firing and Reconcile
// running.
func TestReconcile_NoAM(t *testing.T) {
	h := newHarness(t)
	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "ghost"})
	if err != nil {
		t.Fatalf("Reconcile(missing) err = %v, want nil", err)
	}
	if out.Requeue || out.RequeueAfter != 0 {
		t.Fatalf("Reconcile(missing) outcome = %+v, want zero", out)
	}
}

// TestReconcile_NoTransition pins that a state without any registered
// Tick handler is a benign no-op (ErrNoTransition unwrapped by
// Reconcile).
func TestReconcile_NoTransition(t *testing.T) {
	h := newHarness(t)
	h.seedAM("app1", "Running")

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
	if err != nil {
		t.Fatalf("Reconcile with no matching transition err = %v, want nil", err)
	}
	if out.Requeue || out.RequeueAfter != 0 {
		t.Fatalf("outcome = %+v, want zero", out)
	}
	// No side effects should have hit the AMStore.
	if got := h.amStore.Mutations(); len(got) != 2 { // 1 AddApp + 1 AddAM
		t.Fatalf("Mutations len = %d, want 2 (seed only)", len(got))
	}
}

// TestReconcile_SynthesizedPatchAM pins the "convenience path":
// a handler returns Result{Next, Message, Reason} without an
// explicit PatchAM, and the Runner synthesises one from those fields.
func TestReconcile_SynthesizedPatchAM(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Pending, appv1.Downloading)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Downloading, Message: "starting", Reason: "Ok"}
		},
	})
	h.seedAM("app1", string(appv1.Pending))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("state = %q, want Downloading", got.Status.State)
	}
	if got.Status.Message != "starting" || got.Status.Reason != "Ok" {
		t.Fatalf("message/reason = %q/%q, want starting/Ok", got.Status.Message, got.Status.Reason)
	}
}

// TestReconcile_ExplicitPatchAM_WithOpRecord pins that AppendOpRecord
// is merged into the immediately-following PatchAM (atomic write, no
// double-patch).
func TestReconcile_ExplicitPatchAM_WithOpRecord(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Pending, appv1.Downloading)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Next: appv1.Downloading,
				Effects: []domain.Effect{
					domain.AppendOpRecord{Record: appv1.OpRecord{OpID: "op-123", Message: "started"}},
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{
						State: appv1.Downloading, Message: "downloading",
					}},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Pending))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Downloading {
		t.Fatalf("state = %q, want Downloading", got.Status.State)
	}
	if got.Status.Message != "downloading" {
		t.Fatalf("message = %q, want downloading", got.Status.Message)
	}
	if len(got.Status.OpRecords) != 1 || got.Status.OpRecords[0].OpID != "op-123" {
		t.Fatalf("OpRecords = %+v, want single op-123", got.Status.OpRecords)
	}

	// Exactly ONE PatchStatus mutation should have been recorded
	// (not one for the OpRecord + one for the PatchAM).
	patchCount := 0
	for _, m := range h.amStore.Mutations() {
		if m.Kind == "PatchStatus" {
			patchCount++
		}
	}
	if patchCount != 1 {
		t.Fatalf("PatchStatus count = %d, want 1 (opRecord merged)", patchCount)
	}
}

// TestReconcile_EmitEvent pins that an EmitEvent effect reaches the
// EventRecorder with identity fields backfilled from World.AM.
func TestReconcile_EmitEvent(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Effects: []domain.Effect{
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{State: appv1.Downloading}},
					domain.EmitEvent{
						Type:    "Normal",
						Reason:  "PullStarted",
						Message: "downloading images",
					},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Pending))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	events := h.recorder.Events()
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	ev := events[0]
	if ev.RawAppName != "app1" {
		t.Fatalf("event.RawAppName = %q, want app1", ev.RawAppName)
	}
	if ev.User != "alice" {
		t.Fatalf("event.User = %q, want alice (backfilled from AM.Spec.AppOwner)", ev.User)
	}
	if ev.Reason != "PullStarted" || ev.Message != "downloading images" {
		t.Fatalf("event Reason/Message = %q/%q, want PullStarted/downloading images", ev.Reason, ev.Message)
	}
}

// TestReconcile_EmitEvent_SkipsSameState pins that a phase update
// which stays in the current state does not publish. Image-pull
// percent ticks are published by publishImagePullProgress, not here.
func TestReconcile_EmitEvent_SkipsSameState(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Effects: []domain.Effect{
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{
						State:   appv1.Installing,
						Reason:  "HelmInstallStarted",
						Message: "installing.helm",
					}},
					domain.EmitEvent{
						Type:    "Normal",
						Reason:  "HelmInstallStarted",
						Message: "installing.helm",
					},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Installing))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if events := h.recorder.Events(); len(events) != 0 {
		t.Fatalf("events = %+v, want none for a same-state phase update", events)
	}
}

// TestReconcile_EmitEvent_UsesCatalogRawAppName pins that a state
// event carries the catalog identity, not the AM object name. The
// market card matches Spec.RawAppName; middleware AMs are named
// {namespace}-{app}.
func TestReconcile_EmitEvent_UsesCatalogRawAppName(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Effects: []domain.Effect{
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{State: appv1.Installing}},
					domain.EmitEvent{Type: "Normal", Reason: "DownloadStarted", Message: "downloading"},
				},
			}
		},
	})
	h.amStore.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "citus-middleware-citus",
			Annotations: map[string]string{
				"bytetrade.io/market-source": "market.olares",
			},
		},
		Spec: appv1.ApplicationManagerSpec{
			AppName:    "citus",
			RawAppName: "citus",
			AppOwner:   "alice",
			Type:       appv1.Middleware,
			Config:     `{"title":"Citus","icon":"citus.png"}`,
		},
		Status: appv1.ApplicationManagerStatus{State: appv1.Downloading},
	})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "citus-middleware-citus"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	events := h.recorder.Events()
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	ev := events[0]
	if ev.Name != "citus" || ev.RawAppName != "citus" {
		t.Fatalf("Name/RawAppName = %q/%q, want citus/citus", ev.Name, ev.RawAppName)
	}
	if ev.Title != "Citus" || ev.Icon != "citus.png" || ev.MarketSource != "market.olares" {
		t.Fatalf("Title/Icon/MarketSource = %q/%q/%q", ev.Title, ev.Icon, ev.MarketSource)
	}
	if ev.ChartOwner != "alice" {
		t.Fatalf("ChartOwner = %q, want alice", ev.ChartOwner)
	}
}

func TestReconcile_EmitEvent_SameStateIsDropped(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Initializing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Effects: []domain.Effect{
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{
						State:  appv1.Initializing,
						Reason: "InitializingStarted",
					}},
					domain.EmitEvent{Type: "Normal", Reason: "InitializingStarted", Message: "waiting"},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Initializing))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if events := h.recorder.Events(); len(events) != 0 {
		t.Fatalf("events = %+v, want none for a same-state update", events)
	}
}

// TestReconcile_ScheduleTick pins that ScheduleTick effects collapse
// into ReconcileOutcome.RequeueAfter (the smallest positive
// duration wins).
func TestReconcile_ScheduleTick(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Effects: []domain.Effect{
					domain.ScheduleTick{After: 30 * time.Second},
					domain.ScheduleTick{After: 5 * time.Second},
					domain.ScheduleTick{After: 60 * time.Second},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Pending))

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if out.RequeueAfter != 5*time.Second {
		t.Fatalf("RequeueAfter = %v, want 5s (smallest of the three)", out.RequeueAfter)
	}
}

// TestReconcile_CallWorkload_SuccessDelivery pins the full async
// loop: a CallWorkload spawns a goroutine, the fake workload succeeds,
// the OnSuccess Action lands in PendingActionQueue, and the requeuer
// is poked.
func TestReconcile_CallWorkload_SuccessDelivery(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadInstall,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "InstallFailed"},
				},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop() // drain the launched goroutine before assertions

	if got := h.workloadOps.Calls(); len(got) != 1 || got[0].Method != "Install" {
		t.Fatalf("workload calls = %+v, want single Install", got)
	}
	if _, ok := h.runner.Pending().Peek("app1"); !ok {
		t.Fatal("Pending queue is empty; expected WorkloadReady delivered")
	}
	pending, _ := h.runner.Pending().Peek("app1")
	if _, ok := pending.(domain.WorkloadReady); !ok {
		t.Fatalf("pending action = %T, want WorkloadReady", pending)
	}
	if got := h.requeuer.Calls(); len(got) != 1 || got[0] != "app1" {
		t.Fatalf("requeuer calls = %+v, want [app1]", got)
	}
}

// TestReconcile_CallWorkload_FailureDelivery pins the failure branch:
// a rejecting fake surfaces WorkloadFailed with Runner-populated
// Reason/Message when the handler-emitted OnFailure left them empty.
func TestReconcile_CallWorkload_FailureDelivery(t *testing.T) {
	h := newHarness(t)
	h.workloadOps.InstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HamiUnschedulable", Err: errors.New("no GPU")}
	}
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadInstall,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{}, // deliberately empty; runner backfills
				},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	pending, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("Pending queue is empty; expected WorkloadFailed delivered")
	}
	wf, ok := pending.(domain.WorkloadFailed)
	if !ok {
		t.Fatalf("pending action = %T, want WorkloadFailed", pending)
	}
	if wf.Reason != "HamiUnschedulable" {
		t.Fatalf("Reason = %q, want HamiUnschedulable (backfilled from LandedReason)", wf.Reason)
	}
	if wf.Message == "" {
		t.Fatal("Message empty; expected err.Error() backfill")
	}
}

// TestReconcile_CallbackConsumption pins the closed-loop guarantee: a
// second Reconcile pops the delivered Action, dispatches through the
// callback edge, and applies its follow-up PatchAM. This is the
// "runner keeps the state machine ticking" property.
func TestReconcile_CallbackConsumption(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Downloading, appv1.Installing)
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{Kind: domain.WorkloadInstall, OnSuccess: domain.WorkloadReady{}},
			}}
		},
	})
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.WorkloadReady{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Installing, Message: "install ok"}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))

	// Tick #1 fires the CallWorkload effect.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile #1: %v", err)
	}
	h.runner.Stop() // drain goroutine so WorkloadReady is in Pending
	// Rebuild the runner? No — Stop() has cancelled ctx but the
	// pending queue and clock survive. For a second Reconcile we
	// need a fresh Runner sharing the same ports/queue.
	r2, err := runner.New(runner.Config{
		Engine: h.engine, AMStore: h.amStore, WorkloadOps: h.workloadOps,
		Validator: h.validator, ImagePuller: h.imagePuller,
		Recorder: h.recorder, Clock: h.clock, Requeuer: h.requeuer,
		Pending: h.runner.Pending(),
	})
	if err != nil {
		t.Fatalf("runner.New (second): %v", err)
	}
	t.Cleanup(r2.Stop)

	// Tick #2 consumes the WorkloadReady action and transitions.
	if _, err := r2.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile #2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Installing {
		t.Fatalf("state after callback = %q, want Installing", got.Status.State)
	}
	if got.Status.Message != "install ok" {
		t.Fatalf("message after callback = %q, want install ok", got.Status.Message)
	}
}

// TestReconcile_UndeclaredTransition pins that a handler returning a
// Next that the table does not declare is surfaced as an error (bug
// in handler, not a benign no-op). This is the "kill drift"
// invariant the whole refactor is about.
func TestReconcile_UndeclaredTransition(t *testing.T) {
	h := newHarness(t)
	// Deliberately do NOT AddReachable Pending -> Uninstalling.
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Uninstalling}
		},
	})
	h.seedAM("app1", string(appv1.Pending))

	_, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
	if err == nil {
		t.Fatal("Reconcile err = nil, want ErrUndeclaredTransition surfaced")
	}
	var undeclared domain.ErrUndeclaredTransition
	if !errors.As(err, &undeclared) {
		t.Fatalf("err type = %T (%v), want ErrUndeclaredTransition", err, err)
	}
}

// TestReconcile_Config_MissingFields pins fast-fail construction —
// misconfiguration must not silently no-op the state machine.
func TestReconcile_Config_MissingFields(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*runner.Config)
		wantMsg string
	}{
		{"no engine", func(c *runner.Config) { c.Engine = nil }, "Engine"},
		{"no amstore", func(c *runner.Config) { c.AMStore = nil }, "AMStore"},
		{"no workloadops", func(c *runner.Config) { c.WorkloadOps = nil }, "WorkloadOps"},
		{"no validator", func(c *runner.Config) { c.Validator = nil }, "Validator"},
		{"no imagepuller", func(c *runner.Config) { c.ImagePuller = nil }, "ImagePuller"},
		{"no recorder", func(c *runner.Config) { c.Recorder = nil }, "Recorder"},
		{"no clock", func(c *runner.Config) { c.Clock = nil }, "Clock"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := runner.Config{
				Engine:      domain.NewEngine(domain.NewTransitionTable()),
				AMStore:     fake.NewAMStore(),
				WorkloadOps: &fake.WorkloadOps{},
				Validator:   &fake.InstallValidator{},
				ImagePuller: fake.NewImagePuller(),
				Recorder:    &fake.EventRecorder{},
				Clock:       fake.NewClock(time.Time{}),
			}
			tc.mutate(&cfg)
			_, err := runner.New(cfg)
			if err == nil {
				t.Fatalf("%s: err = nil, want required-field failure", tc.name)
			}
		})
	}
}

// TestReconcile_DownloadSlot_InProcessMutex pins that two Pending
// AMs whose handlers both request Downloading cannot both land
// there when Reconcile runs concurrently. FIFO is bypassed on
// purpose (placeholder always returns Next=Downloading) so the
// process-local slot lock is the only gate.
func TestReconcile_DownloadSlot_InProcessMutex(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Pending, appv1.Downloading)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Downloading}
		},
	})
	h.seedAM("app-a", string(appv1.Pending))
	h.seedAM("app-b", string(appv1.Pending))

	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for _, name := range []string{"app-a", "app-b"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			<-start
			_, err := h.runner.Reconcile(context.Background(), runner.Request{Name: name})
			errs <- err
		}(name)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
	}

	var downloading int
	for _, name := range []string{"app-a", "app-b"} {
		got, _ := h.amStore.Get(context.Background(), name)
		switch got.Status.State {
		case appv1.Downloading:
			downloading++
		case appv1.Pending:
			// lost the slot — stays Pending
		default:
			t.Fatalf("%s state = %q, want Pending or Downloading", name, got.Status.State)
		}
	}
	if downloading != 1 {
		t.Fatalf("Downloading count = %d, want 1 (in-process slot mutex)", downloading)
	}
}

// TestReconcile_DownloadSlot_ReleasedOnLeaveDownloading pins that
// the in-process holder is cleared when the occupying AM leaves
// Downloading, so the next Pending AM can acquire the slot.
func TestReconcile_DownloadSlot_ReleasedOnLeaveDownloading(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Pending, appv1.Downloading)
	h.table.AddReachable(appv1.Downloading, appv1.Installing)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Downloading}
		},
	})
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Installing}
		},
	})
	h.seedAM("holder", string(appv1.Pending))
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "holder"}); err != nil {
		t.Fatalf("holder acquire: %v", err)
	}
	holder, _ := h.amStore.Get(context.Background(), "holder")
	if holder.Status.State != appv1.Downloading {
		t.Fatalf("holder state = %q, want Downloading after acquire", holder.Status.State)
	}

	h.seedAM("waiter", string(appv1.Pending))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"}); err != nil {
		t.Fatalf("waiter while holder downloading: %v", err)
	}
	waiter, _ := h.amStore.Get(context.Background(), "waiter")
	if waiter.Status.State != appv1.Pending {
		t.Fatalf("waiter state = %q, want Pending while slot held", waiter.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "holder"}); err != nil {
		t.Fatalf("holder leave downloading: %v", err)
	}
	holder, _ = h.amStore.Get(context.Background(), "holder")
	if holder.Status.State != appv1.Installing {
		t.Fatalf("holder state = %q, want Installing", holder.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"}); err != nil {
		t.Fatalf("waiter after release: %v", err)
	}
	waiter, _ = h.amStore.Get(context.Background(), "waiter")
	if waiter.Status.State != appv1.Downloading {
		t.Fatalf("waiter state = %q, want Downloading after slot released", waiter.Status.State)
	}
}

// TestReconcile_DownloadSlot_ReleasedWhenAMDeleted pins that a
// NotFound Reconcile drops the in-process download slot. Otherwise a
// deleted Downloading AM pins the slot until process restart.
func TestReconcile_DownloadSlot_ReleasedWhenAMDeleted(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Pending, appv1.Downloading)
	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Next: appv1.Downloading}
		},
	})
	h.seedAM("holder", string(appv1.Pending))
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "holder"}); err != nil {
		t.Fatalf("holder acquire: %v", err)
	}

	h.amStore.Delete("holder")
	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "holder"})
	if err != nil {
		t.Fatalf("Reconcile(deleted) err = %v, want nil", err)
	}
	if out.Requeue || out.RequeueAfter != 0 {
		t.Fatalf("Reconcile(deleted) outcome = %+v, want zero", out)
	}

	h.seedAM("waiter", string(appv1.Pending))
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "waiter"}); err != nil {
		t.Fatalf("waiter after delete: %v", err)
	}
	waiter, _ := h.amStore.Get(context.Background(), "waiter")
	if waiter.Status.State != appv1.Downloading {
		t.Fatalf("waiter state = %q, want Downloading after deleted holder released the slot", waiter.Status.State)
	}
}
