package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the Stopping-state messages. Same
// rationale as uninstalling.go — inline copies simplify diffing
// against pkg/appstate/suspending_app.go.
const (
	stoppingMsgScaleStarted = "stopping.scale: helm scale-to-zero + workload teardown in progress"
	stoppingReasonEntered   = "StopStarted"
	stoppingReasonSucceeded = "StopCompleted"
)

// legacyStoppedReason is the Reason carried onto Stopped. The
// pre-refactor machine published status.reason on that transition,
// and these are the values it wrote: a user stop, an eviction, an
// init failure, an unschedulable pod, or a resource-pressure
// rejection. StopRequested is the runner's name for a user stop;
// the published value stays StopByUser.
//
// K8sRequestCPUPressure is published with the spelling subscribers
// already match. The validator constant still stores the historical
// typo K8sReqeustCPUPressure; both spellings map to the same tag.
func legacyStoppedReason(current string) (string, bool) {
	switch current {
	case constants.AppStopByUser,
		constants.AppStopDueToEvicted,
		constants.AppStopDueToInitFailed,
		constants.AppHamiSchedulable,
		constants.AppUnschedulable,
		string(constants.DiskPressure),
		string(constants.SystemCPUPressure),
		string(constants.SystemMemoryPressure),
		string(constants.K8sRequestMemoryPressure),
		constants.AppStopDueToStartUpFailed:
		return current, true
	case string(constants.K8sRequestCPUPressure), "K8sRequestCPUPressure":
		return "K8sRequestCPUPressure", true
	case runningReasonStopping:
		return constants.AppStopByUser, true
	default:
		return "", false
	}
}

// registerStopping wires the Stopping-state handlers into the
// TransitionTable. Same shape as registerUninstalling but with
// a single phase because the legacy stop path has no long
// namespace-gc poll — helm scale(0) returns synchronously
// (from the perspective of "helm's own state is consistent").
//
// Wiring semantics:
//
//   - (Stopping, Tick{}) — fresh edge. Not declared by the
//     default table (Stopping is a workload-callback state).
//     SetOrAddHandler inserts it.
//
//   - (Stopping, WorkloadReady{}) already has a placeholder
//     targeting Stopped. Replace with the real handler that
//     writes the terminal OpRecord.
//
//   - (Stopping, WorkloadFailed{}) already has a placeholder
//     targeting StopFailed. Replace with the real handler.
//
// Cancel path (mid-stop cancellation) is NOT ported here — the
// legacy Cancel method is a FIXME even in the pre-refactor code
// (suspending_app.go:165). Follow-up alongside the broader
// *Canceling handler batch.
func registerStopping(t *domain.TransitionTable) {
	t.AddReachable(appv1.Stopping, appv1.Stopping)
	t.SetOrAddHandler(appv1.Stopping, domain.Tick{}, OnStoppingTick)
	t.SetOrAddHandler(appv1.Stopping, domain.WorkloadReady{}, OnStoppingWorkloadReady)
	t.SetOrAddHandler(appv1.Stopping, domain.WorkloadFailed{}, OnStoppingWorkloadFailed)
}

// OnStoppingTick is the Tick handler that starts the stop pipeline
// on fresh entry into Stopping.
//
// Legacy correspondence: pkg/appstate/suspending_app.go:83-163 —
// SuspendingApp.exec running scaleOrPatchSuspend → (optional v2
// stop-all cascade) → (optional middleware Stop) →
// compute.DeleteAllocationsForComputeTarget.
//
// The port bundles ALL of the above into the
// WorkloadOps.Stop adapter — the handler just requests the effect.
//
// Phase semantics:
//
//   - phase == ""      → dispatch CallWorkload{Stop}; stamp phase=scale.
//   - phase == scale   → preserve an outstanding claim, or re-dispatch
//     only CallWorkload{Stop} after restart.
//   - unknown phase    → hard fail via Err (safety net).
//
// Stop-all v2 cascade, webhook retry, wait-for-pods, middleware
// Stop, and compute.DeleteForTarget run inside the Runner's
// dispatchStop (same order as suspending_app.go:exec).
//
//   - Mid-stop Cancel. Legacy Cancel is a FIXME; nothing to
//     port.
//
//   - Pod-abnormal-suspend controller reason preservation
//     (SuspendingApp.Exec lines 66-73 re-reads Status.Reason
//     from apiserver in case a peer controller wrote a richer
//     reason mid-stop). Best modeled as a World-materialization
//     enrichment in a follow-up.
func OnStoppingTick(_ context.Context, w domain.World) domain.Result {
	phase := stoppingPhase(w)

	switch phase {
	case "":
		return startHelmStop(w)

	case PhaseStoppingScale:
		return recoverWorkload(w, stopWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("stopping: unknown %s=%q; expected one of {\"\", %q}",
				PayloadStoppingPhase, phase, PhaseStoppingScale),
		}
	}
}

// OnStoppingWorkloadReady completes the single-phase pipeline:
// helm stop succeeded → write the terminal OpRecord and advance
// to Stopped.
func OnStoppingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := stoppingPhase(w)

	switch phase {
	case PhaseStoppingScale:
		return finishToStopped(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("stopping: WorkloadReady with empty %s; every CallWorkload emitted from Stopping must stamp the phase first",
				PayloadStoppingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("stopping: WorkloadReady with unknown %s=%q",
				PayloadStoppingPhase, phase),
		}
	}
}

// OnStoppingWorkloadFailed handles a failed helm stop. Routes to
// StopFailed with an OpRecord + Warning event carrying the
// workload's failure data.
//
// Legacy correspondence: pkg/appstate/suspending_app.go:52-62 —
// the sync-exec error branch that emits makeRecord(StopFailed).
func OnStoppingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "stopping failed"
	}
	if reason == "" {
		reason = string(appv1.StopFailed)
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.StopFailed,
		StateTime: &now,
	}

	// Effect ordering: AppendOpRecord precedes PatchAM so both
	// land atomically. See installing.go for full RV-conflict
	// note.
	return domain.Result{
		Next: appv1.StopFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.StopFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{}, // clear phase; leaving Stopping
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// startHelmStop is the effect bundle for the fresh-entry branch
// of OnStoppingTick.
func startHelmStop(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopping,
				Message: stoppingMsgScaleStarted,
				// Leave Reason alone. The value that entered Stopping
				// (StopByUser, Evicted, Unschedulable, ...) is what
				// finishToStopped publishes, matching the pre-refactor
				// suspend path that copied status.reason onto Stopped.
				Payload: map[string]string{PayloadStoppingPhase: PhaseStoppingScale},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  stoppingReasonEntered,
				Message: stoppingMsgScaleStarted,
			},
			stopWorkload(),
		},
	}
}

func stopWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadStop,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmStopFailed"},
		// Stop is also the handoff from Initializing cancel, which
		// arrives while WaitLaunch still holds the claim. Resume
		// and ApplyingEnv cancel wait for their helm upgrade to
		// return first, so those claims are already clear.
		Preempts: true,
	}
}

// finishToStopped is the terminal branch of
// OnStoppingWorkloadReady. Emits the terminal OpRecord and
// advances to Stopped.
func finishToStopped(w domain.World) domain.Result {
	successMsg := fmt.Sprintf(constants.StopOperationCompletedTpl, w.AM.Spec.AppName)
	now := metav1.NewTime(w.Now)

	// See installing.go for the "OpRecord.Version left empty at
	// handler layer" note.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   successMsg,
		Source:    w.AM.Spec.Source,
		Status:    appv1.Stopped,
		StateTime: &now,
	}

	statusReason := string(appv1.Stopped)
	eventReason := stoppingReasonSucceeded
	if carried, ok := legacyStoppedReason(w.AM.Status.Reason); ok {
		statusReason = carried
		eventReason = carried
	}

	return domain.Result{
		Next: appv1.Stopped,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopped,
				Message: successMsg,
				Reason:  statusReason,
				Payload: map[string]string{}, // clear phase; leaving Stopping
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  eventReason,
				Message: successMsg,
			},
		},
	}
}

// stoppingPhase reads Payload[PayloadStoppingPhase] with the
// standard nil-map guard.
func stoppingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadStoppingPhase]
}
