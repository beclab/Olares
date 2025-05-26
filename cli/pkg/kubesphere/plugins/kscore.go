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
	"bytetrade.io/web3os/installer/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CreateKsCore struct {
	common.KubeAction
}

func (t *CreateKsCore) Execute(runtime connector.Runtime) error {
	//var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	//if err != nil {
	//	return fmt.Errorf("kubectl not found")
	//}

	//var cmd = fmt.Sprintf("%s get pod -n %s -l 'app=redis,tier=database,version=redis-4.0' -o jsonpath='{.items[0].status.phase}'", kubectlpath,
	//	common.NamespaceKubesphereSystem)
	//rphase, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	//if rphase != "Running" {
	//	return fmt.Errorf("Redis State %s", rphase)
	//}

	masterNumIf, ok := t.PipelineCache.Get(common.CacheMasterNum)
	if !ok || masterNumIf == nil {
		return fmt.Errorf("failed to get master num")
	}
	masterNum := masterNumIf.(int64)

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
		"ReplicaCount": fmt.Sprintf("%d", masterNum),
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
			new(NotEqualDesiredVersion),
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
