package terminus

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/storage"
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var BuiltInApps = []string{"portfolio", "vault", "desktop", "message", "wise", "search", "appstore", "notification", "dashboard", "settings", "profile", "agent", "files", "studio"}

type InstallAppsModule struct {
	common.KubeModule
}

func (i *InstallAppsModule) Init() {
	logger.InfoInstallationProgress("Installing built-in apps ...")
	i.Name = "Install Built-in apps"
	i.Desc = "Install Built-in apps"

	prepareAppValues := &task.LocalTask{
		Name:   "PrepareAppValues",
		Desc:   "Prepare app values",
		Action: new(PrepareAppValues),
	}

	installApps := &task.LocalTask{
		Name:   "InstallApps",
		Desc:   "Install apps",
		Action: new(InstallApps),
		Retry:  5,
	}

	clearAppsValues := &task.LocalTask{
		Name:   "ClearAppValues",
		Desc:   "Clear apps values",
		Action: new(ClearAppValues),
	}

	copyFiles := &task.LocalTask{
		Name:   "CopyAppServiceHelmFiles",
		Desc:   "Copy files",
		Action: new(CopyAppServiceHelmFiles),
		Retry:  5,
	}

	i.Tasks = []task.Interface{
		prepareAppValues,
		installApps,
		clearAppsValues,
		copyFiles,
	}

}

type PrepareAppValues struct {
	common.KubeAction
}

func (u *PrepareAppValues) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	ns := fmt.Sprintf("user-space-%s", u.KubeConf.Arg.User.UserName)
	redisPassword, err := getRedisPassword(client, runtime)
	if err != nil {
		return err
	}

	bfDocUrl, _ := getDocUrl(ctx, runtime)
	bflNodeName, err := getBflPod(ctx, ns, client, runtime)
	if err != nil {
		return err
	}
	bflAnnotations, err := getBflAnnotation(ctx, ns, client, runtime)
	if err != nil {
		return err
	}
	fsType := getRootFSType()
	gpuType := getGpuType(u.KubeConf.Arg.GPU.Enable, u.KubeConf.Arg.GPU.Share)
	appValues := getAppSecrets(getAppPatches())

	var values = map[string]interface{}{
		"bfl": map[string]interface{}{
			"nodeport":               30883,
			"nodeport_ingress_http":  30083,
			"nodeport_ingress_https": 30082,
			"username":               u.KubeConf.Arg.User.UserName,
			"admin_user":             true,
			"url":                    bfDocUrl,
			"nodeName":               bflNodeName,
		},
		"pvc": map[string]interface{}{
			"userspace": bflAnnotations["userspace_pv"],
		},
		"userspace": map[string]interface{}{
			"userData": fmt.Sprintf("%s/Home", bflAnnotations["userspace_hostpath"]),
			"appData":  fmt.Sprintf("%s/Data", bflAnnotations["userspace_hostpath"]),
			"appCache": bflAnnotations["appcache_hostpath"],
			"dbdata":   bflAnnotations["dbdata_hostpath"],
		},
		"desktop": map[string]interface{}{
			"nodeport": 30180,
		},
		"global": map[string]interface{}{
			"bfl": map[string]interface{}{
				"username": u.KubeConf.Arg.User.UserName,
			},
		},
		"debugVersion": os.Getenv("DEBUG_VERSION") != "",
		"gpu":          gpuType,
		"fs_type":      fsType,
		"os":           appValues,
		"kubesphere": map[string]interface{}{
			"redis_password": redisPassword,
		},
		common.HelmValuesKeyTerminusGlobalEnvs: common.TerminusGlobalEnvs,
		common.HelmValuesKeyOlaresRootFSPath:   storage.OlaresRootDir,
	}

	u.ModuleCache.Set(common.CacheAppValues, values)

	return nil
}

type InstallApps struct {
	common.KubeAction
}

func (i *InstallApps) Execute(runtime connector.Runtime) error {
	var appPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "apps")
	appsDirEntries, err := os.ReadDir(appPath)
	if err != nil {
		return errors.Wrapf(err, "failed to list %s", appPath)
	}
	var apps []string
	for _, entry := range appsDirEntries {
		if entry.IsDir() {
			apps = append(apps, entry.Name())
		}
	}

	var ns = fmt.Sprintf("user-space-%s", i.KubeConf.Arg.User.UserName)
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	valsCache, ok := i.ModuleCache.Get(common.CacheAppValues)
	if !ok {
		return fmt.Errorf("app values not found")
	}
	vals, ok := valsCache.(map[string]interface{})
	if !ok {
		return fmt.Errorf("app values in cache is not map[string]interface{}")
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	for _, app := range apps {
		if err := utils.UpgradeCharts(ctx, actionConfig, settings, app, path.Join(appPath, app), "", ns, vals, false); err != nil {
			return fmt.Errorf("install app %s failed: %v", app, err)
		}
	}

	return nil
}

type ClearAppValues struct {
	common.KubeAction
}

func (c *ClearAppValues) Execute(runtime connector.Runtime) error {
	// clear apps values.yaml
	_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("cat /dev/null > %s/wizard/config/apps/values.yaml", runtime.GetInstallerDir()), false, false)

	return nil
}

type CopyAppServiceHelmFiles struct {
	common.KubeAction
}

func (c *CopyAppServiceHelmFiles) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	appServiceName, err := getAppServiceName(client, runtime)
	if err != nil {
		return err
	}

	kubeclt, _ := util.GetCommand(common.CommandKubectl)
	for _, app := range []string{"launcher", "apps"} {
		var cmd = fmt.Sprintf("%s cp %s/wizard/config/%s os-system/%s:/userapps -c app-service", kubeclt, runtime.GetInstallerDir(), app, appServiceName)
		if _, err = runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "copy files failed")
		}
	}

	return nil
}

func getAppServiceName(client clientset.Client, runtime connector.Runtime) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pods, err := client.Kubernetes().CoreV1().Pods(common.NamespaceOsSystem).List(ctx, metav1.ListOptions{LabelSelector: "tier=app-service"})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get app-service failed")
	}

	if len(pods.Items) == 0 {
		return "", errors.New("app-service not found")
	}

	return pods.Items[0].Name, nil
}

func getBflPod(ctx context.Context, ns string, client clientset.Client, runtime connector.Runtime) (string, error) {
	pods, err := client.Kubernetes().CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "tier=bfl"})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get bfl failed")
	}

	if len(pods.Items) == 0 {
		return "", errors.New("bfl not found")
	}

	return pods.Items[0].Spec.NodeName, nil
}

func getDocUrl(ctx context.Context, runtime connector.Runtime) (url string, err error) {
	var nodeip string
	var cmd = fmt.Sprintf(`curl --connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 -s http://checkip.dyndns.org/ | grep -o "[[:digit:].]\+"`)
	nodeip, _ = runtime.GetRunner().SudoCmdContext(ctx, cmd, false, false)
	url = fmt.Sprintf("http://%s:30883/bfl/apidocs.json", nodeip)
	return
}

func getBflAnnotation(ctx context.Context, ns string, client clientset.Client, runtime connector.Runtime) (map[string]string, error) {
	sts, err := client.Kubernetes().AppsV1().StatefulSets(ns).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "get bfl sts failed")
	}
	if sts == nil {
		return nil, errors.New("bfl sts not found")
	}

	return sts.Annotations, nil
}

func getAppSecrets(patches map[string]map[string]interface{}) map[string]interface{} {
	var secrets = make(map[string]interface{})
	for _, app := range BuiltInApps {
		s, _ := utils.GeneratePassword(16)
		var v = make(map[string]interface{})
		v["appKey"] = fmt.Sprintf("bytetrade_%s_%d", app, utils.Random())
		v["appSecret"] = s

		p, ok := patches[app]
		if ok && p != nil {
			for pk, pv := range p {
				v[pk] = pv
			}
		}

		secrets[app] = v
	}

	return secrets
}

func getAppPatches() map[string]map[string]interface{} {
	var patches = make(map[string]map[string]interface{})
	var value = make(map[string]interface{})
	value["marketProvider"] = os.Getenv("MARKET_PROVIDER")
	patches["appstore"] = value
	return patches
}
