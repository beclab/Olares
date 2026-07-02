// Package handlers, cancel-pipeline section.
//
// This file collects the three install-pipeline cancel handlers.
// Unlike cancel_handoff.go's soft-handoff pattern (Initializing/
// Resuming/Upgrading/ApplyingEnv → Stopping), these three states
// have to CLEAN UP resources themselves — the source state was
// mid-install with no downstream state that will do it for them.
//
// Structural shape:
//
//	Source state    →  *Canceling state         →  *Canceled terminal
//	────────────────────────────────────────────────────────────────
//	Pending         →  PendingCanceling (Ptn A) →  PendingCanceled
//	Downloading     →  DownloadingCanceling (B) →  DownloadingCanceled
//	Installing      →  InstallingCanceling (C)  →  InstallingCanceled
//
// Pattern A (PendingCanceling) — single phase:
//
//	phase="" → dispatch CallWorkload{WaitForNamespaceDeleted};
//	           stamp phase="wait-ns"
//	WorkloadReady @ wait-ns → PendingCanceled (no OpRecord)
//	WorkloadFailed          → PendingCancelFailed (with OpRecord)
//
//	(The WaitForNamespaceDeleted adapter itself issues the delete
//	before starting the poll — see ports/workload.go
//	WaitForNamespaceDeleted doc. No separate delete-ns effect.)
//
// Pattern B (DownloadingCanceling) — two phases:
//
//	phase="" → dispatch CallWorkload{CancelImagePull};
//	           stamp phase="cancel-pull"
//	WorkloadReady @ cancel-pull → dispatch CallWorkload{WaitForNamespaceDeleted};
//	                              stamp phase="wait-ns"
//	WorkloadReady @ wait-ns     → DownloadingCanceled (no OpRecord)
//	WorkloadFailed              → DownloadingCancelFailed (with OpRecord)
//
// Pattern C (InstallingCanceling) — two phases:
//
//	phase="" → dispatch CallWorkload{Uninstall};
//	           stamp phase="helm-uninstall"
//	WorkloadReady @ helm-uninstall → dispatch CallWorkload{WaitForNamespaceDeleted};
//	                                  stamp phase="wait-ns"
//	WorkloadReady @ wait-ns         → InstallingCanceled
//	                                  (WITH OpRecord — legacy
//	                                  finishCancelToCanceled(..., true))
//	                                  or InstallFailed when the AM
//	                                  carries the install-failure
//	                                  terminal intent
//	WorkloadFailed                  → InstallingCancelFailed (with OpRecord)
//
// Pattern C serves two callers. A user cancel enters from
// (Installing, CancelOp) and finishes at InstallingCanceled. An
// install failure enters from installing.go's failure branch with
// PayloadInstallingCancelingTerminal = InstallFailed, so the same
// tear-down finishes at InstallFailed carrying the original
// diagnostic. Only the terminal write differs.
//
// Legacy correspondence:
//   - pkg/appstate/pending_canceling_app.go
//   - pkg/appstate/download_canceling_app.go
//   - pkg/appstate/installing_canceling_app.go
//
// Legacy quirks preserved:
//
//   - Pattern A/B write NO OpRecord on success (only the
//     Installing variant does — matches legacy
//     finishCancelToCanceled(..., withOpRecord=true) parameter).
//     Downstream OpRecord consumers (NATS push, UI status banner)
//     see the same event stream shape as the legacy stack.
//
//   - The WaitForNamespaceDeleted adapter respects
//     IsProtectedNamespace internally (per port doc); the handler
//     never checks it. If a protected AM ends up in *Canceling,
//     the adapter returns success immediately and the pipeline
//     completes with no side effect — same as legacy line 96 /
//     line 112 skipping the poll for protected namespaces.
//
// SKELETON SCOPE — deliberately NOT ported:
//
//   - The InstallingCanceling shared-chart branch (legacy
//     installing_canceling_app.go:107-150 — detects sharedInstalled
//     across sub-charts and picks between UninstallAll and
//     Uninstall). Both funnel through CallWorkload{Uninstall}
//     here; the adapter's Uninstall implementation is where
//     UninstallAll dispatch belongs, gated by an annotation the
//     handler could stamp into WorkloadRequest.
//
//   - Structured cancel-vs-retry classification. Today
//     WorkloadFailed unconditionally lands in *CancelFailed. A
//     follow-up will branch on the classifier ("network hiccup" →
//     re-dispatch same phase, "hard failure" → *CancelFailed) once
//     the ImagePuller / WorkloadOps errors carry a structured
//     Reason enum.
//
//   - Cancel-during-cancel-in-flight. If the user hits cancel
//     again while a *Canceling is running (say, wait-ns poll is
//     dragging on), the current design silently no-ops (Runner
//     returns ErrNoTransition since *Canceling has no CancelOp
//     handler declared). A follow-up could route it to a
//     re-dispatch of the current phase. Not a common op today.
package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Per-state message + reason constants. Kept string-literal (not
// concatenated at runtime) so log-grepping picks them up. Same
// convention as uninstalling.go / installing.go.
const (
	// Reason strings emitted on the source→*Canceling edge (user-
	// initiated cancel — Normal event; no OpRecord).
	cancelPipelineReasonRequested    = "CancelRequested"
	cancelPipelineMsgFromPending     = "user requested cancel; transitioning to PendingCanceling"
	cancelPipelineMsgFromDownloading = "user requested cancel; transitioning to DownloadingCanceling"
	cancelPipelineMsgFromInstalling  = "user requested cancel; transitioning to InstallingCanceling"

	// Reason strings emitted at phase advance (Normal event).
	cancelPipelineReasonCancelPullStarted    = "CancelPullStarted"
	cancelPipelineMsgCancelPullStarted       = "canceling image pull"
	cancelPipelineReasonHelmUninstallStarted = "CancelHelmUninstallStarted"
	cancelPipelineMsgHelmUninstallStarted    = "tearing down partial helm release"
	cancelPipelineReasonWaitNSStarted        = "CancelWaitNSStarted"
	cancelPipelineMsgWaitNSStarted           = "waiting for namespace garbage collection"

	// Reason strings emitted at the terminal *Canceled state
	// (Normal event; OpRecord only for InstallingCanceled per
	// legacy withOpRecord=true).
	cancelPipelineReasonCompletedPending     = "PendingCancelCompleted"
	cancelPipelineMsgCompletedPending        = "pending install canceled"
	cancelPipelineReasonCompletedDownloading = "DownloadingCancelCompleted"
	cancelPipelineMsgCompletedDownloading    = "downloading canceled"
	cancelPipelineReasonCompletedInstalling  = "InstallingCancelCompleted"
	cancelPipelineMsgCompletedInstalling     = "install canceled"

	// Cancel-fail fallback reasons. Runner-supplied
	// LandedReason overrides these — see failureFromAction for the
	// full priority ladder. These are the last-resort fallbacks
	// when a WorkloadFailed action arrives with empty (Reason,
	// Message).
	cancelPipelineReasonCancelPullFailed    = "ImageCancelFailed"
	cancelPipelineReasonHelmUninstallFailed = "HelmUninstallFailed"
	cancelPipelineReasonWaitNSFailed        = "NamespaceGoneWaitFailed"
)

// =================================================================
// Pending → PendingCanceling → PendingCanceled | PendingCancelFailed
// =================================================================

// registerPendingCanceling wires:
//
//   - (Pending, TriggerOp{CancelOp}) → OnPendingCancel — replaces the
//     source-state placeholder.
//   - (PendingCanceling, Tick{}, WorkloadReady{}, WorkloadFailed{})
//     — three handlers on the *Canceling state itself.
//
// AddReachable declares the PendingCanceling self-loop for the
// phase-progression Result{Next:""} the Tick handler emits.
func registerPendingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.PendingCanceling, appv1.PendingCanceling)
	t.SetOrAddHandler(appv1.Pending, domain.TriggerOp{Op: appv1.CancelOp}, OnPendingCancel)
	t.SetOrAddHandler(appv1.PendingCanceling, domain.Tick{}, OnPendingCancelingTick)
	t.SetOrAddHandler(appv1.PendingCanceling, domain.WorkloadReady{}, OnPendingCancelingWorkloadReady)
	t.SetOrAddHandler(appv1.PendingCanceling, domain.WorkloadFailed{}, OnPendingCancelingWorkloadFailed)
}

// OnPendingCancel is the CancelOp handler on Pending. Flips to
// PendingCanceling; the *Canceling handler's Tick picks up the
// wait-ns dispatch on the next Reconcile.
func OnPendingCancel(_ context.Context, w domain.World) domain.Result {
	return sourceToCanceling(w, appv1.PendingCanceling, cancelPipelineMsgFromPending)
}

// OnPendingCancelingTick is the phase-driver Tick for
// PendingCanceling. Single phase (wait-ns).
func OnPendingCancelingTick(_ context.Context, w domain.World) domain.Result {
	switch pendingCancelingPhase(w) {
	case "":
		return startUnlessSameKind(w, domain.WorkloadWaitNamespaceGone, startCancelWaitNS(w, appv1.PendingCanceling, PayloadPendingCancelingPhase))

	case PhaseCancelWaitNS:
		return recoverWorkload(w, waitNamespaceGoneWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("pending-canceling: unknown %s=%q; expected one of {\"\", %q}",
				PayloadPendingCancelingPhase, pendingCancelingPhase(w), PhaseCancelWaitNS),
		}
	}
}

// OnPendingCancelingWorkloadReady advances phase→terminal.
func OnPendingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := pendingCancelingPhase(w)
	switch phase {
	case PhaseCancelWaitNS:
		return finishToCanceled(w,
			appv1.PendingCanceled,
			cancelPipelineReasonCompletedPending,
			cancelPipelineMsgCompletedPending,
			false, // Pattern A: no OpRecord
		)

	case "":
		return domain.Result{
			Err: fmt.Errorf("pending-canceling: WorkloadReady with empty %s; every CallWorkload emitted from PendingCanceling must stamp the phase first",
				PayloadPendingCancelingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("pending-canceling: WorkloadReady with unknown %s=%q", PayloadPendingCancelingPhase, phase),
		}
	}
}

// OnPendingCancelingWorkloadFailed transitions to PendingCancelFailed.
func OnPendingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	return failToCancelFailed(w,
		appv1.PendingCancelFailed,
		cancelPipelineReasonWaitNSFailed, // only phase in Pattern A
	)
}

// =================================================================
// Downloading → DownloadingCanceling →
//   DownloadingCanceled | DownloadingCancelFailed
// =================================================================

// registerDownloadingCanceling — two phases (cancel-pull, wait-ns).
func registerDownloadingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.DownloadingCanceling, appv1.DownloadingCanceling)
	t.SetOrAddHandler(appv1.Downloading, domain.TriggerOp{Op: appv1.CancelOp}, OnDownloadingCancel)
	t.SetOrAddHandler(appv1.DownloadingCanceling, domain.Tick{}, OnDownloadingCancelingTick)
	t.SetOrAddHandler(appv1.DownloadingCanceling, domain.WorkloadReady{}, OnDownloadingCancelingWorkloadReady)
	t.SetOrAddHandler(appv1.DownloadingCanceling, domain.WorkloadFailed{}, OnDownloadingCancelingWorkloadFailed)
}

// OnDownloadingCancel is the CancelOp handler on Downloading.
func OnDownloadingCancel(_ context.Context, w domain.World) domain.Result {
	return sourceToCanceling(w, appv1.DownloadingCanceling, cancelPipelineMsgFromDownloading)
}

// OnDownloadingCancelingTick is the phase-driver Tick for
// DownloadingCanceling. Two phases: cancel-pull then wait-ns.
func OnDownloadingCancelingTick(_ context.Context, w domain.World) domain.Result {
	phase := downloadingCancelingPhase(w)
	switch phase {
	case "":
		return startUnlessSameKind(w, domain.WorkloadCancelImagePull, startCancelImagePull(w))

	case PhaseCancelImagePull:
		return recoverWorkload(w, cancelImagePullWorkload())
	case PhaseCancelWaitNS:
		return recoverWorkload(w, waitNamespaceGoneWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("downloading-canceling: unknown %s=%q; expected one of {\"\", %q, %q}",
				PayloadDownloadingCancelingPhase, phase, PhaseCancelImagePull, PhaseCancelWaitNS),
		}
	}
}

// OnDownloadingCancelingWorkloadReady advances phase→phase→terminal.
func OnDownloadingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := downloadingCancelingPhase(w)
	switch phase {
	case PhaseCancelImagePull:
		return startCancelWaitNS(w, appv1.DownloadingCanceling, PayloadDownloadingCancelingPhase)

	case PhaseCancelWaitNS:
		return finishToCanceled(w,
			appv1.DownloadingCanceled,
			cancelPipelineReasonCompletedDownloading,
			cancelPipelineMsgCompletedDownloading,
			false, // Pattern B: no OpRecord
		)

	case "":
		return domain.Result{
			Err: fmt.Errorf("downloading-canceling: WorkloadReady with empty %s; every CallWorkload emitted from DownloadingCanceling must stamp the phase first",
				PayloadDownloadingCancelingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("downloading-canceling: WorkloadReady with unknown %s=%q", PayloadDownloadingCancelingPhase, phase),
		}
	}
}

// OnDownloadingCancelingWorkloadFailed transitions to DownloadingCancelFailed.
func OnDownloadingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	fallback := cancelPipelineReasonWaitNSFailed
	if downloadingCancelingPhase(w) == PhaseCancelImagePull {
		fallback = cancelPipelineReasonCancelPullFailed
	}
	return failToCancelFailed(w, appv1.DownloadingCancelFailed, fallback)
}

// =================================================================
// Installing → InstallingCanceling →
//   InstallingCanceled | InstallingCancelFailed
// =================================================================

// registerInstallingCanceling — two phases (helm-uninstall, wait-ns).
func registerInstallingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.InstallingCanceling, appv1.InstallingCanceling)
	t.SetOrAddHandler(appv1.Installing, domain.TriggerOp{Op: appv1.CancelOp}, OnInstallingCancel)
	t.SetOrAddHandler(appv1.InstallingCanceling, domain.Tick{}, OnInstallingCancelingTick)
	t.SetOrAddHandler(appv1.InstallingCanceling, domain.WorkloadReady{}, OnInstallingCancelingWorkloadReady)
	t.SetOrAddHandler(appv1.InstallingCanceling, domain.WorkloadFailed{}, OnInstallingCancelingWorkloadFailed)
}

// OnInstallingCancel is the CancelOp handler on Installing.
func OnInstallingCancel(_ context.Context, w domain.World) domain.Result {
	return sourceToCanceling(w, appv1.InstallingCanceling, cancelPipelineMsgFromInstalling)
}

// OnInstallingCancelingTick is the phase-driver Tick for
// InstallingCanceling. Two phases: helm-uninstall then wait-ns.
func OnInstallingCancelingTick(_ context.Context, w domain.World) domain.Result {
	phase := installingCancelingPhase(w)
	switch phase {
	case "":
		return startUnlessSameKind(w, domain.WorkloadUninstall, startCancelHelmUninstall(w))

	case PhaseCancelHelmUninstall:
		return recoverWorkload(w, cancelUninstallWorkload())
	case PhaseCancelWaitNS:
		return recoverWorkload(w, waitNamespaceGoneWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("installing-canceling: unknown %s=%q; expected one of {\"\", %q, %q}",
				PayloadInstallingCancelingPhase, phase, PhaseCancelHelmUninstall, PhaseCancelWaitNS),
		}
	}
}

// OnInstallingCancelingWorkloadReady advances phase→phase→terminal.
// Pattern C: emits an OpRecord at the terminal step, matching
// legacy finishCancelToCanceled(..., withOpRecord=true).
//
// The terminal is whichever one the caller asked for. An install
// failure routed here stamped PayloadInstallingCancelingTerminal, so
// the same helm-uninstall + wait-ns cleanup lands on InstallFailed
// with the original diagnostic; a user cancel left the key unset and
// lands on InstallingCanceled.
func OnInstallingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := installingCancelingPhase(w)
	switch phase {
	case PhaseCancelHelmUninstall:
		return startCancelWaitNS(w, appv1.InstallingCanceling, PayloadInstallingCancelingPhase)

	case PhaseCancelWaitNS:
		if installingCancelingTerminal(w) == string(appv1.InstallFailed) {
			msg, reason := stashedInstallFailure(w)
			return landInstallFailed(w, msg, reason)
		}
		return finishToCanceled(w,
			appv1.InstallingCanceled,
			cancelPipelineReasonCompletedInstalling,
			cancelPipelineMsgCompletedInstalling,
			true, // Pattern C: WITH OpRecord (legacy withOpRecord=true)
		)

	case "":
		return domain.Result{
			Err: fmt.Errorf("installing-canceling: WorkloadReady with empty %s; every CallWorkload emitted from InstallingCanceling must stamp the phase first",
				PayloadInstallingCancelingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("installing-canceling: WorkloadReady with unknown %s=%q", PayloadInstallingCancelingPhase, phase),
		}
	}
}

// OnInstallingCancelingWorkloadFailed transitions to InstallingCancelFailed.
func OnInstallingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	fallback := cancelPipelineReasonWaitNSFailed
	if installingCancelingPhase(w) == PhaseCancelHelmUninstall {
		fallback = cancelPipelineReasonHelmUninstallFailed
	}
	res := failToCancelFailed(w, appv1.InstallingCancelFailed, fallback)
	for i, effect := range res.Effects {
		patch, ok := effect.(domain.PatchAM)
		if !ok {
			continue
		}
		// A failed cleanup is retryable. Preserve an ordinary
		// install failure's terminal intent + original diagnostic,
		// but remove the failed phase so the retry starts from helm
		// uninstall. A user cancel has no intent to carry.
		patch.Status.Payload = carryInstallFailureIntent(w, nil)
		patch.DeletePayloadKeys = []string{PayloadInstallingCancelingPhase}
		res.Effects[i] = patch
		break
	}
	return res
}

// =================================================================
// Shared helpers
// =================================================================

// sourceToCanceling is the shared shape for every source→*Canceling
// CancelOp handler in this file. Same shape as runningOpTransition
// but with an explicit CancelRequested reason.
//
// No OpRecord — the record for a successful cancel is written at
// the *Canceled terminal (or *CancelFailed for the failure branch),
// following the "record-on-completion" convention.
func sourceToCanceling(
	w domain.World,
	canceling appv1.ApplicationManagerState,
	msg string,
) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("cancel-pipeline: %s handler invoked with nil AM", canceling),
		}
	}
	return domain.Result{
		Next: canceling,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   canceling,
				Reason:  cancelPipelineReasonRequested,
				Message: msg,
				// Clear the source state's phase Payload (installing.phase
				// / downloading.phase) so a fresh *Canceling pipeline
				// starts without stale keys.
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  cancelPipelineReasonRequested,
				Message: msg,
			},
		},
	}
}

// startCancelImagePull is the Pattern B entry effect —
// CallWorkload{CancelImagePull} + stamp phase="cancel-pull".
func startCancelImagePull(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.DownloadingCanceling,
				Reason:  cancelPipelineReasonCancelPullStarted,
				Message: cancelPipelineMsgCancelPullStarted,
				Payload: map[string]string{PayloadDownloadingCancelingPhase: PhaseCancelImagePull},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  cancelPipelineReasonCancelPullStarted,
				Message: cancelPipelineMsgCancelPullStarted,
			},
			cancelImagePullWorkload(),
		},
	}
}

// cancelImagePullWorkload is shared with UpgradeDownloadingCanceling.
// Preempts is set because the pull this cancels is normally still in
// flight, holding the AM's workload claim.
func cancelImagePullWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadCancelImagePull,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: cancelPipelineReasonCancelPullFailed},
		Preempts:  true,
	}
}

// startCancelHelmUninstall is the Pattern C entry effect —
// CallWorkload{Uninstall} + stamp phase="helm-uninstall". Uses the
// same WorkloadUninstall Kind as the full-Uninstalling state; the
// adapter's Uninstall is idempotent so calling it on a partial
// install is safe.
func startCancelHelmUninstall(w domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallingCanceling,
				Reason:  cancelPipelineReasonHelmUninstallStarted,
				Message: cancelPipelineMsgHelmUninstallStarted,
				Payload: carryInstallFailureIntent(w, map[string]string{
					PayloadInstallingCancelingPhase: PhaseCancelHelmUninstall,
				}),
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  cancelPipelineReasonHelmUninstallStarted,
				Message: cancelPipelineMsgHelmUninstallStarted,
			},
			cancelUninstallWorkload(),
		},
	}
}

// cancelUninstallWorkload is Pattern C's first phase. Preempts is
// set because helm install / wait-startup is normally still holding
// the AM's claim when the user cancels.
func cancelUninstallWorkload() domain.CallWorkload {
	cw := uninstallWorkload()
	cw.Preempts = true
	return cw
}

// startCancelWaitNS is the shared final-phase effect for all three
// patterns. Dispatches CallWorkload{WorkloadWaitNamespaceGone} and
// stamps phase="wait-ns" under whichever payload key the caller's
// state uses.
//
// The `state` and `payloadKey` params take the *Canceling state
// name so a single implementation serves all three patterns
// without an internal switch. The PatchAM's State field is set
// to `state` to prevent an accidental cross-state landing if the
// Runner races (which it can't today — state is stable inside
// Reconcile — but this is a defensive belt-and-braces).
func startCancelWaitNS(
	w domain.World,
	state appv1.ApplicationManagerState,
	payloadKey string,
) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   state,
				Reason:  cancelPipelineReasonWaitNSStarted,
				Message: cancelPipelineMsgWaitNSStarted,
				Payload: carryInstallFailureIntent(w, map[string]string{
					payloadKey: PhaseCancelWaitNS,
				}),
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  cancelPipelineReasonWaitNSStarted,
				Message: cancelPipelineMsgWaitNSStarted,
			},
			waitNamespaceGoneWorkload(),
		},
	}
}

// finishToCanceled is the shared terminal-success effect bundle.
// withOpRecord toggles the OpRecord effect — only InstallingCanceled
// writes one, matching legacy finishCancelToCanceled(..., true).
//
// Effect ordering: AppendOpRecord precedes PatchAM (when the
// bool is true) so both land in the same PatchStatus RPC — same
// invariant as installing.finishInstall / uninstalling.finishToUninstalled.
// See installing.go for the full RV-conflict rationale.
func finishToCanceled(
	w domain.World,
	canceled appv1.ApplicationManagerState,
	reason, msg string,
	withOpRecord bool,
) domain.Result {
	patch := domain.PatchAM{Status: appv1.ApplicationManagerStatus{
		State:    canceled,
		Reason:   reason,
		Message:  msg,
		Progress: "100.00",
		Payload:  map[string]string{}, // clear phase; leaving *Canceling
	}}
	emit := domain.EmitEvent{
		Type:    "Normal",
		Reason:  reason,
		Message: msg,
	}

	effects := []domain.Effect{}
	if withOpRecord {
		now := metav1.NewTime(w.Now)
		effects = append(effects, domain.AppendOpRecord{Record: appv1.OpRecord{
			OpID:   w.AM.Status.OpID,
			OpType: w.AM.Status.OpType,
			// Human-readable cancel-completed message. Legacy
			// makeRecord takes the preserved Status.Message from
			// the *Canceling state (e.g. "Install canceled.
			// Operation by user."); the port doesn't
			// have that preserved message available synchronously
			// so we use the terminal-state message the handler
			// is about to patch in — same semantic content ("<op>
			// canceled") for audit / NATS push consumers.
			Message:   msg,
			Source:    w.AM.Spec.Source,
			Status:    canceled,
			StateTime: &now,
		}})
	}
	effects = append(effects, patch, emit)
	return domain.Result{
		Next:    canceled,
		Effects: effects,
	}
}

// failToCancelFailed is the shared *CancelFailed transition. Always
// writes an OpRecord (matches legacy — every cancel-failed path
// through updateStatus writes one, since the operator needs to
// see the failed cancel attempt in the audit trail regardless of
// whether the eventual retry succeeds).
//
// fallbackReason is the phase-specific reason baked in when the
// Runner's WorkloadResult provided no LandedReason (unusual —
// production adapters set one, fakes usually don't).
func failToCancelFailed(
	w domain.World,
	cancelFailed appv1.ApplicationManagerState,
	fallbackReason string,
) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = fmt.Sprintf("cancel failed in phase (state=%q)", cancelFailed)
	}
	if reason == "" {
		reason = fallbackReason
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    cancelFailed,
		StateTime: &now,
	}
	return domain.Result{
		Next: cancelFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   cancelFailed,
				Reason:  reason,
				Message: msg,
				Payload: map[string]string{}, // clear phase; leaving *Canceling
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// pendingCancelingPhase reads Payload[PayloadPendingCancelingPhase]
// with the standard nil-map guard.
func pendingCancelingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadPendingCancelingPhase]
}

// downloadingCancelingPhase reads Payload[PayloadDownloadingCancelingPhase].
func downloadingCancelingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadDownloadingCancelingPhase]
}

// installingCancelingPhase reads Payload[PayloadInstallingCancelingPhase].
func installingCancelingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadInstallingCancelingPhase]
}

// installingCancelingTerminal reads
// Payload[PayloadInstallingCancelingTerminal] — "" for a user cancel,
// string(appv1.InstallFailed) for a cleanup an install failure asked
// for.
func installingCancelingTerminal(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadInstallingCancelingTerminal]
}

// carryInstallFailureIntent re-stamps the terminal intent and the
// stashed install diagnostic onto a phase-advance patch so the
// InstallingCanceling pipeline still knows where to land after it
// finishes. Depending on the store's key-by-key Payload merge would
// work today, but every phase patch then silently relies on it; this
// keeps the invariant inside the handler that owns it.
//
// A no-op for the Pending / Downloading patterns and for a user
// cancel, neither of which ever carries these keys.
func carryInstallFailureIntent(w domain.World, payload map[string]string) map[string]string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return payload
	}
	for _, key := range []string{
		PayloadInstallingCancelingTerminal,
		PayloadInstallingFailMessage,
		PayloadInstallingFailReason,
	} {
		if v := w.AM.Status.Payload[key]; v != "" {
			if payload == nil {
				payload = make(map[string]string, 3)
			}
			payload[key] = v
		}
	}
	return payload
}
