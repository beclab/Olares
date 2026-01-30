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

package precheck

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
)

type RemoveChattrModule struct {
	common.KubeModule
}

func (m *RemoveChattrModule) Init() {
	m.Name = "RemoveWSLChattr"

	removeChattr := &task.RemoteTask{
		Name:     "RemoveWSLChattr",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(RemoveWSLChattr),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		removeChattr,
	}
}

type GetStorageKeyModule struct {
	common.KubeModule
}

func (m *GetStorageKeyModule) Init() {
	m.Name = "GetStorage"

	getStorageKeyTask := &task.RemoteTask{
		Name:     "GetStorageKey",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetStorageKeyTask),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		getStorageKeyTask,
	}
}

type RunPrechecksModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *RunPrechecksModule) Init() {
	m.Name = "RunPrechecks"

	checkers := []Checker{
		new(SystemSupportCheck),
		new(SystemdCheck),
		new(RequiredPortsCheck),
		new(ConflictingContainerdCheck),
		new(NvidiaCardArchChecker),
		new(NouveauChecker),
		new(CudaChecker),
		new(RocmChecker),
	}
	runPreChecks := &task.LocalTask{
		Name: "RunPrechecks",
		Action: &RunChecks{
			Checkers: checkers,
		},
	}

	m.Tasks = []task.Interface{
		runPreChecks,
	}
}

type GreetingsModule struct {
	module.BaseTaskModule
}

func (h *GreetingsModule) Init() {
	h.Name = "GreetingsModule"

	var timeout int64

	for _, v := range h.Runtime.GetAllHosts() {
		timeout += v.GetTimeout()
	}

	hello := &task.RemoteTask{
		Name:     "Greetings",
		Hosts:    h.Runtime.GetAllHosts(),
		Action:   new(GreetingsTask),
		Parallel: false,
		Timeout:  time.Duration(timeout) * time.Second,
	}

	h.Tasks = []task.Interface{
		hello,
	}
}

type NodePreCheckModule struct {
	common.KubeModule
	Skip bool
}

func (n *NodePreCheckModule) IsSkip() bool {
	return n.Skip
}

func (n *NodePreCheckModule) Init() {
	n.Name = "NodePreCheckModule"
	n.Desc = "Do pre-check on cluster nodes"

	preCheck := &task.RemoteTask{
		Name:  "NodePreCheck",
		Desc:  "A pre-check on nodes",
		Hosts: n.Runtime.GetAllHosts(),
		//Prepare: &prepare.FastPrepare{
		//	Inject: func(runtime connector.Runtime) (bool, error) {
		//		if len(n.Runtime.GetHostsByRole(common.ETCD))%2 == 0 {
		//			logger.Error("The number of etcd is even. Please configure it to be odd.")
		//			return false, errors.New("the number of etcd is even")
		//		}
		//		return true, nil
		//	}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	n.Tasks = []task.Interface{
		preCheck,
	}
}
