package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newStoppedAM(name string) *appv1.ApplicationManager {
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
			State:  appv1.Stopped,
			OpType: appv1.StopOp,
			OpID:   "op-42",
		},
	}
}

func stoppedWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnStoppedTick_IsAlwaysNoop pins the steady-state no-op.
func TestOnStoppedTick_IsAlwaysNoop(t *testing.T) {
	am := newStoppedAM("app1")
	res := handlers.OnStoppedTick(context.Background(), stoppedWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" || len(res.Effects) != 0 {
		t.Fatalf("Result = %+v, want zero (steady state)", res)
	}
}

// TestOnStoppedTick_NilAM_IsSafeNoop pins the defensive guard.
func TestOnStoppedTick_NilAM_IsSafeNoop(t *testing.T) {
	res := handlers.OnStoppedTick(context.Background(), domain.World{AM: nil})
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil (nil AM must not error)", res.Err)
	}
}

// TestOnStoppedOp_Transitions pins each of the four op-driven
// exits — same shape as TestOnRunningOp_Transitions.
func TestOnStoppedOp_Transitions(t *testing.T) {
	cases := []struct {
		name    string
		handler func(context.Context, domain.World) domain.Result
		target  appv1.ApplicationManagerState
		reason  string
	}{
		{"resume", handlers.OnStoppedResume, appv1.Resuming, "ResumeRequested"},
		{"uninstall", handlers.OnStoppedUninstall, appv1.Uninstalling, "UninstallRequested"},
		{"upgrade", handlers.OnStoppedUpgrade, appv1.UpgradePending, "UpgradeRequested"},
		{"applyenv", handlers.OnStoppedApplyEnv, appv1.ApplyingEnv, "ApplyEnvRequested"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			am := newStoppedAM("app1")
			res := tc.handler(context.Background(), stoppedWorldFor(am))
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != tc.target {
				t.Fatalf("Next = %q, want %q", res.Next, tc.target)
			}
			patch, _ := findEffect[domain.PatchAM](res.Effects)
			if patch.Status.State != tc.target {
				t.Fatalf("PatchAM.State = %q, want %q", patch.Status.State, tc.target)
			}
			if patch.Status.Reason != tc.reason {
				t.Fatalf("PatchAM.Reason = %q, want %q", patch.Status.Reason, tc.reason)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok {
				t.Fatal("no EmitEvent")
			}
			if ev.Type != "Normal" {
				t.Fatalf("EmitEvent.Type = %q, want Normal (op-driven)", ev.Type)
			}

			// No OpRecord — matches Running's exit convention.
			if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
				t.Fatalf("unexpected AppendOpRecord on Stopped→%q", tc.target)
			}
		})
	}
}

// TestOnStoppedOp_NilAM_HardFails pins the defensive Err path
// for the shared helper.
func TestOnStoppedOp_NilAM_HardFails(t *testing.T) {
	res := handlers.OnStoppedResume(context.Background(), domain.World{AM: nil})
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for nil-AM invocation")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

func (h *e2eHarness) seedStopped(name string) string {
	am := newStoppedAM(name)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

// TestStopped_EndToEnd_TriggerResume pins the resume trigger:
// deliver TriggerOp{ResumeOp} → Reconcile → state = Resuming.
func TestStopped_EndToEnd_TriggerResume(t *testing.T) {
	h := newE2EHarness(t)
	h.seedStopped("myapp")

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.ResumeOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Resuming {
		t.Fatalf("state = %q, want Resuming", got.Status.State)
	}
	if got.Status.Reason != "ResumeRequested" {
		t.Fatalf("Reason = %q, want ResumeRequested", got.Status.Reason)
	}
}

// TestStopped_EndToEnd_TriggerUninstall — sibling coverage for
// the four-way TriggerOp routing.
func TestStopped_EndToEnd_TriggerUninstall(t *testing.T) {
	h := newE2EHarness(t)
	h.seedStopped("myapp")

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.UninstallOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalling {
		t.Fatalf("state = %q, want Uninstalling", got.Status.State)
	}
}

// TestStopped_EndToEnd_HealthyTick pins the steady-state no-op
// end-to-end (watch-driven Reconciles hit this path).
func TestStopped_EndToEnd_HealthyTick(t *testing.T) {
	h := newE2EHarness(t)
	h.seedStopped("myapp")

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("state = %q, want Stopped (steady state must not mutate)", got.Status.State)
	}
}
