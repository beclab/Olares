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

package dns

import (
	"path/filepath"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/plugins/dns/templates"
)

type SetProxyModule struct {
	common.KubeModule
}

func (s *SetProxyModule) Init() {
	s.Name = "SetProxy"
}

type ClusterDNSModule struct {
	common.KubeModule
}

func (c *ClusterDNSModule) Init() {
	c.Name = "ClusterDNSModule"

	// inject coredns svc ip to the global envs before other relative components are created
	common.TerminusGlobalEnvs["COREDNS_SVC"] = c.KubeConf.Cluster.CorednsClusterIP()

	generateCoreDNDService := &task.RemoteTask{
		Name:  "GenerateCoreDNSService",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action: &action.Template{
			Name:     "GenerateCoreDNSService",
			Template: templates.CorednsService,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CorednsService.Name()),
			Data: util.Data{
				"ClusterIP": c.KubeConf.Cluster.CorednsClusterIP(),
				"DNSDomain": c.KubeConf.Cluster.Kubernetes.DNSDomain,
			},
		},
		Parallel: true,
	}

	applyCoreDNSService := &task.RemoteTask{
		Name:  "ApplyCoreDNSService",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CoreDNSExist{Not: true},
		},
		Action:   new(ApplyCoreDNS),
		Retry:    5,
		Delay:    5 * time.Second,
		Parallel: true,
	}

	// disable nodelocaldns as it does not support the hosts plugin
	// and forwards all non-cluster dns request directly to the upstream dns server
	// rather than coredns
	// this can be configured, but in our most common case where only a single node is deployed
	// it only adds an unnecessary layer to the dns process chain

	//generateNodeLocalDNS := &task.RemoteTask{
	//	Name:  "GenerateNodeLocalDNS",
	//	Desc:  "Generate nodelocaldns",
	//	Hosts: c.Runtime.GetHostsByRole(common.Master),
	//	Prepare: &prepare.PrepareCollection{
	//		new(common.OnlyFirstMaster),
	//		new(EnableNodeLocalDNS),
	//	},
	//	Action: &action.Template{
	//		Name:     "GenerateNodeLocalDNS",
	//		Template: templates.NodeLocalDNSService,
	//		Dst:      filepath.Join(common.KubeConfigDir, templates.NodeLocalDNSService.Name()),
	//		Data: util.Data{
	//			"NodelocaldnsImage": images.GetImage(c.Runtime, c.KubeConf, "k8s-dns-node-cache").ImageName(),
	//		},
	//	},
	//	Parallel: true,
	//}
	//
	//applyNodeLocalDNS := &task.RemoteTask{
	//	Name:  "DeployNodeLocalDNS",
	//	Desc:  "Deploy nodelocaldns",
	//	Hosts: c.Runtime.GetHostsByRole(common.Master),
	//	Prepare: &prepare.PrepareCollection{
	//		new(common.OnlyFirstMaster),
	//		new(EnableNodeLocalDNS),
	//	},
	//	Action:   new(DeployNodeLocalDNS),
	//	Parallel: true,
	//	Retry:    5,
	//}
	//
	//generateNodeLocalDNSConfigMap := &task.RemoteTask{
	//	Name:  "GenerateNodeLocalDNSConfigMap",
	//	Desc:  "Generate nodelocaldns configmap",
	//	Hosts: c.Runtime.GetHostsByRole(common.Master),
	//	Prepare: &prepare.PrepareCollection{
	//		new(common.OnlyFirstMaster),
	//		new(EnableNodeLocalDNS),
	//		new(NodeLocalDNSConfigMapNotExist),
	//	},
	//	Action:   new(GenerateNodeLocalDNSConfigMap),
	//	Parallel: true,
	//}
	//
	//applyNodeLocalDNSConfigMap := &task.RemoteTask{
	//	Name:  "ApplyNodeLocalDNSConfigMap",
	//	Desc:  "Apply nodelocaldns configmap",
	//	Hosts: c.Runtime.GetHostsByRole(common.Master),
	//	Prepare: &prepare.PrepareCollection{
	//		new(common.OnlyFirstMaster),
	//		new(EnableNodeLocalDNS),
	//		new(NodeLocalDNSConfigMapNotExist),
	//	},
	//	Action:   new(ApplyNodeLocalDNSConfigMap),
	//	Parallel: true,
	//	Retry:    5,
	//}

	c.Tasks = []task.Interface{
		generateCoreDNDService,
		applyCoreDNSService,
	}
}
