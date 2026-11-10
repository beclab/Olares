package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

type coldStartHarness struct {
	store       *fake.AMStore
	workloadOps *fake.WorkloadOps
	pending     *runner.PendingActionQueue
	clock       *fake.Clock
	runner      *runner.Runner
}

func newColdStartHarness(t *testing.T, workloadOps *fake.WorkloadOps) *coldStartHarness {
	t.Helper()

	table := domain.NewDefaultTable()
	handlers.Register(table)
	pending := runner.NewPendingActionQueue()
	store := fake.NewAMStore()
	clock := fake.NewClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	r, err := runner.New(runner.Config{
		Engine:      domain.NewEngine(table),
		AMStore:     store,
		WorkloadOps: workloadOps,
		Validator:   &fake.InstallValidator{},
		ImagePuller: fake.NewImagePuller(),
		Recorder:    &fake.EventRecorder{},
		Clock:       clock,
		Pending:     pending,
		Requeuer:    &runner.RecordingRequeuer{},
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	t.Cleanup(r.Stop)
	return &coldStartHarness{
		store:       store,
		workloadOps: workloadOps,
		pending:     pending,
		clock:       clock,
		runner:      r,
	}
}

func (h *coldStartHarness) seed(name string, state appv1.ApplicationManagerState, payload map[string]string) {
	h.store.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	h.store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Type:         appv1.App,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   state,
			Payload: payload,
		},
	})
}

func waitForColdStartCallback(t *testing.T, pending *runner.PendingActionQueue, name string) domain.Action {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if action, ok := pending.Peek(name); ok {
			return action
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for callback for %s", name)
	return nil
}

func TestColdStartRecovery_RedispatchesEveryPersistedPhaseAndDeliversCallback(t *testing.T) {
	tests := []struct {
		name       string
		state      appv1.ApplicationManagerState
		payloadKey string
		phase      string
	}{
		{"downloading_pull", appv1.Downloading, handlers.PayloadDownloadingPhase, handlers.PhaseDownloadingPull},

		{"installing_expose_ports", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingExposePorts},
		{"installing_validate", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingValidate},
		{"installing_helm", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingHelmInstall},
		{"installing_post_validate", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingPostValidate},
		{"installing_scale_up", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingScaleUp},
		{"installing_wait_startup", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingWaitStartUp},
		{"installing_wait_launch", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingWaitLaunch},
		{"installing_cleanup", appv1.Installing, handlers.PayloadInstallingPhase, handlers.PhaseInstallingCleanup},

		{"initializing_wait", appv1.Initializing, handlers.PayloadInitializingPhase, handlers.PhaseInitializingWait},
		{"stopping_scale", appv1.Stopping, handlers.PayloadStoppingPhase, handlers.PhaseStoppingScale},
		{"resuming_scale", appv1.Resuming, handlers.PayloadResumingPhase, handlers.PhaseResumingScale},
		{"upgrading_prepare", appv1.Upgrading, handlers.PayloadUpgradingPhase, handlers.PhaseUpgradingPrepare},
		{"upgrading_pull", appv1.Upgrading, handlers.PayloadUpgradingPhase, handlers.PhaseUpgradingPull},
		{"upgrading_helm", appv1.Upgrading, handlers.PayloadUpgradingPhase, handlers.PhaseUpgradingHelm},
		{"upgrading_wait_startup", appv1.Upgrading, handlers.PayloadUpgradingPhase, handlers.PhaseUpgradingWaitStartUp},
		{"applying_env_helm", appv1.ApplyingEnv, handlers.PayloadApplyingEnvPhase, handlers.PhaseApplyingEnvHelm},
		{"uninstalling_helm", appv1.Uninstalling, handlers.PayloadUninstallingPhase, handlers.PhaseUninstallingHelm},
		{"uninstalling_wait_namespace", appv1.Uninstalling, handlers.PayloadUninstallingPhase, handlers.PhaseUninstallingWaitNS},

		{"pending_canceling_wait_namespace", appv1.PendingCanceling, handlers.PayloadPendingCancelingPhase, handlers.PhaseCancelWaitNS},
		{"downloading_canceling_pull", appv1.DownloadingCanceling, handlers.PayloadDownloadingCancelingPhase, handlers.PhaseCancelImagePull},
		{"downloading_canceling_wait_namespace", appv1.DownloadingCanceling, handlers.PayloadDownloadingCancelingPhase, handlers.PhaseCancelWaitNS},
		{"installing_canceling_uninstall", appv1.InstallingCanceling, handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelHelmUninstall},
		{"installing_canceling_wait_namespace", appv1.InstallingCanceling, handlers.PayloadInstallingCancelingPhase, handlers.PhaseCancelWaitNS},

		{"install_failed_cleanup", appv1.InstallFailed, handlers.PayloadInstallFailedCleanup, handlers.PhaseInstallFailedCleanup},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newColdStartHarness(t, &fake.WorkloadOps{})
			name := "cold-" + tc.name
			h.seed(name, tc.state, map[string]string{tc.payloadKey: tc.phase})

			if _, queued := h.pending.Peek(name); queued {
				t.Fatal("fresh PendingActionQueue unexpectedly contains an action")
			}
			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: name}); err != nil {
				t.Fatalf("Reconcile: %v", err)
			}
			if tc.phase == handlers.PhaseDownloadingPull ||
				(tc.state == appv1.Upgrading && tc.phase == handlers.PhaseUpgradingPull) {
				deadline := time.Now().Add(2 * time.Second)
				for h.clock.Pending() == 0 && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if h.clock.Pending() == 0 {
					t.Fatal("image recovery did not start its progress ticker")
				}
				h.clock.Advance(time.Second)
			}

			action := waitForColdStartCallback(t, h.pending, name)
			if _, ok := action.(domain.WorkloadReady); !ok {
				t.Fatalf("callback = %T, want WorkloadReady", action)
			}
		})
	}
}

func TestColdStartRecovery_PersistedUserOperationUsesRealHandler(t *testing.T) {
	h := newColdStartHarness(t, &fake.WorkloadOps{})
	h.seed("durable-stop", appv1.Running, nil)

	am, err := h.store.Get(context.Background(), "durable-stop")
	if err != nil {
		t.Fatalf("get seeded AM: %v", err)
	}
	am.Spec.OpType = appv1.StopOp
	am.Annotations = map[string]string{
		domain.AnnotationOperationRequestID: "request-after-restart",
	}
	if err := h.store.UpdateFull(context.Background(), am); err != nil {
		t.Fatalf("persist operation request: %v", err)
	}
	if _, queued := h.pending.Peek(am.Name); queued {
		t.Fatal("fresh PendingActionQueue unexpectedly contains an action")
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: am.Name}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got, err := h.store.Get(context.Background(), am.Name)
	if err != nil {
		t.Fatalf("get reconciled AM: %v", err)
	}
	if got.Status.State != appv1.Stopping {
		t.Fatalf("state = %q, want %q", got.Status.State, appv1.Stopping)
	}
	if got.Status.OpType != appv1.StopOp || got.Status.OpID != "request-after-restart" {
		t.Fatalf("operation identity = (%q, %q), want (%q, %q)",
			got.Status.OpType, got.Status.OpID, appv1.StopOp, "request-after-restart")
	}
}

func TestColdStartRecovery_SameProcessTickDoesNotDuplicateRealHandlerWorkload(t *testing.T) {
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	defer close(release)
	ops := &fake.WorkloadOps{
		StopFunc: func(context.Context, ports.WorkloadRequest) ports.WorkloadResult {
			started <- struct{}{}
			<-release
			return ports.WorkloadResult{}
		},
	}
	h := newColdStartHarness(t, ops)
	h.seed("same-process-stop", appv1.Stopping, map[string]string{
		handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
	})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "same-process-stop"}); err != nil {
		t.Fatalf("first Reconcile: %v", err)
	}
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first Stop workload")
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "same-process-stop"}); err != nil {
		t.Fatalf("second Reconcile: %v", err)
	}
	if got := len(ops.Calls()); got != 1 {
		t.Fatalf("Stop calls after second Tick = %d, want 1", got)
	}
}
