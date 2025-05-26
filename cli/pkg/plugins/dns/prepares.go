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
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type CoreDNSExist struct {
	common.KubePrepare
	Not bool
}

func (c *CoreDNSExist) PreCheck(runtime connector.Runtime) (bool, error) {
	_, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get svc -n kube-system coredns", false, false)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return c.Not, nil
		}
		return false, err
	}
	return !c.Not, nil
}

type EnableNodeLocalDNS struct {
	common.KubePrepare
}

func (e *EnableNodeLocalDNS) PreCheck(runtime connector.Runtime) (bool, error) {
	if e.KubeConf.Cluster.Kubernetes.EnableNodelocaldns() {
		return true, nil
	}
	return false, nil
}

type NodeLocalDNSConfigMapNotExist struct {
	common.KubePrepare
}

func (n *NodeLocalDNSConfigMapNotExist) PreCheck(runtime connector.Runtime) (bool, error) {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get cm -n kube-system nodelocaldns", false, false); err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
