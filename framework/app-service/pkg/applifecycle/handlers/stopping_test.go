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

func newStoppingAM(name string, payload map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.StopOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Stopping,
			OpType:  appv1.StopOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func stoppingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

func TestOnStoppingTick_EmptyPhase_StartsScale(t *testing.T) {
	am := newStoppingAM("app1", nil)
	res := handlers.OnStoppingTick(context.Background(), stoppingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty", res.Next)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[handlers.PayloadStoppingPhase]; got != handlers.PhaseStoppingScale {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadStoppingPhase, got, handlers.PhaseStoppingScale)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadStop {
		t.Fatalf("Kind = %q, want WorkloadStop", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}
}

func TestOnStoppingTick_ScalePhaseRecoversWithoutStatusEffects(t *testing.T) {
	am := newStoppingAM("app1", map[string]string{
		handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
	})
	assertTickRecovery(t, handlers.OnStoppingTick, stoppingWorldFor(am), domain.CallWorkload{
		Kind:      domain.WorkloadStop,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmStopFailed"},
		Preempts:  true,
	})
}

func TestOnStoppingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newStoppingAM("app1", map[string]string{
		handlers.PayloadStoppingPhase: "alien",
	})
	res := handlers.OnStoppingTick(context.Background(), stoppingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

func TestOnStoppingWorkloadReady_ScalePhase_Finishes(t *testing.T) {
	am := newStoppingAM("app1", map[string]string{
		handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
	})
	res := handlers.OnStoppingWorkloadReady(context.Background(), stoppingWorldFor(am))
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
	if _, ok := patch.Status.Payload[handlers.PayloadStoppingPhase]; ok {
		t.Fatal("Payload[stopping.phase] still set after leaving Stopping; must be cleared")
	}
	rec, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("no AppendOpRecord")
	}
	if rec.Record.Status != appv1.Stopped {
		t.Fatalf("OpRecord.Status = %q, want Stopped", rec.Record.Status)
	}
	// Ordering — AppendOpRecord must precede PatchAM.
	if _, ok := res.Effects[0].(domain.AppendOpRecord); !ok {
		t.Fatalf("Effects[0] = %T, want AppendOpRecord", res.Effects[0])
	}
}

func TestOnStoppingWorkloadReady_PublishesLegacyStoppedReason(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"StopByUser", "StopByUser"},
		{"StopRequested", "StopByUser"},
		{"Evicted", "Evicted"},
		{"InitFailed", "InitFailed"},
		{"HamiUnschedulable", "HamiUnschedulable"},
		{"Unschedulable", "Unschedulable"},
		{"DiskPressure", "DiskPressure"},
		{"SystemCPUPressure", "SystemCPUPressure"},
		{"SystemMemoryPressure", "SystemMemoryPressure"},
		{"K8sReqeustCPUPressure", "K8sRequestCPUPressure"},
		{"K8sRequestCPUPressure", "K8sRequestCPUPressure"},
		{"K8sRequestMemoryPressure", "K8sRequestMemoryPressure"},
		{"StartUpFailed", "StartUpFailed"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			am := newStoppingAM("app1", map[string]string{
				handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
			})
			am.Status.Reason = tc.in
			res := handlers.OnStoppingWorkloadReady(context.Background(), stoppingWorldFor(am))
			patch, _ := findEffect[domain.PatchAM](res.Effects)
			if patch.Status.Reason != tc.want {
				t.Fatalf("status Reason = %q, want %q", patch.Status.Reason, tc.want)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok || ev.Reason != tc.want {
				t.Fatalf("event Reason = %q, want %q", ev.Reason, tc.want)
			}
		})
	}
}

func TestOnStoppingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newStoppingAM("app1", nil)
	res := handlers.OnStoppingWorkloadReady(context.Background(), stoppingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

func TestOnStoppingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newStoppingAM("app1", map[string]string{
		handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
	})
	w := stoppingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "HelmStopFailed",
		Message: "scale rpc failed",
	}

	res := handlers.OnStoppingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.StopFailed {
		t.Fatalf("Next = %q, want StopFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "HelmStopFailed" {
		t.Fatalf("Reason = %q, want HelmStopFailed", patch.Status.Reason)
	}
	if patch.Status.Message != "scale rpc failed" {
		t.Fatalf("Message = %q, want the Action's Message", patch.Status.Message)
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on failure")
	}
	if ev, ok := findEffect[domain.EmitEvent](res.Effects); !ok || ev.Type != "Warning" {
		t.Fatalf("EmitEvent = %+v ok=%v, want Warning", ev, ok)
	}
}

func TestOnStoppingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newStoppingAM("app1", map[string]string{
		handlers.PayloadStoppingPhase: handlers.PhaseStoppingScale,
	})
	w := stoppingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnStoppingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != string(appv1.StopFailed) {
		t.Fatalf("fallback Reason = %q, want %q", patch.Status.Reason, string(appv1.StopFailed))
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

func (h *e2eHarness) seedStopping(name string) string {
	am := newStoppingAM(name, nil)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestStopping_EndToEnd_HappyPath drives Stopping → Stopped:
//
//	R1: Stopping(phase="")      → dispatch WorkloadStop; stamp phase=scale
//	R2: Stopping(phase=scale)   → WorkloadReady → Stopped (+ terminal OpRecord)
func TestStopping_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedStopping("myapp")

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)

	got, _ := h.amStore.Get(context.Background(), "myapp")
	if p := got.Status.Payload[handlers.PayloadStoppingPhase]; p != handlers.PhaseStoppingScale {
		t.Fatalf("after R1, phase = %q, want %q", p, handlers.PhaseStoppingScale)
	}
	calls := h.workloadOps.Calls()
	if len(calls) != 1 || calls[0].Method != "Stop" {
		t.Fatalf("after R1, calls = %+v, want [Stop]", calls)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("state after R2 = %q, want Stopped", got.Status.State)
	}
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after stop completion")
	}
	if got.Status.OpRecords[0].Status != appv1.Stopped {
		t.Fatalf("head OpRecord.Status = %q, want Stopped", got.Status.OpRecords[0].Status)
	}
}

// TestStopping_EndToEnd_HelmFailure pins the failure path.
func TestStopping_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedStopping("myapp")
	h.workloadOps.StopFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("scale failed")}
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.StopFailed {
		t.Fatalf("state = %q, want StopFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError (Runner-classified)", got.Status.Reason)
	}
}
