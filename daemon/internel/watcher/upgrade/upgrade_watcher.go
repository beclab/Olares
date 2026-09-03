package upgrade

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/beclab/Olares/daemon/pkg/utils"

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
	target        *state.UpgradeTarget
	cancel        context.CancelFunc
}

func NewUpgradeWatcher() watcher.Watcher {
	w := &upgradeWatcher{}
	return w
}

func (w *upgradeWatcher) Watch(ctx context.Context) {
	// A post-upgrade reboot can be in progress while the upgrade target file
	// still exists: in the daemon-driven flow the target is only removed by a
	// later phase that the reboot usually preempts, so we cannot rely on
	// target == nil to detect it. Check this up front, regardless of the
	// target: if olares-cli wrote the reboot marker (before flipping the
	// OlaresVersion CR) and the system is actually shutting down, the upgrade
	// is not really done yet - it only takes full effect after the reboot - so
	// surface a dedicated "Rebooting" step instead of letting the rest of the
	// watcher report it as complete. The marker stat is cheap and runs every
	// tick; the shutdown probe (dbus) only runs when the marker is present.
	// Gating on the real shutdown signal also means a stale marker left behind
	// by a reboot that never happened cannot wedge the upgrade state.
	if _, statErr := os.Stat(state.UpgradeRebootMarkFile); statErr == nil {
		if shuttingDown, _ := state.IsSystemShuttingdown(); shuttingDown {
			state.TerminusStateMu.Lock()
			state.CurrentState.UpgradingState = state.InProgress
			state.CurrentState.UpgradingStep = state.UpgradeStepRebooting
			state.CurrentState.UpgradingRetryNum = 0
			state.CurrentState.UpgradingNextRetryAt = nil
			state.CurrentState.UpgradingError = ""
			state.TerminusStateMu.Unlock()
			return
		}
	}

	var err error
	w.target, err = state.GetOlaresUpgradeTarget()
	if err != nil {
		klog.Errorf("failed to check upgrade target: %v", err)
		return
	}

	if w.target == nil {
		if w.cancel != nil {
			w.cancel()
			w.cancel = nil
		}
		w.resetRetryState()
		// The orchestrator does not watch the target file. Stopping the
		// watcher without stopping the run leaves a cluster upgrade going,
		// and a daemon that restarts afterwards resumes it — which is how
		// deleting the file, without the signed cancel that also calls
		// RequestStop, used to keep preparing nodes after a cancel.
		upgrade.StopActiveClusterUpgrade()

		state.TerminusStateMu.Lock()
		state.CurrentState.UpgradingState = ""
		state.CurrentState.UpgradingTarget = ""
		state.CurrentState.UpgradingOperationID = ""
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
	if err != nil || currentVersion.LessThan(&w.target.Version) {
		state.CurrentState.UpgradingTarget = w.target.Version.Original()
	} else if !w.isUpgrading() {
		// The version CR has reached the target. On a single node that means
		// the upgrade is over. On a cluster it does not, and the gap is where
		// the reboot lives: post-upgrade-admin flips the CR, and reboot-nodes
		// runs after it — the stage that takes this very machine down. The
		// daemon that comes back therefore finds the target reached with no
		// upgrade running in this process, while the orchestrator is quietly
		// resuming the rest of the run behind it.
		//
		// Removing the target here does two things, both wrong. It calls
		// StopClusterUpgrade, which cancels the run that is still going. And
		// it clears the upgrading state, so anything that fails from this
		// point on — a node that never comes back from its reboot — is shown
		// to nobody: the status this whole flow is followed by goes blank, and
		// the only remaining trace is an operation record no one is looking at.
		outcome, orchestrated := upgrade.ClusterUpgradeOutcome(w.target.Version.Original())
		switch {
		case orchestrated && !outcome.Settled:
			// Still running. Say so and leave it be — falling through would
			// start a second upgrade alongside the one already going.
			state.CurrentState.UpgradingState = state.InProgress
			state.CurrentState.UpgradingTarget = w.target.Version.Original()
			state.CurrentState.UpgradingOperationID = outcome.ID
			state.CurrentState.UpgradingError = ""
			return
		case orchestrated && !outcome.Succeeded:
			// Over, and it did not work. Keeping the target is the whole of
			// the fix: it puts this back in the hands of the ordinary failure
			// path below, reported as a failed upgrade and retried on the
			// usual backoff, which is the answer a failure at any earlier
			// point in the run already gets. An operator who brings the
			// missing node back gets the rest finished on the next attempt.
			//
			// The state and the error are deliberately left to that path.
			// They describe the attempt happening now — "node olares-worker1:
			// NetworkUnavailable" — where this record describes the one that
			// already ended, and the fresher of the two is the one worth
			// showing. The id is set because it is the handle to what the
			// flat fields cannot say: which stage failed, and on which node.
			state.CurrentState.UpgradingTarget = w.target.Version.Original()
			state.CurrentState.UpgradingOperationID = outcome.ID
		default:
			w.target = nil
			_, err = upgrade.NewRemoveUpgradeTarget().Execute(ctx, nil)
			if err != nil {
				klog.Error("failed to remove upgrade files: ", err)
			}
			return
		}
	}

	if !w.isUpgrading() {
		if !w.isTimeToRetry() {
			return
		}

		exeCtx, cancel := context.WithCancel(ctx)
		w.cancel = cancel

		go func() {
			w.startUpgrading()
			defer w.stopUpgrading()
			if err := w.doUpgradeWithRetry(exeCtx); err != nil {
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
	err := w.doUpgrade(ctx)
	if err != nil {
		w.incrementRetry()

		state.CurrentState.UpgradingRetryNum = w.getRetryCount()
		state.CurrentState.UpgradingNextRetryAt = w.nextRetryTime

		klog.Errorf("upgrade attempt %d failed: %v. Next retry scheduled for %v",
			w.getRetryCount(), err, *w.nextRetryTime)
	}
	return err
}

type upgradePhase struct {
	newCMD         func() commands.Interface
	progressOffset int
	progressSpan   int
}

// weigh attaches this watcher's progress bar to a shared phase sequence.
//
// Which phases there are, and in what order, belongs to the package that
// implements them, because a compute node runs the same ones; only the
// percentages are the watcher's, and only the watcher has a bar to fill. The
// span pairs are positional, so a phase added to the shared list without a
// weight here stops the daemon at startup rather than silently shifting every
// number after it.
func weigh(phases []func() commands.Interface, spans ...int) []upgradePhase {
	if len(spans) != 2*len(phases) {
		panic(fmt.Sprintf("upgrade watcher: %d phases need %d progress bounds, got %d",
			len(phases), 2*len(phases), len(spans)))
	}
	weighted := make([]upgradePhase, 0, len(phases))
	for i, newCMD := range phases {
		weighted = append(weighted, upgradePhase{newCMD, spans[2*i], spans[2*i+1]})
	}
	return weighted
}

func phases(groups ...[]upgradePhase) []upgradePhase {
	var all []upgradePhase
	for _, g := range groups {
		all = append(all, g...)
	}
	return all
}

var downloadPhases = weigh(upgrade.ReleaseDownloadPhases,
	0, 10,
	10, 20,
	30, 10,
	40, 60,
)

// upgradePhases is the control node upgrading itself: check, adopt the
// release, run the upgrade, forget the target.
//
// The middle of it is the same sequence, in the same order, that the
// orchestrator has a compute node run before it may be given a stage. See
// upgrade.ReleaseAdoptPhases for why that is shared rather than written twice.
var upgradePhases = phases(
	[]upgradePhase{{upgrade.NewPreCheck, 0, 10}},
	weigh(upgrade.ReleaseAdoptPhases,
		10, 10,
		20, 10,
		30, 30,
	),
	[]upgradePhase{
		{upgrade.NewUpgrade, 60, 35},
		{upgrade.NewRemoveTarget, 95, 5},
	},
)

func (w *upgradeWatcher) doUpgrade(ctx context.Context) (err error) {
	target := w.target
	if target == nil {
		return nil
	}
	targetVersionLogsDir := filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+target.Version.Original(), "logs")
	prepareLogFile := filepath.Join(targetVersionLogsDir, "install.log")
	upgradeLogFile := filepath.Join(targetVersionLogsDir, "upgrade.log")
	for _, logFile := range []string{prepareLogFile, upgradeLogFile} {
		if err := os.Remove(logFile); err != nil && !os.IsNotExist(err) {
			klog.Errorf("failed to clear log file %s: %v", logFile, err)
		}
	}
	if !target.Downloaded {
		// Execute download phases
		return doDownloadPhases(ctx, *target)
	}

	klog.Info("download already completed, skipping download phases")

	if target.DownloadOnly {
		state.CurrentState.UpgradingState = "WaitingForUserConfirm"
		klog.Info("download completed, waiting for user request to remove upgrade.downloadonly file to proceed with upgrade")
		return nil
	}

	return doUpgradePhases(ctx, *target)
}

func doDownloadPhases(ctx context.Context, target state.UpgradeTarget) (err error) {
	defer func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err != nil {
			state.CurrentState.UpgradingDownloadState = state.Failed
			state.CurrentState.UpgradingDownloadError = err.Error()
			klog.Errorf("download phases failed: %v", err)
		} else {
			state.CurrentState.UpgradingDownloadState = state.Completed
			state.CurrentState.UpgradingDownloadProgress = "100%"
			state.CurrentState.UpgradingDownloadProgressNum = 100
			state.CurrentState.UpgradingDownloadError = ""
			klog.Info("download phases completed successfully")
		}
	}()

	state.CurrentState.UpgradingDownloadState = state.InProgress
	state.CurrentState.UpgradingDownloadError = ""

	for _, phase := range downloadPhases {
		phaseCMD := phase.newCMD()
		state.CurrentState.UpgradingDownloadStep = string(phaseCMD.OperationName())

		res, err := phaseCMD.Execute(ctx, target)
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
	return markDownloaded()
}

func doUpgradePhases(ctx context.Context, target state.UpgradeTarget) (err error) {
	defer func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
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

		res, err := phaseCMD.Execute(ctx, target)
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

func markDownloaded() error {
	target, err := state.GetOlaresUpgradeTarget()
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("no target found to mark downloaded, possibly upgrade cancelled")
	}
	if !target.Downloaded {
		target.Downloaded = true
		return target.Save()
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
