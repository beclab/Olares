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

func newApplyingEnvAM(name string, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.ApplyEnvOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.ApplyingEnv,
			OpType:  appv1.ApplyEnvOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func applyingEnvWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

func TestOnApplyingEnvTick_EmptyPhase_StartsHelm(t *testing.T) {
	am := newApplyingEnvAM("app1", nil)
	res := handlers.OnApplyingEnvTick(context.Background(), applyingEnvWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[handlers.PayloadApplyingEnvPhase]; got != handlers.PhaseApplyingEnvHelm {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadApplyingEnvPhase, got, handlers.PhaseApplyingEnvHelm)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadApplyEnv {
		t.Fatalf("Kind = %q, want WorkloadApplyEnv", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}
}

func TestOnApplyingEnvTick_HelmPhaseRecoversWithoutStatusEffects(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm,
	})
	assertTickRecovery(t, handlers.OnApplyingEnvTick, applyingEnvWorldFor(am), domain.CallWorkload{
		Kind:      domain.WorkloadApplyEnv,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmApplyEnvFailed"},
	})
}

func TestOnApplyingEnvTick_UnknownPhase_HardFails(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: "alien",
	})
	res := handlers.OnApplyingEnvTick(context.Background(), applyingEnvWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

func TestOnApplyingEnvWorkloadReady_HelmPhase_HandsOffToInitializing(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm,
	})
	res := handlers.OnApplyingEnvWorkloadReady(context.Background(), applyingEnvWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Initializing {
		t.Fatalf("Next = %q, want Initializing", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Initializing {
		t.Fatalf("PatchAM.State = %q", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadApplyingEnvPhase]; ok {
		t.Fatal("Payload[applyingenv.phase] still set after leaving ApplyingEnv")
	}
	// No OpRecord — matches legacy.
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on ApplyingEnv→Initializing")
	}
}

func TestOnApplyingEnvWorkloadReady_FromStopped_LandsStopped(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm,
	})
	am.Annotations = map[string]string{
		handlers.AnnotPreUpgradeState: string(appv1.Stopped),
	}
	res := handlers.OnApplyingEnvWorkloadReady(context.Background(), applyingEnvWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopped {
		t.Fatalf("Next = %q, want Stopped", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Stopped {
		t.Fatalf("PatchAM.State = %q, want Stopped", patch.Status.State)
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on apply-env-from-Stopped")
	}
}

func TestOnApplyingEnvWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newApplyingEnvAM("app1", nil)
	res := handlers.OnApplyingEnvWorkloadReady(context.Background(), applyingEnvWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

func TestOnApplyingEnvWorkloadFailed_UsesActionData(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm,
	})
	w := applyingEnvWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "HelmApplyEnvFailed",
		Message: "env update rpc failed",
	}

	res := handlers.OnApplyingEnvWorkloadFailed(context.Background(), w)
	if res.Next != appv1.ApplyEnvFailed {
		t.Fatalf("Next = %q, want ApplyEnvFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "HelmApplyEnvFailed" {
		t.Fatalf("Reason = %q", patch.Status.Reason)
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on failure")
	}
	if ev, ok := findEffect[domain.EmitEvent](res.Effects); !ok || ev.Type != "Warning" {
		t.Fatalf("EmitEvent = %+v ok=%v, want Warning", ev, ok)
	}
}

func TestOnApplyingEnvWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newApplyingEnvAM("app1", map[string]string{
		handlers.PayloadApplyingEnvPhase: handlers.PhaseApplyingEnvHelm,
	})
	w := applyingEnvWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnApplyingEnvWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != string(appv1.ApplyEnvFailed) {
		t.Fatalf("fallback Reason = %q, want %q", patch.Status.Reason, string(appv1.ApplyEnvFailed))
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

func (h *e2eHarness) seedApplyingEnv(name string) string {
	am := newApplyingEnvAM(name, nil)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

func TestApplyingEnv_EndToEnd_FromStopped_LandsStopped(t *testing.T) {
	h := newE2EHarness(t)
	am := newApplyingEnvAM("myapp", nil)
	am.Annotations = map[string]string{
		handlers.AnnotPreUpgradeState: string(appv1.Stopped),
	}
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("state after R2 = %q, want Stopped", got.Status.State)
	}
}

// TestApplyingEnv_EndToEnd_HappyPath drives ApplyingEnv → Initializing.
func TestApplyingEnv_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedApplyingEnv("myapp")

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	calls := h.workloadOps.Calls()
	if len(calls) != 1 || calls[0].Method != "ApplyEnv" {
		t.Fatalf("after R1, calls = %+v, want [ApplyEnv]", calls)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("state after R2 = %q, want Initializing", got.Status.State)
	}
}

// TestApplyingEnv_EndToEnd_HelmFailure pins the failure path.
func TestApplyingEnv_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedApplyingEnv("myapp")
	h.workloadOps.ApplyEnvFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("env update failed")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ApplyEnvFailed {
		t.Fatalf("state = %q, want ApplyEnvFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError", got.Status.Reason)
	}
}

// TestFullLifecycle_ApplyEnvCycle proves the Running → ApplyingEnv
// → Initializing → Running loop. Symmetric to
// TestFullLifecycle_UpgradeCycle.
func TestFullLifecycle_ApplyEnvCycle(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, true)

	// R1: Running + TriggerOp{ApplyEnvOp} → ApplyingEnv
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.ApplyEnvOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.ApplyingEnv {
		t.Fatalf("after R1, state = %q, want ApplyingEnv", got.Status.State)
	}

	// R2: ApplyingEnv(phase="") → dispatch WorkloadApplyEnv
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	// R3: ApplyingEnv(phase=helm) + WorkloadReady → Initializing
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after R3, state = %q, want Initializing", got.Status.State)
	}

	// R4: Initializing(phase="") → dispatch WaitLaunch
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	drainAndRebuild(t, h)
	// R5: Initializing(phase=wait) + WorkloadReady → Running
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after apply-env cycle")
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord.Status = %q, want Running", got.Status.OpRecords[0].Status)
	}
}
