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

package kubesphere

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type DeleteKubeSphereCachesModule struct {
	common.KubeModule
}

func (m *DeleteKubeSphereCachesModule) Init() {
	m.Name = "DeleteKsCache"
	m.Desc = "Delete KubeSphere cache"

	deleteKubeSphereCaches := &task.LocalTask{
		Name:   "DeleteKubeSphereCaches",
		Action: new(DeleteKubeSphereCaches),
	}

	m.Tasks = []task.Interface{
		deleteKubeSphereCaches,
	}
}

type DeployModule struct {
	common.KubeModule
}

func (d *DeployModule) Init() {
	logger.InfoInstallationProgress("Installing kubesphere ...")
	d.Name = "DeployKubeSphereModule"
	d.Desc = "Deploy KubeSphere"

	createNamespace := &task.RemoteTask{
		Name:  "CreateKubeSphereNamespace",
		Desc:  "Create the kubesphere namespace",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CreateNamespace),
		Parallel: false,
	}

	d.Tasks = []task.Interface{
		createNamespace,
	}
}

type CheckResultModule struct {
	common.KubeModule
}

func (c *CheckResultModule) Init() {
	c.Name = "CheckResultModule"
	c.Desc = "Check deploy KubeSphere result"

	check := &task.RemoteTask{
		Name:  "CheckKubeSphereRunning",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(Check),
		Parallel: false,
		Retry:    30,
		Delay:    10 * time.Second,
	}

	getKubeCommand := &task.RemoteTask{
		Name:  "GetKubeCommand",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(GetKubeCommand),
		Parallel: false,
		Retry:    1,
	}

	c.Tasks = []task.Interface{
		check,
		getKubeCommand,
	}
}
