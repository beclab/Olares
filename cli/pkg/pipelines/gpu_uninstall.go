package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"fmt"
	"os"
)

func UninstallGpuDrivers() error {

	arg := common.NewArgument()
	if arg.SystemInfo.IsWsl() {
		fmt.Println("WSL's GPU driver is managed by Windows, does not support uninstalling from inside.")
		os.Exit(1)
	}
	arg.SetConsoleLog("gpuuninstall.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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
