package terminus

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/utils"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	getConfigForNetworkCRDs  = ctrl.GetConfig
	initConfigForNetworkCRDs = utils.InitConfigForAppGateway
	loadValuesForNetworkCRDs = utils.LoadValuesFile
	applyCRDChartFunc        = utils.TemplateAndServerSideApply
	linkerdCRDsPresentFunc   = linkerdPolicyCRDsPresent
	envoyCRDsPresentFunc     = envoyGatewayCRDsPresent
)

// ApplyNetworkCRDs renders network CRD charts and applies them with server-side apply.
type ApplyNetworkCRDs struct {
	common.KubeAction
}

func (t *ApplyNetworkCRDs) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}

	installerDir := resolveInstallerDir(runtime)
	if err := validateAppGatewaySystemInstallerArtifactsFunc(installerDir); err != nil {
		return err
	}

	config, err := getConfigForNetworkCRDs()
	if err != nil {
		return err
	}

	linkerdReady := linkerdCRDsPresentFunc(config)
	envoyReady := envoyCRDsPresentFunc(config)
	if linkerdReady && envoyReady {
		return nil
	}

	actionConfig, settings, err := initConfigForNetworkCRDs(config, resolveAppGatewayNamespace())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	if !linkerdReady {
		if err := applyOneCRDBundle(ctx, actionConfig, settings, filepath.Join(appGatewaySystemCRDsPath(installerDir), "linkerd-crds"), "app-gateway-linkerd-crds"); err != nil {
			return err
		}
	}
	if !envoyReady {
		if err := applyOneCRDBundle(ctx, actionConfig, settings, filepath.Join(appGatewaySystemCRDsPath(installerDir), "envoy-gateway-crds"), "app-gateway-envoy-gateway-crds"); err != nil {
			return err
		}
	}

	return nil
}

func applyOneCRDBundle(
	ctx context.Context,
	actionConfig *action.Configuration,
	settings *cli.EnvSettings,
	bundleDir, releaseName string,
) error {
	chartPath := bundleDir
	if st, err := os.Stat(filepath.Join(bundleDir, "chart", "Chart.yaml")); err == nil && !st.IsDir() {
		chartPath = filepath.Join(bundleDir, "chart")
	}

	values, err := loadValuesForNetworkCRDs(filepath.Join(bundleDir, "values.yaml"))
	if err != nil {
		values = map[string]interface{}{}
	}

	return applyCRDChartFunc(ctx, actionConfig, settings, releaseName, chartPath, "", values)
}

func linkerdPolicyCRDsPresent(cfg *rest.Config) bool {
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false
	}
	if _, err := dc.ServerResourcesForGroupVersion("policy.linkerd.io/v1alpha1"); err != nil {
		return false
	}
	return true
}
