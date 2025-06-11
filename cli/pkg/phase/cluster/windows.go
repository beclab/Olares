package cluster

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/windows"
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
