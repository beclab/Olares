/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package ending

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/pkg/errors"
)

type TaskResult struct {
	mu            sync.Mutex
	ActionResults []*ActionResult
	Status        ResultStatus
	StartTime     time.Time
	EndTime       time.Time
}

func NewTaskResult() *TaskResult {
	return &TaskResult{ActionResults: make([]*ActionResult, 0, 0), Status: NULL, StartTime: time.Now()}
}

func (t *TaskResult) ErrResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = FAILED
}

func (t *TaskResult) NormalResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SUCCESS
}

func (t *TaskResult) SkippedResult() {
	if t.Status != NULL {
		return
	}
	t.EndTime = time.Now()
	t.Status = SKIPPED
}

func (t *TaskResult) AppendSkip(host connector.Host) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	e := &ActionResult{
		Host:      host,
		Status:    SKIPPED,
		Error:     nil,
		StartTime: t.StartTime,
		EndTime:   now,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = now
}

func (t *TaskResult) AppendSuccess(host connector.Host) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	e := &ActionResult{
		Host:      host,
		Status:    SUCCESS,
		Error:     nil,
		StartTime: t.StartTime,
		EndTime:   now,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = now
}

func (t *TaskResult) AppendErr(host connector.Host, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	e := &ActionResult{
		Host:      host,
		Status:    FAILED,
		Error:     err,
		StartTime: t.StartTime,
		EndTime:   now,
	}

	t.ActionResults = append(t.ActionResults, e)
	t.EndTime = now
	t.Status = FAILED
}

func (t *TaskResult) IsFailed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Status == FAILED {
		return true
	}
	return false
}

// CombineErr reports every host this task failed on.
//
// This is the one layer that knows which host each result belongs to, so it owns
// naming the host. The errors reaching it do not carry that themselves: a
// timeout, a failed prepare condition or a runtime setup failure are all
// host-agnostic by the time they are recorded.
func (t *TaskResult) CombineErr() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var msgs []string
	for i := range t.ActionResults {
		result := t.ActionResults[i]
		if result.Status != FAILED {
			continue
		}
		// A task rejected before a host was picked (an action that is nil, for
		// instance) is recorded without one, and must not panic here.
		host := "unknown host"
		if result.Host != nil {
			host = result.Host.GetName()
		}
		err := "unknown error"
		if result.Error != nil {
			err = result.Error.Error()
		}
		msgs = append(msgs, fmt.Sprintf("failed - %s: %s", host, err))
	}
	if len(msgs) == 0 {
		return nil
	}
	return errors.New(strings.Join(msgs, "\n"))
}
