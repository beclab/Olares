package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func NewDownloadWizard(runtime *common.KubeRuntime) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&terminus.InstallWizardDownloadModule{Version: runtime.Arg.OlaresVersion, DownloadCdnUrl: runtime.Arg.DownloadCdnUrl},
	}

	return &pipeline.Pipeline{
		Name:    "Download Installation Wizard",
		Modules: m,
		Runtime: runtime,
	}
}
