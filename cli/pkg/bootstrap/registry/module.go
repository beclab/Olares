/*
 Copyright 2022 The KubeSphere Authors.

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

package registry

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/bootstrap/registry/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	docker_template "bytetrade.io/web3os/installer/pkg/container/templates"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type RegistryCertsModule struct {
	common.KubeModule
	Skip bool
}

func (p *RegistryCertsModule) IsSkip() bool {
	return p.Skip
}

func (i *RegistryCertsModule) Init() {
	i.Name = "InitRegistryModule"
	i.Desc = "Init a local registry"

	fetchCerts := &task.RemoteTask{
		Name:     "FetchRegistryCerts",
		Desc:     "Fetch registry certs",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Prepare:  new(FirstRegistryNode),
		Action:   new(FetchCerts),
		Parallel: false,
	}

	generateCerts := &task.LocalTask{
		Name:   "GenerateRegistryCerts",
		Desc:   "Generate registry Certs",
		Action: new(GenerateCerts),
	}

	syncCertsFile := &task.RemoteTask{
		Name:     "SyncCertsFile",
		Desc:     "Synchronize certs file",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToAllNodes := &task.RemoteTask{
		Name:     "SyncCertsFileToAllNodes",
		Desc:     "Synchronize certs file to all nodes",
		Hosts:    i.Runtime.GetAllHosts(),
		Action:   new(SyncCertsToAllNodes),
		Parallel: true,
		Retry:    1,
	}

	i.Tasks = []task.Interface{
		fetchCerts,
		generateCerts,
		syncCertsFile,
		syncCertsToAllNodes,
	}

}

type InstallRegistryModule struct {
	common.KubeModule
}

func (i *InstallRegistryModule) Init() {
	i.Name = "InstallRegistryModule"
	i.Desc = "Install local registry"

	switch i.KubeConf.Cluster.Registry.Type {
	case common.Harbor:
		i.Tasks = InstallHarbor(i)
	default:
		i.Tasks = InstallRegistry(i)
	}
}

func InstallRegistry(i *InstallRegistryModule) []task.Interface {
	installRegistryBinary := &task.RemoteTask{
		Name:     "InstallRegistryBinary",
		Desc:     "Install local registry",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(InstallRegistryBinary),
		Parallel: true,
		Retry:    1,
	}

	generateRegistryService := &task.RemoteTask{
		Name:  "GenerateRegistryService",
		Desc:  "Generate registry service",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Action: &action.Template{
			Name:     "GenerateRegistryService",
			Template: templates.RegistryServiceTempl,
			Dst:      "/etc/systemd/system/registry.service",
		},
		Parallel: true,
		Retry:    1,
	}

	generateRegistryConfig := &task.RemoteTask{
		Name:  "GenerateRegistryConfig",
		Desc:  "Generate registry config",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Action: &action.Template{
			Name:     "GenerateRegistryConfig",
			Template: templates.RegistryConfigTempl,
			Dst:      "/etc/kubekey/registry/config.yaml",
			Data: util.Data{
				"Certificate": fmt.Sprintf("%s.pem", RegistryCertificateBaseName),
				"Key":         fmt.Sprintf("%s-key.pem", RegistryCertificateBaseName),
			},
		},
		Parallel: true,
		Retry:    1,
	}

	startRgistryService := &task.RemoteTask{
		Name:     "StartRegistryService",
		Desc:     "Start registry service",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(StartRegistryService),
		Parallel: true,
		Retry:    1,
	}

	return []task.Interface{
		installRegistryBinary,
		generateRegistryService,
		generateRegistryConfig,
		startRgistryService,
	}
}

func InstallHarbor(i *InstallRegistryModule) []task.Interface {
	// Install docker
	// syncBinaries := &task.RemoteTask{
	// 	Name:  "SyncDockerBinaries",
	// 	Desc:  "Sync docker binaries",
	// 	Hosts: i.Runtime.GetHostsByRole(common.Registry),
	// 	Prepare: &prepare.PrepareCollection{
	// 		&container.DockerExist{Not: true},
	// 	},
	// 	Action:   new(container.SyncDockerBinaries),
	// 	Parallel: true,
	// 	Retry:    2,
	// }

	generateDockerService := &task.RemoteTask{
		Name:  "GenerateDockerService",
		Desc:  "Generate docker service",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Prepare: &prepare.PrepareCollection{
			&container.DockerExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateDockerService",
			Template: docker_template.DockerService,
			Dst:      filepath.Join("/etc/systemd/system", docker_template.DockerService.Name()),
		},
		Parallel: true,
	}

	generateDockerConfig := &task.RemoteTask{
		Name:  "GenerateDockerConfig",
		Desc:  "Generate docker config",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Prepare: &prepare.PrepareCollection{
			&container.DockerExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateDockerConfig",
			Template: docker_template.DockerConfig,
			Dst:      filepath.Join("/etc/docker/", docker_template.DockerConfig.Name()),
			Data: util.Data{
				"Mirrors":            docker_template.Mirrors(i.KubeConf),
				"InsecureRegistries": docker_template.InsecureRegistries(i.KubeConf),
			},
		},
		Parallel: true,
	}

	// enableDocker := &task.RemoteTask{
	// 	Name:  "EnableDocker",
	// 	Desc:  "Enable docker",
	// 	Hosts: i.Runtime.GetHostsByRole(common.Registry),
	// 	Prepare: &prepare.PrepareCollection{
	// 		&container.DockerExist{Not: true},
	// 	},
	// 	Action:   new(container.EnableDocker),
	// 	Parallel: true,
	// }

	// Install docker compose
	installDockerCompose := &task.RemoteTask{
		Name:     "InstallDockerCompose",
		Desc:     "Install docker compose",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(InstallDockerCompose),
		Parallel: true,
		Retry:    2,
	}

	// Install Harbor
	syncHarborPackage := &task.RemoteTask{
		Name:     "SyncHarborPackage",
		Desc:     "Sync harbor package",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(SyncHarborPackage),
		Parallel: true,
		Retry:    2,
	}

	// generate Harbor Systemd
	generateHarborService := &task.RemoteTask{
		Name:  "GenerateHarborService",
		Desc:  "Generate harbor service",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Action: &action.Template{
			Name:     "GenerateHarborService",
			Template: templates.HarborServiceTempl,
			Dst:      "/etc/systemd/system/harbor.service",
			Data: util.Data{
				"Harbor_install_path": "/opt/harbor",
			},
		},
		Parallel: true,
		Retry:    1,
	}

	generateHarborConfig := &task.RemoteTask{
		Name:  "GenerateHarborConfig",
		Desc:  "Generate harbor config",
		Hosts: i.Runtime.GetHostsByRole(common.Registry),
		Action: &action.Template{
			Name:     "GenerateHarborConfig",
			Template: templates.HarborConfigTempl,
			Dst:      "/opt/harbor/harbor.yml",
			Data: util.Data{
				"Domain":      RegistryCertificateBaseName,
				"Certificate": fmt.Sprintf("%s.pem", RegistryCertificateBaseName),
				"Key":         fmt.Sprintf("%s-key.pem", RegistryCertificateBaseName),
			},
		},
		Parallel: true,
		Retry:    1,
	}

	startHarbor := &task.RemoteTask{
		Name:     "StartHarbor",
		Desc:     "start harbor",
		Hosts:    i.Runtime.GetHostsByRole(common.Registry),
		Action:   new(StartHarbor),
		Parallel: true,
		Retry:    2,
	}

	return []task.Interface{
		// syncBinaries,
		generateDockerService,
		generateDockerConfig,
		// enableDocker,
		installDockerCompose,
		syncHarborPackage,
		generateHarborService,
		generateHarborConfig,
		startHarbor,
	}
}
