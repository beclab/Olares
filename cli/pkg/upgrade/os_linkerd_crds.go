package upgrade

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type upgradeOSLinkerdCRDs struct {
	common.KubeAction
}

func (u *upgradeOSLinkerdCRDs) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get rest config: %s", err)
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceDefault)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	chartPath := path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-linkerd-crds")
	if !util.IsExist(chartPath) {
		return fmt.Errorf("os-linkerd-crds chart not exists")
	}
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSLinkerdCRDs, chartPath, "", common.NamespaceDefault, nil, true); err != nil {
		return err
	}
	return nil
}
