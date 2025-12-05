package v1

import (
	"context"
	"errors"

	"bytetrade.io/web3os/app-service/pkg/helm"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// HelmClient contains config for build a helm client.
type HelmClient struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
	ctx          context.Context
}

// NewHelmClient build a helm client.
func NewHelmClient(ctx context.Context, kubeConfig *rest.Config, namespace string) (*HelmClient, error) {
	actionConfig, settings, err := helm.InitConfig(kubeConfig, namespace)
	if err != nil {
		klog.Errorf("Failed to init helm config namespace=%s err=%v", namespace, err)
		return nil, err
	}

	return &HelmClient{
		actionConfig: actionConfig,
		settings:     settings,
		ctx:          ctx,
	}, nil
}

// IsInstalled returns true if a release was deployed.
func (h *HelmClient) IsInstalled(workflowName string) (bool, error) {
	histClient := action.NewHistory(h.actionConfig)
	histClient.Max = 1
	_, err := histClient.Run(workflowName)

	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, nil
		}

		return false, err
	}
	return true, nil
}

// Status returns a release status.
func (h *HelmClient) Status(workflowName string) (bool, *release.Release, error) {
	histClient := action.NewHistory(h.actionConfig)
	histClient.Max = 1
	histories, err := histClient.Run(workflowName)

	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}

	for _, h := range histories {
		if h.Info.Status == release.StatusDeployed {
			return true, h, nil
		}
	}

	return false, histories[0], nil
}

// Version returns an release version.
func (h *HelmClient) Version(workflowName string) (string, error) {
	getClient := action.NewGet(h.actionConfig)
	release, err := getClient.Run(workflowName)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return "", nil
		}
		return "", err
	}

	return release.Chart.Metadata.Version, nil
}

// Install deploy a release with specified chart and values.
func (h *HelmClient) Install(workflowName, chartsName, repoURL, namespace string, vals map[string]interface{}) error {
	return helm.InstallCharts(h.ctx, h.actionConfig, h.settings, workflowName, chartsName, repoURL, namespace, vals)
}

// Uninstall uninstall a release.
func (h *HelmClient) Uninstall(workflowName string) error {
	return helm.UninstallCharts(h.actionConfig, workflowName)
}

// Upgrade upgrade a release with specified values.
func (h *HelmClient) Upgrade(workflowName, chartsName, repoURL, namespace string, vals map[string]interface{}) error {
	return helm.UpgradeCharts(h.ctx, h.actionConfig, h.settings, workflowName, chartsName, repoURL, namespace, vals, false)
}
