package pipelines

import (
	"context"
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
	if arg.SystemInfo.IsGB10Chip() {
		fmt.Println("NVIDIA DGX Spark / GB10 systems ship and manage their own GPU driver stack as part of the OS, intertwined with vendor packages and an nvidia-branded kernel. We cannot reliably uninstall it without risking the factory NVIDIA driver environment, so the GPU driver uninstall is skipped on these systems.")
		return nil
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

	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	return p.Start(context.Background())

}
