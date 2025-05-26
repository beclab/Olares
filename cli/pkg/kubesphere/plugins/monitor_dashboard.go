package plugins

import (
	"fmt"
	"path"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type InstallMonitorDashboardCrd struct {
	common.KubeAction
}

func (t *InstallMonitorDashboardCrd) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var p = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "ks-monitor", "monitoring-dashboard")
	var cmd = fmt.Sprintf("%s apply -f %s", kubectlpath, p)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

type CreateMonitorDashboardModule struct {
	common.KubeModule
}

func (m *CreateMonitorDashboardModule) Init() {
	m.Name = "CreateMonitorDashboardModule"

	installMonitorDashboardCrd := &task.RemoteTask{
		Name:  "InstallMonitorDashboardCrd",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(InstallMonitorDashboardCrd),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		installMonitorDashboardCrd,
	}

}
