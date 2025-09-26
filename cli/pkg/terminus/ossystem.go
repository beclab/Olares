package terminus

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"bytetrade.io/web3os/app-service/api/sys.bytetrade.io/v1alpha1"
	apputils "bytetrade.io/web3os/app-service/pkg/utils"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/storage"

	"github.com/beclab/Olares/cli/pkg/clientset"
	"github.com/beclab/Olares/cli/pkg/common"
	cc "github.com/beclab/Olares/cli/pkg/core/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/task"
	"github.com/beclab/Olares/cli/pkg/core/util"
	configmaptemplates "github.com/beclab/Olares/cli/pkg/terminus/templates"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InstallOsSystem struct {
	common.KubeAction
}

func (t *InstallOsSystem) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsDarwin() {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown 1000:1000 %s", storage.OlaresSharedLibDir, storage.OlaresSharedLibDir), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to create shared lib dir")
		}
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceOsPlatform)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	vals := map[string]interface{}{
		"backup": map[string]interface{}{
			"bucket":           t.KubeConf.Arg.Storage.BackupClusterBucket,
			"key_prefix":       t.KubeConf.Arg.Storage.StoragePrefix,
			"is_cloud_version": cloudValue(t.KubeConf.Arg.IsCloudInstance),
			"sync_secret":      t.KubeConf.Arg.Storage.StorageSyncSecret,
		},
		"gpu":                                  getGpuType(t.KubeConf.Arg.GPU.Enable),
		"s3_bucket":                            t.KubeConf.Arg.Storage.StorageBucket,
		"fs_type":                              storage.GetRootFSType(),
		common.HelmValuesKeyTerminusGlobalEnvs: common.TerminusGlobalEnvs,
		common.HelmValuesKeyOlaresRootFSPath:   storage.OlaresRootDir,
		"sharedlib":                            storage.OlaresSharedLibDir,
	}

	var platformPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-platform")
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSPlatform, platformPath, "", common.NamespaceOsPlatform, vals, false); err != nil {
		return err
	}

	// TODO: wait for the platform to be ready

	actionConfig, settings, err = utils.InitConfig(config, common.NamespaceOsFramework)
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var frameworkPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "os-framework")
	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameOSFramework, frameworkPath, "", common.NamespaceOsFramework, vals, false); err != nil {
		return err
	}

	return nil
}

type CreateBackupConfigMap struct {
	common.KubeAction
}

func (t *CreateBackupConfigMap) Execute(runtime connector.Runtime) error {
	var backupConfigMapFile = path.Join(runtime.GetInstallerDir(), "deploy", configmaptemplates.BackupConfigMap.Name())
	var data = util.Data{
		"CloudInstance":     cloudValue(t.KubeConf.Arg.IsCloudInstance),
		"StorageBucket":     t.KubeConf.Arg.Storage.BackupClusterBucket,
		"StoragePrefix":     t.KubeConf.Arg.Storage.StoragePrefix,
		"StorageSyncSecret": t.KubeConf.Arg.Storage.StorageSyncSecret,
	}

	backupConfigStr, err := util.Render(configmaptemplates.BackupConfigMap, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render backup configmap template failed")
	}
	if err := util.WriteFile(backupConfigMapFile, []byte(backupConfigStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write backup configmap %s failed", backupConfigMapFile))
	}

	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, backupConfigMapFile), false, true); err != nil {
		return err
	}

	return nil
}

type CreateReverseProxyConfigMap struct {
	common.KubeAction
}

func (c *CreateReverseProxyConfigMap) Execute(runtime connector.Runtime) error {
	var defaultReverseProxyConfigMapFile = path.Join(runtime.GetInstallerDir(), "deploy", configmaptemplates.ReverseProxyConfigMap.Name())
	var data = util.Data{
		"EnableCloudflare": c.KubeConf.Arg.Cloudflare.Enable,
		"EnableFrp":        c.KubeConf.Arg.Frp.Enable,
		"FrpServer":        c.KubeConf.Arg.Frp.Server,
		"FrpPort":          c.KubeConf.Arg.Frp.Port,
		"FrpAuthMethod":    c.KubeConf.Arg.Frp.AuthMethod,
		"FrpAuthToken":     c.KubeConf.Arg.Frp.AuthToken,
	}

	reverseProxyConfigStr, err := util.Render(configmaptemplates.ReverseProxyConfigMap, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render default reverse proxy configmap template failed")
	}
	if err := util.WriteFile(defaultReverseProxyConfigMapFile, []byte(reverseProxyConfigStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write default reverse proxy configmap %s failed", defaultReverseProxyConfigMapFile))
	}

	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, defaultReverseProxyConfigMapFile), false, true); err != nil {
		return err
	}

	return nil
}

type CreateUserEnvConfigMap struct {
	common.KubeAction
}

func (t *CreateUserEnvConfigMap) Execute(runtime connector.Runtime) error {
	userEnvPath := filepath.Join(runtime.GetInstallerDir(), common.OLARES_USER_ENV_FILENAME)
	if !util.IsExist(userEnvPath) {
		logger.Info("user env config file not found, skipping user env configmap apply")
		return nil
	}

	configK8s, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes config")
	}

	scheme := kruntime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to add corev1 to scheme")
	}

	ctrlclient, err := client.New(configK8s, client.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	name := "user-env"
	cm := &corev1.ConfigMap{}
	err = ctrlclient.Get(ctx, types.NamespacedName{Name: name, Namespace: common.NamespaceOsFramework}, cm)
	if apierrors.IsNotFound(err) {
		// create via kubectl from file
		kubectl, _ := util.GetCommand(common.CommandKubectl)
		cmd := fmt.Sprintf("%s -n %s create configmap %s --from-file=%s=%s",
			kubectl, common.NamespaceOsFramework, name, common.OLARES_USER_ENV_FILENAME, userEnvPath,
		)
		if _, cerr := runtime.GetRunner().SudoCmd(cmd, false, true); cerr != nil {
			return errors.Wrap(errors.WithStack(cerr), "failed to create user-env configmap")
		}
		logger.Infof("Created user env configmap from %s", userEnvPath)
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to get user-env configmap")
	}

	// If exists, merge missing envs and update via client
	newDataBytes, err := os.ReadFile(userEnvPath)
	if err != nil {
		return errors.Wrap(err, "failed to read user env config file")
	}

	var newCfg UserEnvConfig
	if err := yaml.Unmarshal(newDataBytes, &newCfg); err != nil {
		return errors.Wrap(err, "failed to parse user env config file")
	}

	var existingCfg UserEnvConfig
	existingContent := cm.Data[common.OLARES_USER_ENV_FILENAME]
	if existingContent != "" {
		if err := yaml.Unmarshal([]byte(existingContent), &existingCfg); err != nil {
			return errors.Wrap(err, "failed to parse existing user env configmap data")
		}
	}

	existingSet := make(map[string]struct{}, len(existingCfg.UserEnvs))
	for _, e := range existingCfg.UserEnvs {
		existingSet[e.EnvName] = struct{}{}
	}

	missing := 0
	for _, e := range newCfg.UserEnvs {
		if _, ok := existingSet[e.EnvName]; !ok {
			existingCfg.UserEnvs = append(existingCfg.UserEnvs, e)
			missing++
		}
	}

	if missing == 0 {
		logger.Infof("No new user envs to add; configmap up to date")
		return nil
	}

	updatedBytes, err := yaml.Marshal(existingCfg)
	if err != nil {
		return errors.Wrap(err, "failed to marshal updated user env config")
	}
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	cm.Data[common.OLARES_USER_ENV_FILENAME] = string(updatedBytes)

	if err := ctrlclient.Update(ctx, cm); err != nil {
		return errors.Wrap(err, "failed to update user-env configmap")
	}

	logger.Infof("Appended %d missing user env(s) and updated configmap", missing)
	return nil
}

type Patch struct {
	common.KubeAction
}

func (p *Patch) Execute(runtime connector.Runtime) error {
	var err error
	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	var globalRoleWorkspaceManager = path.Join(runtime.GetInstallerDir(), "deploy", "patch-globalrole-workspace-manager.yaml")
	if _, err = runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, globalRoleWorkspaceManager), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "patch globalrole workspace manager failed")
	}

	//var notificationManager = path.Join(runtime.GetInstallerDir(), "deploy", "patch-notification-manager.yaml")
	//if _, err = runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, notificationManager), false, true); err != nil {
	//	return errors.Wrap(errors.WithStack(err), "patch notification manager failed")
	//}
	//var notificationManager = path.Join(runtime.GetInstallerDir(), "deploy", "patch-notification-manager.yaml")
	//if _, err = runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, notificationManager), false, true); err != nil {
	//	return errors.Wrap(errors.WithStack(err), "patch notification manager failed")
	//}
	//
	//patchAdminContent := `{"metadata":{"finalizers":["finalizers.kubesphere.io/users"]}}`
	//patchAdminCMD := fmt.Sprintf(
	//	"%s patch user admin -p '%s' --type='merge' ",
	//	kubectl,
	//	patchAdminContent)
	//_, err = runtime.GetRunner().SudoCmd(patchAdminCMD, false, true)
	//if err != nil {
	//	return errors.Wrap(errors.WithStack(err), "patch user admin failed")
	//}
	//patchAdminContent := "{\\\"metadata\\\":{\\\"finalizers\\\":[\\\"finalizers.kubesphere.io/users\\\"]}}"
	//patchAdminCMD := fmt.Sprintf(
	//	"%s patch user admin -p '%s' --type='merge' ",
	//	kubectl,
	//	patchAdminContent)
	//_, err = runtime.GetRunner().Host.SudoCmd(patchAdminCMD, false, true)
	//if err != nil {
	//	return errors.Wrap(errors.WithStack(err), "patch user admin failed")
	//}

	//deleteAdminCMD := fmt.Sprintf("%s delete user admin --ignore-not-found", kubectl)
	//_, err = runtime.GetRunner().SudoCmd(deleteAdminCMD, false, true)
	//if err != nil {
	//	return errors.Wrap(errors.WithStack(err), "failed to delete ks admin user")
	//}
	deleteKubectlAdminCMD := fmt.Sprintf("%s -n kubesphere-controls-system delete deploy kubectl-admin --ignore-not-found", kubectl)
	_, err = runtime.GetRunner().SudoCmd(deleteKubectlAdminCMD, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to delete ks kubectl admin deployment")
	}
	deleteHTTPBackendCMD := fmt.Sprintf("%s -n kubesphere-controls-system delete deploy default-http-backend --ignore-not-found", kubectl)
	_, err = runtime.GetRunner().SudoCmd(deleteHTTPBackendCMD, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to delete ks default http backend")
	}

	patchFelixConfigContent := `{"spec":{"featureDetectOverride": "SNATFullyRandom=false,MASQFullyRandom=false"}}`
	patchFelixConfigCMD := fmt.Sprintf(
		"%s patch felixconfiguration default -p '%s'  --type='merge'",
		kubectl,
		patchFelixConfigContent,
	)
	_, err = runtime.GetRunner().SudoCmd(patchFelixConfigCMD, false, true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to patch felix configuration")
	}

	return nil
}

type ApplySystemEnv struct {
	common.KubeAction
}

// SystemEnvConfig represents the structure of the config.yaml file
type SystemEnvConfig struct {
	APIVersion string                `yaml:"apiVersion"`
	SystemEnvs []v1alpha1.EnvVarSpec `yaml:"systemEnvs"`
}

func (a *ApplySystemEnv) Execute(runtime connector.Runtime) error {
	configPath := filepath.Join(runtime.GetInstallerDir(), common.OLARES_SYSTEM_ENV_FILENAME)
	if !util.IsExist(configPath) {
		logger.Info("system env config file not found, skipping system env apply")
		return nil
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to read system env config file")
	}

	var config SystemEnvConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return errors.Wrap(err, "failed to parse system env config file")
	}

	logger.Debugf("parsed system env config file %s: %#v", configPath, config.SystemEnvs)

	configK8s, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get kubernetes config")
	}

	scheme := kruntime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to add system scheme")
	}

	ctrlclient, err := client.New(configK8s, client.Options{Scheme: scheme})
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, envItem := range config.SystemEnvs {
		resourceName, err := apputils.EnvNameToResourceName(envItem.EnvName)
		if err != nil {
			return fmt.Errorf("invalid system env name: %s", envItem.EnvName)
		}

		var existingSystemEnv v1alpha1.SystemEnv
		err = ctrlclient.Get(ctx, types.NamespacedName{Name: resourceName}, &existingSystemEnv)

		if err == nil {
			logger.Debugf("SystemEnv %s already exists, skipping", resourceName)
			continue
		}

		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get system env")
		}

		// before applying, if process env has the new name set, override Default with that value
		// we do not set the value because this is a default system value from installation
		// and can be reset
		// wheras the value is managed by user
		if procVal := os.Getenv(envItem.EnvName); procVal != "" {
			envItem.Default = procVal
		}

		err = apputils.CheckEnvValueByType(envItem.Value, envItem.Type)
		if err != nil {
			return fmt.Errorf("invalid system env value: %s", envItem.Value)
		}
		err = apputils.CheckEnvValueByType(envItem.Default, envItem.Type)
		if err != nil {
			return fmt.Errorf("invalid system env default value: %s", envItem.Value)
		}

		systemEnv := &v1alpha1.SystemEnv{
			ObjectMeta: metav1.ObjectMeta{
				Name: resourceName,
			},
			EnvVarSpec: envItem,
		}

		if err := ctrlclient.Create(ctx, systemEnv); err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create system env %s: %v", resourceName, err)
		}

		logger.Infof("Created SystemEnv: %s", systemEnv.EnvName)
	}

	return nil
}

type InstallOsSystemModule struct {
	common.KubeModule
}

func (m *InstallOsSystemModule) Init() {
	logger.InfoInstallationProgress("Installing appservice ...")
	m.Name = "InstallOsSystemModule"

	applySystemEnv := &task.LocalTask{
		Name:   "ApplySystemEnv",
		Action: new(ApplySystemEnv),
	}

	createUserEnvConfigMap := &task.LocalTask{
		Name:   "CreateUserEnvConfigMap",
		Action: &CreateUserEnvConfigMap{},
	}

	installOsSystem := &task.LocalTask{
		Name:   "InstallOsSystem",
		Action: &InstallOsSystem{},
		Retry:  1,
	}

	createBackupConfigMap := &task.LocalTask{
		Name:   "CreateBackupConfigMap",
		Action: &CreateBackupConfigMap{},
	}

	createReverseProxyConfigMap := &task.LocalTask{
		Name:   "CreateReverseProxyConfigMap",
		Action: &CreateReverseProxyConfigMap{},
	}

	checkSystemService := &task.LocalTask{
		Name: "CheckSystemServiceStatus",
		Action: &CheckPodsRunning{
			labels: map[string][]string{
				"os-framework": {"tier=app-service"},
			},
		},
		Retry: 20,
		Delay: 10 * time.Second,
	}

	patchOs := &task.LocalTask{
		Name:   "PatchOs",
		Action: &Patch{},
		Retry:  3,
		Delay:  30 * time.Second,
	}

	m.Tasks = []task.Interface{
		applySystemEnv,
		createUserEnvConfigMap,
		installOsSystem,
		createBackupConfigMap,
		createReverseProxyConfigMap,
		checkSystemService,
		patchOs,
	}
}

func getGpuType(gpuEnable bool) (gpuType string) {
	if gpuEnable {
		return "nvidia"
	}
	return "none"
}

func cloudValue(cloudInstance bool) string {
	if cloudInstance {
		return "true"
	}

	return ""
}

func getRedisPassword(client clientset.Client, runtime connector.Runtime) (string, error) {
	secret, err := client.Kubernetes().CoreV1().Secrets(common.NamespaceKubesphereSystem).Get(context.Background(), "redis-secret", metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get redis secret failed")
	}
	if secret == nil || secret.Data == nil || secret.Data["auth"] == nil {
		return "", fmt.Errorf("redis secret not found")
	}

	return string(secret.Data["auth"]), nil

}

type UserEnvConfig struct {
	APIVersion string                `yaml:"apiVersion"`
	UserEnvs   []v1alpha1.EnvVarSpec `yaml:"userEnvs"`
}
