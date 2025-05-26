package pipelines

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func StartOlares() error {
	arg := common.NewArgument()
	arg.SetConsoleLog("start.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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

	return p.Start()
}

func StopOlares(timeout, checkInterval time.Duration) error {
	arg := common.NewArgument()
	arg.SetConsoleLog("stop.log", true)
	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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

	return p.Start()
}
