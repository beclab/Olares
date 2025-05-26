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

package os

import (
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type PvePatchModule struct {
	common.KubeModule
	Skip bool
}

func (p *PvePatchModule) IsSkip() bool {
	return p.Skip
}

func (p *PvePatchModule) Init() {
	p.Name = "PvePatch"

	removePveCNDomain := &task.LocalTask{
		Name:    "RemovePveCNDomain",
		Action:  new(RemoveCNDomain),
		Prepare: new(IsPve),
		Retry:   1,
	}

	pveUpdateSourceCheck := &task.LocalTask{
		Name:    "PveAptUpdateSourceCheck",
		Prepare: new(IsPve),
		Action:  new(PveAptUpdateSourceCheck),
	}

	patchLxcInitScript := &task.LocalTask{
		Name:    "PatchLxcInitScript",
		Action:  new(PatchLxcInitScript),
		Prepare: new(IsPveLxc),
		Retry:   1,
	}

	patchLxcEnvVars := &task.LocalTask{
		Name:    "PatchLxcEnvVars",
		Action:  new(PatchLxcEnvVars),
		Prepare: new(IsPveLxc),
		Retry:   1,
	}

	p.Tasks = []task.Interface{
		removePveCNDomain,
		pveUpdateSourceCheck,
		patchLxcInitScript,
		patchLxcEnvVars,
	}

}

type ConfigSystemModule struct {
	common.KubeModule
}

func (c *ConfigSystemModule) Init() {
	c.Name = "ConfigSystem"

	updateNtpDateTask := &task.RemoteTask{
		Name:     "UpdateNtpDate",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(UpdateNtpDateTask),
		Parallel: false,
		Retry:    1,
	}

	timeSyncTask := &task.RemoteTask{
		Name:     "TimeSync",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(TimeSyncTask),
		Parallel: false,
		Retry:    0,
	}

	configProxyTask := &task.RemoteTask{
		Name:     "ConfigProxy",
		Hosts:    c.Runtime.GetAllHosts(),
		Action:   new(ConfigProxyTask),
		Parallel: false,
		Retry:    0,
	}

	c.Tasks = []task.Interface{
		updateNtpDateTask,
		timeSyncTask,
		configProxyTask,
	}
}

type ConfigureOSModule struct {
	common.KubeModule
}

func (c *ConfigureOSModule) Init() {
	c.Name = "ConfigureOSModule"
	c.Desc = "Init os dependencies"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(GetOSData),
		Parallel: true,
	}

	initOS := &task.RemoteTask{
		Name:     "InitOS",
		Desc:     "Prepare to init OS",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(NodeConfigureOS),
		Parallel: true,
	}

	GenerateScript := &task.RemoteTask{
		Name:    "GenerateScript",
		Desc:    "Generate init os script",
		Hosts:   c.Runtime.GetAllHosts(),
		Prepare: &kubernetes.NodeInCluster{Not: true},
		Action: &action.Template{
			Name:     "GenerateScript",
			Template: templates.InitOsScriptTmpl,
			Dst:      filepath.Join(common.KubeScriptDir, "initOS.sh"),
			Data: util.Data{
				"Hosts": templates.GenerateHosts(c.Runtime, c.KubeConf),
			},
		},
		Parallel: true,
	}

	ExecScript := &task.RemoteTask{
		Name:     "ExecScript",
		Desc:     "Exec init os script",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(NodeExecScript),
		Parallel: true,
	}

	ConfigureNtpServer := &task.RemoteTask{
		Name:  "ConfigureNtpServer",
		Desc:  "configure the ntp server for each node",
		Hosts: c.Runtime.GetAllHosts(),
		Prepare: &prepare.PrepareCollection{
			new(NodeConfigureNtpCheck),
			&kubernetes.NodeInCluster{Not: true},
		},
		Action:   new(NodeConfigureNtpServer),
		Parallel: true,
	}

	configureSwap := &task.RemoteTask{
		Name:     "ConfigureSwap",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(ConfigureSwapTask),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		getOSData,
		initOS,
		GenerateScript,
		ExecScript,
		ConfigureNtpServer,
		configureSwap,
	}
}

type ClearNodeOSModule struct {
	common.KubeModule
}

func (c *ClearNodeOSModule) Init() {
	c.Name = "ClearNodeOSModule"

	resetNetworkConfig := &task.RemoteTask{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	removeFiles := &task.RemoteTask{
		Name:     "RemoveFiles",
		Desc:     "Remove node files",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(RemoveNodeFiles),
		Parallel: true,
	}

	daemonReload := &task.RemoteTask{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.Worker),
		Prepare:  new(DeleteNode),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		resetNetworkConfig,
		removeFiles,
		daemonReload,
	}
}

type ClearOSEnvironmentModule struct {
	common.KubeModule
}

func (c *ClearOSEnvironmentModule) Init() {
	c.Name = "ClearOSModule"

	resetNetworkConfig := &task.RemoteTask{
		Name:     "ResetNetworkConfig",
		Desc:     "Reset os network config",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ResetNetworkConfig),
		Parallel: true,
	}

	uninstallETCD := &task.RemoteTask{
		Name:  "UninstallETCD",
		Desc:  "Uninstall etcd",
		Hosts: c.Runtime.GetHostsByRole(common.ETCD),
		Prepare: &prepare.PrepareCollection{
			new(EtcdTypeIsKubeKey),
		},
		Action:   new(UninstallETCD),
		Parallel: true,
	}

	removeFiles := &task.RemoteTask{
		Name:     "RemoveClusterFiles",
		Desc:     "Remove cluster files",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(RemoveClusterFiles),
		Parallel: true,
	}

	daemonReload := &task.RemoteTask{
		Name:     "DaemonReload",
		Desc:     "Systemd daemon reload",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(DaemonReload),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		resetNetworkConfig,
		uninstallETCD,
		removeFiles,
		daemonReload,
	}
}

type RepositoryOnlineModule struct {
	common.KubeModule
	Skip bool
}

func (r *RepositoryOnlineModule) IsSkip() bool {
	return r.Skip
}

func (r *RepositoryOnlineModule) Init() {
	r.Name = "RepositoryOnlineModule"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	newRepo := &task.RemoteTask{
		Name:     "NewRepoClient",
		Desc:     "New repository client",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(NewRepoClient),
		Parallel: true,
		Retry:    1,
	}

	install := &task.RemoteTask{
		Name:     "InstallPackage",
		Desc:     "Install packages",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(InstallPackage),
		Parallel: true,
		Retry:    1,
	}

	r.Tasks = []task.Interface{
		getOSData,
		newRepo,
		install,
	}
}

type RepositoryModule struct {
	common.KubeModule
	Skip bool
}

func (r *RepositoryModule) IsSkip() bool {
	return r.Skip
}

func (r *RepositoryModule) Init() {
	r.Name = "RepositoryModule"
	r.Desc = "Install local repository"

	getOSData := &task.RemoteTask{
		Name:     "GetOSData",
		Desc:     "Get OS release",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(GetOSData),
		Parallel: true,
	}

	sync := &task.RemoteTask{
		Name:     "SyncRepositoryISOFile",
		Desc:     "Sync repository iso file to all nodes",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(SyncRepositoryFile),
		Parallel: true,
		Retry:    2,
	}

	mount := &task.RemoteTask{
		Name:     "MountISO",
		Desc:     "Mount iso file",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(MountISO),
		Parallel: true,
		Retry:    1,
	}

	newRepo := &task.RemoteTask{
		Name:     "NewRepoClient",
		Desc:     "New repository client",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(NewRepoClient),
		Parallel: true,
		Retry:    1,
		Rollback: new(RollbackUmount),
	}

	backup := &task.RemoteTask{
		Name:     "BackupOriginalRepository",
		Desc:     "Backup original repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(BackupOriginalRepository),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverBackupSuccessNode),
	}

	add := &task.RemoteTask{
		Name:     "AddLocalRepository",
		Desc:     "Add local repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(AddLocalRepository),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverRepository),
	}

	install := &task.RemoteTask{
		Name:     "InstallPackage",
		Desc:     "Install packages",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(InstallPackage),
		Parallel: true,
		Retry:    1,
		Rollback: new(RecoverRepository),
	}

	reset := &task.RemoteTask{
		Name:     "ResetRepository",
		Desc:     "Reset repository to the original repository",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(ResetRepository),
		Parallel: true,
		Retry:    1,
	}

	umount := &task.RemoteTask{
		Name:     "UmountISO",
		Desc:     "Umount ISO file",
		Hosts:    r.Runtime.GetAllHosts(),
		Action:   new(UmountISO),
		Parallel: true,
	}

	r.Tasks = []task.Interface{
		getOSData,
		sync,
		mount,
		newRepo,
		backup,
		add,
		install,
		reset,
		umount,
	}
}
