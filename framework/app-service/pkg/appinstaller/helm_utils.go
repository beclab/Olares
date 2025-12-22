package appinstaller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/appcfg"
	"bytetrade.io/web3os/app-service/pkg/constants"
	"bytetrade.io/web3os/app-service/pkg/generated/clientset/versioned"
	"bytetrade.io/web3os/app-service/pkg/kubesphere"
	"bytetrade.io/web3os/app-service/pkg/tapr"
	userspacev1 "bytetrade.io/web3os/app-service/pkg/users/userspace/v1"
	"bytetrade.io/web3os/app-service/pkg/utils"
	apputils "bytetrade.io/web3os/app-service/pkg/utils/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func (h *HelmOps) SetValues() (values map[string]interface{}, err error) {
	values = make(map[string]interface{})
	values["bfl"] = map[string]interface{}{
		"username": h.app.OwnerName,
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
	h.app.Permission = parseAppPermission(h.app.Permission)
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

	var arch string
	nodes, err := kClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return values, err
	}
	for _, node := range nodes.Items {
		arch = node.Labels["kubernetes.io/arch"]
		break
	}
	values["cluster"] = map[string]interface{}{
		"arch": arch,
	}
	gpuType, err := utils.FindGpuTypeFromNodes(nodes)
	if err != nil {
		klog.Errorf("Failed to get gpuType err=%v", err)
		return values, err
	}
	values["GPU"] = map[string]interface{}{
		"Type": gpuType,
		"Cuda": os.Getenv("OLARES_SYSTEM_CUDA_VERSION"),
	}

	values["gpu"] = gpuType

	if h.app.OIDC.Enabled {
		err = h.createOIDCClient(values, zone, h.app.Namespace)
		if err != nil {
			klog.Errorf("Failed to create OIDCClient err=%v", err)
			return values, err
		}
	}

	sharedLibPath := os.Getenv("SHARED_LIB_PATH")
	values["sharedlib"] = sharedLibPath

	ctx := context.TODO()
	isAdmin, err := kubesphere.IsAdmin(ctx, h.kubeConfig, h.app.OwnerName)
	if err != nil {
		return values, err
	}
	values["isAdmin"] = isAdmin

	var admin string
	appInstalled, installedApps, err := h.getInstalledApps(ctx)
	if err != nil {
		klog.Errorf("Failed to get installed app err=%v", err)
		return values, err
	}

	if appInstalled {
		for _, a := range installedApps {
			if a.IsClusterScoped() {
				admin = a.Spec.Owner
				break
			}
		}
	}

	if admin == "" {
		if isAdmin {
			admin = h.app.OwnerName
		} else {
			// should never happen, but just in case
			klog.Warningf("No admin found for app %s, using the first admin", h.app.AppName)
			admin, err = kubesphere.GetAdminUsername(ctx, h.kubeConfig)
			if err != nil {
				return values, err
			}
		}
	}
	values["admin"] = admin

	rootPath := userspacev1.DefaultRootPath
	if os.Getenv(userspacev1.OlaresRootPath) != "" {
		rootPath = os.Getenv(userspacev1.OlaresRootPath)
	}
	values["rootPath"] = rootPath

	values["downloadCdnURL"] = os.Getenv("OLARES_SYSTEM_CDN_SERVICE")
	values["fs_type"] = utils.EnvOrDefault("OLARES_SYSTEM_ROOTFS_TYPE", "fs")

	terminus, err := utils.GetTerminusVersion(ctx, h.kubeConfig)
	if err != nil {
		return values, err
	}
	values["sysVersion"] = terminus.Spec.Version
	if err := h.AddEnvironmentVariables(values); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return values, err
	}

	nodeInfo, err := utils.GetNodeInfo(ctx)
	if err != nil {
		klog.Errorf("failed to get cluster node info")
		return values, err
	}
	values["nodes"] = nodeInfo

	klog.Infof("values[node]: %#v", values["nodes"])

	deviceName, err := utils.GetDeviceName()
	if err != nil {
		klog.Errorf("failed to get deviceName %v", err)
		return values, err
	}

	values["deviceName"] = deviceName
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

func (h *HelmOps) AddEnvironmentVariables(values map[string]interface{}) error {
	values[constants.OlaresEnvHelmValuesKey] = make(map[string]interface{})
	appEnv, err := h.client.AppClient.SysV1alpha1().AppEnvs(h.app.Namespace).Get(h.ctx, apputils.FormatAppEnvName(h.app.AppName, h.app.OwnerName), metav1.GetOptions{})
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
