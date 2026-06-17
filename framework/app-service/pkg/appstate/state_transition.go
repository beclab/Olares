package appstate

import (
	"fmt"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

var All = []appv1alpha1.ApplicationManagerState{
	appv1alpha1.Pending,
	appv1alpha1.Downloading,
	appv1alpha1.Installing,
	appv1alpha1.Initializing,
	appv1alpha1.Running,
	appv1alpha1.Resuming,
	appv1alpha1.Upgrading,
	appv1alpha1.ApplyingEnv,
	appv1alpha1.Stopping,
	appv1alpha1.Uninstalling,

	appv1alpha1.PendingCanceling,
	appv1alpha1.DownloadingCanceling,
	appv1alpha1.InstallingCanceling,
	appv1alpha1.InitializingCanceling,
	appv1alpha1.ResumingCanceling,
	appv1alpha1.UpgradingCanceling,
	appv1alpha1.ApplyingEnvCanceling,

	appv1alpha1.PendingCancelFailed,
	appv1alpha1.DownloadingCancelFailed,
	appv1alpha1.InstallingCancelFailed,
	appv1alpha1.UpgradingCancelFailed,
	appv1alpha1.ApplyingEnvCancelFailed,
	appv1alpha1.ResumingCancelFailed,

	appv1alpha1.PendingCanceled,
	appv1alpha1.DownloadingCanceled,
	appv1alpha1.InstallingCanceled,

	appv1alpha1.Stopped,
	appv1alpha1.Uninstalled,

	appv1alpha1.DownloadFailed,
	appv1alpha1.InstallFailed,
	appv1alpha1.StopFailed,
	appv1alpha1.UpgradeFailed,
	appv1alpha1.ApplyEnvFailed,
	appv1alpha1.ResumeFailed,
	appv1alpha1.UninstallFailed,
}

// StateTransitions enumerates every (from -> to) edge the runtime can produce.
// It is consumed by IsStateTransitionValid and by updateStatus' transition guard
// in pkg/appstate/types.go, so any path that calls updateStatus(..., newState)
// MUST declare its (currentState -> newState) edge here or the patch will be
// rejected at runtime.
//
// When adding a new edge, also note its source so the next reader can audit
// the table without re-grepping the codebase.
var StateTransitions = map[appv1alpha1.ApplicationManagerState][]appv1alpha1.ApplicationManagerState{
	// "" represents an AM that has just been created and has not been written
	// by the state machine yet; handler_installer_install.go writes Pending
	// against this sentinel via apputils.UpdateAppMgrStatus.
	"": {
		appv1alpha1.Pending,
	},
	appv1alpha1.Pending: {
		appv1alpha1.Downloading,
		appv1alpha1.PendingCanceling,
	},
	appv1alpha1.Downloading: {
		appv1alpha1.Installing,
		appv1alpha1.DownloadFailed,
		appv1alpha1.DownloadingCanceling,
	},
	appv1alpha1.Installing: {
		appv1alpha1.Initializing,
		appv1alpha1.InstallFailed,
		appv1alpha1.InstallingCanceling,
		appv1alpha1.Stopping,
		// installing_app.go: middleware fast-path lands directly in Running
		// after WaitForLaunch instead of going through Initializing.
		appv1alpha1.Running,
	},
	appv1alpha1.Initializing: {
		appv1alpha1.Running,
		appv1alpha1.InitializingCanceling,
	},
	appv1alpha1.Running: {
		appv1alpha1.Stopping,
		appv1alpha1.Upgrading,
		appv1alpha1.ApplyingEnv,
		appv1alpha1.Uninstalling,
		// running_app.go -> forceDeleteApp -> Uninstalled when the
		// Application CR has disappeared (self-heal path).
		appv1alpha1.Uninstalled,
	},
	appv1alpha1.Stopping: {
		appv1alpha1.Stopped,
		appv1alpha1.StopFailed,
	},
	appv1alpha1.Upgrading: {
		appv1alpha1.Initializing,
		appv1alpha1.UpgradeFailed,
		appv1alpha1.UpgradingCanceling,
		// upgrading_app.go: upgrade-from-Stopped lands back in Stopped via
		// landedState (helm release upgraded at replicas=0, nothing to wait
		// for, no need to go through Initializing).
		appv1alpha1.Stopped,
	},
	appv1alpha1.ApplyingEnv: {
		appv1alpha1.Initializing,
		appv1alpha1.ApplyEnvFailed,
		appv1alpha1.ApplyingEnvCanceling,
	},
	appv1alpha1.Uninstalling: {
		appv1alpha1.Uninstalled,
		appv1alpha1.UninstallFailed,
	},
	appv1alpha1.PendingCanceling: {
		appv1alpha1.PendingCanceled,
		appv1alpha1.PendingCancelFailed,
	},
	appv1alpha1.DownloadingCanceling: {
		appv1alpha1.DownloadingCanceled,
		appv1alpha1.DownloadingCancelFailed,
	},
	appv1alpha1.InstallingCanceling: {
		appv1alpha1.InstallingCanceled,
		appv1alpha1.InstallingCancelFailed,
	},

	// initializing state cancel directly turn to stopping
	appv1alpha1.InitializingCanceling: {
		appv1alpha1.Stopping,
		// initializing_canceling_app.go.Cancel writes InstallingCancelFailed
		// (the namespace is reused with installing-cancel for failure paths).
		appv1alpha1.InstallingCancelFailed,
	},

	appv1alpha1.Resuming: {
		appv1alpha1.ResumingCanceling,
		appv1alpha1.ResumeFailed,
		appv1alpha1.Initializing,
		// OperationAllowedInState[Resuming][StopOp]=true, so handler_suspend
		// / pod_abnormal_suspend_app_controller can flip Resuming -> Stopping.
		appv1alpha1.Stopping,
	},
	appv1alpha1.ResumingCanceling: {
		appv1alpha1.Stopping,
		// resuming_canceling_app.go.Cancel writes ResumingCancelFailed.
		appv1alpha1.ResumingCancelFailed,
	},
	appv1alpha1.UpgradingCanceling: {
		appv1alpha1.Stopping,
		appv1alpha1.UpgradingCancelFailed,
	},
	appv1alpha1.ApplyingEnvCanceling: {
		appv1alpha1.Stopping,
		appv1alpha1.ApplyingEnvCancelFailed,
	},
	appv1alpha1.Stopped: {
		appv1alpha1.Resuming,
		appv1alpha1.Uninstalling,
		appv1alpha1.Upgrading,
		appv1alpha1.ApplyingEnv,
	},

	appv1alpha1.DownloadFailed: {
		appv1alpha1.Pending,
	},
	appv1alpha1.InstallFailed: {
		appv1alpha1.Pending,
	},

	appv1alpha1.StopFailed: {
		appv1alpha1.Stopping,
		appv1alpha1.Upgrading,
		appv1alpha1.ApplyingEnv,
		appv1alpha1.Uninstalling,
	},

	appv1alpha1.UpgradeFailed: {
		appv1alpha1.Stopping,
		appv1alpha1.Upgrading,
		appv1alpha1.Uninstalling,
	},
	appv1alpha1.ApplyEnvFailed: {
		appv1alpha1.Stopping,
		appv1alpha1.ApplyingEnv,
		appv1alpha1.Uninstalling,
	},
	appv1alpha1.ResumeFailed: {
		appv1alpha1.Resuming,
		appv1alpha1.Uninstalling,
		// OperationAllowedInState[ResumeFailed] also allows UpgradeOp and
		// ApplyEnvOp; handler_installer_upgrade.go and handler_applyenv.go
		// drive these edges.
		appv1alpha1.Upgrading,
		appv1alpha1.ApplyingEnv,
	},
	appv1alpha1.UninstallFailed: {
		appv1alpha1.Uninstalling,
		// uninstall_failed_app.go -> forceDeleteApp -> Uninstalled.
		appv1alpha1.Uninstalled,
	},
	appv1alpha1.PendingCancelFailed: {
		appv1alpha1.PendingCanceling,
	},
	appv1alpha1.DownloadingCancelFailed: {
		appv1alpha1.DownloadingCanceling,
	},
	appv1alpha1.InstallingCancelFailed: {
		appv1alpha1.InstallingCanceling,
		// OperationAllowedInState[InstallingCancelFailed][UninstallOp]=true.
		appv1alpha1.Uninstalling,
	},
	// UpgradingCancelFailed / ApplyingEnvCancelFailed were missing entirely.
	// Both allow CancelOp (re-enter the canceling state) and UninstallOp.
	appv1alpha1.UpgradingCancelFailed: {
		appv1alpha1.UpgradingCanceling,
		appv1alpha1.Uninstalling,
	},
	appv1alpha1.ApplyingEnvCancelFailed: {
		appv1alpha1.ApplyingEnvCanceling,
		appv1alpha1.Uninstalling,
	},

	// Terminal-reinstallable states: handler_installer_install.go re-targets
	// the same AM by patching Spec and writing Status.State=Pending when the
	// caller still holds the same (app, owner) tuple. The conflict-check
	// only lazy-deletes cross-type (shared vs per-user) collisions, so the
	// same-type same-owner reinstall flips state in place.
	appv1alpha1.PendingCanceled: {
		appv1alpha1.Pending,
	},
	appv1alpha1.DownloadingCanceled: {
		appv1alpha1.Pending,
	},
	appv1alpha1.InstallingCanceled: {
		appv1alpha1.Pending,
	},
	appv1alpha1.Uninstalled: {
		appv1alpha1.Pending,
	},
}

var OperationAllowedInState = map[appv1alpha1.ApplicationManagerState]map[appv1alpha1.OpType]bool{
	// application manager does not exist
	"": {
		appv1alpha1.InstallOp: true,
	},
	appv1alpha1.Pending: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.Downloading: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.Installing: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.Initializing: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.Upgrading: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.ApplyingEnv: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.Resuming: {
		appv1alpha1.CancelOp: true,
		appv1alpha1.StopOp:   true,
	},
	appv1alpha1.Uninstalling:          {},
	appv1alpha1.PendingCanceling:      {},
	appv1alpha1.DownloadingCanceling:  {},
	appv1alpha1.InstallingCanceling:   {},
	appv1alpha1.InitializingCanceling: {},
	appv1alpha1.UpgradingCanceling:    {},
	appv1alpha1.ApplyingEnvCanceling:  {},
	appv1alpha1.ResumingCanceling:     {},

	appv1alpha1.PendingCanceled: {
		appv1alpha1.InstallOp: true,
	},
	appv1alpha1.DownloadingCanceled: {
		appv1alpha1.InstallOp: true,
	},
	appv1alpha1.InstallingCanceled: {
		appv1alpha1.InstallOp: true,
	},
	appv1alpha1.Uninstalled: {
		appv1alpha1.InstallOp: true,
	},
	//appv1alpha1.InitializingCanceled: {
	//	appv1alpha1.UpgradeOp:   true,
	//	appv1alpha1.UninstallOp: true,
	//	appv1alpha1.ResumeOp:    true,
	//},
	//appv1alpha1.UpgradingCanceled: {
	//	appv1alpha1.UpgradeOp:   true,
	//	appv1alpha1.UninstallOp: true,
	//	appv1alpha1.ResumeOp:    true,
	//},
	//appv1alpha1.ResumingCanceled: {
	//	appv1alpha1.UpgradeOp:   true,
	//	appv1alpha1.UninstallOp: true,
	//	appv1alpha1.ResumeOp:    true,
	//},
	appv1alpha1.DownloadFailed: {
		appv1alpha1.InstallOp: true,
	},
	appv1alpha1.InstallFailed: {
		appv1alpha1.InstallOp: true,
	},
	//appv1alpha1.InitialFailed: {
	//	appv1alpha1.UpgradeOp:   true,
	//	appv1alpha1.UninstallOp: true,
	//	appv1alpha1.ResumeOp:    true,
	//},
	appv1alpha1.StopFailed: {
		appv1alpha1.StopOp:      true,
		appv1alpha1.UpgradeOp:   true,
		appv1alpha1.UninstallOp: true,
	},
	appv1alpha1.ResumeFailed: {
		appv1alpha1.ResumeOp:    true,
		appv1alpha1.UpgradeOp:   true,
		appv1alpha1.ApplyEnvOp:  true,
		appv1alpha1.UninstallOp: true,
	},
	appv1alpha1.UninstallFailed: {
		appv1alpha1.UninstallOp: true,
	},
	appv1alpha1.UpgradeFailed: {
		appv1alpha1.UninstallOp: true,
		appv1alpha1.UpgradeOp:   true,
	},
	appv1alpha1.ApplyEnvFailed: {
		appv1alpha1.UninstallOp: true,
		appv1alpha1.ApplyEnvOp:  true,
	},
	appv1alpha1.PendingCancelFailed: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.DownloadingCancelFailed: {
		appv1alpha1.CancelOp: true,
	},
	appv1alpha1.InstallingCancelFailed: {
		appv1alpha1.CancelOp:    true,
		appv1alpha1.UninstallOp: true,
	},

	appv1alpha1.UpgradingCancelFailed: {
		appv1alpha1.CancelOp:    true,
		appv1alpha1.UninstallOp: true,
	},
	appv1alpha1.ApplyingEnvCancelFailed: {
		appv1alpha1.CancelOp:    true,
		appv1alpha1.UninstallOp: true,
	},
	appv1alpha1.Running: {
		appv1alpha1.UninstallOp: true,
		appv1alpha1.UpgradeOp:   true,
		appv1alpha1.ApplyEnvOp:  true,
		appv1alpha1.StopOp:      true,
	},
	appv1alpha1.Stopped: {
		appv1alpha1.UninstallOp: true,
		appv1alpha1.UpgradeOp:   true,
		appv1alpha1.ApplyEnvOp:  true,
		appv1alpha1.ResumeOp:    true,
	},
}

var CancelableStates = map[appv1alpha1.ApplicationManagerState]bool{
	appv1alpha1.Pending:      true,
	appv1alpha1.Downloading:  true,
	appv1alpha1.Installing:   true,
	appv1alpha1.Initializing: true,
	appv1alpha1.Resuming:     true,
	appv1alpha1.Upgrading:    true,
	appv1alpha1.ApplyingEnv:  true,
}

var OperatingStates = map[appv1alpha1.ApplicationManagerState]bool{
	appv1alpha1.Pending:      true,
	appv1alpha1.Downloading:  true,
	appv1alpha1.Installing:   true,
	appv1alpha1.Initializing: true,
	appv1alpha1.Resuming:     true,
	appv1alpha1.Upgrading:    true,
	appv1alpha1.ApplyingEnv:  true,
	appv1alpha1.Stopping:     true,

	appv1alpha1.PendingCanceling:      true,
	appv1alpha1.DownloadingCanceling:  true,
	appv1alpha1.InstallingCanceling:   true,
	appv1alpha1.InitializingCanceling: true,
	appv1alpha1.ResumingCanceling:     true,
	appv1alpha1.UpgradingCanceling:    true,
	appv1alpha1.ApplyingEnvCanceling:  true,

	appv1alpha1.Uninstalling: true,
}

var CancelingStates = map[appv1alpha1.ApplicationManagerState]bool{
	appv1alpha1.PendingCanceling:      true,
	appv1alpha1.DownloadingCanceling:  true,
	appv1alpha1.InstallingCanceling:   true,
	appv1alpha1.InitializingCanceling: true,
	appv1alpha1.ResumingCanceling:     true,
	appv1alpha1.UpgradingCanceling:    true,
	appv1alpha1.ApplyingEnvCanceling:  true,
}

var StateToDurationMap = map[appv1alpha1.ApplicationManagerState]time.Duration{
	appv1alpha1.Pending:      24 * time.Hour,
	appv1alpha1.Downloading:  30 * 24 * time.Hour,
	appv1alpha1.Installing:   30 * time.Minute,
	appv1alpha1.Initializing: time.Hour,
	appv1alpha1.Upgrading:    time.Hour,
	appv1alpha1.ApplyingEnv:  30 * time.Minute,
}

func IsOperationAllowed(curState appv1alpha1.ApplicationManagerState, op appv1alpha1.OpType) bool {
	if allowedOps, exists := OperationAllowedInState[curState]; exists {
		return allowedOps[op]
	}
	return false
}

// ExplainOperationNotAllowed builds the error returned when op is rejected in
// curState. When the state has an in-progress operation that can be cancelled
// (CancelOp is allowed), the message guides the caller to cancel first instead
// of just reporting "not allowed", which otherwise leaves callers retrying the
// same blocked operation (e.g. uninstall stuck behind an initializing app).
func ExplainOperationNotAllowed(curState appv1alpha1.ApplicationManagerState, op appv1alpha1.OpType) error {
	if IsOperationAllowed(curState, appv1alpha1.CancelOp) {
		return fmt.Errorf("%s operation is not allowed while the app is in %s state: an operation is in progress; "+
			"cancel it first via POST /app-service/v1/apps/{name}/cancel and wait until the app reaches a terminal state, then retry %s",
			op, curState, op)
	}
	return fmt.Errorf("%s operation is not allowed for %s state", op, curState)
}

func IsCancelable(curState appv1alpha1.ApplicationManagerState) bool {
	return CancelableStates[curState]
}

func IsCanceling(curState appv1alpha1.ApplicationManagerState) bool {
	return CancelingStates[curState]
}

func IsStateTransitionValid(from, to appv1alpha1.ApplicationManagerState) bool {
	if validTransitions, exists := StateTransitions[from]; exists {
		for _, validState := range validTransitions {
			if validState == to {
				return true
			}
		}
	}
	return false
}

// IsStateTransitionAllowed is the runtime guard used by updateStatus. It is
// looser than IsStateTransitionValid in exactly one way: a same-state write
// (from == to) is always allowed, so that idempotent retries of updateStatus
// (e.g. RetryOnConflict re-reading the latest state, or a deferred Finally
// re-asserting a terminal state with an updated message/reason) are not
// rejected.
func IsStateTransitionAllowed(from, to appv1alpha1.ApplicationManagerState) bool {
	if from == to {
		return true
	}
	return IsStateTransitionValid(from, to)
}

func StateToDuration(state appv1alpha1.ApplicationManagerState) time.Duration {
	if t, ok := StateToDurationMap[state]; ok {
		return t
	}
	return 10 * time.Minute
}

// IsTerminalReinstallable reports whether an AM is in a terminal state from
// which an Install operation is allowed. This is the broadest set; it matches
// OperationAllowedInState[state][InstallOp] == true with a non-empty state.
//
// NOTE: "reinstallable" does not imply "safe to delete the AM CR". InstallFailed
// historically left dangling helm/permission/provider/shared-NS resources before
// the install-failure cleanup helper plugged the gaps; use IsSafelyDeletable
// when the intent is to remove the AM.
func IsTerminalReinstallable(state appv1alpha1.ApplicationManagerState) bool {
	switch state {
	case appv1alpha1.PendingCanceled,
		appv1alpha1.DownloadingCanceled,
		appv1alpha1.InstallingCanceled,
		appv1alpha1.Uninstalled,
		appv1alpha1.DownloadFailed,
		appv1alpha1.InstallFailed:
		return true
	}
	return false
}

// IsSafelyDeletable reports whether the AM CR can be deleted directly without
// leaking cluster resources. Every terminal state in this set is entered only
// AFTER the state machine has run a full cleanup pass AND confirmed that the
// app namespace is gone:
//   - Uninstalled            ← UninstallingApp.exec ran ops.Uninstall(All) +
//     waitForDeleteNamespace
//   - InstallingCanceled     ← InstallingCancelingApp ran ops.Uninstall(All) +
//     poll() waited for NS IsNotFound
//   - InstallFailed          ← cleanupAfterInstallFailure runs ops.Uninstall()
//     AND polls for NS IsNotFound before the transition; InstallFailedApp.Exec
//     retries on every reconcile if the synchronous wait timed out
//   - PendingCanceled / DownloadingCanceled / DownloadFailed
//     ← cluster resources never created
//
// Callers that delete the AM based on this predicate should still re-verify the
// app namespace is IsNotFound to defend against pre-upgrade dirty data where
// the invariant may not yet hold.
func IsSafelyDeletable(state appv1alpha1.ApplicationManagerState) bool {
	switch state {
	case appv1alpha1.PendingCanceled,
		appv1alpha1.DownloadingCanceled,
		appv1alpha1.InstallingCanceled,
		appv1alpha1.Uninstalled,
		appv1alpha1.DownloadFailed,
		appv1alpha1.InstallFailed:
		return true
	}
	return false
}
