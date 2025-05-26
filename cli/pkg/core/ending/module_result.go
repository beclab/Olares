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
	"time"

	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type ModuleResult struct {
	HostResults   map[string]Interface
	CombineResult error
	Status        ResultStatus
	StartTime     time.Time
	EndTime       time.Time
}

func NewModuleResult() *ModuleResult {
	return &ModuleResult{HostResults: make(map[string]Interface), StartTime: time.Now(), Status: NULL}
}

func (m *ModuleResult) IsFailed() bool {
	if m.Status == FAILED {
		return true
	}
	return false
}

func (m *ModuleResult) AppendHostResult(p Interface) {
	if m.HostResults == nil {
		return
	}
	m.HostResults[p.GetHost().GetName()] = p
}

func (m *ModuleResult) LocalErrResult(err error) {
	now := time.Now()
	r := &ActionResult{
		Host:      &connector.BaseHost{Name: common.LocalHost},
		Status:    FAILED,
		Error:     err,
		StartTime: m.StartTime,
		EndTime:   now,
	}

	m.HostResults[r.Host.GetName()] = r
	m.CombineResult = err
	m.EndTime = now
	m.Status = FAILED
}

func (m *ModuleResult) ErrResult(combineErr error) {
	m.EndTime = time.Now()
	m.Status = FAILED
	m.CombineResult = combineErr
}

func (m *ModuleResult) NormalResult() {
	m.EndTime = time.Now()
	m.Status = SUCCESS
}
