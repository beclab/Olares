package handlers

import (
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// Register wires every real handler in this package into `table`. It
// is the single entry point production wiring should call after
// domain.NewDefaultTable():
//
//	table := domain.NewDefaultTable()
//	handlers.Register(table)
//	engine := domain.NewEngine(table)
//	runner, _ := runner.New(runner.Config{Engine: engine, ...})
//
// Idempotent: calling Register twice on the same table re-installs
// the same handlers on the same edges — the second call is a no-op
// with respect to behaviour but redundant work.
//
// Each per-state registration function (registerInstalling,
// registerDownloading, ...) is package-private so external callers
// cannot cherry-pick a subset. This forces the "all handlers move
// together" discipline: if a caller wants to keep placeholder
// behaviour for state X, that must be a deliberate table-editing
// step AFTER Register, not a decision made by omission.
func Register(table *domain.TransitionTable) {
	registerPending(table)
	registerDownloading(table)
	registerInstalling(table)
	registerInitializing(table)
	registerRunning(table)
	registerUninstalling(table)
	registerStopping(table)
	registerStopped(table)
	registerResuming(table)

	registerUpgrading(table)
	registerUpgradePending(table)
	registerUpgradeDownloading(table)
	registerUpgradeDownloadCancel(table)
	registerApplyingEnv(table)

	registerUninstalled(table)
	registerInstallFailed(table)
	registerDownloadFailed(table)
	registerUninstallFailed(table)
	registerStopFailed(table)
	registerResumeFailed(table)
	registerUpgradeFailed(table)
	registerApplyEnvFailed(table)

	// Cancel-handoff group — Pattern C (Initializing/Resuming/
	// Upgrading/ApplyingEnv Canceling → Stopping). See
	// cancel_handoff.go for the full four-state pattern doc.
	registerInitializingCanceling(table)
	registerResumingCanceling(table)
	registerUpgradingCanceling(table)
	registerApplyingEnvCanceling(table)

	// Cancel-pipeline group — Pattern A/B (Pending/Downloading/
	// Installing Canceling → *Canceled / *CancelFailed). See
	// cancel_pipeline.go for the per-state phase breakdown.
	registerPendingCanceling(table)
	registerDownloadingCanceling(table)
	registerInstallingCanceling(table)

	// Cancel-terminal group — the 3 *Canceled + 7 *CancelFailed
	// states, all with Tick no-op + op-driven-restart handlers.
	// See terminal_states.go for the full family taxonomy.
	registerPendingCanceled(table)
	registerDownloadingCanceled(table)
	registerInstallingCanceled(table)
	registerPendingCancelFailed(table)
	registerDownloadingCancelFailed(table)
	registerInstallingCancelFailed(table)
	registerInitializingCancelFailed(table)
	registerUpgradingCancelFailed(table)
	registerApplyingEnvCancelFailed(table)
	registerResumingCancelFailed(table)

	registerTimeouts(table)

	// Additional handlers land here as they are ported. Kept as
	// explicit one-liners (rather than a slice iterated over)
	// because a reviewer scanning this file benefits from seeing
	// each registration and its filename side by side.
}
