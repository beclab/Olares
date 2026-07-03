package terminus

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type InstallLinkerdFullAddonModule struct {
	common.KubeModule
}

func (m *InstallLinkerdFullAddonModule) Init() {
	m.Name = "InstallLinkerdFullAddonModule"
	m.Tasks = []task.Interface{
		&task.LocalTask{Name: "ValidateLinkerdDeployAssets", Action: &ValidateLinkerdDeployAssets{}},
		&task.LocalTask{Name: "SyncLinkerdPKIAndIdentity", Action: &SyncLinkerdPKIAndIdentity{}},
		&task.LocalTask{Name: "ApplyLinkerdMeshBootstrapNP", Action: &ApplyLinkerdMeshBootstrapNP{}},
		&task.LocalTask{Name: "WaitLinkerdControlPlaneReady", Action: &WaitLinkerdControlPlaneReady{}},
	}
}
