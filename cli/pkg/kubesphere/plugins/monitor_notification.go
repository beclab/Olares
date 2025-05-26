package plugins

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CreateMonitorNotification struct {
	common.KubeAction
}

func (t *CreateMonitorNotification) Execute(runtime connector.Runtime) error {
	nodeNumIf, ok := t.PipelineCache.Get(common.CacheNodeNum)
	if !ok || nodeNumIf == nil {
		return fmt.Errorf("node get failed")
	}
	var replicas int
	var nodeNum = nodeNumIf.(int64)
	if nodeNum < 3 {
		replicas = 1
	} else {
		replicas = 2
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	var appName = common.ChartNameMonitorNotification
	var appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "ks-monitor", "notification-manager")

	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceKubesphereMonitoringSystem)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var values = make(map[string]interface{})
	values["Release"] = map[string]string{
		"Namespace": common.NamespaceKubesphereMonitoringSystem,
		"Replicas":  fmt.Sprintf("%d", replicas),
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, appName, appPath, "", common.NamespaceKubesphereMonitoringSystem, values, false); err != nil {
		return err
	}

	return nil
}

// +

type CreateNotificationModule struct {
	common.KubeModule
}

func (m *CreateNotificationModule) Init() {
	m.Name = "CreateMonitorNotification"

	createMonitorNotifiction := &task.RemoteTask{
		Name:  "CreateNotification",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
			new(common.GetNodeNum),
		},
		Action:   new(CreateMonitorNotification),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		createMonitorNotifiction,
	}
}
