package upgrade

import (
	"bytetrade.io/web3os/terminusd/pkg/cli"
	"bytetrade.io/web3os/terminusd/pkg/cluster/state"
	"bytetrade.io/web3os/terminusd/pkg/commands"
	"context"
	"errors"
	"fmt"
	"github.com/nxadm/tail"
	"io"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type upgrade struct {
	commands.Operation
	logFile          string
	progressKeywords []progressKeyword
	progress         int
	progressChan     chan<- int
}

var _ commands.Interface = &upgrade{}

func NewUpgrade() commands.Interface {
	return &upgrade{
		Operation: commands.Operation{
			Name: commands.Upgrade,
		},
		progressKeywords: []progressKeyword{
			{"PrepareUserInfoForUpgrade success", 5},
			{"ClearAppChartValues success", 10},
			{"ClearBFLChartValues success", 15},
			{"UpdateChartsInAppService success", 20},
			{"UpgradeUserComponents success", 25},
			{"UpdateReleaseFile success", 30},
			{"UpgradeSystemComponents success", 35},
			{"[Job] UpgradeOlares execute successfully", commands.ProgressNumFinished},
		},
	}
}

func (i *upgrade) Execute(ctx context.Context, p any) (res any, err error) {
	version, ok := p.(string)
	if !ok {
		return nil, errors.New("invalid param")
	}

	i.logFile = filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "logs", "upgrade.log")
	if err := i.refreshProgress(); err != nil {
		return nil, fmt.Errorf("could not determine whether upgrade is finished: %v", err)
	}
	if i.progress == commands.ProgressNumFinished {
		return newExecutionRes(true, nil), nil
	}

	progressChan := make(chan int, 100)
	i.progressChan = progressChan

	cmd := commands.NewBaseCommand()
	cmd.WithWatchDog_(i.watch)

	params := []string{
		"upgrade",
		"--version", version,
		"--base-dir", commands.TERMINUS_BASE_DIR,
	}
	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, params...); err != nil {
		return nil, err
	}

	return newExecutionRes(false, progressChan), nil
}

func (i *upgrade) watch(ctx context.Context) {
	go func() {
		defer close(i.progressChan)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if i.progress != commands.ProgressNumFinished {
					if err := i.refreshProgress(); err != nil {
						klog.Errorf("failed to refresh upgrade progress upon context done: %v", err)
					}
				}
				return
			case <-ticker.C:
				if err := i.refreshProgress(); err != nil {
					klog.Errorf("failed to refresh upgrade progress: %v", err)
				}
			}
		}
	}()
}

func (i *upgrade) refreshProgress() error {
	info, err := os.Stat(i.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		klog.Errorf("error stat upgrade log file %s: %v", i.logFile, err)
		return err
	}

	filesize := info.Size()
	tailsize := min(filesize, 4096)

	t, err := tail.TailFile(i.logFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Errorf("error tail upgrade file %s: %v", i.logFile, err)
		return err
	}

	updated := false
	for line := range t.Lines {
		for _, p := range i.progressKeywords {
			if strings.Contains(line.Text, p.KeyWord) {
				if i.progress < p.ProgressNum {
					i.progress = p.ProgressNum
					updated = true
				}
			}
		}
	}

	// smooth progress
	if !updated {
		next := i.progress
		for _, p := range i.progressKeywords {
			if p.ProgressNum > i.progress {
				next = p.ProgressNum
				break
			}
		}

		if next > i.progress+1 {
			i.progress += 1
			updated = true
		}
	}

	if updated && i.progressChan != nil {
		i.progressChan <- i.progress
	}

	return nil
}

type createTarget struct {
	commands.Operation
}

var _ commands.Interface = &createTarget{}

func NewCreateTarget() commands.Interface {
	return &createTarget{
		Operation: commands.Operation{
			Name: commands.CreateUpgradeTarget,
		},
	}
}

func (i *createTarget) Execute(ctx context.Context, p any) (res any, err error) {
	version, ok := p.(string)
	if !ok {
		err = errors.New("invalid param")
		return
	}

	if err := createUpgradeTarget(version); err != nil {
		return nil, fmt.Errorf("failed to create upgrade target: %v", err)
	}

	state.StateTrigger <- struct{}{}

	return nil, nil

}

func createUpgradeTarget(version string) error {
	return os.WriteFile(commands.UPGRADE_TARGET_FILE, []byte(version), 0755)
}

type removeTarget struct {
	commands.Operation
}

var _ commands.Interface = &removeTarget{}

func NewRemoveTarget() commands.Interface {
	return &removeTarget{
		Operation: commands.Operation{
			Name: commands.RemoveUpgradeTarget,
		},
	}
}

func (i *removeTarget) Execute(_ context.Context, _ any) (res any, err error) {
	if err := os.Remove(commands.UPGRADE_TARGET_FILE); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return newExecutionRes(true, nil), nil
}
