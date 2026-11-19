package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newHandoffAM(name string, state appv1.ApplicationManagerState) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.CancelOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:  state,
			OpType: appv1.CancelOp,
			OpID:   "op-42",
		},
	}
}

func handoffWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — CancelOp source→*Canceling matrix
// -----------------------------------------------------------------
//
// The four handoff CancelOp handlers share sourceToCanceling
// semantics. One table-driven test locks the (source, handler,
// target, reason) invariant for all of them.

type handoffCancelCase struct {
	name    string
	source  appv1.ApplicationManagerState
	handler func(context.Context, domain.World) domain.Result
	target  appv1.ApplicationManagerState
}

func handoffCancelMatrix() []handoffCancelCase {
	return []handoffCancelCase{
		{"initializing", appv1.Initializing, handlers.OnInitializingCancel, appv1.InitializingCanceling},
		{"resuming", appv1.Resuming, handlers.OnResumingCancel, appv1.ResumingCanceling},
		{"upgrading", appv1.Upgrading, handlers.OnUpgradingCancel, appv1.UpgradingCanceling},
		{"applyingenv", appv1.ApplyingEnv, handlers.OnApplyingEnvCancel, appv1.ApplyingEnvCanceling},
	}
}

// TestHandoffCancel_Transitions pins the source→*Canceling flip
// for all four Pattern C handlers.
func TestHandoffCancel_Transitions(t *testing.T) {
	for _, tc := range handoffCancelMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newHandoffAM("app1", tc.source)
			res := tc.handler(context.Background(), handoffWorldFor(am))
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != tc.target {
				t.Fatalf("Next = %q, want %q", res.Next, tc.target)
			}
			patch, ok := findEffect[domain.PatchAM](res.Effects)
			if !ok {
				t.Fatal("no PatchAM")
			}
			if patch.Status.State != tc.target {
				t.Fatalf("PatchAM.State = %q, want %q", patch.Status.State, tc.target)
			}
			if patch.Status.Reason != "CancelRequested" {
				t.Fatalf("Reason = %q, want CancelRequested", patch.Status.Reason)
			}
			// Payload cleared to drop any source-state phase key
			// (installing.phase, initializing.phase, …) so the
			// *Canceling state starts fresh.
			if patch.Status.Payload == nil || len(patch.Status.Payload) != 0 {
				t.Fatalf("Payload = %+v, want empty non-nil map (clears source phase)", patch.Status.Payload)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok {
				t.Fatal("no EmitEvent")
			}
			if ev.Type != "Normal" {
				t.Fatalf("EmitEvent.Type = %q, want Normal", ev.Type)
			}
			// No OpRecord at CancelOp entry — record fires at
			// the downstream Stopping terminal.
			if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
				t.Fatal("unexpected AppendOpRecord on source→*Canceling edge")
			}
			// No CallWorkload — Pattern C is a pure state flip.
			if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
				t.Fatal("unexpected CallWorkload on Pattern C entry")
			}
		})
	}
}

// -----------------------------------------------------------------
// Layer 1 — *Canceling Tick handoff-to-Stopping matrix
// -----------------------------------------------------------------

type handoffTickCase struct {
	name    string
	state   appv1.ApplicationManagerState
	handler func(context.Context, domain.World) domain.Result
}

func handoffTickMatrix() []handoffTickCase {
	return []handoffTickCase{
		{"initializingCanceling", appv1.InitializingCanceling, handlers.OnInitializingCancelingTick},
		{"resumingCanceling", appv1.ResumingCanceling, handlers.OnResumingCancelingTick},
		{"applyingEnvCanceling", appv1.ApplyingEnvCanceling, handlers.OnApplyingEnvCancelingTick},
	}
}

// TestHandoffTick_TransitionsToStopping pins the *Canceling → Stopping
// handoff for all four Pattern C states.
func TestHandoffTick_TransitionsToStopping(t *testing.T) {
	for _, tc := range handoffTickMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newHandoffAM("app1", tc.state)
			res := tc.handler(context.Background(), handoffWorldFor(am))
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != appv1.Stopping {
				t.Fatalf("Next = %q, want Stopping", res.Next)
			}
			patch, ok := findEffect[domain.PatchAM](res.Effects)
			if !ok {
				t.Fatal("no PatchAM")
			}
			if patch.Status.State != appv1.Stopping {
				t.Fatalf("PatchAM.State = %q, want Stopping", patch.Status.State)
			}
			if patch.Status.Reason != "CancelHandoffToStopping" {
				t.Fatalf("Reason = %q, want CancelHandoffToStopping", patch.Status.Reason)
			}
			if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
				t.Fatal("unexpected CallWorkload on handoff (Stopping owns the helm call)")
			}
		})
	}
}

func TestHandoffTick_KeepsInitFailed(t *testing.T) {
	am := newHandoffAM("app1", appv1.InitializingCanceling)
	am.Status.Reason = "InitFailed"
	res := handlers.OnInitializingCancelingTick(context.Background(), handoffWorldFor(am))
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok || patch.Status.Reason != "InitFailed" {
		t.Fatalf("Reason = %q, want InitFailed", patch.Status.Reason)
	}
	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok || ev.Reason != "InitFailed" {
		t.Fatalf("event Reason = %q, want InitFailed", ev.Reason)
	}
}

func TestOnResumingCancelingTick_InFlight_WaitsForHelm(t *testing.T) {
	am := newHandoffAM("app1", appv1.ResumingCanceling)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadResume
	res := handlers.OnResumingCancelingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in ResumingCanceling until helm returns)", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight (helm context must be canceled without dropping the claim)")
	}
	if _, ok := findEffect[domain.PatchAM](res.Effects); ok {
		t.Fatal("in-flight tick must not patch toward Stopping")
	}
}

func TestOnResumingCancel_InFlight_CancelsHelm(t *testing.T) {
	am := newHandoffAM("app1", appv1.Resuming)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadResume
	res := handlers.OnResumingCancel(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.ResumingCanceling {
		t.Fatalf("Next = %q, want ResumingCanceling", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight")
	}
}

func TestOnResumingCancelingWorkloadFailed_HandsOffToStopping(t *testing.T) {
	am := newHandoffAM("app1", appv1.ResumingCanceling)
	w := handoffWorldFor(am)
	w.Action = domain.WorkloadFailed{Reason: "HelmResumeFailed", Message: "context canceled"}
	res := handlers.OnResumingCancelingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
}

func TestOnResumingCancelingWorkloadReady_HandsOffToStopping(t *testing.T) {
	am := newHandoffAM("app1", appv1.ResumingCanceling)
	res := handlers.OnResumingCancelingWorkloadReady(context.Background(), handoffWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
}

func TestOnApplyingEnvCancelingTick_InFlight_WaitsForHelm(t *testing.T) {
	am := newHandoffAM("app1", appv1.ApplyingEnvCanceling)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadApplyEnv
	res := handlers.OnApplyingEnvCancelingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in ApplyingEnvCanceling until helm returns)", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight (helm context must be canceled without dropping the claim)")
	}
	if _, ok := findEffect[domain.PatchAM](res.Effects); ok {
		t.Fatal("in-flight tick must not patch toward Stopping")
	}
}

func TestOnApplyingEnvCancel_InFlight_CancelsHelm(t *testing.T) {
	am := newHandoffAM("app1", appv1.ApplyingEnv)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadApplyEnv
	res := handlers.OnApplyingEnvCancel(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.ApplyingEnvCanceling {
		t.Fatalf("Next = %q, want ApplyingEnvCanceling", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight")
	}
}

func TestOnApplyingEnvCancelingWorkloadFailed_HandsOffToStopping(t *testing.T) {
	am := newHandoffAM("app1", appv1.ApplyingEnvCanceling)
	w := handoffWorldFor(am)
	w.Action = domain.WorkloadFailed{Reason: "HelmApplyEnvFailed", Message: "context canceled"}
	res := handlers.OnApplyingEnvCancelingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
}

func TestOnApplyingEnvCancelingWorkloadReady_HandsOffToStopping(t *testing.T) {
	am := newHandoffAM("app1", appv1.ApplyingEnvCanceling)
	res := handlers.OnApplyingEnvCancelingWorkloadReady(context.Background(), handoffWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
}

func TestOnUpgradingCancel_InFlightHelm_CancelsWithoutDroppingClaim(t *testing.T) {
	am := newHandoffAM("app1", appv1.Upgrading)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadUpgrade
	res := handlers.OnUpgradingCancel(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.UpgradingCanceling {
		t.Fatalf("Next = %q, want UpgradingCanceling", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight")
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("UpgradingCanceling must not cancel the image pull")
	}
}

func TestOnUpgradingCancelingTick_InFlight_Waits(t *testing.T) {
	am := newHandoffAM("app1", appv1.UpgradingCanceling)
	w := handoffWorldFor(am)
	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadUpgrade
	res := handlers.OnUpgradingCancelingTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay until the upgrade call returns)", res.Next)
	}
	if _, ok := findEffect[domain.CancelInFlight](res.Effects); !ok {
		t.Fatal("missing CancelInFlight")
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("UpgradingCanceling must not cancel the image pull")
	}
}

func TestOnUpgradingCancelingTick_Idle_LandsUpgradeFailed(t *testing.T) {
	am := newHandoffAM("app1", appv1.UpgradingCanceling)
	res := handlers.OnUpgradingCancelingTick(context.Background(), handoffWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.UpgradeFailed {
		t.Fatalf("Next = %q, want UpgradeFailed", res.Next)
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("UpgradingCanceling must not cancel the image pull")
	}
}

func TestOnUpgradingCancelingWorkloadReady_LandsInUpgradeFailed(t *testing.T) {
	am := newHandoffAM("app1", appv1.UpgradingCanceling)
	am.Spec.Config = `{"version":"2.0.0"}`
	am.Status.Message = "user requested cancel"
	res := handlers.OnUpgradingCancelingWorkloadReady(context.Background(), handoffWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.UpgradeFailed {
		t.Fatalf("Next = %q, want UpgradeFailed", res.Next)
	}
	if _, ok := findEffect[domain.PatchSpec](res.Effects); ok {
		t.Fatal("in-progress cancel must not restore config or rewrite the deployed version")
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("UpgradingCanceling must not cancel the image pull")
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok || patch.Status.State != appv1.UpgradeFailed {
		t.Fatalf("PatchAM = %+v ok=%v, want UpgradeFailed", patch, ok)
	}
	if patch.Status.Reason != string(appv1.UpgradeFailed) {
		t.Fatalf("Reason = %q, want UpgradeFailed", patch.Status.Reason)
	}
}

func TestOnUpgradingCancelingWorkloadFailed_LandsUpgradeFailed(t *testing.T) {
	am := newHandoffAM("app1", appv1.UpgradingCanceling)
	w := handoffWorldFor(am)
	w.Action = domain.WorkloadFailed{Reason: "HelmUpgradeFailed", Message: "context canceled"}
	res := handlers.OnUpgradingCancelingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.UpgradeFailed {
		t.Fatalf("Next = %q, want UpgradeFailed", res.Next)
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("UpgradingCanceling must not cancel the image pull")
	}
}

// TestHandoffCancel_NilAM_HardFails pins the defensive Err
// path across all handoff handlers.
func TestHandoffCancel_NilAM_HardFails(t *testing.T) {
	cancelFns := []func(context.Context, domain.World) domain.Result{
		handlers.OnInitializingCancel, handlers.OnResumingCancel,
		handlers.OnUpgradingCancel, handlers.OnApplyingEnvCancel,
	}
	tickFns := []func(context.Context, domain.World) domain.Result{
		handlers.OnInitializingCancelingTick, handlers.OnResumingCancelingTick,
		handlers.OnUpgradingCancelingTick, handlers.OnApplyingEnvCancelingTick,
	}
	for i, fn := range cancelFns {
		if res := fn(context.Background(), domain.World{AM: nil}); res.Err == nil {
			t.Errorf("cancel handler #%d: Err = nil, want failure for nil AM", i)
		}
	}
	for i, fn := range tickFns {
		if res := fn(context.Background(), domain.World{AM: nil}); res.Err == nil {
			t.Errorf("tick handler #%d: Err = nil, want failure for nil AM", i)
		}
	}
}

// -----------------------------------------------------------------
// Layer 2 — end-to-end lifecycle: Running → Upgrading → cancel → Stopping
// -----------------------------------------------------------------

// TestFullLifecycle_UpgradeCancel drives a Running AM through the
// upgrade download queue, cancels while the image pull is still at
// 40%, and restores Running. Helm has not started, so cancel does
// not go through Stopping.
func TestFullLifecycle_UpgradeCancel(t *testing.T) {
	h := newE2EHarness(t)
	am := newRunningAM("myapp", appv1.App)
	h.amStore.Add(am)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

	h.imagePuller.SetProgress("myapp", ports.ImageProgress{Percent: 40})
	h.imagePuller.CancelJobFunc = func(_ context.Context, name string) error {
		h.imagePuller.SetProgress(name, ports.ImageProgress{
			Percent: 40,
			Failed:  true,
			Reason:  appv1.DownloadingCanceled.String(),
		})
		return nil
	}

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.UpgradeOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradePending {
		t.Fatalf("after R1, state = %q, want UpgradePending", got.Status.State)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradeDownloading {
		t.Fatalf("after R2, state = %q, want UpgradeDownloading", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadUpgradeDownloadingPhase]; p != handlers.PhaseUpgradingPull {
		t.Fatalf("after R4, phase = %q, want pull", p)
	}
	awaitPullParked(t, h)

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradeDownloadingCanceling {
		t.Fatalf("after R5, state = %q, want UpgradeDownloadingCanceling", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	drainAndRebuild(t, h)

	canceled := false
	for _, c := range h.imagePuller.Calls() {
		if c.Method == "CancelJob" {
			canceled = true
			break
		}
	}
	if !canceled {
		t.Fatal("expected ImagePuller.CancelJob during upgrade download cancel")
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("after cancel, state = %q, want Running", got.Status.State)
	}
}

// TestFullLifecycle_ApplyEnvCancel waits for the in-flight helm
// apply-env to return before Stopping starts Scale(0). Same shape
// as TestFullLifecycle_ResumingCancel_MidScale.
func TestFullLifecycle_ApplyEnvCancel(t *testing.T) {
	h := newE2EHarness(t)
	am := newRunningAM("myapp", appv1.App)
	h.amStore.Add(am)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

	started := make(chan struct{})
	returned := make(chan struct{})
	h.workloadOps.ApplyEnvFunc = func(ctx context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		close(started)
		<-ctx.Done()
		close(returned)
		return ports.WorkloadResult{Err: ctx.Err(), LandedReason: "HelmApplyEnvFailed"}
	}

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.ApplyEnvOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ApplyingEnv {
		t.Fatalf("after R1, state = %q, want ApplyingEnv", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	awaitChan(t, started, "apply-env to hold the claim")
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ApplyingEnvCanceling {
		t.Fatalf("after R3, state = %q, want ApplyingEnvCanceling", got.Status.State)
	}
	awaitChan(t, returned, "apply-env helm to return before Stopping")
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ApplyingEnvCanceling {
		t.Fatalf("before helm callback, state = %q, want ApplyingEnvCanceling", got.Status.State)
	}

	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("after helm WorkloadFailed, state = %q, want Stopping", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("final state = %q, want Stopped", got.Status.State)
	}
}

func TestFullLifecycle_InitializingCancel_MidLaunch(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInitializing("myapp", appv1.InstallOp)

	started, release, fn := parkedWorkload()
	h.workloadOps.WaitForLaunchFunc = fn

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	awaitChan(t, started, "wait-launch to hold the claim")
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InitializingCanceling {
		t.Fatalf("after R2, state = %q, want InitializingCanceling", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("after R3, state = %q, want Stopping", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	release()
	drainAndRebuild(t, h)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("final state = %q, want Stopped", got.Status.State)
	}
}

func TestFullLifecycle_ResumingCancel_MidScale(t *testing.T) {
	h := newE2EHarness(t)
	h.seedResuming("myapp")

	started := make(chan struct{})
	returned := make(chan struct{})
	h.workloadOps.ResumeFunc = func(ctx context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		close(started)
		<-ctx.Done()
		close(returned)
		return ports.WorkloadResult{Err: ctx.Err(), LandedReason: "HelmResumeFailed"}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	awaitChan(t, started, "resume to hold the claim")
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ResumingCanceling {
		t.Fatalf("after R2, state = %q, want ResumingCanceling", got.Status.State)
	}
	awaitChan(t, returned, "resume helm to return before Stopping")
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ResumingCanceling {
		t.Fatalf("before helm callback, state = %q, want ResumingCanceling", got.Status.State)
	}

	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("after helm WorkloadFailed, state = %q, want Stopping", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("final state = %q, want Stopped", got.Status.State)
	}
}
