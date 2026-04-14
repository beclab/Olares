package amdgpu

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

// InstallAmdContainerToolkitModule installs AMD container toolkit on supported Ubuntu if ROCm is installed.
type InstallAmdContainerToolkitModule struct {
	common.KubeModule
	Skip          bool // conditional execution based on ROCm detection
	SkipRocmCheck bool
}

func (m *InstallAmdContainerToolkitModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallAmdContainerToolkitModule) Init() {
	m.Name = "InstallAmdContainerToolkit"
	if m.IsSkip() {
		return
	}

	prepareCollection := prepare.PrepareCollection{}
	if !m.SkipRocmCheck {
		prepareCollection = append(prepareCollection, new(RocmInstalled))
	}

	updateAmdSource := &task.RemoteTask{
		Name:     "UpdateAmdContainerToolkitSource",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(UpdateAmdContainerToolkitSource),
		Prepare:  &prepareCollection,
		Parallel: false,
		Retry:    1,
	}

	installAmdContainerToolkit := &task.RemoteTask{
		Name:     "InstallAmdContainerToolkit",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepareCollection,
		Action:   new(InstallAmdContainerToolkit),
		Parallel: false,
		Retry:    1,
	}

	generateAndValidateCDI := &task.RemoteTask{
		Name:     "GenerateAndValidateAmdCDI",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepareCollection,
		Action:   new(GenerateAndValidateAmdCDI),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		updateAmdSource,
		installAmdContainerToolkit,
		generateAndValidateCDI,
	}
}

// InstallAmdPluginModule installs AMD GPU device plugin on Kubernetes.
type InstallAmdPluginModule struct {
	common.KubeModule
	Skip bool // conditional execution based on GPU enablement
}

func (m *InstallAmdPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallAmdPluginModule) Init() {
	m.Name = "InstallAmdPlugin"

	// update node with AMD GPU labels
	updateNode := &task.RemoteTask{
		Name:  "UpdateNodeAmdGPUInfo",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(UpdateNodeAmdGPUInfo),
		Parallel: false,
		Retry:    1,
	}

	installPlugin := &task.RemoteTask{
		Name:  "InstallAmdPlugin",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(InstallAmdPlugin),
		Parallel: false,
		Retry:    1,
	}

	checkGpuState := &task.RemoteTask{
		Name:  "CheckAmdGPUState",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(RocmInstalled),
		},
		Action:   new(CheckAmdGpuStatus),
		Parallel: false,
		Retry:    50,
		Delay:    10 * time.Second,
	}

	m.Tasks = []task.Interface{
		updateNode,
		installPlugin,
		checkGpuState,
	}
}
