package gpu

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
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

	// currently, this is only used to skip the nvidia-smi check after driver upgrade
	// because the nvidia-smi will not work after upgrade (Failed to initialize NVML: Driver/library version mismatch)
	// otherwise, always check the driver is running properly after installation to fail early and avoid other issues
	SkipNVMLCheckAfterInstall bool
}

func (m *InstallDriversModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDriversModule) Init() {
	m.Name = "InstallGPUDriver"

	installCudaDeps := &task.RemoteTask{
		Name:  "InstallCudaKeyRing",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(CudaNotInstalled),
			&NvidiaGraphicsCard{ExitOnNotFound: m.FailOnNoInstallation},
		},
		Action: &InstallCudaDeps{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	installCudaDriver := &task.RemoteTask{ // not for WSL
		Name:  "InstallNvidiaDriver",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(CudaNotInstalled),
			&NvidiaGraphicsCard{ExitOnNotFound: m.FailOnNoInstallation},
		},
		Action:   &InstallCudaDriver{SkipNVMLCheckAfterInstall: m.SkipNVMLCheckAfterInstall},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installCudaDeps,
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

	updateCudaSource := &task.RemoteTask{
		Name:  "UpdateNvidiaToolkitSource",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Action: &UpdateNvidiaContainerToolkitSource{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Prepare:  &prepareCollection,
		Parallel: false,
		Retry:    1,
	}

	installNvidiaContainerToolkit := &task.RemoteTask{
		Name:     "InstallNvidiaToolkit",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepareCollection,
		Action:   new(InstallNvidiaContainerToolkit),
		Parallel: false,
		Retry:    1,
	}

	configureContainerdRuntime := &task.RemoteTask{
		Name:     "ConfigureContainerdRuntime",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  &prepareCollection,
		Action:   new(ConfigureContainerdRuntime),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		updateCudaSource,
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

	patchK3sDriver := &task.RemoteTask{
		Name:  "PatchK3sDriver",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(PatchK3sDriver),
		Parallel: false,
		Retry:    1,
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

	restartContainerd := &task.RemoteTask{
		Name:  "RestartContainerd",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(ContainerdInstalled),
		},
		Action:   new(RestartContainerd),
		Parallel: false,
		Retry:    1,
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
	updateNode := &task.RemoteTask{
		Name:  "UpdateNode",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(UpdateNodeLabels),
		Parallel: false,
		Retry:    1,
	}

	installPlugin := &task.RemoteTask{
		Name:  "InstallPlugin",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(InstallPlugin),
		Parallel: false,
		Retry:    1,
	}

	checkGpuState := &task.RemoteTask{
		Name:  "CheckGPUState",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CudaInstalled),
		},
		Action:   new(CheckGpuStatus),
		Parallel: false,
		Retry:    50,
		Delay:    10 * time.Second,
	}

	installGPUShared := &task.RemoteTask{
		Name:  "InstallGPUShared",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(GPUSharePrepare),
		},
		Action:   new(InstallGPUShared),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		updateNode,
		installPlugin,
		checkGpuState,
		installGPUShared,
	}
}

type GetCudaVersionModule struct {
	common.KubeModule
}

func (g *GetCudaVersionModule) Init() {
	g.Name = "GetCudaVersion"

	getCudaVersion := &task.LocalTask{
		Name:   "GetCudaVersion",
		Action: new(GetCudaVersion),
	}

	g.Tasks = []task.Interface{
		getCudaVersion,
	}
}

type NodeLabelingModule struct {
	common.KubeModule
}

func (l *NodeLabelingModule) Init() {
	l.Name = "NodeLabeling"

	updateNode := &task.RemoteTask{
		Name:  "UpdateNode",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CudaInstalled),
			new(K8sNodeInstalled),
		},
		Action:   new(UpdateNodeLabels),
		Parallel: false,
		Retry:    1,
	}

	restartPlugin := &task.RemoteTask{
		Name:  "RestartPlugin",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(CudaInstalled),
			new(K8sNodeInstalled),
		},
		Action:   new(RestartPlugin),
		Parallel: false,
		Retry:    1,
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

	removeNode := &task.RemoteTask{
		Name:  "RemoveNodeLabels",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(K8sNodeInstalled),
		},
		Action:   new(RemoveNodeLabels),
		Parallel: false,
		Retry:    1,
	}

	restartPlugin := &task.RemoteTask{
		Name:  "RestartPlugin",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(K8sNodeInstalled),
			new(GpuDevicePluginInstalled),
		},
		Action:   new(RestartPlugin),
		Parallel: false,
		Retry:    1,
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

	uninstallCuda := &task.RemoteTask{
		Name:  "UninstallCuda",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(UninstallNvidiaDrivers),
		Parallel: false,
		Retry:    1,
	}

	removeRuntime := &task.RemoteTask{
		Name:  "RemoveRuntime",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(ContainerdInstalled),
		},
		Action: new(RemoveContainerRuntimeConfig),
	}

	l.Tasks = []task.Interface{
		uninstallCuda,
		removeRuntime,
	}

}

type ExitIfNoDriverUpgradeNeededModule struct {
	common.KubeModule
}

func (l *ExitIfNoDriverUpgradeNeededModule) Init() {
	l.Tasks = append(l.Tasks, &task.LocalTask{
		Action: new(ExitIfNoDriverUpgradeNeeded),
	})

}
