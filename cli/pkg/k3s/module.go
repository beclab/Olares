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

package k3s

import (
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"path/filepath"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	containertemplates "bytetrade.io/web3os/installer/pkg/container/templates"
	"bytetrade.io/web3os/installer/pkg/core/action"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s/templates"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/registry"
)

type InstallContainerModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (i *InstallContainerModule) IsSkip() bool {
	return i.Skip
}

func (i *InstallContainerModule) Init() {
	i.Name = "InstallContainerModule(k3s)"
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
		Name:  "ZfsMountReset",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
			&container.ZfsResetPrepare{},
		},
		Action:   new(container.ZfsReset),
		Parallel: false,
		Retry:    5,
		Delay:    5 * time.Second,
	}

	createZfsMount := &task.RemoteTask{
		Name:  "CreateZfsMount",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
			&container.ZfsResetPrepare{},
		},
		Action:   new(container.CreateZfsMount),
		Parallel: false,
		Retry:    1,
	}

	syncContainerd := &task.RemoteTask{
		Name:  "SyncContainerd",
		Desc:  "Sync containerd binaries",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
		},
		Action: &container.SyncContainerd{
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
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.CrictlExist{Not: true},
		},
		Action: &container.SyncCrictlBinaries{
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
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateContainerdService",
			Template: containertemplates.ContainerdService,
			Dst:      filepath.Join("/etc/systemd/system", containertemplates.ContainerdService.Name()),
		},
		Parallel: true,
	}

	generateContainerdConfig := &task.RemoteTask{
		Name:  "GenerateContainerdConfig",
		Desc:  "Generate containerd config",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateContainerdConfig",
			Template: containertemplates.ContainerdConfig,
			Dst:      filepath.Join("/etc/containerd/", containertemplates.ContainerdConfig.Name()),
			Data: util.Data{
				"Mirrors":            containertemplates.Mirrors(m.KubeConf),
				"InsecureRegistries": m.KubeConf.Cluster.Registry.InsecureRegistries,
				"SandBoxImage":       images.GetImage(m.Runtime, m.KubeConf, "pause").ImageName(),
				"Auths":              registry.DockerRegistryAuthEntries(m.KubeConf.Cluster.Registry.Auths),
				"DataRoot":           containertemplates.DataRoot(m.KubeConf),
				"FsType":             m.KubeConf.Arg.SystemInfo.GetFsType(),
				"ZfsRootPath":        cc.ZfsSnapshotter,
			},
		},
		Parallel: true,
	}

	generateCrictlConfig := &task.RemoteTask{
		Name:  "GenerateCrictlConfig",
		Desc:  "Generate crictl config",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateCrictlConfig",
			Template: containertemplates.CrictlConfig,
			Dst:      filepath.Join("/etc/", containertemplates.CrictlConfig.Name()),
			Data: util.Data{
				"Endpoint": m.KubeConf.Cluster.Kubernetes.ContainerRuntimeEndpoint,
			},
		},
		Parallel: true,
	}

	enableContainerd := &task.RemoteTask{
		Name:  "EnableContainerd",
		Desc:  "Enable containerd",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&container.ContainerdExist{Not: true},
		},
		Action: &container.EnableContainerd{
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

type StatusModule struct {
	common.KubeModule
}

func (s *StatusModule) Init() {
	s.Name = "StatusModule"
	s.Desc = "Get cluster status"

	cluster := NewK3sStatus()
	s.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:     "GetClusterStatus(k3s)",
		Desc:     "Get k3s cluster status",
		Hosts:    s.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	s.Tasks = []task.Interface{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"
	i.Desc = "Install k3s cluster"

	syncBinary := &task.RemoteTask{
		Name:    "SyncKubeBinary(k3s)",
		Desc:    "Synchronize k3s binaries",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &kubernetes.NodeInCluster{Not: true},
		Action: &SyncKubeBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  i.BaseDir,
				Manifest: i.Manifest,
			},
		},
		Parallel: true,
		Retry:    2,
	}

	killAllScript := &task.RemoteTask{
		Name:    "GenerateK3sKillAllScript",
		Desc:    "Generate k3s killall.sh script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &kubernetes.NodeInCluster{Not: true},
		Action: &action.Template{
			Name:     "GenerateK3sKillAllScript",
			Template: templates.K3sKillallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sKillallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	uninstallScript := &task.RemoteTask{
		Name:    "GenerateK3sUninstallScript",
		Desc:    "Generate k3s uninstall script",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &kubernetes.NodeInCluster{Not: true},
		Action: &action.Template{
			Name:     "GenerateK3sUninstallScript",
			Template: templates.K3sUninstallScript,
			Dst:      filepath.Join("/usr/local/bin", templates.K3sUninstallScript.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	chmod := &task.RemoteTask{
		Name:     "ChmodScript(k3s)",
		Desc:     "Chmod +x k3s script ",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(ChmodScript),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []task.Interface{
		syncBinary,
		killAllScript,
		uninstallScript,
		chmod,
	}
}

type InitClusterModule struct {
	common.KubeModule
}

func (i *InitClusterModule) Init() {
	i.Name = "K3sInitClusterModule"
	i.Desc = "Init k3s cluster"

	k3sService := &task.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: false,
	}

	k3sEnv := &task.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: false,
	}

	k3sRegistryConfig := &task.RemoteTask{
		Name:  "GenerateK3sRegistryConfig",
		Desc:  "Generate k3s registry config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			//&UsePrivateRegstry{Not: false},
		},
		Action:   new(GenerateK3sRegistryConfig),
		Parallel: false,
	}

	enableK3s := &task.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: false,
	}

	// preload image
	// preloadImages := &task.RemoteTask{
	// 	Name:  "PreloadImagesService",
	// 	Desc:  "Preload Images",
	// 	Hosts: i.Runtime.GetHostsByRole(common.Master),
	// 	Prepare: &prepare.PrepareCollection{
	// 		new(common.OnlyFirstMaster),
	// 		&ClusterIsExist{Not: true},
	// 	},
	// 	Action:   new(PreloadImagesService), // ! herre
	// 	Parallel: false,
	// 	Retry:    1,
	// }

	copyKubeConfig := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
		Parallel: true,
	}

	addMasterTaint := &task.RemoteTask{
		Name:  "AddMasterTaint(k3s)",
		Desc:  "Add master taint",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    200,
		Delay:    10 * time.Second,
	}

	addWorkerLabel := &task.RemoteTask{
		Name:  "AddWorkerLabel(k3s)",
		Desc:  "Add worker label",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    200,
		Delay:    10 * time.Second,
	}

	i.Tasks = []task.Interface{
		k3sService,
		k3sEnv,
		k3sRegistryConfig,
		enableK3s,
		// preloadImages,
		copyKubeConfig,
		addMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "K3sJoinNodesModule"
	j.Desc = "Join k3s nodes"

	k3sService := &task.RemoteTask{
		Name:  "GenerateK3sService",
		Desc:  "Generate k3s Service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sService),
		Parallel: true,
	}

	k3sEnv := &task.RemoteTask{
		Name:  "GenerateK3sServiceEnv",
		Desc:  "Generate k3s service env",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
		},
		Action:   new(GenerateK3sServiceEnv),
		Parallel: true,
	}

	k3sRegistryConfig := &task.RemoteTask{
		Name:  "GenerateK3sRegistryConfig",
		Desc:  "Generate k3s registry config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&UsePrivateRegstry{Not: false},
		},
		Action:   new(GenerateK3sRegistryConfig),
		Parallel: true,
	}

	enableK3s := &task.RemoteTask{
		Name:  "EnableK3sService",
		Desc:  "Enable k3s service",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
		},
		Action:   new(EnableK3sService),
		Parallel: false,
	}

	copyKubeConfigForMaster := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy k3s.yaml to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
		},
		Action:   new(CopyK3sKubeConfig),
		Parallel: true,
	}

	syncKubeConfigToWorker := &task.RemoteTask{
		Name:  "SyncKubeConfigToWorker(k3s)",
		Desc:  "Synchronize kube config to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
	}

	addMasterTaint := &task.RemoteTask{
		Name:  "AddMasterTaint(k3s)",
		Desc:  "Add master taint",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			&common.IsWorker{Not: true},
		},
		Action:   new(AddMasterTaint),
		Parallel: true,
		Retry:    200,
		Delay:    10 * time.Second,
	}

	addWorkerLabel := &task.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&kubernetes.NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    200,
		Delay:    10 * time.Second,
	}

	j.Tasks = []task.Interface{
		k3sService,
		k3sEnv,
		k3sRegistryConfig,
		enableK3s,
		copyKubeConfigForMaster,
		syncKubeConfigToWorker,
		addMasterTaint,
		addWorkerLabel,
	}
}

type DeleteClusterModule struct {
	common.KubeModule
}

func (d *DeleteClusterModule) Init() {
	d.Name = "DeleteClusterModule"
	d.Desc = "Delete k3s cluster"

	deleteCurrentNode := &task.LocalTask{
		Name:   "DeleteCurrentNode",
		Action: new(kubernetes.KubectlDeleteCurrentWorkerNode),
	}

	killScript := &task.RemoteTask{
		Name:     "ExecKillAllScript(k3s)",
		Hosts:    d.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(CheckK3sKillAllScript),
		Action:   new(ExecKillAllScript),
		Parallel: false,
	}

	execScript := &task.RemoteTask{
		Name:     "ExecUninstallScript(k3s)",
		Hosts:    d.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(CheckK3sUninstallScript),
		Action:   new(ExecUninstallScript),
		Parallel: false,
	}

	d.Tasks = []task.Interface{
		deleteCurrentNode,
		killScript,
		execScript,
	}
}

type SaveKubeConfigModule struct {
	common.KubeModule
}

func (s *SaveKubeConfigModule) Init() {
	s.Name = "SaveKubeConfigModule"
	s.Desc = "Save kube config file as a configmap"

	save := &task.LocalTask{
		Name:   "SaveKubeConfig(k3s)",
		Desc:   "Save kube config as a configmap",
		Action: new(SaveKubeConfig),
	}

	s.Tasks = []task.Interface{
		save,
	}
}

type UninstallK3sModule struct {
	common.KubeModule
}

func (m *UninstallK3sModule) Init() {
	m.Name = "UninstallK3s"

	uninstallK3s := &task.LocalTask{
		Name:    "UninstallK3s",
		Prepare: new(CheckK3sUninstallScript),
		Action:  new(UninstallK3s),
	}

	deleteCalicoCNI := &task.LocalTask{
		Name:   "DeleteCalicoCNI",
		Action: new(DeleteCalicoCNI),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		uninstallK3s,
		deleteCalicoCNI,
	}
}

func checkContainerExists(runtime connector.Runtime) bool {
	var cmd = "if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; " +
		"then echo 'not exist'; " +
		"fi"
	var runner = runtime.GetRunner()
	if output, err := runner.Host.SudoCmd(cmd, false, false); err != nil || strings.Contains(output, "not exist") {
		return false
	}
	return true
}
