package middlewareinstaller

import (
	"context"
	v1 "github.com/beclab/Olares/framework/app-service/pkg/workflowinstaller/v1"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

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
