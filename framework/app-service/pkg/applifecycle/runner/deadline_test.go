package runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestReconcile_SameStatePatchDoesNotRefreshStatusTime(t *testing.T) {
	h := newHarness(t)
	entered := h.clock.Now().Add(-time.Minute)
	h.table.AddReachable(appv1.Installing, appv1.Installing)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
			State:   appv1.Installing,
			Message: "still installing",
		}}}}
	})
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	stamp := metav1.NewTime(entered)
	am.Status.StatusTime = &stamp
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if !got.Status.StatusTime.Time.Equal(entered) {
		t.Fatalf("StatusTime = %s, want %s", got.Status.StatusTime.Time, entered)
	}
	if got.Status.UpdateTime == nil || !got.Status.UpdateTime.Time.Equal(h.clock.Now()) {
		t.Fatalf("UpdateTime = %#v, want %s", got.Status.UpdateTime, h.clock.Now())
	}
}

func TestReconcile_StateChangeRefreshesStatusTime(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Installing, appv1.Downloading)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Next: appv1.Downloading, Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
			State: appv1.Downloading,
		}}}}
	})
	h.seedAM("app1", string(appv1.Installing))
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.StatusTime == nil || !got.Status.StatusTime.Time.Equal(h.clock.Now()) {
		t.Fatalf("StatusTime = %#v, want %s", got.Status.StatusTime, h.clock.Now())
	}
}

// TestReconcile_InstallFailedCleanupRetriesUntilNamespaceGone pins that
// a cleanup whose namespace outlives the wait keeps the AM in
// InstallFailed and retries after the backoff, landing Uninstalled
// only once the namespace is gone.
func TestReconcile_InstallFailedCleanupRetriesUntilNamespaceGone(t *testing.T) {
	h := newHarness(t)
	handlers.Register(h.table)
	nsWaits := 0
	h.workloadOps.WaitForNamespaceDeletedFunc = func(context.Context, ports.NamespaceWaitRequest) ports.WorkloadResult {
		nsWaits++
		if nsWaits == 1 {
			return ports.WorkloadResult{LandedReason: "NSPollTimeout", Err: errors.New("namespace still terminating")}
		}
		return ports.WorkloadResult{}
	}
	entered := metav1.NewTime(h.clock.Now().Add(-time.Hour))
	h.seedAM("app1", string(appv1.InstallFailed))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	h.amStore.Add(am)

	reconcile := func(step string) runner.ReconcileOutcome {
		t.Helper()
		out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
		if err != nil {
			t.Fatalf("%s: %v", step, err)
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) && h.runner.InFlight() != 0 {
			time.Sleep(time.Millisecond)
		}
		if h.runner.InFlight() != 0 {
			t.Fatalf("%s: workload did not finish", step)
		}
		return out
	}

	reconcile("start cleanup")
	out := reconcile("consume failure")
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.InstallFailed {
		t.Fatalf("after failed cleanup, state = %s, want InstallFailed", got.Status.State)
	}
	if got.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanupRetry {
		t.Fatalf("payload = %#v, want retry phase", got.Status.Payload)
	}
	if out.RequeueAfter != time.Minute {
		t.Fatalf("RequeueAfter = %s, want 1m", out.RequeueAfter)
	}

	reconcile("tick during backoff")
	if nsWaits != 1 {
		t.Fatalf("namespace waits = %d during backoff, want 1", nsWaits)
	}

	h.clock.Advance(time.Minute)
	reconcile("retry cleanup")
	reconcile("consume success")
	got, _ = h.amStore.Get(context.Background(), "app1")
	if nsWaits != 2 {
		t.Fatalf("namespace waits = %d, want 2", nsWaits)
	}
	if got.Status.State != appv1.Uninstalled {
		t.Fatalf("state = %s, want Uninstalled", got.Status.State)
	}
	if _, ok := got.Status.Payload[handlers.PayloadInstallFailedCleanupRetryAt]; ok {
		t.Fatalf("Uninstalled retained retry stamp: %#v", got.Status.Payload)
	}
}

func TestReconcile_InstallFailedCleanupDone_BecomesUninstalled(t *testing.T) {
	h := newHarness(t)
	handlers.Register(h.table)
	entered := metav1.NewTime(h.clock.Now().Add(-time.Minute))
	h.seedAM("app1", string(appv1.InstallFailed))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	am.Status.Payload = map[string]string{handlers.PayloadInstallFailedCleanup: handlers.PhaseInstallFailedCleanupDone}
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Uninstalled {
		t.Fatalf("state = %s, want Uninstalled", got.Status.State)
	}
	if got.Status.Reason != "InstallFailureCleanedUp" {
		t.Fatalf("reason = %q, want InstallFailureCleanedUp", got.Status.Reason)
	}
}

// TestReconcile_StaleTimeoutFiredDoesNotFailFreshCanceling pins that a
// TimeoutFired produced for the previous state's deadline is not applied
// after the AM has already entered InitializingCanceling. The new state's
// own deadline has not elapsed, so the reconcile ticks through to Stopping
// instead of writing InitializingCancelFailed.
func TestReconcile_StaleTimeoutFiredDoesNotFailFreshCanceling(t *testing.T) {
	h := newHarness(t)
	handlers.Register(h.table)
	h.table.AddReachable(appv1.InitializingCanceling, appv1.Stopping)
	h.table.SetTimeout(appv1.InitializingCanceling, 10*time.Second)
	h.seedAM("app1", string(appv1.InitializingCanceling))
	am, _ := h.amStore.Get(context.Background(), "app1")
	stamp := metav1.NewTime(h.clock.Now())
	am.Status.StatusTime = &stamp
	am.Status.Reason = constants.InitializeCancelBySystem
	h.amStore.Add(am)
	h.runner.Pending().Deliver("app1", domain.TimeoutFired{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("state = %s, want Stopping", got.Status.State)
	}
	if _, queued := h.runner.Pending().Peek("app1"); queued {
		t.Fatal("stale TimeoutFired was restored to the pending queue")
	}
}

func TestReconcile_ExpiredCancelingDeadlineStillFails(t *testing.T) {
	h := newHarness(t)
	handlers.Register(h.table)
	h.table.AddReachable(appv1.InitializingCanceling, appv1.InitializingCancelFailed)
	h.table.SetTimeout(appv1.InitializingCanceling, 10*time.Second)
	h.seedAM("app1", string(appv1.InitializingCanceling))
	am, _ := h.amStore.Get(context.Background(), "app1")
	stamp := metav1.NewTime(h.clock.Now().Add(-10 * time.Second))
	am.Status.StatusTime = &stamp
	h.amStore.Add(am)
	h.runner.Pending().Deliver("app1", domain.TimeoutFired{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.InitializingCancelFailed {
		t.Fatalf("state = %s, want InitializingCancelFailed", got.Status.State)
	}
	if got.Status.Reason != "InitializingCancelTimedOut" {
		t.Fatalf("reason = %q, want InitializingCancelTimedOut", got.Status.Reason)
	}
}

func TestReconcile_ExpiredDeadlineBeatsQueuedCallbackAndUserOp(t *testing.T) {
	h := newHarness(t)
	h.table.SetTimeout(appv1.Installing, 30*time.Minute)
	h.table.AddReachable(appv1.Installing, appv1.InstallingCanceling)
	handlers.Register(h.table)
	entered := metav1.NewTime(h.clock.Now().Add(-30 * time.Minute))
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	am.Status.OpID = "old"
	am.Spec.OpType = appv1.CancelOp
	if am.Annotations == nil {
		am.Annotations = map[string]string{}
	}
	am.Annotations[domain.AnnotationOperationRequestID] = "new"
	h.amStore.Add(am)
	h.runner.Pending().Deliver("app1", domain.WorkloadReady{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.InstallingCanceling {
		t.Fatalf("state = %s, want InstallingCanceling", got.Status.State)
	}
	if got.Status.Payload[handlers.PayloadInstallingFailReason] != constants.InstallCancelBySystem {
		t.Fatalf("reason stash = %#v", got.Status.Payload)
	}
}

// TestReconcile_CompletedCallbackBeatsExpiredDeadline pins that a
// workload which already returned success is consumed even when this
// Reconcile runs after the state deadline. The deadline still preempts
// a workload that is running; it must not tear down work whose result
// is already workloadResultQueued.
func TestReconcile_CompletedCallbackBeatsExpiredDeadline(t *testing.T) {
	h := newHarness(t)
	h.workloadOps.InstallFunc = func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{}
	}
	h.table.SetTimeout(appv1.Installing, 10*time.Minute)
	h.table.AddReachable(appv1.Installing, appv1.Initializing)
	h.table.AddReachable(appv1.Installing, appv1.InstallingCanceling)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{domain.CallWorkload{
			Kind: domain.WorkloadInstall, OnSuccess: domain.WorkloadReady{}, OnFailure: domain.WorkloadFailed{},
		}}}
	})
	h.table.SetOrAddHandler(appv1.Installing, domain.WorkloadReady{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{
			Next: appv1.Initializing,
			Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State: appv1.Initializing,
			}}},
		}
	})
	h.table.SetOrAddHandler(appv1.Installing, domain.TimeoutFired{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{
			Next: appv1.InstallingCanceling,
			Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State: appv1.InstallingCanceling,
			}}},
		}
	})
	entered := metav1.NewTime(h.clock.Now().Add(-time.Minute))
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && h.runner.InFlight() != 0 {
		time.Sleep(time.Millisecond)
	}
	if h.runner.InFlight() != 0 {
		t.Fatal("install workload did not finish")
	}
	if _, queued := h.runner.Pending().Peek("app1"); !queued {
		t.Fatal("completed install did not queue its callback")
	}

	h.clock.Advance(9 * time.Minute)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("state = %s, want Initializing (completed callback beats the deadline)", got.Status.State)
	}
}

func TestLaunchCallWorkload_RequestTimeoutMatchesRemainingDeadline(t *testing.T) {
	h := newHarness(t)
	h.table.SetTimeout(appv1.Installing, 30*time.Minute)
	h.table.AddReachable(appv1.Installing, appv1.Installing)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{domain.CallWorkload{
			Kind: domain.WorkloadInstall, OnSuccess: domain.WorkloadReady{}, OnFailure: domain.WorkloadFailed{},
		}}}
	})
	entered := metav1.NewTime(h.clock.Now().Add(-10 * time.Minute))
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	h.amStore.Add(am)

	var got time.Duration
	h.workloadOps.InstallFunc = func(_ context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
		got = req.Timeout
		return ports.WorkloadResult{}
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && h.runner.InFlight() != 0 {
		time.Sleep(time.Millisecond)
	}
	if h.runner.InFlight() != 0 {
		t.Fatal("install workload did not finish")
	}
	if got != 20*time.Minute {
		t.Fatalf("req.Timeout = %s, want 20m", got)
	}
}

func TestReconcile_NilStatusTimeIsInitialized(t *testing.T) {
	h := newHarness(t)
	h.table.AddReachable(appv1.Installing, appv1.Installing)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
			State:   appv1.Installing,
			Message: "progress",
		}}}}
	})
	h.seedAM("app1", string(appv1.Installing))
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.StatusTime == nil || !got.Status.StatusTime.Time.Equal(h.clock.Now()) {
		t.Fatalf("StatusTime = %#v, want %s", got.Status.StatusTime, h.clock.Now())
	}
}

func TestReconcile_TimeoutDispatchConflictRestoresTimeoutFired(t *testing.T) {
	h := newHarness(t)
	h.table.SetTimeout(appv1.Installing, time.Minute)
	h.table.SetOrAddHandler(appv1.Installing, domain.TimeoutFired{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{State: appv1.Installing, Message: "first"}},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{State: appv1.Installing, Message: "second"}},
		}}
	})
	entered := metav1.NewTime(h.clock.Now().Add(-time.Minute))
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	h.amStore.Add(am)
	h.runner.Pending().Deliver("app1", domain.WorkloadReady{})

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Requeue {
		t.Fatalf("outcome = %+v, want Requeue", out)
	}
	restored, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty")
	}
	if _, ok := restored.(domain.TimeoutFired); !ok {
		t.Fatalf("restored action = %T, want TimeoutFired", restored)
	}
}

func TestReconcile_ExpiredDeadlineCancelsBlockedWorkload(t *testing.T) {
	h := newHarness(t)
	started := make(chan struct{}, 1)
	h.workloadOps.InstallFunc = func(ctx context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		started <- struct{}{}
		<-ctx.Done()
		return ports.WorkloadResult{}
	}
	h.table.SetTimeout(appv1.Installing, 10*time.Minute)
	h.table.AddReachable(appv1.Installing, appv1.InstallingCanceling)
	h.table.SetOrAddHandler(appv1.Installing, domain.Tick{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{Effects: []domain.Effect{domain.CallWorkload{
			Kind: domain.WorkloadInstall, OnSuccess: domain.WorkloadReady{}, OnFailure: domain.WorkloadFailed{},
		}}}
	})
	h.table.SetOrAddHandler(appv1.Installing, domain.TimeoutFired{}, func(context.Context, domain.World) domain.Result {
		return domain.Result{
			Next: appv1.InstallingCanceling,
			Effects: []domain.Effect{domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State: appv1.InstallingCanceling,
			}}},
		}
	})
	entered := metav1.NewTime(h.clock.Now().Add(-time.Minute))
	h.seedAM("app1", string(appv1.Installing))
	am, _ := h.amStore.Get(context.Background(), "app1")
	am.Status.StatusTime = &entered
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("install did not start")
	}

	h.clock.Advance(9 * time.Minute)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && h.runner.InFlight() != 0 {
		time.Sleep(time.Millisecond)
	}
	if h.runner.InFlight() != 0 {
		t.Fatal("blocked install survived the deadline")
	}
	got, _ := h.amStore.Get(context.Background(), "app1")
	if got.Status.State != appv1.InstallingCanceling {
		t.Fatalf("state = %s, want InstallingCanceling", got.Status.State)
	}
	if action, queued := h.runner.Pending().Peek("app1"); queued {
		t.Fatalf("late workload result queued as %T", action)
	}
}
