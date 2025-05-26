package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/gpu"
)

func GpuDriverStatus() error {
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

	// get gpu status
	if err := new(gpu.PrintGpuStatus).Execute(runtime); err != nil {
		return err
	}

	// get device plugin status
	if err := new(gpu.PrintPluginsStatus).Execute(runtime); err != nil {
		return err
	}

	return nil

}
