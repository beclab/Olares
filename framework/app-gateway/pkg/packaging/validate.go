package packaging

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	archdocVendorDoc = "archdoc/app-gateway/app-gateway-olares-release-and-install-guide-2026-05-15.md"
	vendorLockFile   = "VENDOR_VERSION.lock.yaml"
)

// VendorSourceRel is the path to vendored Helm charts inside the Olares repo (same pattern as GPU .olares/config/gpu).
const VendorSourceRel = "framework/app-gateway/.olares/config/app-gateway-vendor"

// VendorDir returns the absolute vendor directory under an Olares repository root.
func VendorDir(olaresRepoRoot string) string {
	return filepath.Join(olaresRepoRoot, filepath.FromSlash(VendorSourceRel))
}

var requiredVendorChartDirs = []string{
	"envoy-gateway-crds-helm",
	"envoy-gateway-helm",
	"linkerd-crds-chart",
	"linkerd-control-plane-chart",
}

// Installer assets copied from framework/app-gateway/hack and deploy (see hack/sync-vendor-values.sh).
const (
	vendorLinkerdCertScriptRel = "generate-linkerd-identity-certs.sh"
	vendorMeshNetworkPolicyRel = "network-policies/linkerd-mesh-ingress.yaml"
)

var linkerdChartDirs = []string{
	"linkerd-crds-chart",
	"linkerd-control-plane-chart",
}

type chartMeta struct {
	Version string `yaml:"version"`
}

type vendorVersionLock struct {
	LinkerdEdgeChartVersion string `yaml:"linkerd_edge_chart_version"`
	EnvoyGateway            string `yaml:"envoy_gateway"`
}

// ValidateVendorDir ensures vendor contains required charts, lock file, and approved versions.
func ValidateVendorDir(vendorDir string) error {
	for _, name := range requiredVendorChartDirs {
		p := filepath.Join(vendorDir, name)
		st, err := os.Stat(p)
		if err != nil || !st.IsDir() {
			return fmt.Errorf("missing %s under %s (commit complete vendor; see %s)", name, vendorDir, archdocVendorDoc)
		}
	}
	for _, rel := range []string{vendorLinkerdCertScriptRel, vendorMeshNetworkPolicyRel} {
		p := filepath.Join(vendorDir, rel)
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			return fmt.Errorf("missing installer asset %s under %s (run framework/app-gateway/hack/sync-vendor-values.sh)", rel, vendorDir)
		}
	}
	return validateApprovedVersions(vendorDir)
}

func validateApprovedVersions(vendorDir string) error {
	lockPath := filepath.Join(vendorDir, vendorLockFile)
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return fmt.Errorf("missing %s under %s (see %s)", vendorLockFile, vendorDir, archdocVendorDoc)
	}
	var lock vendorVersionLock
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("invalid %s: %w", vendorLockFile, err)
	}
	if lock.LinkerdEdgeChartVersion != LinkerdEdgeChartVersion {
		return fmt.Errorf("%s linkerd_edge_chart_version=%q, want %q (see %s)",
			vendorLockFile, lock.LinkerdEdgeChartVersion, LinkerdEdgeChartVersion, archdocVendorDoc)
	}
	if lock.EnvoyGateway != EnvoyGatewayVersion {
		return fmt.Errorf("%s envoy_gateway=%q, want %q (see %s)",
			vendorLockFile, lock.EnvoyGateway, EnvoyGatewayVersion, archdocVendorDoc)
	}
	for _, dir := range linkerdChartDirs {
		ver, err := readHelmChartVersion(filepath.Join(vendorDir, dir))
		if err != nil {
			return fmt.Errorf("%s: %w", dir, err)
		}
		if ver != LinkerdEdgeChartVersion {
			return fmt.Errorf("%s Chart.yaml version=%q, want %q (see %s)",
				dir, ver, LinkerdEdgeChartVersion, archdocVendorDoc)
		}
	}
	return nil
}

func readHelmChartVersion(chartDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	if err != nil {
		return "", err
	}
	var meta chartMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return "", err
	}
	if meta.Version == "" {
		return "", fmt.Errorf("Chart.yaml has no version field")
	}
	return meta.Version, nil
}

// ValidateInstallerBundle checks paths inside an Olares installer dist (wizard/config/...).
func ValidateInstallerBundle(installerDir string) error {
	systemDir := filepath.Join(installerDir, "wizard", "config", AppGatewaySystemChartName)
	if err := ValidateAppGatewaySystemDir(systemDir); err != nil {
		return fmt.Errorf("installer incomplete for unified ingress: %w", err)
	}
	return nil
}

// ValidateAppGatewaySystemDir ensures app-gateway-system chart assets are present in installer bundle.
func ValidateAppGatewaySystemDir(systemDir string) error {
	chartFile := filepath.Join(systemDir, "Chart.yaml")
	if _, err := os.Stat(chartFile); err != nil {
		return fmt.Errorf("missing app-gateway-system chart file: %s", chartFile)
	}
	if _, err := readHelmChartVersion(systemDir); err != nil {
		return fmt.Errorf("invalid app-gateway-system chart metadata at %s: %w", chartFile, err)
	}

	requiredDirs := []string{
		"charts",
		"crds",
	}
	for _, rel := range requiredDirs {
		p := filepath.Join(systemDir, rel)
		st, err := os.Stat(p)
		if err != nil {
			return fmt.Errorf("missing app-gateway-system asset directory: %s", p)
		}
		if !st.IsDir() {
			return fmt.Errorf("app-gateway-system asset is not a directory: %s", p)
		}
	}
	return nil
}
