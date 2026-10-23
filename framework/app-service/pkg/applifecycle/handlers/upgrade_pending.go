package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

const (
	upgradePendingMsgStartDownload = "start to download"
	upgradePendingMsgBedFull       = "waiting: another app is currently downloading"
	upgradePendingReasonAdvance    = "UpgradeDownloading"
	upgradePendingReasonWaiting    = "WaitingForDownloadingSlot"
)

// registerUpgradePending wires the queue in front of an upgrade image
// pull. UpgradePending shares the cluster-wide download slot with
// install Pending: only the FIFO head advances, and only when no AM
// is already Downloading or UpgradeDownloading.
func registerUpgradePending(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradePending, appv1.UpgradePending)
	t.SetOrAddHandler(appv1.UpgradePending, domain.Tick{}, OnUpgradePendingTick)
}

// OnUpgradePendingTick advances UpgradePending to UpgradeDownloading
// when this AM is the oldest waiter and the download slot is free.
func OnUpgradePendingTick(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("upgrade-pending: Tick invoked with nil AM")}
	}
	if w.ClusterStats.DownloadingCount >= pendingDownloadingSlotSize {
		return waitForUpgradeSlot(w)
	}
	if head := w.ClusterStats.NextDownloading; head != "" && w.AM.Name != head {
		return waitForUpgradeSlot(w)
	}
	return advanceToUpgradeDownloading(w)
}

func advanceToUpgradeDownloading(w domain.World) domain.Result {
	return domain.Result{
		Next: appv1.UpgradeDownloading,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeDownloading,
				Reason:  upgradePendingReasonAdvance,
				Message: upgradePendingMsgStartDownload,
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradePendingReasonAdvance,
				Message: upgradePendingMsgStartDownload,
			},
		},
	}
}

func waitForUpgradeSlot(w domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradePending,
				Reason:  upgradePendingReasonWaiting,
				Message: upgradePendingMsgBedFull,
			}},
			domain.ScheduleTick{After: pendingBedFullBackoff},
		},
	}
}
