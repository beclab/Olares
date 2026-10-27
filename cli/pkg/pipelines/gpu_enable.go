package pipelines

import (
	"context"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/gpu"
	"github.com/beclab/Olares/cli/pkg/gpu/amdgpu"
	"github.com/beclab/Olares/cli/pkg/gpu/intelgpu"
	"github.com/beclab/Olares/cli/pkg/gpu/mtgpu"
)

func EnableGpuNode() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpuenable.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	// DisableGpuNode wipes the per-mode labels of every accelerator, so enabling
	// has to relabel all of them to be its inverse. Each vendor module is
	// idempotent and no-ops when its hardware isn't present, so this is also what
	// makes `gpu enable` meaningful on a machine with no NVIDIA card at all.
	p := &pipeline.Pipeline{
		Name: "EnableGpuNode",
		Modules: []module.Module{
			&gpu.NodeLabelingModule{},
			&intelgpu.LabelNodeModule{},
			&amdgpu.LabelNodeModule{},
			&mtgpu.LabelNodeModule{},
		},
		Runtime: runtime,
	}

	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	return p.Start(context.Background())

}
