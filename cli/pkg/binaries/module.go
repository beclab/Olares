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
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type CriBinariesModule struct {
	common.KubeModule
}

func (i *CriBinariesModule) Init() {
	i.Name = "CriBinariesModule"
	i.Desc = "Download Cri package"
	switch i.KubeConf.Arg.Type {
	case common.Docker:
		i.Tasks = CriBinaries(i)
	case common.Containerd:
		i.Tasks = CriBinaries(i)
	default:
	}

}

func CriBinaries(p *CriBinariesModule) []task.Interface {

	download := &task.LocalTask{
		Name:   "DownloadCriPackage",
		Desc:   "Download Cri package",
		Action: new(CriDownload),
	}

	p.Tasks = []task.Interface{
		download,
	}
	return p.Tasks
}

// TODO: install helm
