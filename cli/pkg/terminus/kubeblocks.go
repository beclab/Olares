package terminus

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/pkg/errors"
)

type InstallKubeblocks struct {
	common.KubeAction
}

func (t *InstallKubeblocks) Execute(runtime connector.Runtime) error {
	kubectl := "kubectl"

	kubeblocksCRDsPath := filepath.Join(runtime.GetInstallerDir(), "wizard/config/kubeblocks/kubeblocks_crds.yaml")

	applyCRDsCmd := fmt.Sprintf("%s apply -f %s --server-side", kubectl, kubeblocksCRDsPath)
	_, err := runtime.GetRunner().SudoCmd(applyCRDsCmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to apply kubeblocks_crds.yaml")
	}
	return nil
}

type InstallKubeblocksModule struct {
	common.KubeModule
}

func (m *InstallKubeblocksModule) Init() {
	logger.InfoInstallationProgress("Installing KubeBlocks ...")
	m.Name = "InstallKubeblocksModule"

	installKubeblocks := &task.LocalTask{
		Name:   "InstallKubeblocks",
		Action: &InstallKubeblocks{},
		Retry:  3,
		Delay:  10 * time.Second,
	}

	m.Tasks = []task.Interface{
		installKubeblocks,
	}
}
