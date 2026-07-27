package appinstaller

import (
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
)

func (h *HelmOps) ApplyEnv() error {
	_, err := h.status()
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}

	values := make(map[string]interface{})
	if err := h.AddEnvironmentVariables(values, true); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return err
	}

	// ReuseValues: only env-related overrides change; do not absorb new chart defaults.
	err = helm.UpgradeCharts(h.ctx, h.actionConfig, h.settings, h.app.AppName, h.app.ChartsName, h.app.RepoURL, h.app.Namespace, values, helm.ReuseValues)
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
	if h.options.SkipWaitForStartUp {
		// App was Stopped (release scaled to zero); the env upgrade keeps it at
		// zero replicas, so there are no pods to wait for.
		klog.Infof("App %s applyenv with skipWaitForStartUp, not waiting for pods", h.app.AppName)
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
