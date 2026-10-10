package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	corev1 "k8s.io/api/core/v1"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
)

type InstallAccount struct {
	common.KubeAction
}

func (t *InstallAccount) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	ns := corev1.NamespaceDefault
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var accountPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "account")

	if !util.IsExist(accountPath) {
		return fmt.Errorf("account not exists")
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, accountPath, "", ns, nil, false); err != nil {
		return err
	}

	return nil
}

type InstallAccountModule struct {
	common.KubeModule
}

func (m *InstallAccountModule) Init() {
	logger.InfoInstallationProgress("Installing account ...")
	m.Name = "InstallAccount"

	installAccount := &task.LocalTask{
		Name:   "InstallAccount",
		Action: &InstallAccount{},
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		installAccount,
	}
}
