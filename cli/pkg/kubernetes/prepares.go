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
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

type NodesInfoGetter interface {
	GetNodesInfo() map[string]string
}

type NodeInCluster struct {
	common.KubePrepare
	Not         bool
	NoneCluster bool
}

func (n *NodeInCluster) PreCheck(runtime connector.Runtime) (bool, error) {
	if n.NoneCluster {
		return true, nil
	}
	host := runtime.RemoteHost()
	if v, ok := n.PipelineCache.Get(common.ClusterStatus); ok {
		nodesInfoGetter, ok := v.(NodesInfoGetter)
		if !ok {
			return false, errors.New("get cluster status by pipeline cache failed")
		}
		nodesInfo := nodesInfoGetter.GetNodesInfo()
		var versionOk bool
		if res, ok := nodesInfo[host.GetName()]; ok && res != "" {
			versionOk = true
		}
		_, ipOk := nodesInfo[host.GetInternalAddress()]
		if n.Not {
			return !(versionOk || ipOk), nil
		}
		return versionOk || ipOk, nil
	} else {
		return false, errors.New("get cluster status by pipeline cache failed")
	}
}

type ClusterIsExist struct {
	common.KubePrepare
	Not bool
}

func (c *ClusterIsExist) PreCheck(_ connector.Runtime) (bool, error) {
	if exist, ok := c.PipelineCache.GetMustBool(common.ClusterExist); ok {
		if c.Not {
			return !exist, nil
		}
		return exist, nil
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
}

type CheckKubeadmExist struct {
	common.KubePrepare
}

func (p *CheckKubeadmExist) PreCheck(runtime connector.Runtime) (bool, error) {
	if util.IsExist("/usr/local/bin/kubeadm") {
		return true, nil
	}
	return false, nil
}
