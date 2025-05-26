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

package etcd

import (
	"path/filepath"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/etcd/templates"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type PreCheckModule struct {
	common.KubeModule
	Skip bool
}

func (p *PreCheckModule) IsSkip() bool {
	return p.Skip
}

func (p *PreCheckModule) Init() {
	p.Name = "ETCDPreCheckModule"
	p.Desc = "Get ETCD cluster status"

	getStatus := &task.RemoteTask{
		Name:     "GetETCDStatus",
		Desc:     "Get etcd status",
		Hosts:    p.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(GetStatus),
		Parallel: false,
		Retry:    0,
	}
	p.Tasks = []task.Interface{
		getStatus,
	}
}

type CertsModule struct {
	common.KubeModule
	Skip bool
}

func (p *CertsModule) IsSkip() bool {
	return p.Skip
}

func (c *CertsModule) Init() {
	c.Name = "CertsModule"
	c.Desc = "Sign ETCD cluster certs"

	switch c.KubeConf.Cluster.Etcd.Type {
	case kubekeyapiv1alpha2.KubeKey:
		c.Tasks = CertsModuleForKubeKey(c)
	case kubekeyapiv1alpha2.External:
		c.Tasks = CertsModuleForExternal(c)
	}
}

func CertsModuleForKubeKey(c *CertsModule) []task.Interface {
	// If the etcd cluster already exists, obtain the certificate in use from the etcd node.
	fetchCerts := &task.RemoteTask{
		Name:     "FetchETCDCerts",
		Desc:     "Fetch etcd certs",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(FetchCerts),
		Parallel: false,
	}

	generateCerts := &task.LocalTask{
		Name:   "GenerateETCDCerts",
		Desc:   "Generate etcd Certs",
		Action: new(GenerateCerts),
	}

	syncCertsFile := &task.RemoteTask{
		Name:     "SyncCertsFile",
		Desc:     "Synchronize certs file",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	syncCertsToMaster := &task.RemoteTask{
		Name:     "SyncCertsFileToMaster",
		Desc:     "Synchronize certs file to master",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  &common.OnlyETCD{Not: true},
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	return []task.Interface{
		fetchCerts,
		generateCerts,
		syncCertsFile,
		syncCertsToMaster,
	}
}

func CertsModuleForExternal(c *CertsModule) []task.Interface {
	fetchCerts := &task.LocalTask{
		Name:   "FetchETCDCerts",
		Desc:   "Fetch etcd certs",
		Action: new(FetchCertsForExternalEtcd),
	}

	syncCertsToMaster := &task.RemoteTask{
		Name:     "SyncCertsFileToMaster",
		Desc:     "Synchronize certs file to master",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Action:   new(SyncCertsFile),
		Parallel: true,
		Retry:    1,
	}

	return []task.Interface{
		fetchCerts,
		syncCertsToMaster,
	}
}

type InstallETCDBinaryModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (p *InstallETCDBinaryModule) IsSkip() bool {
	return p.Skip
}

func (i *InstallETCDBinaryModule) Init() {
	i.Name = "InstallETCDBinaryModule"
	i.Desc = "Install ETCD cluster"

	installETCDBinary := &task.RemoteTask{
		Name:  "InstallETCDBinary",
		Desc:  "Install etcd using binary",
		Hosts: i.Runtime.GetHostsByRole(common.ETCD),
		Action: &InstallETCDBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  i.BaseDir,
				Manifest: i.Manifest,
			},
		},
		Parallel: true,
		Retry:    1,
	}

	generateETCDService := &task.RemoteTask{
		Name:  "GenerateETCDService",
		Desc:  "Generate etcd service",
		Hosts: i.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Name:     "GenerateETCDService",
			Template: templates.ETCDService,
			Dst:      "/etc/systemd/system/etcd.service",
		},
		Parallel: true,
		Retry:    1,
	}

	accessAddress := &task.RemoteTask{
		Name:     "GenerateAccessAddress",
		Desc:     "Generate access address",
		Hosts:    i.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(FirstETCDNode),
		Action:   new(GenerateAccessAddress),
		Parallel: true,
		Retry:    1,
	}

	i.Tasks = []task.Interface{
		installETCDBinary,
		generateETCDService,
		accessAddress,
	}
}

type ConfigureModule struct {
	common.KubeModule
	Skip bool
}

func (p *ConfigureModule) IsSkip() bool {
	return p.Skip
}

func (e *ConfigureModule) Init() {
	e.Name = "ETCDConfigureModule"
	e.Desc = "Configure ETCD cluster"

	if v, ok := e.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		if !cluster.clusterExist {
			e.Tasks = handleNewCluster(e)
		} else {
			e.Tasks = handleExistCluster(e)
		}
	}
}

func handleNewCluster(c *ConfigureModule) []task.Interface {

	existETCDHealthCheck := &task.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &task.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	allRefreshETCDConfig := &task.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	restart := &task.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	allETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	refreshETCDConfigToExist := &task.RemoteTask{
		Name:     "RefreshETCDConfigToExist",
		Desc:     "Refresh etcd.env config to exist mode on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   &RefreshConfig{ToExisting: true},
		Parallel: false,
	}

	tasks := []task.Interface{
		existETCDHealthCheck,
		generateETCDConfig,
		allRefreshETCDConfig,
		restart,
		allETCDNodeHealthCheck,
		refreshETCDConfigToExist,
		allETCDNodeHealthCheck,
	}
	return tasks
}

func handleExistCluster(c *ConfigureModule) []task.Interface {

	existETCDHealthCheck := &task.RemoteTask{
		Name:     "ExistETCDHealthCheck",
		Desc:     "Health check on exist etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  new(NodeETCDExist),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	generateETCDConfig := &task.RemoteTask{
		Name:     "GenerateETCDConfig",
		Desc:     "Generate etcd.env config on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(GenerateConfig),
		Parallel: false,
	}

	joinMember := &task.RemoteTask{
		Name:     "JoinETCDMember",
		Desc:     "Join etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(JoinMember),
		Parallel: false,
	}

	restart := &task.RemoteTask{
		Name:     "RestartETCD",
		Desc:     "Restart etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(RestartETCD),
		Parallel: true,
	}

	newETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "NewETCDNodeHealthCheck",
		Desc:     "Health check on new etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	checkMember := &task.RemoteTask{
		Name:     "CheckETCDMember",
		Desc:     "Check etcd member",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Prepare:  &NodeETCDExist{Not: true},
		Action:   new(CheckMember),
		Parallel: true,
	}

	allRefreshETCDConfig := &task.RemoteTask{
		Name:     "AllRefreshETCDConfig",
		Desc:     "Refresh etcd.env config on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(RefreshConfig),
		Parallel: false,
	}

	allETCDNodeHealthCheck := &task.RemoteTask{
		Name:     "AllETCDNodeHealthCheck",
		Desc:     "Health check on all etcd",
		Hosts:    c.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(HealthCheck),
		Parallel: true,
		Retry:    20,
	}

	tasks := []task.Interface{
		existETCDHealthCheck,
		generateETCDConfig,
		joinMember,
		restart,
		newETCDNodeHealthCheck,
		checkMember,
		allRefreshETCDConfig,
		allETCDNodeHealthCheck,
	}
	return tasks
}

type BackupModule struct {
	common.KubeModule
	Skip bool
}

func (p *BackupModule) IsSkip() bool {
	return p.Skip
}

func (b *BackupModule) Init() {
	b.Name = "ETCDBackupModule"
	b.Desc = "Backup ETCD cluster data"

	backupETCD := &task.RemoteTask{
		Name:     "BackupETCD",
		Desc:     "Backup etcd data regularly",
		Hosts:    b.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(BackupETCD),
		Parallel: true,
	}

	generateBackupETCDService := &task.RemoteTask{
		Name:  "GenerateBackupETCDService",
		Desc:  "Generate backup ETCD service",
		Hosts: b.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Name:     "GenerateBackupETCDService",
			Template: templates.BackupETCDService,
			Dst:      filepath.Join("/etc/systemd/system/", templates.BackupETCDService.Name()),
			Data: util.Data{
				"ScriptPath": filepath.Join(b.KubeConf.Cluster.Etcd.BackupScriptDir, "etcd-backup.sh"),
			},
		},
		Parallel: true,
	}

	generateBackupETCDTimer := &task.RemoteTask{
		Name:  "GenerateBackupETCDTimer",
		Desc:  "Generate backup ETCD timer",
		Hosts: b.Runtime.GetHostsByRole(common.ETCD),
		Action: &action.Template{
			Name:     "GenerateBackupETCDTimer",
			Template: templates.BackupETCDTimer,
			Dst:      filepath.Join("/etc/systemd/system/", templates.BackupETCDTimer.Name()),
			Data: util.Data{
				"OnCalendarStr": templates.BackupTimeOnCalendar(b.KubeConf.Cluster.Etcd.BackupPeriod),
			},
		},
		Parallel: true,
	}

	enable := &task.RemoteTask{
		Name:     "EnableBackupETCDService",
		Desc:     "Enable backup etcd service",
		Hosts:    b.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(EnableBackupETCDService),
		Parallel: true,
	}

	b.Tasks = []task.Interface{
		backupETCD,
		generateBackupETCDService,
		generateBackupETCDTimer,
		enable,
	}
}
