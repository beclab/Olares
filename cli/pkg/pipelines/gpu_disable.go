package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
)

func DisableGpuNode() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpudisable.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "DisableGpuNode",
		Modules: []module.Module{
			&gpu.NodeUnlabelingModule{},
		},
		Runtime: runtime,
	}

	return p.Start()

}
