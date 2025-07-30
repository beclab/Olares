package upgrade

import (
	"context"
	"errors"
	"fmt"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cli"
	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/nxadm/tail"
	"k8s.io/klog/v2"
)

type prepareImages struct {
	commands.Operation
	logFile          string
	progressKeywords []progressKeyword
	progress         int
	progressChan     chan<- int
}

var _ commands.Interface = &prepareImages{}

func NewImportImages() commands.Interface {
	return &prepareImages{
		Operation: commands.Operation{
			Name: commands.ImportImages,
		},
		progressKeywords: []progressKeyword{
			{"Preload Container Images execute successfully", commands.ProgressNumFinished},
		},
	}
}

func (i *prepareImages) Execute(ctx context.Context, p any) (res any, err error) {
	target, ok := p.(state.UpgradeTarget)
	if !ok {
		return nil, errors.New("invalid param")
	}

	i.logFile = filepath.Join(commands.TERMINUS_BASE_DIR, "versions", "v"+target.Version.Original(), "logs", "install.log")
	if err := i.refreshProgress(); err != nil {
		return nil, fmt.Errorf("could not determine whether images prepare is finished: %v", err)
	}
	if i.progress == commands.ProgressNumFinished {
		return newExecutionRes(true, nil), nil
	}

	progressChan := make(chan int, 100)
	i.progressChan = progressChan

	cmd := commands.NewBaseCommand()
	cmd.WithWatchDog_(i.watch)

	params := []string{
		"prepare", "images",
		"--version", target.Version.Original(),
		"--base-dir", commands.TERMINUS_BASE_DIR,
	}
	if err = cmd.RunAsync_(ctx, cli.TERMINUS_CLI, params...); err != nil {
		return nil, err
	}

	return newExecutionRes(false, progressChan), nil
}

func (i *prepareImages) watch(ctx context.Context) {
	go func() {
		defer close(i.progressChan)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				if i.progress != commands.ProgressNumFinished {
					if err := i.refreshProgress(); err != nil {
						klog.Errorf("failed to refresh images prepare progress upon context done: %v", err)
					}
				}
				return
			case <-ticker.C:
				if err := i.refreshProgress(); err != nil {
					klog.Errorf("failed to refresh images prepare progress: %v", err)
				}
			}
		}
	}()
}

func (i *prepareImages) refreshProgress() error {
	info, err := os.Stat(i.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		klog.Errorf("error stat images prepare log file %s: %v", i.logFile, err)
		return err
	}

	filesize := info.Size()
	tailsize := min(filesize, 40960)

	t, err := tail.TailFile(i.logFile,
		tail.Config{Follow: false, Location: &tail.SeekInfo{Offset: -tailsize, Whence: io.SeekEnd}})
	if err != nil {
		klog.Errorf("error tail images prepare file %s: %v", i.logFile, err)
		return err
	}

	updated := false
	for line := range t.Lines {
		for _, p := range i.progressKeywords {
			var lineProgress, nextProgress int
			if strings.Contains(line.Text, p.KeyWord) {
				lineProgress = p.ProgressNum
			} else {
				lineProgress, nextProgress = parseImagePrepareProgressByItemProgress(line.Text)
			}
			if i.progress < lineProgress {
				i.progress = lineProgress
				updated = true
			} else if i.progress+1 < nextProgress {
				i.progress += 1
				updated = true
			}
		}
	}

	if updated && i.progressChan != nil {
		i.progressChan <- i.progress
	}

	return nil
}

func parseImagePrepareProgressByItemProgress(line string) (int, int) {
	// filter out other item progress lines to avoid confusion
	if !strings.Contains(line, "imported image") {
		return 0, 0
	}
	return parseProgressFromItemProgress(line)
}
