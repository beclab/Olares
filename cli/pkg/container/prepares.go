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

package container

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type DockerExist struct {
	common.KubePrepare
	Not bool
}

func (d *DockerExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd("if [ -z $(which docker) ] || [ ! -e /var/run/docker.sock ]; "+
		"then echo 'not exist'; "+
		"fi", false, false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return d.Not, nil
	}
	return !d.Not, nil
}

type CrictlExist struct {
	common.KubePrepare
	Not bool
}

func (c *CrictlExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which crictl) ]; "+
			"then echo 'not exist'; "+
			"fi", false, false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return c.Not, nil
	} else {
		return !c.Not, nil
	}
}

type ContainerdExist struct {
	common.KubePrepare
	Not bool
}

func (c *ContainerdExist) PreCheck(runtime connector.Runtime) (bool, error) {
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; "+
			"then echo 'not exist'; "+
			"fi", false, false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return c.Not, nil
	}
	return !c.Not, nil
}

type PrivateRegistryAuth struct {
	common.KubePrepare
}

func (p *PrivateRegistryAuth) PreCheck(runtime connector.Runtime) (bool, error) {
	if len(p.KubeConf.Cluster.Registry.Auths.Raw) == 0 {
		return false, nil
	}
	return true, nil
}

type ZfsResetPrepare struct {
	common.KubePrepare
}

func (p *ZfsResetPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.GetSystemInfo().GetFsType() == "zfs" {
		return true, nil
	}
	return false, nil
}
