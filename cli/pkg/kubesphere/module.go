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
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/version/kubesphere/templates"

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
	Skip bool
}

func (d *DeployModule) IsSkip() bool {
	return d.Skip
}

func (d *DeployModule) Init() {
	logger.InfoInstallationProgress("Installing kubesphere ...")
	d.Name = "DeployKubeSphereModule"
	d.Desc = "Deploy KubeSphere"

	generateManifests := &task.RemoteTask{
		Name:  "GenerateKsInstallerCRD",
		Desc:  "Generate KubeSphere ks-installer crd manifests",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &action.Template{
			Name:     "GenerateKsInstallerCRD",
			Template: templates.KsInstaller,
			Dst:      filepath.Join(common.KubeAddonsDir, templates.KsInstaller.Name()),
		},
		Parallel: false,
	}

	addConfig := &task.RemoteTask{
		Name:  "AddKsInstallerConfig",
		Desc:  "Add config to ks-installer manifests",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(AddInstallerConfig),
		Parallel: false,
	}

	createNamespace := &task.RemoteTask{
		Name:  "CreateKubeSphereNamespace",
		Desc:  "Create the kubesphere namespace",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(CreateNamespace),
		Parallel: false,
	}

	setup := &task.RemoteTask{
		Name:  "SetupKsInstallerConfig",
		Desc:  "Setup ks-installer config",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(Setup), // todo
		Parallel: false,
		Retry:    1,
	}

	apply := &task.RemoteTask{
		Name:  "ApplyKsInstaller",
		Desc:  "Apply ks-installer",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(Apply),
		Parallel: false,
		Retry:    10,
		Delay:    5 * time.Second,
	}

	d.Tasks = []task.Interface{
		generateManifests,
		// apply crd installer.kubesphere.io/v1alpha1
		// apply,
		addConfig,
		createNamespace,
		setup,
		apply,
	}
}

type CheckResultModule struct {
	common.KubeModule
	Skip bool
}

func (c *CheckResultModule) IsSkip() bool {
	return c.Skip
}

func (c *CheckResultModule) Init() {
	c.Name = "CheckResultModule"
	c.Desc = "Check deploy KubeSphere result"

	check := &task.RemoteTask{
		Name:  "CheckKubeSphereRunning",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
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
			new(NotEqualDesiredVersion),
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
