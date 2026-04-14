package mtgpu

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallMThreadsPluginModule installs MThreads GPU device plugin on Kubernetes.
type InstallMThreadsPluginModule struct {
	common.KubeModule
	Skip bool // conditional execution based on GPU enablement
}

func (m *InstallMThreadsPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallMThreadsPluginModule) Init() {
	m.Name = "InstallMThreadsPlugin"

	// update node with MThreads GPU labels
	updateNode := &task.RemoteTask{
		Name:     "UpdateNodeMThreadsGPUInfo",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(UpdateNodeMThreadsGPUInfo),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		updateNode,
	}
}
