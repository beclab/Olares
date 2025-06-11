package download

import (
	"github.com/beclab/Olares/cli/pkg/bootstrap/download"
	"github.com/beclab/Olares/cli/pkg/bootstrap/precheck"
	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/module"
	"github.com/beclab/Olares/cli/pkg/core/pipeline"
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
