package utils

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	helmLoader "helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// InitConfigForAppGateway initializes a Helm action configuration for the app-gateway install path.
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

// UpgradeChartsSkipCRDsWait upgrades/installs a chart with CRDs skipped and Helm --wait (including Jobs).
func UpgradeChartsSkipCRDsWait(ctx context.Context, actionConfig *action.Configuration, settings *cli.EnvSettings,
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

// TemplateAndServerSideApply renders a chart (helm template semantics) and applies the manifests
// with kubectl server-side apply. Used for large CRD sets that exceed the Helm release Secret limit.
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
		logger.Infof("[helm] %s: no manifests to apply (check chart values)", releaseName)
		return nil
	}
	logger.Infof("[helm] applying %s manifests via kubectl server-side apply", releaseName)
	return kubectlServerSideApply(ctx, settings, rel.Manifest)
}

func kubectlServerSideApply(ctx context.Context, settings *cli.EnvSettings, manifest string) error {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		return errors.Wrap(err, "kubectl not found in PATH (required for Envoy Gateway CRDs install)")
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

// LoadValuesFile reads a YAML values file into a map (empty/missing file yields an empty map).
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
