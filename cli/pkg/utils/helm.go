package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// InitConfigForAppGateway initializes Helm for the app-gateway / Linkerd install path.
func InitConfigForAppGateway(kubeConfig *rest.Config, namespace string) (*action.Configuration, *cli.EnvSettings, error) {
	actionConfig := new(action.Configuration)
	settings := cli.New()
	helmDriver := os.Getenv("HELM_DRIVER")
	settings.SetNamespace(namespace)

	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		settings.KubeConfig = kc
	} else if _, err := os.Stat(clientcmd.RecommendedHomeFile); err == nil {
		settings.KubeConfig = clientcmd.RecommendedHomeFile
	} else {
		settings.KubeAPIServer = kubeConfig.Host
		if kubeConfig.BearerToken != "" {
			settings.KubeToken = kubeConfig.BearerToken
		}
		settings.KubeInsecureSkipTLSVerify = kubeConfig.Insecure
	}

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

// TemplateAndServerSideApply renders a chart and applies manifests with kubectl server-side apply.
func TemplateAndServerSideApply(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
	releaseName, chartPath, namespace string, vals map[string]interface{}) error {
	inst := action.NewInstall(actionConfig)
	inst.ReleaseName = releaseName
	inst.Namespace = namespace
	inst.DryRun = true
	inst.ClientOnly = true
	inst.Replace = true
	inst.IncludeCRDs = true
	inst.Timeout = 300 * time.Second
	if namespace != "" {
		inst.CreateNamespace = true
	}

	chartRequested, err := helmLoader.Load(chartPath)
	if err != nil {
		return err
	}

	rel, err := inst.Run(chartRequested, vals)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace([]byte(rel.Manifest))) == 0 {
		logger.Infof("[helm] %s: no manifests to apply", releaseName)
		return nil
	}
	logger.Infof("[helm] applying %s manifests via kubectl server-side apply", releaseName)
	return kubectlServerSideApply(ctx, settings, rel.Manifest)
}

// KubectlApplyFile runs kubectl apply --server-side on a manifest file.
func KubectlApplyFile(ctx context.Context, settings *cli.EnvSettings, manifestPath string) error {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		return errors.Wrap(err, "kubectl not found in PATH")
	}
	cmd := exec.CommandContext(ctx, kubectl, "apply", "--server-side", "--force-conflicts", "-f", manifestPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if settings != nil && settings.KubeConfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+settings.KubeConfig)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "kubectl apply -f")
	}
	return nil
}

// KubectlApplyDirectory applies all YAML files in a directory via server-side apply.
func KubectlApplyDirectory(ctx context.Context, settings *cli.EnvSettings, dir string) error {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		return errors.Wrap(err, "kubectl not found in PATH")
	}
	cmd := exec.CommandContext(ctx, kubectl, "apply", "--server-side", "--force-conflicts", "-f", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if settings != nil && settings.KubeConfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+settings.KubeConfig)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "kubectl apply -f")
	}
	return nil
}

func kubectlServerSideApply(ctx context.Context, settings *cli.EnvSettings, manifest string) error {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		return errors.Wrap(err, "kubectl not found in PATH")
	}
	cmd := exec.CommandContext(ctx, kubectl, "apply", "--server-side", "--force-conflicts", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(manifest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if settings != nil && settings.KubeConfig != "" {
		cmd.Env = append(os.Environ(), "KUBECONFIG="+settings.KubeConfig)
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "kubectl apply --server-side")
	}
	return nil
}

// LoadValuesFile reads a Helm values YAML file into a map (missing file returns empty map).
func LoadValuesFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return map[string]interface{}{}, nil
	}
	out := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}
