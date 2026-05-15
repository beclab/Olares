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
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	appGatewayVendorDirName = "app-gateway-vendor"
	helmReleaseLinkerd      = "linkerd"
	helmReleaseEGCRDs       = "eg-crds"
	helmReleaseEG           = "eg"
	helmReleaseAppGateway   = "app-gateway"
)

func resolveAppGatewayNamespace() string {
	if ns := os.Getenv("APP_GATEWAY_NAMESPACE"); ns != "" {
		return ns
	}
	return agwconfig.Namespace()
}

func vendorNamespaces() (appGatewayNS, linkerdNS string) {
	appGatewayNS = resolveAppGatewayNamespace()
	linkerdNS = agwconfig.LinkerdNamespace()
	return appGatewayNS, linkerdNS
}

func resolveInstallerDir(runtime connector.Runtime) string {
	if d := os.Getenv("OLARES_INSTALLER_DIR"); d != "" {
		return d
	}
	return runtime.GetInstallerDir()
}

func appGatewayVendorPath(installerDir string) string {
	return filepath.Join(installerDir, "wizard", "config", appGatewayVendorDirName)
}

func appGatewayHelmChartPath(installerDir string) string {
	return filepath.Join(installerDir, "wizard", "config", "apps", "app-gateway")
}

func appGatewayStackEnabled() bool {
	v := os.Getenv("APP_GATEWAY_STACK_ENABLED")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

// InstallAppGatewayVendor installs Linkerd control plane and Envoy Gateway using helm SDK only.
type InstallAppGatewayVendor struct {
	common.KubeAction
}

func (t *InstallAppGatewayVendor) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		logger.Info("APP_GATEWAY_STACK_ENABLED is off; skipping app-gateway vendor install")
		return nil
	}

	vendor := appGatewayVendorPath(resolveInstallerDir(runtime))
	if _, err := os.Stat(filepath.Join(vendor, "envoy-gateway-helm")); err != nil {
		return errors.Wrapf(err, "app-gateway vendor charts not found under installer (run Olares package build with bundle-vendor-charts)")
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	appGatewayNS, linkerdNamespace := vendorNamespaces()

	// Linkerd control plane
	linkerdChart := filepath.Join(vendor, "linkerd-control-plane-chart")
	if st, err := os.Stat(linkerdChart); err == nil && st.IsDir() {
		actionConfig, settings, err := utils.InitConfig(config, linkerdNamespace)
		if err != nil {
			return err
		}
		linkerdVals, err := utils.LoadValuesFile(filepath.Join(vendor, "linkerd-values.yaml"))
		if err != nil {
			linkerdVals = map[string]interface{}{}
		}
		logger.InfoInstallationProgress("Installing Linkerd control plane (helm SDK) ...")
		if err := utils.UpgradeCharts(ctx, actionConfig, settings, helmReleaseLinkerd, linkerdChart, "", linkerdNamespace, linkerdVals, false); err != nil {
			return errors.Wrap(err, "install linkerd control plane")
		}
	} else {
		logger.Warn("linkerd-control-plane-chart not bundled; skip Linkerd install")
	}

	// Envoy Gateway CRDs + control plane
	egVals, err := utils.LoadValuesFile(filepath.Join(vendor, "envoy-gateway-values.yaml"))
	if err != nil {
		egVals = map[string]interface{}{}
	}

	actionEG, settingsEG, err := utils.InitConfig(config, appGatewayNS)
	if err != nil {
		return err
	}

	crdsChart := filepath.Join(vendor, "envoy-gateway-crds-helm")
	logger.InfoInstallationProgress(fmt.Sprintf("Installing Envoy Gateway CRDs into %s (helm SDK) ...", appGatewayNS))
	if err := utils.UpgradeCharts(ctx, actionEG, settingsEG, helmReleaseEGCRDs, crdsChart, "", appGatewayNS, egVals, false); err != nil {
		return errors.Wrap(err, "install envoy gateway crds")
	}

	egChart := filepath.Join(vendor, "envoy-gateway-helm")
	logger.InfoInstallationProgress(fmt.Sprintf("Installing Envoy Gateway control plane into %s (helm SDK) ...", appGatewayNS))
	if err := utils.UpgradeCharts(ctx, actionEG, settingsEG, helmReleaseEG, egChart, "", appGatewayNS, egVals, false); err != nil {
		return errors.Wrap(err, "install envoy gateway")
	}

	return nil
}

// WaitAppGatewayReady waits for EG control plane and optional demo gateway programmed.
type WaitAppGatewayReady struct {
	common.KubeAction
}

func (t *WaitAppGatewayReady) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	c, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	appGatewayNS, _ := vendorNamespaces()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		var eg appsv1.Deployment
		err = c.Get(ctx, types.NamespacedName{Namespace: appGatewayNS, Name: "envoy-gateway"}, &eg)
		if err == nil && eg.Status.ReadyReplicas >= 1 {
			logger.InfoInstallationProgress("Envoy Gateway control plane is ready")
			return nil
		}
		if err != nil {
			err = errors.Wrap(err, "envoy-gateway deployment not found")
		} else {
			err = fmt.Errorf("envoy-gateway not ready yet")
		}

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
		}
	}
}

// InstallAppGatewayChart installs Gateway API resources into namespace from config/defaults.yaml.
type InstallAppGatewayChart struct {
	common.KubeAction
}

func (t *InstallAppGatewayChart) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		return nil
	}

	installerDir := resolveInstallerDir(runtime)
	chartPath := appGatewayHelmChartPath(installerDir)
	if _, err := os.Stat(chartPath); err != nil {
		return errors.Wrapf(err, "app-gateway helm chart not found at %s (run Olares package.sh)", chartPath)
	}

	ns := resolveAppGatewayNamespace()
	def, err := agwconfig.Load()
	if err != nil {
		def = agwconfig.Defaults{}
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	vals := map[string]interface{}{
		"namespace": ns,
		"gateway": map[string]interface{}{
			"name":             def.Gateway.Name,
			"gatewayClassName": def.Gateway.GatewayClassName,
		},
		"demo": map[string]interface{}{
			"enabled": def.Demo.Enabled,
			"host":    def.Demo.Host,
		},
	}
	if def.Gateway.Name == "" || def.Gateway.GatewayClassName == "" {
		vals["gateway"] = map[string]interface{}{
			"name":             "app-gateway",
			"gatewayClassName": "olares-app-gateway",
		}
	}

	logger.InfoInstallationProgress(fmt.Sprintf("Installing app-gateway chart into namespace %s (helm SDK) ...", ns))
	return utils.UpgradeCharts(ctx, actionConfig, settings, helmReleaseAppGateway, chartPath, "", ns, vals, false)
}
