package handlers_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// -----------------------------------------------------------------
// Fixtures — thin, shared across all six terminal-state tests
// -----------------------------------------------------------------

// newTerminalAM builds an AM sitting in `state`. Terminal states
// carry no Payload phase, so this is a minimal fixture.
func newTerminalAM(name string, state appv1.ApplicationManagerState) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Source:       "user-upload",
			Type:         appv1.App,
			OpType:       appv1.InstallOp, // arbitrary — never read here
		},
		Status: appv1.ApplicationManagerStatus{
			State:  state,
			OpType: appv1.InstallOp,
			OpID:   "op-42",
		},
	}
}

func terminalWorldFor(am *appv1.ApplicationManager) domain.World {
	return domain.World{
		AM:  am,
		Now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// -----------------------------------------------------------------
// Layer 1 — TICK NO-OP contract
// -----------------------------------------------------------------
//
// All six terminal ticks must return a zero Result. One
// table-driven test locks in the contract for every state at
// once — a new op or state mistakenly wired into a Tick handler
// would break exactly one row and be trivially diagnosable.

func TestTerminalTicks_AreAllNoop(t *testing.T) {
	cases := []struct {
		state   appv1.ApplicationManagerState
		handler func(context.Context, domain.World) domain.Result
	}{
		{appv1.Uninstalled, handlers.OnUninstalledTick},
		{appv1.DownloadFailed, handlers.OnDownloadFailedTick},
		{appv1.UninstallFailed, handlers.OnUninstallFailedTick},
		{appv1.StopFailed, handlers.OnStopFailedTick},
		{appv1.ResumeFailed, handlers.OnResumeFailedTick},
		{appv1.UpgradeFailed, handlers.OnUpgradeFailedTick},
		{appv1.ApplyEnvFailed, handlers.OnApplyEnvFailedTick},

		// Cancel-terminal group additions.
		{appv1.PendingCanceled, handlers.OnPendingCanceledTick},
		{appv1.DownloadingCanceled, handlers.OnDownloadingCanceledTick},
		{appv1.InstallingCanceled, handlers.OnInstallingCanceledTick},
		{appv1.PendingCancelFailed, handlers.OnPendingCancelFailedTick},
		{appv1.DownloadingCancelFailed, handlers.OnDownloadingCancelFailedTick},
		{appv1.InstallingCancelFailed, handlers.OnInstallingCancelFailedTick},
		{appv1.UpgradingCancelFailed, handlers.OnUpgradingCancelFailedTick},
		{appv1.ApplyingEnvCancelFailed, handlers.OnApplyingEnvCancelFailedTick},
		{appv1.ResumingCancelFailed, handlers.OnResumingCancelFailedTick},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(string(tc.state), func(t *testing.T) {
			am := newTerminalAM("app1", tc.state)
			res := tc.handler(context.Background(), terminalWorldFor(am))
			if res.Err != nil {
				t.Fatalf("Err = %v, want nil", res.Err)
			}
			if res.Next != "" {
				t.Fatalf("Next = %q, want empty (terminal Tick must not transition)", res.Next)
			}
			if len(res.Effects) != 0 {
				t.Fatalf("Effects = %+v, want none", res.Effects)
			}
		})
	}
}

// TestTerminalTicks_NilAM_AreSafeNoop pins the defensive guard
// across every terminal Tick — none may panic on a hand-crafted
// zero World.
func TestTerminalTicks_NilAM_AreSafeNoop(t *testing.T) {
	fns := []func(context.Context, domain.World) domain.Result{
		handlers.OnUninstalledTick,
		handlers.OnInstallFailedTick,
		handlers.OnDownloadFailedTick,
		handlers.OnUninstallFailedTick,
		handlers.OnStopFailedTick,
		handlers.OnResumeFailedTick,
		handlers.OnUpgradeFailedTick,
		handlers.OnApplyEnvFailedTick,

		// Cancel-terminal group additions.
		handlers.OnPendingCanceledTick,
		handlers.OnDownloadingCanceledTick,
		handlers.OnInstallingCanceledTick,
		handlers.OnPendingCancelFailedTick,
		handlers.OnDownloadingCancelFailedTick,
		handlers.OnInstallingCancelFailedTick,
		handlers.OnUpgradingCancelFailedTick,
		handlers.OnApplyingEnvCancelFailedTick,
		handlers.OnResumingCancelFailedTick,
	}
	for i, fn := range fns {
		res := fn(context.Background(), domain.World{AM: nil})
		if res.Err != nil {
			t.Fatalf("handler #%d Err = %v, want nil for nil AM", i, res.Err)
		}
	}
}

func TestOnInstallFailedTick_CleanupPhase_RecoversOutstandingWorkload(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallFailedCleanup: handlers.PhaseInstallFailedCleanup,
	}
	w := terminalWorldFor(am)
	wantCall := domain.CallWorkload{
		Kind:      domain.WorkloadCleanupInstallFailure,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "InstallCleanupFailed"},
	}

	recovered := handlers.OnInstallFailedTick(context.Background(), w)
	want := domain.Result{Effects: []domain.Effect{wantCall}}
	if !reflect.DeepEqual(recovered, want) {
		t.Fatalf("recovered Result = %#v, want exactly %#v", recovered, want)
	}

	w.WorkloadInFlight = true
	inFlight := handlers.OnInstallFailedTick(context.Background(), w)
	if !reflect.DeepEqual(inFlight, domain.Result{}) {
		t.Fatalf("in-flight Result = %#v, want empty", inFlight)
	}
}

func TestOnInstallFailedTick_EmptyCleanup_StartsCleanupWithStatusPatch(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	res := handlers.OnInstallFailedTick(context.Background(), terminalWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if len(res.Effects) != 2 {
		t.Fatalf("Effects = %#v, want exactly PatchAM and CallWorkload", res.Effects)
	}
	patch, ok := res.Effects[0].(domain.PatchAM)
	if !ok {
		t.Fatalf("first effect = %T, want PatchAM", res.Effects[0])
	}
	if patch.Status.State != appv1.InstallFailed ||
		patch.Status.Reason != "InstallCleanupStarted" ||
		patch.Status.Message != "installing.cleanup: tearing down partial install after failure" ||
		patch.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanup {
		t.Fatalf("PatchAM.Status = %#v, want cleanup-start status", patch.Status)
	}
	wantCall := domain.CallWorkload{
		Kind:      domain.WorkloadCleanupInstallFailure,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "InstallCleanupFailed"},
	}
	if !reflect.DeepEqual(res.Effects[1], wantCall) {
		t.Fatalf("second effect = %#v, want exactly %#v", res.Effects[1], wantCall)
	}
}

func TestOnInstallFailedTick_CleanupDone_BecomesUninstalled(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallFailedCleanup: handlers.PhaseInstallFailedCleanupDone,
	}
	res := handlers.OnInstallFailedTick(context.Background(), terminalWorldFor(am))
	if res.Next != appv1.Uninstalled {
		t.Fatalf("Next = %q, want Uninstalled", res.Next)
	}
	patch, ok := res.Effects[0].(domain.PatchAM)
	if !ok {
		t.Fatalf("first effect = %T, want PatchAM", res.Effects[0])
	}
	if patch.Status.State != appv1.Uninstalled ||
		patch.Status.Reason != "InstallFailureCleanedUp" ||
		len(patch.Status.Payload) != 0 {
		t.Fatalf("PatchAM.Status = %#v, want cleaned-up Uninstalled", patch.Status)
	}
}

func TestOnInstallFailedCleanupDone_LandsUninstalled(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	w := terminalWorldFor(am)
	w.Action = domain.WorkloadReady{}
	res := handlers.OnInstallFailedCleanupDone(context.Background(), w)
	if res.Next != appv1.Uninstalled {
		t.Fatalf("Next = %q, want Uninstalled", res.Next)
	}
	patch, ok := res.Effects[0].(domain.PatchAM)
	if !ok || patch.Status.Reason != "InstallFailureCleanedUp" {
		t.Fatalf("patch = %#v", patch)
	}

}

func TestOnInstallFailedCleanupDone_FailureStaysInstallFailedAndSchedulesRetry(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	w := terminalWorldFor(am)
	w.Action = domain.WorkloadFailed{Reason: "InstallCleanupFailed"}
	res := handlers.OnInstallFailedCleanupDone(context.Background(), w)
	if res.Next != "" {
		t.Fatalf("Next = %q, want empty (stay InstallFailed)", res.Next)
	}
	patch, ok := res.Effects[0].(domain.PatchAM)
	if !ok {
		t.Fatalf("first effect = %T, want PatchAM", res.Effects[0])
	}
	wantRetryAt := w.Now.Add(time.Minute).Format(time.RFC3339)
	if patch.Status.State != appv1.InstallFailed ||
		patch.Status.Reason != "InstallCleanupRetrying" ||
		patch.Status.Message != "install cleanup failed; retrying: InstallCleanupFailed" ||
		patch.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanupRetry ||
		patch.Status.Payload[handlers.PayloadInstallFailedCleanupRetryAt] != wantRetryAt {
		t.Fatalf("PatchAM.Status = %#v, want retry status", patch.Status)
	}
	var tick domain.ScheduleTick
	for _, e := range res.Effects {
		if st, ok := e.(domain.ScheduleTick); ok {
			tick = st
		}
	}
	if tick.After != time.Minute {
		t.Fatalf("ScheduleTick.After = %s, want 1m", tick.After)
	}
}

func TestOnInstallFailedTick_RetryPhase_WaitsOutBackoff(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	w := terminalWorldFor(am)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallFailedCleanup:        handlers.PhaseInstallFailedCleanupRetry,
		handlers.PayloadInstallFailedCleanupRetryAt: w.Now.Add(30 * time.Second).Format(time.RFC3339),
	}
	res := handlers.OnInstallFailedTick(context.Background(), w)
	want := domain.Result{Effects: []domain.Effect{domain.ScheduleTick{After: 30 * time.Second}}}
	if !reflect.DeepEqual(res, want) {
		t.Fatalf("Result = %#v, want %#v", res, want)
	}
}

func TestOnInstallFailedTick_RetryPhase_RestartsCleanupAfterBackoff(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallFailed)
	w := terminalWorldFor(am)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallFailedCleanup:        handlers.PhaseInstallFailedCleanupRetry,
		handlers.PayloadInstallFailedCleanupRetryAt: w.Now.Add(-time.Second).Format(time.RFC3339),
	}
	res := handlers.OnInstallFailedTick(context.Background(), w)
	if len(res.Effects) != 2 {
		t.Fatalf("Effects = %#v, want PatchAM and CallWorkload", res.Effects)
	}
	patch, ok := res.Effects[0].(domain.PatchAM)
	if !ok || patch.Status.Payload[handlers.PayloadInstallFailedCleanup] != handlers.PhaseInstallFailedCleanup {
		t.Fatalf("first effect = %#v, want cleanup-phase PatchAM", res.Effects[0])
	}
	if !reflect.DeepEqual(patch.DeletePayloadKeys, []string{handlers.PayloadInstallFailedCleanupRetryAt}) {
		t.Fatalf("DeletePayloadKeys = %v, want retry stamp removed", patch.DeletePayloadKeys)
	}
	call, ok := res.Effects[1].(domain.CallWorkload)
	if !ok || call.Kind != domain.WorkloadCleanupInstallFailure {
		t.Fatalf("second effect = %#v, want cleanup CallWorkload", res.Effects[1])
	}

	w.WorkloadInFlight = true
	w.InFlightKind = domain.WorkloadCleanupInstallFailure
	if got := handlers.OnInstallFailedTick(context.Background(), w); !reflect.DeepEqual(got, domain.Result{}) {
		t.Fatalf("in-flight Result = %#v, want empty", got)
	}
}

// -----------------------------------------------------------------
// Layer 1 — OP-ROUTING contract
// -----------------------------------------------------------------
//
// One row per (state, op, expected-target, expected-reason).
// Twelve rows total — the entire terminal-state op surface pinned
// in a single table. A future edit that swaps target/reason on
// any row will fail loudly.

type terminalOpCase struct {
	name    string
	state   appv1.ApplicationManagerState
	handler func(context.Context, domain.World) domain.Result
	target  appv1.ApplicationManagerState
	reason  string
}

func terminalOpMatrix() []terminalOpCase {
	return []terminalOpCase{
		// Uninstalled
		{"uninstalled/install", appv1.Uninstalled, handlers.OnUninstalledInstall,
			appv1.Pending, "InstallRequested"},

		// InstallFailed
		{"installfailed/uninstall", appv1.InstallFailed, handlers.OnInstallFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// DownloadFailed
		{"downloadfailed/install", appv1.DownloadFailed, handlers.OnDownloadFailedInstall,
			appv1.Pending, "InstallRequested"},

		// UninstallFailed
		{"uninstallfailed/uninstall", appv1.UninstallFailed, handlers.OnUninstallFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// StopFailed
		{"stopfailed/stop", appv1.StopFailed, handlers.OnStopFailedStop,
			appv1.Stopping, "StopRequested"},
		{"stopfailed/uninstall", appv1.StopFailed, handlers.OnStopFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// ResumeFailed
		{"resumefailed/resume", appv1.ResumeFailed, handlers.OnResumeFailedResume,
			appv1.Resuming, "ResumeRequested"},
		{"resumefailed/applyenv", appv1.ResumeFailed, handlers.OnResumeFailedApplyEnv,
			appv1.ApplyingEnv, "ApplyEnvRequested"},
		{"resumefailed/uninstall", appv1.ResumeFailed, handlers.OnResumeFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// UpgradeFailed
		{"upgradefailed/upgrade", appv1.UpgradeFailed, handlers.OnUpgradeFailedUpgrade,
			appv1.UpgradePending, "UpgradeRequested"},
		{"upgradefailed/uninstall", appv1.UpgradeFailed, handlers.OnUpgradeFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// ApplyEnvFailed
		{"applyenvfailed/applyenv", appv1.ApplyEnvFailed, handlers.OnApplyEnvFailedApplyEnv,
			appv1.ApplyingEnv, "ApplyEnvRequested"},
		{"applyenvfailed/stop", appv1.ApplyEnvFailed, handlers.OnApplyEnvFailedStop,
			appv1.Stopping, "StopRequested"},
		{"applyenvfailed/uninstall", appv1.ApplyEnvFailed, handlers.OnApplyEnvFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},

		// Cancel-terminal group — Canceled family (InstallOp
		// re-target).
		{"pendingcanceled/install", appv1.PendingCanceled, handlers.OnPendingCanceledInstall,
			appv1.Pending, "InstallRequested"},
		{"downloadingcanceled/install", appv1.DownloadingCanceled, handlers.OnDownloadingCanceledInstall,
			appv1.Pending, "InstallRequested"},
		{"installingcanceled/install", appv1.InstallingCanceled, handlers.OnInstallingCanceledInstall,
			appv1.Pending, "InstallRequested"},

		// Cancel-terminal group — CancelFailed family
		// (CancelRetry back to *Canceling; UninstallOp abandon
		// on install-pipeline variants).
		{"pendingcancelfailed/cancel", appv1.PendingCancelFailed, handlers.OnPendingCancelFailedCancel,
			appv1.PendingCanceling, "CancelRetryRequested"},
		{"downloadingcancelfailed/cancel", appv1.DownloadingCancelFailed, handlers.OnDownloadingCancelFailedCancel,
			appv1.DownloadingCanceling, "CancelRetryRequested"},
		{"installingcancelfailed/cancel", appv1.InstallingCancelFailed, handlers.OnInstallingCancelFailedCancel,
			appv1.InstallingCanceling, "CancelRetryRequested"},
		{"installingcancelfailed/uninstall", appv1.InstallingCancelFailed, handlers.OnInstallingCancelFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},
		{"upgradingcancelfailed/cancel", appv1.UpgradingCancelFailed, handlers.OnUpgradingCancelFailedCancel,
			appv1.UpgradingCanceling, "CancelRetryRequested"},
		{"upgradingcancelfailed/uninstall", appv1.UpgradingCancelFailed, handlers.OnUpgradingCancelFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},
		{"applyingenvcancelfailed/cancel", appv1.ApplyingEnvCancelFailed, handlers.OnApplyingEnvCancelFailedCancel,
			appv1.ApplyingEnvCanceling, "CancelRetryRequested"},
		{"applyingenvcancelfailed/uninstall", appv1.ApplyingEnvCancelFailed, handlers.OnApplyingEnvCancelFailedUninstall,
			appv1.Uninstalling, "UninstallRequested"},
		{"resumingcancelfailed/cancel", appv1.ResumingCancelFailed, handlers.OnResumingCancelFailedCancel,
			appv1.ResumingCanceling, "CancelRetryRequested"},
	}
}

func TestTerminalOp_Transitions(t *testing.T) {
	for _, tc := range terminalOpMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			am := newTerminalAM("app1", tc.state)
			res := tc.handler(context.Background(), terminalWorldFor(am))
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
			// No OpRecord at terminal-exit — matches the
			// terminalOpTransition contract (destination state
			// owns the terminal record).
			if _, ok := findEffect[domain.AppendOpRecord](res.Effects); ok {
				t.Fatalf("unexpected AppendOpRecord on %s→%s", tc.state, tc.target)
			}
		})
	}
}

func TestInstallingCancelFailedCancel_PreservesInstallFailureIntent(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallingCancelFailed)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallingCancelingPhase:    handlers.PhaseCancelWaitNS,
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	}

	res := handlers.OnInstallingCancelFailedCancel(context.Background(), terminalWorldFor(am))
	if res.Err != nil {
		t.Fatalf("Err = %v", res.Err)
	}
	if res.Next != appv1.InstallingCanceling {
		t.Fatalf("Next = %q, want InstallingCanceling", res.Next)
	}
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	for key, want := range map[string]string{
		handlers.PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
		handlers.PayloadInstallingFailMessage:       "chart apply failed",
		handlers.PayloadInstallingFailReason:        "HelmError",
	} {
		if got := patch.Status.Payload[key]; got != want {
			t.Fatalf("retry patch Payload[%q] = %q, want %q", key, got, want)
		}
	}
	if got := patch.Status.Payload[handlers.PayloadInstallingCancelingPhase]; got != "" {
		t.Fatalf("retry patch phase = %q, want cleared", got)
	}
}

func TestInstallingCancelFailedCancel_UserCancelKeepsIntentAbsent(t *testing.T) {
	am := newTerminalAM("app1", appv1.InstallingCancelFailed)
	am.Status.Payload = map[string]string{
		handlers.PayloadInstallingCancelingPhase: handlers.PhaseCancelWaitNS,
	}

	res := handlers.OnInstallingCancelFailedCancel(context.Background(), terminalWorldFor(am))
	patch, _ := findEffect[domain.PatchAM](res.Effects)
	if _, ok := patch.Status.Payload[handlers.PayloadInstallingCancelingTerminal]; ok {
		t.Fatal("user-originated cancel retry unexpectedly stamped InstallFailed intent")
	}
}

// TestTerminalOp_NilAM_HardFails pins the defensive Err path
// for the shared helper.
func TestTerminalOp_NilAM_HardFails(t *testing.T) {
	res := handlers.OnUninstalledInstall(context.Background(), domain.World{AM: nil})
	if res.Err == nil {
		t.Fatal("Err = nil, want failure for nil-AM invocation")
	}
}

// -----------------------------------------------------------------
// Layer 2 — END-TO-END OP ROUTING
// -----------------------------------------------------------------
//
// One Reconcile per state per representative op, proving the
// TriggerOp{Op:…} → state-flip path works end-to-end for every
// terminal state. Kept to one representative op per state (rather
// than 12 e2e cases) — the Layer-1 matrix already pins each
// handler individually; this layer is about "the Runner's
// dispatch actually delivers TriggerOp into these handlers".

type terminalReconcileCase struct {
	name     string
	state    appv1.ApplicationManagerState
	op       appv1.OpType
	expected appv1.ApplicationManagerState
}

func terminalReconcileMatrix() []terminalReconcileCase {
	return []terminalReconcileCase{
		{"uninstalled_install", appv1.Uninstalled, appv1.InstallOp, appv1.Pending},
		{"installfailed_uninstall", appv1.InstallFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"downloadfailed_install", appv1.DownloadFailed, appv1.InstallOp, appv1.Pending},
		{"uninstallfailed_uninstall", appv1.UninstallFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"stopfailed_stop", appv1.StopFailed, appv1.StopOp, appv1.Stopping},
		{"stopfailed_uninstall", appv1.StopFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"resumefailed_resume", appv1.ResumeFailed, appv1.ResumeOp, appv1.Resuming},
		{"resumefailed_applyenv", appv1.ResumeFailed, appv1.ApplyEnvOp, appv1.ApplyingEnv},
		{"resumefailed_uninstall", appv1.ResumeFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"upgradefailed_upgrade", appv1.UpgradeFailed, appv1.UpgradeOp, appv1.UpgradePending},
		{"upgradefailed_uninstall", appv1.UpgradeFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"applyenvfailed_applyenv", appv1.ApplyEnvFailed, appv1.ApplyEnvOp, appv1.ApplyingEnv},
		{"applyenvfailed_stop", appv1.ApplyEnvFailed, appv1.StopOp, appv1.Stopping},
		{"applyenvfailed_uninstall", appv1.ApplyEnvFailed, appv1.UninstallOp, appv1.Uninstalling},

		// Cancel-terminal group — Canceled family.
		{"pendingcanceled_install", appv1.PendingCanceled, appv1.InstallOp, appv1.Pending},
		{"downloadingcanceled_install", appv1.DownloadingCanceled, appv1.InstallOp, appv1.Pending},
		{"installingcanceled_install", appv1.InstallingCanceled, appv1.InstallOp, appv1.Pending},

		// Cancel-terminal group — CancelFailed family.
		{"pendingcancelfailed_cancel", appv1.PendingCancelFailed, appv1.CancelOp, appv1.PendingCanceling},
		{"downloadingcancelfailed_cancel", appv1.DownloadingCancelFailed, appv1.CancelOp, appv1.DownloadingCanceling},
		{"installingcancelfailed_cancel", appv1.InstallingCancelFailed, appv1.CancelOp, appv1.InstallingCanceling},
		{"installingcancelfailed_uninstall", appv1.InstallingCancelFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"upgradingcancelfailed_cancel", appv1.UpgradingCancelFailed, appv1.CancelOp, appv1.UpgradingCanceling},
		{"upgradingcancelfailed_uninstall", appv1.UpgradingCancelFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"applyingenvcancelfailed_cancel", appv1.ApplyingEnvCancelFailed, appv1.CancelOp, appv1.ApplyingEnvCanceling},
		{"applyingenvcancelfailed_uninstall", appv1.ApplyingEnvCancelFailed, appv1.UninstallOp, appv1.Uninstalling},
		{"resumingcancelfailed_cancel", appv1.ResumingCancelFailed, appv1.CancelOp, appv1.ResumingCanceling},
	}
}

func TestTerminal_EndToEnd_OpRouting(t *testing.T) {
	for _, tc := range terminalReconcileMatrix() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := newE2EHarness(t)
			h.amStore.Add(newTerminalAM("myapp", tc.state))
			h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

			h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: tc.op})

			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
				t.Fatalf("Reconcile: %v", err)
			}
			got, _ := h.amStore.Get(context.Background(), "myapp")
			if got.Status.State != tc.expected {
				t.Fatalf("state = %q, want %q", got.Status.State, tc.expected)
			}
		})
	}
}

// TestTerminal_EndToEnd_HealthyTicks pins the steady-state no-op
// end-to-end for every terminal state. A watch-driven Reconcile
// firing on an unrelated change (annotation update, sibling
// controller write) must NOT mutate state or emit spurious
// effects.
func TestTerminal_EndToEnd_HealthyTicks(t *testing.T) {
	states := []appv1.ApplicationManagerState{
		appv1.Uninstalled,
		appv1.DownloadFailed,
		appv1.UninstallFailed,
		appv1.StopFailed,
		appv1.ResumeFailed,
		appv1.UpgradeFailed,
		appv1.ApplyEnvFailed,

		// Cancel-terminal group additions.
		appv1.PendingCanceled,
		appv1.DownloadingCanceled,
		appv1.InstallingCanceled,
		appv1.PendingCancelFailed,
		appv1.DownloadingCancelFailed,
		appv1.InstallingCancelFailed,
		appv1.UpgradingCancelFailed,
		appv1.ApplyingEnvCancelFailed,
		appv1.ResumingCancelFailed,
	}
	for _, s := range states {
		s := s
		t.Run(string(s), func(t *testing.T) {
			h := newE2EHarness(t)
			h.amStore.Add(newTerminalAM("myapp", s))
			h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
				t.Fatalf("Reconcile: %v", err)
			}
			got, _ := h.amStore.Get(context.Background(), "myapp")
			if got.Status.State != s {
				t.Fatalf("state = %q, want %q (terminal Tick must not mutate)", got.Status.State, s)
			}
		})
	}
}

// -----------------------------------------------------------------
// Layer 2 — LOAD-BEARING lifecycle: Uninstalled → reinstall → Running
// -----------------------------------------------------------------

// TestFullLifecycle_UninstalledReinstall proves the terminal-
// reinstallable contract end-to-end: a previously-uninstalled AM
// can be brought back to Running by delivering a single
// TriggerOp{InstallOp} — the state machine handles all the
// intermediate Pending → Downloading → Installing → Initializing
// → Running hops automatically.
//
// This is the analogue of TestFullLifecycle_InstallThroughUninstall
// but starting from the terminal end: proves that the state
// machine is *idempotent under re-installation*, i.e. an operator
// can retry install from a clean terminal state without any
// out-of-band cleanup.
func TestFullLifecycle_UninstalledReinstall(t *testing.T) {
	h := newE2EHarness(t)
	// Seed an already-uninstalled AM (no paired Application).
	h.amStore.Add(newTerminalAM("myapp", appv1.Uninstalled))

	// R1: Uninstalled + TriggerOp{InstallOp} → Pending
	h.runner.Pending().Deliver("myapp", domain.TriggerOp{Op: appv1.InstallOp})
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	got, _ := h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Pending {
		t.Fatalf("after R1, state = %q, want Pending", got.Status.State)
	}

	// R2: Pending → Downloading
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R2: %v", err)
	}
	// R3: Downloading(phase="") → dispatch pull
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R3: %v", err)
	}
	drainAndRebuild(t, h)
	// R4: Downloading(phase=pull) + WorkloadReady → Installing
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R4: %v", err)
	}
	// The Installing pipeline: validate, helm-install, scale-up,
	// wait-startup, then the hand-off to Initializing.
	driveInstallToInitializing(t, h, "myapp")

	// Synthesize paired Application post-helm (real helm adapter
	// creates it; the fake doesn't).
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "myapp"}})

	// R8: Initializing(phase="") → dispatch WaitLaunch
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R8: %v", err)
	}
	drainAndRebuild(t, h)
	// R9: Initializing(phase=wait) + WorkloadReady → Running
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "myapp"}); err != nil {
		t.Fatalf("R9: %v", err)
	}
	got, _ = h.amStore.Get(context.Background(), "myapp")
	if got.Status.State != appv1.Running {
		t.Fatalf("final state = %q, want Running", got.Status.State)
	}

	// Sanity: the Running OpRecord landed (from Initializing's
	// terminal write). We're not asserting the earlier lineage
	// here — that's TestFullLifecycle_InstallThroughUninstall's
	// job.
	if len(got.Status.OpRecords) == 0 {
		t.Fatal("no OpRecords after reinstall")
	}
	if got.Status.OpRecords[0].Status != appv1.Running {
		t.Fatalf("head OpRecord.Status = %q, want Running", got.Status.OpRecords[0].Status)
	}
}
