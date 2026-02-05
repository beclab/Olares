package pipelines

import (
	"fmt"
	"os"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func MasterInfoPipeline() error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: Only Linux nodes can be added to an Olares cluster!")
		os.Exit(1)
	}
	arg.SetConsoleLog("masterinfo.log", true)

	if err := arg.MasterHostConfig.Validate(); err != nil {
		return fmt.Errorf("invalid master host config: %w", err)
	}

	runtime, err := common.NewKubeRuntime(*arg)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	p := &pipeline.Pipeline{
		Name:    "Get Master Info",
		Modules: []module.Module{&terminus.GetMasterInfoModule{Print: true}},
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
