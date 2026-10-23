package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// Package-local mirrors of the Running-state messages. Kept in
// sync with the legacy `AppSelfHeal*` / target-state initial
// messages so operator dashboards + audit logs stay diff-clean
// against a rollback binary.
const (
	runningReasonSelfHeal = "AppMissingSelfHeal"
	runningMsgSelfHeal    = "paired Application CR is missing; self-healing to Uninstalled"

	// Op-driven transitions each use a "starting <op>" template
	// so events fired at the moment of the state flip are still
	// legible even before the target state's own handler writes
	// its richer message.
	runningReasonStopping  = "StopRequested"
	runningMsgStopping     = "user requested stop; transitioning to Stopping"
	runningReasonUninstall = "UninstallRequested"
	runningMsgUninstall    = "user requested uninstall; transitioning to Uninstalling"
	runningReasonUpgrade   = "UpgradeRequested"
	runningMsgUpgrade      = "user requested upgrade; transitioning to UpgradePending"
	runningReasonApplyEnv  = "ApplyEnvRequested"
	runningMsgApplyEnv     = "user requested applyEnv; transitioning to ApplyingEnv"
)

// registerRunning wires the Running-state handlers into the
// TransitionTable. Running is the first "stable" state we port —
// unlike Installing/Downloading/Initializing, it does NOT own a
// long-running CallWorkload. Its job is to observe cluster
// facts (AppExists) and route user-initiated operations.
//
// Wiring semantics:
//
//   - (Running, Tick{}) → OnRunningTick — self-heal to Uninstalled
//     when the paired Application CR has been GC'd out from under
//     us AND the AM is an app-type. Middleware types stay in
//     Running (they don't have paired Application CRs in the
//     same way; matches legacy RunningApp constructor).
//
//   - (Running, TriggerOp{Op:StopOp}) → OnRunningStop
//
//   - (Running, TriggerOp{Op:UninstallOp}) → OnRunningUninstall
//
//   - (Running, TriggerOp{Op:UpgradeOp}) → OnRunningUpgrade
//
//   - (Running, TriggerOp{Op:ApplyEnvOp}) → OnRunningApplyEnv
//
//     Each op handler is a thin PatchAM + EmitEvent that flips
//     the state to the canonical target; the target state's own
//     handler owns the actual work (dispatching helm ops etc.).
//     This keeps Running's surface tiny and the per-op logic
//     colocated with the state that performs it.
//
// AddReachable(Running, Running) declares the self-loop for the
// "AppExists && no op requested" branch of OnRunningTick.
func registerRunning(t *domain.TransitionTable) {
	t.AddReachable(appv1.Running, appv1.Running)
	t.SetOrAddHandler(appv1.Running, domain.Tick{}, OnRunningTick)

	// Op-driven exits. Each edge has a placeholder installed by
	// registerEdges + actionForEdge picking TriggerOp{Op: <canonical>}
	// for the op-target combination. SetOrAddHandler replaces
	// the placeholder with the real (still trivial) handler.
	t.SetOrAddHandler(appv1.Running, domain.TriggerOp{Op: appv1.StopOp}, OnRunningStop)
	t.SetOrAddHandler(appv1.Running, domain.TriggerOp{Op: appv1.UninstallOp}, OnRunningUninstall)
	t.SetOrAddHandler(appv1.Running, domain.TriggerOp{Op: appv1.UpgradeOp}, OnRunningUpgrade)
	t.SetOrAddHandler(appv1.Running, domain.TriggerOp{Op: appv1.ApplyEnvOp}, OnRunningApplyEnv)
}

// OnRunningTick is the reconcile-driven handler for the Running
// state. It has TWO behaviours:
//
//   - Self-heal: legacy pkg/appstate/running_app.go:50-64 detects
//     that the paired Application CR is missing (helm was
//     uninstalled out-of-band, or the AM survived a namespace
//     purge) and flips the AM to Uninstalled. Ported verbatim:
//     when World.AppExists is false AND the AM is an app-type,
//     transition to Uninstalled and let the Uninstalled handler
//     drive the actual cleanup.
//
//   - Steady-state: everything is fine → return empty Result, a
//     no-op reconcile. Controller-runtime's watch-driven Reconcile
//     firing on unrelated changes (annotations, spec updates that
//     the controller doesn't care about) hits this path.
//
// Note: the LEGACY forceDeleteApp goroutine also performs helm
// uninstall + compute release before flipping the status. This
// handler does NOT do that work directly — the Uninstalled state
// handler (a follow-up PR) owns the cleanup effect chain. Until
// then, self-healing to Uninstalled leaves the helm release +
// compute allocation in place; that's a KNOWN skeleton gap.
//
// SKELETON SCOPE — NOT ported:
//
//   - Middleware Downtime aggregation (legacy uses World-level
//     "middleware downtime" tracking to log slow middlewares).
//   - pod_abnormal_suspend_app_controller's Running → Stopping
//     trigger. That controller fires a TriggerOp{StopOp} from a
//     separate controller path, which OnRunningStop already
//     handles — but the OBSERVATION logic (Deployment condition
//     check) is still in the pod-abnormal controller and has
//     not been ported to World.Workload.
func OnRunningTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{}
	}

	// Self-heal only fires for app-type AMs. Middleware AMs use a
	// different Application-tracking mechanism and their absence
	// is not (yet) a self-heal signal.
	if !w.AppExists && w.AM.Spec.Type == appv1.App {
		return selfHealToUninstalled(w)
	}

	// Steady state — no-op reconcile.
	return domain.Result{}
}

// selfHealToUninstalled is the effect bundle for the missing-
// Application branch of OnRunningTick. Factored out so tests can
// build the expected Result directly.
func selfHealToUninstalled(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Uninstalled,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Uninstalled,
				Reason:  runningReasonSelfHeal,
				Message: runningMsgSelfHeal,
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  runningReasonSelfHeal,
				Message: runningMsgSelfHeal,
			},
		},
	}
}

// OnRunningStop is invoked when the apiserver or the pod-abnormal
// controller delivers TriggerOp{Op:StopOp} to a Running AM. Flips
// to Stopping. A pod-abnormal reason stamped on the AM replaces
// StopRequested; a user stop leaves the annotations empty and
// keeps StopRequested.
func OnRunningStop(_ context.Context, w domain.World) domain.Result {
	return stopToStopping(w, runningReasonStopping, runningMsgStopping)
}

// stopToStopping enters Stopping from any StopOp entry this
// process accepts. finishToStopped later publishes Status.Reason,
// so the reason written here is the one Stopped emits.
//
// Pod-abnormal suspend stamps api.AppStopReasonKey and
// AppStopMessageKey before Trigger. Those replace the fallback
// and are deleted in the same transition, so a later user StopOp
// does not reuse an eviction reason. Absent keys keep the fallback
// (StopRequested for a user stop).
func stopToStopping(w domain.World, fallbackReason, fallbackMsg string) domain.Result {
	reason, msg := fallbackReason, fallbackMsg
	var clear []string
	if w.AM != nil && w.AM.Annotations != nil {
		if r := w.AM.Annotations[api.AppStopReasonKey]; r != "" {
			reason = r
			clear = append(clear, api.AppStopReasonKey)
		}
		if m := w.AM.Annotations[api.AppStopMessageKey]; m != "" {
			msg = m
			clear = append(clear, api.AppStopMessageKey)
		}
	}
	res := runningOpTransition(w, appv1.Stopping, reason, msg)
	if res.Err != nil || len(clear) == 0 {
		return res
	}
	res.Effects = append([]domain.Effect{
		domain.PatchSpec{DeleteAnnotations: clear},
	}, res.Effects...)
	return res
}

// OnRunningUninstall is invoked for TriggerOp{Op:UninstallOp}.
func OnRunningUninstall(_ context.Context, w domain.World) domain.Result {
	return runningOpTransition(w, appv1.Uninstalling, runningReasonUninstall, runningMsgUninstall)
}

// OnRunningUpgrade is invoked for TriggerOp{Op:UpgradeOp}.
func OnRunningUpgrade(_ context.Context, w domain.World) domain.Result {
	return runningOpTransition(w, appv1.UpgradePending, runningReasonUpgrade, runningMsgUpgrade)
}

// OnRunningApplyEnv is invoked for TriggerOp{Op:ApplyEnvOp}.
func OnRunningApplyEnv(_ context.Context, w domain.World) domain.Result {
	return runningOpTransition(w, appv1.ApplyingEnv, runningReasonApplyEnv, runningMsgApplyEnv)
}

// runningOpTransition is the shared shape for every op-driven
// Running exit. Kept as a helper (rather than the four handlers
// being duplicated) because:
//
//   - The handlers are structurally identical — a single PatchAM +
//     EmitEvent — and the divergence is data (target state, message).
//     Sharing the shape lets a future audit spot inconsistency
//     between them at a glance.
//
//   - It DOES NOT hide the per-op decision behind a switch: each
//     of the four OnRunning* functions is a named symbol handlers/
//     register.go can point at directly, so the analyzer +
//     handlerpurity linter see each one as an independent target.
//
// The transition intentionally emits NO OpRecord — legacy also
// writes none at the Running-exit point. The full-op OpRecord
// lands at the target state's terminal handler (Uninstalled,
// Stopped, Running-after-upgrade, etc.).
func runningOpTransition(w domain.World, target appv1.ApplicationManagerState, reason, msg string) domain.Result {
	// Guard against a bogus World construction — production is
	// safe because Runner always populates AM, but a hand-crafted
	// test World might not.
	if w.AM == nil {
		return domain.Result{
			Err: fmt.Errorf("running: %s handler invoked with nil AM", reason),
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
