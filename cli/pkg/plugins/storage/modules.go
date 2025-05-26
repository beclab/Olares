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
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/plugins/storage/templates"
)

type DeployLocalVolumeModule struct {
	common.KubeModule
	Skip bool
}

func (d *DeployLocalVolumeModule) IsSkip() bool {
	return d.Skip
}

func (d *DeployLocalVolumeModule) Init() {
	d.Name = "DeployStorageClassModule"
	d.Desc = "Deploy cluster storage-class"

	generate := &task.RemoteTask{
		Name:  "GenerateOpenEBSManifest",
		Desc:  "Generate OpenEBS manifest",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CheckDefaultStorageClass),
		},
		Action: &action.Template{
			Name:     "GenerateOpenEBSManifest",
			Template: templates.OpenEBS,
			Dst:      filepath.Join(common.KubeAddonsDir, templates.OpenEBS.Name()),
			Data: util.Data{
				"ProvisionerLocalPVImage": images.GetImage(d.Runtime, d.KubeConf, "provisioner-localpv").ImageName(),
				"LinuxUtilsImage":         images.GetImage(d.Runtime, d.KubeConf, "linux-utils").ImageName(),
			},
		},
		Parallel: true,
	}

	deploy := &task.RemoteTask{
		Name:  "DeployOpenEBS",
		Desc:  "Deploy OpenEBS as cluster default StorageClass",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CheckDefaultStorageClass),
		},
		Action:   new(DeployLocalVolume),
		Parallel: true,
	}

	d.Tasks = []task.Interface{
		generate,
		deploy,
	}
}
