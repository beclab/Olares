package pipelines

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
)

func DisableConflictingGPUDrivers() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("gpu-disable-conflicts.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "DisableConflictingGPUDrivers",
		Modules: []module.Module{
			&gpu.DisableConflictingGPUDriversModule{},
		},
		Runtime: runtime,
	}

	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	return p.Start(context.Background())
}
