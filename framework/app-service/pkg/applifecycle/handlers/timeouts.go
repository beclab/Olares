package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// registerTimeouts wires TimeoutFired for every state that dispatches
// CallWorkload. The Runner only delivers the action after the persisted
// state deadline; these handlers choose the landing state.
func registerTimeouts(t *domain.TransitionTable) {
	for _, state := range []appv1.ApplicationManagerState{
		appv1.Downloading,
		appv1.UpgradeDownloading,
		appv1.Installing,
		appv1.Initializing,
		appv1.Upgrading,
		appv1.ApplyingEnv,
		appv1.Resuming,
		appv1.Stopping,
		appv1.Uninstalling,
		appv1.PendingCanceling,
		appv1.DownloadingCanceling,
		appv1.UpgradePendingCanceling,
		appv1.UpgradeDownloadingCanceling,
		appv1.InstallingCanceling,
		appv1.UpgradingCanceling,
	} {
		t.SetOrAddHandler(state, domain.TimeoutFired{}, OnTimeout)
	}
}

// OnTimeout is the TimeoutFired handler for every deadline-enabled
// state. Installing is the exception: timeout is an install failure,
// so cleanup still runs through InstallingCanceling and the stashed
// intent makes the successful cleanup land InstallFailed.
func OnTimeout(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("timeout: handler invoked with nil AM")}
	}
	switch w.AM.Status.State {
	case appv1.Downloading:
		return timeoutToCanceling(w, appv1.DownloadingCanceling, constants.DownloadCancelBySystem, constants.DownloadCanceledByTimeout)
	case appv1.UpgradeDownloading:
		return timeoutToCanceling(w, appv1.UpgradeDownloadingCanceling, constants.UpgradeCancelBySystem, constants.UpgradeCanceledByTimeout)
	case appv1.Installing:
		return handoffToInstallingCanceling(w, constants.InstallCanceledByTimeout, constants.InstallCancelBySystem)
	case appv1.Initializing:
		return timeoutToCanceling(w, appv1.InitializingCanceling, constants.InitializeCancelBySystem, constants.InitializeCanceledByTimeout)
	case appv1.Upgrading:
		return timeoutToCanceling(w, appv1.UpgradingCanceling, constants.UpgradeCancelBySystem, constants.UpgradeCanceledByTimeout)
	case appv1.ApplyingEnv:
		return timeoutToCanceling(w, appv1.ApplyingEnvCanceling, constants.ApplyEnvCancelBySystem, constants.ApplyEnvCanceledByTimeout)
	case appv1.Resuming:
		return timeoutToCanceling(w, appv1.ResumingCanceling, constants.ResumeCancelBySystem, constants.ResumeCanceledByTimeout)
	case appv1.Stopping:
		return timeoutToFailed(w, appv1.StopFailed, "StopTimedOut", "stop timed out")
	case appv1.Uninstalling:
		return timeoutToFailed(w, appv1.UninstallFailed, "UninstallTimedOut", "uninstall timed out")
	case appv1.PendingCanceling:
		return timeoutToFailed(w, appv1.PendingCancelFailed, "PendingCancelTimedOut", "pending cancel timed out")
	case appv1.DownloadingCanceling:
		return timeoutToFailed(w, appv1.DownloadingCancelFailed, "DownloadingCancelTimedOut", "downloading cancel timed out")
	case appv1.UpgradePendingCanceling:
		return timeoutToFailed(w, appv1.UpgradePendingCancelFailed, "UpgradePendingCancelTimedOut", "upgrade pending cancel timed out")
	case appv1.UpgradeDownloadingCanceling:
		return timeoutToFailed(w, appv1.UpgradeDownloadingCancelFailed, "UpgradeDownloadingCancelTimedOut", "upgrade downloading cancel timed out")
	case appv1.InstallingCanceling:
		return timeoutInstallingCanceling(w)
	case appv1.UpgradingCanceling:
		return timeoutToFailed(w, appv1.UpgradingCancelFailed, "UpgradingCancelTimedOut", "upgrading cancel timed out")
	default:
		return domain.Result{Err: fmt.Errorf("timeout: %s has no TimeoutFired handler", w.AM.Status.State)}
	}
}

func timeoutToCanceling(w domain.World, next appv1.ApplicationManagerState, reason, msg string) domain.Result {
	return domain.Result{
		Next: next,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   next,
				Reason:  reason,
				Message: msg,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{Type: "Normal", Reason: reason, Message: msg},
		},
	}
}

func timeoutToFailed(w domain.World, next appv1.ApplicationManagerState, reason, msg string) domain.Result {
	now := metav1.NewTime(w.Now)
	record := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    next,
		StateTime: &now,
	}
	return domain.Result{
		Next: next,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: record},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   next,
				Reason:  reason,
				Message: msg,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{Type: "Warning", Reason: reason, Message: msg},
		},
	}
}

// timeoutInstallingCanceling lands InstallingCancelFailed but keeps an
// ordinary install-failure intent so a later retry still finishes at
// InstallFailed. A user cancel has nothing to carry.
func timeoutInstallingCanceling(w domain.World) domain.Result {
	res := timeoutToFailed(w, appv1.InstallingCancelFailed, "InstallingCancelTimedOut", "installing cancel timed out")
	for i, effect := range res.Effects {
		patch, ok := effect.(domain.PatchAM)
		if !ok {
			continue
		}
		patch.Status.Payload = carryInstallFailureIntent(w, nil)
		patch.DeletePayloadKeys = []string{PayloadInstallingCancelingPhase}
		res.Effects[i] = patch
		break
	}
	return res
}
