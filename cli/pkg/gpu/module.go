package gpu

import (
	"time"

	"github.com/beclab/Olares/cli/pkg/container"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/manifest"
)

type InstallDriversModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool // enableGPU && ubuntuVersionSupport

	// log a failure message and then exit
	// instead of silently skip the jobs when:
	// 1. no card is found (which skips the driver installation)
	// 2. no driver is found (which skips the container toolkit installation)
	FailOnNoInstallation bool
}

func (m *InstallDriversModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDriversModule) Init() {
	m.Name = "InstallGPUDriver"

	installCudaDriver := &task.LocalTask{
		Name: "InstallNvidiaDriver",
		Prepare: &prepare.PrepareCollection{
			new(CudaNotInstalled),
			&NvidiaGraphicsCard{ExitOnNotFound: m.FailOnNoInstallation},
		},
		Action: &InstallCudaDriver{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		installCudaDriver,
	}
}

type InstallContainerToolkitModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip          bool // enableGPU && ubuntuVersionSupport
	SkipCudaCheck bool
}

func (m *InstallContainerToolkitModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallContainerToolkitModule) Init() {
	m.Name = "InstallContainerToolkit"
	prepareCollection := prepare.PrepareCollection{
		new(ContainerdInstalled),
	}
	if !m.SkipCudaCheck {
		prepareCollection = append(prepareCollection, new(CudaInstalled))
	}

	installNvidiaContainerToolkit := &task.LocalTask{
		Name:    "InstallNvidiaToolkit",
		Prepare: &prepareCollection,
		Action: &InstallNvidiaContainerToolkit{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Retry: 1,
	}

	configureContainerdRuntime := &task.LocalTask{
		Name:    "ConfigureContainerdRuntime",
		Prepare: &prepareCollection,
		Action:  new(ConfigureContainerdRuntime),
		Retry:   1,
	}

	m.Tasks = []task.Interface{
		installNvidiaContainerToolkit,
		configureContainerdRuntime,
	}

}

type RestartK3sServiceModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *RestartK3sServiceModule) IsSkip() bool {
	return m.Skip
}

func (m *RestartK3sServiceModule) Init() {
	m.Name = "RestartK3sService"

	patchK3sDriver := &task.LocalTask{
		Name: "PatchK3sDriver",
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action: new(PatchK3sDriver),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		patchK3sDriver,
	}
}

type RestartContainerdModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *RestartContainerdModule) IsSkip() bool {
	return m.Skip
}

func (m *RestartContainerdModule) Init() {
	m.Name = "RestartContainerd"

	restartContainerd := &task.LocalTask{
		Name: "RestartContainerd",
		Prepare: &prepare.PrepareCollection{
			new(ContainerdInstalled),
		},
		Action: new(container.RestartContainerd),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		restartContainerd,
	}
}

type InstallPluginModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallPluginModule) Init() {
	m.Name = "InstallPlugin"

	// update node with gpu labels, to make plugins enabled
	updateNode := &task.LocalTask{
		Name:   "UpdateNodeGPUInfo",
		Action: new(UpdateNodeGPUInfo),
		Retry:  1,
	}

	installPlugin := &task.LocalTask{
		Name:   "InstallGPUPlugin",
		Action: new(InstallPlugin),
		Retry:  1,
	}

	checkGpuState := &task.LocalTask{
		Name: "CheckGPUState",
		Prepare: &prepare.PrepareCollection{
			new(CudaInstalled),
		},
		Action: new(CheckGpuStatus),
		Retry:  50,
		Delay:  10 * time.Second,
	}

	m.Tasks = []task.Interface{
		updateNode,
		installPlugin,
		checkGpuState,
	}
}

type NodeLabelingModule struct {
	common.KubeModule
}

func (l *NodeLabelingModule) Init() {
	l.Name = "NodeLabeling"

	updateNode := &task.LocalTask{
		Name: "UpdateNode",
		Prepare: &prepare.PrepareCollection{
			new(CudaInstalled),
			new(CurrentNodeInK8s),
		},
		Action: new(UpdateNodeGPUInfo),
		Retry:  1,
	}

	restartPlugin := &task.LocalTask{
		Name: "RestartPlugin",
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CudaInstalled),
			new(CurrentNodeInK8s),
		},
		Action: new(RestartPlugin),
		Retry:  1,
	}

	l.Tasks = []task.Interface{
		updateNode,
		restartPlugin,
	}
}

type NodeUnlabelingModule struct {
	common.KubeModule
}

func (l *NodeUnlabelingModule) Init() {
	l.Name = "NodeUnlabeling"

	removeNode := &task.LocalTask{
		Name: "RemoveNodeLabels",
		Prepare: &prepare.PrepareCollection{
			new(CurrentNodeInK8s),
		},
		Action: new(RemoveNodeLabels),
		Retry:  1,
	}

	restartPlugin := &task.LocalTask{
		Name: "RestartPlugin",
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CurrentNodeInK8s),
			new(GpuDevicePluginInstalled),
		},
		Action: new(RestartPlugin),
		Retry:  1,
	}

	l.Tasks = []task.Interface{
		removeNode,
		restartPlugin,
	}
}

// NvidiaNodeUnlabelingModule clears only the NVIDIA-owned node labels. It is
// used by the driver uninstall path, where other accelerators on the same node
// (e.g. an Intel iGPU) are unaffected and must keep advertising their modes.
// The full wipe lives in NodeUnlabelingModule, used by `gpu disable`.
type NvidiaNodeUnlabelingModule struct {
	common.KubeModule
}

func (l *NvidiaNodeUnlabelingModule) Init() {
	l.Name = "NvidiaNodeUnlabeling"

	removeNode := &task.LocalTask{
		Name: "RemoveNvidiaNodeLabels",
		Prepare: &prepare.PrepareCollection{
			new(CurrentNodeInK8s),
		},
		Action: new(RemoveNvidiaNodeLabels),
		Retry:  1,
	}

	restartPlugin := &task.LocalTask{
		Name: "RestartPlugin",
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CurrentNodeInK8s),
			new(GpuDevicePluginInstalled),
		},
		Action: new(RestartPlugin),
		Retry:  1,
	}

	l.Tasks = []task.Interface{
		removeNode,
		restartPlugin,
	}
}

type UninstallCudaModule struct {
	common.KubeModule
}

func (l *UninstallCudaModule) Init() {
	l.Name = "UninstallCuda"

	uninstallCuda := &task.LocalTask{
		Name:   "UninstallCuda",
		Action: new(UninstallNvidiaDrivers),
		Retry:  1,
	}

	removeRuntime := &task.LocalTask{
		Name: "RemoveRuntime",
		Prepare: &prepare.PrepareCollection{
			new(ContainerdInstalled),
		},
		Action: new(RemoveContainerRuntimeConfig),
	}

	l.Tasks = []task.Interface{
		uninstallCuda,
		removeRuntime,
	}

}

type DisableConflictingGPUDriversModule struct {
	common.KubeModule
}

func (m *DisableConflictingGPUDriversModule) Init() {
	m.Name = "DisableConflictingGPUDrivers"

	writeBlacklist := &task.LocalTask{
		Name:   "WriteConflictingGPUDriverBlacklist",
		Action: new(WriteConflictingGPUDriverBlacklist),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		writeBlacklist,
	}
}
