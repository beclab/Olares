package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/download"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func NewCheckDownload(mainifest string, runtime *common.KubeRuntime) *pipeline.Pipeline {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&download.CheckDownloadModule{Manifest: mainifest, BaseDir: runtime.GetBaseDir()},
	}

	return &pipeline.Pipeline{
		Name:    "Check Downloaded Olares Installation Package",
		Modules: m,
		Runtime: runtime,
	}
}
