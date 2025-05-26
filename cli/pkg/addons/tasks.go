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

package addons

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
)

type Install struct {
	common.KubeAction
}

func (i *Install) Execute(runtime connector.Runtime) error {
	nums := len(i.KubeConf.Cluster.Addons)
	for index, addon := range i.KubeConf.Cluster.Addons {
		logger.Infof("%s Install addon [%v-%v]: %s", runtime.RemoteHost().GetName(), nums, index, addon.Name)
		if err := InstallAddons(i.KubeConf, &addon, filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))); err != nil {
			return err
		}
	}
	return nil
}
