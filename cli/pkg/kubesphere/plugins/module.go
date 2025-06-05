package plugins

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type CopyEmbed struct {
	common.KubeModule
}

func (t *CopyEmbed) Init() {
	t.Name = "CopyEmbed"

	copyEmbed := &task.LocalTask{
		Name:   "CopyEmbedFiles",
		Action: new(CopyEmbedFiles),
	}

	t.Tasks = []task.Interface{
		copyEmbed,
	}
}

type DeployKsPluginsModule struct {
	common.KubeModule
}

func (t *DeployKsPluginsModule) Init() {
	t.Name = "DeployKsPlugins"

	checkNodeState := &task.RemoteTask{
		Name:  "CheckNodeState",
		Hosts: t.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(CheckNodeState),
		Parallel: false,
		Retry:    20,
		Delay:    10 * time.Second,
	}

	initNs := &task.RemoteTask{
		Name:  "InitKsNamespace",
		Hosts: t.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(InitNamespace),
		Parallel: false,
	}

	t.Tasks = []task.Interface{
		checkNodeState,
		initNs,
	}
}
