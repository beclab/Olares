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

package certs

import (
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/certs/templates"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type AutoRenewCertsEnabled struct {
	common.KubePrepare
	Not bool
}

func (a *AutoRenewCertsEnabled) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist(filepath.Join("/etc/systemd/system/", templates.K8sCertsRenewService.Name()))
	if err != nil {
		return false, err
	}
	if exist {
		return !a.Not, nil
	}
	return a.Not, nil
}
