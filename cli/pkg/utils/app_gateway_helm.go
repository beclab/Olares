package utils

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// UpgradeChartsSkipCRDsWait upgrades/installs a chart with CRDs skipped and Helm --wait (including Jobs).
// App-gateway Envoy Gateway install only; does not alter UpgradeCharts / upgradeCharts behavior.
func UpgradeChartsSkipCRDsWait(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartName, repoURL, namespace string, vals map[string]interface{}, reuseValue bool) error {
	return upgradeChartsSkipCRDsWait(ctx, actionConfig, settings, appName, chartName, repoURL, namespace, vals, reuseValue)
}

func upgradeChartsSkipCRDsWait(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartName, repoURL, namespace string, vals map[string]interface{}, reuseValue bool) error {
	client := action.NewUpgrade(actionConfig)
	client.Namespace = namespace
	client.Timeout = 10 * time.Minute
	client.SkipCRDs = true
	client.Wait = true
	client.WaitForJobs = true
	if reuseValue {
		client.ReuseValues = true
	}
	if repoURL != "" {
		client.RepoURL = repoURL
	}
	r, err := runUpgrade(ctx, []string{appName, chartName}, client, settings, vals)
	if err != nil {
		if !errors.Is(err, driver.ErrNoDeployedReleases) {
			return err
		}
		return installChartsSkipCRDsWait(ctx, actionConfig, settings, appName, chartName, repoURL, namespace, vals)
	}
	logReleaseUpgrade(r)
	return nil
}

func installChartsSkipCRDsWait(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartsName, repoURL, namespace string, vals map[string]interface{}) error {
	instClient := action.NewInstall(actionConfig)
	instClient.CreateNamespace = namespace != ""
	instClient.Namespace = namespace
	instClient.Timeout = 10 * time.Minute
	instClient.SkipCRDs = true
	instClient.Wait = true
	instClient.WaitForJobs = true
	if repoURL != "" {
		instClient.RepoURL = repoURL
	}
	r, err := runInstall(ctx, []string{appName, chartsName}, instClient, settings, vals)
	if err != nil {
		return err
	}
	logReleaseInfo(r)
	return nil
}
