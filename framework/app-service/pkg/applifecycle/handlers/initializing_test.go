package handlers_test

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

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

// newInitializingAM builds an AM in Initializing state with the
// given op type + payload. Op type matters because
// finishToRunning branches on it to pick the success message.
func newInitializingAM(name string, op appv1.OpType, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       op,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Initializing,
			OpType:  op,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func initializingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnInitializingTick_EmptyPhase_StartsWait pins the fresh-entry
// branch: dispatch WorkloadWaitLaunch + stamp phase=wait.
func TestOnInitializingTick_EmptyPhase_StartsWait(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, nil)
	res := handlers.OnInitializingTick(context.Background(), initializingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in Initializing)", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if got := patch.Status.Payload[handlers.PayloadInitializingPhase]; got != handlers.PhaseInitializingWait {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadInitializingPhase, got, handlers.PhaseInitializingWait)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload effect")
	}
	if cw.Kind != domain.WorkloadWaitLaunch {
		t.Fatalf("CallWorkload.Kind = %q, want WorkloadWaitLaunch", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("CallWorkload OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}

	// OnFailure must carry the legacy fallback reason so a
	// port that returns bare failure still produces a legible AM.
	wf, ok := cw.OnFailure.(domain.WorkloadFailed)
	if !ok {
		t.Fatalf("OnFailure = %T, want WorkloadFailed", cw.OnFailure)
	}
	if wf.Reason != constants.AppStopDueToInitFailed {
		t.Fatalf("OnFailure.Reason = %q, want %q", wf.Reason, constants.AppStopDueToInitFailed)
	}
}

func TestOnInitializingTick_WaitPhaseRecoversWithoutStatusEffects(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	assertTickRecovery(t, handlers.OnInitializingTick, initializingWorldFor(am), domain.CallWorkload{
		Kind:      domain.WorkloadWaitLaunch,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{
			Reason:  constants.AppStopDueToInitFailed,
			Message: constants.InitializeCancelDueToWaitForLaunchTimedOut,
		},
	})
}

// TestOnInitializingTick_UnknownPhase_HardFails — same safety net
// as its siblings.
func TestOnInitializingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, map[string]string{
		handlers.PayloadInitializingPhase: "grabbed-by-alien",
	})
	res := handlers.OnInitializingTick(context.Background(), initializingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnInitializingWorkloadReady_InstallOp_Advances pins the
// terminal success branch for Install: → Running with Install
// success message template.
func TestOnInitializingWorkloadReady_InstallOp_Advances(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	res := handlers.OnInitializingWorkloadReady(context.Background(), initializingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Running {
		t.Fatalf("Next = %q, want Running", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Running {
		t.Fatalf("PatchAM.State = %q, want Running", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadInitializingPhase]; ok {
		t.Fatal("Payload[initializing.phase] still set after leaving Initializing; must be cleared")
	}

	rec, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord effect")
	}
	if rec.Record.Status != appv1.Running {
		t.Fatalf("OpRecord.Status = %q, want Running", rec.Record.Status)
	}
	// The message should be the Install success template.
	if rec.Record.Message == "" || !containsInstallTemplate(rec.Record.Message) {
		t.Fatalf("OpRecord.Message = %q, want an Install-completion template", rec.Record.Message)
	}
}

// containsInstallTemplate returns true if s looks like the
// Install-op success template ("Successfully installed ..."). Kept
// as a shallow substring check to avoid coupling tests to the
// exact wording — the point is to prove that InstallOp routes to
// the Install template and not to Upgrade/ApplyEnv.
func containsInstallTemplate(s string) bool {
	return len(s) >= len("Successfully installed") && s[:len("Successfully installed")] == "Successfully installed"
}

// TestOnInitializingWorkloadReady_UpgradeOp_UsesUpgradeMessage pins
// the OpType-branching in finishToRunning: UpgradeOp must produce
// the Upgrade template, not Install.
func TestOnInitializingWorkloadReady_UpgradeOp_UsesUpgradeMessage(t *testing.T) {
	am := newInitializingAM("app1", appv1.UpgradeOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	res := handlers.OnInitializingWorkloadReady(context.Background(), initializingWorldFor(am))
	rec, _ := findEffect[domain.AppendOpRecord](res.Effects)
	if rec.Record.Message == "" || rec.Record.Message[:len("Successfully upgraded")] != "Successfully upgraded" {
		t.Fatalf("OpRecord.Message = %q, want Upgrade template", rec.Record.Message)
	}
}

// TestOnInitializingWorkloadReady_ApplyEnvOp_UsesApplyEnvMessage
// pins the third branch of the OpType switch.
func TestOnInitializingWorkloadReady_ApplyEnvOp_UsesApplyEnvMessage(t *testing.T) {
	am := newInitializingAM("app1", appv1.ApplyEnvOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	res := handlers.OnInitializingWorkloadReady(context.Background(), initializingWorldFor(am))
	rec, _ := findEffect[domain.AppendOpRecord](res.Effects)
	if rec.Record.Message == "" || rec.Record.Message[:len("Successfully applied env")] != "Successfully applied env" {
		t.Fatalf("OpRecord.Message = %q, want ApplyEnv template", rec.Record.Message)
	}
}

// TestOnInitializingWorkloadReady_EmptyPhase_HardFails pins the
// dispatch-invariant: every CallWorkload must stamp the phase.
func TestOnInitializingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, nil)
	res := handlers.OnInitializingWorkloadReady(context.Background(), initializingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

// TestOnInitializingWorkloadFailed_TransitionsToInitializingCanceling
// pins the LEGACY failure routing: wait-launch failure does NOT
// go to a "*Failed" state; it goes to InitializingCanceling so
// the unwind path can rollback the helm release + IM.
func TestOnInitializingWorkloadFailed_TransitionsToInitializingCanceling(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	w := initializingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  constants.AppStopDueToInitFailed,
		Message: constants.InitializeCancelDueToWaitForLaunchTimedOut,
	}

	res := handlers.OnInitializingWorkloadFailed(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InitializingCanceling {
		t.Fatalf("Next = %q, want InitializingCanceling (legacy funnels wait-launch fail into cancel)", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.InitializingCanceling {
		t.Fatalf("PatchAM.State = %q, want InitializingCanceling", patch.Status.State)
	}
	if patch.Status.Reason != constants.AppStopDueToInitFailed {
		t.Fatalf("Reason = %q, want %q", patch.Status.Reason, constants.AppStopDueToInitFailed)
	}

	// No OpRecord on wait-launch failure — legacy passes nil for
	// opRecord in updateStatus (initializing_app.go:99-108). The
	// terminal record for the whole lifecycle lands when
	// InitializingCanceling unwind completes.
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on Initializing→InitializingCanceling; legacy writes none here")
	}

	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok {
		t.Fatal("no EmitEvent")
	}
	if ev.Type != "Warning" {
		t.Fatalf("EmitEvent.Type = %q, want Warning", ev.Type)
	}
}

// TestOnInitializingWorkloadFailed_FallsBackWhenActionEmpty pins
// the safety net: legacy fallback strings when the Action carries
// no data.
func TestOnInitializingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newInitializingAM("app1", appv1.InstallOp, map[string]string{
		handlers.PayloadInitializingPhase: handlers.PhaseInitializingWait,
	})
	w := initializingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnInitializingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != constants.AppStopDueToInitFailed {
		t.Fatalf("fallback Reason = %q, want %q", patch.Status.Reason, constants.AppStopDueToInitFailed)
	}
	if patch.Status.Message != constants.InitializeCancelDueToWaitForLaunchTimedOut {
		t.Fatalf("fallback Message = %q, want %q", patch.Status.Message, constants.InitializeCancelDueToWaitForLaunchTimedOut)
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// seedInitializing seeds an Initializing AM with empty phase.
func (h *e2eHarness) seedInitializing(name string, op appv1.OpType) string {
	am := newInitializingAM(name, op, nil)
	// Initializing state is entered after Installing completed
	// helm install — so the paired Application CR EXISTS.
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestInit_EndToEnd_HappyPath drives Initializing → Running:
//
//	R1: Initializing(phase="")  → dispatch WaitLaunch
//	R2: Initializing(phase=wait) → WorkloadReady → Running
func TestInit_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInitializing("myapp", appv1.InstallOp)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after R1, state = %q, want Initializing", got.Status.State)
	}
	if p := got.Status.Payload[handlers.PayloadInitializingPhase]; p != handlers.PhaseInitializingWait {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseInitializingWait)
	}
	// WaitForLaunch was dispatched.
	calls := h.workloadOps.Calls()
	if len(calls) == 0 || calls[len(calls)-1].Method != "WaitForLaunch" {
		t.Fatalf("workload calls = %+v, want last method WaitForLaunch", calls)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("after R2, state = %q, want Running", got.Status.State)
	}
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecord after Running")
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord.Status = %q, want Running", got.Status.OpRecords[0].Status)
	}
}

// TestInit_EndToEnd_WaitLaunchFailure pins the legacy failure
// routing: real WaitForLaunch failure → InitializingCanceling.
func TestInit_EndToEnd_WaitLaunchFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedInitializing("myapp", appv1.InstallOp)
	h.workloadOps.WaitForLaunchFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{
			LandedReason: constants.AppStopDueToInitFailed,
			Err:          errors.New("wait for launch timed out"),
		}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.InitializingCanceling {
		t.Fatalf("state = %q, want InitializingCanceling", got.Status.State)
	}
	if got.Status.Reason != constants.AppStopDueToInitFailed {
		t.Fatalf("Reason = %q, want %q", got.Status.Reason, constants.AppStopDueToInitFailed)
	}
}

// TestFullLifecycle_PendingThroughRunning is the load-bearing
// chained-driver test. Extends the earlier
// TestInstallLifecycle_EndToEnd_PendingThroughInitializing by
// adding the Initializing → Running leg. All four handlers
// (pending, downloading, installing, initializing) fire in
// sequence.
func TestFullLifecycle_PendingThroughRunning(t *testing.T) {
	h := newE2EHarness(t)
	// Seed Pending; happy path requires AppExists=false at Pending.
	h.amStore.Add(newPendingAM("myapp"))

	// R1: Pending → Downloading (no CallWorkload).
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	// R2: Downloading(phase="") → dispatch pull.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	// R3: Downloading(phase=pull) + WorkloadReady → Installing.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	// The Installing pipeline: validate, helm-install, scale-up,
	// wait-startup, then the hand-off to Initializing.
	driveInstallToInitializing(t, h, "myapp")

	// The Initializing state now needs the paired Application CR.
	// (Pending was seeded without one; add it now — the legacy
	// helm install adapter would have created it. For the fake
	// harness we synthesise it here.)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

	// R7: Initializing(phase="") → dispatch WaitLaunch.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	drainAndRebuild(t, h)
	// R8: Initializing(phase=wait) + WorkloadReady → Running.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R8: %v", err)
	}

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}
	// Terminal OpRecord is the Running (install-completed) one.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after full lifecycle")
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord.Status = %q, want Running (final install success)", got.Status.OpRecords[0].Status)
	}
}
