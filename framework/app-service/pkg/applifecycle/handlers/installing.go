package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the legacy install messages. Held here
// rather than pulled from a helper so a reviewer diffing this file
// against pkg/appstate/installing_app.go can eyeball the string
// contents without spelunking. If either wording drifts, the legacy
// consumers (bfl UI / operator playbooks) MUST be checked first.
const (
	installingMsgExposePortsStarted  = "installing.expose-ports: assigning expose ports"
	installingMsgValidateStarted     = "installing.validate: pre-helm resource/quota check in progress"
	installingMsgHelmStarted         = "installing.helm-install: helm release create in progress"
	installingMsgPostValidateStarted = "installing.post-validate: post-helm resource/quota check in progress"
	installingMsgScaleUpStarted      = "installing.scale-up: restoring manifest-declared replica counts"
	installingMsgWaitStartUpStarted  = "installing.wait-startup: waiting for pods to become ready"
	installingMsgWaitLaunchStarted   = "installing.wait-launch: waiting for middleware to launch"
	installingMsgCleanupStarted      = "installing.cleanup: tearing down partial install after failure"
	installingReasonEntered          = "InstallStarted"
	installingReasonExposePorts      = "InstallExposePortsStarted"
	installingReasonHelmStarted      = "HelmInstallStarted"
	installingReasonPostValidate     = "InstallPostValidateStarted"
	installingReasonScaleUpStarted   = "InstallScaleUpStarted"
	installingReasonWaitStartUp      = "InstallWaitStartUpStarted"
	installingReasonWaitLaunch       = "InstallWaitLaunchStarted"
	installingReasonCleanupStarted   = "InstallCleanupStarted"
	installingReasonSucceeded        = "InstallSucceeded"
)

// registerInstalling wires the Installing-state handlers into the
// TransitionTable. Exported via Register (register.go); kept private
// here so a reviewer opening this file sees the handler + its wiring
// as a single unit.
//
// Wiring semantics:
//
//   - (Installing, Tick{}) is a fresh edge — the default table does
//     not declare it because no legacy Installing → Installing self-
//     loop exists. The port needs it to drive the
//     multi-phase install without an intermediate CRD state.
//     SetOrAddHandler inserts it.
//   - (Installing, WorkloadReady{}) already has a placeholder
//     transitioning to Initializing; we replace the handler with
//     OnWorkloadReady, which additionally supports the middleware
//     → Running fast-path and the intra-Installing phase advance.
//   - (Installing, WorkloadFailed{}) already has a placeholder
//     transitioning to InstallFailed; we replace with the real
//     handler that stamps an OpRecord and an EmitEvent alongside
//     the state flip.
//   - AddReachable(Installing, Installing) declares the self-loop the
//     multi-phase design relies on. Without it, Engine.Step's
//     Allows() check would reject a Result{Next: Installing} return.
//     (Same-state writes are permitted by AllowsSelfWrite, but the
//     OnTick handler returns "" for Next when it advances to the
//     next sub-phase, so this is defensive.)
func registerInstalling(t *domain.TransitionTable) {
	t.AddReachable(appv1.Installing, appv1.Installing)
	t.SetOrAddHandler(appv1.Installing, domain.Tick{}, OnInstallingTick)
	t.SetOrAddHandler(appv1.Installing, domain.WorkloadReady{}, OnInstallingWorkloadReady)
	t.SetOrAddHandler(appv1.Installing, domain.WorkloadFailed{}, OnInstallingWorkloadFailed)
}

// OnInstallingTick is invoked when the Runner delivers a Tick{}
// action to an AM currently in the Installing state. It is the
// FIRST-CLASS entry point that starts the multi-phase install
// pipeline.
//
// Legacy correspondence: the head of pkg/appstate/installing_app.go's
// goroutine — the `twoPhase` branch that decides whether validation
// runs before or after helm.
//
// The pipeline, and the reason it has this shape: helm renders a
// two-phase app's workloads at replicas=0 (see
// appinstaller.HelmOps.SetValues), so the resource-pressure check
// cannot run until the release exists and has told us what it costs.
// Single-phase apps have no such staging and validate up front.
//
//	entry → expose-ports ──┬─ single-phase ─→ validate ─┐
//	                       └─ two-phase ────────────────┴─→ helm-install
//	              helm-install ─┬─ two-phase ─→ post-validate ─┐
//	                            └─ single-phase ───────────────┴─→ scale-up
//	                    scale-up ─┬─ middleware ─→ wait-launch ─→ Running
//	                              └─ otherwise ──→ wait-startup ─→ Initializing
//
//	any non-Stopping failure → InstallingCanceling (terminal intent
//	InstallFailed) → helm uninstall + wait-ns → InstallFailed
//
// Every phase other than the entry preserves an active workload claim.
// If no claim survived a restart, Tick re-dispatches only that phase's
// workload without rewriting status or emitting another event.
func OnInstallingTick(_ context.Context, w domain.World) domain.Result {
	phase := installingPhase(w)

	switch phase {
	case "":
		return startExposePorts(w)

	case PhaseInstallingExposePorts:
		return recoverWorkload(w, exposePortsWorkload())
	case PhaseInstallingValidate:
		return recoverWorkload(w, validateInstallWorkload())
	case PhaseInstallingHelmInstall:
		return recoverWorkload(w, helmInstallWorkload())
	case PhaseInstallingPostValidate:
		return recoverWorkload(w, validateInstallWorkload())
	case PhaseInstallingScaleUp:
		return recoverWorkload(w, scaleUpWorkload())
	case PhaseInstallingWaitStartUp:
		return recoverWorkload(w, waitStartUpWorkload())
	case PhaseInstallingWaitLaunch:
		return recoverWorkload(w, installWaitLaunchWorkload())
	case PhaseInstallingCleanup:
		return recoverWorkload(w, cleanupInstallFailureWorkload())

	default:
		// An unrecognised phase value in Payload is either a bug
		// (typo, missed handler update) or an in-place upgrade
		// landed on a rolled-back binary. Surface it loudly so
		// operators / test suites catch it.
		return domain.Result{
			Err: fmt.Errorf("installing: unknown %s=%q in Payload; expected one of %v",
				PayloadInstallingPhase, phase, installingPhases),
		}
	}
}

// installingPhases is the set of legal PayloadInstallingPhase values
// (excluding ""), used only to render a helpful error message when an
// AM carries something else.
var installingPhases = []string{
	PhaseInstallingExposePorts,
	PhaseInstallingValidate,
	PhaseInstallingHelmInstall,
	PhaseInstallingPostValidate,
	PhaseInstallingScaleUp,
	PhaseInstallingWaitStartUp,
	PhaseInstallingWaitLaunch,
	PhaseInstallingCleanup,
}

// OnInstallingWorkloadReady is invoked when a CallWorkload dispatched
// by the Installing pipeline completes successfully. The current
// sub-phase (read from Payload) selects the next step.
//
// Legacy correspondence: the sequential statements inside
// pkg/appstate/installing_app.go's goroutine. Each "the previous call
// returned nil, do the next one" point in that function is one edge
// here; see the ASCII pipeline on OnInstallingTick.
func OnInstallingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := installingPhase(w)

	switch phase {
	case PhaseInstallingExposePorts:
		// Two-phase apps skip the pre-helm check entirely — running
		// it here would measure a release that does not exist yet
		// and then measure it again in post-validate.
		if w.HasWorkloadReplicas {
			return startHelmInstall(w)
		}
		return startValidate(w)

	case PhaseInstallingValidate:
		return startHelmInstall(w)

	case PhaseInstallingCleanup:
		// Only reachable for an AM a pre-handoff binary parked in the
		// in-state cleanup phase; fresh failures clean up through
		// InstallingCanceling now.
		failMsg, failReason := stashedInstallFailure(w)
		return landInstallFailed(w, failMsg, failReason)

	case PhaseInstallingHelmInstall:
		// The release exists. For a two-phase app it is deployed at
		// replicas=0, which is exactly the state the post-helm check
		// needs to measure.
		if w.HasWorkloadReplicas {
			return startPostValidate(w)
		}
		return startScaleUp(w)

	case PhaseInstallingPostValidate:
		return startScaleUp(w)

	case PhaseInstallingScaleUp:
		// Middleware never goes through Initializing: HelmOps.Install
		// already skips startup polling for it, and legacy waits for
		// launch inline and lands straight in Running.
		if w.AM != nil && w.AM.Spec.Type == appv1.Middleware {
			return startWaitLaunch(w)
		}
		return startWaitStartUp(w)

	case PhaseInstallingWaitStartUp:
		return finishInstall(w)

	case PhaseInstallingWaitLaunch:
		return finishMiddlewareInstall(w)

	case "":
		// A WorkloadReady with no phase context is a Runner bug or
		// a stale queued action from a pre-phase-tracking binary.
		// Fail loudly so the invariant "phase Payload is set
		// before every CallWorkload" gets fixed instead of
		// silently mis-routing.
		return domain.Result{
			Err: fmt.Errorf("installing: WorkloadReady with empty %s; every CallWorkload emitted from Installing must stamp the phase first",
				PayloadInstallingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("installing: WorkloadReady with unknown %s=%q",
				PayloadInstallingPhase, phase),
		}
	}
}

// OnInstallingWorkloadFailed is invoked when a CallWorkload
// dispatched by the Installing pipeline fails (either the workload
// itself errored, or, for ValidateInstall, the validator returned
// !decision.OK — the Runner unifies both via
// validationRejectedError).
//
// Not every failure is an install failure. Two of them leave the
// helm release deployed and route the AM to Stopping instead, so the
// user can free up capacity and resume rather than reinstall from
// scratch:
//
//   - pods cannot be scheduled. The adapter classified it (see
//     helmops.classifyInstallErr) and the Runner forwarded the
//     verdict on WorkloadFailed.LandedState. That covers both helm
//     install and WaitForStartUp: ErrPodPending or
//     ErrServerSidePodPending, with ErrHamiUnschedulable only
//     changing the reason. The sentinel is wrapped inside the
//     workload error and unreadable from here.
//   - the post-helm resource check rejected. Legacy additionally
//     converts the operation itself into a stop, so the AM's
//     Spec.OpType stops advertising an install that is not going to
//     finish.
//
// Any other WaitForStartUp error enters InstallingCanceling with
// reason StartUpFailed. Cleanup then lands InstallFailed with that
// same reason.
//
// A middleware launch timeout is the odd one out: it enters
// InstallingCanceling WITHOUT a terminal intent, so it keeps landing
// on the user-cancel terminal the legacy stack used.
//
// Every other failure hands off to InstallingCanceling with the
// InstallFailed terminal intent, so the release tear-down and the
// namespace wait run once, in one pipeline, before the AM is allowed
// to land in InstallFailed.
func OnInstallingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	phase := installingPhase(w)

	// Prefer failure data carried on the Action itself (populated
	// by the Runner from WorkloadResult.LandedReason /
	// err.Error()). Fall back to a phase-derived pair so we always
	// have SOMETHING to write, even for tests / production paths
	// that dispatch WorkloadFailed with empty Reason / Message.
	msg, reason := failureFromAction(w.Action)
	if msg == "" || reason == "" {
		fbMsg, fbReason := failureFallback(phase)
		if msg == "" {
			msg = fbMsg
		}
		if reason == "" {
			reason = fbReason
		}
	}
	wf, _ := w.Action.(domain.WorkloadFailed)

	switch {
	case wf.LandedState == appv1.Stopping:
		// Unschedulable pods during helm install. The operation
		// stays an install (legacy does not rewrite Spec.OpType on
		// this path) — only the state moves.
		return landStopping(w, msg, reason, wf.StopAll, false)

	case phase == PhaseInstallingPostValidate:
		return landStopping(w, msg, reason, wf.StopAll, true)

	case phase == PhaseInstallingWaitStartUp:
		// Unschedulable already returned above via LandedState.
		if msg == "" {
			msg = "installing failed in phase \"wait-startup\""
		}
		return enterInstallingCanceling(w, msg, msg, constants.AppStopDueToStartUpFailed, constants.AppStopDueToStartUpFailed)

	case phase == PhaseInstallingWaitLaunch:
		return landMiddlewareLaunchFailed(w)

	case phase == PhaseInstallingCleanup:
		// Cleanup itself failed or timed out. Legacy still
		// transitions to InstallFailed so the state machine
		// cannot stall; operators see the original reason.
		failMsg, failReason := stashedInstallFailure(w)
		return landInstallFailed(w, failMsg, failReason)
	}

	return handoffToInstallingCanceling(w, msg, reason)
}

// startExposePorts is the first install step: assign host expose
// ports and persist them into Spec.Config, matching
// installing_app.go:66-83.
func startExposePorts(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgExposePortsStarted,
				Reason:  installingReasonExposePorts,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingExposePorts},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonExposePorts,
				Message: installingMsgExposePortsStarted,
			},
			exposePortsWorkload(),
		},
	}
}

func exposePortsWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadSetExposePorts,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "SetExposePortsFailed"},
	}
}

// handoffToInstallingCanceling tears a failed install down through the
// SAME pipeline a user cancel uses (helm uninstall → wait for the
// namespace to be reaped) instead of an in-state cleanup workload.
// PayloadInstallingCancelingTerminal is what keeps the two callers
// distinguishable: it tells OnInstallingCancelingWorkloadReady to land
// InstallFailed rather than InstallingCanceled once the cleanup is
// done. The original diagnostic rides along so that terminal write
// reports the install failure, not the cancel.
//
// No CallWorkload here: the InstallingCanceling Tick owns the
// uninstall dispatch, and its Preempts claim is what stops the
// still-running install workload.
func handoffToInstallingCanceling(w domain.World, msg, reason string) domain.Result {
	return enterInstallingCanceling(w, installingMsgCleanupStarted, msg, reason, installingReasonCleanupStarted)
}

// enterInstallingCanceling moves a failed install into the cleanup
// pipeline. statusReason is the reason operators see on this
// transition. failReason is written again when cleanup lands
// InstallFailed. WaitForStartUp passes StartUpFailed for both.
func enterInstallingCanceling(_ domain.World, statusMsg, failMsg, failReason, statusReason string) domain.Result {
	return domain.Result{
		Next: appv1.InstallingCanceling,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallingCanceling,
				Message: statusMsg,
				Reason:  statusReason,
				Payload: map[string]string{
					// Blank rather than absent: Payload merges
					// key-by-key, so the Installing sub-phase would
					// otherwise survive into InstallingCanceling.
					PayloadInstallingPhase:             "",
					PayloadInstallingCancelingTerminal: string(appv1.InstallFailed),
					PayloadInstallingFailMessage:       failMsg,
					PayloadInstallingFailReason:        failReason,
				},
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  statusReason,
				Message: statusMsg,
			},
		},
	}
}

func cleanupInstallFailureWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadCleanupInstallFailure,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "InstallCleanupFailed"},
	}
}

// landInstallFailed is the write after the installing-cancel pipeline
// has already torn the release down. Cleanup is marked done so the
// next InstallFailed Tick moves to Uninstalled instead of cleaning again.
func landInstallFailed(w domain.World, msg, reason string) domain.Result {
	if msg == "" {
		msg = "installing failed"
	}
	if reason == "" {
		reason = string(appv1.InstallFailed)
	}
	now := metav1.NewTime(w.Now)
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.InstallFailed,
		StateTime: &now,
	}
	return domain.Result{
		Next: appv1.InstallFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{
					PayloadInstallFailedCleanup: PhaseInstallFailedCleanupDone,
				},
			}, DeletePayloadKeys: []string{
				PayloadInstallingPhase,
				PayloadInstallingCancelingPhase,
				PayloadInstallingCancelingTerminal,
				PayloadInstallingFailMessage,
				PayloadInstallingFailReason,
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// stashedInstallFailure reads the diagnostic parked by
// handoffToInstallingCanceling. Falls back to a generic pair when the
// keys are missing (cleanup-phase WorkloadFailed with no prior stash).
func stashedInstallFailure(w domain.World) (msg, reason string) {
	if w.AM != nil && w.AM.Status.Payload != nil {
		msg = w.AM.Status.Payload[PayloadInstallingFailMessage]
		reason = w.AM.Status.Payload[PayloadInstallingFailReason]
	}
	if msg == "" {
		msg, reason = failureFromAction(w.Action)
	}
	if msg == "" {
		msg = "installing failed"
	}
	if reason == "" {
		reason = string(appv1.InstallFailed)
	}
	return msg, reason
}

// startValidate is the effect bundle for the fresh-entry branch of
// OnInstallingTick. Factored out so both the handler and its unit
// tests can construct the exact same Result.
func startValidate(w domain.World) domain.Result {
	return domain.Result{
		// Next is left blank: the state stays as Installing while
		// we transition Payload from "" to "validate". Runner will
		// still apply the PatchAM because we included one
		// explicitly.
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgValidateStarted,
				Reason:  installingReasonEntered,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingValidate},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonEntered,
				Message: installingMsgValidateStarted,
			},
			validateInstallWorkload(),
		},
	}
}

func validateInstallWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadValidateInstall,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "InstallValidationRejected"},
	}
}

// startHelmInstall is the effect bundle for the validate-→-helm
// transition inside OnInstallingWorkloadReady.
func startHelmInstall(w domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgHelmStarted,
				Reason:  installingReasonHelmStarted,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingHelmInstall},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonHelmStarted,
				Message: installingMsgHelmStarted,
			},
			helmInstallWorkload(),
		},
	}
}

func helmInstallWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadInstall,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmInstallFailed"},
	}
}

// startPostValidate is the two-phase-only resource check that runs
// against the deployed-but-scaled-to-zero release.
//
// It dispatches the same WorkloadValidateInstall kind as
// startValidate; the phase in Payload is what tells the callback
// handlers which of the two checks just answered.
func startPostValidate(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgPostValidateStarted,
				Reason:  installingReasonPostValidate,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingPostValidate},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonPostValidate,
				Message: installingMsgPostValidateStarted,
			},
			validateInstallWorkload(),
		},
	}
}

// startScaleUp restores the manifest-declared replica counts via a
// second helm upgrade — WorkloadResume is helm Scale(-1).
//
// This is the step whose absence left two-phase apps deployed at
// replicas=0: helm install had succeeded, the AM advanced to
// Initializing, and the readiness wait there polled forever for pods
// nothing was ever going to create.
//
// Dispatched for single-phase apps too. HelmOps.Scale short-circuits
// to a no-op when the app declares no workloadReplicas and
// HelmOpsV2.Scale is a no-op outright, so an unconditional dispatch
// costs one callback and keeps this handler from carrying a copy of
// those adapters' opt-out conditions.
func startScaleUp(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgScaleUpStarted,
				Reason:  installingReasonScaleUpStarted,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingScaleUp},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonScaleUpStarted,
				Message: installingMsgScaleUpStarted,
			},
			scaleUpWorkload(),
		},
	}
}

func scaleUpWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadResume,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "InstallScaleUpFailed"},
	}
}

// startWaitStartUp polls the pods the scale-up just created for
// readiness. Non-middleware apps only.
func startWaitStartUp(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgWaitStartUpStarted,
				Reason:  installingReasonWaitStartUp,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingWaitStartUp},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonWaitStartUp,
				Message: installingMsgWaitStartUpStarted,
			},
			waitStartUpWorkload(),
		},
	}
}

func waitStartUpWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadWaitStartUp,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{
			Reason: constants.AppStopDueToStartUpFailed,
		},
	}
}

// startWaitLaunch is the middleware counterpart of startWaitStartUp.
// Middleware bypasses Initializing entirely, so this is the last
// phase before Running.
func startWaitLaunch(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Installing,
				Message: installingMsgWaitLaunchStarted,
				Reason:  installingReasonWaitLaunch,
				Payload: map[string]string{PayloadInstallingPhase: PhaseInstallingWaitLaunch},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonWaitLaunch,
				Message: installingMsgWaitLaunchStarted,
			},
			installWaitLaunchWorkload(),
		},
	}
}

func installWaitLaunchWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadWaitLaunch,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{
			Reason: constants.AppStopDueToStartUpFailed,
		},
	}
}

// finishInstall is the non-middleware terminal branch: pods are
// ready, so hand off to Initializing, which runs its own
// WaitForLaunch before declaring the app Running.
func finishInstall(w domain.World) domain.Result {
	successMsg := fmt.Sprintf(constants.InstallOperationCompletedTpl, w.AM.Spec.Type, w.AM.Spec.AppName)
	now := metav1.NewTime(w.Now)
	// See OnInstallingWorkloadFailed for the note on why Version is
	// left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   successMsg,
		Source:    w.AM.Spec.Source,
		Status:    appv1.Initializing,
		StateTime: &now,
	}

	// Effect order matters: AppendOpRecord precedes PatchAM so the
	// two land in the same PatchStatus RPC. See the same-shaped
	// note in OnInstallingWorkloadFailed for the RV-conflict
	// hazard when the order is swapped.
	return domain.Result{
		Next: appv1.Initializing,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Initializing,
				Message: successMsg,
				Reason:  installingReasonSucceeded,
				Payload: map[string]string{}, // clear phase; leaving Installing
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  installingReasonSucceeded,
				Message: successMsg,
			},
		},
	}
}

// finishMiddlewareInstall is the middleware terminal branch: the
// launch wait already confirmed readiness, so the AM goes straight to
// Running without passing through Initializing.
func finishMiddlewareInstall(w domain.World) domain.Result {
	successMsg := fmt.Sprintf(constants.InstallOperationCompletedTpl, w.AM.Spec.Type, w.AM.Spec.AppName)
	now := metav1.NewTime(w.Now)
	// See OnInstallingWorkloadFailed for the note on why Version is
	// left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   successMsg,
		Source:    w.AM.Spec.Source,
		Status:    appv1.Running,
		StateTime: &now,
	}

	// AppendOpRecord precedes PatchAM so both land in one
	// PatchStatus RPC — see finishInstall for the RV-conflict note.
	return domain.Result{
		Next: appv1.Running,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Running,
				Message: string(appv1.Running),
				Reason:  string(appv1.Running),
				Payload: map[string]string{}, // clear phase; leaving Installing
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  string(appv1.Running),
				Message: successMsg,
			},
		},
	}
}

// landStopping is the shared tail for every Installing failure that
// leaves the helm release in the cluster. The app ends up deployed
// but scaled to zero, which is a state the user can recover from with
// a resume once capacity frees up — unlike InstallFailed, which tears
// the release down.
//
// stopAll stamps the flag that widens the stop to the shared server
// the app owns; convertOp additionally rewrites the operation itself
// from an install into a stop, which legacy does only on the
// post-helm rejection path (an install that will not finish should
// stop advertising itself as in progress).
//
// No OpRecord is written: legacy passes nil on all of these paths,
// since the operation has not reached a terminal outcome yet — the
// Stopping state's own handler records one when the stop completes.
func landStopping(w domain.World, msg, reason string, stopAll, convertOp bool) domain.Result {
	var effects []domain.Effect

	// PatchSpec first: it bumps the AM's ResourceVersion and the
	// Runner refreshes its cached copy afterwards, so the status
	// patch below still satisfies its optimistic-lock precondition.
	// See domain.PatchSpec's ordering note.
	if stopAll || convertOp {
		spec := domain.PatchSpec{}
		if convertOp {
			spec.OpType = appv1.StopOp
		}
		if stopAll {
			spec.SetAnnotations = map[string]string{AnnotStopAll: "true"}
		}
		effects = append(effects, spec)
	}

	status := appv1.ApplicationManagerStatus{
		State:   appv1.Stopping,
		Message: msg,
		Reason:  reason,
		// Payload is cleared: phase tracking only applies while
		// state == Installing. An empty non-nil map makes the
		// store's field-level merge drop any leftover keys.
		Payload: map[string]string{},
	}
	if convertOp {
		status.OpType = appv1.StopOp
	}

	effects = append(effects,
		domain.PatchAM{Status: status},
		domain.EmitEvent{
			Type:    "Warning",
			Reason:  reason,
			Message: msg,
		},
	)

	return domain.Result{Next: appv1.Stopping, Effects: effects}
}

// landMiddlewareLaunchFailed handles a middleware that never
// launched. Legacy routes this to InstallingCanceling rather than
// InstallFailed or Stopping: a middleware that will not come up has
// to have its release rolled back, and the InstallingCanceling
// handler is what owns that tear-down.
//
// The message is the state name, matching legacy verbatim — the
// actionable detail is in the Reason.
func landMiddlewareLaunchFailed(_ domain.World) domain.Result {
	msg := string(appv1.InstallingCanceling)
	return domain.Result{
		Next: appv1.InstallingCanceling,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.InstallingCanceling,
				Message: msg,
				Reason:  constants.AppStopDueToStartUpFailed,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  constants.AppStopDueToStartUpFailed,
				Message: msg,
			},
		},
	}
}

// failureFromAction pulls the (message, reason) pair off a
// WorkloadFailed action. Runner populates these from
// WorkloadResult.LandedReason and err.Error() when the workload
// terminates. Returns ("", "") for any non-WorkloadFailed action
// (e.g. a bare Tick), which lets the caller fall back cleanly.
func failureFromAction(a domain.Action) (msg, reason string) {
	wf, ok := a.(domain.WorkloadFailed)
	if !ok {
		return "", ""
	}
	return wf.Message, wf.Reason
}

// failureFallback derives a fixed (message, reason) pair from the
// sub-phase when the Action carries no failure data. Kept phase-
// specific rather than a single "install failed" string because
// operators skimming events benefit from knowing whether the
// failure was pre-helm (rejected) or during helm apply.
func failureFallback(phase string) (msg, reason string) {
	switch phase {
	case PhaseInstallingExposePorts:
		reason = "SetExposePortsFailed"
	case PhaseInstallingValidate, PhaseInstallingPostValidate:
		reason = "InstallValidationRejected"
	case PhaseInstallingHelmInstall:
		reason = "HelmInstallFailed"
	case PhaseInstallingScaleUp:
		reason = "InstallScaleUpFailed"
	case PhaseInstallingWaitStartUp, PhaseInstallingWaitLaunch:
		reason = constants.AppStopDueToStartUpFailed
	case PhaseInstallingCleanup:
		reason = "InstallCleanupFailed"
	default:
		reason = string(appv1.InstallFailed)
	}
	msg = fmt.Sprintf("installing failed in phase %q", phase)
	return msg, reason
}

// installingPhase reads Payload[PayloadInstallingPhase] with a safe
// nil-map guard. Returns "" both when the AM has no Payload and when
// the key is not set — either case means "no sub-phase yet", which
// is the fresh-entry branch.
func installingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadInstallingPhase]
}
