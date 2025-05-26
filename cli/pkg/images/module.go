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

package images

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type PreloadImagesModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (p *PreloadImagesModule) IsSkip() bool {
	return p.Skip
}

func (p *PreloadImagesModule) Init() {
	p.Name = "PreloadImages"

	preload := &task.RemoteTask{
		Name:  "PreloadImages",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// &MasterPullImages{Not: true},
			&plugins.IsCloudInstance{Not: true},
			// &CheckImageManifest{},
			&ContainerdInstalled{},
		},
		Action: &LoadImages{
			ManifestAction: manifest.ManifestAction{
				Manifest: p.Manifest,
				BaseDir:  p.BaseDir,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	pinImages := &task.LocalTask{
		Name: "PinImages",
		Prepare: &prepare.PrepareCollection{
			&ContainerdInstalled{},
		},
		Action: &PinImages{
			ManifestAction: manifest.ManifestAction{
				Manifest: p.Manifest,
				BaseDir:  p.BaseDir,
			},
		},
	}

	p.Tasks = []task.Interface{
		preload,
		pinImages,
	}
}
