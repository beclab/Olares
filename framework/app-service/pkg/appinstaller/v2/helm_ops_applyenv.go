package v2

import (
	"errors"
	"fmt"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"k8s.io/klog/v2"
)

func (h *HelmOpsV2) ApplyEnv() error {
	if !h.isMultiCharts() {
		return h.HelmOps.ApplyEnv()
	}

	if h.isMultiCharts() && !h.hasClusterSharedCharts() {
		err := errors.New("multi-charts app must have at least one cluster shared chart")
		klog.Error(err)
		return err
	}

	status, err := h.status(h.App().AppName)
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}
	if status.Info.Status != helmrelease.StatusDeployed {
		return fmt.Errorf("cannot upgrade release %s/%s, current state is %s", h.App().Namespace, h.App().AppName, status.Info.Status)
	}

	values := make(map[string]interface{})
	if err := h.AddEnvironmentVariables(values); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return err
	}

	if err := h.upgrade(values); err != nil {
		klog.Errorf("Failed to upgrade app %s err=%v", h.App().AppName, err)
		return err
	}

	if err := h.AddApplicationLabelsToDeployment(); err != nil {
		klog.Errorf("Failed to add application labels to deployment for app %s err=%v", h.App().AppName, err)
		return err
	}

	if h.App().Type == appv1alpha1.Middleware.String() {
		return nil
	}

	ok, err := h.WaitForStartUp()
	if err != nil && (errors.Is(err, errcode.ErrPodPending) || errors.Is(err, errcode.ErrServerSidePodPending)) {
		return err
	}

	if !ok {
		klog.Errorf("Failed to wait for app %s startup", h.App().AppName)
		return fmt.Errorf("app %s failed to start up", h.App().AppName)
	}

	klog.Infof("App (V2) %s applyenv successfully", h.App().AppName)
	return nil
}
