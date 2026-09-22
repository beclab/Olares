package chart

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

// ErrArchiveExists is returned by Package when the target .tgz is already
// present and overwrite is false. Callers match on it to suggest --force.
var ErrArchiveExists = errors.New("archive already exists")

// Package loads the Helm chart at chartDir and writes a <name>-<version>.tgz
// into outputDir (defaults to the current directory), returning the path to
// the created archive. It mirrors `helm package`, so the result is accepted
// as-is by both `olares-cli chart lint` and `olares-cli market upload`. The
// archive name and version come from the chart's Chart.yaml. All non-standard
// files (notably OlaresManifest.yaml) are preserved because the helm loader
// captures them in the chart's raw file set.
//
// An existing archive at the target path is left alone unless overwrite is
// set: because the name is derived from Chart.yaml rather than from the
// output flag, repackaging after an edit without bumping the version
// otherwise replaces the previous archive with no mention of it, and an
// upload that follows can send either one depending on which step the user
// re-ran.
func Package(chartDir, outputDir string, overwrite bool) (string, error) {
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
	target := filepath.Join(outputDir, fmt.Sprintf("%s-%s.tgz", c.Name(), c.Metadata.Version))
	if !overwrite {
		if _, err := os.Stat(target); err == nil {
			return "", fmt.Errorf("%w: %s", ErrArchiveExists, target)
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("stat %s: %w", target, err)
		}
	}
	out, err := chartutil.Save(c, outputDir)
	if err != nil {
		return "", fmt.Errorf("package chart: %w", err)
	}
	return out, nil
}
