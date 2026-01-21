package pipelines

import (
	"strings"

	"github.com/beclab/Olares/cli/pkg/amdgpu"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/action"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/core/task"
)

type singleTaskModule struct {
	common.KubeModule
	name string
	act  action.Action
}

func (m *singleTaskModule) Init() {
	m.Name = m.name
	m.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   m.name,
			Action: m.act,
		},
	}
}

func AmdGpuInstall() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("amdgpuinstall.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}
	p := &pipeline.Pipeline{
		Name:    "InstallAMDGPUDrivers",
		Runtime: runtime,
		Modules: []module.Module{
			&amdgpu.InstallAmdRocmModule{},
		},
	}
	return p.Start()
}

func AmdGpuUninstall() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("amdgpuuninstall.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}
	p := &pipeline.Pipeline{
		Name:    "UninstallAMDGPUDrivers",
		Runtime: runtime,
		Modules: []module.Module{
			&singleTaskModule{name: "AmdgpuUninstall", act: new(amdgpu.AmdgpuUninstallAction)},
		},
	}
	return p.Start()
}

func AmdGpuStatus() error {
	arg := common.NewArgument()
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}
	runtime.SetRunner(
		&connector.Runner{
			Host: &connector.BaseHost{
				Name: common.LocalHost,
				Arch: runtime.GetSystemInfo().GetOsArch(),
				Os:   runtime.GetSystemInfo().GetOsType(),
			},
		},
	)
	amdModel, _ := runtime.GetRunner().SudoCmd("lspci | grep -iE 'VGA|3D|Display' | grep -iE 'AMD|ATI' | head -1 || true", false, false)
	drvVer, _ := runtime.GetRunner().SudoCmd("modinfo amdgpu 2>/dev/null | awk -F': ' '/^version:/{print $2}' || true", false, false)
	rocmVer, _ := runtime.GetRunner().SudoCmd("cat /opt/rocm/.info/version 2>/dev/null || true", false, false)

	if strings.TrimSpace(amdModel) != "" {
		logger.Infof("AMD GPU: %s", strings.TrimSpace(amdModel))
	} else {
		logger.Info("AMD GPU: not detected")
	}
	if strings.TrimSpace(drvVer) != "" {
		logger.Infof("AMDGPU driver %s", strings.TrimSpace(drvVer))
	} else {
		logger.Info("AMDGPU driver version: unknown")
	}
	if strings.TrimSpace(rocmVer) != "" {
		logger.Infof("ROCm version: %s", strings.TrimSpace(rocmVer))
	} else {
		logger.Info("ROCm version: not installed")
	}
	return nil
}
