package intelgpu

import (
	"time"

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

	// nfd.yaml installs CRDs; keep it in its own task and retry so the CRDs are
	// established before the dependent CR / plugin manifests are applied.
	applyNFD := &task.LocalTask{
		Name:   "ApplyIntelNFD",
		Action: new(ApplyIntelNFD),
		Retry:  5,
		Delay:  5 * time.Second,
	}

	// depends on the NodeFeatureRule CRD created by applyNFD, retry until ready
	applyNodeFeatureRules := &task.LocalTask{
		Name:   "ApplyIntelNodeFeatureRules",
		Action: new(ApplyIntelNodeFeatureRules),
		Retry:  10,
		Delay:  6 * time.Second,
	}

	applyGPUPlugin := &task.LocalTask{
		Name:   "ApplyIntelGPUPlugin",
		Action: new(ApplyIntelGPUPlugin),
		Retry:  3,
		Delay:  5 * time.Second,
	}

	checkGPUPlugin := &task.LocalTask{
		Name:   "CheckIntelGpu",
		Action: new(CheckIntelGpu),
		Retry:  50,
		Delay:  10 * time.Second,
	}

	m.Tasks = []task.Interface{
		updateNode,
		applyNFD,
		applyNodeFeatureRules,
		applyGPUPlugin,
		checkGPUPlugin,
	}
}
