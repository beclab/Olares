package plugins

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
)

type DeploySnapshotController struct {
	common.KubeAction
}

func (t *DeploySnapshotController) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var buildFilesDir = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir)
	var scrd = path.Join(buildFilesDir, "snapshot-controller", "crds", "snapshot.storage.k8s.io_volumesnapshot.yaml")
	var cmd = fmt.Sprintf("%s apply -f %s --force", kubectlpath, scrd)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("Install snapshot controller failed: %v", err)
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	var appName = common.ChartNameSnapshotController
	var appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, appName)

	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceKubeSystem)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var values = make(map[string]interface{})
	values["Release"] = map[string]string{
		"Namespace": common.NamespaceKubeSystem,
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, appName, appPath, "",
		common.NamespaceKubeSystem, values, false); err != nil {
		return err
	}

	return nil
}

type DeploySnapshotControllerModule struct {
	common.KubeModule
}

func (d *DeploySnapshotControllerModule) Init() {
	d.Name = "DeploySnapshotController"

	createSnapshotController := &task.RemoteTask{
		Name:  "CreateSnapshotController",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(DeploySnapshotController),
		Retry:    2,
		Parallel: false,
	}

	d.Tasks = []task.Interface{
		createSnapshotController,
	}
}
