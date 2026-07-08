package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type InstallOSLinkerdCRDs struct {
	common.KubeAction
}

func (t *InstallOSLinkerdCRDs) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	ns := corev1.NamespaceDefault
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var chartPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-linkerd-crds")
	if !util.IsExist(chartPath) {
		return fmt.Errorf("os-linkerd-crds chart not exists")
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSLinkerdCRDs, chartPath, "", ns, nil, false); err != nil {
		return err
	}

	return nil
}
