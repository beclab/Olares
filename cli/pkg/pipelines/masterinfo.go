package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
	"fmt"
	"github.com/pkg/errors"
	"os"
)

func MasterInfoPipeline(opts *options.MasterInfoOptions) error {
	arg := common.NewArgument()
	if !arg.SystemInfo.IsLinux() {
		fmt.Println("error: Only Linux nodes can be added to an Olares cluster!")
		os.Exit(1)
	}
	arg.SetBaseDir(opts.BaseDir)

	if err := arg.LoadMasterHostConfigIfAny(); err != nil {
		return errors.Wrap(err, "failed to load master host config")
	}
	arg.SetMasterHostOverride(opts.MasterHostConfig)
	if err := arg.MasterHostConfig.Validate(); err != nil {
		return fmt.Errorf("invalid master host config: %w", err)
	}
	arg.SetConsoleLog("masterinfo.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
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
