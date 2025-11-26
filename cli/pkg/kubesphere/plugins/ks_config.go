package plugins

import (
	"context"
	"path"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
)

type ApplyKsConfigManifests struct {
	common.KubeAction
}

func (t *ApplyKsConfigManifests) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	var appKsConfigName = common.ChartNameKsConfig
	var appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, appKsConfigName)

	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceKubesphereSystem)
	if err != nil {
		return err
	}

	var values = make(map[string]interface{})
	if err := utils.UpgradeCharts(context.Background(), actionConfig, settings, appKsConfigName,
		appPath, "", common.NamespaceKubesphereSystem, values, false); err != nil {
		logger.Errorf("failed to install %s chart: %v", appKsConfigName, err)
		return err
	}
	return nil
}
