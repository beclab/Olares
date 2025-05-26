package plugins

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

var kscorecrds = []map[string]string{
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "serviceaccounts",
		"resource": "kubesphere-cluster-admin",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "serviceaccounts",
		"resource": "kubesphere-router-serviceaccount",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "role",
		"resource": "system:kubesphere-router-role",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "rolebinding",
		"resource": "nginx-ingress-role-nisa-binding",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "deployment",
		"resource": "default-http-backend",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-controls-system",
		"kind":     "service",
		"resource": "default-http-backend",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "secrets",
		"resource": "ks-controller-manager-webhook-cert",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "serviceaccounts",
		"resource": "kubesphere",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "clusterroles",
		"resource": "system:kubesphere-router-clusterrole",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "clusterrolebindings",
		"resource": "system:nginx-ingress-clusterrole-nisa-binding",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "clusterrolebindings",
		"resource": "system:kubesphere-cluster-admin",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "clusterrolebindings",
		"resource": "kubesphere",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "services",
		"resource": "ks-apiserver",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "services",
		"resource": "ks-controller-manager",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "deployments",
		"resource": "ks-apiserver",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "deployments",
		"resource": "ks-controller-manager",
		"release":  "ks-core",
	},
	//{
	//	"ns":       "kubesphere-system",
	//	"kind":     "validatingwebhookconfigurations",
	//	"resource": "users.iam.kubesphere.io",
	//	"release":  "ks-core",
	//},
	{
		"ns":       "kubesphere-system",
		"kind":     "validatingwebhookconfigurations",
		"resource": "resourcesquotas.quota.kubesphere.io",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "validatingwebhookconfigurations",
		"resource": "network.kubesphere.io",
		"release":  "ks-core",
	},
	{
		"ns":       "kubesphere-system",
		"kind":     "users.iam.kubesphere.io",
		"resource": "admin",
		"release":  "ks-core",
	},
}

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

type PatchKsCoreStatus struct {
	common.KubeAction
}

func (t *PatchKsCoreStatus) Execute(runtime connector.Runtime) error {
	//var kubectlpath, _ = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	//if kubectlpath == "" {
	//	kubectlpath = path.Join(common.BinDir, common.CommandKubectl)
	//}
	//
	//var jsonPath = fmt.Sprintf(`{\"status\": {\"core\": {\"status\": \"enabled\", \"enabledTime\": \"%s\"}}}`, time.Now().Format("2006-01-02T15:04:05Z"))
	//var cmd = fmt.Sprintf("%s patch cc ks-installer --type merge -p '%s' -n %s", kubectlpath, jsonPath, common.NamespaceKubesphereSystem)
	//
	//_, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true)
	//if err != nil {
	//	return errors.Wrap(errors.WithStack(err), "patch ks-core status failed")
	//}

	return nil
}

type CreateKsCoreConfig struct {
	common.KubeAction
}

func (t *CreateKsCoreConfig) Execute(runtime connector.Runtime) error {
	jwtSecretIf, ok := t.PipelineCache.Get(common.CacheJwtSecret)
	if !ok || jwtSecretIf == nil {
		return fmt.Errorf("failed to get jwt secret")
	}

	kubeVersionIf, ok := t.PipelineCache.Get(common.CacheKubeletVersion)
	if !ok || kubeVersionIf == nil {
		return fmt.Errorf("failed to get kubelet version")
	}

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
	values = make(map[string]interface{})
	values["Release"] = map[string]interface{}{
		"JwtSecret":   jwtSecretIf.(string),
		"TokenMaxAge": t.KubeConf.Arg.TokenMaxAge * int64(time.Second),
	}
	if err := utils.UpgradeCharts(context.Background(), actionConfig, settings, appKsConfigName,
		appPath, "", common.NamespaceKubesphereSystem, values, false); err != nil {
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

type PacthKsCore struct {
	common.KubeAction
}

func (t *PacthKsCore) Execute(runtime connector.Runtime) error {
	var secretsNum int64
	var crdNum int64
	var secretsNumIf, ok = t.PipelineCache.Get(common.CacheSecretsNum)
	if ok && secretsNumIf != nil {
		secretsNum = secretsNumIf.(int64)
	}

	crdNumIf, ok := t.PipelineCache.Get(common.CacheCrdsNUm)
	if ok && crdNumIf != nil {
		crdNum = crdNumIf.(int64)
	}

	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	if secretsNum == 0 && crdNum != 0 {
		for _, item := range kscorecrds {
			var cmd = fmt.Sprintf("%s -n %s annotate --overwrite %s %s meta.helm.sh/release-name=%s && %s -n %s annotate --overwrite %s %s meta.helm.sh/release-namespace=%s && %s -n %s label --overwrite %s %s app.kubernetes.io/managed-by=Helm",
				kubectlpath, item["ns"], item["kind"], item["resource"], item["release"],
				kubectlpath, item["ns"], item["kind"], item["resource"], common.NamespaceKubesphereSystem,
				kubectlpath, item["ns"], item["kind"], item["resource"])

			if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
				return errors.Wrap(errors.WithStack(err), "patch ks-core crd")
			}
		}
	}

	return nil
}

type CheckKsCoreExist struct {
	common.KubeAction
}

func (t *CheckKsCoreExist) Execute(runtime connector.Runtime) error {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var cmd string

	cmd = fmt.Sprintf("%s -n %s get secrets --field-selector=type=helm.sh/release.v1 | grep ks-core |wc -l",
		kubectlpath,
		common.NamespaceKubesphereSystem)
	stdout, _ := runtime.GetRunner().SudoCmd(cmd, false, false)

	secretNum, err := strconv.ParseInt(stdout, 10, 64)
	if err != nil {
		secretNum = 0
	}

	cmd = fmt.Sprintf("%s get crd users.iam.kubesphere.io | grep 'users.iam.kubesphere.io' |wc -l", kubectlpath)
	stdout, _ = runtime.GetRunner().SudoCmd(cmd, false, false)

	usersCrdNum, err := strconv.ParseInt(stdout, 10, 64)
	if err != nil {
		usersCrdNum = 0
	}

	logger.Debugf("secretNum: %d, usersCrdNum: %d", secretNum, usersCrdNum)

	t.ModuleCache.Set(common.CacheSecretsNum, secretNum)
	t.ModuleCache.Set(common.CacheCrdsNUm, usersCrdNum)

	return nil
}

type DeployKsCoreConfigModule struct {
	common.KubeModule
}

func (m *DeployKsCoreConfigModule) Init() {
	m.Name = "DeployKsCoreConfig"

	checkKsCoreExist := &task.RemoteTask{
		Name:  "CheckKsCoreExist",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
			new(common.GetMasterNum),
		},
		Action:   new(CheckKsCoreExist),
		Parallel: false,
		Retry:    0,
	}

	pacthKsCore := &task.RemoteTask{
		Name:  "PacthKsCore",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(PacthKsCore),
		Parallel: false,
		Retry:    0,
	}

	createKsCoreConfigManifests := &task.RemoteTask{
		Name:  "CreateKsCoreConfigManifests",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
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
			new(NotEqualDesiredVersion),
		},
		Action:   new(CreateKsCoreConfig),
		Parallel: true,
		Retry:    0,
	}

	patchKsCoreStatus := &task.RemoteTask{
		Name:  "PatchKsCoreStatus",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(PatchKsCoreStatus),
		Parallel: true,
		Retry:    0,
	}

	createKsRole := &task.RemoteTask{
		Name:  "CreateKsRole",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
		},
		Action:   new(CreateKsRole),
		Parallel: true,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		checkKsCoreExist,
		pacthKsCore,
		createKsCoreConfigManifests,
		createKsCoreConfig,
		patchKsCoreStatus,
		createKsRole,
	}
}
