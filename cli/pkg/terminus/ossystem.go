package terminus

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/beclab/Olares/framework/app-service/api/sys.bytetrade.io/v1alpha1"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/storage"

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
		"gpu":                                getGpuType(t.KubeConf.Arg.GPU.Enable),
		"s3_bucket":                          t.KubeConf.Arg.Storage.StorageBucket,
		"fs_type":                            storage.GetRootFSType(),
		common.HelmValuesKeyOlaresRootFSPath: storage.OlaresRootDir,
		"sharedlib":                          storage.OlaresSharedLibDir,
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

type CreateUserEnvConfigMap struct {
	common.KubeAction
}

func (t *CreateUserEnvConfigMap) Execute(runtime connector.Runtime) error {
	userEnvPath := filepath.Join(runtime.GetInstallerDir(), common.OLARES_USER_ENV_FILENAME)
	if !util.IsExist(userEnvPath) {
		logger.Info("user env config file not found, skipping user env configmap apply")
		return nil
	}

	desiredBytes, err := os.ReadFile(userEnvPath)
	if err != nil {
		return errors.Wrap(err, "failed to read user env config file")
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
	namespace := common.NamespaceOsFramework
	cm := &corev1.ConfigMap{}
	err = ctrlclient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, cm)
	if apierrors.IsNotFound(err) {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: map[string]string{
				common.OLARES_USER_ENV_FILENAME: string(desiredBytes),
			},
		}

		if err := ctrlclient.Create(ctx, cm); err != nil && !apierrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "failed to create user-env configmap")
		}

		logger.Infof("Created user env configmap from %s", userEnvPath)
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "failed to get user-env configmap")
	}

	if cm.Data == nil {
		cm.Data = map[string]string{}
	}
	if cm.Data[common.OLARES_USER_ENV_FILENAME] == string(desiredBytes) {
		logger.Infof("user-env configmap is up to date")
		return nil
	}

	cm.Data[common.OLARES_USER_ENV_FILENAME] = string(desiredBytes)

	if err := ctrlclient.Update(ctx, cm); err != nil {
		return errors.Wrap(err, "failed to update user-env configmap")
	}

	logger.Infof("Updated user env configmap from %s", userEnvPath)
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

		err = envItem.ValidateValue(envItem.Value)
		if err != nil {
			return fmt.Errorf("invalid system env value: %s", envItem.Value)
		}
		err = envItem.ValidateValue(envItem.Default)
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
		Retry:  5,
		Delay:  15 * time.Second,
	}

	createUserEnvConfigMap := &task.LocalTask{
		Name:   "CreateUserEnvConfigMap",
		Action: &CreateUserEnvConfigMap{},
	}

	createSharedLibDir := &task.LocalTask{
		Name:   "CreateSharedLibDir",
		Action: &storage.CreateSharedLibDir{},
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
		createSharedLibDir,
		installOsSystem,
		createBackupConfigMap,
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

type UserEnvConfig struct {
	APIVersion string                `yaml:"apiVersion"`
	UserEnvs   []v1alpha1.EnvVarSpec `yaml:"userEnvs"`
}
