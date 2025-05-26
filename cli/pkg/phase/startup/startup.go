package startup

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
)

func GetMachineInfo() error {
	runtime, err := common.NewKubeRuntime(common.AllInOne, common.Argument{})
	if err != nil {
		return err
	}

	m := []module.Module{
		&plugins.CopyEmbed{},
	}

	p := pipeline.Pipeline{
		Name:    "Startup",
		Modules: m,
		Runtime: runtime,
	}

	return p.Start()
}
