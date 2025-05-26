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

package addons

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type AddonsModule struct {
	common.KubeModule
	Skip bool
}

func (a *AddonsModule) IsSkip() bool {
	return a.Skip
}

func (a *AddonsModule) Init() {
	a.Name = "AddonsModule"
	a.Desc = "Install addons"

	install := &task.LocalTask{
		Name:   "InstallAddons",
		Desc:   "Install addons",
		Action: new(Install),
	}

	a.Tasks = []task.Interface{
		install,
	}
}
