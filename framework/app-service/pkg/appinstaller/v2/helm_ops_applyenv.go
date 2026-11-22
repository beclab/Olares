package v2

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/Olares/framework/app-service/pkg/helm"
	"github.com/beclab/Olares/framework/app-service/pkg/kubesphere"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/klog/v2"
)

// applyEnvToSubReleases is the dependency-free core of applyEnvToReleases.
// skip reports whether a chart must be left alone (shared chart, non-admin
// owner, mirroring HelmOpsV2.upgrade); configFor returns the action
// configuration whose storage holds that chart's release.
func applyEnvToSubReleases(ctx context.Context, charts []appcfg.Chart,
	skip func(appcfg.Chart) (bool, error),
	configFor func(appcfg.Chart) (*action.Configuration, error),
	values map[string]interface{}) error {
	for _, chart := range charts {
		skipped, err := skip(chart)
		if err != nil {
			return err
		}
		if skipped {
			klog.Infof("Skipping applyEnv of shared chart %s for non-admin owner", chart.Name)
			continue
		}
		cfg, err := configFor(chart)
		if err != nil {
			return fmt.Errorf("action config for chart %s: %w", chart.Name, err)
		}
		rel, err := action.NewStatus(cfg).Run(chart.Name)
		if err != nil {
			return fmt.Errorf("get release of chart %s: %w", chart.Name, err)
		}
		if err := helm.UpgradeReleaseCharts(ctx, cfg, chart.Name, rel.Namespace, rel.Chart, values); err != nil {
			return fmt.Errorf("apply env to chart %s: %w", chart.Name, err)
		}
	}
	return nil
}

func (h *HelmOpsV2) ApplyEnv() error {
	if !h.isMultiCharts() {
		return h.HelmOps.ApplyEnv()
	}

	if h.isMultiCharts() && !h.hasClusterSharedCharts() {
		err := errors.New("multi-charts app must have at least one cluster shared chart")
		klog.Error(err)
		return err
	}

	_, err := h.status(h.App().AppName)
	if err != nil {
		klog.Errorf("get release status failed %v", err)
		return err
	}

	values := make(map[string]interface{})
	if err := h.AddEnvironmentVariables(values, true); err != nil {
		klog.Errorf("Failed to add environment variables: %v", err)
		return err
	}

	if err := h.applyEnvFromReleases(values); err != nil {
		klog.Errorf("Failed to apply env to app %s err=%v", h.App().AppName, err)
		return err
	}

	if err := h.AddApplicationLabelsToDeployment(); err != nil {
		klog.Errorf("Failed to add application labels to deployment for app %s err=%v", h.App().AppName, err)
		return err
	}

	if h.App().Type == appv1alpha1.Middleware.String() {
		return nil
	}

	if h.Options().SkipWaitForStartUp {
		// App was Stopped (release scaled to zero); the env upgrade keeps it at
		// zero replicas, so there are no pods to wait for.
		klog.Infof("App (V2) %s applyenv with skipWaitForStartUp, not waiting for pods", h.App().AppName)
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

// applyEnvFromReleases updates env values on each sub-chart release using the
// chart already stored in that release. It does not read ChartsName from disk,
// and it does not call upgrade(), which loads a new chart for version upgrades.
func (h *HelmOpsV2) applyEnvFromReleases(values map[string]interface{}) error {
	for _, chart := range h.App().SubCharts {
		if chart.Shared {
			isAdmin, err := kubesphere.IsAdmin(h.Context(), h.KubeConfig(), h.App().OwnerName)
			if err != nil {
				klog.Errorf("Failed to check if user is admin for chart %s: %v", chart.Name, err)
				return err
			}
			if !isAdmin {
				klog.Infof("Skipping apply env of shared chart %s for non-admin user %s", chart.Name, h.App().OwnerName)
				continue
			}
		}

		release, err := h.status(chart.Name)
		if err != nil {
			klog.Errorf("Failed to get status for chart %s: %v", chart.Name, err)
			return err
		}

		actionConfig := h.ActionConfig()
		namespace := release.Namespace
		if chart.Shared {
			ns := appcfg.ChartNamespace(&chart, h.App().OwnerName)
			actionConfig, _, err = helm.InitConfig(h.KubeConfig(), ns)
			if err != nil {
				klog.Errorf("Failed to create action config for shared chart %s: %v", chart.Name, err)
				return err
			}
			if namespace == "" {
				namespace = ns
			}
		} else if namespace == "" {
			namespace = h.App().Namespace
		}

		if err := helm.UpgradeReleaseCharts(h.Context(), actionConfig, chart.Name, namespace, release.Chart, values); err != nil {
			klog.Errorf("Failed to apply env to chart name=%s err=%v", chart.Name, err)
			return err
		}
	}
	return nil
}
