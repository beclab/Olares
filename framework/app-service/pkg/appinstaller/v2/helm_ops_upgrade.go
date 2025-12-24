package v2

import (
	"errors"
	"fmt"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	v1 "github.com/beclab/Olares/framework/app-service/pkg/appinstaller"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"

	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/klog/v2"
)

func (h *HelmOpsV2) Upgrade() error {
	if !h.isMultiCharts() {
		return h.HelmOps.Upgrade()
	}

	if h.isMultiCharts() && !h.hasClusterSharedCharts() {
		err := errors.New("multi-charts app must have at least one cluster shared chart")
		klog.Error(err)
		return err
	}

	// We just need to check the status of the main chart,
	// as the upgrade operation will be applied to all sub-charts.
	status, err := h.status(h.App().AppName)
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}
	if status.Info.Status == helmrelease.StatusDeployed {
		values, err := h.SetValues()
		if err != nil {
			klog.Errorf("set values err %v", err)
			return err
		}

		// in v2, if app is multi-charts and has a cluster shared chart,
		//  middleware namespace is always os-platform
		middlewareNamespace := "os-platform"
		err = h.TaprApply(values, middlewareNamespace)
		if err != nil {
			klog.Errorf("tapr apply err %v", err)
			return err
		}

		err = h.upgrade(values)
		if err != nil {
			klog.Errorf("Failed to upgrade app %s err=%v", h.App().AppName, err)
			return err
		}

		// just need to add labels to the main ( client ) chart's deployment
		err = h.AddApplicationLabelsToDeployment()
		if err != nil {
			klog.Errorf("Failed to add application labels to deployment for app %s err=%v", h.App().AppName, err)
			return err
		}

		// add application labels to shared namespace
		err = h.addApplicationLabelsToSharedNamespace()
		if err != nil {
			klog.Errorf("Failed to add application labels to shared namespace err=%v", err)
			return err
		}

		// v2 && multi-charts app, add labels to namespace for cluster scoped apps
		err = h.AddLabelToNamespaceForDependClusterApp()
		if err != nil {
			return err
		}

		if h.App().Type == appv1alpha1.Middleware.String() {
			return nil
		}

		err = h.UpdatePolicy()
		if err != nil {
			klog.Errorf("Failed to update policy for app %s err=%v", h.App().AppName, err)
			return err
		}

		if err = h.RegisterOrUnregisterAppProvider(v1.Register); err != nil {
			klog.Errorf("Failed to register app provider err=%v", err)
			return err
		}

		ok, err := h.WaitForStartUp()
		if err != nil && errors.Is(err, errcode.ErrPodPending) {
			return err
		}

		if !ok {
			klog.Errorf("Failed to wait for app %s startup", h.App().AppName)
			return fmt.Errorf("app %s failed to start up", h.App().AppName)
		}

		klog.Infof("App (V2) %s upgraded successfully", h.App().AppName)
		return nil
	}

	return fmt.Errorf("cannot upgrade release %s/%s, current state is %s", h.App().Namespace, h.App().AppName, status.Info.Status)
}

func (h *HelmOpsV2) upgrade(values map[string]interface{}) error {
	for _, chart := range h.App().SubCharts {
		if chart.Shared {
			isAdmin, err := kubesphere.IsAdmin(h.Context(), h.KubeConfig(), h.App().OwnerName)
			if err != nil {
				klog.Errorf("Failed to check if user is admin for chart %s: %v", chart.Name, err)
				return err
			}

			if !isAdmin {
				klog.Infof("Skipping upgrading of shared chart %s for non-admin user %s", chart.Name, h.App().OwnerName)
				continue
			}
		}

		status, err := h.status(chart.Name)
		if err != nil {
			if errors.Is(err, driver.ErrReleaseNotFound) && chart.Shared {
				// upgrading single chart to multi-charts app,
				klog.Infof("[upgrading] chart %s not found, installing it", chart.Name)
				actionConfig, settings, err := helm.InitConfig(h.KubeConfig(), chart.Namespace(h.App().OwnerName))
				if err != nil {
					klog.Errorf("Failed to create action config for shared chart %s: %v", chart.Name, err)
					return err
				}

				err = helm.InstallCharts(
					h.Context(),
					actionConfig,
					settings,
					chart.Name,
					chart.ChartPath(h.App().AppName),
					h.App().RepoURL,
					chart.Namespace(h.App().OwnerName),
					values,
				)
				if err != nil {
					klog.Errorf("Failed to install chart %s: %v", chart.Name, err)
					return err
				}
			}
			klog.Errorf("Failed to get status for chart %s: %v", chart.Name, err)
			return err
		}

		if status.Info.Status == helmrelease.StatusDeployed {
			actionConfig := h.ActionConfig()
			settings := h.Settings()
			if chart.Shared {
				// re-create action config for shared chart
				actionConfig, settings, err = helm.InitConfig(h.KubeConfig(), chart.Namespace(h.App().OwnerName))
				if err != nil {
					klog.Errorf("Failed to create action config for shared chart %s: %v", chart.Name, err)
					return err
				}
			}

			err = helm.UpgradeCharts(
				h.Context(),
				actionConfig,
				settings,
				chart.Name,
				chart.ChartPath(h.App().AppName),
				h.App().RepoURL,
				h.App().Namespace,
				values,
				true,
			)
			if err != nil {
				klog.Errorf("Failed to upgrade chart name=%s err=%v", chart.Name, err)
				return err
			}
		} else {
			return fmt.Errorf("chart %s is not deployed, cannot upgrade", chart.Name)
		}
	}

	return nil
}
