package plugins

import (
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/utils"
)

type CopyEmbedFiles struct {
	common.KubeAction
}

func (t *CopyEmbedFiles) Execute(runtime connector.Runtime) error {
	var dst = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir)
	return utils.CopyEmbed(assets, ".", dst)
}
