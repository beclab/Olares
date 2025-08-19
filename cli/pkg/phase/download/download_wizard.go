package download

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
	"github.com/beclab/Olares/cli/pkg/terminus"
)

func NewDownloadWizard(runtime *common.KubeRuntime, urlOverride, releaseID string) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&terminus.InstallWizardDownloadModule{Version: runtime.Arg.OlaresVersion, DownloadCdnUrl: runtime.Arg.DownloadCdnUrl, UrlOverride: urlOverride, ReleaseID: releaseID},
	}

	return &pipeline.Pipeline{
		Name:    "Download Installation Wizard",
		Modules: m,
		Runtime: runtime,
	}
}
