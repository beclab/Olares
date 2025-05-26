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

package confirm

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallConfirmModule struct {
	common.KubeModule
	Skip bool
}

func (i *InstallConfirmModule) IsSkip() bool {
	return i.Skip
}

func (i *InstallConfirmModule) Init() {
	i.Name = "ConfirmModule"
	i.Desc = "Display confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(InstallationConfirm),
	}

	i.Tasks = []task.Interface{
		display,
	}
}

type DeleteClusterConfirmModule struct {
	common.KubeModule
}

func (d *DeleteClusterConfirmModule) Init() {
	d.Name = "DeleteClusterConfirmModule"
	d.Desc = "Display delete confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "cluster"},
	}

	d.Tasks = []task.Interface{
		display,
	}
}

type DeleteNodeConfirmModule struct {
	common.KubeModule
}

func (d *DeleteNodeConfirmModule) Init() {
	d.Name = "DeleteNodeConfirmModule"
	d.Desc = "Display delete node confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &DeleteConfirm{Content: "node"},
	}

	d.Tasks = []task.Interface{
		display,
	}
}

type UpgradeConfirmModule struct {
	common.KubeModule
	Skip bool
}

func (u *UpgradeConfirmModule) IsSkip() bool {
	return u.Skip
}

func (u *UpgradeConfirmModule) Init() {
	u.Name = "UpgradeConfirmModule"
	u.Desc = "Display upgrade confirmation form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: new(UpgradeConfirm),
	}

	u.Tasks = []task.Interface{
		display,
	}
}

type CheckFileExistModule struct {
	module.BaseTaskModule
	FileName string
}

func (c *CheckFileExistModule) Init() {
	c.Name = "CheckFileExist"
	c.Desc = "Check file if is existed"

	check := &task.LocalTask{
		Name:   "CheckExist",
		Desc:   "Check output file if existed",
		Action: &CheckFile{FileName: c.FileName},
	}

	c.Tasks = []task.Interface{
		check,
	}
}

type MigrateCriConfirmModule struct {
	common.KubeModule
}

func (d *MigrateCriConfirmModule) Init() {
	d.Name = "MigrateCriConfirmModule"
	d.Desc = "Display Migrate Cri form"

	display := &task.LocalTask{
		Name:   "ConfirmForm",
		Desc:   "Display confirmation form",
		Action: &MigrateCri{},
	}

	d.Tasks = []task.Interface{
		display,
	}

}
