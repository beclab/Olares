package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/windows"
)

type windowsInstallPhaseBuilder struct {
	runtime *common.KubeRuntime
}

func (w *windowsInstallPhaseBuilder) build() []module.Module {
	return []module.Module{
		&windows.InstallWSLModule{},
		&windows.InstallWSLUbuntuDistroModule{},
		&windows.GetDiskPartitionModule{},
		&windows.MoveDistroModule{},
		&windows.ConfigWslModule{},
		&windows.InstallTerminusModule{},
	}
}

type windowsUninstallPhaseBuilder struct {
	runtime *common.KubeRuntime
}

func (w *windowsUninstallPhaseBuilder) build() []module.Module {
	return []module.Module{
		&windows.UninstallOlaresModule{},
	}
}
