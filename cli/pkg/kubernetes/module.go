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

package kubernetes

import (
	"fmt"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type StatusModule struct {
	common.KubeModule
}

func (k *StatusModule) Init() {
	k.Name = "KubernetesStatusModule"
	k.Desc = "Get kubernetes cluster status"

	cluster := NewKubernetesStatus()
	k.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:  "GetClusterStatus",
		Desc:  "Get kubernetes cluster status",
		Hosts: k.Runtime.GetHostsByRole(common.Master),
		//Prepare:  new(NoClusterInfo),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	k.Tasks = []task.Interface{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"
	i.Desc = "Install kubernetes cluster"

	syncBinary := &task.RemoteTask{
		Name:    "SyncKubeBinary",
		Desc:    "Synchronize kubernetes binaries",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &SyncKubeBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  i.BaseDir,
				Manifest: i.Manifest,
			},
		},
		Parallel: true,
		Retry:    2,
	}

	syncKubelet := &task.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "Synchronize kubelet",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &task.RemoteTask{
		Name:     "GenerateKubeletService",
		Desc:     "Generate kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   &GenerateKubeletService{},
		Parallel: true,
		Retry:    2,
	}

	enableKubelet := &task.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "Enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &task.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "Generate kubelet env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(GenerateKubeletEnv),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []task.Interface{
		syncBinary,
		syncKubelet,
		generateKubeletService,
		enableKubelet,
		generateKubeletEnv,
	}
}

type InitKubernetesModule struct {
	common.KubeModule
}

func (i *InitKubernetesModule) Init() {
	i.Name = "InitKubernetesModule"
	i.Desc = "Init kubernetes cluster"

	generateKubeadmConfig := &task.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action: &GenerateKubeadmConfig{
			IsInitConfiguration:     true,
			WithSecurityEnhancement: i.KubeConf.Arg.SecurityEnhancement,
		},
		Parallel: true,
	}

	kubeadmInit := &task.RemoteTask{
		Name:  "KubeadmInit",
		Desc:  "Init cluster using kubeadm",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(KubeadmInit),
		Retry:    3,
		Parallel: true,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy admin.conf to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: true,
	}

	removeMasterTaint := &task.RemoteTask{
		Name:  "RemoveMasterTaint",
		Desc:  "Remove master taint",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(RemoveMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabel := &task.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	i.Tasks = []task.Interface{
		generateKubeadmConfig,
		kubeadmInit,
		copyKubeConfig,
		removeMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "JoinNodesModule"
	j.Desc = "Join kubernetes nodes"

	j.PipelineCache.Set(common.ClusterExist, true)

	generateKubeadmConfig := &task.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action: &GenerateKubeadmConfig{
			IsInitConfiguration:     false,
			WithSecurityEnhancement: j.KubeConf.Arg.SecurityEnhancement,
		},
		Parallel: true,
	}

	joinMasterNode := &task.RemoteTask{
		Name:  "JoinControlPlaneNode(k8s)",
		Desc:  "Join control-plane node",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(JoinNode),
		Parallel: true,
		Retry:    5,
	}

	joinWorkerNode := &task.RemoteTask{
		Name:  "JoinWorkerNode(k8s)",
		Desc:  "Join worker node",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(JoinNode),
		Parallel: true,
		Retry:    5,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "copyKubeConfig",
		Desc:  "Copy admin.conf to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: true,
		Retry:    2,
	}

	removeMasterTaint := &task.RemoteTask{
		Name:  "RemoveMasterTaint",
		Desc:  "Remove master taint",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(RemoveMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabelToMaster := &task.RemoteTask{
		Name:  "AddWorkerLabelToMaster",
		Desc:  "Add worker label to master",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	syncKubeConfig := &task.RemoteTask{
		Name:  "SyncKubeConfig(k8s)",
		Desc:  "Synchronize kube config to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
		Retry:    3,
	}

	addWorkerLabelToWorker := &task.RemoteTask{
		Name:  "AddWorkerLabelToWorker",
		Desc:  "Add worker label to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	j.Tasks = []task.Interface{
		generateKubeadmConfig,
		joinMasterNode,
		joinWorkerNode,
		copyKubeConfig,
		removeMasterTaint,
		addWorkerLabelToMaster,
		syncKubeConfig,
		addWorkerLabelToWorker,
	}
}

type ResetClusterModule struct {
	common.KubeModule
}

func (r *ResetClusterModule) Init() {
	r.Name = "ResetClusterModule"
	r.Desc = "Reset kubernetes cluster"

	deleteCurrentNode := &task.LocalTask{
		Name:   "DeleteCurrentNode",
		Action: new(KubectlDeleteCurrentWorkerNode),
	}

	kubeadmReset := &task.RemoteTask{
		Name:     "KubeadmReset(k8s)",
		Desc:     "Reset the cluster using kubeadm",
		Hosts:    r.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(CheckKubeadmExist),
		Action:   new(KubeadmReset),
		Parallel: false,
	}

	r.Tasks = []task.Interface{
		deleteCurrentNode,
		kubeadmReset,
	}
}

type UmountKubeModule struct {
	common.KubeModule
}

func (c *UmountKubeModule) Init() {
	c.Name = "UmountKubeModule(k8s)"

	umountKubelet := &task.RemoteTask{
		Name:     "Umountkube(k8s)",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(UmountKubelet),
		Parallel: false,
		Retry:    1,
	}

	c.Tasks = []task.Interface{
		umountKubelet,
	}
}

type CompareConfigAndClusterInfoModule struct {
	common.KubeModule
}

func (c *CompareConfigAndClusterInfoModule) Init() {
	c.Name = "CompareConfigAndClusterInfoModule"
	c.Desc = "Compare config and cluster nodes info"

	check := &task.RemoteTask{
		Name:    "FindNode(k8s)",
		Desc:    "Find information about nodes that are expected to be deleted",
		Hosts:   c.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(FindNode),
	}

	c.Tasks = []task.Interface{
		check,
	}
}

type DeleteKubeNodeModule struct {
	common.KubeModule
}

func (d *DeleteKubeNodeModule) Init() {
	d.Name = "DeleteKubeNodeModule"
	d.Desc = "Delete kubernetes node"

	drain := &task.RemoteTask{
		Name:    "DrainNode(k8s)",
		Desc:    "Node safely evict all pods",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(DrainNode),
		Retry:   2,
	}

	deleteNode := &task.RemoteTask{
		Name:    "DeleteNode(k8s)",
		Desc:    "Delete the node using kubectl",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(KubectlDeleteNode),
		Retry:   5,
	}

	d.Tasks = []task.Interface{
		drain,
		deleteNode,
	}
}

type SetUpgradePlanModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (s *SetUpgradePlanModule) Init() {
	s.Name = fmt.Sprintf("SetUpgradePlanModule %d/%d", s.Step, len(UpgradeStepList))
	s.Desc = "Set upgrade plan"

	plan := &task.LocalTask{
		Name:   "SetUpgradePlan(k8s)",
		Desc:   "Set upgrade plan",
		Action: &SetUpgradePlan{Step: s.Step},
	}

	s.Tasks = []task.Interface{
		plan,
	}
}

type SaveKubeConfigModule struct {
	common.KubeModule
}

func (s *SaveKubeConfigModule) Init() {
	s.Name = "SaveKubeConfigModule"
	s.Desc = "Save kube config file as a configmap"

	save := &task.LocalTask{
		Name:   "SaveKubeConfig(k8s)",
		Desc:   "Save kube config as a configmap",
		Action: new(SaveKubeConfig),
		Retry:  5,
	}

	s.Tasks = []task.Interface{
		save,
	}
}

type ConfigureKubernetesModule struct {
	common.KubeModule
}

func (c *ConfigureKubernetesModule) Init() {
	c.Name = "ConfigureKubernetesModule"
	c.Desc = "Configure kubernetes"

	configure := &task.RemoteTask{
		Name:     "ConfigureKubernetes",
		Desc:     "Configure kubernetes",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(ConfigureKubernetes),
		Retry:    6,
		Delay:    10 * time.Second,
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		configure,
	}
}

type SecurityEnhancementModule struct {
	common.KubeModule
	Skip bool
}

func (s *SecurityEnhancementModule) IsSkip() bool {
	return s.Skip
}

func (s *SecurityEnhancementModule) Init() {
	s.Name = "SecurityEnhancementModule"
	s.Desc = "Security enhancement for the cluster"

	etcdSecurityEnhancement := &task.RemoteTask{
		Name:     "EtcdSecurityEnhancementTask(k8s)",
		Desc:     "Security enhancement for etcd",
		Hosts:    s.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(EtcdSecurityEnhancemenAction),
		Parallel: true,
	}

	masterSecurityEnhancement := &task.RemoteTask{
		Name:     "K8sSecurityEnhancementTask(k8s)",
		Desc:     "Security enhancement for kubernetes",
		Hosts:    s.Runtime.GetHostsByRole(common.Master),
		Action:   new(MasterSecurityEnhancemenAction),
		Parallel: true,
	}

	nodesSecurityEnhancement := &task.RemoteTask{
		Name:     "K8sSecurityEnhancementTask(k8s)",
		Desc:     "Security enhancement for kubernetes",
		Hosts:    s.Runtime.GetHostsByRole(common.Worker),
		Action:   new(NodesSecurityEnhancemenAction),
		Parallel: true,
	}

	s.Tasks = []task.Interface{
		etcdSecurityEnhancement,
		masterSecurityEnhancement,
		nodesSecurityEnhancement,
	}
}
