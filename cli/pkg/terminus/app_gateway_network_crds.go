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
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	getConfigForNetworkCRDs  = ctrl.GetConfig
	initConfigForNetworkCRDs = utils.InitConfigForAppGateway
	loadValuesForNetworkCRDs = utils.LoadValuesFile
	applyCRDChartFunc        = utils.TemplateAndServerSideApply
	envoyCRDsPresentFunc     = envoyGatewayCRDsPresent
)

// ApplyNetworkCRDs renders the Envoy Gateway / Gateway API CRD chart and applies it
// with kubectl server-side apply (CRD set exceeds the Helm release Secret limit).
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

	if envoyCRDsPresentFunc(config) {
		return nil
	}

	actionConfig, settings, err := initConfigForNetworkCRDs(config, resolveAppGatewayNamespace())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	return applyOneCRDBundle(ctx, actionConfig, settings,
		filepath.Join(appGatewaySystemCRDsPath(installerDir), "envoy-gateway-crds"),
		"app-gateway-envoy-gateway-crds")
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
