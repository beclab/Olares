package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/cli/pkg/core/util"
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
	modules := []string{"apps", "framework", "daemon", "infrastructure", "platform", "vendor"}
	buildTemplate := "build/base-package"

	// Copy template files
	if err := util.CopyDirectory(buildTemplate, m.distPath); err != nil {
		return err
	}

	osChartTemplatePath := "wizard/config/os-chart-template"
	for _, osm := range []string{"os-platform", "os-framework"} {
		if err := util.CopyDirectory(filepath.Join(buildTemplate, osChartTemplatePath), filepath.Join(m.distPath, fmt.Sprintf("/wizard/config/%s", osm))); err != nil {
			return err
		}
	}

	if err := util.RemoveDir(filepath.Join(m.distPath, osChartTemplatePath)); err != nil {
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
	var distDeployType string
	switch mod {
	case "platform":
		distDeployType = "os-platform"
	case "framework":
		distDeployType = "os-framework"
	}
	modPath := filepath.Join(m.olaresRepoRoot, mod)
	err := filepath.Walk(modPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !strings.EqualFold(info.Name(), ".olares") {
			return nil
		}

		fmt.Printf("packaging %s ... \n", path)

		// Package user app charts
		chartPath := filepath.Join(path, "config/user/helm-charts")
		if err := util.CopyDirectoryIfExists(chartPath, filepath.Join(m.distPath, "wizard/config/apps")); err != nil {
			return err
		}

		// Package cluster CRDs
		crdPath := filepath.Join(path, "config/cluster/crds")
		if err := util.CopyDirectoryIfExists(crdPath, filepath.Join(m.distPath, "wizard/config/settings/templates/crds")); err != nil {
			return err
		}

		// Package cluster deployments
		deployPath := filepath.Join(path, "config/cluster/deploy")
		if err := util.CopyDirectoryIfExists(deployPath, filepath.Join(m.distPath, fmt.Sprintf("wizard/config/%s/templates/deploy", distDeployType))); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (m *Manager) packageLauncher() error {
	fmt.Println("packaging launcher ...")
	return util.CopyDirectory(
		filepath.Join(m.olaresRepoRoot, "framework/bfl/.olares/config/launcher"),
		filepath.Join(m.distPath, "wizard/config/launcher"),
	)
}

func (m *Manager) packageGPU() error {
	fmt.Println("packaging gpu ...")
	return util.CopyDirectory(
		filepath.Join(m.olaresRepoRoot, "infrastructure/gpu/.olares/config/gpu"),
		filepath.Join(m.distPath, "wizard/config/gpu"),
	)
}
