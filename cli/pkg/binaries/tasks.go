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

package binaries

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/manifest"
)

type InstallAppArmorTask struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallAppArmorTask) Execute(runtime connector.Runtime) error {
	fileName, err := GetUbutun24AppArmor(t.BaseDir, t.Manifest)
	if err != nil {
		logger.Fatal("failed to download apparmor: %v", err)
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("dpkg -i %s", fileName), false, true); err != nil {
		logger.Errorf("failed to install apparmor: %v", err)
		return err
	}

	return nil
}
