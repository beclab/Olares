package upgrade

import (
	"bytetrade.io/web3os/terminusd/pkg/cli"
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

type downloadWizard struct {
	commands.Operation
	logFile             string
	assumeFinishedFiles []string
	progressKeywords    []progressKeyword
	progress            int
	progressChan        chan<- int
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
	version, ok := p.(string)
	if !ok {
		return nil, errors.New("invalid param")
	}

	i.logFile = filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "logs", "install.log")
	i.assumeFinishedFiles = []string{
		filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "version.hint"),
		filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "installation.manifest"),
		filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+version, "wizard"),
	}
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
	if commands.CDN_URL != "" {
		params = append(params, "--download-cdn-url", commands.CDN_URL)
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
	allExisting := true
	for _, f := range i.assumeFinishedFiles {
		_, err := os.Stat(f)
		if os.IsNotExist(err) {
			allExisting = false
			break
		}
		if err != nil {
			return fmt.Errorf("could not stat file %s: %v", f, err)
		}
	}
	if allExisting {
		i.progress = commands.ProgressNumFinished
		return nil
	}
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
