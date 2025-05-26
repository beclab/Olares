package plugins

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type EnableKsMonitorStatus struct {
	common.KubeAction
}

func (t *EnableKsMonitorStatus) Execute(runtime connector.Runtime) error {
	return nil
}

type CreatePrometheusComponent struct {
	common.KubeAction
	Component  string
	Force      string
	ServerSide string
}

func (t *CreatePrometheusComponent) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var f = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", t.Component)
	if !util.IsExist(f) {
		return fmt.Errorf("file %s not found", f)
	}

	var cmd = fmt.Sprintf("%s apply -f %s %s %s", kubectlpath, f, t.Force, t.ServerSide)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("create crd %s failed: %v", f, err)
		return err
	}

	return nil
}

type CreateOperator struct {
	common.KubeAction
}

func (t *CreateOperator) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var f = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "prometheus", "prometheus-operator")

	var crds []string
	var ress []string

	if !util.IsExist(f) {
		return fmt.Errorf("file %s not found", f)
	}

	if err := filepath.Walk(f, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		var fileName = info.Name()
		if strings.Contains(fileName, "CustomResourceDefinition.yaml") {
			crds = append(crds, path)
		} else {
			ress = append(ress, path)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk %s failed: %v", f, err)
	}

	for _, crd := range crds {
		var cmd = fmt.Sprintf("%s apply -f %s --force-conflicts --server-side", kubectlpath, crd)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			logger.Errorf("create crd %s failed: %v", crd, err)
			return err
		}
	}

	for _, res := range ress {
		var cmd = fmt.Sprintf("%s apply -f %s --force-conflicts --server-side", kubectlpath, res)
		if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			logger.Errorf("create crd %s failed: %v", res, err)
			return err
		}
	}

	return nil
}

type DeployPrometheusModule struct {
	common.KubeModule
}

func (m *DeployPrometheusModule) Init() {
	m.Name = "DeployPrometheus"

	createOperator := &task.RemoteTask{
		Name:  "CreatePrometheusOperator",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(CreateOperator),
		Parallel: false,
		Retry:    0,
	}

	createNodeExporter := &task.RemoteTask{
		Name:  "CreateNodeExporter",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &CreatePrometheusComponent{
			Component: "node-exporter",
			Force:     "--force",
		},
		Parallel: false,
		Retry:    0,
	}

	createKubeStateMetrics := &task.RemoteTask{
		Name:  "CreateKubeStateMetrics",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &CreatePrometheusComponent{
			Component: "kube-state-metrics",
			Force:     "--force",
		},
		Parallel: false,
		Retry:    0,
	}

	createPrometheus := &task.RemoteTask{
		Name:  "CreatePrometheus",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &CreatePrometheusComponent{
			Component: "prometheus",
		},
		Parallel: false,
	}

	createKubeMonitor := &task.RemoteTask{
		Name:  "CreateKubeMonitor",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action: &CreatePrometheusComponent{
			Component: "kubernetes",
			Force:     "--force",
		},
		Parallel: false,
	}

	//createAlertManager := &task.RemoteTask{
	//	Name:  "CreateAlertManager",
	//	Hosts: m.Runtime.GetHostsByRole(common.Master),
	//	Prepare: &prepare.PrepareCollection{
	//		new(common.OnlyFirstMaster),
	//		new(NotEqualDesiredVersion),
	//	},
	//	Action: &CreatePrometheusComponent{
	//		Component: "alertmanager",
	//	},
	//	Parallel: false,
	//}

	m.Tasks = []task.Interface{
		createOperator,
		createNodeExporter,
		createKubeStateMetrics,
		createPrometheus,
		createKubeMonitor,
		//createAlertManager,
	}

}
