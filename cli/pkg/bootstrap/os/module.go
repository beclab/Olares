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
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/kubernetes"

	"github.com/beclab/Olares/cli/pkg/bootstrap/os/templates"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
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

	c.Tasks = []task.Interface{
		updateNtpDateTask,
		timeSyncTask,
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

	symlinkSysconf := &task.RemoteTask{
		Name:     "SymlinkSysconf",
		Desc:     "Create symbolic link to sysconf if non-existing",
		Hosts:    c.Runtime.GetAllHosts(),
		Prepare:  &kubernetes.NodeInCluster{Not: true},
		Action:   new(SymLinkSysconf),
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
		symlinkSysconf,
		ConfigureNtpServer,
		configureSwap,
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
