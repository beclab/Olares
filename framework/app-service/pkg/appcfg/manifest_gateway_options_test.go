package appcfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyManifestGatewayOptionsFromChart(t *testing.T) {
	dir := t.TempDir()
	manifest := `apiVersion: v2
options:
  gatewayRouteMode: gateway
  inCluster: gateway
`
	if err := os.WriteFile(filepath.Join(dir, "OlaresManifest.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &ApplicationConfig{}
	ApplyManifestGatewayOptionsFromChart(cfg, dir)
	if cfg.GatewayRouteMode != "gateway" || cfg.InClusterMode != "gateway" {
		t.Fatalf("got route=%q inCluster=%q", cfg.GatewayRouteMode, cfg.InClusterMode)
	}
}
