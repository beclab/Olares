package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// Package-local mirrors of the Stopped-state messages. Kept
// symmetric with running.go's op-transition messages so operator
// dashboards see consistent phrasing across stable states.
const (
	stoppedReasonResume    = "ResumeRequested"
	stoppedMsgResume       = "user requested resume; transitioning to Resuming"
	stoppedReasonUninstall = "UninstallRequested"
	stoppedMsgUninstall    = "user requested uninstall; transitioning to Uninstalling"
	stoppedReasonUpgrade   = "UpgradeRequested"
	stoppedMsgUpgrade      = "user requested upgrade; transitioning to UpgradePending"
	stoppedReasonApplyEnv  = "ApplyEnvRequested"
	stoppedMsgApplyEnv     = "user requested applyEnv; transitioning to ApplyingEnv"
)

// registerStopped wires the Stopped-state handlers into the
// TransitionTable. Stopped is a "stable-with-op-routing" state,
// architecturally identical to Running — its four op-driven
// exits are the only edges out.
//
// Wiring semantics:
//
//   - (Stopped, Tick{}) — fresh edge. The default table does not
//     declare it because Stopped has no reconcile-driven work
//     in legacy (compute cleanup is idempotent and can be re-run
//     by any Reconcile that hits Stopped, or deferred to the
//     WorkloadOps.Stop adapter). SetOrAddHandler inserts a
//     no-op Tick that lets a watch-driven Reconcile hit Stopped
//     without erroring.
//
//   - (Stopped, TriggerOp{Op:ResumeOp}) → OnStoppedResume
//
//   - (Stopped, TriggerOp{Op:UninstallOp}) → OnStoppedUninstall
//
//   - (Stopped, TriggerOp{Op:UpgradeOp}) → OnStoppedUpgrade
//
//   - (Stopped, TriggerOp{Op:ApplyEnvOp}) → OnStoppedApplyEnv
//
//     Each op handler is a thin PatchAM + EmitEvent that flips
//     the state to the canonical target; the target state's own
//     handler owns the actual work. Same pattern as Running.
//
// AddReachable(Stopped, Stopped) declares the self-loop for the
// no-op Tick.
//
// SKELETON SCOPE — NOT ported:
//
//   - Guarded compute-allocation cleanup (legacy StoppedApp.
//     cleanupComputeAllocation, stopped_app.go:56-76). The
//     legacy code re-runs EnsureAllocationsDeletedForComputeTarget
//     on every Reconcile as a safety net for AMs that landed in
//     Stopped without going through SuspendingApp. In the
//     port that responsibility belongs to the
//     WorkloadOps.Stop adapter (idempotent on repeat calls) or
//     a dedicated ReleaseCompute effect. Deferred until the
//     ComputeOps port has real semantics.
func registerStopped(t *domain.TransitionTable) {
	t.AddReachable(appv1.Stopped, appv1.Stopped)
	t.SetOrAddHandler(appv1.Stopped, domain.Tick{}, OnStoppedTick)

	// Op-driven exits. Each edge has a placeholder installed by
	// registerEdges + actionForEdge picking TriggerOp{Op: <canonical>}
	// for the op-target combination. SetOrAddHandler replaces
	// the placeholder with the real (still trivial) handler.
	t.SetOrAddHandler(appv1.Stopped, domain.TriggerOp{Op: appv1.ResumeOp}, OnStoppedResume)
	t.SetOrAddHandler(appv1.Stopped, domain.TriggerOp{Op: appv1.UninstallOp}, OnStoppedUninstall)
	t.SetOrAddHandler(appv1.Stopped, domain.TriggerOp{Op: appv1.UpgradeOp}, OnStoppedUpgrade)
	t.SetOrAddHandler(appv1.Stopped, domain.TriggerOp{Op: appv1.ApplyEnvOp}, OnStoppedApplyEnv)
}

// OnStoppedTick is the reconcile-driven handler for the Stopped
// state. It has ONE behaviour today: no-op.
//
// A watch-driven Reconcile firing on unrelated changes
// (annotations, spec updates) must not error out or emit
// spurious effects. See the SKELETON SCOPE note on the register
// function for the compute-cleanup deferral.
func OnStoppedTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{}
	}
	return domain.Result{}
}

// OnStoppedResume is invoked when the apiserver delivers
// TriggerOp{Op:ResumeOp} to a Stopped AM. Flips to Resuming —
// the Resuming handler dispatches the actual helm scale(-1).
func OnStoppedResume(_ context.Context, w domain.World) domain.Result {
	return stoppedOpTransition(w, appv1.Resuming, stoppedReasonResume, stoppedMsgResume)
}

// OnStoppedUninstall is invoked for TriggerOp{Op:UninstallOp}.
func OnStoppedUninstall(_ context.Context, w domain.World) domain.Result {
	return stoppedOpTransition(w, appv1.Uninstalling, stoppedReasonUninstall, stoppedMsgUninstall)
}

// OnStoppedUpgrade is invoked for TriggerOp{Op:UpgradeOp}.
func OnStoppedUpgrade(_ context.Context, w domain.World) domain.Result {
	return stoppedOpTransition(w, appv1.UpgradePending, stoppedReasonUpgrade, stoppedMsgUpgrade)
}

// OnStoppedApplyEnv is invoked for TriggerOp{Op:ApplyEnvOp}.
func OnStoppedApplyEnv(_ context.Context, w domain.World) domain.Result {
	return stoppedOpTransition(w, appv1.ApplyingEnv, stoppedReasonApplyEnv, stoppedMsgApplyEnv)
}

// stoppedOpTransition is the shared shape for every op-driven
// Stopped exit — structurally identical to Running's
// runningOpTransition. Kept as a helper for the same reasons
// (audit-clean divergence, per-op named symbols for the analyzer,
// no OpRecord at exit — the target state's terminal handler
// owns the final OpRecord).
func stoppedOpTransition(w domain.World, target appv1.ApplicationManagerState, reason, msg string) domain.Result {
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("stopped: %s handler invoked with nil AM", reason),
		}
	}

	return domain.Result{
		Next: target,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   target,
				Reason:  reason,
				Message: msg,
				// New operation, fresh progress bar — see
				// terminalOpTransition for the full rationale.
				Progress: "0.00",
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}
