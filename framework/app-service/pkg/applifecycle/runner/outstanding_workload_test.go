package runner

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
)

func newOutstandingTestRunner(t *testing.T, table *domain.TransitionTable, ops *fake.WorkloadOps) (*Runner, *fake.AMStore) {
	t.Helper()
	store := fake.NewAMStore()
	r, err := New(Config{
		Engine:      domain.NewEngine(table),
		AMStore:     store,
		WorkloadOps: ops,
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       fake.NewClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(r.Stop)
	return r, store
}

func outstandingTestAM(name string, state appv1.ApplicationManagerState) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Type:         appv1.App,
		},
		Status: appv1.ApplicationManagerStatus{State: state},
	}
}

func seedOutstandingTestAM(store *fake.AMStore, am *appv1.ApplicationManager) {
	store.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: am.Name}})
	store.Add(am)
}

func installCall() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadInstall,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{},
	}
}

func waitForTest(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestLaunchCallWorkload_SameAMCannotStartTwiceWhileRunning(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	defer close(release)
	ops := &fake.WorkloadOps{
		InstallFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-release
			return ports.WorkloadResult{}
		},
	}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)
	world := domain.World{AM: outstandingTestAM("app1", appv1.Installing)}

	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("first launch: %v", err)
	}
	<-started
	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("duplicate launch: %v", err)
	}
	if got := len(ops.Calls()); got != 1 {
		t.Fatalf("Install calls while first is running = %d, want 1", got)
	}
	if got := r.InFlight(); got != 1 {
		t.Fatalf("InFlight = %d, want actual goroutine count 1", got)
	}
}

func TestLaunchCallWorkload_SameAMRemainsClaimedWhileCallbackPending(t *testing.T) {
	ops := &fake.WorkloadOps{}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)
	world := domain.World{AM: outstandingTestAM("app1", appv1.Installing)}

	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("first launch: %v", err)
	}
	waitForTest(t, "first callback", func() bool {
		_, queued := r.Pending().Peek("app1")
		return queued && r.InFlight() == 0
	})
	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("duplicate launch: %v", err)
	}
	if got := len(ops.Calls()); got != 1 {
		t.Fatalf("Install calls with callback pending = %d, want 1", got)
	}
}

func TestReconcile_ConsumingCallbackAllowsNextPhaseWorkload(t *testing.T) {
	table := domain.NewTransitionTable()
	var handlerSawInFlight bool
	table.Add(domain.Transition{
		From: appv1.Installing,
		On:   domain.WorkloadReady{},
		Handler: func(_ context.Context, world domain.World) domain.Result {
			handlerSawInFlight = world.WorkloadInFlight
			return domain.Result{Effects: []domain.Effect{installCall()}}
		},
	})
	ops := &fake.WorkloadOps{}
	r, store := newOutstandingTestRunner(t, table, ops)
	am := outstandingTestAM("app1", appv1.Installing)
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatalf("first launch: %v", err)
	}
	waitForTest(t, "first callback", func() bool {
		_, queued := r.Pending().Peek("app1")
		return queued
	})
	if _, err := r.Reconcile(context.Background(), Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile callback: %v", err)
	}
	if handlerSawInFlight {
		t.Fatal("callback handler WorkloadInFlight = true, want false after acknowledge")
	}
	waitForTest(t, "next phase workload", func() bool { return len(ops.Calls()) == 2 })
}

func TestReconcile_UserActionCannotClearRunningWorkloadClaim(t *testing.T) {
	table := domain.NewTransitionTable()
	var handlerSawInFlight bool
	table.Add(domain.Transition{
		From: appv1.Running,
		On:   domain.TriggerOp{Op: appv1.StopOp},
		Handler: func(_ context.Context, world domain.World) domain.Result {
			handlerSawInFlight = world.WorkloadInFlight
			return domain.Result{Effects: []domain.Effect{installCall()}}
		},
	})
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	defer close(release)
	ops := &fake.WorkloadOps{
		InstallFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-release
			return ports.WorkloadResult{}
		},
	}
	r, store := newOutstandingTestRunner(t, table, ops)
	am := outstandingTestAM("app1", appv1.Running)
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatalf("first launch: %v", err)
	}
	<-started
	r.Pending().Deliver("app1", domain.TriggerOp{Op: appv1.StopOp})
	if _, err := r.Reconcile(context.Background(), Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile user action: %v", err)
	}
	if got := len(ops.Calls()); got != 1 {
		t.Fatalf("Install calls after user action = %d, want 1", got)
	}
	if !handlerSawInFlight {
		t.Fatal("user-action handler WorkloadInFlight = false, want true while workload runs")
	}
	if got := r.InFlight(); got != 1 {
		t.Fatalf("InFlight after user action = %d, want actual goroutine count 1", got)
	}
}

func TestReconcile_ConflictRestoresCompletedCallbackClaim(t *testing.T) {
	table := domain.NewTransitionTable()
	handlerCalls := 0
	table.Add(domain.Transition{
		From: appv1.Installing,
		On:   domain.WorkloadReady{},
		Handler: func(context.Context, domain.World) domain.Result {
			handlerCalls++
			if handlerCalls > 1 {
				return domain.Result{}
			}
			return domain.Result{Effects: []domain.Effect{
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Installing, Message: "first",
				}},
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Running, Message: "second",
				}},
			}}
		},
	})
	r, store := newOutstandingTestRunner(t, table, &fake.WorkloadOps{})
	am := outstandingTestAM("app1", appv1.Installing)
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatalf("launch: %v", err)
	}
	waitForTest(t, "completed callback", func() bool {
		_, queued := r.Pending().Peek("app1")
		return queued && r.InFlight() == 0
	})

	out, err := r.Reconcile(context.Background(), Request{Name: "app1"})
	if err != nil {
		t.Fatalf("conflicting Reconcile: %v", err)
	}
	if !out.Requeue {
		t.Fatalf("conflicting outcome = %+v, want Requeue=true", out)
	}
	if action, queued := r.Pending().Peek("app1"); !queued {
		t.Fatal("callback was not restored after conflict")
	} else if _, ok := action.(domain.WorkloadReady); !ok {
		t.Fatalf("restored action = %T, want WorkloadReady", action)
	}
	restoredWorld, ok, err := r.buildWorld(context.Background(), "app1")
	if err != nil || !ok {
		t.Fatalf("buildWorld after restore = (_, %t, %v), want ok", ok, err)
	}
	if !restoredWorld.WorkloadInFlight {
		t.Fatal("WorkloadInFlight = false after callback restore, want true")
	}

	if _, err := r.Reconcile(context.Background(), Request{Name: "app1"}); err != nil {
		t.Fatalf("retry Reconcile: %v", err)
	}
	if handlerCalls != 2 {
		t.Fatalf("handler calls = %d, want callback retried", handlerCalls)
	}
	if _, queued := r.Pending().Peek("app1"); queued {
		t.Fatal("callback remains queued after successful retry")
	}
	retriedWorld, ok, err := r.buildWorld(context.Background(), "app1")
	if err != nil || !ok {
		t.Fatalf("buildWorld after retry = (_, %t, %v), want ok", ok, err)
	}
	if retriedWorld.WorkloadInFlight {
		t.Fatal("WorkloadInFlight = true after successful callback retry")
	}
}

func TestLaunchCallWorkload_DifferentAMsRunConcurrently(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	defer close(release)
	ops := &fake.WorkloadOps{
		InstallFunc: func(_ context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
			started <- req.AM.Name
			<-release
			return ports.WorkloadResult{}
		},
	}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)

	for _, name := range []string{"app-a", "app-b"} {
		if err := r.launchCallWorkload(context.Background(), domain.World{
			AM: outstandingTestAM(name, appv1.Installing),
		}, installCall()); err != nil {
			t.Fatalf("launch %s: %v", name, err)
		}
	}
	seen := map[string]bool{}
	for len(seen) < 2 {
		select {
		case name := <-started:
			seen[name] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("started AMs = %v, want both", seen)
		}
	}
	if got := r.InFlight(); got != 2 {
		t.Fatalf("InFlight = %d, want 2 actual goroutines", got)
	}
}

func TestBuildWorld_ReportsOutstandingWorkAndFreshRunnerReportsFalse(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	ops := &fake.WorkloadOps{
		InstallFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			close(started)
			<-release
			return ports.WorkloadResult{}
		},
	}
	table := domain.NewTransitionTable()
	r, store := newOutstandingTestRunner(t, table, ops)
	am := outstandingTestAM("app1", appv1.Installing)
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatalf("launch: %v", err)
	}
	<-started
	runningWorld, ok, err := r.buildWorld(context.Background(), "app1")
	if err != nil || !ok {
		t.Fatalf("buildWorld running = (_, %t, %v), want ok", ok, err)
	}
	if !runningWorld.WorkloadInFlight {
		t.Fatal("WorkloadInFlight = false while workload is running")
	}

	releaseOnce.Do(func() { close(release) })
	waitForTest(t, "callback pending", func() bool {
		_, queued := r.Pending().Peek("app1")
		return queued && r.InFlight() == 0
	})
	pendingWorld, ok, err := r.buildWorld(context.Background(), "app1")
	if err != nil || !ok {
		t.Fatalf("buildWorld pending = (_, %t, %v), want ok", ok, err)
	}
	if !pendingWorld.WorkloadInFlight {
		t.Fatal("WorkloadInFlight = false while callback is pending")
	}

	fresh, _ := newOutstandingTestRunner(t, table, &fake.WorkloadOps{})
	fresh.amStore = store
	freshWorld, ok, err := fresh.buildWorld(context.Background(), "app1")
	if err != nil || !ok {
		t.Fatalf("fresh buildWorld = (_, %t, %v), want ok", ok, err)
	}
	if freshWorld.WorkloadInFlight {
		t.Fatal("fresh Runner WorkloadInFlight = true, want false")
	}
}

func TestOutstandingWorkloadRegistry_ConcurrentClaimsAreAtomic(t *testing.T) {
	started := make(chan struct{}, 8)
	release := make(chan struct{})
	defer close(release)
	ops := &fake.WorkloadOps{
		InstallFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-release
			return ports.WorkloadResult{}
		},
	}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)
	world := domain.World{AM: outstandingTestAM("app1", appv1.Installing)}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
				t.Errorf("launch: %v", err)
			}
		}()
	}
	wg.Wait()
	<-started
	if got := len(ops.Calls()); got != 1 {
		t.Fatalf("concurrent Install calls = %d, want 1", got)
	}
}

// TestLaunchCallWorkload_PreemptingCallRunsWhileAnotherIsInFlight
// pins the cancel-mid-download path at Runner level: a CallWorkload
// marked Preempts must take the AM's claim away from the goroutine
// already running, and that goroutine's late result must be dropped
// rather than delivered into the pipeline that replaced it.
func TestLaunchCallWorkload_PreemptingCallRunsWhileAnotherIsInFlight(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	ops := &fake.WorkloadOps{
		InstallFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-release
			return ports.WorkloadResult{}
		},
	}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)
	world := domain.World{AM: outstandingTestAM("app1", appv1.Downloading)}

	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("long-running launch: %v", err)
	}
	<-started

	preempting := domain.CallWorkload{
		Kind:      domain.WorkloadCancelImagePull,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "ImageCancelFailed"},
		Preempts:  true,
	}
	if err := r.launchCallWorkload(context.Background(), world, preempting); err != nil {
		t.Fatalf("preempting launch: %v", err)
	}
	waitForTest(t, "preempting callback", func() bool {
		action, queued := r.Pending().Peek("app1")
		_, ready := action.(domain.WorkloadReady)
		return queued && ready
	})

	// Let the preempted workload finish. Its WorkloadReady must not
	// reach the queue — the slot still holds the cancel's callback.
	close(release)
	waitForTest(t, "preempted goroutine to exit", func() bool {
		return r.InFlight() == 0
	})
	if !r.workloads.outstanding("app1") {
		t.Fatal("cancel's callback claim was released by the preempted goroutine")
	}
	if !r.workloads.acknowledgeCallback("app1") {
		t.Fatal("queued callback is not the cancel's; the preempted result overwrote it")
	}
}

func TestLaunchCallWorkload_DeadlineExceededDeliversTimeoutFired(t *testing.T) {
	table := domain.NewTransitionTable()
	table.SetTimeout(appv1.Installing, time.Minute)
	ops := &fake.WorkloadOps{
		InstallFunc: func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
			if req.Timeout > 0 {
				t.Errorf("req.Timeout = %s, want an already elapsed deadline", req.Timeout)
			}
			return ports.WorkloadResult{Err: ctx.Err()}
		},
	}
	r, store := newOutstandingTestRunner(t, table, ops)
	am := outstandingTestAM("app1", appv1.Installing)
	entered := metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Hour))
	am.Status.StatusTime = &entered
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatal(err)
	}
	waitForTest(t, "timeout callback", func() bool {
		action, queued := r.Pending().Peek("app1")
		_, ok := action.(domain.TimeoutFired)
		return queued && ok && r.InFlight() == 0
	})
}

func TestLaunchCallWorkload_DeadlineExceededWithSuccessDeliversReady(t *testing.T) {
	table := domain.NewTransitionTable()
	table.SetTimeout(appv1.Installing, time.Minute)
	ops := &fake.WorkloadOps{
		InstallFunc: func(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
			if req.Timeout > 0 {
				t.Errorf("req.Timeout = %s, want an already elapsed deadline", req.Timeout)
			}
			if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Errorf("ctx.Err() = %v, want DeadlineExceeded", ctx.Err())
			}
			return ports.WorkloadResult{}
		},
	}
	r, store := newOutstandingTestRunner(t, table, ops)
	am := outstandingTestAM("app1", appv1.Installing)
	entered := metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Hour))
	am.Status.StatusTime = &entered
	seedOutstandingTestAM(store, am)

	if err := r.launchCallWorkload(context.Background(), domain.World{AM: am}, installCall()); err != nil {
		t.Fatal(err)
	}
	waitForTest(t, "success callback", func() bool {
		action, queued := r.Pending().Peek("app1")
		_, ok := action.(domain.WorkloadReady)
		return queued && ok && r.InFlight() == 0
	})
}

// TestLaunchCallWorkload_PreemptCancelsAbandonedWorkloadContext pins
// that preemption tears down the abandoned goroutine's context.
// Polling ports have no deadline of their own — WaitForLaunch loops
// on a 1s ticker until its context is done — so without this the
// goroutine outlives the cancel it was interrupted by and keeps
// hitting the apiserver until the process exits.
//
// The workload here returns ONLY on ctx.Done(); nothing else releases
// it, so reaching InFlight()==0 is the assertion.
func TestLaunchCallWorkload_PreemptCancelsAbandonedWorkloadContext(t *testing.T) {
	started := make(chan struct{}, 1)
	ops := &fake.WorkloadOps{
		InstallFunc: func(ctx context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-ctx.Done()
			return ports.WorkloadResult{Err: ctx.Err()}
		},
	}
	r, _ := newOutstandingTestRunner(t, domain.NewTransitionTable(), ops)
	world := domain.World{AM: outstandingTestAM("app1", appv1.Downloading)}

	if err := r.launchCallWorkload(context.Background(), world, installCall()); err != nil {
		t.Fatalf("long-running launch: %v", err)
	}
	<-started

	preempting := domain.CallWorkload{
		Kind:      domain.WorkloadCancelImagePull,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "ImageCancelFailed"},
		Preempts:  true,
	}
	if err := r.launchCallWorkload(context.Background(), world, preempting); err != nil {
		t.Fatalf("preempting launch: %v", err)
	}

	waitForTest(t, "preempted goroutine to observe its canceled context", func() bool {
		return r.InFlight() == 0
	})
}

func TestStateEntryTimeFallback(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	am := &appv1.ApplicationManager{}
	if got := stateEntryTime(am, now); !got.Equal(now) {
		t.Fatalf("empty AM entry = %s, want now", got)
	}
	created := metav1.NewTime(now.Add(-2 * time.Hour))
	am.CreationTimestamp = created
	if got := stateEntryTime(am, now); !got.Equal(created.Time) {
		t.Fatalf("creation fallback = %s, want %s", got, created.Time)
	}
	updated := metav1.NewTime(now.Add(-time.Hour))
	am.Status.UpdateTime = &updated
	if got := stateEntryTime(am, now); !got.Equal(updated.Time) {
		t.Fatalf("update fallback = %s, want %s", got, updated.Time)
	}
	entered := metav1.NewTime(now.Add(-time.Minute))
	am.Status.StatusTime = &entered
	if got := stateEntryTime(am, now); !got.Equal(entered.Time) {
		t.Fatalf("status time = %s, want %s", got, entered.Time)
	}
}
