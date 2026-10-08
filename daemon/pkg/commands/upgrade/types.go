package upgrade

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"

	"github.com/beclab/Olares/daemon/pkg/commands"
	"github.com/distribution/distribution/v3/manifest/ocischema"
)

type ExecutionRes interface {
	Finished() bool
	Progress() <-chan int
	Completion() <-chan error
}

type executionRes struct {
	finished     bool
	progressChan <-chan int
	completion   <-chan error
}

func (r *executionRes) Finished() bool {
	return r.finished
}

func (r *executionRes) Progress() <-chan int {
	return r.progressChan
}

func (r *executionRes) Completion() <-chan error {
	return r.completion
}

func newExecutionRes(finished bool, progressChan <-chan int) ExecutionRes {
	return &executionRes{
		finished:     finished,
		progressChan: progressChan,
	}
}

func newExecutionResWithCompletion(progressChan <-chan int, completion <-chan error) ExecutionRes {
	return &executionRes{progressChan: progressChan, completion: completion}
}

// AwaitExecution requires both the success log marker and a successful result
// when the phase provides a completion channel. Already-finished phases and
// legacy phases without a completion channel retain their progress behavior.
func AwaitExecution(ctx context.Context, res ExecutionRes, onProgress func(int)) error {
	if res.Finished() {
		return nil
	}
	progress := res.Progress()
	completion := res.Completion()
	var latest int
	for {
		if progress == nil && completion == nil {
			if latest >= commands.ProgressNumFinished {
				return nil
			}
			return errors.New("command execution did not succeed")
		}
		if latest >= commands.ProgressNumFinished && completion == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case p, ok := <-progress:
			if !ok {
				progress = nil
				continue
			}
			if p > latest {
				latest = p
				if onProgress != nil {
					onProgress(p)
				}
			}
		case err, ok := <-completion:
			completion = nil
			if !ok {
				return errors.New("command exited without a result")
			}
			if err != nil {
				return err
			}
		}
	}
}

func NewExecutionRes(finished bool, progressChan <-chan int) ExecutionRes {
	return newExecutionRes(finished, progressChan)
}

type progressKeyword struct {
	KeyWord     string
	ProgressNum int
}

// matches against "(x/y)"
var itemProcessProgressRE = regexp.MustCompile(`\((\d+)/(\d+)\)`)

func parseProgressFromItemProgress(line string) (int, int) {
	matches := itemProcessProgressRE.FindAllStringSubmatch(line, 2)
	if len(matches) != 1 || len(matches[0]) != 3 {
		return 0, 0
	}
	indexStr, totalStr := matches[0][1], matches[0][2]
	index, err := strconv.ParseFloat(indexStr, 64)
	if index == 0 || err != nil {
		return 0, 0
	}
	total, err := strconv.ParseFloat(totalStr, 64)
	if total == 0 || err != nil {
		return 0, 0
	}
	cur := int(math.Round(index / total * 90))
	var next int
	if index < total {
		next = int(math.Round((index + 1) / total * 90))
	} else {
		next = 99
	}
	return cur, next
}

type manifestComponent struct {
	Type     string             `json:"type"`
	Path     string             `json:"path"`
	FileID   string             `json:"fileid"`
	Size     uint64             `json:"size"`
	Manifest ocischema.Manifest `json:"manifest,omitempty"`
}
