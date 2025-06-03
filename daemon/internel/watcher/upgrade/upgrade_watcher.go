package upgrade

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"bytetrade.io/web3os/terminusd/internel/watcher"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"bytetrade.io/web3os/terminusd/pkg/commands/upgrade"
	"k8s.io/klog/v2"
)

type upgradeWatcher struct {
	watcher.Watcher
	sync.Mutex
	upgrading bool
}

func NewUpgradeWatcher() watcher.Watcher {
	w := &upgradeWatcher{}
	return w
}

func (w *upgradeWatcher) Watch(ctx context.Context) {
	switch state.CurrentState.TerminusState {
	// indicates an upgrade target exists
	case state.Upgrading:
		// if the upgrade process is running, just wait for it to finish
		if !w.isUpgrading() {
			go func() {
				w.startUpgrading()
				defer w.stopUpgrading()
				if err := doUpgrade(ctx); err != nil {
					klog.Errorf("upgrading error: %v", err)
				}
			}()
		}
	}
}

func (w *upgradeWatcher) isUpgrading() bool {
	w.Lock()
	defer w.Unlock()
	return w.upgrading
}

func (w *upgradeWatcher) startUpgrading() {
	w.Lock()
	defer w.Unlock()
	w.upgrading = true
}

func (w *upgradeWatcher) stopUpgrading() {
	w.Lock()
	defer w.Unlock()
	w.upgrading = false
}

type upgradePhase struct {
	newCMD         func() commands.Interface
	progressOffset int
	progressSpan   int
}

// todo: add a phase to upgrade olares-cli after the version of olares-cli and olares has been unified
var upgradePhases = []upgradePhase{
	{upgrade.NewUpgradeCli, 0, 5},
	{upgrade.NewDownloadWizard, 5, 10},
	{upgrade.NewVersionCompatibilityCheck, 15, 0},
	{upgrade.NewHealthCheck, 15, 0},
	{upgrade.NewDownloadComponent, 15, 30},
	{upgrade.NewPrepareImages, 45, 30},
	{upgrade.NewPrepareOlaresd, 75, 5},
	{upgrade.NewUpgrade, 80, 15},
	{upgrade.NewRemoveTarget, 95, 5},
}

func doUpgrade(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			state.CurrentState.UpgradingState = state.Failed
			state.CurrentState.UpgradingError = err.Error()

			// clear logs after every failed attempt
			// in case any under layer change that bypassed olaresd, e.g., manual removal of files
			// is causing the upgrade retry to stuck forever
			targetVersionDir := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+state.CurrentState.UpgradingTarget)
			prepareLogFile := filepath.Join(targetVersionDir, "install.log")
			upgradeLogFile := filepath.Join(targetVersionDir, "upgrade.log")
			for _, logFile := range []string{prepareLogFile, upgradeLogFile} {
				if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
					klog.Errorf("failed to clear log file %s of current upgrade attempt (%d): %v", logFile, state.CurrentState.UpgradingRetryNum, err)
				}
			}
		}
	}()

	state.CurrentState.UpgradingState = state.InProgress
	state.CurrentState.UpgradingError = ""
	state.CurrentState.UpgradingRetryNum += 1

	for _, phase := range upgradePhases {
		phaseCMD := phase.newCMD()
		state.CurrentState.UpgradingStep = string(phaseCMD.OperationName())
		res, err := phaseCMD.Execute(ctx, state.CurrentState.UpgradingTarget)
		if err != nil {
			return fmt.Errorf("error: upgrade phase %s: %v", phaseCMD.OperationName(), err)
		}
		executionRes, ok := res.(upgrade.ExecutionRes)
		if !ok {
			return fmt.Errorf("unexpected result type for upgrade phase %s", phaseCMD.OperationName())
		}
		if executionRes.Finished() {
			// for now, do not update progress here
			// as it may revert back the progress
			// todo: if the retry num will be presented by the frontend to user, maybe we can update progress here
			continue
		}
		var phaseProgress int
		for phaseProgress < 100 {
			select {
			case <-ctx.Done():
				return nil
			case p, ok := <-executionRes.Progress():
				// the command completed and the progress channel is closed
				if !ok {
					if phaseProgress != commands.ProgressNumFinished {
						return fmt.Errorf("error: upgrade phase %s: command execution did not succeed", phaseCMD.OperationName())
					}
				} else if p > phaseProgress {
					klog.Infof("refreshing upgrading phase %s, progress: %d", phaseCMD.OperationName(), phaseProgress)
					phaseProgress = p
				}
			}
			refreshUpgradeProgressFromPhase(phase, phaseProgress)
		}
	}
	// if the upgrade succeeded, the upgrade target will be removed
	// and the upgrade status cleared
	return nil
}

func refreshUpgradeProgressFromPhase(phase upgradePhase, phaseProgress int) {
	spanProgress := math.Min(float64(phaseProgress)*float64(phase.progressSpan)/float64(commands.ProgressNumFinished), float64(phase.progressSpan))
	newProgress := phase.progressOffset + int(math.Round(spanProgress))
	if state.CurrentState.UpgradingProgressNum >= newProgress {
		return
	}
	state.CurrentState.UpgradingProgressNum = newProgress
	state.CurrentState.UpgradingProgress = fmt.Sprintf("%d%%", state.CurrentState.UpgradingProgressNum)
}
