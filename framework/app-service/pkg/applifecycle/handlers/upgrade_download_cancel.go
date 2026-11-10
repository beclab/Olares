package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

const (
	upgradeDownloadCancelMsgPending = "canceling upgrade while it waits for the download slot"
	upgradeDownloadCancelMsgPull    = "canceling upgrade image pull"
	upgradePendingCancelReason      = "UpgradePendingCanceled"
	upgradeDownloadCancelReason     = "UpgradeDownloadCanceled"
	upgradeDownloadCancelPullReason = "UpgradeDownloadCancelPull"
)

// registerUpgradeDownloadCancel wires cancel of the queue and of the
// image pull. Both land back on the pre-upgrade steady state
// (Stopped when that annotation says so, otherwise Running) instead
// of Stopping. Canceling an upgrade that has already reached helm
// stays on UpgradingCanceling.
func registerUpgradeDownloadCancel(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradePendingCanceling, appv1.UpgradePendingCanceling)
	t.SetOrAddHandler(appv1.UpgradePending, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradePendingCancel)
	t.SetOrAddHandler(appv1.UpgradePendingCanceling, domain.Tick{}, OnUpgradePendingCancelingTick)

	t.AddReachable(appv1.UpgradeDownloadingCanceling, appv1.UpgradeDownloadingCanceling)
	t.SetOrAddHandler(appv1.UpgradeDownloading, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradeDownloadingCancel)
	t.SetOrAddHandler(appv1.UpgradeDownloadingCanceling, domain.Tick{}, OnUpgradeDownloadingCancelingTick)
	t.SetOrAddHandler(appv1.UpgradeDownloadingCanceling, domain.WorkloadReady{}, OnUpgradeDownloadingCancelingWorkloadReady)
	t.SetOrAddHandler(appv1.UpgradeDownloadingCanceling, domain.WorkloadFailed{}, OnUpgradeDownloadingCancelingWorkloadFailed)

	t.AddReachable(appv1.UpgradePendingCancelFailed, appv1.UpgradePendingCancelFailed)
	t.SetOrAddHandler(appv1.UpgradePendingCancelFailed, domain.Tick{}, OnUpgradePendingCancelFailedTick)
	t.SetOrAddHandler(appv1.UpgradePendingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradePendingCancelFailedCancel)

	t.AddReachable(appv1.UpgradeDownloadingCancelFailed, appv1.UpgradeDownloadingCancelFailed)
	t.SetOrAddHandler(appv1.UpgradeDownloadingCancelFailed, domain.Tick{}, OnUpgradeDownloadingCancelFailedTick)
	t.SetOrAddHandler(appv1.UpgradeDownloadingCancelFailed, domain.TriggerOp{Op: appv1.CancelOp}, OnUpgradeDownloadingCancelFailedCancel)
}

func OnUpgradePendingCancel(_ context.Context, w domain.World) domain.Result {
	return cancelToCanceling(w, appv1.UpgradePendingCanceling, upgradeDownloadCancelMsgPending)
}

// OnUpgradePendingCancelingTick restores the pre-upgrade state. The
// queue has not started an image pull, so there is nothing to cancel.
func OnUpgradePendingCancelingTick(_ context.Context, w domain.World) domain.Result {
	return restorePreUpgradeState(w, upgradePendingCancelReason)
}

func OnUpgradeDownloadingCancel(_ context.Context, w domain.World) domain.Result {
	return cancelToCanceling(w, appv1.UpgradeDownloadingCanceling, upgradeDownloadCancelMsgPull)
}

func OnUpgradeDownloadingCancelingTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("upgrade-downloading-canceling: Tick invoked with nil AM")}
	}
	switch upgradeDownloadingCancelingPhase(w) {
	case "":
		return startUnlessSameKind(w, domain.WorkloadCancelImagePull, startUpgradeDownloadCancelImagePull(w))
	case PhaseCancelImagePull:
		return recoverWorkload(w, cancelImagePullWorkload())
	default:
		return domain.Result{
			Err: fmt.Errorf("upgrade-downloading-canceling: unknown %s=%q",
				PayloadUpgradeDownloadingCancelingPhase, upgradeDownloadingCancelingPhase(w)),
		}
	}
}

func OnUpgradeDownloadingCancelingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	if upgradeDownloadingCancelingPhase(w) != PhaseCancelImagePull {
		return domain.Result{
			Err: fmt.Errorf("upgrade-downloading-canceling: WorkloadReady with %s=%q; expected %q",
				PayloadUpgradeDownloadingCancelingPhase, upgradeDownloadingCancelingPhase(w), PhaseCancelImagePull),
		}
	}
	return restorePreUpgradeState(w, upgradeDownloadCancelReason)
}

func OnUpgradeDownloadingCancelingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "upgrade download cancel: image pull cancel failed"
	}
	if reason == "" {
		reason = string(appv1.UpgradeDownloadingCancelFailed)
	}
	return domain.Result{
		Next: appv1.UpgradeDownloadingCancelFailed,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeDownloadingCancelFailed,
				Message: msg,
				Reason:  reason,
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

func OnUpgradePendingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

func OnUpgradePendingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.UpgradePendingCanceling, terminalReasonCancelRetry, "user requested cancel retry; transitioning back to UpgradePendingCanceling")
}

func OnUpgradeDownloadingCancelFailedTick(ctx context.Context, w domain.World) domain.Result {
	return noopTerminalTick(ctx, w)
}

func OnUpgradeDownloadingCancelFailedCancel(_ context.Context, w domain.World) domain.Result {
	return terminalOpTransition(w, appv1.UpgradeDownloadingCanceling, terminalReasonCancelRetry, "user requested cancel retry; transitioning back to UpgradeDownloadingCanceling")
}

func startUpgradeDownloadCancelImagePull(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeDownloadingCanceling,
				Reason:  upgradeDownloadCancelPullReason,
				Message: upgradeDownloadCancelMsgPull,
				Payload: map[string]string{PayloadUpgradeDownloadingCancelingPhase: PhaseCancelImagePull},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradeDownloadCancelPullReason,
				Message: upgradeDownloadCancelMsgPull,
			},
			cancelImagePullWorkload(),
		},
	}
}

func restorePreUpgradeState(w domain.World, reason string) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("upgrade-download: restore invoked with nil AM")}
	}
	landed := appv1.Running
	if upgradeFromStopped(w) {
		landed = appv1.Stopped
	}
	msg := "upgrade canceled; restored to " + landed.String()
	spec := domain.PatchSpec{
		DeleteAnnotations: []string{api.AppPreUpgradeConfigKey, api.AppTargetChartVersionKey},
	}
	if w.AM.Annotations != nil {
		if config, ok := w.AM.Annotations[api.AppPreUpgradeConfigKey]; ok {
			spec.Config = config
		}
	}
	return domain.Result{
		Next: landed,
		Effects: []domain.Effect{
			spec,
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   landed,
				Reason:  reason,
				Message: msg,
				Payload: map[string]string{},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

func upgradeDownloadingCancelingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadUpgradeDownloadingCancelingPhase]
}
