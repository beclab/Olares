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

package certs

import (
	"path/filepath"

	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/beclab/Olares/cli/pkg/certs/templates"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
)

type AutoRenewCertsModule struct {
	common.KubeModule
	Skip bool
}

func (a *AutoRenewCertsModule) IsSkip() bool {
	return a.Skip
}

func (a *AutoRenewCertsModule) Init() {
	a.Name = "AutoRenewCertsModule"
	a.Desc = "Install auto renew control-plane certs"

	generateK8sCertsRenewScript := &task.RemoteTask{
		Name:  "GenerateK8sCertsRenewScript",
		Desc:  "Generate k8s certs renew script",
		Hosts: a.Runtime.GetHostsByRole(common.Master),
		Action: &action.Template{
			Name:     "GenerateK8sCertsRenewScript",
			Template: templates.K8sCertsRenewScript,
			Dst:      filepath.Join("/usr/local/bin/kube-scripts/", templates.K8sCertsRenewScript.Name()),
			Data: util.Data{
				"IsDocker":            a.KubeConf.Cluster.Kubernetes.ContainerManager == common.Docker,
				"IsKubeadmAlphaCerts": versionutil.MustParseSemantic(a.KubeConf.Cluster.Kubernetes.Version).LessThan(versionutil.MustParseGeneric("v1.20.0")),
			},
		},
		Parallel: true,
	}

	generateK8sCertsRenewService := &task.RemoteTask{
		Name:  "GenerateK8sCertsRenewService",
		Desc:  "Generate k8s certs renew service",
		Hosts: a.Runtime.GetHostsByRole(common.Master),
		Action: &action.Template{
			Template: templates.K8sCertsRenewService,
			Dst:      filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewService.Name()),
		},
		Parallel: true,
	}

	generateK8sCertsRenewTimer := &task.RemoteTask{
		Name:  "GenerateK8sCertsRenewTimer",
		Desc:  "Generate k8s certs renew timer",
		Hosts: a.Runtime.GetHostsByRole(common.Master),
		Action: &action.Template{
			Name:     "GenerateK8sCertsRenewTimer",
			Template: templates.K8sCertsRenewTimer,
			Dst:      filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewTimer.Name()),
		},
		Parallel: true,
	}

	enable := &task.RemoteTask{
		Name:     "EnableK8sCertsRenewService",
		Desc:     "Enable k8s certs renew service",
		Hosts:    a.Runtime.GetHostsByRole(common.Master),
		Action:   new(EnableRenewService),
		Parallel: true,
	}

	a.Tasks = []task.Interface{
		generateK8sCertsRenewScript,
		generateK8sCertsRenewService,
		generateK8sCertsRenewTimer,
		enable,
	}
}

type UninstallAutoRenewCertsModule struct {
	common.KubeModule
}

func (u *UninstallAutoRenewCertsModule) Init() {
	u.Name = "UninstallAutoRenewCertsModule"
	u.Desc = "UnInstall auto renew control-plane certs"

	uninstall := &task.RemoteTask{
		Name:     "UnInstallAutoRenewCerts",
		Desc:     "UnInstall auto renew control-plane certs",
		Hosts:    u.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(AutoRenewCertsEnabled),
		Action:   new(UninstallAutoRenewCerts),
		Parallel: true,
	}

	u.Tasks = []task.Interface{
		uninstall,
	}
}

type UninstallCertsFilesModule struct {
	common.KubeModule
}

func (m *UninstallCertsFilesModule) Init() {
	m.Name = "UninstallCertsFilesModule"

	uninstall := &task.RemoteTask{
		Name:     "UninstallCertsFiles",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(UninstallCertsFiles),
		Parallel: true,
	}

	m.Tasks = []task.Interface{
		uninstall,
	}
}
