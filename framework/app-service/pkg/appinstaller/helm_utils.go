package appinstaller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	"github.com/beclab/Olares/framework/app-service/pkg/tapr"
	userspacev1 "github.com/beclab/Olares/framework/app-service/pkg/users/userspace/v1"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/api/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// BuildBaseHelmValues builds the Helm values map used for both image-download
// (dry-run) and real install/upgrade, keeping template rendering consistent.
//
// When dryRun is true the function fills placeholder values for keys that are
// expensive to obtain or have side effects (OIDC, permissions, userspace paths,
// cluster-scoped deps, middleware, etc.) so that a Helm template dry-run still
// produces a complete manifest. When dryRun is false those keys are left unset
// and the caller (SetValues) is expected to fill them with real data.
func BuildBaseHelmValues(ctx context.Context, kubeConfig *rest.Config, appConfig *appcfg.ApplicationConfig, ownerName string, dryRun bool) (values map[string]interface{}, err error) {
	values = make(map[string]interface{})

	values["bfl"] = map[string]interface{}{
		"username": ownerName,
	}

	isAdmin, err := kubesphere.IsAdmin(ctx, kubeConfig, ownerName)
	if err != nil {
		return values, err
	}
	values["isAdmin"] = isAdmin

	admin, err := kubesphere.GetAdminUsername(ctx, kubeConfig)
	if err != nil {
		return values, err
	}
	if isAdmin {
		admin = ownerName
	}
	values["admin"] = admin

	values["GPU"] = map[string]interface{}{
		"Type": appConfig.GetSelectedGpuTypeValue(),
		"Cuda": os.Getenv("OLARES_SYSTEM_CUDA_VERSION"),
	}
	values["gpu"] = appConfig.GetSelectedGpuTypeValue()

	terminus, err := utils.GetTerminusVersion(ctx, kubeConfig)
	if err != nil {
		return values, err
	}
	values["sysVersion"] = terminus.Spec.Version

	nodeInfo, err := utils.GetNodeInfo(ctx)
	if err != nil {
		return values, err
	}
	values["nodes"] = nodeInfo

	deviceName, err := utils.GetDeviceName()
	if err != nil {
		return values, err
	}
	values["deviceName"] = deviceName

	kClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return values, err
	}
	nodes, err := kClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return values, err
	}
	var arch string
	for _, node := range nodes.Items {
		arch = node.Labels["kubernetes.io/arch"]
		break
	}
	values["cluster"] = map[string]interface{}{
		"arch": arch,
	}

	values["sharedlib"] = os.Getenv("SHARED_LIB_PATH")

	rootPath := userspacev1.DefaultRootPath
	if os.Getenv(userspacev1.OlaresRootPath) != "" {
		rootPath = os.Getenv(userspacev1.OlaresRootPath)
	}
	values["rootPath"] = rootPath

	values["downloadCdnURL"] = os.Getenv("OLARES_SYSTEM_CDN_SERVICE")
	values["fs_type"] = utils.EnvOrDefault("OLARES_SYSTEM_ROOTFS_TYPE", "fs")

	if err := addEnvironmentVariables(ctx, kubeConfig, appConfig, ownerName, values); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return values, err
	}

	if dryRun {
		// Placeholder values for keys that require HelmOps context, have side
		// effects, or are otherwise hard to obtain during the image-download
		// phase. These ensure the Helm template dry-run renders without errors.
		values["user"] = map[string]interface{}{"zone": "user-zone"}
		values["schedule"] = map[string]interface{}{"nodeName": "node"}
		values["oidc"] = map[string]interface{}{
			"client": map[string]interface{}{},
			"issuer": "issuer",
		}
		values["userspace"] = map[string]interface{}{
			"appCache": "appcache",
			"userData": "userspace/Home",
		}
		values["os"] = map[string]interface{}{
			"appKey":    "appKey",
			"appSecret": "appSecret",
		}
		values["domain"] = map[string]string{}
		values["dep"] = map[string]interface{}{}
		values["svcs"] = map[string]interface{}{}

		// Middleware placeholders
		values["postgres"] = map[string]interface{}{"databases": map[string]interface{}{}}
		values["mariadb"] = map[string]interface{}{"databases": map[string]interface{}{}}
		values["mysql"] = map[string]interface{}{"databases": map[string]interface{}{}}
		values["redis"] = map[string]interface{}{}
		values["mongodb"] = map[string]interface{}{"databases": map[string]interface{}{}}
		values["minio"] = map[string]interface{}{"buckets": map[string]interface{}{}}
		values["rabbitmq"] = map[string]interface{}{"vhosts": map[string]interface{}{}}
		values["elasticsearch"] = map[string]interface{}{"indexes": map[string]interface{}{}}
		values["clickhouse"] = map[string]interface{}{"databases": map[string]interface{}{}}
		values["nats"] = map[string]interface{}{
			"subjects": map[string]interface{}{},
			"refs":     map[string]interface{}{},
		}
	}

	return values, nil
}

func (h *HelmOps) SetValues() (values map[string]interface{}, err error) {
	ctx := context.TODO()

	values, err = BuildBaseHelmValues(ctx, h.kubeConfig, h.app, h.app.OwnerName, false)
	if err != nil {
		return values, err
	}
	err = h.AddEnvironmentVariables(values, false)
	if err != nil {
		return values, err
	}

	// Refine admin: prefer the owner of an already-installed cluster-scoped instance.
	appInstalled, installedApps, err := h.getInstalledApps(ctx)
	if err != nil {
		klog.Errorf("Failed to get installed app err=%v", err)
		return values, err
	}
	if appInstalled {
		for _, a := range installedApps {
			if appcfg.IsClusterScoped(a) {
				values["admin"] = a.Spec.Owner
				break
			}
		}
	}

	zone, err := h.userZone()
	if err != nil {
		klog.Errorf("Failed to find user zone on crd err=%v", err)
	} else if zone != "" {
		values["user"] = map[string]interface{}{
			"zone": zone,
		}
	}

	entries := make(map[string]interface{})
	for i, entrance := range h.app.Entrances {
		var url string
		if len(h.app.Entrances) == 1 {
			url = fmt.Sprintf("%s.%s", h.app.AppID, zone)
		} else {
			url = fmt.Sprintf("%s%d.%s", h.app.AppID, i, zone)
		}
		entries[entrance.Name] = url
	}

	values["domain"] = entries
	userspace := make(map[string]interface{})
	h.app.Permission = ParseAppPermission(h.app.Permission)
	for _, p := range h.app.Permission {
		switch perm := p.(type) {
		case appcfg.AppDataPermission, appcfg.AppCachePermission, appcfg.UserDataPermission:

			// app requests app data permission
			// set .Values.schedule.nodeName and .Values.userspace.appCache to app
			// since app data on the bfl's local hostpath, app will schedule to the same node of bfl
			node, appCachePath, userspacePath, err := h.selectNode()
			if err != nil {
				klog.Errorf("Failed select node err=%v", err)
				return values, err
			}
			values["schedule"] = map[string]interface{}{
				"nodeName": node,
			}

			// appData = userspacePath + /Data
			// userData = userspacePath + /Home

			if perm == appcfg.AppCacheRW {
				userspace["appCache"] = filepath.Join(appCachePath, h.app.AppName)
			}
			if perm == appcfg.UserDataRW {
				userspace["userData"] = fmt.Sprintf("%s/Home", userspacePath)
			}
			if perm == appcfg.AppDataRW {
				appData := fmt.Sprintf("%s/Data", userspacePath)
				userspace["appData"] = filepath.Join(appData, h.app.AppName)
			}

		case []appcfg.ProviderPermission:
			permCfgs, err := apputils.ProviderPermissionsConvertor(perm).ToPermissionCfg(h.ctx, h.app.OwnerName, h.options.MarketSource)
			if err != nil {
				klog.Errorf("Failed to convert app permissions for %s: %v", h.app.AppName, err)
				return values, err
			}
			appReg, err := h.registerAppPerm(h.app.ServiceAccountName, h.app.OwnerName, permCfgs)
			if err != nil {
				klog.Errorf("Failed to register err=%v", err)
				return values, err
			}

			values["os"] = map[string]interface{}{
				"appKey":    appReg.Data.AppKey,
				"appSecret": appReg.Data.AppSecret,
			}
		}
	}
	values["userspace"] = userspace

	// set service entrance for app that depend on cluster-scoped app
	type Service struct {
		EntranceName string
		Host         string
		Port         int
	}
	var services []Service
	appClient := versioned.NewForConfigOrDie(h.kubeConfig)
	apps, err := appClient.AppV1alpha1().Applications().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return values, err
	}

	clusterScopedAppNamespaces := sets.String{}

	for _, dep := range h.app.Dependencies {
		// if app is cluster-scoped get its host and port
		for _, app := range apps.Items {
			if dep.Type == constants.DependencyTypeApp && app.Spec.Name == dep.Name && app.Spec.Settings["clusterScoped"] == "true" {
				clusterScopedAppNamespaces.Insert(app.Spec.Namespace)
				for _, e := range app.Spec.Entrances {
					services = append(services, Service{
						Host:         e.Host + "." + app.Spec.Namespace,
						Port:         int(e.Port),
						EntranceName: e.Name,
					})
				}
			}
		}
	}
	// set cluster-scoped app's host and port to helm Values
	dep := make(map[string]interface{})
	for _, svc := range services {
		dep[fmt.Sprintf("%s_host", svc.EntranceName)] = svc.Host
		dep[fmt.Sprintf("%s_port", svc.EntranceName)] = svc.Port
	}
	values["dep"] = dep

	kClient, err := kubernetes.NewForConfig(h.kubeConfig)
	if err != nil {
		return values, err
	}
	svcs := make(map[string]interface{})
	for ns := range clusterScopedAppNamespaces {
		servicesList, _ := kClient.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
		for _, svc := range servicesList.Items {
			ports := make([]int32, 0)
			for _, p := range svc.Spec.Ports {
				ports = append(ports, p.Port)
			}
			svcs[fmt.Sprintf("%s_host", svc.Name)] = fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)
			svcs[fmt.Sprintf("%s_ports", svc.Name)] = ports
		}
	}
	values["svcs"] = svcs
	klog.Info("svcs: ", svcs)

	if h.app.OIDC.Enabled {
		err = h.createOIDCClient(values, zone, h.app.Namespace)
		if err != nil {
			klog.Errorf("Failed to create OIDCClient err=%v", err)
			return values, err
		}
	}

	klog.Infof("values[node]: %#v", values["nodes"])

	return values, err
}

func (h *HelmOps) TaprApply(values map[string]interface{}, namespace string) error {
	if namespace == "" {
		namespace = fmt.Sprintf("%s-%s", "user-system", h.app.OwnerName)
	}

	if err := tapr.Apply(h.app.Middleware, h.kubeConfig, h.app.AppName, h.app.Namespace,
		namespace, h.token, h.app.OwnerName, values); err != nil {
		klog.Errorf("Failed to apply middleware err=%v", err)
		return err
	}

	return nil
}

func (h *HelmOps) getInstalledApps(ctx context.Context) (installed bool, app []*v1alpha1.Application, err error) {
	var apps *v1alpha1.ApplicationList
	apps, err = h.client.AppClient.AppV1alpha1().Applications().List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list applications err=%v", err)
		return
	}

	for _, a := range apps.Items {
		if a.Spec.Name == h.app.AppName {
			installed = true
			app = append(app, &a)
		}
	}

	return
}

// addEnvironmentVariables reads AppEnv from the cluster and populates the
// olaresEnv helm values key. This is a read-only operation usable from both
// BuildBaseHelmValues and SetValues.
func addEnvironmentVariables(ctx context.Context, kubeConfig *rest.Config, appConfig *appcfg.ApplicationConfig, ownerName string, values map[string]interface{}) error {
	values[constants.OlaresEnvHelmValuesKey] = make(map[string]interface{})

	appClient, err := versioned.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	appEnv, err := appClient.SysV1alpha1().AppEnvs(appConfig.Namespace).Get(ctx, apputils.FormatAppEnvName(appConfig.AppName, ownerName), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, env := range appEnv.Envs {
		values[constants.OlaresEnvHelmValuesKey].(map[string]interface{})[env.EnvName] = env.GetEffectiveValue()
	}

	klog.Infof("Added environment variables to Helm values: %+v", values[constants.OlaresEnvHelmValuesKey])
	return nil
}

// AddEnvironmentVariables populates olaresEnv values and marks the AppEnv as
// applied. When loadValues is true, it reads the env vars from the cluster into
// values first; when false it skips the read (useful when BuildBaseHelmValues
// has already populated them) and only performs the NeedApply side-effect.
func (h *HelmOps) AddEnvironmentVariables(values map[string]interface{}, loadValues bool) error {
	if loadValues {
		if err := addEnvironmentVariables(h.ctx, h.kubeConfig, h.app, h.app.OwnerName, values); err != nil {
			return err
		}
	}

	appEnv, err := h.client.AppClient.SysV1alpha1().AppEnvs(h.app.Namespace).Get(h.ctx, apputils.FormatAppEnvName(h.app.AppName, h.app.OwnerName), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !appEnv.NeedApply {
		return nil
	}

	appEnv.NeedApply = false
	_, err = h.client.AppClient.SysV1alpha1().AppEnvs(h.app.Namespace).Update(h.ctx, appEnv, metav1.UpdateOptions{})
	if err != nil {
		// ignore update error, we use update rather than patch to avoid race condition
		// and the update error may indicate that the appenv is already updated by other request
		// which might need to be applied again
		klog.Errorf("Failed to update appenv %s/%s needApply to false err=%v", appEnv.Namespace, appEnv.Name, err)
	}

	return nil
}
