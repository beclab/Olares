package packaging

import (
	"os"
	"path/filepath"
	"testing"
)

func writeChart(t *testing.T, dir, version string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("version: "+version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeVendorLock(t *testing.T, vendorDir string) {
	t.Helper()
	body := "linkerd_edge_chart_version: \"2026.5.1\"\nenvoy_gateway: \"v1.8.0\"\n"
	if err := os.WriteFile(filepath.Join(vendorDir, vendorLockFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func populateVendor(t *testing.T, vendorDir string) {
	t.Helper()
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeVendorLock(t, vendorDir)
	for _, name := range requiredVendorChartDirs {
		ver := "0.0.0"
		if name == "linkerd-crds-chart" || name == "linkerd-control-plane-chart" {
			ver = LinkerdEdgeChartVersion
		}
		writeChart(t, filepath.Join(vendorDir, name), ver)
	}
}

func TestValidateVendorDir_missing(t *testing.T) {
	dir := t.TempDir()
	if err := ValidateVendorDir(dir); err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestValidateVendorDir_ok(t *testing.T) {
	dir := t.TempDir()
	populateVendor(t, dir)
	if err := ValidateVendorDir(dir); err != nil {
		t.Fatal(err)
	}
}

func TestValidateVendorDir_linkerdVersionMismatch(t *testing.T) {
	dir := t.TempDir()
	populateVendor(t, dir)
	writeChart(t, filepath.Join(dir, "linkerd-crds-chart"), "2099.0.0")
	if err := ValidateVendorDir(dir); err == nil {
		t.Fatal("expected version mismatch error")
	}
}

func TestValidateInstallerBundle_missingChart(t *testing.T) {
	dir := t.TempDir()
	vendor := filepath.Join(dir, "wizard", "config", "app-gateway-vendor")
	populateVendor(t, vendor)
	if err := ValidateInstallerBundle(dir); err == nil {
		t.Fatal("expected error when app-gateway chart missing")
	}
}
