package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func StartPreCheckPipeline(opt *options.PreCheckOptions) error {
	terminusVersion := opt.Version

	var arg = common.NewArgument()
	arg.SetOlaresVersion(terminusVersion)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("precheck.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name: "PreCheck",
		Modules: []module.Module{
			&precheck.RunPrechecksModule{},
		},
		Runtime: runtime,
	}
	return p.Start()

}
