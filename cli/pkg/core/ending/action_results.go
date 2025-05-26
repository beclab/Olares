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

	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type ActionResult struct {
	Host      connector.Host
	Status    ResultStatus
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

func (a *ActionResult) GetHost() connector.Host {
	return a.Host
}

func (a *ActionResult) GetStatus() ResultStatus {
	return a.Status
}

func (a *ActionResult) GetErr() error {
	return a.Error
}

func (a *ActionResult) GetStartTime() time.Time {
	return a.StartTime
}

func (a *ActionResult) GetEndTime() time.Time {
	return a.EndTime
}
