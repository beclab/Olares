package handlers_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

// newRunningAM builds an AM in Running state with configurable
// Type (App vs Middleware) — the self-heal branch of OnRunningTick
// only fires for App-typed AMs.
func newRunningAM(name string, appType appv1.Type) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appType,
			OpType:       appv1.InstallOp, // set once at install
		},
		Status: appv1.ApplicationManagerStatus{
			State:  appv1.Running,
			OpType: appv1.InstallOp,
			OpID:   "op-42",
		},
	}
}

func runningWorldFor(am *appv1.ApplicationManager, appExists bool) domain.World {
	return domain.World{
		AM:        am,
		Now:       time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		AppExists: appExists,
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

// TestOnRunningTick_HealthyApp_IsNoop pins the steady-state path:
// AppExists=true → empty Result, no effects.
func TestOnRunningTick_HealthyApp_IsNoop(t *testing.T) {
	am := newRunningAM("app1", appv1.App)
	res := handlers.OnRunningTick(context.Background(), runningWorldFor(am, true))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (steady state)", res.Next)
	}
	if len(res.Effects) != 0 {
		t.Fatalf("Effects = %+v, want none", res.Effects)
	}
}

// TestOnRunningTick_MissingApp_SelfHeals pins the self-heal path:
// AppExists=false + Type=App → advance to Uninstalling, so helm
// uninstall and the namespace wait run before Uninstalled.
func TestOnRunningTick_MissingApp_SelfHeals(t *testing.T) {
	am := newRunningAM("app1", appv1.App)
	res := handlers.OnRunningTick(context.Background(), runningWorldFor(am, false))
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil", res.Err)
	}
	if res.Next != appv1.Uninstalling {
		t.Fatalf("Next = %q, want Uninstalling", res.Next)
	}

	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM effect")
	}
	if patch.Status.State != appv1.Uninstalling {
		t.Fatalf("PatchAM.State = %q, want Uninstalling", patch.Status.State)
	}
	if patch.Status.Reason != "AppMissingSelfHeal" {
		t.Fatalf("PatchAM.Reason = %q, want AppMissingSelfHeal", patch.Status.Reason)
	}
	if _, ok := findEffect[domain.CallWorkload](res.Effects); ok {
		t.Fatal("self-heal dispatched a workload; Uninstalling's Tick owns helm uninstall")
	}

	ev, ok := findEffect[domain.EmitEvent](res.Effects)
	if !ok {
		t.Fatal("no EmitEvent effect")
	}
	if ev.Type != "Warning" {
		t.Fatalf("EmitEvent.Type = %q, want Warning (self-heal is anomalous)", ev.Type)
	}
}

// TestOnRunningTick_MissingMiddleware_DoesNotSelfHeal pins the
// legacy carve-out: middleware-typed AMs whose Application CR is
// missing do NOT self-heal — a different subsystem manages them.
func TestOnRunningTick_MissingMiddleware_DoesNotSelfHeal(t *testing.T) {
	am := newRunningAM("mw1", appv1.Middleware)
	res := handlers.OnRunningTick(context.Background(), runningWorldFor(am, false))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (middleware self-heal not ported)", res.Next)
	}
	if len(res.Effects) != 0 {
		t.Fatalf("Effects = %+v, want none for middleware missing case", res.Effects)
	}
}

// TestOnRunningTick_NilAM_IsSafeNoop pins the defensive guard:
// a bogus World with nil AM must NOT panic.
func TestOnRunningTick_NilAM_IsSafeNoop(t *testing.T) {
	w := domain.World{AM: nil, AppExists: false}
	res := handlers.OnRunningTick(context.Background(), w)
	if res.Err != nil {
		t.Fatalf("Err = %v, want nil (nil AM must not error)", res.Err)
	}
	if res.Next != "" || len(res.Effects) != 0 {
		t.Fatalf("Result = %+v, want zero for nil AM", res)
	}
}

// TestOnRunningOp_Transitions pins each of the four op-driven
// exits: (op, expectedTarget, expectedReason).
func TestOnRunningOp_Transitions(t *testing.T) {
	cases := []struct {
		name    string
		handler func(context.Context, domain.World) domain.Result
		target  appv1.ApplicationManagerState
		reason  string
	}{
		{"stop", handlers.OnRunningStop, appv1.Stopping, "StopRequested"},
		{"uninstall", handlers.OnRunningUninstall, appv1.Uninstalling, "UninstallRequested"},
		{"upgrade", handlers.OnRunningUpgrade, appv1.UpgradePending, "UpgradeRequested"},
		{"applyenv", handlers.OnRunningApplyEnv, appv1.ApplyingEnv, "ApplyEnvRequested"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			am := newRunningAM("app1", appv1.App)
			res := tc.handler(context.Background(), runningWorldFor(am, true))
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
				t.Fatalf("EmitEvent.Type = %q, want Normal (op-driven transition)", ev.Type)
			}

			// No OpRecord at Running-exit — legacy defers the
			// terminal record to the target state's handler.
			if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
				t.Fatalf("unexpected AppendOpRecord on Running→%q (legacy writes none here)", tc.target)
			}
		})
	}
}

// TestStopHandlers_CarryAnnotatedAbnormalReason pins the pod-abnormal
// path: the controller stamps the eviction or unschedulable reason
// before Trigger, and every StopOp entry that this controller can
// legally fire copies it onto Status and the event, then clears the
// annotations so a later user stop is not stuck with them.
func TestStopHandlers_CarryAnnotatedAbnormalReason(t *testing.T) {
	handlersUnderTest := []struct {
		name    string
		handler func(context.Context, domain.World) domain.Result
	}{
		{"running", handlers.OnRunningStop},
		{"applyingenv", handlers.OnApplyingEnvStop},
		{"resuming", handlers.OnResumingStop},
		{"stopfailed", handlers.OnStopFailedStop},
		{"applyenvfailed", handlers.OnApplyEnvFailedStop},
	}
	for _, tc := range handlersUnderTest {
		t.Run(tc.name, func(t *testing.T) {
			am := newRunningAM("app1", appv1.App)
			am.Annotations = map[string]string{
				api.AppStopReasonKey:  "Evicted",
				api.AppStopMessageKey: "evicted pod: ns/pod",
				api.AppStopAllKey:     "true",
			}
			res := tc.handler(context.Background(), runningWorldFor(am, true))
			if res.Err != nil {
				t.Fatalf("Err = %v", res.Err)
			}
			if res.Next != appv1.Stopping {
				t.Fatalf("Next = %q, want Stopping", res.Next)
			}
			if len(res.Effects) == 0 {
				t.Fatal("no effects")
			}
			spec, ok := res.Effects[0].(domain.PatchSpec)
			if !ok {
				t.Fatalf("effects[0] = %T, want PatchSpec clearing the carried reason", res.Effects[0])
			}
			if len(spec.DeleteAnnotations) != 2 ||
				spec.DeleteAnnotations[0] != api.AppStopReasonKey ||
				spec.DeleteAnnotations[1] != api.AppStopMessageKey {
				t.Fatalf("DeleteAnnotations = %v, want reason and message keys only", spec.DeleteAnnotations)
			}
			patch, ok := findEffect[domain.PatchAM](res.Effects)
			if !ok {
				t.Fatal("no PatchAM")
			}
			if patch.Status.Reason != "Evicted" || patch.Status.Message != "evicted pod: ns/pod" {
				t.Fatalf("PatchAM reason/message = %q / %q", patch.Status.Reason, patch.Status.Message)
			}
			ev, ok := findEffect[domain.EmitEvent](res.Effects)
			if !ok {
				t.Fatal("no EmitEvent")
			}
			if ev.Reason != "Evicted" || ev.Message != "evicted pod: ns/pod" {
				t.Fatalf("EmitEvent reason/message = %q / %q", ev.Reason, ev.Message)
			}
		})
	}
}

// TestOnRunningOp_NilAM_HardFails pins the defensive Err path for
// the shared helper. Nil-AM here is a genuine bug (Runner always
// populates it) — surface it instead of silently no-op'ing.
func TestOnRunningOp_NilAM_HardFails(t *testing.T) {
	w := domain.World{AM: nil}
	res := handlers.OnRunningStop(context.Background(), w)
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for nil-AM invocation")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

// seedRunning seeds a Running AM. appExists controls whether the
// paired Application CR is also present in the AMStore.
func (h *e2eHarness) seedRunning(name string, appType appv1.Type, appExists bool) string {
	am := newRunningAM(name, appType)
	if appExists {
		h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	}
	return h.amStore.Add(am)
}

// TestRunning_EndToEnd_HealthyTick pins the steady-state no-op:
// a healthy Running AM's periodic tick must not mutate state.
func TestRunning_EndToEnd_HealthyTick(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, true)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("state = %q, want Running (steady state must not mutate)", got.Status.State)
	}
}

// TestRunning_EndToEnd_MissingAppHeals pins the self-heal
// end-to-end: seed Running with NO paired Application → Tick →
// Uninstalling → helm uninstall → namespace wait → Uninstalled.
func TestRunning_EndToEnd_MissingAppHeals(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, false) // appExists=false

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalling {
		t.Fatalf("after R1, state = %q, want Uninstalling (self-heal on missing App CR)", got.Status.State)
	}

	// R2 dispatches helm uninstall, R3 the namespace wait, R4 lands.
	for i := 2; i <= 4; i++ {
		if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
			t.Fatalf("R%d: %v", i, err)
		}
		drainAndRebuild(t, h)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R5: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalled {
		t.Fatalf("final state = %q, want Uninstalled", got.Status.State)
	}
	methods := map[string]bool{}
	for _, c := range h.workloadOps.Calls() {
		methods[c.Method] = true
	}
	if !methods["Uninstall"] || !methods["WaitForNamespaceDeleted"] {
		t.Fatalf("workload calls = %v, want Uninstall and WaitForNamespaceDeleted", methods)
	}
}

// TestRunning_EndToEnd_TriggerUninstall pins the op-driven exit:
// deliver TriggerOp{UninstallOp} → Reconcile → state = Uninstalling.
func TestRunning_EndToEnd_TriggerUninstall(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, true)

	// Simulate the apiserver: deliver TriggerOp into the pending
	// queue (matches how the collapsed apiserver-controller path
	// will feed user-initiated ops into the state machine).
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.UninstallOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Uninstalling {
		t.Fatalf("state = %q, want Uninstalling", got.Status.State)
	}
	if got.Status.Reason != "UninstallRequested" {
		t.Fatalf("Reason = %q, want UninstallRequested", got.Status.Reason)
	}
}

// TestRunning_EndToEnd_TriggerStop pins the stop path for the
// same reason — different target state, same code path shape.
// Together with TriggerUninstall this locks that the four-way
// TriggerOp dispatch (keyed by Op:<opName>) works correctly.
func TestRunning_EndToEnd_TriggerStop(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, true)
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.StopOp})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopping {
		t.Fatalf("state = %q, want Stopping", got.Status.State)
	}
}
