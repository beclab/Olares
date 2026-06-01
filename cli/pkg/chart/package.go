package chart

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Package loads the Helm chart at chartDir and writes a <name>-<version>.tgz
// into outputDir (defaults to the current directory), returning the path to
// the created archive. It mirrors `helm package`, so the result is accepted
// as-is by both `olares-cli chart lint` and `olares-cli market upload`. The
// archive name and version come from the chart's Chart.yaml. All non-standard
// files (notably OlaresManifest.yaml) are preserved because the helm loader
// captures them in the chart's raw file set.
func Package(chartDir, outputDir string) (string, error) {
	if chartDir == "" {
		return "", fmt.Errorf("chart directory is required")
	}
	if outputDir == "" {
		outputDir = "."
	}
	c, err := loader.LoadDir(chartDir)
	if err != nil {
		return "", fmt.Errorf("load chart %q: %w", chartDir, err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}
	out, err := chartutil.Save(c, outputDir)
	if err != nil {
		return "", fmt.Errorf("package chart: %w", err)
	}
	return out, nil
}
