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

package storage

import (
	"fmt"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
)

type DeployLocalVolume struct {
	common.KubeAction
}

func (d *DeployLocalVolume) Execute(runtime connector.Runtime) error {
	cmd := fmt.Sprintf("/usr/local/bin/kubectl apply -f %s", filepath.Join(common.KubeAddonsDir, "local-volume.yaml"))
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy local-volume.yaml failed")
	}
	return nil
}
