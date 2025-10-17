package download

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/download"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
)

func NewDownloadPackage(mainifest string, runtime *common.KubeRuntime) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&download.PackageDownloadModule{Manifest: mainifest, BaseDir: runtime.GetBaseDir(), CDNService: runtime.Arg.OlaresCDNService},
	}

	return &pipeline.Pipeline{
		Name:    "Download Installation Package",
		Modules: m,
		Runtime: runtime,
	}
}
