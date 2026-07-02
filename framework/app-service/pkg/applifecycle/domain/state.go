// Package domain is the pure core of the application-lifecycle state machine.
//
// This package MUST NOT import any of the following, directly or transitively:
//
//   - sigs.k8s.io/controller-runtime/pkg/client
//   - k8s.io/client-go (except the CRD type re-exports via appsv1)
//   - helm.sh/helm/v3
//   - github.com/nats-io/nats.go
//
// The only Kubernetes-adjacent import allowed is the CRD type package
// (github.com/beclab/api/api/app.bytetrade.io/v1alpha1) because
// ApplicationManager / ApplicationManagerStatus / OpRecord are our domain
// aggregate root and are needed to express PatchAM effects. See CONTRACT.md
// for the full list of handler rules.
package domain

import (
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// State is a type alias for the CRD state enum. Kept as an alias so
// handlers and adapters can pass values freely between domain and CRD
// boundaries without conversion.
type State = appsv1.ApplicationManagerState

// OpType is a type alias for the CRD operation enum. Same rationale as
// State.
type OpType = appsv1.OpType

// InitialState represents an ApplicationManager that has just been created
// and has not been written by the state machine yet. handler_installer_install.go
// writes Pending against this sentinel via apputils.UpdateAppMgrStatus.
const InitialState State = ""

// AllStates enumerates every declared state, plus the empty sentinel.
var AllStates = []State{
	InitialState,

	appsv1.Pending,
	appsv1.Downloading,
	appsv1.UpgradePending,
	appsv1.UpgradeDownloading,
	appsv1.Installing,
	appsv1.Initializing,
	appsv1.Running,
	appsv1.Resuming,
	appsv1.Upgrading,
	appsv1.ApplyingEnv,
	appsv1.Stopping,
	appsv1.Uninstalling,

	appsv1.PendingCanceling,
	appsv1.DownloadingCanceling,
	appsv1.UpgradePendingCanceling,
	appsv1.UpgradeDownloadingCanceling,
	appsv1.InstallingCanceling,
	appsv1.InitializingCanceling,
	appsv1.ResumingCanceling,
	appsv1.UpgradingCanceling,
	appsv1.ApplyingEnvCanceling,

	appsv1.PendingCancelFailed,
	appsv1.DownloadingCancelFailed,
	appsv1.UpgradePendingCancelFailed,
	appsv1.UpgradeDownloadingCancelFailed,
	appsv1.InstallingCancelFailed,
	appsv1.UpgradingCancelFailed,
	appsv1.ApplyingEnvCancelFailed,
	appsv1.ResumingCancelFailed,

	appsv1.PendingCanceled,
	appsv1.DownloadingCanceled,
	appsv1.InstallingCanceled,

	appsv1.Stopped,
	appsv1.Uninstalled,

	appsv1.DownloadFailed,
	appsv1.InstallFailed,
	appsv1.StopFailed,
	appsv1.UpgradeFailed,
	appsv1.ApplyEnvFailed,
	appsv1.ResumeFailed,
	appsv1.UninstallFailed,
}

// AllOps enumerates every declared operation type.
var AllOps = []OpType{
	appsv1.InstallOp,
	appsv1.UninstallOp,
	appsv1.UpgradeOp,
	appsv1.StopOp,
	appsv1.ResumeOp,
	appsv1.CancelOp,
	appsv1.ApplyEnvOp,
}

// defaultTTL is the fallback state timeout when a state does not appear in
// the TransitionTable's TTL map. Matches pkg/appstate.StateToDuration's
// default of 10 minutes.
const defaultTTL = 10 * time.Minute
