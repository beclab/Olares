package plugins

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CreateKsCore struct {
	common.KubeAction
}

func (t *CreateKsCore) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	var appKsCoreName = common.ChartNameKsCore
	var appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, appKsCoreName)

	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceKubesphereSystem)
	if err != nil {
		return err
	}

	var values = make(map[string]interface{})
	values["Release"] = map[string]string{
		"Namespace":    common.NamespaceKubesphereSystem,
		"ReplicaCount": fmt.Sprintf("%d", 1),
	}
	if err := utils.UpgradeCharts(context.Background(), actionConfig, settings, appKsCoreName,
		appPath, "", common.NamespaceKubesphereSystem, values, false); err != nil {
		logger.Errorf("failed to install %s chart: %v", appKsCoreName, err)
		return err
	}

	return nil
}

type DeployKsCoreModule struct {
	common.KubeModule
}

func (m *DeployKsCoreModule) Init() {
	m.Name = "DeployKsCore"

	createKsCore := &task.RemoteTask{
		Name:  "CreateKsCore",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CreateKsCore),
		Parallel: false,
		Retry:    10,
		Delay:    10 * time.Second,
	}

	m.Tasks = []task.Interface{
		createKsCore,
	}
}
