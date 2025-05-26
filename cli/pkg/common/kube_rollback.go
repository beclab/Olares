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

package common

import (
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/rollback"
)

type KubeRollback struct {
	rollback.BaseRollback
	KubeConf *KubeConf
}

func (k *KubeRollback) AutoAssert(runtime connector.Runtime) {
	kubeRuntime := runtime.(*KubeRuntime)
	conf := &KubeConf{
		Cluster:     kubeRuntime.Cluster,
		ClusterName: kubeRuntime.ClusterName,
		Kubeconfig:  kubeRuntime.Kubeconfig,
		ClientSet:   kubeRuntime.ClientSet,
		Arg:         kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
