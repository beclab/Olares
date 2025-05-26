package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
)

func EnableGpuNode() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpuenable.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "EnableGpuNode",
		Modules: []module.Module{
			&gpu.NodeLabelingModule{},
		},
		Runtime: runtime,
	}

	return p.Start()

}
