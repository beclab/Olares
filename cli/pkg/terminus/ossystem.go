package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/storage"

	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	configmaptemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type InstallOsSystem struct {
	common.KubeAction
}

func (t *InstallOsSystem) Execute(runtime connector.Runtime) error {
	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}

	if !runtime.GetSystemInfo().IsDarwin() {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown 1000:1000 %s", storage.OlaresSharedLibDir, storage.OlaresSharedLibDir), false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "failed to create shared lib dir")
		}
	}

	var cmd = fmt.Sprintf("%s get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d", kubectl)
	redisPwd, err := runtime.GetRunner().SudoCmd(cmd, false, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get redis secret error")
	}

	if redisPwd == "" {
		return fmt.Errorf("redis secret not found")
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceOsSystem)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var systemPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "system")
	vals := map[string]interface{}{
		"kubesphere": map[string]interface{}{"redis_password": redisPwd},
		"backup": map[string]interface{}{
			"bucket":           t.KubeConf.Arg.Storage.BackupClusterBucket,
			"key_prefix":       t.KubeConf.Arg.Storage.StoragePrefix,
			"is_cloud_version": cloudValue(t.KubeConf.Arg.IsCloudInstance),
			"sync_secret":      t.KubeConf.Arg.Storage.StorageSyncSecret,
		},
		"gpu":                                  getGpuType(t.KubeConf.Arg.GPU.Enable, t.KubeConf.Arg.GPU.Share),
		"s3_bucket":                            t.KubeConf.Arg.Storage.StorageBucket,
		"fs_type":                              getRootFSType(),
		common.HelmValuesKeyTerminusGlobalEnvs: common.TerminusGlobalEnvs,
		common.HelmValuesKeyOlaresRootFSPath:   storage.OlaresRootDir,
	}

	if !runtime.GetSystemInfo().IsDarwin() {
		vals["sharedlib"] = storage.OlaresSharedLibDir
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameSystem, systemPath, "", common.NamespaceOsSystem, vals, false); err != nil {
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

type InstallOsSystemModule struct {
	common.KubeModule
}

func (m *InstallOsSystemModule) Init() {
	logger.InfoInstallationProgress("Installing appservice ...")
	m.Name = "InstallOsSystemModule"

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
				"os-system": {"tier=app-service"},
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
		installOsSystem,
		createBackupConfigMap,
		createReverseProxyConfigMap,
		checkSystemService,
		patchOs,
	}
}

func getGpuType(gpuEnable, gpuShare bool) (gpuType string) {
	gpuType = "none"
	if gpuEnable {
		if gpuShare {
			gpuType = "nvshare"
		} else {
			gpuType = "nvidia"
		}
	}

	return gpuType
}

func cloudValue(cloudInstance bool) string {
	if cloudInstance {
		return "true"
	}

	return ""
}

func getRootFSType() string {
	if util.IsExist(storage.JuiceFsServiceFile) {
		return "jfs"
	}
	return "fs"
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
