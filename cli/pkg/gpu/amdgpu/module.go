package amdgpu

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

// LabelNodeModule only (re-)applies the AMD mode label(s), without touching the
// device plugin or ROCm. `gpu enable` uses it to restore labels that
// `gpu disable` wiped; the action is a no-op without a supported AMD GPU or ROCm.
type LabelNodeModule struct {
	common.KubeModule
}

func (m *LabelNodeModule) Init() {
	m.Name = "LabelAMDGPUNode"

	labelNode := &task.LocalTask{
		Name:    "UpdateNodeAMDInfo",
		Prepare: new(gpu.CurrentNodeInK8s),
		Action:  new(UpdateNodeAMDInfo),
		Retry:   1,
	}

	m.Tasks = []task.Interface{
		labelNode,
	}
}

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
// On discrete AMD GPUs (amd-gpu) it also installs the device-metrics-exporter chart.
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
	updateNode := &task.LocalTask{
		Name:   "UpdateNodeAMDInfo",
		Action: new(UpdateNodeAMDInfo),
		Retry:  1,
	}

	installPlugin := &task.LocalTask{
		Name:   "InstallAmdPlugin",
		Action: new(InstallAmdPlugin),
		Retry:  1,
	}

	checkGpuState := &task.LocalTask{
		Name: "CheckAmdGPUState",
		Prepare: &prepare.PrepareCollection{
			new(RocmInstalled),
		},
		Action: new(CheckAmdGpuStatus),
		Retry:  50,
		Delay:  10 * time.Second,
	}

	m.Tasks = append([]task.Interface{
		updateNode,
		installPlugin,
		checkGpuState,
	}, MetricsExporterTasks()...)
}

// MetricsExporterTasks installs/upgrades the device-metrics-exporter chart. It
// is gated by HasAmdDiscreteGPU so it only runs on discrete AMD GPU (amd-gpu)
// nodes, and is safe to run standalone from the upgrade path.
func MetricsExporterTasks() []task.Interface {
	installMetricsExporter := &task.LocalTask{
		Name:    "InstallDeviceMetricsExporter",
		Prepare: new(HasAmdDiscreteGPU),
		Action:  new(InstallDeviceMetricsExporter),
		Retry:   3,
		Delay:   5 * time.Second,
	}

	checkMetricsExporter := &task.LocalTask{
		Name:    "CheckDeviceMetricsExporter",
		Prepare: new(HasAmdDiscreteGPU),
		Action:  new(CheckDeviceMetricsExporter),
		Retry:   30,
		Delay:   10 * time.Second,
	}

	return []task.Interface{
		installMetricsExporter,
		checkMetricsExporter,
	}
}
