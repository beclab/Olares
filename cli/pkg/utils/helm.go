package utils

import (
	"context"
	"fmt"
	"helm.sh/helm/v3/pkg/storage/driver"
	"os"
	"time"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
)

// Config helm client config.
type Config struct {
	ActionCfg *action.Configuration
	Settings  *cli.EnvSettings
}

func debug(format string, v ...interface{}) {
	if false {
		format = fmt.Sprintf("[helm] [debug] %s\n", format)
		logger.Debug(fmt.Sprintf(format, v...))
	}
}

// InitConfig initializes the configuration for executing actions.
func InitConfig(kubeConfig *rest.Config, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	actionConfig := new(action.Configuration)
	var settings = cli.New()
	helmDriver := os.Getenv("HELM_DRIVER")
	settings.KubeAPIServer = kubeConfig.Host
	settings.KubeToken = kubeConfig.BearerToken
	settings.KubeInsecureSkipTLSVerify = true
	settings.SetNamespace(namespace)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, helmDriver, debug); err != nil {
		logger.Error(err, "helm config init error")
		return nil, nil, err
	}

	return actionConfig, settings, nil
}

// InstallCharts installs helm chart using action config and environment settings.
func InstallCharts(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartsName, repoURL, namespace string, vals map[string]interface{}) error {

	instClient := action.NewInstall(actionConfig)
	if namespace == "" {
		instClient.CreateNamespace = false
	} else {
		instClient.CreateNamespace = true
	}
	instClient.Namespace = namespace
	instClient.Timeout = 300 * time.Second

	if repoURL != "" {
		instClient.RepoURL = repoURL
	}

	r, err := runInstall(ctx, []string{appName, chartsName}, instClient, settings, vals)
	if err != nil {
		// todo uninstall
		// delete failed install
		// do not need delete release, helm will delete it automatically if failed
		// if r != nil {
		// 	deleteCli := hin.newUninstallClient(hin.actionconfig)
		// 	errDel := hin.runUninstall(hin.app.AppName, deleteCli)
		// 	if errDel != nil {
		// 		ctrl.Log.Error(errDel, "delete the app error", "appname", hin.app.AppName, "namespace", hin.app.Namespace)
		// 	}
		// }
		return err
	}
	logReleaseInfo(r)

	return nil
}

// UpgradeCharts upgrades helm chart using action config and environment settings.
func UpgradeCharts(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	appName, chartName, repoURL, namespace string, vals map[string]interface{}, reuseValue bool) error {
	client := action.NewUpgrade(actionConfig)
	client.Namespace = namespace
	client.Timeout = 300 * time.Second
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
		return InstallCharts(ctx, actionConfig, settings, appName, chartName, repoURL, namespace, vals)
	}
	logReleaseUpgrade(r)
	return nil
}

// UninstallCharts upgrades helm chart using action config.
func UninstallCharts(cfg *action.Configuration, releaseName string) error {
	uninstall := action.NewUninstall(cfg)
	uninstall.KeepHistory = false
	r, err := uninstall.Run(releaseName)
	if err != nil {
		if r != nil && r.Release != nil && r.Release.Info != nil &&
			r.Release.Info.Status == release.StatusUninstalled {
			return nil
		}
		return err
	}
	logUninstallReleaseInfo(r)
	return nil
}

// RollbackCharts rollback helm chart using action config.
func RollbackCharts(cfg *action.Configuration, releaseName string) error {
	rollback := action.NewRollback(cfg)
	err := rollback.Run(releaseName)
	if err != nil {
		return err
	}
	return nil
}

func runUpgrade(ctx context.Context, args []string, client *action.Upgrade, settings *cli.EnvSettings, vals map[string]interface{}) (*release.Release, error) {
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	cp, err := client.ChartPathOptions.LocateChart(args[1], settings)
	if err != nil {
		return nil, err
	}
	p := getter.All(settings)

	chartRequested, err := helmLoader.Load(cp)
	if err != nil {
		return nil, err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = helmLoader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}
	return client.RunWithContext(ctx, args[0], chartRequested, vals)

}

func runInstall(ctx context.Context, args []string, client *action.Install, settings *cli.EnvSettings, vals map[string]interface{}) (*release.Release, error) {
	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}

	name, c, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(c, settings)
	if err != nil {
		return nil, err
	}
	p := getter.All(settings)
	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := helmLoader.Load(cp)

	if err != nil {
		return nil, err
	}
	if req := chartRequested.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					return nil, err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = helmLoader.Load(cp); err != nil {
					return nil, errors.Wrap(err, "failed reloading chart after repo update")
				}
			} else {
				return nil, err
			}
		}
	}

	return client.RunWithContext(ctx, chartRequested, vals)
}

// ReleaseName returns application release name.
func ReleaseName(appname, owner string) string {
	return fmt.Sprintf("%s-%s", appname, owner)
}

func logReleaseInfo(release *release.Release) {
	logger.Infow("[helm] app installed success",
		"NAME", release.Name,
		"LAST DEPLOYED", release.Info.LastDeployed.Format(time.ANSIC),
		"NAMESPACE", release.Namespace,
		"STATUS", release.Info.Status.String(),
		"REVISION", release.Version)
}

func logUninstallReleaseInfo(release *release.UninstallReleaseResponse) {
	logger.Infow("[helm] app uninstalled success",
		"NAME", release.Release.Name,
		"NAMESPACE", release.Release.Namespace,
		"INFO", release.Info)
}

func logReleaseUpgrade(release *release.Release) {
	logger.Infow("[helm] app upgrade success",
		"NAME", release.Name,
		"LAST DEPLOYED", release.Info.LastDeployed.Format(time.ANSIC),
		"NAMESPACE", release.Namespace,
		"STATUS", release.Info.Status.String(),
		"REVISION", release.Version)
}
