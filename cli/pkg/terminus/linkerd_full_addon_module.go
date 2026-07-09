package terminus

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

const linkerdInstallTaskRetry = 3

type InstallLinkerdFullAddonModule struct {
	common.KubeModule
}

func (m *InstallLinkerdFullAddonModule) Init() {
	m.Name = "InstallLinkerdFullAddonModule"
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "WaitAppGatewayMeshNP",
			Action: &WaitAppGatewayMeshNP{},
			Retry:  linkerdInstallTaskRetry,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "SyncLinkerdPKIAndIdentity",
			Action: &SyncLinkerdPKIAndIdentity{},
			Retry:  linkerdInstallTaskRetry,
			Delay:  15 * time.Second,
		},
		&task.LocalTask{
			Name:   "WaitLinkerdControlPlaneReady",
			Action: &WaitLinkerdControlPlaneReady{},
			Retry:  linkerdInstallTaskRetry,
			Delay:  15 * time.Second,
		},
	}
}
