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

package os

import (
	kubekeyv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type CheckHwClock struct {
	common.KubePrepare
}

func (p *CheckHwClock) PreCheck(_ connector.Runtime) (bool, error) {
	hwclockPath, err := util.GetCommand(common.CommandHwclock)
	if err != nil {
		logger.Errorf("hwclock lookup error %v", err)
		return true, nil
	}

	if hwclockPath == "" {
		logger.Errorf("hwclock not found")
		return true, nil
	}

	return false, nil
}

type NodeConfigureNtpCheck struct {
	common.KubePrepare
}

func (n *NodeConfigureNtpCheck) PreCheck(_ connector.Runtime) (bool, error) {
	// skip when both NtpServers and Timezone was not set in cluster config
	if len(n.KubeConf.Cluster.System.NtpServers) == 0 && len(n.KubeConf.Cluster.System.Timezone) == 0 {
		return false, nil
	}

	return true, nil
}

type EtcdTypeIsKubeKey struct {
	common.KubePrepare
}

func (e *EtcdTypeIsKubeKey) PreCheck(_ connector.Runtime) (bool, error) {
	if len(e.KubeConf.Cluster.Etcd.Type) == 0 || e.KubeConf.Cluster.Etcd.Type == kubekeyv1alpha2.KubeKey {
		return true, nil
	}

	return false, nil
}

type DeleteNode struct {
	common.KubePrepare
}

func (d *DeleteNode) PreCheck(runtime connector.Runtime) (bool, error) {
	nodeName, ok := d.PipelineCache.Get("dstNode")
	if !ok {
		return true, nil
	}

	host := runtime.RemoteHost()
	if host.GetName() == nodeName {
		return true, nil
	}

	return false, nil
}

type IsPveLxc struct {
	common.KubePrepare
}

func (r *IsPveLxc) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.GetSystemInfo().IsPveLxc() {
		return true, nil
	}
	return false, nil
}

type IsPve struct {
	common.KubePrepare
}

func (r *IsPve) PreCheck(runtime connector.Runtime) (bool, error) {
	sys := runtime.GetSystemInfo()
	if sys.IsPve() {
		return true, nil
	}
	return false, nil
}
