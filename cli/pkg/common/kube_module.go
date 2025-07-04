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

package common

import (
	kubekeyapiv1alpha2 "github.com/beclab/Olares/cli/apis/kubekey/v1alpha2"
	kubekeyclientset "github.com/beclab/Olares/cli/clients/clientset/versioned"
	"github.com/beclab/Olares/cli/pkg/core/module"
)

type KubeConf struct {
	ClusterHosts []string
	ClusterName  string
	Cluster      *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig   string
	ClientSet    *kubekeyclientset.Clientset
	Arg          *Argument
}

type KubeModule struct {
	module.BaseTaskModule
	KubeConf *KubeConf
}

func (k *KubeModule) IsSkip() bool {
	return k.Skip
}

func (k *KubeModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterName: kubeRuntime.ClusterName,
		Cluster:     kubeRuntime.Cluster,
		Kubeconfig:  kubeRuntime.Kubeconfig,
		ClientSet:   kubeRuntime.ClientSet,
		Arg:         kubeRuntime.Arg,
	}

	k.KubeConf = conf
}

type KubeCustomModule struct {
	module.CustomModule
	KubeConf *KubeConf
}

func (k *KubeCustomModule) AutoAssert() {
	kubeRuntime := k.Runtime.(*KubeRuntime)
	conf := &KubeConf{
		ClusterName: kubeRuntime.ClusterName,
		Cluster:     kubeRuntime.Cluster,
		Kubeconfig:  kubeRuntime.Kubeconfig,
		ClientSet:   kubeRuntime.ClientSet,
		Arg:         kubeRuntime.Arg,
	}

	k.KubeConf = conf
}
