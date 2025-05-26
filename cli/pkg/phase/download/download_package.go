package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/download"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func NewDownloadPackage(mainifest string, runtime *common.KubeRuntime) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&download.PackageDownloadModule{Manifest: mainifest, BaseDir: runtime.GetBaseDir(), DownloadCdnUrl: runtime.Arg.DownloadCdnUrl},
	}

	return &pipeline.Pipeline{
		Name:    "Download Installation Package",
		Modules: m,
		Runtime: runtime,
	}
}
