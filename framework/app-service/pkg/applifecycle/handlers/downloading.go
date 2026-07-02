package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the legacy Downloading-state messages.
// Same rationale as installing.go — inline copies simplify the
// reviewer's diff against pkg/appstate/downloading_app.go.
const (
	downloadingMsgPullStarted    = "downloading.pull: image pull in progress"
	downloadingMsgHandoffInstall = "start to install"
	downloadingReasonEntered     = "DownloadStarted"
	downloadingReasonComplete    = "DownloadCompleted"
)

// registerDownloading wires the Downloading-state handlers into the
// TransitionTable. Same shape as registerInstalling; see that
// function for the SetOrAddHandler vs Add rationale.
//
// Wiring semantics:
//
//   - (Downloading, Tick{}) already has a placeholder targeting
//     Installing (via actionForEdge). Replace it with OnDownloadingTick
//     so entry fires the image-pull CallWorkload instead of the
//     placeholder's straight advance.
//   - (Downloading, WorkloadReady{}) already has a placeholder
//     targeting Installing; replace with OnDownloadingWorkloadReady
//     which additionally clears the phase Payload.
//   - (Downloading, WorkloadFailed{}) already has a placeholder
//     targeting DownloadFailed; replace with
//     OnDownloadingWorkloadFailed to stamp an OpRecord + Warning
//     event alongside the transition.
//   - AddReachable(Downloading, Downloading) declares the self-loop
//     used by the "waiting for callback" branch of OnDownloadingTick
//     (Result{Next:""} keeps the state; defensive documentation).
//
// Cancel path — (Downloading, TriggerOp{CancelOp}) →
// DownloadingCanceling — remains on the placeholder until the
// cancel-op handler lands.
func registerDownloading(t *domain.TransitionTable) {
	t.AddReachable(appv1.Downloading, appv1.Downloading)
	t.SetOrAddHandler(appv1.Downloading, domain.Tick{}, OnDownloadingTick)
	t.SetOrAddHandler(appv1.Downloading, domain.WorkloadReady{}, OnDownloadingWorkloadReady)
	t.SetOrAddHandler(appv1.Downloading, domain.WorkloadFailed{}, OnDownloadingWorkloadFailed)
}

// OnDownloadingTick is invoked when the Runner delivers a Tick{}
// action to an AM currently in the Downloading state. It is the
// entry point that starts the image-pull pipeline.
//
// Legacy correspondence: pkg/appstate/downloading_app.go:100-147 —
// the `exec` method that unmarshals AppCfg, calls ResolveImageRefs,
// and creates the ImageManager CR via `imageClient.Create`.
//
// Skeleton scope branches:
//
//   - phase == "" (fresh entry from Pending) → dispatch
//     CallWorkload{WorkloadPullImage}; stamp phase = "pull"; emit
//     an EmitEvent so operators see the state fan out.
//   - phase == "pull" → preserve an active workload claim, or
//     re-dispatch only WorkloadPullImage after restart recovery.
//   - unknown phase → hard fail via Err (same safety net as
//     OnInstallingTick).
//
// The WorkloadReady callback arrives once every node's bytes have
// landed (Percent>=100). Runner.dispatchImagePull polls for
// that; unpack into containerd continues after the handoff and is
// absorbed by Installing. The phase == "pull" branch covers the
// whole download, which for a fresh pull runs minutes to tens of
// minutes.
//
// Legacy differences NOT ported (each is a separate follow-up):
//
//   - Intermediate AM.Status.Progress: the poll loop publishes each
//     percent change on NATS (legacy updateProgress) but does not
//     PatchAM. Live percentages also stay readable from the
//     ImageManager CR.
//   - Cancel-mid-pull (Downloading → DownloadingCanceling): needs
//     the cancel-op handler in this state's TriggerOp{CancelOp}
//     edge, plus a WorkloadCancelImagePull CallWorkload routing
//     to ImagePuller.CancelJob. Placeholder remains until then.
//   - Repair path (DownloadingCancelFailed → repair): needs
//     UpdateJobStatus wiring; tracked in ADAPTERS_TODO.md.
func OnDownloadingTick(_ context.Context, w domain.World) domain.Result {
	phase := downloadingPhase(w)

	switch phase {
	case "":
		return startImagePull(w)

	case PhaseDownloadingPull:
		return recoverWorkload(w, pullImageWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("downloading: unknown %s=%q in Payload; expected one of {\"\", %q}",
				PayloadDownloadingPhase, phase, PhaseDownloadingPull),
		}
	}
}

// OnDownloadingWorkloadReady is invoked when a
// CallWorkload{WorkloadPullImage} dispatched from Downloading
// completes successfully. Advances the AM to Installing.
//
// Legacy correspondence: pkg/appstate/downloading_app.go:44-53 —
// the WaitAsync success branch's `updateStatus(Installing, "start
// to install", ...)`.
func OnDownloadingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := downloadingPhase(w)

	switch phase {
	case PhaseDownloadingPull:
		return handoffToInstalling(w)

	case "":
		// WorkloadReady arriving with empty phase means the
		// Runner delivered a callback we didn't dispatch — a
		// bug. Fail loudly so it gets fixed instead of silently
		// mis-routing to Installing.
		return domain.Result{
			Err: fmt.Errorf("downloading: WorkloadReady with empty %s; every CallWorkload emitted from Downloading must stamp the phase first",
				PayloadDownloadingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("downloading: WorkloadReady with unknown %s=%q",
				PayloadDownloadingPhase, phase),
		}
	}
}

// OnDownloadingWorkloadFailed is invoked when the image pull's
// CallWorkload fails (either ResolveRefs errored, CreateJob
// errored, or — once polling is added — Progress reports
// Failed=true). Transitions unconditionally to DownloadFailed.
//
// Legacy correspondence: pkg/appstate/downloading_app.go:33-42
// (poll-loop WaitAsync failure) plus :102-111 (Exec-time failure).
// Both legacy paths write DownloadFailed with the same OpRecord
// shape, so the port collapses them into one
// handler branch.
//
// Skeleton scope: does NOT branch on specific error codes. A
// follow-up will teach the handler to route "canceled" (from a
// concurrent cancel-op) to DownloadingCanceled instead of
// DownloadFailed — needs typed sentinel errors on WorkloadFailed
// which the Runner does not currently thread through.
func OnDownloadingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "image pull failed"
	}
	if reason == "" {
		reason = string(appv1.DownloadFailed)
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.DownloadFailed,
		StateTime: &now,
	}

	// Effect ordering rule (see installing.go for the full RV-
	// conflict note): AppendOpRecord MUST precede PatchAM so the
	// buffered OpRecord merges into the same PatchStatus RPC.
	return domain.Result{
		Next: appv1.DownloadFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.DownloadFailed,
				Message: msg,
				Reason:  reason,
				// Clear phase — leaving Downloading.
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

// startImagePull is the effect bundle for the "fresh entry"
// branch of OnDownloadingTick. Factored so tests can hand-
// construct the same Result to compare against.
//
// The CallWorkload's OnFailure carries the fallback reason
// "ImagePullFailed", which the Runner overwrites with
// LandedReason (e.g. "ImageResolveFailed",
// "ImagePullStartFailed") when the workload provides one. See
// runner/effects.go's failure-backfill priority comment.
func startImagePull(_ domain.World) domain.Result {
	return domain.Result{
		// Next: "" keeps the state as Downloading while the
		// phase Payload advances from "" to "pull". Runner
		// still applies the PatchAM because we include one
		// explicitly.
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:    appv1.Downloading,
				Message:  downloadingMsgPullStarted,
				Reason:   downloadingReasonEntered,
				Payload:  map[string]string{PayloadDownloadingPhase: PhaseDownloadingPull},
				Progress: "0.00",
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  downloadingReasonEntered,
				Message: downloadingMsgPullStarted,
			},
			pullImageWorkload(),
		},
	}
}

func pullImageWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadPullImage,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "ImagePullFailed"},
	}
}

// handoffToInstalling is the terminal branch of
// OnDownloadingWorkloadReady. The Downloading state's job is done —
// hand off to Installing with a status update and event so
// operators see the state transition.
//
// Legacy behaviour: no OpRecord is written at this transition
// (only at the terminal Success/Failed states). Match that here;
// the OpRecord for the whole install lifecycle is written when
// Installing completes (see finishInstall in installing.go).
func handoffToInstalling(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Installing,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: downloadingMsgHandoffInstall,
				Reason:  downloadingReasonComplete,
				// Clear phase — leaving Downloading; the
				// Installing state manages its own phase key.
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  downloadingReasonComplete,
				Message: downloadingMsgHandoffInstall,
			},
		},
	}
}

// downloadingPhase reads Payload[PayloadDownloadingPhase] with
// nil-map guard. Same shape as installingPhase.
func downloadingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadDownloadingPhase]
}
