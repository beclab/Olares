package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the Uninstalling-state messages. Same
// rationale as installing.go — inline copies simplify diffing
// against pkg/appstate/uninstalling_app.go.
const (
	uninstallingMsgHelmStarted   = "uninstalling.helm: helm release delete + cleanup in progress"
	uninstallingMsgWaitNSStarted = "uninstalling.wait-ns: waiting for namespace garbage collection"
	uninstallingReasonEntered    = "UninstallStarted"
	uninstallingReasonHelmDone   = "UninstallHelmCompleted"
	uninstallingReasonSucceeded  = "UninstallCompleted"
)

// registerUninstalling wires the Uninstalling-state handlers into
// the TransitionTable. Same shape as registerInstalling.
//
// Wiring semantics:
//
//   - (Uninstalling, Tick{}) — a fresh edge; the default table
//     does not declare it because Uninstalling is a workload-
//     callback state in legacy (no Tick-driven action). The two-
//     phase port needs it to drive phase-to-phase advance
//     via CallWorkload chaining. SetOrAddHandler inserts it.
//
//   - (Uninstalling, WorkloadReady{}) already has a placeholder
//     targeting Uninstalled. Replace with the real handler which
//     branches on phase to either advance the phase or complete.
//
//   - (Uninstalling, WorkloadFailed{}) already has a placeholder
//     targeting UninstallFailed. Replace with the real handler.
//
//   - AddReachable(Uninstalling, Uninstalling) declares the
//     phase-progression self-loop (Result{Next:""} while phase
//     advances internally).
//
// Cancel path — (Uninstalling, TriggerOp{CancelOp}) → UninstallFailed
// per the legacy Cancel method — is NOT ported here; the current
// scope does not include mid-uninstall cancellation. Follow-up.
func registerUninstalling(t *domain.TransitionTable) {
	t.AddReachable(appv1.Uninstalling, appv1.Uninstalling)
	t.SetOrAddHandler(appv1.Uninstalling, domain.Tick{}, OnUninstallingTick)
	t.SetOrAddHandler(appv1.Uninstalling, domain.WorkloadReady{}, OnUninstallingWorkloadReady)
	t.SetOrAddHandler(appv1.Uninstalling, domain.WorkloadFailed{}, OnUninstallingWorkloadFailed)
}

// OnUninstallingTick is the Tick handler that starts the uninstall
// pipeline on fresh entry into Uninstalling.
//
// Legacy correspondence: pkg/appstate/uninstalling_app.go:49-93 —
// the `Exec` method spawning a goroutine that runs
// helm uninstall → compute release → waitForDeleteNamespace →
// (optional) waitForDeleteSharedNamespaces.
//
// The port collapses these into TWO CallWorkload phases
// (helm and wait-ns), with the fast synchronous cleanups
// (compute release, PVC deletion, permission unregister) bundled
// INTO the Uninstall port's adapter — see ports/workload.go
// Uninstall doc-comment for the full list.
//
// Phase semantics:
//
//   - phase == ""                → dispatch CallWorkload{Uninstall};
//     stamp phase=helm.
//   - phase == helm              → preserve an outstanding claim, or
//     re-dispatch only CallWorkload{Uninstall} after restart.
//   - phase == wait-ns           → preserve an outstanding claim, or
//     re-dispatch only CallWorkload{WaitNamespaceGone} after restart.
//   - unknown phase              → hard fail via Err (safety net).
//
// UninstallAll, compute.DeleteForTarget, shared-namespace wait,
// and the os-platform/mongodb special case run inside the
// Runner's dispatchUninstall / WaitForNamespaceDeleted.
//
// Mid-uninstall Cancel (legacy Cancel method: flip to
// UninstallFailed) is still not ported.
func OnUninstallingTick(_ context.Context, w domain.World) domain.Result {
	phase := uninstallingPhase(w)

	switch phase {
	case "":
		return startHelmUninstall(w)

	case PhaseUninstallingHelm:
		return recoverWorkload(w, uninstallWorkload())
	case PhaseUninstallingWaitNS:
		return recoverWorkload(w, waitNamespaceGoneWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("uninstalling: unknown %s=%q; expected one of {\"\", %q, %q}",
				PayloadUninstallingPhase, phase, PhaseUninstallingHelm, PhaseUninstallingWaitNS),
		}
	}
}

// OnUninstallingWorkloadReady advances the pipeline through its
// two phases:
//
//   - phase == helm    → helm uninstall completed; dispatch the
//     namespace-gone wait; stamp phase=wait-ns.
//   - phase == wait-ns → namespace fully deleted; write the
//     terminal OpRecord and advance to Uninstalled.
func OnUninstallingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := uninstallingPhase(w)

	switch phase {
	case PhaseUninstallingHelm:
		return startWaitNamespaceGone(w)

	case PhaseUninstallingWaitNS:
		return finishToUninstalled(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("uninstalling: WorkloadReady with empty %s; every CallWorkload emitted from Uninstalling must stamp the phase first",
				PayloadUninstallingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("uninstalling: WorkloadReady with unknown %s=%q",
				PayloadUninstallingPhase, phase),
		}
	}
}

// OnUninstallingWorkloadFailed handles a failed CallWorkload from
// EITHER phase. Both phases route to UninstallFailed with an
// OpRecord + Warning event carrying the workload's failure data.
//
// Legacy correspondence: pkg/appstate/uninstalling_app.go:66-77 —
// the `err != nil` branch of the goroutine.
//
// A phase-specific fallback reason is still provided in case the
// Runner did not backfill a LandedReason (rare).
func OnUninstallingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	phase := uninstallingPhase(w)
	msg, reason := failureFromAction(w.Action)
	if msg == "" || reason == "" {
		fbMsg, fbReason := uninstallFailureFallback(phase)
		if msg == "" {
			msg = fbMsg
		}
		if reason == "" {
			reason = fbReason
		}
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.UninstallFailed,
		StateTime: &now,
	}

	// Effect ordering: AppendOpRecord precedes PatchAM so both
	// land atomically. See installing.go for full RV-conflict
	// note.
	return domain.Result{
		Next: appv1.UninstallFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UninstallFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{}, // clear phase; leaving Uninstalling
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

// startHelmUninstall is the effect bundle for the fresh-entry
// branch of OnUninstallingTick.
func startHelmUninstall(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Uninstalling,
				Message: uninstallingMsgHelmStarted,
				Reason:  uninstallingReasonEntered,
				Payload: map[string]string{PayloadUninstallingPhase: PhaseUninstallingHelm},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  uninstallingReasonEntered,
				Message: uninstallingMsgHelmStarted,
			},
			uninstallWorkload(),
		},
	}
}

func uninstallWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadUninstall,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
	}
}

// startWaitNamespaceGone is the effect bundle for the helm-→-wait
// phase advance inside OnUninstallingWorkloadReady.
func startWaitNamespaceGone(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Uninstalling,
				Message: uninstallingMsgWaitNSStarted,
				Reason:  uninstallingReasonHelmDone,
				Payload: map[string]string{PayloadUninstallingPhase: PhaseUninstallingWaitNS},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  uninstallingReasonHelmDone,
				Message: uninstallingMsgWaitNSStarted,
			},
			waitNamespaceGoneWorkload(),
		},
	}
}

func waitNamespaceGoneWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadWaitNamespaceGone,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "NamespaceGoneWaitFailed"},
	}
}

// finishToUninstalled is the terminal branch of
// OnUninstallingWorkloadReady after both phases succeed.
//
// Emits the terminal OpRecord and advances to Uninstalled. The
// Uninstalled state's own handler (follow-up PR) will own any
// remaining cleanup / final housekeeping.
func finishToUninstalled(w domain.World) domain.Result {
	successMsg := fmt.Sprintf(constants.UninstallOperationCompletedTpl, w.AM.Spec.Type, w.AM.Spec.AppName)
	now := metav1.NewTime(w.Now)

	// See installing.go for the "OpRecord.Version left empty at
	// handler layer" note.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   successMsg,
		Source:    w.AM.Spec.Source,
		Status:    appv1.Uninstalled,
		StateTime: &now,
	}

	return domain.Result{
		Next: appv1.Uninstalled,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Uninstalled,
				Message: string(appv1.Uninstalled),
				Reason:  string(appv1.Uninstalled),
				Payload: map[string]string{}, // clear phase; leaving Uninstalling
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  uninstallingReasonSucceeded,
				Message: successMsg,
			},
		},
	}
}

// uninstallFailureFallback returns a phase-specific (message,
// reason) pair for OnUninstallingWorkloadFailed when the Action
// carries no data.
func uninstallFailureFallback(phase string) (msg, reason string) {
	switch phase {
	case PhaseUninstallingHelm:
		reason = "HelmUninstallFailed"
	case PhaseUninstallingWaitNS:
		reason = "NamespaceGoneWaitFailed"
	default:
		reason = string(appv1.UninstallFailed)
	}
	msg = fmt.Sprintf("uninstalling failed in phase %q", phase)
	return msg, reason
}

// uninstallingPhase reads Payload[PayloadUninstallingPhase] with
// the standard nil-map guard.
func uninstallingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadUninstallingPhase]
}
