package pipelines

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/spf13/viper"
)

func StartPreCheckPipeline() error {
	var arg = common.NewArgument()
	arg.SetOlaresVersion(viper.GetString(common.FlagVersion))
	arg.SetConsoleLog("precheck.log", true)

	runtime, err := common.NewKubeRuntime(*arg)
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
