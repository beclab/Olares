package intelgpu

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallIntelPluginModule advertises Intel integrated-GPU support on the node.
// There is no device plugin to install in this version (Intel iGPUs use unified
// system memory and are scheduled like Apple Silicon), so the module only
// updates the node's mode label.
type InstallIntelPluginModule struct {
	common.KubeModule
	Skip bool // conditional execution based on Intel GPU detection
}

func (m *InstallIntelPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallIntelPluginModule) Init() {
	m.Name = "InstallIntelPlugin"

	// update node with Intel GPU label
	updateNode := &task.LocalTask{
		Name:   "UpdateNodeIntelGPUInfo",
		Action: new(UpdateNodeIntelGPUInfo),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		updateNode,
	}
}
