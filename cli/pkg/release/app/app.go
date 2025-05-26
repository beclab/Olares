package app

import (
	"fmt"
	"os"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/core/util"
)

type Manager struct {
	olaresRepoRoot string
	distPath       string
}

func NewManager(olaresRepoRoot, distPath string) *Manager {
	return &Manager{
		olaresRepoRoot: olaresRepoRoot,
		distPath:       distPath,
	}
}

func (m *Manager) Package() error {
	modules := []string{"frameworks", "libs", "apps", "third-party"}
	buildTemplate := "build/installer"

	// Create dist directory if not exists
	if err := os.MkdirAll(m.distPath, 0755); err != nil {
		return err
	}

	// Copy template files
	if err := util.CopyDirectory(buildTemplate, m.distPath); err != nil {
		return err
	}

	// Package modules
	for _, mod := range modules {
		if err := m.packageModule(mod); err != nil {
			return err
		}
	}

	// Package launcher and GPU
	if err := m.packageLauncher(); err != nil {
		return err
	}

	if err := m.packageGPU(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) packageModule(mod string) error {
	modPath := filepath.Join(m.olaresRepoRoot, mod)
	entries, err := os.ReadDir(modPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		app := entry.Name()

		fmt.Printf("packaging %s ... \n", app)

		// Package user app charts
		chartPath := filepath.Join(modPath, app, "config/user/helm-charts")
		if err := util.CopyDirectoryIfExists(chartPath, filepath.Join(m.distPath, "wizard/config/apps")); err != nil {
			return err
		}

		// Package cluster CRDs
		crdPath := filepath.Join(modPath, app, "config/cluster/crds")
		if err := util.CopyDirectoryIfExists(crdPath, filepath.Join(m.distPath, "wizard/config/settings/templates/crds")); err != nil {
			return err
		}

		// Package cluster deployments
		deployPath := filepath.Join(modPath, app, "config/cluster/deploy")
		if err := util.CopyDirectoryIfExists(deployPath, filepath.Join(m.distPath, "wizard/config/system/templates/deploy")); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) packageLauncher() error {
	fmt.Println("packaging launcher ...")
	return util.CopyDirectory(
		filepath.Join(m.olaresRepoRoot, "frameworks/bfl/config/launcher"),
		filepath.Join(m.distPath, "wizard/config/launcher"),
	)
}

func (m *Manager) packageGPU() error {
	fmt.Println("packaging gpu ...")
	return util.CopyDirectory(
		filepath.Join(m.olaresRepoRoot, "frameworks/GPU/config/gpu"),
		filepath.Join(m.distPath, "wizard/config/gpu"),
	)
}
