package appstate

import (
	"time"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
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

	appv1alpha1.PendingCanceled,
	appv1alpha1.DownloadingCanceled,
	appv1alpha1.InstallingCanceled,

	appv1alpha1.Stopped,

	appv1alpha1.DownloadFailed,
	appv1alpha1.InstallFailed,
	appv1alpha1.StopFailed,
	appv1alpha1.UpgradeFailed,
	appv1alpha1.ApplyEnvFailed,
	appv1alpha1.ResumeFailed,
	appv1alpha1.UninstallFailed,
}

var StateTransitions = map[appv1alpha1.ApplicationManagerState][]appv1alpha1.ApplicationManagerState{
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
	},
	appv1alpha1.Stopping: {
		appv1alpha1.Stopped,
		appv1alpha1.StopFailed,
	},
	appv1alpha1.Upgrading: {
		appv1alpha1.Initializing,
		appv1alpha1.UpgradeFailed,
		appv1alpha1.UpgradingCanceling,
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
	},

	appv1alpha1.Resuming: {
		appv1alpha1.ResumingCanceling,
		appv1alpha1.ResumeFailed,
		appv1alpha1.Initializing,
	},
	appv1alpha1.ResumingCanceling: {
		appv1alpha1.Stopping,
	},
	appv1alpha1.UpgradingCanceling: {
		appv1alpha1.Stopping,
	},
	appv1alpha1.ApplyingEnvCanceling: {
		appv1alpha1.Stopping,
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
	},
	appv1alpha1.UninstallFailed: {
		appv1alpha1.Uninstalling,
	},
	appv1alpha1.PendingCancelFailed: {
		appv1alpha1.PendingCanceling,
	},
	appv1alpha1.DownloadingCancelFailed: {
		appv1alpha1.DownloadingCanceling,
	},
	appv1alpha1.InstallingCancelFailed: {
		appv1alpha1.InstallingCanceling,
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

func StateToDuration(state appv1alpha1.ApplicationManagerState) time.Duration {
	if t, ok := StateToDurationMap[state]; ok {
		return t
	}
	return 10 * time.Minute
}
