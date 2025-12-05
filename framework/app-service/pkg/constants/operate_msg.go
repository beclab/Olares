package constants

// describes the template for operation to record operate history.
const (
	// InstallOperationCompletedTpl is for successful install operation.
	InstallOperationCompletedTpl = "Successfully installed %s: %s"
	// OperationCanceledByUserTpl is for cancel operation by user.
	OperationCanceledByUserTpl = "Install canceled by user."
	// OperationCanceledByTerminusTpl is for cancel operation by system.
	OperationCanceledByTerminusTpl = "Install canceled. Operation timed out."
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
