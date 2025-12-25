package middlewareinstaller

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	v1 "github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// Install a middleware with helm client.
func Install(ctx context.Context, kubeConfig *rest.Config, middleware *MiddlewareConfig) error {
	helmClient, err := v1.NewHelmClient(ctx, kubeConfig, middleware.Namespace)
	if err != nil {
		return err
	}

	if installed, err := helmClient.IsInstalled(middleware.MiddlewareName); err != nil {
		klog.Errorf("Failed to get install history workflowName=%s err=%v", middleware.MiddlewareName, err)
		return err
	} else if installed {
		klog.Errorf("workflowName=%s already installed", middleware.MiddlewareName)
		return errors.New(middleware.MiddlewareName + " is already installed")
	}

	vals, err := getSettings(ctx, kubeConfig, middleware)
	if err != nil {
		return err
	}

	err = helmClient.Install(middleware.MiddlewareName, middleware.ChartsName, middleware.RepoURL, middleware.Namespace, vals)
	if err != nil {
		klog.Errorf("Failed to install middleware chart name=%s err=%v", middleware.MiddlewareName, err)
		return err
	}
	return nil
}

// Uninstall a helm release for middleware.
func Uninstall(ctx context.Context, kubeConfig *rest.Config, middleware *MiddlewareConfig) error {
	helmClient, err := v1.NewHelmClient(ctx, kubeConfig, middleware.Namespace)
	if err != nil {
		return err
	}

	installed, err := helmClient.IsInstalled(middleware.MiddlewareName)
	if err != nil {
		klog.Errorf("Failed to get install history middlewareName=%s err=%v", middleware.MiddlewareName, err)
		return err
	}
	if !installed {
		klog.Infof("middleware %s is not installed", middleware.MiddlewareName)
		return nil
	}

	err = helmClient.Uninstall(middleware.MiddlewareName)
	if err != nil {
		klog.Errorf("Failed to uninstall middleware chart name=%s err=%v", middleware.ChartsName, err)
		return err
	}

	return nil
}

func getSettings(ctx context.Context, kubeConfig *rest.Config, middleware *MiddlewareConfig) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	values["bfl"] = map[string]interface{}{
		"username": middleware.OwnerName,
	}

	appData, appCache, userdata, err := getAppData(ctx, kubeConfig, middleware.OwnerName)
	if err != nil {
		klog.Errorf("Failed to get user appdata err=%v", err)
		return nil, err
	}
	userspace := make(map[string]interface{})
	if middleware.Cfg.Permission.AppData {
		userspace["appData"] = appData
	}
	if middleware.Cfg.Permission.AppCache {
		userspace["appCache"] = appCache
	}
	if len(middleware.Cfg.Permission.UserData) > 0 {
		userspace["userData"] = userdata
	}
	values["userspace"] = userspace
	return values, nil
}

func getAppData(ctx context.Context, kubeConfig *rest.Config, owner string) (applicationdata, appdata, userdata string, err error) {
	k8s, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return "", "", "", err
	}

	bfl, err := k8s.AppsV1().StatefulSets("user-space-"+owner).Get(ctx, "bfl",
		metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get bfl err=%v", err)
		return "", "", "", err
	}

	appdata = bfl.Annotations[constants.UserAppDataDirKey]
	if appdata == "" {
		return "", "", "", errors.New("appdata not found")
	}

	userspace, ok := bfl.Annotations[constants.UserSpaceDirKey]
	if !ok {
		return "", "", "", errors.New("userspace not found")
	}

	applicationdata = filepath.Join(userspace, "Data")
	userdata = filepath.Join(userspace, "Home")

	return
}
