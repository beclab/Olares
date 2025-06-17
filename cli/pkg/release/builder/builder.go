package builder

import (
	"fmt"
	"github.com/beclab/Olares/cli/pkg/core/util"
	"github.com/beclab/Olares/cli/pkg/release/app"
	"github.com/beclab/Olares/cli/pkg/release/manifest"
	"os"
	"path/filepath"
)

type Builder struct {
	olaresRepoRoot  string
	distPath        string
	version         string
	manifestManager *manifest.Manager
	appManager      *app.Manager
}

func NewBuilder(olaresRepoRoot, version, cdnURL string, ignoreMissingImages bool) *Builder {
	distPath := filepath.Join(olaresRepoRoot, ".dist/install-wizard")
	return &Builder{
		olaresRepoRoot:  olaresRepoRoot,
		distPath:        distPath,
		version:         version,
		manifestManager: manifest.NewManager(olaresRepoRoot, distPath, cdnURL, ignoreMissingImages),
		appManager:      app.NewManager(olaresRepoRoot, distPath),
	}
}

func (b *Builder) Build() (string, error) {
	// Clean previous build
	if err := os.RemoveAll(b.distPath); err != nil {
		return "", fmt.Errorf("failed to clean previous dist directory: %v", err)
	}

	// Create dist directory if not exists
	if err := os.MkdirAll(b.distPath, 0755); err != nil {
		return "", err
	}

	// Package apps
	if err := b.appManager.Package(); err != nil {
		return "", fmt.Errorf("package apps failed: %v", err)
	}

	// Generate manifest
	if err := b.manifestManager.Generate(); err != nil {
		return "", fmt.Errorf("failed to generate manifest: %v", err)
	}

	// archive the install-wizard
	return b.archive()

}

func (b *Builder) archive() (string, error) {
	versionStr := "v" + b.version
	files := []string{
		filepath.Join(b.distPath, "wizard/config/settings/templates/terminus_cr.yaml"),
		filepath.Join(b.distPath, "install.sh"),
		filepath.Join(b.distPath, "install.ps1"),
		filepath.Join(b.distPath, "joincluster.sh"),
		filepath.Join(b.distPath, "installation.manifest"),
	}

	for _, file := range files {
		if err := util.ReplaceInFile(file, "#__VERSION__", b.version); err != nil {
			return "", err
		}
	}

	tarFile := filepath.Join(b.olaresRepoRoot, fmt.Sprintf("install-wizard-%s.tar.gz", versionStr))
	return tarFile, util.Tar(b.distPath, tarFile, b.distPath)
}
