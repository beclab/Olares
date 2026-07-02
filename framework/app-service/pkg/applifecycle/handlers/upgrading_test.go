package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// -----------------------------------------------------------------
// Fixtures
// -----------------------------------------------------------------

func newUpgradingAM(name string, payload, annotations map[string]string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.UpgradeOp,
		},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Upgrading,
			OpType:  appv1.UpgradeOp,
			OpID:    "op-42",
			Payload: payload,
		},
	}
}

func upgradingWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — DIRECT HANDLER TESTS
// -----------------------------------------------------------------

func TestOnUpgradingTick_EmptyPhase_StartsPrepare(t *testing.T) {
	am := newUpgradingAM("app1", nil, nil)
	res := handlers.OnUpgradingTick(context.Background(), upgradingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if got := patch.Status.Payload[handlers.PayloadUpgradingPhase]; got != handlers.PhaseUpgradingPrepare {
		t.Fatalf("Payload[%q] = %q, want %q", handlers.PayloadUpgradingPhase, got, handlers.PhaseUpgradingPrepare)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok {
		t.Fatal("no CallWorkload")
	}
	if cw.Kind != domain.WorkloadPrepareUpgrade {
		t.Fatalf("Kind = %q, want WorkloadPrepareUpgrade", cw.Kind)
	}
	if cw.OnSuccess == nil || cw.OnFailure == nil {
		t.Fatal("OnSuccess/OnFailure nil (violates CONTRACT.md rule 5)")
	}
}

func TestOnUpgradingTick_KnownPhasesRecoverWithoutStatusEffects(t *testing.T) {
	tests := []struct {
		phase string
		call  domain.CallWorkload
	}{
		{
			phase: handlers.PhaseUpgradingPrepare,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadPrepareUpgrade,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "UpgradePrepareFailed"},
			},
		},
		{
			phase: handlers.PhaseUpgradingPull,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadPullImage,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "UpgradeImagePullFailed"},
			},
		},
		{
			phase: handlers.PhaseUpgradingHelm,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadUpgrade,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: "HelmUpgradeFailed"},
			},
		},
		{
			phase: handlers.PhaseUpgradingWaitStartUp,
			call: domain.CallWorkload{
				Kind:      domain.WorkloadWaitStartUp,
				OnSuccess: domain.WorkloadReady{},
				OnFailure: domain.WorkloadFailed{Reason: constants.AppStopDueToStartUpFailed},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.phase, func(t *testing.T) {
			am := newUpgradingAM("app1", map[string]string{
				handlers.PayloadUpgradingPhase: tc.phase,
			}, nil)
			assertTickRecovery(t, handlers.OnUpgradingTick, upgradingWorldFor(am), tc.call)
		})
	}
}

func TestOnUpgradingWorkloadReady_Prepare_StartsPull(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingPrepare,
	}, nil)
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok || cw.Kind != domain.WorkloadPullImage {
		t.Fatalf("Kind = %+v, want WorkloadPullImage", cw)
	}
}

func TestOnUpgradingWorkloadReady_Pull_StartsHelm(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingPull,
	}, nil)
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok || cw.Kind != domain.WorkloadUpgrade {
		t.Fatalf("Kind = %+v, want WorkloadUpgrade", cw)
	}
}

func TestOnUpgradingTick_UnknownPhase_HardFails(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: "alien",
	}, nil)
	res := handlers.OnUpgradingTick(context.Background(), upgradingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for unknown phase")
	}
}

// TestOnUpgradingWorkloadReady_HelmPhase_HandsOffToInitializing pins
// the normal path (no preState annotation → Initializing).
func TestOnUpgradingWorkloadReady_HelmPhase_StartsWaitStartUp(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingHelm,
	}, nil)
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (wait-startup still in Upgrading)", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if got := patch.Status.Payload[handlers.PayloadUpgradingPhase]; got != handlers.PhaseUpgradingWaitStartUp {
		t.Fatalf("phase = %q, want wait-startup", got)
	}
	cw, ok := findEffect[domain.CallWorkload](res.Effects)
	if !ok || cw.Kind != domain.WorkloadWaitStartUp {
		t.Fatalf("Kind = %+v, want WorkloadWaitStartUp", cw)
	}
}

func TestOnUpgradingWorkloadReady_WaitStartUp_HandsOffToInitializing(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingWaitStartUp,
	}, map[string]string{
		api.AppVersionKey:            "1.0.0",
		api.AppTargetChartVersionKey: "2.0.0",
		api.AppPreUpgradeConfigKey:   `{"version":"1.0.0"}`,
	})
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Initializing {
		t.Fatalf("Next = %q, want Initializing", res.Next)
	}
	spec, ok := findEffect[domain.PatchSpec](res.Effects)
	if !ok {
		t.Fatal("no PatchSpec committing the target chart version")
	}
	if spec.SetAnnotations[api.AppVersionKey] != "2.0.0" {
		t.Fatalf("committed version = %q, want 2.0.0", spec.SetAnnotations[api.AppVersionKey])
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Initializing {
		t.Fatalf("PatchAM.State = %q", patch.Status.State)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadUpgradingPhase]; ok {
		t.Fatal("Payload[upgrading.phase] still set after leaving Upgrading")
	}
	// No OpRecord — the terminal record lands at Initializing→Running.
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
		t.Fatal("unexpected AppendOpRecord on Upgrading→Initializing")
	}
}

func TestOnUpgradingWorkloadReady_AlreadyCommitted_HandsOff(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingWaitStartUp,
	}, map[string]string{
		api.AppVersionKey: "2.0.0",
	})
	am.Spec.Config = `{"Version":"2.0.0"}`
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Initializing {
		t.Fatalf("Next = %q, want Initializing", res.Next)
	}
	if _, ok := findEffect[domain.PatchSpec](res.Effects); ok {
		t.Fatal("commit was already applied; PatchSpec should be skipped")
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok || patch.Status.State != appv1.Initializing {
		t.Fatalf("PatchAM = %+v, want Initializing", patch.Status.State)
	}
}

func TestOnUpgradingWorkloadReady_MissingTargetVersion_Fails(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingWaitStartUp,
	}, map[string]string{api.AppVersionKey: "1.0.0"})
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want missing target chart version")
	}
	if res.Next != "" {
		t.Fatalf("Next = %q, want no transition", res.Next)
	}
}

// TestOnUpgradingWorkloadReady_HelmPhase_StoppedFastPath pins the
// upgrade-from-Stopped branch (preState annotation = "stopped" →
// land in Stopped with reason AppStopByUser).
func TestOnUpgradingWorkloadReady_HelmPhase_StoppedFastPath(t *testing.T) {
	am := newUpgradingAM("app1",
		map[string]string{handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingHelm},
		map[string]string{
			handlers.AnnotPreUpgradeState: appv1.Stopped.String(),
			api.AppVersionKey:             "1.0.0",
			api.AppTargetChartVersionKey:  "2.0.0",
			api.AppPreUpgradeConfigKey:    `{"version":"1.0.0"}`,
		},
	)
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.Stopped {
		t.Fatalf("Next = %q, want Stopped (upgrade-from-Stopped fast-path)", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.State != appv1.Stopped {
		t.Fatalf("PatchAM.State = %q, want Stopped", patch.Status.State)
	}
	if patch.Status.Reason != constants.AppStopByUser {
		t.Fatalf("PatchAM.Reason = %q, want %q", patch.Status.Reason, constants.AppStopByUser)
	}
	if _, ok := patch.Status.Payload[handlers.PayloadUpgradingPhase]; ok {
		t.Fatal("Payload[upgrading.phase] still set after leaving Upgrading")
	}
	record, ok := findEffect[domain.AppendOpRecord](res.Effects)
	if !ok {
		t.Fatal("Stopped fast-path must record the committed chart version")
	}
	if record.Record.Version != "2.0.0" {
		t.Fatalf("OpRecord.Version = %q, want 2.0.0", record.Record.Version)
	}
}

func TestOnUpgradingWorkloadReady_EmptyPhase_HardFails(t *testing.T) {
	am := newUpgradingAM("app1", nil, nil)
	res := handlers.OnUpgradingWorkloadReady(context.Background(), upgradingWorldFor(am))
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for missing phase")
	}
}

func TestOnUpgradingWorkloadFailed_UsesActionData(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingHelm,
	}, nil)
	w := upgradingWorldFor(am)
	w.Action = domain.WorkloadFailed{
		Reason:  "DowngradeNotAllowed",
		Message: "requested version 1.0 is < deployed version 1.2",
	}

	res := handlers.OnUpgradingWorkloadFailed(context.Background(), w)
	if res.Next != appv1.UpgradeFailed {
		t.Fatalf("Next = %q, want UpgradeFailed", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != "DowngradeNotAllowed" {
		t.Fatalf("Reason = %q", patch.Status.Reason)
	}
	if patch.Status.Message != "requested version 1.0 is < deployed version 1.2" {
		t.Fatalf("Message = %q", patch.Status.Message)
	}
	if _, ok := findEffect[domain.AppendOpRecord](res.Effects); !ok {
		t.Fatal("no AppendOpRecord on failure")
	}
	if ev, ok := findEffect[domain.EmitEvent](res.Effects); !ok || ev.Type != "Warning" {
		t.Fatalf("EmitEvent = %+v ok=%v, want Warning", ev, ok)
	}
}

func TestOnUpgradingWorkloadFailed_FallsBackWhenActionEmpty(t *testing.T) {
	am := newUpgradingAM("app1", map[string]string{
		handlers.PayloadUpgradingPhase: handlers.PhaseUpgradingHelm,
	}, nil)
	w := upgradingWorldFor(am)
	w.Action = domain.WorkloadFailed{}

	res := handlers.OnUpgradingWorkloadFailed(context.Background(), w)
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if patch.Status.Reason != string(appv1.UpgradeFailed) {
		t.Fatalf("fallback Reason = %q, want %q", patch.Status.Reason, string(appv1.UpgradeFailed))
	}
	if patch.Status.Message == "" {
		t.Fatal("fallback Message empty")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END TESTS
// -----------------------------------------------------------------

func (h *e2eHarness) seedUpgrading(name string, annotations map[string]string) string {
	ann := map[string]string{
		api.AppVersionKey:            "1.0.0",
		api.AppTargetChartVersionKey: "2.0.0",
		api.AppPreUpgradeConfigKey:   `{"version":"1.0.0"}`,
	}
	for k, v := range annotations {
		ann[k] = v
	}
	am := newUpgradingAM(name, nil, ann)
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
	return h.amStore.Add(am)
}

func driveUpgradeThrough(t *testing.T, h *e2eHarness, name string, phases []string) {
	t.Helper()
	drivePhases(t, h, name, handlers.PayloadUpgradingPhase, phases)
}

func drivePhases(t *testing.T, h *e2eHarness, name, key string, phases []string) {
	t.Helper()
	for i, phase := range phases {
		if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: name}); err != nil {
			t.Fatalf("upgrade round %d (want phase %s): %v", i+1, phase, err)
		}
		drainAndRebuild(t, h)
		got, err := h.amStore.Get(context.Background(), name)
		if err != nil {
			t.Fatalf("upgrade round %d: get: %v", i+1, err)
		}
		if p := got.Status.Payload[key]; p != phase {
			t.Fatalf("upgrade round %d: phase = %q, want %q", i+1, p, phase)
		}
	}
}

// TestUpgrading_EndToEnd_HappyPath drives Upgrading → Initializing:
//
//	prepare → pull → helm → wait-startup → Initializing
func TestUpgrading_EndToEnd_HappyPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUpgrading("myapp", nil)

	driveUpgradeThrough(t, h, "myapp", []string{
		handlers.PhaseUpgradingPrepare,
		handlers.PhaseUpgradingPull,
		handlers.PhaseUpgradingHelm,
		handlers.PhaseUpgradingWaitStartUp,
	})
	calls := h.workloadOps.Calls()
	if len(calls) == 0 || calls[len(calls)-1].Method != "WaitForStartUp" {
		t.Fatalf("workload calls = %+v, want last WaitForStartUp", calls)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("finish: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("state after finish = %q, want Initializing", got.Status.State)
	}
	// No OpRecord at Upgrading exit.
	if len(got.Status.OpRecords) != 0 {
		t.Fatalf("OpRecords = %+v, want empty (terminal record lands at Initializing→Running)", got.Status.OpRecords)
	}
	if got.Annotations[api.AppVersionKey] != "2.0.0" {
		t.Fatalf("chart version = %q, want committed 2.0.0", got.Annotations[api.AppVersionKey])
	}
	if _, ok := got.Annotations[api.AppTargetChartVersionKey]; ok {
		t.Fatal("target version was not cleared")
	}
	if _, ok := got.Annotations[api.AppPreUpgradeConfigKey]; ok {
		t.Fatal("pre-upgrade config snapshot was not cleared")
	}
}

// TestUpgrading_EndToEnd_StoppedFastPath drives Upgrading → Stopped
// when the preState annotation says the upgrade was requested from
// a stopped app.
func TestUpgrading_EndToEnd_StoppedFastPath(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUpgrading("myapp", map[string]string{
		handlers.AnnotPreUpgradeState: appv1.Stopped.String(),
	})

	driveUpgradeThrough(t, h, "myapp", []string{
		handlers.PhaseUpgradingPrepare,
		handlers.PhaseUpgradingPull,
		handlers.PhaseUpgradingHelm,
	})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("finish: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Stopped {
		t.Fatalf("state = %q, want Stopped (fast-path)", got.Status.State)
	}
	if got.Status.Reason != constants.AppStopByUser {
		t.Fatalf("Reason = %q, want %q", got.Status.Reason, constants.AppStopByUser)
	}
	if got.Annotations[api.AppVersionKey] != "2.0.0" {
		t.Fatalf("chart version = %q, want committed 2.0.0", got.Annotations[api.AppVersionKey])
	}
	if len(got.Status.OpRecords) != 1 || got.Status.OpRecords[0].Version != "2.0.0" {
		t.Fatalf("op records = %#v, want committed target version", got.Status.OpRecords)
	}
}

// TestUpgrading_EndToEnd_HelmFailure pins the failure path.
func TestUpgrading_EndToEnd_HelmFailure(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUpgrading("myapp", nil)
	h.workloadOps.UpgradeFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{LandedReason: "HelmError", Err: errors.New("upgrade failed")}
	}

	driveUpgradeThrough(t, h, "myapp", []string{
		handlers.PhaseUpgradingPrepare,
		handlers.PhaseUpgradingPull,
		handlers.PhaseUpgradingHelm,
	})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("finish: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradeFailed {
		t.Fatalf("state = %q, want UpgradeFailed", got.Status.State)
	}
	if got.Status.Reason != "HelmError" {
		t.Fatalf("Reason = %q, want HelmError", got.Status.Reason)
	}
	if got.Annotations[api.AppVersionKey] != "1.0.0" {
		t.Fatalf("chart version = %q, want deployed 1.0.0", got.Annotations[api.AppVersionKey])
	}
	if got.Annotations[api.AppTargetChartVersionKey] != "2.0.0" {
		t.Fatalf("target version = %q, want 2.0.0 kept for retry", got.Annotations[api.AppTargetChartVersionKey])
	}
}

// TestUpgrading_EndToEnd_CancelDuringHelm lands UpgradeFailed and
// leaves the target config plus the deployed chart version in place.
// Stopping is not on this path: a later UpgradeOp retries the same target.
func TestUpgrading_EndToEnd_CancelDuringHelm(t *testing.T) {
	h := newE2EHarness(t)
	h.seedUpgrading("myapp", nil)
	seeded, err := h.amStore.Get(context.Background(), "myapp")
	if err != nil {
		t.Fatal(err)
	}
	seeded.Spec.Config = `{"version":"2.0.0"}`
	if err := h.amStore.UpdateFull(context.Background(), seeded); err != nil {
		t.Fatal(err)
	}

	driveUpgradeThrough(t, h, "myapp", []string{
		handlers.PhaseUpgradingPrepare,
		handlers.PhaseUpgradingPull,
	})
	started, release, fn := parkedWorkload()
	h.workloadOps.UpgradeFunc = fn
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("dispatch helm: %v", err)
	}
	awaitChan(t, started, "helm upgrade to hold the claim")

	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.CancelOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("enter canceling: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradingCanceling {
		t.Fatalf("state = %q, want UpgradingCanceling", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("wait for helm: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradingCanceling {
		t.Fatalf("state = %q, want UpgradingCanceling while helm is still in flight", got.Status.State)
	}
	release()
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("consume cancel: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradeFailed {
		t.Fatalf("state = %q, want UpgradeFailed", got.Status.State)
	}
	if got.Spec.Config != `{"version":"2.0.0"}` {
		t.Fatalf("config = %q, want target config kept", got.Spec.Config)
	}
	if got.Annotations[api.AppVersionKey] != "1.0.0" {
		t.Fatalf("chart version = %q, want deployed 1.0.0", got.Annotations[api.AppVersionKey])
	}
	if got.Annotations[api.AppTargetChartVersionKey] != "2.0.0" {
		t.Fatalf("target version = %q, want 2.0.0 kept for retry", got.Annotations[api.AppTargetChartVersionKey])
	}
	if got.Annotations[api.AppPreUpgradeConfigKey] != `{"version":"1.0.0"}` {
		t.Fatalf("snapshot = %q, want pre-upgrade config kept", got.Annotations[api.AppPreUpgradeConfigKey])
	}
}

// TestFullLifecycle_UpgradeCycle proves the Running → Upgrading →
// Initializing → Running loop. Seven handlers fire across six
// Reconciles. Also verifies the terminal OpRecord chain is intact
// after upgrade completion.
func TestFullLifecycle_UpgradeCycle(t *testing.T) {
	h := newE2EHarness(t)
	h.seedRunning("myapp", appv1.App, true)
	seeded, _ := h.amStore.Get(context.Background(), "myapp")
	seeded.Annotations = map[string]string{
		api.AppVersionKey:            "1.0.0",
		api.AppTargetChartVersionKey: "2.0.0",
		api.AppPreUpgradeConfigKey:   `{"version":"1.0.0"}`,
	}
	if err := h.amStore.UpdateFull(context.Background(), seeded); err != nil {
		t.Fatalf("stamp upgrade metadata: %v", err)
	}

	// R1: Running + TriggerOp{UpgradeOp} → UpgradePending
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.UpgradeOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradePending {
		t.Fatalf("after R1, state = %q, want UpgradePending", got.Status.State)
	}
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("enter download: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.UpgradeDownloading {
		t.Fatalf("after slot, state = %q, want UpgradeDownloading", got.Status.State)
	}

	drivePhases(t, h, "myapp", handlers.PayloadUpgradeDownloadingPhase, []string{
		handlers.PhaseUpgradingPrepare,
		handlers.PhaseUpgradingPull,
	})
	driveUpgradeThrough(t, h, "myapp", []string{
		handlers.PhaseUpgradingHelm,
		handlers.PhaseUpgradingWaitStartUp,
	})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("upgrade finish: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Initializing {
		t.Fatalf("after upgrade, state = %q, want Initializing", got.Status.State)
	}

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("initializing dispatch: %v", err)
	}
	drainAndRebuild(t, h)
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("initializing finish: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}

	// Terminal OpRecord chain: head = Running (from
	// Initializing's terminal write). The op recorded is
	// UpgradeOp from the R1 trigger, but the *status* of the
	// terminal record is Running.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after upgrade cycle")
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord.Status = %q, want Running", got.Status.OpRecords[0].Status)
	}
}
