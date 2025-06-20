package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
	semver "github.com/Masterminds/semver/v3"
)

type prepareOlaresd struct {
	commands.Operation
	logFile          string
	progressKeywords []progressKeyword
	progress         int
	progressChan     chan<- int
}

var _ commands.Interface = &prepareOlaresd{}

func NewInstallOlaresd() commands.Interface {
	return &prepareOlaresd{
		Operation: commands.Operation{
			Name: commands.InstallOlaresd,
		},
		progressKeywords: []progressKeyword{
			{"ReplaceOlaresdBinary success", 30},
			// if executed by olaresd
			// this will be the last log printed by olares-cli
			// as the olares-cli process will exit along with
			// the olaresd when olares-cli restarts olaresd
			{"UpdateOlaresdEnv success", commands.ProgressNumFinished},
			// if executed manually by user
			// these logs will be seen,
			// but they're very likely to never be processed by us
			{"RestartOlaresd success", commands.ProgressNumFinished},
			{"[Job] Prepare Olaresd daemon execute successfully", commands.ProgressNumFinished},
		},
	}
}

func (i *prepareOlaresd) Execute(ctx context.Context, p any) (res any, err error) {
	version, ok := p.(string)
	if !ok {
		return nil, errors.New("invalid param")
	}
	targetVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid target version %s: %v", version, err)
	}

	currentVersion, err := getCurrentDaemonVersion()
	if err != nil {
		klog.Warningf("Failed to get current olaresd version: %v, proceeding with installation", err)
	} else {
		if !currentVersion.LessThan(targetVersion) {
			return newExecutionRes(true, nil), nil
		}
	}

	i.logFile = filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "logs", "install.log")
	if err := i.refreshProgress(); err != nil {
		return nil, fmt.Errorf("could not determine whether olaresd prepare is finished: %v", err)
	}
	if i.progress == commands.ProgressNumFinished {
		return newExecutionRes(true, nil), nil
	}

	progressChan := make(chan int, 100)
	i.progressChan = progressChan

	cmd := commands.NewBaseCommand()
	cmd.WithWatchDog_(i.watch)

	params := []string{
		"prepare", "olaresd",
		"--version", version,
		"--base-dir", commands.TERMINUS_BASE_DIR,
	}
	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, params...); err != nil {
		return nil, err
	}

	return newExecutionRes(false, progressChan), nil
}

func (i *prepareOlaresd) watch(ctx context.Context) {
	go func() {
		defer close(i.progressChan)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if i.progress != commands.ProgressNumFinished {
					if err := i.refreshProgress(); err != nil {
						klog.Errorf("failed to refresh olaresd prepare progress upon context done: %v", err)
					}
				}
				return
			case <-ticker.C:
				if err := i.refreshProgress(); err != nil {
					klog.Errorf("failed to refresh olaresd prepare progress: %v", err)
				}
			}
		}
	}()
}

// todo: check finish state by current olaresd binary version after the version of olaresd and olaresd has been unified
func (i *prepareOlaresd) refreshProgress() error {
	info, err := os.Stat(i.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		klog.Errorf("error stat olaresd prepare log file %s: %v", i.logFile, err)
		return err
	}

	filesize := info.Size()
	tailsize := min(filesize, 40960)

	t, err := tail.TailFile(i.logFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Errorf("error tail olaresd prepare file %s: %v", i.logFile, err)
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
