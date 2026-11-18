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
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newResumingAM(name string, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.ResumeOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Resuming,
			OpType:  appv1.ResumeOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func resumingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

func TestOnResumingTick_EmptyPhase_StartsScale(t *testing.T) {
	am := newResumingAM("app1", nil)
	res := handlers.OnResumingTick(context.Background(), resumingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[handlers.PayloadResumingPhase]; got != handlers.PhaseResumingScale {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadResumingPhase, got, handlers.PhaseResumingScale)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadResume {
		t.Fatalf("Kind = %q, want WorkloadResume", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}
}

func TestOnResumingTick_ScalePhaseRecoversWithoutStatusEffects(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: handlers.PhaseResumingScale,
	})
	assertTickRecovery(t, handlers.OnResumingTick, resumingWorldFor(am), domain.CallWorkload{
		Kind:      domain.WorkloadResume,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmResumeFailed"},
	})
}

func TestOnResumingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: "alien",
	})
	res := handlers.OnResumingTick(context.Background(), resumingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnResumingWorkloadReady_ScalePhase_HandsOffToInitializing
// pins the key architectural point: Resuming's success DOES NOT
// go to Running directly — it advances to Initializing so the
// launch wait happens exactly once (in Initializing's
// WaitForLaunch dispatch). This matches legacy
// resuming_app.go:171.
func TestOnResumingWorkloadReady_ScalePhase_HandsOffToInitializing(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: handlers.PhaseResumingScale,
	})
	res := handlers.OnResumingWorkloadReady(context.Background(), resumingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Initializing {
		t.Fatalf("Next = %q, want Initializing (Resuming hands off to Initializing, not Running)", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Initializing {
		t.Fatalf("PatchAM.State = %q, want Initializing", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadResumingPhase]; ok {
		t.Fatal("Payload[resuming.phase] still set after leaving Resuming; must be cleared")
	}
	// NO OpRecord at the Resuming exit — matches legacy: the
	// terminal record for "resume completed" lands when
	// Initializing reaches Running.
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on Resuming→Initializing (legacy defers to Initializing→Running)")
	}
}

func TestOnResumingWorkloadReady_PendingPod_HandsOffToStopping(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: handlers.PhaseResumingScale,
	})
	am.Annotations = map[string]string{
		handlers.AnnotPendingPod: "true",
	}
	res := handlers.OnResumingWorkloadReady(context.Background(), resumingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopping {
		t.Fatalf("Next = %q, want Stopping", res.Next)
	}
}

func TestOnResumingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newResumingAM("app1", nil)
	res := handlers.OnResumingWorkloadReady(context.Background(), resumingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

func TestOnResumingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: handlers.PhaseResumingScale,
	})
	w := resumingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "HelmResumeFailed",
		Message: "scale-up rpc failed",
	}

	res := handlers.OnResumingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.ResumeFailed {
		t.Fatalf("Next = %q, want ResumeFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "HelmResumeFailed" {
		t.Fatalf("Reason = %q, want HelmResumeFailed", patch.Status.Reason)
	}
	if patch.Status.Message != "scale-up rpc failed" {
		t.Fatalf("Message = %q", patch.Status.Message)
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on failure")
	}
	if ev, ok := findEffect[domain.EmitEvent](res.Effects); !ok || ev.Type != "Warning" {
		t.Fatalf("EmitEvent = %+v ok=%v, want Warning", ev, ok)
	}
}

func TestOnResumingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newResumingAM("app1", map[string]string{
		handlers.PayloadResumingPhase: handlers.PhaseResumingScale,
	})
	w := resumingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnResumingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != string(appv1.ResumeFailed) {
		t.Fatalf("fallback Reason = %q, want %q", patch.Status.Reason, string(appv1.ResumeFailed))
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

func (h *e2eHarness) seedResuming(name string) string {
	am := newResumingAM(name, nil)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestResuming_EndToEnd_HappyPath drives Resuming → Initializing:
//
//	R1: Resuming(phase="")       → dispatch WorkloadResume; stamp phase=scale
//	R2: Resuming(phase=scale)    → WorkloadReady → Initializing (no OpRecord here)
func TestResuming_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedResuming("myapp")

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadResumingPhase]; p != handlers.PhaseResumingScale {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseResumingScale)
	}
	calls := h.workloadOps.Calls()
	if len(calls) != 1 || calls[0].Method != "Resume" {
		t.Fatalf("after R1, calls = %+v, want [Resume]", calls)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("state after R2 = %q, want Initializing", got.Status.State)
	}
	// No OpRecord at the resume-scale exit.
	if len(got.Status.OpRecords) != 0 {
		t.Fatalf("OpRecords = %+v, want empty (terminal record lands at Initializing→Running)", got.Status.OpRecords)
	}
}

// TestResuming_EndToEnd_HelmFailure pins the failure path.
func TestResuming_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedResuming("myapp")
	h.workloadOps.ResumeFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("scale-up failed")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ResumeFailed {
		t.Fatalf("state = %q, want ResumeFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError", got.Status.Reason)
	}
}

// TestFullLifecycle_StopResumeLoop is the load-bearing test for
// the stop/resume subsystem. Starting from Running, it drives:
//
//	Running → (StopOp) → Stopping → Stopped → (ResumeOp) → Resuming
//	  → Initializing → Running
//
// Eight handlers fire across seven Reconciles. Exercises the full
// TriggerOp routing table on both Running and Stopped, plus the
// two-phase Stopping and Resuming workload chains.
func TestFullLifecycle_StopResumeLoop(t *testing.T) {
	h := newE2EHarness(t)
	// Seed Running (via runningTest helpers).
	h.seedRunning("myapp", appv1.App, true)

	// R1: Running + TriggerOp{StopOp} → Stopping
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.StopOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("after R1, state = %q, want Stopping", got.Status.State)
	}

	// R2: Stopping(phase="") → dispatch WorkloadStop
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	// R3: Stopping(phase=scale) + WorkloadReady → Stopped
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("after R3, state = %q, want Stopped", got.Status.State)
	}
	if got.Status.OpRecords[0].Status != appv1.Stopped {
		t.Fatalf("after R3, head OpRecord = %q, want Stopped", got.Status.OpRecords[0].Status)
	}

	// R4: Stopped + TriggerOp{ResumeOp} → Resuming
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.ResumeOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Resuming {
		t.Fatalf("after R4, state = %q, want Resuming", got.Status.State)
	}

	// R5: Resuming(phase="") → dispatch WorkloadResume
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	drainAndRebuild(t, h)
	// R6: Resuming(phase=scale) + WorkloadReady → Initializing
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R6: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after R6, state = %q, want Initializing", got.Status.State)
	}

	// R7: Initializing(phase="") → dispatch WorkloadWaitLaunch
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	drainAndRebuild(t, h)
	// R8: Initializing(phase=wait) + WorkloadReady → Running
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R8: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}

	// Verify OpRecord chain: head = Running (from
	// Initializing's terminal write on resume completion),
	// second = Stopped (from the earlier stop).
	if len(got.Status.OpRecords) < 2 {
		t.Fatalf("OpRecords len = %d, want >= 2 (stop + resume-final)", len(got.Status.OpRecords))
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord = %q, want Running (from Initializing→Running on resume)", got.Status.OpRecords[0].Status)
	}
	if got.Status.OpRecords[1].Status != appv1.Stopped {
		t.Fatalf("OpRecords[1] = %q, want Stopped (from the earlier Stopping→Stopped)", got.Status.OpRecords[1].Status)
	}
}
