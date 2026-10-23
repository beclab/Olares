// Package handlers, cancel-handoff section.
//
// This file collects the four related cancel-handoff paths. They
// share the source -> *Canceling state flip, but their in-flight
// workload handling and terminal destinations differ:
//
//	Source state          →  *Canceling state           →  Destination
//	─────────────────────────────────────────────────────────────────
//	Initializing          →  InitializingCanceling      →  Stopping
//	Resuming              →  ResumingCanceling          →  Stopping
//	Upgrading             →  UpgradingCanceling         →  UpgradeFailed
//	ApplyingEnv           →  ApplyingEnvCanceling       →  Stopping
//
// Current workload/claim semantics:
//
//   - Initializing flips to InitializingCanceling, whose next Tick
//     hands off directly to Stopping. Stopping's preempting workload
//     cancels any outstanding WaitForLaunch claim.
//   - Resuming and ApplyingEnv append CancelInFlight while their helm
//     upgrade is outstanding. Their *Canceling Tick stays put, and
//     WorkloadReady / WorkloadFailed hands off only after helm returns,
//     preventing Stopping's Scale(0) from overlapping that upgrade.
//   - Upgrading waits with CancelInFlight while its helm upgrade or
//     startup wait is outstanding, then lands UpgradeFailed. Image
//     pull cancel belongs to UpgradeDownloadingCanceling; Upgrading
//     is only reached after that pull has finished.
//
// Every OnXCancel writes the *Canceling state and a "cancel requested"
// event, but no OpRecord. The downstream terminal transition owns the
// completion record. Tick and workload-callback handlers then implement
// the per-state behavior above.
//
// Legacy correspondence: the legacy *_canceling_app.go paths were
// synchronous, so their helm calls could not overlap. CancelInFlight
// preserves that ordering now that Runner owns asynchronous workloads.
//
// SKELETON SCOPE — deliberately NOT ported:
//
//   - The non-upgrading *CancelFailed edges that represented a
//     Kubernetes status-write failure in legacy code are not emitted
//     by handlers. Runner handles PatchStatus conflicts and retries
//     instead. Those terminal states remain registered for existing
//     objects and administrative recovery operations.
package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// Package-local message constants. Kept per-source-state so
// operator dashboards can differentiate "cancel while installing"
// from "cancel while upgrading" at a glance without decoding the
// state field.
const (
	cancelHandoffReasonRequested = "CancelRequested"

	cancelHandoffMsgFromInitializing = "user requested cancel; transitioning through InitializingCanceling to Stopping"
	cancelHandoffMsgFromResuming     = "user requested cancel; transitioning through ResumingCanceling to Stopping"
	cancelHandoffMsgFromUpgrading    = "user requested cancel; transitioning through UpgradingCanceling to UpgradeFailed"
	cancelHandoffMsgFromApplyingEnv  = "user requested cancel; transitioning through ApplyingEnvCanceling to Stopping"

	cancelHandoffReasonToStopping = "CancelHandoffToStopping"
	cancelHandoffMsgToStopping    = "cancel accepted; handing off to Stopping for teardown"
)

// -----------------------------------------------------------------
// Shared helpers
// -----------------------------------------------------------------

// cancelToCanceling emits the (state=XCanceling, reason=CancelRequested,
// message=msg) patch + a Normal EmitEvent for the source→canceling
// transition. Shared across the four On*Cancel handlers because
// their divergence is data (target state, message) not logic —
// identical to how runningOpTransition / stoppedOpTransition /
// terminalOpTransition factor.
//
// The Runner's actionForEdge picks TriggerOp{CancelOp} for the
// source→*Canceling edge (see table_default.go:opForEdge's
// isCancelingState branch), so the handler produces the effect
// bundle without touching Effect ordering.
func cancelToCanceling(
	w domain.World,
	canceling appv1.ApplicationManagerState,
	msg string,
) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("cancel-handoff: %s handler invoked with nil AM", canceling),
		}
	}
	return domain.Result{
		Next: canceling,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   canceling,
				Reason:  cancelHandoffReasonRequested,
				Message: msg,
				// Clear the source state's phase Payload
				// (initializing.phase / upgrading.phase / etc.)
				// so downstream Stopping starts without stale
				// keys leaking cross-state. Symmetric with
				// cancel_pipeline.go's sourceToCanceling.
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  cancelHandoffReasonRequested,
				Message: msg,
			},
		},
	}
}

// handoffToStopping emits the (state=Stopping, reason=CancelHandoffToStopping)
// patch + a Normal EmitEvent for the *Canceling → Stopping
// transition. Shared across the four On*CancelingTick handlers.
//
// The Payload map is explicitly cleared (empty non-nil map — see
// ports.StatusPatch.Payload contract) so a source-state phase does
// not leak into Stopping.
//
// The Stopping handler (see stopping.go) picks up on its own Tick
// and dispatches CallWorkload{Stop} to run the helm scale-to-zero
// / compute release / etc.
func handoffToStopping(w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("cancel-handoff: Tick invoked with nil AM"),
		}
	}
	reason := cancelHandoffReasonToStopping
	if carried, ok := legacyStoppedReason(w.AM.Status.Reason); ok && carried == w.AM.Status.Reason {
		reason = carried
	}
	return domain.Result{
		Next: appv1.Stopping,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopping,
				Reason:  reason,
				Message: cancelHandoffMsgToStopping,
				Payload: map[string]string{}, // clear (defensive; no phase set here today)
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: cancelHandoffMsgToStopping,
			},
		},
	}
}

// -----------------------------------------------------------------
// Initializing → InitializingCanceling → Stopping
// -----------------------------------------------------------------
//
// Legacy correspondence: pkg/appstate/initializing_canceling_app.go —
// Exec calls finishCancelToStopping unconditionally.
//
// Note: the default table declares TWO outgoing edges from
// InitializingCanceling: → Stopping (happy path, ported here) and
// → InstallingCancelFailed ("reuses installing-cancel failure ns"
// per the table comment). The second edge fires ONLY on an API-write
// failure inside the legacy updateStatus, which the port
// folds into the Runner's PatchStatus retry loop. No handler
// dispatches it; it stays as a placeholder for the domain-lint
// exhaustiveness check.
func registerInitializingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.InitializingCanceling, appv1.InitializingCanceling)
	t.SetOrAddHandler(appv1.Initializing, domain.TriggerOp{Op: appv1.CancelOp}, OnInitializingCancel)
	t.SetOrAddHandler(appv1.InitializingCanceling, domain.Tick{}, OnInitializingCancelingTick)
}

// OnInitializingCancel is the CancelOp handler on Initializing.
// Flips to InitializingCanceling with a "cancel requested" event.
func OnInitializingCancel(_ context.Context, w domain.World) domain.Result {
	return cancelToCanceling(w, appv1.InitializingCanceling, cancelHandoffMsgFromInitializing)
}

// OnInitializingCancelingTick is the Tick handler on InitializingCanceling.
// Unconditional handoff to Stopping.
func OnInitializingCancelingTick(_ context.Context, w domain.World) domain.Result {
	return handoffToStopping(w)
}

// -----------------------------------------------------------------
// Resuming → ResumingCanceling → Stopping
// -----------------------------------------------------------------
//
// Legacy correspondence: pkg/appstate/resuming_canceling_app.go.
// Unlike InitializingCanceling, this waits for the in-flight helm
// Scale to return before handoff. Stopping's Scale(0) is another
// helm upgrade; starting it while Scale(-1) is still creating a
// revision fails with "release: already exists".
//
// Second declared edge (→ ResumingCancelFailed) is again an API-write
// failure path only, unreachable via a handler in the port.
// See registerInitializingCanceling doc for the full rationale.
func registerResumingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.ResumingCanceling, appv1.ResumingCanceling)
	t.SetOrAddHandler(appv1.Resuming, domain.TriggerOp{Op: appv1.CancelOp}, OnResumingCancel)
	t.SetOrAddHandler(appv1.ResumingCanceling, domain.Tick{}, OnResumingCancelingTick)
	t.SetOrAddHandler(appv1.ResumingCanceling, domain.WorkloadFailed{}, OnResumingCancelingWorkloadFailed)
	t.SetOrAddHandler(appv1.ResumingCanceling, domain.WorkloadReady{}, OnResumingCancelingWorkloadReady)
}

// OnResumingCancel is the CancelOp handler on Resuming.
// When a helm Scale is still in flight, CancelInFlight asks it to
// return without dropping the claim. Stopping waits for that callback.
func OnResumingCancel(_ context.Context, w domain.World) domain.Result {
	res := cancelToCanceling(w, appv1.ResumingCanceling, cancelHandoffMsgFromResuming)
	if res.Err != nil || !w.WorkloadInFlight {
		return res
	}
	res.Effects = append(res.Effects, domain.CancelInFlight{})
	return res
}

// OnResumingCancelingTick stays in ResumingCanceling while the
// canceled Scale is still outstanding. A tick with no claim means
// the helm call has already returned (or the process restarted
// after it did) and Stopping may start Scale(0).
func OnResumingCancelingTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("cancel-handoff: Tick invoked with nil AM"),
		}
	}
	if w.WorkloadInFlight {
		return domain.Result{Effects: []domain.Effect{domain.CancelInFlight{}}}
	}
	return handoffToStopping(w)
}

// OnResumingCancelingWorkloadFailed is the canceled Scale returning
// an error (context canceled → HelmResumeFailed). The release
// revision is terminal, so Stopping can Scale(0).
func OnResumingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	return handoffToStopping(w)
}

// OnResumingCancelingWorkloadReady covers the race where Scale
// finishes successfully before the cancel context lands. The user
// cancel still proceeds to Stopping.
func OnResumingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	return handoffToStopping(w)
}

// -----------------------------------------------------------------
// Upgrading → UpgradingCanceling → UpgradeFailed
// -----------------------------------------------------------------
//
// Image pull cancel is UpgradeDownloadingCanceling's job. Upgrading
// starts only after that pull has finished, so this path never
// dispatches WorkloadCancelImagePull. It waits for the outstanding
// helm upgrade or startup wait, then lands UpgradeFailed. Config
// and the deployed chart-version annotation are left unchanged.
//
// UpgradingCancelFailed is the timeout edge, not a workload-failure
// edge. A helm callback that returns an error still means the
// release call is done, so the cancel lands UpgradeFailed.
func registerUpgradingCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradingCanceling, appv1.UpgradingCanceling)
	t.SetOrAddHandler(appv1.Upgrading, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradingCancel)
	t.SetOrAddHandler(appv1.UpgradingCanceling, domain.Tick{}, OnUpgradingCancelingTick)
	t.SetOrAddHandler(appv1.UpgradingCanceling, domain.WorkloadReady{}, OnUpgradingCancelingWorkloadReady)
	t.SetOrAddHandler(appv1.UpgradingCanceling, domain.WorkloadFailed{}, OnUpgradingCancelingWorkloadFailed)
}

// OnUpgradingCancel is the CancelOp handler on Upgrading.
// An outstanding helm upgrade or startup wait is asked to return
// via CancelInFlight without dropping the claim. The callback lands
// UpgradeFailed.
func OnUpgradingCancel(_ context.Context, w domain.World) domain.Result {
	res := cancelToCanceling(w, appv1.UpgradingCanceling, cancelHandoffMsgFromUpgrading)
	if res.Err != nil || !w.WorkloadInFlight {
		return res
	}
	res.Effects = append(res.Effects, domain.CancelInFlight{})
	return res
}

// OnUpgradingCancelingTick stays in UpgradingCanceling while helm
// or the startup wait is still outstanding. A tick with no claim
// means that call has already returned, so the cancel lands
// UpgradeFailed.
func OnUpgradingCancelingTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("cancel-handoff: Tick invoked with nil AM")}
	}
	if w.WorkloadInFlight {
		return domain.Result{Effects: []domain.Effect{domain.CancelInFlight{}}}
	}
	return handoffToUpgradeFailed(w)
}

// OnUpgradingCancelingWorkloadReady lands in UpgradeFailed. The
// target config stays on Spec.Config and the deployed chart version
// is not rewritten, so a later UpgradeOp retries the same target.
func OnUpgradingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	return handoffToUpgradeFailed(w)
}

func handoffToUpgradeFailed(w domain.World) domain.Result {
	msg := ""
	if w.AM != nil {
		msg = w.AM.Status.Message
	}
	if msg == "" {
		msg = string(appv1.UpgradeFailed)
	}
	return domain.Result{
		Next: appv1.UpgradeFailed,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeFailed,
				Message: msg,
				Reason:  string(appv1.UpgradeFailed),
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  string(appv1.UpgradeFailed),
				Message: msg,
			},
		},
	}
}

// OnUpgradingCancelingWorkloadFailed is the canceled helm call
// returning an error. The release revision is terminal, so the
// cancel still lands UpgradeFailed.
func OnUpgradingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	return handoffToUpgradeFailed(w)
}

// -----------------------------------------------------------------
// ApplyingEnv → ApplyingEnvCanceling → Stopping
// -----------------------------------------------------------------
//
// Legacy correspondence: pkg/appstate/applying_env_canceling_app.go.
// Same wait as Resuming: ApplyEnv is a helm upgrade, and Stopping's
// Scale(0) is another one. Starting it while ApplyEnv is still
// creating a revision fails with "release: already exists".
//
// Third declared edge (→ ApplyingEnvCancelFailed) is again an
// API-write failure path. A canceled ApplyEnv's WorkloadFailed
// means the helm call returned, so it hands off to Stopping rather
// than that failure terminal.
func registerApplyingEnvCanceling(t *domain.TransitionTable) {
	t.AddReachable(appv1.ApplyingEnvCanceling, appv1.ApplyingEnvCanceling)
	t.SetOrAddHandler(appv1.ApplyingEnv, domain.TriggerOp{Op: appv1.CancelOp}, OnApplyingEnvCancel)
	t.SetOrAddHandler(appv1.ApplyingEnvCanceling, domain.Tick{}, OnApplyingEnvCancelingTick)
	t.SetOrAddHandler(appv1.ApplyingEnvCanceling, domain.WorkloadFailed{}, OnApplyingEnvCancelingWorkloadFailed)
	t.SetOrAddHandler(appv1.ApplyingEnvCanceling, domain.WorkloadReady{}, OnApplyingEnvCancelingWorkloadReady)
}

// OnApplyingEnvCancel is the CancelOp handler on ApplyingEnv.
// When a helm apply-env is still in flight, CancelInFlight asks it to
// return without dropping the claim. Stopping waits for that callback.
func OnApplyingEnvCancel(_ context.Context, w domain.World) domain.Result {
	res := cancelToCanceling(w, appv1.ApplyingEnvCanceling, cancelHandoffMsgFromApplyingEnv)
	if res.Err != nil || !w.WorkloadInFlight {
		return res
	}
	res.Effects = append(res.Effects, domain.CancelInFlight{})
	return res
}

// OnApplyingEnvCancelingTick stays in ApplyingEnvCanceling while the
// canceled apply-env is still outstanding. A tick with no claim means
// the helm call has already returned (or the process restarted
// after it did) and Stopping may start Scale(0).
func OnApplyingEnvCancelingTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("cancel-handoff: Tick invoked with nil AM"),
		}
	}
	if w.WorkloadInFlight {
		return domain.Result{Effects: []domain.Effect{domain.CancelInFlight{}}}
	}
	return handoffToStopping(w)
}

// OnApplyingEnvCancelingWorkloadFailed is the canceled apply-env
// returning an error. The release revision is terminal, so Stopping
// can Scale(0).
func OnApplyingEnvCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	return handoffToStopping(w)
}

// OnApplyingEnvCancelingWorkloadReady covers the race where apply-env
// finishes successfully before the cancel context lands. The user
// cancel still proceeds to Stopping.
func OnApplyingEnvCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	return handoffToStopping(w)
}
