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

package network

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type OldK8sVersion struct {
	common.KubePrepare
	Not bool
}

func (o *OldK8sVersion) PreCheck(_ connector.Runtime) (bool, error) {
	cmp, err := versionutil.MustParseSemantic(o.KubeConf.Cluster.Kubernetes.Version).Compare("v1.16.0")
	if err != nil {
		return false, err
	}
	// old version
	if cmp == -1 {
		return !o.Not, nil
	}
	// new version
	return o.Not, nil
}

type EnableSSL struct {
	common.KubePrepare
}

func (e *EnableSSL) PreCheck(_ connector.Runtime) (bool, error) {
	if e.KubeConf.Cluster.Network.Kubeovn.EnableSSL {
		return true, nil
	}
	return false, nil
}
