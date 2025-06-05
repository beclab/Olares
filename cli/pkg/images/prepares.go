// Copyright 2022 bytetrade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package images

import (
	"fmt"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type ContainerdInstalled struct {
	common.KubePrepare
}

func (c *ContainerdInstalled) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.RemoteHost().GetOs() == common.Darwin {
		return true, nil
	}
	output, err := runtime.GetRunner().SudoCmd(
		"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; "+
			"then echo 'not exist'; "+
			"fi", false, false)
	if err != nil {
		return false, err
	}
	if strings.Contains(output, "not exist") {
		return false, fmt.Errorf("containerd service not installed")
	}
	return true, nil
}
