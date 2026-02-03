package pipelines

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

func EnableGpuNode() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpuenable.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
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
