package pipelines

import (
	"github.com/beclab/Olares/cli/cmd/ctl/options"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
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
