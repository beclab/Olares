package upgrade

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/internel/watcher"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/beclab/Olares/daemon/pkg/commands/upgrade"
	"k8s.io/klog/v2"
)

type upgradeWatcher struct {
	watcher.Watcher
	sync.Mutex
	upgrading bool
	// Internal retry state
	retryCount    int
	nextRetryTime *time.Time
}

func NewUpgradeWatcher() watcher.Watcher {
	w := &upgradeWatcher{}
	return w
}

func (w *upgradeWatcher) Watch(ctx context.Context) {
	targetVersion, err := state.GetOlaresUpgradeTarget()
	if err != nil {
		klog.Errorf("failed to check upgrade target: %v", err)
		return
	}

	if targetVersion == nil {
		w.resetRetryState()

		state.TerminusStateMu.Lock()
		state.CurrentState.UpgradingState = ""
		state.CurrentState.UpgradingTarget = ""
		state.CurrentState.UpgradingRetryNum = 0
		state.CurrentState.UpgradingNextRetryAt = nil
		state.CurrentState.UpgradingStep = ""
		state.CurrentState.UpgradingProgressNum = 0
		state.CurrentState.UpgradingProgress = ""
		state.CurrentState.UpgradingError = ""

		state.CurrentState.UpgradingDownloadState = ""
		state.CurrentState.UpgradingDownloadStep = ""
		state.CurrentState.UpgradingDownloadProgressNum = 0
		state.CurrentState.UpgradingDownloadProgress = ""
		state.CurrentState.UpgradingDownloadError = ""
		state.TerminusStateMu.Unlock()

		return
	}

	dynamicClient, err := utils.GetDynamicClient()
	if err != nil {
		return
	}

	currentVersionStr, err := utils.GetTerminusVersion(ctx, dynamicClient)
	if err != nil {
		klog.Error("failed to get current version, skip upgrading check: ", err)
		return
	}
	if currentVersionStr == nil {
		klog.Error("current version is nil, skip upgrading check")
		return
	}
	currentVersion, err := semver.NewVersion(*currentVersionStr)
	if err != nil || currentVersion.LessThan(targetVersion) {
		state.CurrentState.UpgradingTarget = targetVersion.Original()
	} else {
		err = upgrade.RemoveUpgradeFiles()
		if err != nil {
			klog.Error("failed to remove upgrade files: ", err)
		}
		return
	}

	if !w.isUpgrading() {
		if !w.isTimeToRetry() {
			return
		}

		go func() {
			w.startUpgrading()
			defer w.stopUpgrading()
			if err := w.doUpgradeWithRetry(ctx); err != nil {
				klog.Errorf("upgrading error: %v", err)
			}
		}()
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

func (w *upgradeWatcher) isTimeToRetry() bool {
	w.Lock()
	defer w.Unlock()

	if w.nextRetryTime == nil {
		return true
	}

	now := time.Now()
	if now.Before(*w.nextRetryTime) {
		klog.V(2).Infof("upgrade retry scheduled for %v (in %v)",
			*w.nextRetryTime,
			w.nextRetryTime.Sub(now))
		return false
	}

	return true
}

func (w *upgradeWatcher) resetRetryState() {
	w.Lock()
	defer w.Unlock()
	w.retryCount = 0
	w.nextRetryTime = nil
}

func (w *upgradeWatcher) incrementRetry() {
	w.Lock()
	defer w.Unlock()
	w.retryCount++
	nextRetry := state.CalculateNextRetryTime(w.retryCount)
	w.nextRetryTime = &nextRetry
}

func (w *upgradeWatcher) getRetryCount() int {
	w.Lock()
	defer w.Unlock()
	return w.retryCount
}

func (w *upgradeWatcher) doUpgradeWithRetry(ctx context.Context) error {
	err := doUpgrade(ctx)
	if err != nil {
		w.incrementRetry()

		state.CurrentState.UpgradingRetryNum = w.getRetryCount()
		state.CurrentState.UpgradingNextRetryAt = w.nextRetryTime

		klog.Errorf("upgrade attempt %d failed: %v. Next retry scheduled for %v",
			w.getRetryCount(), err, *w.nextRetryTime)

		targetVersionDir := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+state.CurrentState.UpgradingTarget)
		prepareLogFile := filepath.Join(targetVersionDir, "install.log")
		upgradeLogFile := filepath.Join(targetVersionDir, "upgrade.log")
		for _, logFile := range []string{prepareLogFile, upgradeLogFile} {
			if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
				klog.Errorf("failed to clear log file %s: %v", logFile, err)
			}
		}
	}
	return err
}

type upgradePhase struct {
	newCMD         func() commands.Interface
	progressOffset int
	progressSpan   int
}

var downloadPhases = []upgradePhase{
	{upgrade.NewDownloadCLI, 0, 10},
	{upgrade.NewDownloadWizard, 10, 20},
	{upgrade.NewDownloadComponent, 30, 40},
}

var upgradePhases = []upgradePhase{
	{upgrade.NewVersionCompatibilityCheck, 0, 5},
	{upgrade.NewHealthCheck, 5, 5},
	{upgrade.NewInstallCLI, 10, 10},
	{upgrade.NewImportImages, 20, 30},
	{upgrade.NewInstallOlaresd, 50, 10},
	{upgrade.NewUpgrade, 60, 35},
	{upgrade.NewRemoveTarget, 95, 5},
}

func doUpgrade(ctx context.Context) (err error) {
	downloadCompleted, err := state.IsUpgradeDownloadCompleted()
	if err != nil {
		return fmt.Errorf("failed to check download status: %v", err)
	}

	if !downloadCompleted {
		// Execute download phases
		if err := doDownloadPhases(ctx); err != nil {
			return err
		}
	} else {
		klog.Info("download already completed, skipping download phases")
		state.CurrentState.UpgradingDownloadState = state.Completed
		state.CurrentState.UpgradingDownloadProgress = "100%"
		state.CurrentState.UpgradingDownloadProgressNum = 100
	}

	downloadOnly, err := state.IsUpgradeDownloadOnly()
	if err != nil {
		return fmt.Errorf("failed to check download-only status: %v", err)
	}

	if downloadOnly {
		state.CurrentState.UpgradingState = "WaitingForUserConfirm"
		klog.Info("download completed, waiting for user request to remove upgrade.downloadonly file to proceed with upgrade")
		return nil
	}

	return doUpgradePhases(ctx)
}

func doDownloadPhases(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			state.CurrentState.UpgradingDownloadState = state.Failed
			state.CurrentState.UpgradingDownloadError = err.Error()
			klog.Errorf("download phases failed: %v", err)
		} else {
			state.CurrentState.UpgradingDownloadState = state.Completed
			state.CurrentState.UpgradingDownloadError = ""
			if err := createUpgradeDownloadedFile(); err != nil {
				klog.Errorf("failed to create upgrade.downloaded file: %v", err)
			}
			klog.Info("download phases completed successfully")
		}
	}()

	state.CurrentState.UpgradingDownloadState = state.InProgress
	state.CurrentState.UpgradingDownloadError = ""

	for _, phase := range downloadPhases {
		phaseCMD := phase.newCMD()
		state.CurrentState.UpgradingDownloadStep = string(phaseCMD.OperationName())

		res, err := phaseCMD.Execute(ctx, state.CurrentState.UpgradingTarget)
		if err != nil {
			return fmt.Errorf("error: download phase %s: %v", phaseCMD.OperationName(), err)
		}
		executionRes, ok := res.(upgrade.ExecutionRes)
		if !ok {
			return fmt.Errorf("unexpected result type for download phase %s", phaseCMD.OperationName())
		}
		if executionRes.Finished() {
			continue
		}
		var phaseProgress int
		for phaseProgress < 100 {
			select {
			case <-ctx.Done():
				return nil
			case p, ok := <-executionRes.Progress():
				if !ok {
					if phaseProgress != commands.ProgressNumFinished {
						return fmt.Errorf("error: download phase %s: command execution did not succeed", phaseCMD.OperationName())
					}
				} else if p > phaseProgress {
					klog.Infof("refreshing download phase %s, progress: %d", phaseCMD.OperationName(), phaseProgress)
					phaseProgress = p
				}
			}
			refreshDownloadProgressFromPhase(phase, phaseProgress)
		}
	}
	return nil
}

func doUpgradePhases(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			state.CurrentState.UpgradingState = state.Failed
			state.CurrentState.UpgradingError = err.Error()
		}
	}()

	state.CurrentState.UpgradingState = state.InProgress
	state.CurrentState.UpgradingError = ""

	state.StateTrigger <- struct{}{}

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
			continue
		}
		var phaseProgress int
		for phaseProgress < 100 {
			select {
			case <-ctx.Done():
				return nil
			case p, ok := <-executionRes.Progress():
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

func refreshDownloadProgressFromPhase(phase upgradePhase, phaseProgress int) {
	spanProgress := math.Min(float64(phaseProgress)*float64(phase.progressSpan)/float64(commands.ProgressNumFinished), float64(phase.progressSpan))
	newProgress := phase.progressOffset + int(math.Round(spanProgress))
	if state.CurrentState.UpgradingDownloadProgressNum >= newProgress {
		return
	}
	state.CurrentState.UpgradingDownloadProgressNum = newProgress
	state.CurrentState.UpgradingDownloadProgress = fmt.Sprintf("%d%%", state.CurrentState.UpgradingDownloadProgressNum)
}

func createUpgradeDownloadedFile() error {
	return os.WriteFile(commands.UPGRADE_DOWNLOADED_FILE, []byte(""), 0644)
}
