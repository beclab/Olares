package builder

import (
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/release/app"
	"bytetrade.io/web3os/installer/pkg/release/manifest"
	"fmt"
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
		manifestManager: manifest.NewManager(olaresRepoRoot, cdnURL, ignoreMissingImages),
		appManager:      app.NewManager(olaresRepoRoot, distPath),
	}
}

func (b *Builder) Build() (string, error) {
	// Clean previous build
	if err := os.RemoveAll(filepath.Join(b.olaresRepoRoot, ".dist")); err != nil {
		return "", fmt.Errorf("failed to clean previous dist directory: %v", err)
	}

	// Package apps
	if err := b.appManager.Package(); err != nil {
		return "", fmt.Errorf("package apps failed: %v", err)
	}

	// Build manifest
	if err := b.manifestManager.Build(); err != nil {
		return "", fmt.Errorf("manifest build failed: %v", err)
	}

	// Copy upgrade script, as the current build.sh does
	if err := util.CopyFile(
		filepath.Join(b.olaresRepoRoot, "scripts/upgrade.sh"),
		filepath.Join(b.distPath, "upgrade.sh"),
	); err != nil {
		return "", fmt.Errorf("failed to copy upgrade script: %v", err)
	}

	// archive the install-wizard
	return b.createFinalPackage()

}

func (b *Builder) createFinalPackage() (string, error) {
	if err := os.RemoveAll(filepath.Join(b.distPath, "images")); err != nil {
		return "", err
	}

	manifestSrc := filepath.Join(b.olaresRepoRoot, ".manifest/installation.manifest")
	manifestDst := filepath.Join(b.distPath, "installation.manifest")
	if err := util.MoveFile(manifestSrc, manifestDst); err != nil {
		return "", err
	}

	imagesSrc := filepath.Join(b.olaresRepoRoot, ".manifest")
	imagesDst := filepath.Join(b.distPath, "images")
	if err := util.MoveDirectory(imagesSrc, imagesDst); err != nil {
		return "", err
	}

	versionStr := "v" + b.version
	files := []string{
		filepath.Join(b.distPath, "wizard/config/settings/templates/terminus_cr.yaml"),
		filepath.Join(b.distPath, "install.sh"),
		filepath.Join(b.distPath, "install.ps1"),
	}

	for _, file := range files {
		if err := util.ReplaceInFile(file, "#__VERSION__", b.version); err != nil {
			return "", err
		}
	}

	tarFile := filepath.Join(b.olaresRepoRoot, fmt.Sprintf("install-wizard-%s.tar.gz", versionStr))
	return tarFile, util.Tar(b.distPath, tarFile, b.distPath)
}
