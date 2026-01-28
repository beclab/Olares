package pipelines

import (
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

func DisableGpuNode() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpudisable.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
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
