package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the Resuming-state messages.
const (
	resumingMsgScaleStarted     = "resuming.scale: helm scale-up in progress"
	resumingReasonEntered       = "ResumeStarted"
	resumingReasonHandoffLaunch = "ResumeHelmDone"
	resumingMsgHandoffLaunch    = "helm scale-up done; advancing to Initializing for launch wait"
)

// registerResuming wires the Resuming-state handlers into the
// TransitionTable. Structurally identical to Stopping — a
// single-phase helm scale operation — but the SUCCESS branch
// hands off to Initializing (not Running) so the launch-wait
// happens exactly once in the launch-aware state.
//
// Wiring semantics:
//
//   - (Resuming, Tick{}) — fresh edge (Resuming is a workload-
//     callback state in legacy). SetOrAddHandler inserts it.
//
//   - (Resuming, WorkloadReady{}) already has a placeholder
//     targeting Initializing. Replace with the real handler.
//
//   - (Resuming, WorkloadFailed{}) already has a placeholder
//     targeting ResumeFailed. Replace with the real handler.
//
// Cancel path — (Resuming, TriggerOp{CancelOp}) → ResumingCanceling
// — lives in cancel_handoff.go. (Resuming, TriggerOp{StopOp}) →
// Stopping is the pod-abnormal shortcut: OnResumingStop copies a
// stamped reason when present.
func registerResuming(t *domain.TransitionTable) {
	t.AddReachable(appv1.Resuming, appv1.Resuming)
	t.SetOrAddHandler(appv1.Resuming, domain.Tick{}, OnResumingTick)
	t.SetOrAddHandler(appv1.Resuming, domain.WorkloadReady{}, OnResumingWorkloadReady)
	t.SetOrAddHandler(appv1.Resuming, domain.WorkloadFailed{}, OnResumingWorkloadFailed)
	t.SetOrAddHandler(appv1.Resuming, domain.TriggerOp{Op: appv1.StopOp}, OnResumingStop)
}

// OnResumingStop is the pod-abnormal shortcut out of Resuming.
// It enters Stopping immediately; Stopping then preempts the
// in-flight scale, matching the controller's previous direct
// Status write. A stamped abnormal reason replaces StopRequested.
func OnResumingStop(_ context.Context, w domain.World) domain.Result {
	return stopToStopping(w, runningReasonStopping, runningMsgStopping)
}

// OnResumingTick is the Tick handler that starts the resume
// pipeline on fresh entry into Resuming.
//
// Legacy correspondence: pkg/appstate/resuming_app.go:48-81 —
// ResumingApp.exec running scaleOrPatchResume → (optional
// middleware Start). The IsStartUp poll (lines 201-218) that
// legacy also does inside Resuming is INTENTIONALLY dropped —
// the target state Initializing already dispatches
// WorkloadWaitLaunch which covers pod readiness. Duplicating
// the wait was a legacy asymmetry (helm-Install → Initializing
// does not do a pre-wait either); the port unifies
// the pattern.
//
// Phase semantics:
//
//   - phase == ""      → dispatch CallWorkload{Resume}; stamp phase=scale.
//   - phase == scale   → preserve an outstanding claim, or re-dispatch
//     only CallWorkload{Resume} after restart.
//   - unknown phase    → hard fail via Err (safety net).
//
// Pending-pod abort (bytetrade.io/pending-pod) is checked after
// scale succeeds and routes to Stopping. Mid-resume CancelOp is
// handled by cancel_handoff.go.
func OnResumingTick(_ context.Context, w domain.World) domain.Result {
	phase := resumingPhase(w)

	switch phase {
	case "":
		return startHelmResume(w)

	case PhaseResumingScale:
		return recoverWorkload(w, resumeWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("resuming: unknown %s=%q; expected one of {\"\", %q}",
				PayloadResumingPhase, phase, PhaseResumingScale),
		}
	}
}

// OnResumingWorkloadReady completes the single-phase pipeline:
// helm resume succeeded → hand off to Initializing for the
// launch-wait. Emits NO OpRecord here — the terminal "resume
// completed" record lands when Initializing transitions to
// Running (matching legacy: the makeRecord in the resume path
// is written by Initializing/Running handlers, not by
// resuming_app.go's success branch).
func OnResumingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := resumingPhase(w)

	switch phase {
	case PhaseResumingScale:
		if pendingPodAbort(w) {
			return handoffToStoppingFromResume(w)
		}
		return handoffToInitializing(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("resuming: WorkloadReady with empty %s; every CallWorkload emitted from Resuming must stamp the phase first",
				PayloadResumingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("resuming: WorkloadReady with unknown %s=%q",
				PayloadResumingPhase, phase),
		}
	}
}

// OnResumingWorkloadFailed handles a failed helm resume. Routes
// to ResumeFailed with an OpRecord + Warning event.
//
// Legacy correspondence: pkg/appstate/resuming_app.go:157-169 —
// the waitAsync error branch (which emits an OpRecord). The
// synchronous scale-failure branch at lines 49-56 does NOT emit
// an OpRecord in legacy; that asymmetry is corrected here for
// consistency with the other *Failed transitions.
func OnResumingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "resuming failed"
	}
	if reason == "" {
		reason = string(appv1.ResumeFailed)
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.ResumeFailed,
		StateTime: &now,
	}

	// Effect ordering: AppendOpRecord precedes PatchAM so both
	// land atomically. See installing.go for full RV-conflict
	// note.
	return domain.Result{
		Next: appv1.ResumeFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.ResumeFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{}, // clear phase; leaving Resuming
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// startHelmResume is the effect bundle for the fresh-entry
// branch of OnResumingTick.
func startHelmResume(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:    appv1.Resuming,
				Message:  resumingMsgScaleStarted,
				Reason:   resumingReasonEntered,
				Payload:  map[string]string{PayloadResumingPhase: PhaseResumingScale},
				Progress: "0.00",
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  resumingReasonEntered,
				Message: resumingMsgScaleStarted,
			},
			resumeWorkload(),
		},
	}
}

func resumeWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadResume,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmResumeFailed"},
	}
}

// pendingPodAbort reports whether the pod-abnormal-suspend
// controller asked to abort this resume. Legacy resuming_app.go
// checks the annotation during IsStartUp; the wait now lives in
// Initializing, so the closest equivalent is this post-scale check.
func pendingPodAbort(w domain.World) bool {
	if w.AM == nil || w.AM.Annotations == nil {
		return false
	}
	return w.AM.Annotations[AnnotPendingPod] == "true"
}

// handoffToStoppingFromResume aborts a resume whose pending-pod
// annotation was stamped mid-scale. Stopping owns teardown.
func handoffToStoppingFromResume(w domain.World) domain.Result {
	reason := "PendingPodStop"
	if w.AM != nil {
		if carried, ok := legacyStoppedReason(w.AM.Status.Reason); ok && carried == w.AM.Status.Reason {
			reason = carried
		}
	}
	return domain.Result{
		Next: appv1.Stopping,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopping,
				Message: "resume aborted: pending-pod stop requested",
				Reason:  reason,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: "resume aborted: pending-pod stop requested",
			},
		},
	}
}

// handoffToInitializing is the terminal-for-Resuming branch of
// OnResumingWorkloadReady. No OpRecord — the terminal record
// lands at Running (via Initializing).
func handoffToInitializing(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Initializing,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Initializing,
				Message: resumingMsgHandoffLaunch,
				Reason:  resumingReasonHandoffLaunch,
				Payload: map[string]string{}, // clear phase; leaving Resuming
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  resumingReasonHandoffLaunch,
				Message: resumingMsgHandoffLaunch,
			},
		},
	}
}

// resumingPhase reads Payload[PayloadResumingPhase] with the
// standard nil-map guard.
func resumingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadResumingPhase]
}
