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

func newUninstallingAM(name string, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.UninstallOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Uninstalling,
			OpType:  appv1.UninstallOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func uninstallingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnUninstallingTick_EmptyPhase_StartsHelm pins the fresh-
// entry branch: dispatch CallWorkload{Uninstall} + stamp phase=helm.
func TestOnUninstallingTick_EmptyPhase_StartsHelm(t *testing.T) {
	am := newUninstallingAM("app1", nil)
	res := handlers.OnUninstallingTick(context.Background(), uninstallingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay in Uninstalling)", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[handlers.PayloadUninstallingPhase]; got != handlers.PhaseUninstallingHelm {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadUninstallingPhase, got, handlers.PhaseUninstallingHelm)
	}

	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadUninstall {
		t.Fatalf("Kind = %q, want WorkloadUninstall", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}
}

func TestOnUninstallingTick_KnownPhasesRecoverWithoutStatusEffects(t *testing.T) {
	tests := []struct {
		phase string
		call  domain.CallWorkload
	}{
		{
			phase: handlers.PhaseUninstallingHelm,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadUninstall,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
			},
		},
		{
			phase: handlers.PhaseUninstallingWaitNS,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitNamespaceGone,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			am := newUninstallingAM("app1", map[string]string{
				handlers.PayloadUninstallingPhase: tc.phase,
			})
			assertTickRecovery(t, handlers.OnUninstallingTick, uninstallingWorldFor(am), tc.call)
		})
	}
}

// TestOnUninstallingTick_UnknownPhase_HardFails — same safety net
// as siblings.
func TestOnUninstallingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newUninstallingAM("app1", map[string]string{
		handlers.PayloadUninstallingPhase: "grabbed-by-alien",
	})
	res := handlers.OnUninstallingTick(context.Background(), uninstallingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnUninstallingWorkloadReady_HelmPhase_StartsWaitNS pins the
// phase-1 → phase-2 advance.
func TestOnUninstallingWorkloadReady_HelmPhase_StartsWaitNS(t *testing.T) {
	am := newUninstallingAM("app1", map[string]string{
		handlers.PayloadUninstallingPhase: handlers.PhaseUninstallingHelm,
	})
	res := handlers.OnUninstallingWorkloadReady(context.Background(), uninstallingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadUninstallingPhase]; got != handlers.PhaseUninstallingWaitNS {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadUninstallingPhase, got, handlers.PhaseUninstallingWaitNS)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadWaitNamespaceGone {
		t.Fatalf("Kind = %q, want WorkloadWaitNamespaceGone", cw.Kind)
	}
}

// TestOnUninstallingWorkloadReady_WaitNSPhase_Finishes pins the
// terminal branch → Uninstalled with OpRecord.
func TestOnUninstallingWorkloadReady_WaitNSPhase_Finishes(t *testing.T) {
	am := newUninstallingAM("app1", map[string]string{
		handlers.PayloadUninstallingPhase: handlers.PhaseUninstallingWaitNS,
	})
	res := handlers.OnUninstallingWorkloadReady(context.Background(), uninstallingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Uninstalled {
		t.Fatalf("Next = %q, want Uninstalled", res.Next)
	}

	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Uninstalled {
		t.Fatalf("PatchAM.State = %q, want Uninstalled", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadUninstallingPhase]; ok {
		t.Fatal("Payload[uninstalling.phase] still set after leaving Uninstalling; must be cleared")
	}

	rec, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord")
	}
	if rec.Record.Status != appv1.Uninstalled {
		t.Fatalf("OpRecord.Status = %q, want Uninstalled", rec.Record.Status)
	}
	// Ordering — AppendOpRecord must precede PatchAM.
	if _, ok := res.Effects[0].(domain.AppendOpRecord); !ok {
		t.Fatalf("Effects[0] = %T, want AppendOpRecord", res.Effects[0])
	}
}

// TestOnUninstallingWorkloadReady_EmptyPhase_HardFails pins the
// dispatch invariant.
func TestOnUninstallingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newUninstallingAM("app1", nil)
	res := handlers.OnUninstallingWorkloadReady(context.Background(), uninstallingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

// TestOnUninstallingWorkloadFailed_UsesActionData pins that
// failure at either phase → UninstallFailed with Action's Reason/
// Message preserved.
func TestOnUninstallingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newUninstallingAM("app1", map[string]string{
		handlers.PayloadUninstallingPhase: handlers.PhaseUninstallingHelm,
	})
	w := uninstallingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "HelmUninstallFailed",
		Message: "chart delete rpc failed",
	}

	res := handlers.OnUninstallingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.UninstallFailed {
		t.Fatalf("Next = %q, want UninstallFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "HelmUninstallFailed" {
		t.Fatalf("Reason = %q, want HelmUninstallFailed", patch.Status.Reason)
	}
	if patch.Status.Message != "chart delete rpc failed" {
		t.Fatalf("Message = %q, want the Action's Message", patch.Status.Message)
	}
	// OpRecord + Warning event emitted.
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on failure")
	}
	if ev, ok := findEffect[domain.EmitEvent](res.Effects); !ok || ev.Type != "Warning" {
		t.Fatalf("EmitEvent = %+v ok=%v, want Warning", ev, ok)
	}
}

// TestOnUninstallingWorkloadFailed_FallsBackWhenActionEmpty pins
// the phase-derived fallback strings.
func TestOnUninstallingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newUninstallingAM("app1", map[string]string{
		handlers.PayloadUninstallingPhase: handlers.PhaseUninstallingWaitNS,
	})
	w := uninstallingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnUninstallingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "NamespaceGoneWaitFailed" {
		t.Fatalf("fallback Reason = %q, want NamespaceGoneWaitFailed", patch.Status.Reason)
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// seedUninstalling seeds an Uninstalling AM with empty phase.
func (h *e2eHarness) seedUninstalling(name string) string {
	am := newUninstallingAM(name, nil)
	// Paired Application still exists — helm hasn't run yet.
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestUninstall_EndToEnd_HappyPath drives Uninstalling → Uninstalled:
//
//	R1: Uninstalling(phase="")      → dispatch WorkloadUninstall
//	R2: Uninstalling(phase=helm)    → WorkloadReady → dispatch WaitNS
//	R3: Uninstalling(phase=wait-ns) → WorkloadReady → Uninstalled
func TestUninstall_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUninstalling("myapp")

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadUninstallingPhase]; p != handlers.PhaseUninstallingHelm {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseUninstallingHelm)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ = h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadUninstallingPhase]; p != handlers.PhaseUninstallingWaitNS {
		t.Fatalf("after R2, phase = %q, want %q", p, handlers.PhaseUninstallingWaitNS)
	}
	// Both workload calls in order: Uninstall then WaitForNamespaceDeleted.
	calls := h.workloadOps.Calls()
	if len(calls) < 2 {
		t.Fatalf("workload calls = %+v, want at least 2", calls)
	}
	if calls[0].Method != "Uninstall" {
		t.Fatalf("calls[0].Method = %q, want Uninstall", calls[0].Method)
	}
	if calls[len(calls)-1].Method != "WaitForNamespaceDeleted" {
		t.Fatalf("last call = %q, want WaitForNamespaceDeleted", calls[len(calls)-1].Method)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalled {
		t.Fatalf("state after R3 = %q, want Uninstalled", got.Status.State)
	}
	// Terminal OpRecord landed.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after uninstall completion")
	}
	if got.Status.OpRecords[0].Status != appv1.Uninstalled {
		t.Fatalf("head OpRecord.Status = %q, want Uninstalled", got.Status.OpRecords[0].Status)
	}
}

// TestUninstall_EndToEnd_HelmFailure pins the phase-1 failure
// path: Uninstall returns err → WorkloadFailed → UninstallFailed.
func TestUninstall_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUninstalling("myapp")
	h.workloadOps.UninstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("chart delete failed")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UninstallFailed {
		t.Fatalf("state = %q, want UninstallFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError (Runner-classified)", got.Status.Reason)
	}
}

// TestUninstall_EndToEnd_WaitNSFailure pins the phase-2 failure:
// WaitForNamespaceDeleted returns err → UninstallFailed.
func TestUninstall_EndToEnd_WaitNSFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUninstalling("myapp")
	h.workloadOps.WaitForNamespaceDeletedFunc = func(_ context.Context, _ ports.NamespaceWaitRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "NSPollTimeout", Err: errors.New("timed out after 30m")}
	}

	// Advance through phase 1 (helm ok) → phase 2 (wait-ns fails) →
	// UninstallFailed.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UninstallFailed {
		t.Fatalf("state = %q, want UninstallFailed", got.Status.State)
	}
	if got.Status.Reason != "NSPollTimeout" {
		t.Fatalf("Reason = %q, want NSPollTimeout", got.Status.Reason)
	}
}

// TestFullLifecycle_InstallThroughUninstall is the load-bearing
// end-to-end test: extends TestFullLifecycle_PendingThroughRunning
// by triggering a UninstallOp from Running and driving all the way
// to Uninstalled. Six handlers (pending, downloading, installing,
// initializing, running, uninstalling) fire in sequence.
func TestFullLifecycle_InstallThroughUninstall(t *testing.T) {
	h := newE2EHarness(t)
	// Same install-half prefix as TestFullLifecycle_Pending
	// ThroughRunning.
	h.amStore.Add(newPendingAM("myapp"))

	// R1: Pending → Downloading
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	// R2: Downloading(phase="") → dispatch pull
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	drainAndRebuild(t, h)
	// R3: Downloading(phase=pull) + WorkloadReady → Installing
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	// The Installing pipeline: validate, helm-install, scale-up,
	// wait-startup, then the hand-off to Initializing.
	driveInstallToInitializing(t, h, "myapp")

	// Synthesize paired Application (real helm adapter would create it).
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

	// R7: Initializing(phase="") → dispatch WaitLaunch
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R7: %v", err)
	}
	drainAndRebuild(t, h)
	// R8: Initializing(phase=wait) + WorkloadReady → Running
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R8: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("after R8, state = %q, want Running", got.Status.State)
	}

	// Now trigger uninstall from Running.
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.UninstallOp})

	// R9: Running + TriggerOp{UninstallOp} → Uninstalling
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R9: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalling {
		t.Fatalf("after R9, state = %q, want Uninstalling", got.Status.State)
	}

	// R10: Uninstalling(phase="") → dispatch helm uninstall
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R10: %v", err)
	}
	drainAndRebuild(t, h)
	// R11: Uninstalling(phase=helm) + WorkloadReady → dispatch wait-ns
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R11: %v", err)
	}
	drainAndRebuild(t, h)
	// R12: Uninstalling(phase=wait-ns) + WorkloadReady → Uninstalled
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R12: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalled {
		t.Fatalf("final state = %q, want Uninstalled", got.Status.State)
	}

	// Head OpRecord is the Uninstalled terminal one — it must
	// have been prepended by the Uninstalling handler.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after full lifecycle")
	}
	if got.Status.OpRecords[0].Status != appv1.Uninstalled {
		t.Fatalf("head OpRecord.Status = %q, want Uninstalled (final terminal)", got.Status.OpRecords[0].Status)
	}
	// Second-most-recent OpRecord is the earlier Running (install-complete) one.
	if len(got.Status.OpRecords) < 2 {
		t.Fatalf("OpRecords len = %d, want >= 2 (install-complete + uninstall)", len(got.Status.OpRecords))
	}
	if got.Status.OpRecords[1].Status != appv1.Running {
		t.Fatalf("OpRecords[1].Status = %q, want Running (from install lifecycle)", got.Status.OpRecords[1].Status)
	}
}
