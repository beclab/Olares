package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the Initializing-state messages. Same
// rationale as installing.go — inline copies simplify the reviewer's
// diff against pkg/appstate/initializing_app.go.
const (
	initializingMsgWaitStarted = "initializing.wait: waiting for workload to become ready"
	initializingReasonEntered  = "InitializingStarted"
	initializingReasonReady    = "Running"
)

// registerInitializing wires the Initializing-state handlers into
// the TransitionTable. Same SetOrAddHandler pattern as
// registerInstalling / registerDownloading.
//
// Wiring semantics:
//
//   - (Initializing, Tick{}) already has a placeholder targeting
//     Running (via actionForEdge's default). Replace with
//     OnInitializingTick so entry fires the wait-launch
//     CallWorkload instead of jumping straight to Running.
//   - (Initializing, WorkloadReady{}) already has a placeholder
//     targeting Running; replace with OnInitializingWorkloadReady
//     which additionally writes the terminal OpRecord and emits
//     the "install completed" event.
//   - (Initializing, WorkloadFailed{}) does NOT have a placeholder
//     targeting a failure state — the legacy code funnels wait-
//     launch failures back into the cancel-path (see the doc-
//     comment on OnInitializingWorkloadFailed). We ADD the edge
//     to InitializingCanceling explicitly and wire the handler.
//   - AddReachable(Initializing, Initializing) declares the
//     "still waiting" self-loop (Result{Next:""} while phase=wait).
//
// Cancel path — (Initializing, TriggerOp{CancelOp}) →
// InitializingCanceling — stays on the placeholder until the
// cancel-op handler lands.
func registerInitializing(t *domain.TransitionTable) {
	t.AddReachable(appv1.Initializing, appv1.Initializing)
	t.AddReachable(appv1.Initializing, appv1.InitializingCanceling)
	t.SetOrAddHandler(appv1.Initializing, domain.Tick{}, OnInitializingTick)
	t.SetOrAddHandler(appv1.Initializing, domain.WorkloadReady{}, OnInitializingWorkloadReady)
	t.SetOrAddHandler(appv1.Initializing, domain.WorkloadFailed{}, OnInitializingWorkloadFailed)
}

// OnInitializingTick is invoked when the Runner delivers a Tick{}
// action to an AM currently in the Initializing state. Entry point
// that dispatches WaitForLaunch.
//
// Legacy correspondence: pkg/appstate/initializing_app.go:79-127 —
// the goroutine wrapping `ops.WaitForLaunch()`. The
// port collapses the two "outcome-writes-status" branches into two
// pure handlers (OnWorkloadReady, OnWorkloadFailed) plus the
// Runner-owned goroutine spawned by CallWorkload{WaitLaunch}.
//
// Phase semantics identical to Downloading / Installing:
//
//   - phase == ""    → fresh entry from Installing; dispatch
//     CallWorkload{WaitLaunch}; stamp phase=wait.
//   - phase == wait  → preserve an active claim, or re-dispatch
//     only WaitLaunch after restart recovery.
//   - unknown phase  → hard fail via Err (safety net).
//
// Legacy differences NOT ported (each a separate follow-up):
//
//   - Caller-driven cancel disambiguation (legacy lines 82-95):
//     WaitForLaunch's `!ok && (err==nil || ctx.Err()!=nil)` branch
//     "leaves terminal state to the initiator". This handler does
//     NOT need that logic because the Runner's cancel-op path
//     will have already flipped the state OUT of Initializing
//     before the WaitLaunch callback lands — so
//     OnInitializingWorkloadReady/Failed simply won't fire when
//     an initiator-driven cancel has beaten them to it. Requires
//     the cancel-op handler though; documented dependency.
//
//   - OpType-specific success message: OnInitializingWorkloadReady
//     branches on Spec.OpType (Install / Upgrade / ApplyEnv) to
//     pick the "Successfully %sed %s: %s" template. This IS
//     ported here since Spec.OpType is readable from World.AM.
func OnInitializingTick(_ context.Context, w domain.World) domain.Result {
	phase := initializingPhase(w)

	switch phase {
	case "":
		return startWaitForLaunch(w)

	case PhaseInitializingWait:
		return recoverWorkload(w, waitForLaunchWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("initializing: unknown %s=%q in Payload; expected one of {\"\", %q}",
				PayloadInitializingPhase, phase, PhaseInitializingWait),
		}
	}
}

// OnInitializingWorkloadReady is invoked when a
// CallWorkload{WorkloadWaitLaunch} dispatched from Initializing
// completes successfully. Advances to Running with an OpRecord
// carrying the op-type-specific completion message and emits the
// success event.
//
// Legacy correspondence: pkg/appstate/initializing_app.go:111-126
// — the WaitForLaunch success branch.
func OnInitializingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := initializingPhase(w)

	switch phase {
	case PhaseInitializingWait:
		return finishToRunning(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("initializing: WorkloadReady with empty %s; every CallWorkload emitted from Initializing must stamp the phase first",
				PayloadInitializingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("initializing: WorkloadReady with unknown %s=%q",
				PayloadInitializingPhase, phase),
		}
	}
}

// OnInitializingWorkloadFailed is invoked when WaitForLaunch times
// out or the workload otherwise fails to become ready.
//
// LEGACY BEHAVIOUR NOTE — target is InitializingCanceling, NOT
// InitializingFailed:
//
// The legacy code (pkg/appstate/initializing_app.go:99-108) drives
// a wait-launch failure to InitializingCanceling with:
//
//   - Reason  = constants.AppStopDueToInitFailed
//   - Message = constants.InitializeCancelDueToWaitForLaunchTimedOut
//
// The state machine then unwinds via the InitializingCanceling
// handler (which owns the actual helm rollback + IM cleanup).
// This is subtle: "the workload failed" and "we cancel init"
// share a downstream path, so we route both here.
//
// Distinguishing "real failure" from "caller-driven cancel" is a
// Runner-level concern (see legacy lines 82-95). Once the
// cancel-op handler lands and races with WaitLaunch, the caller-
// driven cancel case is handled BEFORE the WorkloadFailed callback
// can land (state has already left Initializing, so
// Engine.Step's ErrNoTransition path swallows the stale
// callback). Until then, treat every WaitLaunch failure as a
// real init timeout.
func OnInitializingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = constants.InitializeCancelDueToWaitForLaunchTimedOut
	}
	if reason == "" {
		reason = constants.AppStopDueToInitFailed
	}

	// No OpRecord written on the wait-launch failure path — the
	// terminal OpRecord for the whole install/upgrade lifecycle
	// lands when the InitializingCanceling handler completes its
	// unwind. Legacy does the same (initializing_app.go:99-108
	// passes `nil` for opRecord in updateStatus).
	return domain.Result{
		Next: appv1.InitializingCanceling,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InitializingCanceling,
				Message: msg,
				Reason:  reason,
				// Clear phase — leaving Initializing.
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// startWaitForLaunch is the effect bundle for the fresh-entry
// branch of OnInitializingTick.
//
// OnFailure carries the fallback reason "InitializeCancelDueToWait
// ForLaunchTimedOut" so a workload that fails with no LandedReason
// still produces a legible AM Status. The Runner overrides this
// with LandedReason if the port provides one.
func startWaitForLaunch(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Initializing,
				Message: initializingMsgWaitStarted,
				Reason:  initializingReasonEntered,
				Payload: map[string]string{PayloadInitializingPhase: PhaseInitializingWait},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  initializingReasonEntered,
				Message: initializingMsgWaitStarted,
			},
			waitForLaunchWorkload(),
		},
	}
}

func waitForLaunchWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadWaitLaunch,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{
			Reason:  constants.AppStopDueToInitFailed,
			Message: constants.InitializeCancelDueToWaitForLaunchTimedOut,
		},
	}
}

// finishToRunning is the terminal branch of
// OnInitializingWorkloadReady. Writes the op-type-specific success
// message + the terminal OpRecord and advances to Running.
func finishToRunning(w domain.World) domain.Result {
	successMsg := successMessageFor(w.AM)
	now := metav1.NewTime(w.Now)

	// See installing.go for the "OpRecord.Version is left empty
	// at the handler layer" note.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   successMsg,
		Source:    w.AM.Spec.Source,
		Status:    appv1.Running,
		StateTime: &now,
	}

	// AppendOpRecord precedes PatchAM so both land in one
	// PatchStatus RPC — same RV-conflict prevention as
	// installing.finishInstall.
	return domain.Result{
		Next: appv1.Running,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:    appv1.Running,
				Message:  string(appv1.Running),
				Reason:   string(appv1.Running),
				Progress: "100.00",
				Payload:  map[string]string{}, // clear phase; leaving Initializing
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  initializingReasonReady,
				Message: successMsg,
			},
		},
	}
}

// successMessageFor picks the op-type-specific completion message,
// matching legacy behaviour in initializing_app.go:113-119.
//
// The three op types Initializing can transition to Running for
// are Install, Upgrade, ApplyEnv. Anything else falls back to the
// Install template (matches legacy default; the other op types
// that reach Initializing all use a copy of the Install path in
// practice).
func successMessageFor(am *appv1.ApplicationManager) string {
	tpl := constants.InstallOperationCompletedTpl
	switch am.Spec.OpType {
	case appv1.UpgradeOp:
		tpl = constants.UpgradeOperationCompletedTpl
	case appv1.ApplyEnvOp:
		tpl = constants.ApplyEnvOperationCompletedTpl
	}
	return fmt.Sprintf(tpl, am.Spec.Type, am.Spec.AppName)
}

// initializingPhase reads Payload[PayloadInitializingPhase] with
// the same nil-map guard pattern as downloadingPhase /
// installingPhase.
func initializingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadInitializingPhase]
}
