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
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
)

type KubePrepare struct {
	prepare.BasePrepare
	KubeConf *KubeConf
}

func (k *KubePrepare) AutoAssert(runtime connector.Runtime) {
	kubeRuntime := runtime.(*KubeRuntime)
	conf := &KubeConf{
		Cluster:    kubeRuntime.Cluster,
		Kubeconfig: kubeRuntime.Kubeconfig,
		ClientSet:  kubeRuntime.ClientSet,
		Arg:        kubeRuntime.Arg,
	}

	k.KubeConf = conf
}

type OnlyFirstMaster struct {
	KubePrepare
	Not bool
}

func (o *OnlyFirstMaster) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Master) &&
		runtime.RemoteHost().GetName() == runtime.GetHostsByRole(Master)[0].GetName() {
		return !o.Not, nil
	}
	return o.Not, nil
}

type IsMaster struct {
	KubePrepare
}

func (i *IsMaster) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Master) {
		return true, nil
	}
	return false, nil
}

type IsWorker struct {
	KubePrepare
	Not bool
}

func (i *IsWorker) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Worker) {
		return !i.Not, nil
	}
	return i.Not, nil
}

type OnlyWorker struct {
	KubePrepare
}

func (o *OnlyWorker) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(Worker) && !runtime.RemoteHost().IsRole(Master) {
		return true, nil
	}
	return false, nil
}

type OnlyETCD struct {
	KubePrepare
	Not bool
}

func (o *OnlyETCD) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().IsRole(ETCD) {
		return !o.Not, nil
	}
	return o.Not, nil
}

type OnlyK3s struct {
	KubePrepare
}

func (o *OnlyK3s) PreCheck(_ connector.Runtime) (bool, error) {
	if o.KubeConf.Cluster.Kubernetes.Type == "k3s" {
		return true, nil
	}
	return false, nil
}

type OnlyKubernetes struct {
	KubePrepare
}

func (o *OnlyKubernetes) PreCheck(_ connector.Runtime) (bool, error) {
	if o.KubeConf.Cluster.Kubernetes.Type != "k3s" {
		return true, nil
	}
	return false, nil
}

type EnableKubeProxy struct {
	KubePrepare
}

func (e *EnableKubeProxy) PreCheck(_ connector.Runtime) (bool, error) {
	if !e.KubeConf.Cluster.Kubernetes.DisableKubeProxy {
		return true, nil
	}
	return false, nil
}
