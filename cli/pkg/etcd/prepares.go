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
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
)

type FirstETCDNode struct {
	common.KubePrepare
	Not bool
}

func (f *FirstETCDNode) PreCheck(runtime connector.Runtime) (bool, error) {
	v, ok := f.PipelineCache.Get(common.ETCDCluster)
	if !ok {
		return false, errors.New("get etcd cluster status by pipeline cache failed")
	}
	cluster := v.(*EtcdCluster)

	if (!cluster.clusterExist && runtime.GetHostsByRole(common.ETCD)[0].GetName() == runtime.RemoteHost().GetName()) ||
		(cluster.clusterExist && strings.Contains(cluster.peerAddresses[0], runtime.RemoteHost().GetInternalAddress())) {
		return !f.Not, nil
	}
	return f.Not, nil
}

type NodeETCDExist struct {
	common.KubePrepare
	Not bool
}

func (n *NodeETCDExist) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := host.GetCache().GetMustBool(common.ETCDExist); ok {
		if v {
			return !n.Not, nil
		}
		return n.Not, nil
	}
	return false, errors.New("get etcd node status by host label failed")
}
