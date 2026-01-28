package pipelines

import (
	"fmt"
	"os"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

func UninstallGpuDrivers() error {

	arg := common.NewArgument()
	if arg.SystemInfo.IsWsl() {
		fmt.Println("WSL's GPU driver is managed by Windows, does not support uninstalling from inside.")
		os.Exit(1)
	}
	arg.SetConsoleLog("gpuuninstall.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name:    "UninstallGpuDrivers",
		Runtime: runtime,
		Modules: []module.Module{
			&gpu.NodeUnlabelingModule{},
			&gpu.UninstallCudaModule{},
			&gpu.RestartContainerdModule{},
		},
	}

	return p.Start()

}
