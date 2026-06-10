package terminus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
	agwpack "github.com/beclab/Olares/framework/app-gateway/pkg/packaging"

	"github.com/beclab/Olares/cli/pkg/common"
	"github.com/beclab/Olares/cli/pkg/core/connector"
	"github.com/beclab/Olares/cli/pkg/core/logger"
	"github.com/beclab/Olares/cli/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	appGatewayVendorDirName = "app-gateway-vendor"
	helmReleaseLinkerdCRDs  = "linkerd-crds"
	helmReleaseLinkerd      = "linkerd"
	helmReleaseEGCRDs       = "eg-crds"
	helmReleaseEG           = "eg"
	helmReleaseAppGateway   = "app-gateway"

	envMeshProfileBootstrap = "APP_GATEWAY_MESH_PROFILE"
	meshProfileFull         = "full"
	meshProfileLite         = "lite"
	clusterConfigSingleton  = "cluster"
)

var (
	upgradeChartsForVendorInstallFunc = utils.UpgradeCharts
)

var clusterConfigGVK = schema.GroupVersionKind{
	Group:   "cluster.olares.io",
	Version: "v1alpha1",
	Kind:    "ClusterConfig",
}

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

func resolveMeshProfileFromEnv() string {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(envMeshProfileBootstrap)))
	if v == meshProfileLite {
		return meshProfileLite
	}
	return meshProfileFull
}

func meshProfileSkipsLinkerdInstall(profile string) bool {
	return profile == meshProfileLite
}

func validateMeshProfileHelmAlignment(profile string, def agwconfig.Defaults) error {
	if profile == meshProfileLite && def.MeshLinkerdEnabled() {
		return fmt.Errorf("mesh profile lite conflicts with defaults.yaml mesh.linkerd.enabled=true")
	}
	return nil
}

func applyLiteMeshHelmOverrides(vals map[string]interface{}) {
	mesh, ok := vals["mesh"].(map[string]interface{})
	if !ok {
		mesh = map[string]interface{}{}
		vals["mesh"] = mesh
	}
	linkerd, ok := mesh["linkerd"].(map[string]interface{})
	if !ok {
		linkerd = map[string]interface{}{}
		mesh["linkerd"] = linkerd
	}
	linkerd["enabled"] = false
	vals["linkerdPkiGuardian"] = map[string]interface{}{"enabled": false}
}

func bootstrapClusterConfigMeshProfile(ctx context.Context, c client.Client, profile string) error {
	if c == nil || profile == "" {
		return nil
	}
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(clusterConfigGVK)
	err := c.Get(ctx, types.NamespacedName{Name: clusterConfigSingleton}, obj)
	if apierrors.IsNotFound(err) {
		obj.SetName(clusterConfigSingleton)
		if err := unstructured.SetNestedField(obj.Object, profile, "spec", "meshProfile"); err != nil {
			return err
		}
		return c.Create(ctx, obj)
	}
	if err != nil {
		return err
	}
	current, _, err := unstructured.NestedString(obj.Object, "spec", "meshProfile")
	if err != nil {
		return err
	}
	if current == profile {
		return nil
	}
	if err := unstructured.SetNestedField(obj.Object, profile, "spec", "meshProfile"); err != nil {
		return err
	}
	return c.Update(ctx, obj)
}

// appGatewayStackEnabled reports whether the unified ingress stack is part of this run.
// Default is on for Olares install/upgrade; set APP_GATEWAY_STACK_ENABLED=0 only for exceptional dev skips.
func appGatewayStackEnabled() bool {
	v := os.Getenv("APP_GATEWAY_STACK_ENABLED")
	return v == "" || v == "1" || v == "true" || v == "TRUE"
}

// ValidateAppGatewayInstallerArtifacts ensures the Olares installer bundle contains vendor + chart.
func ValidateAppGatewayInstallerArtifacts(installerDir string) error {
	if err := agwpack.ValidateInstallerBundle(installerDir); err != nil {
		return errors.Wrap(err, "app-gateway")
	}
	return nil
}

// InstallAppGatewayVendor installs Linkerd (CRDs chart then control plane) and Envoy Gateway using helm SDK only.
type InstallAppGatewayVendor struct {
	common.KubeAction
}

func (t *InstallAppGatewayVendor) Execute(runtime connector.Runtime) error {
	if !appGatewayStackEnabled() {
		logger.Info("APP_GATEWAY_STACK_ENABLED is off; skipping app-gateway vendor install")
		return nil
	}

	installerDir := resolveInstallerDir(runtime)
	if err := ValidateAppGatewayInstallerArtifacts(installerDir); err != nil {
		return err
	}
	vendor := appGatewayVendorPath(installerDir)

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	meshProfile := resolveMeshProfileFromEnv()
	defaults, err := agwconfig.Load()
	if err != nil {
		defaults = agwconfig.Defaults{}
	}
	if err := validateMeshProfileHelmAlignment(meshProfile, defaults); err != nil {
		return err
	}

	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	if err := bootstrapClusterConfigMeshProfile(ctx, k8sClient, meshProfile); err != nil {
		return errors.Wrap(err, "bootstrap ClusterConfig.spec.meshProfile")
	}

	appGatewayNS, linkerdNamespace := vendorNamespaces()

	if !meshProfileSkipsLinkerdInstall(meshProfile) {
		actionLinkerd, settingsLinkerd, err := utils.InitConfigForAppGateway(config, linkerdNamespace)
		if err != nil {
			return err
		}
		linkerdCRDsChart := filepath.Join(vendor, "linkerd-crds-chart")
		linkerdCPChart := filepath.Join(vendor, "linkerd-control-plane-chart")
		crdsVals, err := utils.LoadValuesFile(filepath.Join(vendor, "linkerd-crds-values.yaml"))
		if err != nil {
			crdsVals = map[string]interface{}{}
		}
		logger.InfoInstallationProgress("Installing Linkerd CRDs chart (helm SDK) ...")
		if err := upgradeChartsForVendorInstallFunc(ctx, actionLinkerd, settingsLinkerd, helmReleaseLinkerdCRDs, linkerdCRDsChart, "", linkerdNamespace, crdsVals, false); err != nil {
			return errors.Wrap(err, "install linkerd-crds")
		}

		linkerdVals, err := utils.LoadValuesFile(filepath.Join(vendor, "linkerd-values.yaml"))
		if err != nil {
			linkerdVals = map[string]interface{}{}
		}
		if err := enrichLinkerdHelmValues(ctx, k8sClient, linkerdNamespace, vendor, linkerdVals); err != nil {
			return errors.Wrap(err, "prepare linkerd identity certificates")
		}
		logger.InfoInstallationProgress("Installing Linkerd control plane (helm SDK) ...")
		if err := upgradeChartsForVendorInstallFunc(ctx, actionLinkerd, settingsLinkerd, helmReleaseLinkerd, linkerdCPChart, "", linkerdNamespace, linkerdVals, false); err != nil {
			return errors.Wrap(err, "install linkerd control plane")
		}
		if err := applyLinkerdMeshNetworkPolicies(ctx, k8sClient, settingsLinkerd, vendor); err != nil {
			return errors.Wrap(err, "apply linkerd mesh network policies")
		}
	} else {
		logger.InfoInstallationProgress("meshProfile=lite: skipping Linkerd CRDs and control plane install")
	}

	// Envoy Gateway CRDs + control plane
	egVals, err := utils.LoadValuesFile(filepath.Join(vendor, "envoy-gateway-values.yaml"))
	if err != nil {
		egVals = map[string]interface{}{}
	}
	egCRDsVals, err := utils.LoadValuesFile(filepath.Join(vendor, "envoy-gateway-crds-values.yaml"))
	if err != nil {
		egCRDsVals = map[string]interface{}{}
	}

	actionEG, settingsEG, err := utils.InitConfigForAppGateway(config, appGatewayNS)
	if err != nil {
		return err
	}

	if err := ensureAppGatewayNamespace(ctx, k8sClient, appGatewayNS); err != nil {
		return errors.Wrap(err, "ensure app-gateway namespace")
	}

	crdsChart := filepath.Join(vendor, "envoy-gateway-crds-helm")
	if envoyGatewayCRDsPresent(config) {
		logger.InfoInstallationProgress("Envoy Gateway CRDs already present; skip server-side apply")
	} else {
		logger.InfoInstallationProgress(fmt.Sprintf(
			"Installing Envoy Gateway CRDs into %s (helm template + kubectl server-side apply; first install may take several minutes) ...",
			appGatewayNS))
		if err := utils.TemplateAndServerSideApply(ctx, actionEG, settingsEG, helmReleaseEGCRDs, crdsChart, appGatewayNS, egCRDsVals); err != nil {
			return errors.Wrap(err, "install envoy gateway crds")
		}
	}

	egChart := filepath.Join(vendor, "envoy-gateway-helm")
	logger.InfoInstallationProgress(fmt.Sprintf(
		"Installing Envoy Gateway control plane into %s (helm SDK, wait for certgen Job + envoy-gateway deployment; typically 1–3 min) ...",
		appGatewayNS))
	if err := UpgradeEnvoyGatewayHelmWait(ctx, actionEG, settingsEG, helmReleaseEG, egChart, "", appGatewayNS, egVals, false); err != nil {
		return errors.Wrap(err, "install envoy gateway")
	}
	logger.InfoInstallationProgress("Envoy Gateway Helm release eg is deployed; verifying control plane ...")
	if err := waitEnvoyGatewayControlPlaneReady(ctx, k8sClient, appGatewayNS, 3*time.Minute); err != nil {
		return errors.Wrap(err, "verify envoy gateway control plane")
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	appGatewayNS, _ := vendorNamespaces()
	return waitEnvoyGatewayControlPlaneReady(ctx, c, appGatewayNS, 10*time.Minute)
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
	if err := ValidateAppGatewayInstallerArtifacts(installerDir); err != nil {
		return err
	}
	chartPath := appGatewayHelmChartPath(installerDir)

	ns := resolveAppGatewayNamespace()
	def, err := agwconfig.Load()
	if err != nil {
		def = agwconfig.Defaults{}
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfigForAppGateway(config, ns)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	meshProfile := resolveMeshProfileFromEnv()
	if err := validateMeshProfileHelmAlignment(meshProfile, def); err != nil {
		return err
	}

	vals := buildAppGatewayHelmValues(ns, def)
	if meshProfileSkipsLinkerdInstall(meshProfile) {
		applyLiteMeshHelmOverrides(vals)
	}

	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return err
	}
	if err := bootstrapClusterConfigMeshProfile(ctx, k8sClient, meshProfile); err != nil {
		return errors.Wrap(err, "bootstrap ClusterConfig.spec.meshProfile")
	}
	if err := ensureAppGatewayNamespaceMetadata(ctx, k8sClient, ns); err != nil {
		return errors.Wrap(err, "prepare app-gateway namespace")
	}

	logger.InfoInstallationProgress(fmt.Sprintf("Installing app-gateway chart into namespace %s (helm SDK, EnvoyProxy mesh=%v) ...", ns, def.MeshLinkerdEnabled()))
	if err := utils.UpgradeChartsInExistingNamespace(ctx, actionConfig, settings, helmReleaseAppGateway, chartPath, "", ns, vals, false); err != nil {
		return err
	}
	if def.MeshLinkerdEnabled() {
		if err := applyLinkerdMeshNetworkPolicies(ctx, k8sClient, settings, appGatewayVendorPath(installerDir)); err != nil {
			return errors.Wrap(err, "apply linkerd mesh network policies")
		}
	}
	return finalizeAppGatewayMesh(ctx, k8sClient, ns, def)
}

// ensureAppGatewayNamespaceMetadata applies Olares platform labels/annotations when EG install
// created the namespace via helm --create-namespace (no chart-owned Namespace manifest).
func ensureAppGatewayNamespaceMetadata(ctx context.Context, c client.Client, ns string) error {
	var existing corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: ns}, &existing); err != nil {
		return err
	}
	patch := client.MergeFrom(existing.DeepCopy())
	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	existing.Labels["app.kubernetes.io/name"] = "app-gateway"
	existing.Labels["applications.app.bytetrade.io/author"] = "bytetrade.io"
	if existing.Annotations == nil {
		existing.Annotations = map[string]string{}
	}
	// Do not set linkerd.io/inject on the namespace: EG data plane uses EnvoyProxy pod annotations only.
	delete(existing.Annotations, "linkerd.io/inject")
	existing.Annotations["bytetrade.io/ns-type"] = "platform"
	existing.Labels["bytetrade.io/ns-type"] = "system"
	if err := c.Patch(ctx, &existing, patch); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	return nil
}
