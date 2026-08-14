package intelgpu

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

// LabelNodeModule only (re-)applies the Intel mode labels, without touching the
// device plugin or the host drivers. `gpu enable` uses it to restore the labels
// that `gpu disable` wiped; the action is a no-op on nodes without an Intel GPU.
type LabelNodeModule struct {
	common.KubeModule
}

func (m *LabelNodeModule) Init() {
	m.Name = "LabelIntelGPUNode"

	labelNode := &task.LocalTask{
		Name:    "LabelIntelGPUs",
		Prepare: new(gpu.CurrentNodeInK8s),
		Action:  new(LabelIntelGPUs),
		Retry:   1,
	}

	m.Tasks = []task.Interface{
		labelNode,
	}
}

// InstallIntelPluginModule installs/upgrades the Intel GPU stack on the node
// (mode labels, optional dGPU host drivers, NFD, device plugin, xpumd).
type InstallIntelPluginModule struct {
	common.KubeModule
	Skip bool // conditional execution based on Intel GPU detection
}

func (m *InstallIntelPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallIntelPluginModule) Init() {
	m.Name = "InstallIntelPlugin"
	m.Tasks = PluginTasks()
}

// PluginTasks is the shared install/upgrade task list for the Intel GPU stack
// (labels, optional dGPU drivers, NFD, device plugin, xpumd). Each general
// task is gated by HasAnyIntelGPU so the same list is safe to run from upgrade
// on nodes without an Intel GPU; dGPU-only steps keep HasQualifyingIntelDGPU.
func PluginTasks() []task.Interface {
	// label the node with the intel / intel-gpu modes based on the detected
	// integrated / discrete Intel GPUs that meet the kernel requirement
	labelNode := &task.LocalTask{
		Name:    "LabelIntelGPUs",
		Prepare: new(HasAnyIntelGPU),
		Action:  new(LabelIntelGPUs),
		Retry:   1,
	}

	// install the discrete-GPU host driver stack; only runs when a discrete Intel
	// GPU is present, in-tree and kernel-supported
	installDGPUDrivers := &task.LocalTask{
		Name:    "InstallIntelDGPUDrivers",
		Action:  new(InstallIntelDGPUDrivers),
		Prepare: new(HasQualifyingIntelDGPU),
		Retry:   1,
	}

	// nfd.yaml installs CRDs; keep it in its own task and retry so the CRDs are
	// established before the dependent CR / plugin manifests are applied.
	applyNFD := &task.LocalTask{
		Name:    "ApplyIntelNFD",
		Prepare: new(HasAnyIntelGPU),
		Action:  new(ApplyIntelNFD),
		Retry:   5,
		Delay:   5 * time.Second,
	}

	// depends on the NodeFeatureRule CRD created by applyNFD, retry until ready
	applyNodeFeatureRules := &task.LocalTask{
		Name:    "ApplyIntelNodeFeatureRules",
		Prepare: new(HasAnyIntelGPU),
		Action:  new(ApplyIntelNodeFeatureRules),
		Retry:   10,
		Delay:   6 * time.Second,
	}

	applyGPUPlugin := &task.LocalTask{
		Name:    "ApplyIntelGPUPlugin",
		Prepare: new(HasAnyIntelGPU),
		Action:  new(ApplyIntelGPUPlugin),
		Retry:   3,
		Delay:   5 * time.Second,
	}

	checkGPUPlugin := &task.LocalTask{
		Name:    "CheckIntelGpu",
		Prepare: new(HasAnyIntelGPU),
		Action:  new(CheckIntelGpu),
		Retry:   50,
		Delay:   10 * time.Second,
	}

	// xpumd needs gpu.intel.com/monitoring from the device plugin; install only
	// on nodes with a qualifying discrete Intel GPU.
	installXpumd := &task.LocalTask{
		Name:    "InstallXpumd",
		Action:  new(InstallXpumd),
		Prepare: new(HasQualifyingIntelDGPU),
		Retry:   3,
		Delay:   5 * time.Second,
	}

	checkXpumd := &task.LocalTask{
		Name:    "CheckXpumd",
		Action:  new(CheckXpumd),
		Prepare: new(HasQualifyingIntelDGPU),
		Retry:   30,
		Delay:   10 * time.Second,
	}

	return []task.Interface{
		labelNode,
		installDGPUDrivers,
		applyNFD,
		applyNodeFeatureRules,
		applyGPUPlugin,
		checkGPUPlugin,
		installXpumd,
		checkXpumd,
	}
}
