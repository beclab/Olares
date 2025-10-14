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
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/commands"

	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
)

type downloadWizard struct {
	commands.Operation
	logFile          string
	progressKeywords []progressKeyword
	progress         int
	progressChan     chan<- int
}

var _ commands.Interface = &downloadWizard{}

func NewDownloadWizard() commands.Interface {
	return &downloadWizard{
		Operation: commands.Operation{
			Name: commands.DownloadWizard,
		},
		progressKeywords: []progressKeyword{
			{"[Module] DownloadInstallWizard", 10},
			{"[Job] Download Installation Wizard execute successfully", commands.ProgressNumFinished},
		},
	}
}

func (i *downloadWizard) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}

	version := target.Version.Original()

	i.logFile = filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "logs", "install.log")
	if err := i.refreshProgress(); err != nil {
		return nil, fmt.Errorf("could not determine whether wizard download is finished: %v", err)
	}
	if i.progress == commands.ProgressNumFinished {
		return newExecutionRes(true, nil), nil
	}

	progressChan := make(chan int, 100)
	i.progressChan = progressChan

	cmd := commands.NewBaseCommand()
	cmd.WithWatchDog_(i.watch)

	params := []string{
		"download", "wizard",
		"--version", version,
		"--base-dir", commands.TERMINUS_BASE_DIR,
	}
	if commands.OLARES_CDN_SERVICE != "" {
		params = append(params, "--download-cdn-url", commands.OLARES_CDN_SERVICE)
	}
	if target.WizardURL != "" {
		params = append(params, "--url-override", target.WizardURL)
	}
	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, params...); err != nil {
		return nil, err
	}

	return newExecutionRes(false, progressChan), nil
}

func (i *downloadWizard) watch(ctx context.Context) {
	go func() {
		defer close(i.progressChan)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if i.progress != commands.ProgressNumFinished {
					if err := i.refreshProgress(); err != nil {
						klog.Errorf("failed to refresh wizard download progress upon context done: %v", err)
					}
				}
				return
			case <-ticker.C:
				if err := i.refreshProgress(); err != nil {
					klog.Errorf("failed to refresh wizard download progress: %v", err)
				}
			}
		}
	}()
}

func (i *downloadWizard) refreshProgress() error {
	info, err := os.Stat(i.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		klog.Errorf("error stat wizard download log file %s: %v", i.logFile, err)
		return err
	}

	filesize := info.Size()
	tailsize := min(filesize, 40960)

	t, err := tail.TailFile(i.logFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Errorf("error tail wizard download log file %s: %v", i.logFile, err)
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
