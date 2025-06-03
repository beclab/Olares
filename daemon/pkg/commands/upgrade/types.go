package upgrade

import (
	"math"
	"regexp"
	"strconv"
)

type ExecutionRes interface {
	Finished() bool
	Progress() <-chan int
}

type executionRes struct {
	finished     bool
	progressChan <-chan int
}

func (r *executionRes) Finished() bool {
	return r.finished
}

func (r *executionRes) Progress() <-chan int {
	return r.progressChan
}

func newExecutionRes(finished bool, progressChan <-chan int) ExecutionRes {
	return &executionRes{
		finished:     finished,
		progressChan: progressChan,
	}
}

type progressKeyword struct {
	KeyWord     string
	ProgressNum int
}

// matches against "(x/y)"
var itemProcessProgressRE = regexp.MustCompile(`\((\d+)/(\d+)\)`)

func parseProgressFromItemProgress(line string) int {
	matches := itemProcessProgressRE.FindAllStringSubmatch(line, 2)
	if len(matches) != 3 {
		return 0
	}
	indexStr, totalStr := matches[0][1], matches[0][2]
	index, err := strconv.ParseFloat(indexStr, 64)
	if index == 0 || err != nil {
		return 0
	}
	total, err := strconv.ParseFloat(totalStr, 64)
	if total == 0 || err != nil {
		return 0
	}
	return int(math.Round(index / total * 90))
}
