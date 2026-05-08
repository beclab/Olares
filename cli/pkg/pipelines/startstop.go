package pipelines

import (
	"context"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func StartOlares() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("start.log", true)
	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "StartOlares",
		Modules: []module.Module{
			&terminus.StartOlaresModule{},
		},
		Runtime: runtime,
	}

	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	return p.Start(context.Background())
}

func StopOlares(timeout, checkInterval time.Duration) error {
	arg := common.NewArgument()
	arg.SetConsoleLog("stop.log", true)
	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "StopOlares",
		Modules: []module.Module{
			&terminus.StopOlaresModule{
				Timeout:       timeout,
				CheckInterval: checkInterval,
			},
		},
		Runtime: runtime,
	}

	// TODO(ctx): plumb ctx in a follow-up; this entry point is not yet ctx-aware.
	return p.Start(context.Background())
}
