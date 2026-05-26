package appcfg

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ApplyManifestGatewayOptionsFromChart copies options.gatewayRouteMode and
// options.inCluster from OlaresManifest.yaml when present.
func ApplyManifestGatewayOptionsFromChart(cfg *ApplicationConfig, chartPath string) {
	if cfg == nil || chartPath == "" {
		return
	}
	for _, name := range []string{"OlaresManifest.yaml", "OlaresManifest.yml"} {
		data, err := os.ReadFile(filepath.Join(chartPath, name))
		if err != nil {
			continue
		}
		var doc struct {
			Options struct {
				GatewayRouteMode string `yaml:"gatewayRouteMode" json:"gatewayRouteMode"`
				InCluster        string `yaml:"inCluster" json:"inCluster"`
			} `yaml:"options" json:"options"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			continue
		}
		if v := strings.TrimSpace(doc.Options.GatewayRouteMode); v != "" && cfg.GatewayRouteMode == "" {
			cfg.GatewayRouteMode = v
		}
		if v := strings.TrimSpace(doc.Options.InCluster); v != "" && cfg.InClusterMode == "" {
			cfg.InClusterMode = v
		}
		return
	}
}
