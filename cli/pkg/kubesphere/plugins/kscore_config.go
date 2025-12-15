package plugins

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/core/prepare"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CreateKsRole struct {
	common.KubeAction
}

func (t *CreateKsRole) Execute(runtime connector.Runtime) error {
	var f = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, "ks-init", "role-templates.yaml")
	if !utils.IsExist(f) {
		return fmt.Errorf("file %s not found", f)
	}

	var kubectlpath, _ = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	}

	cmd := fmt.Sprintf("%s apply -f %s", kubectlpath, f)
	_, err := runtime.GetRunner().SudoCmd(cmd, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "create ks role failed")
	}
	return nil
}

type CreateKsCoreConfig struct {
	common.KubeAction
}

func (t *CreateKsCoreConfig) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	var appKsCoreConfigName = common.ChartNameKsCoreConfig
	var appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, appKsCoreConfigName)

	// create ks-core-config
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceKubesphereSystem)
	if err != nil {
		return err
	}

	var values = make(map[string]interface{})
	values["Release"] = map[string]string{
		"Namespace": common.NamespaceKubesphereSystem,
	}
	if err := utils.UpgradeCharts(context.Background(), actionConfig, settings, appKsCoreConfigName,
		appPath, "", common.NamespaceKubesphereSystem, values, false); err != nil {
		logger.Errorf("failed to install %s chart: %v", appKsCoreConfigName, err)
		return err
	}

	// create ks-config
	var appKsConfigName = common.ChartNameKsConfig
	appPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, appKsConfigName)
	if err := utils.UpgradeCharts(context.Background(), actionConfig, settings, appKsConfigName,
		appPath, "", common.NamespaceKubesphereSystem, nil, false); err != nil {
		logger.Errorf("failed to install %s chart: %v", appKsConfigName, err)
		return err
	}

	return nil
}

type CreateKsCoreConfigManifests struct {
	common.KubeAction
}

func (t *CreateKsCoreConfigManifests) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var kscoreConfigCrdsPath = path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir, cc.BuildDir, common.ChartNameKsCoreConfig, "crds")

	filepath.Walk(kscoreConfigCrdsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, path), false, true)
			if err != nil {
				logger.Errorf("failed to apply %s: %v", path, err)
				return err
			}
		}
		return nil
	})

	return nil
}

type DeployKsCoreConfigModule struct {
	common.KubeModule
}

func (m *DeployKsCoreConfigModule) Init() {
	m.Name = "DeployKsCoreConfig"

	createKsCoreConfigManifests := &task.RemoteTask{
		Name:  "CreateKsCoreConfigManifests",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CreateKsCoreConfigManifests),
		Parallel: false,
		Retry:    30,
		Delay:    5 * time.Second,
	}

	createKsCoreConfig := &task.RemoteTask{
		Name:  "CreateKsCoreConfig",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CreateKsCoreConfig),
		Parallel: true,
		Retry:    0,
	}

	createKsRole := &task.RemoteTask{
		Name:  "CreateKsRole",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CreateKsRole),
		Parallel: true,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		createKsCoreConfigManifests,
		createKsCoreConfig,
		createKsRole,
	}
}
