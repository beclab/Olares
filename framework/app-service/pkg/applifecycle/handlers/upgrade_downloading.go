package handlers

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

const (
	upgradeDownloadingMsgPrepare = "upgrade-downloading.prepare: version guard + chart fetch in progress"
	upgradeDownloadingMsgPull    = "upgrade-downloading.pull: pulling upgraded image refs"
	upgradeDownloadingMsgHandoff = "images ready; starting helm upgrade"
	upgradeDownloadingReasonPrep = "UpgradeDownloadPrepareStarted"
	upgradeDownloadingReasonPull = "UpgradeDownloadPullStarted"
	upgradeDownloadingReasonDone = "UpgradeDownloadCompleted"
)

// registerUpgradeDownloading wires chart refresh and image pull.
// Success hands off to Upgrading already in the helm phase, so helm
// does not pull images again. Failure lands in UpgradeFailed.
func registerUpgradeDownloading(t *domain.TransitionTable) {
	t.AddReachable(appv1.UpgradeDownloading, appv1.UpgradeDownloading)
	t.SetOrAddHandler(appv1.UpgradeDownloading, domain.Tick{}, OnUpgradeDownloadingTick)
	t.SetOrAddHandler(appv1.UpgradeDownloading, domain.WorkloadReady{}, OnUpgradeDownloadingWorkloadReady)
	t.SetOrAddHandler(appv1.UpgradeDownloading, domain.WorkloadFailed{}, OnUpgradeDownloadingWorkloadFailed)
}

// OnUpgradeDownloadingTick drives prepare then pull. An empty phase
// starts chart refresh; a persisted phase re-dispatches only after
// restart, when no in-process claim survived.
func OnUpgradeDownloadingTick(_ context.Context, w domain.World) domain.Result {
	switch upgradeDownloadingPhase(w) {
	case "":
		return startUnlessSameKind(w, domain.WorkloadPrepareUpgrade, startUpgradeDownloadPrepare(w))
	case PhaseUpgradingPrepare:
		return recoverWorkload(w, prepareUpgradeWorkload())
	case PhaseUpgradingPull:
		return recoverWorkload(w, upgradePullImageWorkload())
	default:
		return domain.Result{
			Err: fmt.Errorf("upgrade-downloading: unknown %s=%q",
				PayloadUpgradeDownloadingPhase, upgradeDownloadingPhase(w)),
		}
	}
}

// OnUpgradeDownloadingWorkloadReady advances prepare to pull, and
// pull to Upgrading with the helm phase already selected.
func OnUpgradeDownloadingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	switch upgradeDownloadingPhase(w) {
	case PhaseUpgradingPrepare:
		return startUpgradeDownloadPull(w)
	case PhaseUpgradingPull:
		return handoffUpgradeDownloadingToUpgrading(w)
	case "":
		return domain.Result{
			Err: fmt.Errorf("upgrade-downloading: WorkloadReady with empty %s",
				PayloadUpgradeDownloadingPhase),
		}
	default:
		return domain.Result{
			Err: fmt.Errorf("upgrade-downloading: WorkloadReady with unknown %s=%q",
				PayloadUpgradeDownloadingPhase, upgradeDownloadingPhase(w)),
		}
	}
}

// OnUpgradeDownloadingWorkloadFailed records the failed upgrade and
// leaves UpgradeDownloading. Chart-fetch and image-pull failures
// share this landing state.
func OnUpgradeDownloadingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	if w.AM == nil {
		return domain.Result{Err: fmt.Errorf("upgrade-downloading: WorkloadFailed invoked with nil AM")}
	}
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "upgrade download failed"
	}
	if reason == "" {
		reason = string(appv1.UpgradeFailed)
	}
	now := metav1.NewTime(w.Now)
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.UpgradeFailed,
		StateTime: &now,
	}
	return domain.Result{
		Next: appv1.UpgradeFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeFailed,
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

func startUpgradeDownloadPrepare(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeDownloading,
				Message: upgradeDownloadingMsgPrepare,
				Reason:  upgradeDownloadingReasonPrep,
				Payload: map[string]string{PayloadUpgradeDownloadingPhase: PhaseUpgradingPrepare},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradeDownloadingReasonPrep,
				Message: upgradeDownloadingMsgPrepare,
			},
			prepareUpgradeWorkload(),
		},
	}
}

func startUpgradeDownloadPull(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeDownloading,
				Message: upgradeDownloadingMsgPull,
				Reason:  upgradeDownloadingReasonPull,
				Payload: map[string]string{PayloadUpgradeDownloadingPhase: PhaseUpgradingPull},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradeDownloadingReasonPull,
				Message: upgradeDownloadingMsgPull,
			},
			upgradePullImageWorkload(),
		},
	}
}

func handoffUpgradeDownloadingToUpgrading(_ domain.World) domain.Result {
	return domain.Result{
		Next: appv1.Upgrading,
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Upgrading,
				Message: upgradeDownloadingMsgHandoff,
				Reason:  upgradeDownloadingReasonDone,
				Payload: map[string]string{PayloadUpgradingPhase: PhaseUpgradingHelm},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradeDownloadingReasonDone,
				Message: upgradeDownloadingMsgHandoff,
			},
			upgradeWorkload(),
		},
	}
}

func upgradeDownloadingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadUpgradeDownloadingPhase]
}
