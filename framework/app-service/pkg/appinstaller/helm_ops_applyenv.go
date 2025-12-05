package appinstaller

import (
	"fmt"

	appv1alpha1 "bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/app-service/pkg/helm"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"k8s.io/klog/v2"
)

func (h *HelmOps) ApplyEnv() error {
	status, err := h.status()
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}
	if status.Info.Status != helmrelease.StatusDeployed {
		return fmt.Errorf("cannot upgrade release %s/%s, current state is %s", h.app.Namespace, h.app.AppName, status.Info.Status)
	}

	values := make(map[string]interface{})
	if err := h.AddEnvironmentVariables(values); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return err
	}

	err = helm.UpgradeCharts(h.ctx, h.actionConfig, h.settings, h.app.AppName, h.app.ChartsName, h.app.RepoURL, h.app.Namespace, values, true)
	if err != nil {
		klog.Errorf("Failed to upgrade chart name=%s err=%v", h.app.AppName, err)
		return err
	}

	if err = h.AddApplicationLabelsToDeployment(); err != nil {
		return err
	}
	if h.app.Type == appv1alpha1.Middleware.String() {
		return nil
	}
	ok, err := h.WaitForStartUp()
	if !ok {
		klog.Errorf("Failed to wait for app %s startup", h.app.AppName)
		return fmt.Errorf("app %s failed to start up", h.app.AppName)
	}
	klog.Infof("App %s applyenv successfully", h.app.AppName)
	return nil
}
