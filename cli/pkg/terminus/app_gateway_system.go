package terminus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/utils"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	appGatewaySystemReleaseName = "app-gateway-system"
	appGatewaySystemDirName     = "app-gateway-system"
)

var (
	validateAppGatewaySystemInstallerArtifactsFunc = ValidateAppGatewaySystemInstallerArtifacts
	getConfigForSystemInstall                      = ctrl.GetConfig
	newClientForSystemInstall                      = func(cfg *rest.Config) (client.Client, error) { return client.New(cfg, client.Options{}) }
	initConfigForSystemInstall                     = utils.InitConfigForAppGateway
	upgradeChartsSkipCRDsWaitFunc                  = utils.UpgradeChartsSkipCRDsWait
	loadAppGatewayDefaultsFunc                     = agwconfig.Load
)

func appGatewaySystemPath(installerDir string) string {
	return filepath.Join(installerDir, "wizard", "config", appGatewaySystemDirName)
}

func appGatewaySystemCRDsPath(installerDir string) string {
	return filepath.Join(appGatewaySystemPath(installerDir), "crds")
}

// ValidateAppGatewaySystemInstaller checks release bundle before cluster install.
type ValidateAppGatewaySystemInstaller struct {
	common.KubeAction
}

func (t *ValidateAppGatewaySystemInstaller) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}
	return validateAppGatewaySystemInstallerArtifactsFunc(resolveInstallerDir(runtime))
}

// InstallAppGatewaySystem installs the app-gateway system umbrella chart.
type InstallAppGatewaySystem struct {
	common.KubeAction
}

func (t *InstallAppGatewaySystem) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}

	installerDir := resolveInstallerDir(runtime)
	if err := validateAppGatewaySystemInstallerArtifactsFunc(installerDir); err != nil {
		return err
	}

	config, err := getConfigForSystemInstall()
	if err != nil {
		return err
	}

	releaseNamespace := resolveAppGatewayNamespace()
	actionConfig, settings, err := initConfigForSystemInstall(config, releaseNamespace)
	if err != nil {
		return err
	}

	defaults, err := loadAppGatewayDefaultsFunc()
	if err != nil {
		defaults = agwconfig.Defaults{}
	}
	vals := buildAppGatewayHelmValues(releaseNamespace, defaults)
	chartPath := appGatewaySystemPath(installerDir)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	k8sClient, err := newClientForSystemInstall(config)
	if err != nil {
		return err
	}
	linkerdNS := agwconfigLinkerdNamespace()
	mat, ok, err := loadLinkerdPKISecret(ctx, k8sClient, linkerdNS)
	if err != nil {
		return fmt.Errorf("load linkerd pki secret: %w", err)
	}
	if !ok || mat == nil {
		return fmt.Errorf("missing %s/%s after PrepareLinkerdPKI", linkerdNS, linkerdPKISecretName)
	}
	linkerdVals, _ := vals["linkerd"].(map[string]interface{})
	if linkerdVals == nil {
		linkerdVals = map[string]interface{}{}
	}
	if err := applyLinkerdPKIToSubchartValues(linkerdVals, mat); err != nil {
		return err
	}
	vals["linkerd"] = linkerdVals

	return upgradeChartsSkipCRDsWaitFunc(
		ctx,
		actionConfig,
		settings,
		appGatewaySystemReleaseName,
		chartPath,
		"",
		releaseNamespace,
		vals,
		false,
	)
}

// ValidateAppGatewaySystemInstallerArtifacts ensures the Olares installer bundle contains app-gateway-system.
func ValidateAppGatewaySystemInstallerArtifacts(installerDir string) error {
	systemDir := appGatewaySystemPath(installerDir)
	checks := []struct {
		relative string
		wantDir  bool
	}{
		{relative: "Chart.yaml"},
		{relative: "values.yaml"},
		{relative: "charts", wantDir: true},
		{relative: filepath.Join("crds", "linkerd-crds"), wantDir: true},
		{relative: filepath.Join("crds", "envoy-gateway-crds"), wantDir: true},
	}

	for _, item := range checks {
		path := filepath.Join(systemDir, item.relative)
		st, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("installer incomplete for app-gateway-system: missing %s", path)
		}
		if item.wantDir && !st.IsDir() {
			return fmt.Errorf("installer incomplete for app-gateway-system: %s is not a directory", path)
		}
	}

	return nil
}
