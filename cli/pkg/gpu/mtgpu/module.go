package mtgpu

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

// LabelNodeModule only (re-)applies the Moore Threads mode label. `gpu enable`
// uses it to restore the label that `gpu disable` wiped; the action is a no-op
// on nodes without an MThreads AI Book M1000.
type LabelNodeModule struct {
	common.KubeModule
}

func (m *LabelNodeModule) Init() {
	m.Name = "LabelMThreadsGPUNode"

	labelNode := &task.LocalTask{
		Name:    "UpdateNodeMThreadsGPUInfo",
		Prepare: new(gpu.CurrentNodeInK8s),
		Action:  new(UpdateNodeMThreadsGPUInfo),
		Retry:   1,
	}

	m.Tasks = []task.Interface{
		labelNode,
	}
}

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
