package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

// Package-local mirrors of the Upgrading-state messages.
const (
	upgradingMsgPrepareStarted     = "upgrading.prepare: version guard + chart fetch in progress"
	upgradingMsgPullStarted        = "upgrading.pull: pulling upgraded image refs"
	upgradingMsgHelmStarted        = "upgrading.helm: helm upgrade in progress"
	upgradingMsgWaitStartUpStarted = "upgrading.wait-startup: waiting for pods to become ready"
	upgradingReasonEntered         = "UpgradeStarted"
	upgradingReasonPullStarted     = "UpgradePullStarted"
	upgradingReasonHelmStarted     = "UpgradeHelmStarted"
	upgradingReasonWaitStartUp     = "UpgradeWaitStartUpStarted"
	upgradingReasonHandoffLaunch   = "UpgradeHelmDone"
	upgradingMsgHandoffLaunch      = "helm upgrade done; advancing to Initializing for launch wait"
	upgradingMsgHandoffStoppedFast = "helm upgrade at replicas=0 done; landing back in Stopped"
)

// registerUpgrading wires the Upgrading-state handlers into the
// TransitionTable. Structurally similar to Uninstalling: a single
// helm-driven phase that ends either in success (with a two-way
// success branch — Initializing OR Stopped fast-path) or failure.
//
// Wiring semantics:
//
//   - (Upgrading, Tick{}) — fresh edge. The default table declares
//     no Tick edge for Upgrading (it's a workload-callback state
//     in legacy). SetOrAddHandler inserts it.
//
//   - (Upgrading, WorkloadReady{}) — a placeholder already targets
//     Initializing. SetOrAddHandler replaces it; the real handler
//     branches on the AnnotPreUpgradeState annotation to route
//     into EITHER Initializing OR Stopped.
//
//   - (Upgrading, WorkloadFailed{}) — a placeholder already targets
//     UpgradeFailed. SetOrAddHandler replaces it.
//
// AddReachable(Upgrading, Upgrading) declares the phase-progression
// self-loop.
//
// Cancel path — (Upgrading, TriggerOp{CancelOp}) → UpgradingCanceling
// — is NOT ported here; it belongs with the broader *Canceling
// handler batch.
func registerUpgrading(t *domain.TransitionTable) {
	t.AddReachable(appv1.Upgrading, appv1.Upgrading)
	t.SetOrAddHandler(appv1.Upgrading, domain.Tick{}, OnUpgradingTick)
	t.SetOrAddHandler(appv1.Upgrading, domain.WorkloadReady{}, OnUpgradingWorkloadReady)
	t.SetOrAddHandler(appv1.Upgrading, domain.WorkloadFailed{}, OnUpgradingWorkloadFailed)
}

// OnUpgradingTick is the Tick handler that starts the upgrade
// pipeline on fresh entry into Upgrading.
//
// Legacy correspondence: pkg/appstate/upgrading_app.go:154-320 —
// UpgradingApp.exec running:
//  1. GetDeployedReleaseVersion + version guard (>= deployed)
//  2. Get admin + GetAppConfig / GetIndexAndDownloadChart (chart download)
//  3. ResolveImageRefs + imageClient.Create + PollDownloadProgress
//     (image pull)
//  4. NewHelmOps(SkipWaitForStartUp based on preState)
//  5. ops.Upgrade()
//  6. If preState=Stopped → return; else ops.WaitForStartUp
//
// Each legacy step is its own CallWorkload so the Runner can reuse
// the original helpers (GetDeployedReleaseVersion, GetAppConfig,
// ImagePuller, HelmOps.Upgrade, WaitForStartUp) rather than
// reimplementing them inside a single adapter method.
//
//	entry → prepare → pull → helm ─┬─ preState=Stopped → Stopped
//	                               └─ otherwise → wait-startup → Initializing
func OnUpgradingTick(_ context.Context, w domain.World) domain.Result {
	phase := upgradingPhase(w)

	switch phase {
	case "":
		return startUpgradePrepare(w)

	case PhaseUpgradingPrepare:
		return recoverWorkload(w, prepareUpgradeWorkload())
	case PhaseUpgradingPull:
		return recoverWorkload(w, upgradePullImageWorkload())
	case PhaseUpgradingHelm:
		return recoverWorkload(w, upgradeWorkload())
	case PhaseUpgradingWaitStartUp:
		return recoverWorkload(w, upgradeWaitStartUpWorkload())

	default:
		return domain.Result{
			Err: fmt.Errorf("upgrading: unknown %s=%q; expected one of %v",
				PayloadUpgradingPhase, phase, upgradingPhases),
		}
	}
}

var upgradingPhases = []string{
	PhaseUpgradingPrepare,
	PhaseUpgradingPull,
	PhaseUpgradingHelm,
	PhaseUpgradingWaitStartUp,
}

// OnUpgradingWorkloadReady completes the single-phase pipeline.
// Branches on the pre-upgrade-state annotation:
//
//   - If AnnotPreUpgradeState == "stopped" → land in Stopped
//     (fast-path — helm rendered replicas=0, no launch wait
//     needed; matches legacy upgrading_app.go:299-307).
//   - Otherwise → hand off to Initializing which owns the launch
//     wait via WorkloadWaitLaunch.
//
// Both success branches commit AppTargetChartVersionKey onto
// AppVersionKey and drop the upgrade scratch annotations. The
// Initializing path writes no OpRecord — Initializing→Running
// captures completion. The Stopped fast-path writes one, because
// that app never passes through Running and the ledger would
// otherwise keep reporting the pre-upgrade version.
func OnUpgradingWorkloadReady(_ context.Context, w domain.World) domain.Result {
	phase := upgradingPhase(w)

	switch phase {
	case PhaseUpgradingPrepare:
		return startUpgradePull(w)

	case PhaseUpgradingPull:
		return startHelmUpgrade(w)

	case PhaseUpgradingHelm:
		if upgradeFromStopped(w) {
			return handoffToStoppedFastPath(w)
		}
		return startUpgradeWaitStartUp(w)

	case PhaseUpgradingWaitStartUp:
		return handoffToInitializingAfterUpgrade(w)

	case "":
		return domain.Result{
			Err: fmt.Errorf("upgrading: WorkloadReady with empty %s; every CallWorkload emitted from Upgrading must stamp the phase first",
				PayloadUpgradingPhase),
		}

	default:
		return domain.Result{
			Err: fmt.Errorf("upgrading: WorkloadReady with unknown %s=%q",
				PayloadUpgradingPhase, phase),
		}
	}
}

// OnUpgradingWorkloadFailed handles a failed helm upgrade (or an
// upstream version-guard / image-pull failure bubbled up by the
// adapter). Routes to UpgradeFailed with an OpRecord + Warning
// event carrying the workload's failure data.
//
// Legacy correspondence: upgrading_app.go:105-122 — the
// defer-recover branch that maps errPodPending / errHamiUnschedulable
// to reason strings AppUnschedulable / AppHamiSchedulable before
// writing UpgradeFailed. That mapping is expected to live in the
// adapter (via WorkloadResult.LandedReason) — the handler just
// forwards whatever LandedReason the workload emitted.
func OnUpgradingWorkloadFailed(_ context.Context, w domain.World) domain.Result {
	msg, reason := failureFromAction(w.Action)
	if msg == "" {
		msg = "upgrading failed"
	}
	if reason == "" {
		reason = string(appv1.UpgradeFailed)
	}

	now := metav1.NewTime(w.Now)
	// See installing.go's failure branch for the note on why
	// OpRecord.Version is left empty at the handler layer.
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.OperationFailedTpl, w.AM.Status.OpType, msg),
		Source:    w.AM.Spec.Source,
		Status:    appv1.UpgradeFailed,
		StateTime: &now,
	}

	// Effect ordering: AppendOpRecord precedes PatchAM so both
	// land atomically. See installing.go for full RV-conflict
	// note.
	return domain.Result{
		Next: appv1.UpgradeFailed,
		Effects: []domain.Effect{
			domain.AppendOpRecord{Record: opRecord},
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.UpgradeFailed,
				Message: msg,
				Reason:  reason,
				Payload: map[string]string{}, // clear phase; leaving Upgrading
			}},
			domain.EmitEvent{
				Type:    "Warning",
				Reason:  reason,
				Message: msg,
			},
		},
	}
}

func startUpgradePrepare(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Upgrading,
				Message: upgradingMsgPrepareStarted,
				Reason:  upgradingReasonEntered,
				Payload: map[string]string{PayloadUpgradingPhase: PhaseUpgradingPrepare},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradingReasonEntered,
				Message: upgradingMsgPrepareStarted,
			},
			prepareUpgradeWorkload(),
		},
	}
}

func prepareUpgradeWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadPrepareUpgrade,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "UpgradePrepareFailed"},
	}
}

func startUpgradePull(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Upgrading,
				Message: upgradingMsgPullStarted,
				Reason:  upgradingReasonPullStarted,
				Payload: map[string]string{PayloadUpgradingPhase: PhaseUpgradingPull},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradingReasonPullStarted,
				Message: upgradingMsgPullStarted,
			},
			upgradePullImageWorkload(),
		},
	}
}

func upgradePullImageWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadPullImage,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "UpgradeImagePullFailed"},
	}
}

func startHelmUpgrade(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Upgrading,
				Message: upgradingMsgHelmStarted,
				Reason:  upgradingReasonHelmStarted,
				Payload: map[string]string{PayloadUpgradingPhase: PhaseUpgradingHelm},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradingReasonHelmStarted,
				Message: upgradingMsgHelmStarted,
			},
			upgradeWorkload(),
		},
	}
}

func upgradeWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadUpgrade,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: "HelmUpgradeFailed"},
	}
}

func startUpgradeWaitStartUp(_ domain.World) domain.Result {
	return domain.Result{
		Next: "",
		Effects: []domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Upgrading,
				Message: upgradingMsgWaitStartUpStarted,
				Reason:  upgradingReasonWaitStartUp,
				Payload: map[string]string{PayloadUpgradingPhase: PhaseUpgradingWaitStartUp},
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradingReasonWaitStartUp,
				Message: upgradingMsgWaitStartUpStarted,
			},
			upgradeWaitStartUpWorkload(),
		},
	}
}

func upgradeWaitStartUpWorkload() domain.CallWorkload {
	return domain.CallWorkload{
		Kind:      domain.WorkloadWaitStartUp,
		OnSuccess: domain.WorkloadReady{},
		OnFailure: domain.WorkloadFailed{Reason: constants.AppStopDueToStartUpFailed},
	}
}

// handoffToInitializingAfterUpgrade is the normal-path success
// branch — same shape as resuming.handoffToInitializing but with
// upgrade-specific messages so operator dashboards can tell the
// two apart at the event-log level.
func handoffToInitializingAfterUpgrade(w domain.World) domain.Result {
	spec, _, err := commitUpgradeMetadata(w)
	if err != nil {
		return domain.Result{Err: err}
	}
	return domain.Result{
		Next: appv1.Initializing,
		Effects: appendCommit([]domain.Effect{
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  upgradingReasonHandoffLaunch,
				Message: upgradingMsgHandoffLaunch,
			},
		}, spec, domain.PatchAM{Status: appv1.ApplicationManagerStatus{
			State:   appv1.Initializing,
			Message: upgradingMsgHandoffLaunch,
			Reason:  upgradingReasonHandoffLaunch,
			Payload: map[string]string{}, // clear phase; leaving Upgrading
		}}),
	}
}

// handoffToStoppedFastPath is the upgrade-from-Stopped success
// branch. Land the app back in Stopped with reason
// AppStopByUser, and record the committed chart version. This
// path never reaches Running, so the op ledger is the only place
// the new version is visible while the app stays Stopped.
func handoffToStoppedFastPath(w domain.World) domain.Result {
	spec, target, err := commitUpgradeMetadata(w)
	if err != nil {
		return domain.Result{Err: err}
	}
	now := metav1.NewTime(w.Now)
	opRecord := appv1.OpRecord{
		OpID:      w.AM.Status.OpID,
		OpType:    w.AM.Status.OpType,
		Message:   fmt.Sprintf(constants.UpgradeOperationCompletedTpl, w.AM.Spec.Type.String(), w.AM.Spec.AppName),
		Source:    w.AM.Spec.Source,
		Version:   target,
		Status:    appv1.Stopped,
		StateTime: &now,
	}
	return domain.Result{
		Next: appv1.Stopped,
		Effects: appendCommit([]domain.Effect{
			domain.PatchAM{Status: appv1.ApplicationManagerStatus{
				State:   appv1.Stopped,
				Message: upgradingMsgHandoffStoppedFast,
				Reason:  constants.AppStopByUser,
				Payload: map[string]string{}, // clear phase
			}},
			domain.EmitEvent{
				Type:    "Normal",
				Reason:  constants.AppStopByUser,
				Message: upgradingMsgHandoffStoppedFast,
			},
		}, spec, domain.AppendOpRecord{Record: opRecord}),
	}
}

// commitUpgradeMetadata copies the requested chart version onto the
// deployed-version annotation and drops the scratch keys. AppVersionKey
// stays unchanged until this point, so a canceled upgrade still reports
// the release that is actually installed.
//
// A completed attempt that already wrote the annotation and then failed
// before the status transition landed retries with the scratch keys gone.
// That retry must still hand off: chart-version already matches Spec.Config,
// so there is nothing left to write.
func commitUpgradeMetadata(w domain.World) (domain.PatchSpec, string, error) {
	if w.AM == nil || w.AM.Annotations == nil {
		return domain.PatchSpec{}, "", fmt.Errorf("upgrading: completed upgrade is missing target chart version")
	}
	target := w.AM.Annotations[api.AppTargetChartVersionKey]
	if target == "" {
		committed := w.AM.Annotations[api.AppVersionKey]
		if committed != "" && committed == upgradeSpecConfigVersion(w) {
			return domain.PatchSpec{}, committed, nil
		}
		return domain.PatchSpec{}, "", fmt.Errorf("target chart version missing for completed upgrade of %s", w.AM.Name)
	}
	return domain.PatchSpec{
		SetAnnotations: map[string]string{
			api.AppVersionKey: target,
		},
		DeleteAnnotations: []string{
			api.AppTargetChartVersionKey,
			api.AppPreUpgradeConfigKey,
		},
	}, target, nil
}

// appendCommit puts a metadata patch ahead of the status effects, and
// skips it when the commit already happened on a previous attempt.
func appendCommit(rest []domain.Effect, spec domain.PatchSpec, first domain.Effect) []domain.Effect {
	out := make([]domain.Effect, 0, len(rest)+2)
	if spec.Config != "" || spec.OpType != "" || len(spec.SetAnnotations) > 0 || len(spec.DeleteAnnotations) > 0 {
		out = append(out, spec)
	}
	out = append(out, first)
	return append(out, rest...)
}

// upgradeSpecConfigVersion reads the chart version stored in Spec.Config.
// Live configs use "Version"; older fixtures use "version".
func upgradeSpecConfigVersion(w domain.World) string {
	if w.AM == nil || w.AM.Spec.Config == "" {
		return ""
	}
	var cfg struct {
		Version  string `json:"Version"`
		Version2 string `json:"version"`
	}
	if err := json.Unmarshal([]byte(w.AM.Spec.Config), &cfg); err != nil {
		return ""
	}
	if cfg.Version != "" {
		return cfg.Version
	}
	return cfg.Version2
}

// upgradeFromStopped reports whether the upgrade was requested
// from a Stopped source (so its success branch should skip the
// launch wait). Reads AnnotPreUpgradeState with the standard
// nil-map guard.
//
// The comparison is case-insensitive against the state string
// because the legacy annotation writer uses ApplicationManagerState.
// String() which produces "stopped" — but a future writer might
// use metadata.serializer values with different casing.
func upgradeFromStopped(w domain.World) bool {
	if w.AM == nil || w.AM.Annotations == nil {
		return false
	}
	v := w.AM.Annotations[AnnotPreUpgradeState]
	return strings.EqualFold(v, appv1.Stopped.String())
}

// upgradingPhase reads Payload[PayloadUpgradingPhase] with the
// standard nil-map guard.
func upgradingPhase(w domain.World) string {
	if w.AM == nil || w.AM.Status.Payload == nil {
		return ""
	}
	return w.AM.Status.Payload[PayloadUpgradingPhase]
}
