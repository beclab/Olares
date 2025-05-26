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

package container

import (
	"path/filepath"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/registry"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container/templates"
	"bytetrade.io/web3os/installer/pkg/core/action"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
)

type InstallContainerModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip        bool
	NoneCluster bool
}

func (i *InstallContainerModule) IsSkip() bool {
	return i.Skip
}

func (i *InstallContainerModule) Init() {
	i.Name = "InstallContainerModule(k8s)"
	i.Desc = "Install container manager"

	switch i.KubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Containerd:
		i.Tasks = InstallContainerd(i)
	case common.Crio:
		// TODO: Add the steps of cri-o's installation.
	case common.Isula:
		// TODO: Add the steps of iSula's installation.
	default:
		logger.Fatalf("Unsupported container runtime: %s", strings.TrimSpace(i.KubeConf.Cluster.Kubernetes.ContainerManager))
	}
}

func InstallContainerd(m *InstallContainerModule) []task.Interface {
	fsReset := &task.RemoteTask{
		Name:  "DeleteZfsMount",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
			&ZfsResetPrepare{},
		},
		Action:   new(ZfsReset),
		Parallel: false,
		Retry:    5,
		Delay:    5 * time.Second,
	}

	createZfsMount := &task.RemoteTask{
		Name:  "CreateZfsMount",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
			&ZfsResetPrepare{},
		},
		Action:   new(CreateZfsMount),
		Parallel: false,
		Retry:    1,
	}

	syncContainerd := &task.RemoteTask{
		Name:  "SyncContainerd",
		Desc:  "Sync containerd binaries",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
		},
		Action: &SyncContainerd{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: true,
		Retry:    2,
	}

	syncCrictlBinaries := &task.RemoteTask{
		Name:  "SyncCrictlBinaries",
		Desc:  "Sync crictl binaries",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&CrictlExist{Not: true},
		},
		Action: &SyncCrictlBinaries{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: true,
		Retry:    2,
	}

	generateContainerdService := &task.RemoteTask{
		Name:  "GenerateContainerdService",
		Desc:  "Generate containerd service",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateContainerdService",
			Template: templates.ContainerdService,
			Dst:      filepath.Join("/etc/systemd/system", templates.ContainerdService.Name()),
		},
		Parallel: true,
	}

	generateContainerdConfig := &task.RemoteTask{
		Name:  "GenerateContainerdConfig",
		Desc:  "Generate containerd config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateContainerdConfig",
			Template: templates.ContainerdConfig,
			Dst:      filepath.Join("/etc/containerd/", templates.ContainerdConfig.Name()),
			Data: util.Data{
				"Mirrors":            templates.Mirrors(m.KubeConf),
				"InsecureRegistries": m.KubeConf.Cluster.Registry.InsecureRegistries,
				"SandBoxImage":       images.GetImage(m.Runtime, m.KubeConf, "pause").ImageName(),
				"Auths":              registry.DockerRegistryAuthEntries(m.KubeConf.Cluster.Registry.Auths),
				"DataRoot":           templates.DataRoot(m.KubeConf),
				"FsType":             m.KubeConf.Arg.SystemInfo.GetFsType(),
				"ZfsRootPath":        cc.ZfsSnapshotter,
			},
		},
		Parallel: true,
	}

	generateCrictlConfig := &task.RemoteTask{
		Name:  "GenerateCrictlConfig",
		Desc:  "Generate crictl config",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateCrictlConfig",
			Template: templates.CrictlConfig,
			Dst:      filepath.Join("/etc/", templates.CrictlConfig.Name()),
			Data: util.Data{
				"Endpoint": m.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
			},
		},
		Parallel: true,
	}

	enableContainerd := &task.RemoteTask{
		Name:  "EnableContainerd",
		Desc:  "Enable containerd",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true, NoneCluster: m.NoneCluster},
			&ContainerdExist{Not: true},
		},
		Action: &EnableContainerd{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: true,
	}

	return []task.Interface{
		fsReset,
		createZfsMount,
		syncContainerd,
		syncCrictlBinaries,
		generateContainerdService,
		generateContainerdConfig,
		generateCrictlConfig,
		enableContainerd,
	}
}

type UninstallContainerModule struct {
	common.KubeModule
	Skip bool
}

func (i *UninstallContainerModule) IsSkip() bool {
	return i.Skip
}

func (i *UninstallContainerModule) Init() {
	i.Name = "UninstallContainerModule"
	i.Desc = "Uninstall container manager"

	switch i.KubeConf.Cluster.Kubernetes.ContainerManager {
	case common.Docker:
		i.Tasks = UninstallDocker(i)
	case common.Containerd:
		i.Tasks = UninstallContainerd(i)
	case common.Crio:
		// TODO: Add the steps of cri-o's installation.
	case common.Isula:
		// TODO: Add the steps of iSula's installation.
	default:
		logger.Fatalf("Unsupported container runtime: %s", strings.TrimSpace(i.KubeConf.Cluster.Kubernetes.ContainerManager))
	}
}

func UninstallDocker(m *UninstallContainerModule) []task.Interface {

	disableDocker := &task.RemoteTask{
		Name:  "DisableDocker",
		Desc:  "Disable docker",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&DockerExist{Not: false},
		},
		Action:   new(DisableDocker),
		Parallel: true,
	}

	return []task.Interface{
		disableDocker,
	}
}

func UninstallContainerd(m *UninstallContainerModule) []task.Interface {

	disableContainerd := &task.RemoteTask{
		Name:     "UninstallContainerd",
		Desc:     "Uninstall containerd",
		Hosts:    m.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DisableContainerd),
		Parallel: false,
	}

	return []task.Interface{
		disableContainerd,
	}
}

type CriMigrateModule struct {
	common.KubeModule

	Skip bool
}

func (i *CriMigrateModule) IsSkip() bool {
	return i.Skip
}

func (p *CriMigrateModule) Init() {
	p.Name = "CriMigrateModule"
	p.Desc = "Cri Migrate manager"

	if p.KubeConf.Arg.Role == common.Worker {
		p.Tasks = MigrateWCri(p)
	} else if p.KubeConf.Arg.Role == common.Master {
		p.Tasks = MigrateMCri(p)
	} else if p.KubeConf.Arg.Role == "all" {
		p.Tasks = MigrateACri(p)
	} else {
		logger.Fatalf("Unsupported Role: %s", strings.TrimSpace(p.KubeConf.Arg.Role))
	}
}

func MigrateWCri(p *CriMigrateModule) []task.Interface {

	MigrateWCri := &task.RemoteTask{
		Name:     "MigrateToDocker",
		Desc:     "Migrate To Docker",
		Hosts:    p.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(common.OnlyWorker),
		Action:   new(MigrateSelfNodeCri),
		Parallel: false,
	}

	p.Tasks = []task.Interface{
		MigrateWCri,
	}

	return p.Tasks
}

func MigrateMCri(p *CriMigrateModule) []task.Interface {

	MigrateMCri := &task.RemoteTask{
		Name:     "MigrateMasterToDocker",
		Desc:     "Migrate Master To Docker",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   new(MigrateSelfNodeCri),
		Parallel: false,
	}

	p.Tasks = []task.Interface{
		MigrateMCri,
	}

	return p.Tasks
}

func MigrateACri(p *CriMigrateModule) []task.Interface {

	MigrateACri := &task.RemoteTask{
		Name:     "MigrateMasterToDocker",
		Desc:     "Migrate Master To Docker",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Action:   new(MigrateSelfNodeCri),
		Parallel: false,
	}

	p.Tasks = []task.Interface{
		MigrateACri,
	}

	return p.Tasks
}

type DeleteZfsMountModule struct {
	common.KubeModule
	Skip bool
}

func (i *DeleteZfsMountModule) IsSkip() bool {
	return i.Skip
}

func (m *DeleteZfsMountModule) Init() {
	m.Name = "DeleteZfsMount"

	zfsReset := &task.RemoteTask{
		Name:  "DeleteZfsMount",
		Hosts: m.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			new(ZfsResetPrepare),
		},
		Action:   new(ZfsReset),
		Parallel: false,
		Retry:    5,
		Delay:    5 * time.Second,
	}

	m.Tasks = []task.Interface{
		zfsReset,
	}
}

type KillContainerdProcessModule struct {
	common.KubeModule
}

func (m *KillContainerdProcessModule) Init() {
	m.Name = "KillContainerdProcess"

	killContainerdProcess := &task.RemoteTask{
		Name:     "KillContainerdProcess",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   &KillContainerdProcess{Signal: "KILL"},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		killContainerdProcess,
	}

}
