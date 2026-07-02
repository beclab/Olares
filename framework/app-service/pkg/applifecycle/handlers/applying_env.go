package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the ApplyingEnv-state messages.
const (
	applyingEnvMsgHelmStarted        = "applyingenv.helm: helm apply-env in progress"
	applyingEnvReasonEntered         = "ApplyEnvStarted"
	applyingEnvReasonHandoffLaunch   = "ApplyEnvHelmDone"
	applyingEnvMsgHandoffLaunch      = "Environment variables applied, waiting for application to initialize"
	applyingEnvMsgHandoffStoppedFast = "apply-env at replicas=0 done; landing back in Stopped"
)

// registerApplyingEnv wires the ApplyingEnv-state handlers into
// the TransitionTable. Structurally the simplest of the *ing
// states — a single helm-apply-env phase that ends either in
// success (→ Initializing, always) or failure (→ ApplyEnvFailed).
//
// Wiring semantics:
//
//   - (ApplyingEnv, Tick{}) — fresh edge (workload-callback
//     state in legacy). SetOrAddHandler inserts it.
//
//   - (ApplyingEnv, WorkloadReady{}) — placeholder already
//     targets Initializing. SetOrAddHandler replaces it.
//
//   - (ApplyingEnv, WorkloadFailed{}) — placeholder already
//     targets ApplyEnvFailed. SetOrAddHandler replaces it.
//
// Cancel path — (ApplyingEnv, TriggerOp{CancelOp}) →
// ApplyingEnvCanceling — lives in cancel_handoff.go.
// (ApplyingEnv, TriggerOp{StopOp}) → Stopping is the pod-abnormal
// shortcut: OnApplyingEnvStop copies a stamped reason when present.
//
// The (ApplyingEnv → Stopped) edge is the applyEnv-from-Stopped
// fast-path: legacy applying_env_app.go reads
// AppPreUpgradeStateKey == stopped, sets SkipWaitForStartUp, and
// lands Stopped. OnApplyingEnvWorkloadReady branches the same way.
func registerApplyingEnv(t *domain.TransitionTable) {
	t.AddReachable(appv1.ApplyingEnv, appv1.ApplyingEnv)
	t.SetOrAddHandler(appv1.ApplyingEnv, domain.Tick{}, OnApplyingEnvTick)
	t.SetOrAddHandler(appv1.ApplyingEnv, domain.WorkloadReady{}, OnApplyingEnvWorkloadReady)
	t.SetOrAddHandler(appv1.ApplyingEnv, domain.WorkloadFailed{}, OnApplyingEnvWorkloadFailed)
	t.SetOrAddHandler(appv1.ApplyingEnv, domain.TriggerOp{Op: appv1.StopOp}, OnApplyingEnvStop)
}

// OnApplyingEnvStop is the pod-abnormal shortcut out of ApplyingEnv.
// It enters Stopping immediately; Stopping then preempts the
// in-flight apply, matching the controller's previous direct
// Status write. A stamped abnormal reason replaces StopRequested.
func OnApplyingEnvStop(_ context.Context, w domain.World) domain.Result {
	return stopToStopping(w, runningReasonStopping, runningMsgStopping)
}

// OnApplyingEnvTick is the Tick handler that starts the apply-env
// pipeline on fresh entry into ApplyingEnv.
//
// Legacy correspondence: pkg/appstate/applying_env_app.go:88-118 —
// ApplyingEnvApp.exec running:
//  1. KubeConfig() lookup
//  2. Unmarshal AppCfg
//  3. NewHelmOps(...) + helmOps.ApplyEnv()
//
// All three are bundled into the WorkloadOps.ApplyEnv adapter.
//
// Phase semantics:
//
//   - phase == ""     → dispatch CallWorkload{ApplyEnv}; stamp phase=helm.
//   - phase == helm   → preserve an outstanding claim, or re-dispatch
//     only CallWorkload{ApplyEnv} after restart.
//   - unknown phase   → hard fail via Err (safety net).
//
// Mid-apply-env Cancel lives in cancel_handoff.go. StopOp is
// OnApplyingEnvStop.
func OnApplyingEnvTick(_ context.Context, w domain.World) domain.Result {
	phase := applyingEnvPhase(w)

	switch phase {
	case "":
		return startHelmApplyEnv(w)

	case PhaseApplyingEnvHelm:
		return recoverWorkload(w, applyEnvWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("applyingenv: unknown %s=%q; expected one of {\"\", %q}",
				PayloadApplyingEnvPhase, phase, PhaseApplyingEnvHelm),
		}
	}
}

// OnApplyingEnvWorkloadReady completes the single-phase pipeline:
// helm apply-env succeeded → Initializing (normal) or Stopped
// (apply-env-from-Stopped fast-path). Emits NO OpRecord — matches
// legacy: the terminal record lands when Initializing→Running
// fires, and the Stopped fast-path writes none.
func OnApplyingEnvWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := applyingEnvPhase(w)

	switch phase {
	case PhaseApplyingEnvHelm:
		if upgradeFromStopped(w) {
			return handoffToStoppedAfterApplyEnv(w)
		}
		return handoffToInitializingAfterApplyEnv(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("applyingenv: WorkloadReady with empty %s; every CallWorkload emitted from ApplyingEnv must stamp the phase first",
				PayloadApplyingEnvPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("applyingenv: WorkloadReady with unknown %s=%q",
				PayloadApplyingEnvPhase, phase),
		}
	}
}

// OnApplyingEnvWorkloadFailed handles a failed helm apply-env.
// Routes to ApplyEnvFailed with an OpRecord + Warning event.
//
// Legacy correspondence: applying_env_app.go:60-72 — the sync-
// exec error branch that emits makeRecord(ApplyEnvFailed).
func OnApplyingEnvWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "applyingenv failed"
	}
	if reason == "" {
		reason = string(appv1.ApplyEnvFailed)
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.ApplyEnvFailed,
		StateTime: &now,
	}

	// Effect ordering: AppendOpRecord precedes PatchAM so both
	// land atomically. See installing.go for full RV-conflict
	// note.
	return domain.Result{
		Next: appv1.ApplyEnvFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.ApplyEnvFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{}, // clear phase; leaving ApplyingEnv
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// startHelmApplyEnv is the effect bundle for the fresh-entry
// branch of OnApplyingEnvTick.
func startHelmApplyEnv(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:    appv1.ApplyingEnv,
				Message:  applyingEnvMsgHelmStarted,
				Reason:   applyingEnvReasonEntered,
				Payload:  map[string]string{PayloadApplyingEnvPhase: PhaseApplyingEnvHelm},
				Progress: "0.00",
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  applyingEnvReasonEntered,
				Message: applyingEnvMsgHelmStarted,
			},
			applyEnvWorkload(),
		},
	}
}

func applyEnvWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadApplyEnv,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmApplyEnvFailed"},
	}
}

// handoffToStoppedAfterApplyEnv is the apply-env-from-Stopped
// success branch. Reason AppStopByUser matches
// applying_env_app.go:landedReason.
func handoffToStoppedAfterApplyEnv(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Stopped,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopped,
				Message: applyingEnvMsgHandoffStoppedFast,
				Reason:  constants.AppStopByUser,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  constants.AppStopByUser,
				Message: applyingEnvMsgHandoffStoppedFast,
			},
		},
	}
}

// handoffToInitializingAfterApplyEnv is the success branch.
// Message text mirrors legacy applying_env_app.go:77 verbatim so
// operator dashboards diff-clean between old and new binaries.
func handoffToInitializingAfterApplyEnv(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Initializing,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Initializing,
				Message: applyingEnvMsgHandoffLaunch,
				Reason:  applyingEnvReasonHandoffLaunch,
				Payload: map[string]string{}, // clear phase; leaving ApplyingEnv
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  applyingEnvReasonHandoffLaunch,
				Message: applyingEnvMsgHandoffLaunch,
			},
		},
	}
}

// applyingEnvPhase reads Payload[PayloadApplyingEnvPhase] with
// the standard nil-map guard.
func applyingEnvPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadApplyingEnvPhase]
}
