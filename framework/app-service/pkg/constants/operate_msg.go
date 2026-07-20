package constants

// describes the template for operation to record operate history.
const (
	// InstallOperationCompletedTpl is for successful install operation.
	InstallOperationCompletedTpl = "Successfully installed %s: %s"

	// Cancel-by-timeout messages, written by *App.Cancel() when the
	// reconcile loop detects IsTimeout()==true and transitions the AM to
	// the corresponding *Canceling state (see appmgr_controller.go's
	// timeout branch — Cancel() is invoked exclusively from there, not
	// from the user-initiated cancel handler which writes status
	// directly). Each is operation-specific so the resulting NATS push /
	// UI status banner / audit log can distinguish *which* operation
	// timed out rather than seeing a generic state-name string. These
	// messages are also preserved through the *Canceling -> Stopping
	// transition by baseStatefulApp.finishCancelToStopping, so the
	// Stopping push for cancel-cleanup carries the same context.
	InstallCanceledByTimeout    = "Install canceled. Operation timed out."
	DownloadCanceledByTimeout   = "Download canceled. Operation timed out."
	UpgradeCanceledByTimeout    = "Upgrade canceled. Operation timed out."
	ResumeCanceledByTimeout     = "Resume canceled. Operation timed out."
	ApplyEnvCanceledByTimeout   = "Apply env canceled. Operation timed out."
	InitializeCanceledByTimeout = "Initialization canceled. Operation timed out."

	// Cancel-by-user messages, written by handler_installer_cancel.go::cancel
	// when a user posts POST /apps/{name}/cancel (default ?type=operate,
	// the only call site today). Parallel to the *CanceledByTimeout
	// constants above so downstream consumers see consistent wording
	// across both cancel paths and can still distinguish them by the
	// trailing "Operation timed out." vs "Operation by user." sentence.
	// The handler maps (cancelType, targetCancelingState) -> one of
	// these via cancelStatus; see also the lineage test's
	// mirrorCancelStatus which must stay in sync.
	InstallCanceledByUser    = "Install canceled. Operation by user."
	DownloadCanceledByUser   = "Download canceled. Operation by user."
	UpgradeCanceledByUser    = "Upgrade canceled. Operation by user."
	ResumeCanceledByUser     = "Resume canceled. Operation by user."
	ApplyEnvCanceledByUser   = "Apply env canceled. Operation by user."
	InitializeCanceledByUser = "Initialization canceled. Operation by user."

	// Cancel reason tags, written alongside the *Canceled* messages above
	// to Status.Reason. These structured camelCase identifiers encode
	// both the operation that was canceled *and* the trigger (user vs.
	// system/timeout), so consumers reading Status.Reason on the
	// *Canceling state — and on the subsequent Stopping state, which
	// inherits Reason through baseStatefulApp.finishCancelToStopping's
	// preserve-on-empty semantics — can branch on the cancel kind
	// without parsing the human-readable Message string.
	//
	// Pairing convention:
	//   - *CancelByUser   pairs with *CanceledByUser   (handler path).
	//   - *CancelBySystem pairs with *CanceledByTimeout (reconcile path
	//     in *App.Cancel(), which today only fires on IsTimeout()).
	// "System" is used rather than "Timeout" here so the tag still reads
	// sensibly if non-timeout reconcile-driven cancels are ever added.
	InstallCancelByUser    = "installCancelByUser"
	DownloadCancelByUser   = "downloadCancelByUser"
	UpgradeCancelByUser    = "upgradeCancelByUser"
	ResumeCancelByUser     = "resumeCancelByUser"
	ApplyEnvCancelByUser   = "applyEnvCancelByUser"
	InitializeCancelByUser = "initializeCancelByUser"

	InitializeCancelDueToWaitForLaunchTimedOut = "initialize waitFor timed out"

	InstallCancelBySystem    = "installCancelBySystem"
	DownloadCancelBySystem   = "downloadCancelBySystem"
	UpgradeCancelBySystem    = "upgradeCancelBySystem"
	ResumeCancelBySystem     = "resumeCancelBySystem"
	ApplyEnvCancelBySystem   = "applyEnvCancelBySystem"
	InitializeCancelBySystem = "initializeCancelBySystem"

	// UninstallOperationCompletedTpl is for successful uninstall operation.
	UninstallOperationCompletedTpl = "Successfully uninstalled %s: %s"
	// UpgradeOperationCompletedTpl is for successful upgrade operation.
	UpgradeOperationCompletedTpl = "Successfully upgraded %s: %s"
	// ApplyEnvOperationCompletedTpl is for successful upgrade operation.
	ApplyEnvOperationCompletedTpl = "Successfully applied env to %s: %s"
	// StopOperationCompletedTpl is for suspend operation.
	StopOperationCompletedTpl = "%s stopped."

	// OperationFailedTpl is for failed opration.
	OperationFailedTpl = "Failed to %s: %s"
)
