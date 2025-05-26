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

package k3s

import (
	"path"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

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
		return false, errors.New("get k3s cluster status by pipeline cache failed")
	}
}

type UsePrivateRegstry struct {
	common.KubePrepare
	Not bool
}

func (c *UsePrivateRegstry) PreCheck(_ connector.Runtime) (bool, error) {
	return c.KubeConf.Cluster.Registry.PrivateRegistry != "", nil
}

type CheckK3sUninstallScript struct {
	common.KubePrepare
}

func (p *CheckK3sUninstallScript) PreCheck(_ connector.Runtime) (bool, error) {
	var scriptPath = path.Join(common.BinDir, "k3s-uninstall.sh")
	if util.IsExist(scriptPath) {
		return true, nil
	}

	return false, nil
}

type CheckK3sKillAllScript struct {
	common.KubePrepare
}

func (p *CheckK3sKillAllScript) PreCheck(_ connector.Runtime) (bool, error) {
	var scriptPath = path.Join(common.BinDir, "k3s-killall.sh")
	if util.IsExist(scriptPath) {
		return true, nil
	}

	return false, nil
}
