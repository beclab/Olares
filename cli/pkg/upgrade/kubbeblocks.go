package upgrade

import (
	"fmt"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/pkg/errors"
)

type upgradeKubeblocksComponents struct {
	common.KubeAction
}

func (u *upgradeKubeblocksComponents) Execute(runtime connector.Runtime) error {
	kubectl := "kubectl"

	kubeblocksCRDsPath := filepath.Join(runtime.GetInstallerDir(), "wizard/config/kubeblocks/kubeblocks_crds.yaml")

	applyCRDsCmd := fmt.Sprintf("%s apply -f %s --server-side", kubectl, kubeblocksCRDsPath)
	_, err := runtime.GetRunner().SudoCmd(applyCRDsCmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to apply kubeblocks_crds.yaml")
	}
	return nil

}
