// Package handlers, terminal-states section.
//
// This file collects the pure-terminal-with-op-restart states in
// one batch, because they share an identical shape:
//
// "Failed" family (op-restart / abandon):
//
//   - Uninstalled       (op: InstallOp)
//   - InstallFailed     (op: UninstallOp; Tick cleans up then Uninstalled)
//   - DownloadFailed    (op: InstallOp)
//   - UninstallFailed   (op: UninstallOp)
//   - StopFailed        (ops: StopOp, UninstallOp)
//   - ResumeFailed      (ops: ResumeOp, ApplyEnvOp, UninstallOp)
//   - UpgradeFailed     (ops: UpgradeOp, UninstallOp)
//   - ApplyEnvFailed    (ops: ApplyEnvOp, StopOp, UninstallOp)
//
// "Canceled" family (install-pipeline cancel terminal — only
// InstallOp allowed, to re-install a previously-canceled AM):
//
//   - PendingCanceled     (op: InstallOp)
//   - DownloadingCanceled (op: InstallOp)
//   - InstallingCanceled  (op: InstallOp)
//
// "CancelFailed" family (cancel-retry / abandon — the cleanup
// pipeline itself failed and the operator can retry the cancel or
// abandon via uninstall):
//
//   - PendingCancelFailed     (op: CancelOp)
//   - DownloadingCancelFailed (op: CancelOp)
//   - InstallingCancelFailed     (ops: CancelOp, UninstallOp)
//   - InitializingCancelFailed   (ops: CancelOp, UninstallOp)
//   - UpgradingCancelFailed      (ops: CancelOp, UninstallOp)
//   - ApplyingEnvCancelFailed (ops: CancelOp, UninstallOp)
//   - ResumingCancelFailed    (op: CancelOp; Tick stays put)
//
// Each handler is a Tick no-op plus one op-driven transition per
// legal op. The op transitions are all "flip state to the canonical
// target + emit a Normal event" — the target state's own handler
// owns every side effect. Zero CallWorkload dispatches, zero
// OpRecord writes at this layer.
//
// Why one file? These states are structurally identical (all
// delegate to terminalOpTransition) and per-state files would
// multiply LOC ~5x without adding discoverability. Register/Tick/
// op-handler triples are grouped per-state within this file, each
// with its own doc comment; the pattern reads top-to-bottom.
package handlers

import (
	"context"
	"fmt"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// -----------------------------------------------------------------
// Shared reason/message constants
// -----------------------------------------------------------------
//
// One (reason, message) pair per Op — regardless of which terminal
// state the request came from. Rationale: the operator UX keys off
// Reason for "what is happening" (State already provides "from
// where"), so uniformity across states makes filtering + audit-
// grepping much easier.
//
// Values match running.go / stopped.go verbatim where the same op
// exists, so all callers converge on identical event text.
const (
	terminalReasonInstall   = "InstallRequested"
	terminalMsgInstall      = "user requested install; transitioning to Pending"
	terminalReasonUninstall = "UninstallRequested"
	terminalMsgUninstall    = "user requested uninstall; transitioning to Uninstalling"
	terminalReasonUpgrade   = "UpgradeRequested"
	terminalMsgUpgrade      = "user requested upgrade; transitioning to UpgradePending"
	terminalReasonApplyEnv  = "ApplyEnvRequested"
	terminalMsgApplyEnv     = "user requested applyEnv; transitioning to ApplyingEnv"
	terminalReasonResume    = "ResumeRequested"
	terminalMsgResume       = "user requested resume; transitioning to Resuming"
	terminalReasonStop      = "StopRequested"
	terminalMsgStop         = "user requested stop; transitioning to Stopping"

	// Cancel-retry template messages. Each *CancelFailed state
	// routes CancelOp back into its matching *Canceling state,
	// giving the operator a way to retry the failed cleanup
	// without having to hand-craft an escape hatch. The message
	// carries the target *Canceling state name so a UI banner
	// can distinguish "retrying pending cancel" from "retrying
	// upgrading cancel" at a glance.
	terminalReasonCancelRetry          = "CancelRetryRequested"
	terminalMsgCancelRetryPending      = "user requested cancel retry; transitioning back to PendingCanceling"
	terminalMsgCancelRetryDownloading  = "user requested cancel retry; transitioning back to DownloadingCanceling"
	terminalMsgCancelRetryInstalling   = "user requested cancel retry; transitioning back to InstallingCanceling"
	terminalMsgCancelRetryInitializing = "user requested cancel retry; transitioning back to InitializingCanceling"
	terminalMsgCancelRetryUpgrading    = "user requested cancel retry; transitioning back to UpgradingCanceling"
	terminalMsgCancelRetryApplyingEnv  = "user requested cancel retry; transitioning back to ApplyingEnvCanceling"
	terminalMsgCancelRetryResuming     = "user requested cancel retry; transitioning back to ResumingCanceling"
)

// terminalOpTransition is the shared op-transition helper for
// every terminal state. Same shape as running.runningOpTransition
// and stopped.stoppedOpTransition — extracted here so all
// terminal states converge on one implementation.
//
// The op emits NO OpRecord: the terminal record for a successful
// restart / uninstall / etc. lands at the destination state's
// terminal handler (Running / Stopped / Uninstalled …). This
// matches the legacy "makeRecord fires at operation completion"
// convention.
//
// Nil-AM guard mirrors runningOpTransition: production is safe
// because the Runner always populates AM, but a hand-crafted
// test World must not silently no-op.
func terminalOpTransition(
	w domain.World,
	target appv1.ApplicationManagerState,
	reason, msg string,
) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("terminal: %s handler invoked with nil AM", reason),
		}
	}
	return domain.Result{
		Next: target,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   target,
				Reason:  reason,
				Message: msg,
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// noopTerminalTick is the shared Tick body for every terminal
// state. Watch-driven Reconciles firing on unrelated changes
// (annotations, spec updates, sibling controller writes) hit
// this path — must not error, must not mutate.
func noopTerminalTick(_ context.Context, _ domain.World) domain.Result {
	return domain.Result{}
}

// -----------------------------------------------------------------
// Uninstalled
// -----------------------------------------------------------------
//
// The clean terminal state after a successful uninstall. Only
// legal op is InstallOp — a re-install of a previously-uninstalled
// app. Target: Pending (fresh install pipeline start).
//
// Legacy correspondence: pkg/appstate/uninstalled_app.go is
// intentionally NOT read here — the legacy StatefulApp for
// Uninstalled has no exec body (it's a pure terminal). The
// re-install path lives in the higher-level state-transition
// tables (opTarget) which we already ported into table_default.
func registerUninstalled(t *domain.TransitionTable) {
	t.AddReachable(appv1.Uninstalled, appv1.Uninstalled)
	t.SetOrAddHandler(appv1.Uninstalled, domain.Tick{}, OnUninstalledTick)
	t.SetOrAddHandler(appv1.Uninstalled, domain.TriggerOp{Op: appv1.InstallOp}, OnUninstalledInstall)
}

// OnUninstalledTick is a no-op — see noopTerminalTick.
func OnUninstalledTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnUninstalledInstall re-installs a previously-uninstalled app.
func OnUninstalledInstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Pending, terminalReasonInstall, terminalMsgInstall)
}

// -----------------------------------------------------------------
// InstallFailed
// -----------------------------------------------------------------
//
// Terminal state after a failed install (validator rejection, helm
// failure, etc.). Install is not accepted here. The Tick owns
// background cleanup and retries it until the namespace is gone; only
// then does the AM become Uninstalled, which is where a fresh install
// starts. There is no deadline: leaving early would let a reinstall
// race a namespace that is still terminating. UninstallOp remains, so
// an operator can abandon into the uninstall pipeline at any time.
func registerInstallFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.InstallFailed, appv1.InstallFailed)
	t.AddReachable(appv1.InstallFailed, appv1.Uninstalled)
	t.SetOrAddHandler(appv1.InstallFailed, domain.Tick{}, OnInstallFailedTick)
	t.SetOrAddHandler(appv1.InstallFailed, domain.WorkloadReady{}, OnInstallFailedCleanupDone)
	t.SetOrAddHandler(appv1.InstallFailed, domain.WorkloadFailed{}, OnInstallFailedCleanupDone)
	t.SetOrAddHandler(appv1.InstallFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnInstallFailedUninstall)
}

const (
	installFailedReasonCleanedUp       = "InstallFailureCleanedUp"
	installFailedMsgCleanedUp          = "install failure resources cleaned up"
	installFailedReasonCleanupRetrying = "InstallCleanupRetrying"
	installFailedMsgCleanupRetrying    = "install cleanup failed; retrying"

	// installFailedCleanupRetryBackoff spaces out cleanup attempts.
	// A stuck namespace already spends the full namespace wait per
	// attempt; the backoff only keeps fast errors from spinning.
	installFailedCleanupRetryBackoff = time.Minute
)

// OnInstallFailedTick runs install-failure cleanup in the background.
// An empty phase starts it. An in-progress phase recovers the
// workload after a restart. A retry phase waits out its backoff and
// then starts another attempt. A done phase leaves InstallFailed for
// Uninstalled so the app can be installed again.
func OnInstallFailedTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{}
	}
	phase := ""
	if w.AM.Status.Payload != nil {
		phase = w.AM.Status.Payload[PayloadInstallFailedCleanup]
	}
	switch phase {
	case PhaseInstallFailedCleanupDone:
		return finishInstallFailedToUninstalled(installFailedMsgCleanedUp, installFailedReasonCleanedUp, "Normal")
	case PhaseInstallFailedCleanup:
		return recoverWorkload(w, cleanupInstallFailureWorkload())
	case PhaseInstallFailedCleanupRetry:
		if wait := installFailedRetryWait(w); wait > 0 {
			return domain.Result{Effects: []domain.Effect{domain.ScheduleTick{After: wait}}}
		}
		return startUnlessSameKind(w, domain.WorkloadCleanupInstallFailure, startInstallFailedCleanup(w))
	default:
		return startInstallFailedCleanup(w)
	}
}

// installFailedRetryWait returns how long a retry phase must still
// wait. A missing or unparsable stamp retries immediately.
func installFailedRetryWait(w domain.World) time.Duration {
	at, err := time.Parse(time.RFC3339, w.AM.Status.Payload[PayloadInstallFailedCleanupRetryAt])
	if err != nil {
		return 0
	}
	return at.Sub(w.Now)
}

func startInstallFailedCleanup(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallFailed,
				Message: installingMsgCleanupStarted,
				Reason:  installingReasonCleanupStarted,
				Payload: map[string]string{
					PayloadInstallFailedCleanup: PhaseInstallFailedCleanup,
				},
			}, DeletePayloadKeys: []string{PayloadInstallFailedCleanupRetryAt}},
			cleanupInstallFailureWorkload(),
		},
	}
}

// OnInstallFailedCleanupDone handles the cleanup workload's result.
// Success means the namespace is gone, so the AM becomes Uninstalled.
// Failure keeps InstallFailed and schedules another attempt: landing
// Uninstalled with the namespace still present would allow a
// reinstall into it and block appmgr GC.
func OnInstallFailedCleanupDone(_ context.Context, w domain.World) domain.Result {
	failed, ok := w.Action.(domain.WorkloadFailed)
	if !ok {
		return finishInstallFailedToUninstalled(installFailedMsgCleanedUp, installFailedReasonCleanedUp, "Normal")
	}
	msg := installFailedMsgCleanupRetrying
	if failed.Reason != "" {
		msg = fmt.Sprintf("%s: %s", installFailedMsgCleanupRetrying, failed.Reason)
	}
	retryAt := w.Now.Add(installFailedCleanupRetryBackoff).UTC().Format(time.RFC3339)
	return domain.Result{
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallFailed,
				Message: msg,
				Reason:  installFailedReasonCleanupRetrying,
				Payload: map[string]string{
					PayloadInstallFailedCleanup:        PhaseInstallFailedCleanupRetry,
					PayloadInstallFailedCleanupRetryAt: retryAt,
				},
			}},
			domain.EmitEvent{Type: "Warning", Reason: installFailedReasonCleanupRetrying, Message: msg},
			domain.ScheduleTick{After: installFailedCleanupRetryBackoff},
		},
	}
}

// finishInstallFailedToUninstalled is the only exit from InstallFailed
// that the background cleanup uses. Payload from the failed install
// is cleared; the failure itself already lives on the OpRecord written
// when the AM entered InstallFailed.
func finishInstallFailedToUninstalled(msg, reason, eventType string) domain.Result {
	return domain.Result{
		Next: appv1.Uninstalled,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Uninstalled,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{},
			}, DeletePayloadKeys: []string{
				PayloadInstallFailedCleanup,
				PayloadInstallFailedCleanupRetryAt,
				PayloadInstallingPhase,
				PayloadInstallingCancelingPhase,
				PayloadInstallingCancelingTerminal,
				PayloadInstallingFailMessage,
				PayloadInstallingFailReason,
			}},
			domain.EmitEvent{Type: eventType, Reason: reason, Message: msg},
		},
	}
}

// OnInstallFailedUninstall abandons the failed install by
// triggering the uninstall pipeline (which cleans any partial
// helm release / namespace / etc. left behind).
func OnInstallFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// DownloadFailed
// -----------------------------------------------------------------
//
// Terminal state after image pull failure. Only legal op is
// InstallOp — retry the full pipeline (pull + install) from
// Pending. UninstallOp is NOT legal because a DownloadFailed
// AM has no helm release yet — nothing to uninstall.
func registerDownloadFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.DownloadFailed, appv1.DownloadFailed)
	t.SetOrAddHandler(appv1.DownloadFailed, domain.Tick{}, OnDownloadFailedTick)
	t.SetOrAddHandler(appv1.DownloadFailed, domain.TriggerOp{Op: appv1.InstallOp}, OnDownloadFailedInstall)
}

// OnDownloadFailedTick is a no-op — see noopTerminalTick.
func OnDownloadFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnDownloadFailedInstall retries the failed pipeline from Pending.
func OnDownloadFailedInstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Pending, terminalReasonInstall, terminalMsgInstall)
}

// -----------------------------------------------------------------
// UninstallFailed
// -----------------------------------------------------------------
//
// Terminal state after uninstall failure. Only legal op is
// UninstallOp — retry the uninstall (the target namespace's
// GC might have unstuck, the helm release might be re-deletable
// now, etc.).
func registerUninstallFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.UninstallFailed, appv1.UninstallFailed)
	t.SetOrAddHandler(appv1.UninstallFailed, domain.Tick{}, OnUninstallFailedTick)
	t.SetOrAddHandler(appv1.UninstallFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnUninstallFailedUninstall)
}

// OnUninstallFailedTick is a no-op — see noopTerminalTick.
func OnUninstallFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnUninstallFailedUninstall retries the failed uninstall.
func OnUninstallFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// StopFailed
// -----------------------------------------------------------------
//
// Terminal state after stop failure. Two legal ops:
//   - StopOp:      retry the stop.
//   - UninstallOp: abandon — go straight to uninstall.
//
// UpgradeOp is rejected. A failed stop must reach Running or
// Stopped before an upgrade can start.
func registerStopFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.StopFailed, appv1.StopFailed)
	t.SetOrAddHandler(appv1.StopFailed, domain.Tick{}, OnStopFailedTick)
	t.SetOrAddHandler(appv1.StopFailed, domain.TriggerOp{Op: appv1.StopOp}, OnStopFailedStop)
	t.SetOrAddHandler(appv1.StopFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnStopFailedUninstall)
}

// OnStopFailedTick is a no-op — see noopTerminalTick.
func OnStopFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnStopFailedStop retries the failed stop. A pod-abnormal reason
// stamped on the AM replaces the generic StopRequested.
func OnStopFailedStop(_ context.Context, w domain.World) domain.Result {
	return stopToStopping(w, terminalReasonStop, terminalMsgStop)
}

// OnStopFailedUninstall abandons the failed stop by triggering
// the uninstall pipeline.
func OnStopFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// ResumeFailed
// -----------------------------------------------------------------
//
// Terminal state after resume failure. Three legal ops:
//   - ResumeOp:    retry the resume.
//   - ApplyEnvOp:  apply env changes and re-resume.
//   - UninstallOp: abandon.
//
// UpgradeOp is rejected, same as StopFailed.
func registerResumeFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.ResumeFailed, appv1.ResumeFailed)
	t.SetOrAddHandler(appv1.ResumeFailed, domain.Tick{}, OnResumeFailedTick)
	t.SetOrAddHandler(appv1.ResumeFailed, domain.TriggerOp{Op: appv1.ResumeOp}, OnResumeFailedResume)
	t.SetOrAddHandler(appv1.ResumeFailed, domain.TriggerOp{Op: appv1.ApplyEnvOp}, OnResumeFailedApplyEnv)
	t.SetOrAddHandler(appv1.ResumeFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnResumeFailedUninstall)
}

// OnResumeFailedTick is a no-op — see noopTerminalTick.
func OnResumeFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnResumeFailedResume retries the failed resume.
func OnResumeFailedResume(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Resuming, terminalReasonResume, terminalMsgResume)
}

// OnResumeFailedApplyEnv triggers an env apply from ResumeFailed.
func OnResumeFailedApplyEnv(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.ApplyingEnv, terminalReasonApplyEnv, terminalMsgApplyEnv)
}

// OnResumeFailedUninstall abandons the failed resume.
func OnResumeFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// UpgradeFailed
// -----------------------------------------------------------------
//
// Terminal state after upgrade failure. Two legal ops:
//   - UpgradeOp:   retry the upgrade.
//   - UninstallOp: abandon — go straight to uninstall.
//
// The default table also declares a UpgradeFailed → Stopping
// edge with a `// dead edge today` comment; not wired here.
func registerUpgradeFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradeFailed, appv1.UpgradeFailed)
	t.SetOrAddHandler(appv1.UpgradeFailed, domain.Tick{}, OnUpgradeFailedTick)
	t.SetOrAddHandler(appv1.UpgradeFailed, domain.TriggerOp{Op: appv1.UpgradeOp}, OnUpgradeFailedUpgrade)
	t.SetOrAddHandler(appv1.UpgradeFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnUpgradeFailedUninstall)
}

// OnUpgradeFailedTick is a no-op — see noopTerminalTick.
func OnUpgradeFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnUpgradeFailedUpgrade retries the failed upgrade.
func OnUpgradeFailedUpgrade(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.UpgradePending, terminalReasonUpgrade, terminalMsgUpgrade)
}

// OnUpgradeFailedUninstall abandons the failed upgrade.
func OnUpgradeFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// ApplyEnvFailed
// -----------------------------------------------------------------
//
// Terminal state after env-apply failure. Three legal ops:
//   - ApplyEnvOp:  retry the env apply.
//   - StopOp:      stop the (possibly-partially-env-applied) app.
//   - UninstallOp: abandon.
func registerApplyEnvFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.ApplyEnvFailed, appv1.ApplyEnvFailed)
	t.SetOrAddHandler(appv1.ApplyEnvFailed, domain.Tick{}, OnApplyEnvFailedTick)
	t.SetOrAddHandler(appv1.ApplyEnvFailed, domain.TriggerOp{Op: appv1.ApplyEnvOp}, OnApplyEnvFailedApplyEnv)
	t.SetOrAddHandler(appv1.ApplyEnvFailed, domain.TriggerOp{Op: appv1.StopOp}, OnApplyEnvFailedStop)
	t.SetOrAddHandler(appv1.ApplyEnvFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnApplyEnvFailedUninstall)
}

// OnApplyEnvFailedTick is a no-op — see noopTerminalTick.
func OnApplyEnvFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnApplyEnvFailedApplyEnv retries the failed env apply.
func OnApplyEnvFailedApplyEnv(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.ApplyingEnv, terminalReasonApplyEnv, terminalMsgApplyEnv)
}

// OnApplyEnvFailedStop stops the app after env-apply failure. A
// pod-abnormal reason stamped on the AM replaces the generic
// StopRequested.
func OnApplyEnvFailedStop(_ context.Context, w domain.World) domain.Result {
	return stopToStopping(w, terminalReasonStop, terminalMsgStop)
}

// OnApplyEnvFailedUninstall abandons the failed env-apply.
func OnApplyEnvFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// =================================================================
// "Canceled" family — install-pipeline cancel terminals
// =================================================================
//
// Reachable via the cancel_pipeline.go handlers. Only legal op is
// InstallOp (re-install a previously-canceled AM). The default
// table declares the *Canceled → Pending edge for each; we register
// the InstallOp handler that fires it.
//
// No UninstallOp because a canceled AM has no helm release / no
// live resources — the cancel pipeline already cleaned everything
// up. Legacy allows only InstallOp here (see opAllowed table).

// -----------------------------------------------------------------
// PendingCanceled
// -----------------------------------------------------------------
//
// Terminal after a successful PendingCanceling. Re-install goes
// to Pending (fresh pipeline start). Legacy correspondence: no
// dedicated *_app.go file — the state is a pure terminal, the
// install re-target is driven by the opTarget table which we
// ported into table_default.
func registerPendingCanceled(t *domain.TransitionTable) {
	t.AddReachable(appv1.PendingCanceled, appv1.PendingCanceled)
	t.SetOrAddHandler(appv1.PendingCanceled, domain.Tick{}, OnPendingCanceledTick)
	t.SetOrAddHandler(appv1.PendingCanceled, domain.TriggerOp{Op: appv1.InstallOp}, OnPendingCanceledInstall)
}

// OnPendingCanceledTick is a no-op — see noopTerminalTick.
func OnPendingCanceledTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnPendingCanceledInstall re-installs a previously-pending-canceled app.
func OnPendingCanceledInstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Pending, terminalReasonInstall, terminalMsgInstall)
}

// -----------------------------------------------------------------
// DownloadingCanceled
// -----------------------------------------------------------------
//
// Symmetric to PendingCanceled — same shape, different source.
// Re-install triggers a fresh Pending pipeline (the previously-
// downloaded image blob may or may not still be cached; that's
// an ImagePuller concern, not a state-machine one).
func registerDownloadingCanceled(t *domain.TransitionTable) {
	t.AddReachable(appv1.DownloadingCanceled, appv1.DownloadingCanceled)
	t.SetOrAddHandler(appv1.DownloadingCanceled, domain.Tick{}, OnDownloadingCanceledTick)
	t.SetOrAddHandler(appv1.DownloadingCanceled, domain.TriggerOp{Op: appv1.InstallOp}, OnDownloadingCanceledInstall)
}

// OnDownloadingCanceledTick is a no-op — see noopTerminalTick.
func OnDownloadingCanceledTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnDownloadingCanceledInstall re-installs a previously-download-canceled app.
func OnDownloadingCanceledInstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Pending, terminalReasonInstall, terminalMsgInstall)
}

// -----------------------------------------------------------------
// InstallingCanceled
// -----------------------------------------------------------------
//
// Symmetric to PendingCanceled — this AM had a partial helm release
// which the InstallingCanceling handler tore down (see
// cancel_pipeline.go). Re-install starts fresh from Pending.
func registerInstallingCanceled(t *domain.TransitionTable) {
	t.AddReachable(appv1.InstallingCanceled, appv1.InstallingCanceled)
	t.SetOrAddHandler(appv1.InstallingCanceled, domain.Tick{}, OnInstallingCanceledTick)
	t.SetOrAddHandler(appv1.InstallingCanceled, domain.TriggerOp{Op: appv1.InstallOp}, OnInstallingCanceledInstall)
}

// OnInstallingCanceledTick is a no-op — see noopTerminalTick.
func OnInstallingCanceledTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnInstallingCanceledInstall re-installs a previously-install-canceled app.
func OnInstallingCanceledInstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Pending, terminalReasonInstall, terminalMsgInstall)
}

// =================================================================
// "CancelFailed" family — cancel-retry / abandon terminals
// =================================================================
//
// Reachable when a cancel_pipeline.go or cancel_handoff.go handler's
// CallWorkload fails. CancelOp routes back into the matching
// *Canceling state (retry); UninstallOp (where declared)
// abandons the AM via the full uninstall pipeline.

// -----------------------------------------------------------------
// PendingCancelFailed
// -----------------------------------------------------------------
//
// Only legal op is CancelOp — a PendingCancel AM has no helm
// release, so UninstallOp is not declared in the default table
// (opAllowed[PendingCancelFailed] = {CancelOp: true}). Retry goes
// back into PendingCanceling for another wait-ns attempt.
func registerPendingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.PendingCancelFailed, appv1.PendingCancelFailed)
	t.SetOrAddHandler(appv1.PendingCancelFailed, domain.Tick{}, OnPendingCancelFailedTick)
	t.SetOrAddHandler(appv1.PendingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnPendingCancelFailedCancel)
}

// OnPendingCancelFailedTick is a no-op — see noopTerminalTick.
func OnPendingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnPendingCancelFailedCancel retries the failed pending-cancel.
func OnPendingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.PendingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryPending)
}

// -----------------------------------------------------------------
// DownloadingCancelFailed
// -----------------------------------------------------------------
//
// Same shape as PendingCancelFailed — CancelOp-only. A downloading
// cancel that failed also has no helm release to uninstall.
func registerDownloadingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.DownloadingCancelFailed, appv1.DownloadingCancelFailed)
	t.SetOrAddHandler(appv1.DownloadingCancelFailed, domain.Tick{}, OnDownloadingCancelFailedTick)
	t.SetOrAddHandler(appv1.DownloadingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnDownloadingCancelFailedCancel)
}

// OnDownloadingCancelFailedTick is a no-op — see noopTerminalTick.
func OnDownloadingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnDownloadingCancelFailedCancel retries the failed downloading-cancel.
func OnDownloadingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.DownloadingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryDownloading)
}

// -----------------------------------------------------------------
// InstallingCancelFailed
// -----------------------------------------------------------------
//
// Two ops: CancelOp (retry the helm-uninstall + wait-ns pipeline)
// or UninstallOp (abandon via the full Uninstalling flow). The
// AM here has a partial helm release from the failed install, so
// UninstallOp is a real escape hatch (unlike PendingCancelFailed).
func registerInstallingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.InstallingCancelFailed, appv1.InstallingCancelFailed)
	t.SetOrAddHandler(appv1.InstallingCancelFailed, domain.Tick{}, OnInstallingCancelFailedTick)
	t.SetOrAddHandler(appv1.InstallingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnInstallingCancelFailedCancel)
	t.SetOrAddHandler(appv1.InstallingCancelFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnInstallingCancelFailedUninstall)
}

// OnInstallingCancelFailedTick is a no-op — see noopTerminalTick.
func OnInstallingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnInstallingCancelFailedCancel retries the failed installing-cancel.
func OnInstallingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	res := terminalOpTransition(w, appv1.InstallingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryInstalling)
	for i, effect := range res.Effects {
		patch, ok := effect.(domain.PatchAM)
		if !ok {
			continue
		}
		// An ordinary install failure keeps its requested terminal
		// and original diagnostic across cleanup retries. Reset only
		// the failed phase so InstallingCanceling restarts at helm
		// uninstall. User cancellation carries no intent.
		patch.Status.Payload = carryInstallFailureIntent(w, nil)
		patch.DeletePayloadKeys = []string{PayloadInstallingCancelingPhase}
		res.Effects[i] = patch
		break
	}
	return res
}

// OnInstallingCancelFailedUninstall abandons the AM via the full
// Uninstalling pipeline — same route as InstallFailed's UninstallOp.
func OnInstallingCancelFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// InitializingCancelFailed
// -----------------------------------------------------------------
//
// Timeout terminal of InitializingCanceling. CancelOp retries the
// handoff to Stopping. UninstallOp abandons through Uninstalling.
// Tick stays here.
func registerInitializingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.InitializingCancelFailed, appv1.InitializingCancelFailed)
	t.SetOrAddHandler(appv1.InitializingCancelFailed, domain.Tick{}, OnInitializingCancelFailedTick)
	t.SetOrAddHandler(appv1.InitializingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnInitializingCancelFailedCancel)
	t.SetOrAddHandler(appv1.InitializingCancelFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnInitializingCancelFailedUninstall)
}

// OnInitializingCancelFailedTick is a no-op — see noopTerminalTick.
func OnInitializingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnInitializingCancelFailedCancel retries the failed initializing-cancel.
func OnInitializingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.InitializingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryInitializing)
}

// OnInitializingCancelFailedUninstall abandons the failed initializing-cancel.
func OnInitializingCancelFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// UpgradingCancelFailed
// -----------------------------------------------------------------
//
// Two ops: CancelOp (retry the handoff-to-Stopping) or UninstallOp
// (abandon via full Uninstalling). An upgrade-cancel that failed
// leaves the AM with a partially-upgraded helm release; both retry
// (via Stopping's own tear-down) and abandon are legitimate.
func registerUpgradingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradingCancelFailed, appv1.UpgradingCancelFailed)
	t.SetOrAddHandler(appv1.UpgradingCancelFailed, domain.Tick{}, OnUpgradingCancelFailedTick)
	t.SetOrAddHandler(appv1.UpgradingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradingCancelFailedCancel)
	t.SetOrAddHandler(appv1.UpgradingCancelFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnUpgradingCancelFailedUninstall)
}

// OnUpgradingCancelFailedTick is a no-op — see noopTerminalTick.
func OnUpgradingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnUpgradingCancelFailedCancel retries the failed upgrading-cancel.
func OnUpgradingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.UpgradingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryUpgrading)
}

// OnUpgradingCancelFailedUninstall abandons the failed upgrade-cancel.
func OnUpgradingCancelFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// ApplyingEnvCancelFailed
// -----------------------------------------------------------------
//
// Two ops: CancelOp (retry) or UninstallOp (abandon). Same
// rationale as UpgradingCancelFailed — a partially-env-applied
// AM's helm release is still there.
func registerApplyingEnvCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.ApplyingEnvCancelFailed, appv1.ApplyingEnvCancelFailed)
	t.SetOrAddHandler(appv1.ApplyingEnvCancelFailed, domain.Tick{}, OnApplyingEnvCancelFailedTick)
	t.SetOrAddHandler(appv1.ApplyingEnvCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnApplyingEnvCancelFailedCancel)
	t.SetOrAddHandler(appv1.ApplyingEnvCancelFailed, domain.TriggerOp{Op: appv1.UninstallOp}, OnApplyingEnvCancelFailedUninstall)
}

// OnApplyingEnvCancelFailedTick is a no-op — see noopTerminalTick.
func OnApplyingEnvCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnApplyingEnvCancelFailedCancel retries the failed apply-env-cancel.
func OnApplyingEnvCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.ApplyingEnvCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryApplyingEnv)
}

// OnApplyingEnvCancelFailedUninstall abandons the failed apply-env-cancel.
func OnApplyingEnvCancelFailedUninstall(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.Uninstalling, terminalReasonUninstall, terminalMsgUninstall)
}

// -----------------------------------------------------------------
// ResumingCancelFailed
// -----------------------------------------------------------------
//
// Kept for ApplicationManagers already persisted in this state.
// ResumingCanceling does not write it. Tick stays here; CancelOp
// re-enters ResumingCanceling, which then hands off to Stopping.
// UninstallOp is not allowed, matching OperationAllowedInState.
func registerResumingCancelFailed(t *domain.TransitionTable) {
	t.AddReachable(appv1.ResumingCancelFailed, appv1.ResumingCancelFailed)
	t.SetOrAddHandler(appv1.ResumingCancelFailed, domain.Tick{}, OnResumingCancelFailedTick)
	t.SetOrAddHandler(appv1.ResumingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnResumingCancelFailedCancel)
}

// OnResumingCancelFailedTick is a no-op — see noopTerminalTick.
func OnResumingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

// OnResumingCancelFailedCancel retries the failed resume-cancel.
func OnResumingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.ResumingCanceling, terminalReasonCancelRetry, terminalMsgCancelRetryResuming)
}
