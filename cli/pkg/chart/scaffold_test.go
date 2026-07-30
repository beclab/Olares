package chart

import (
	"os"
	"path/filepath"
	"testing"

	oac "github.com/beclab/Olares/framework/oac"
	"github.com/beclab/api/manifest"
	"sigs.k8s.io/yaml"
)

func TestFromComposeProducesV3LintableChart(t *testing.T) {
	root := t.TempDir()
	composePath := filepath.Join(root, "compose.yaml")
	compose := []byte(`services:
  web:
    image: beclab/hello:1.0.0
    labels:
      olares.service.type: Entrance
    ports:
      - "8080:80"
`)
	if err := os.WriteFile(composePath, compose, 0o600); err != nil {
		t.Fatal(err)
	}

	chartDir := filepath.Join(root, "testapp")
	if err := FromCompose(Options{
		ComposeFiles: []string{composePath},
		OutputDir:    chartDir,
		Name:         "testapp",
		Title:        "Test App",
	}); err != nil {
		t.Fatalf("FromCompose() error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(chartDir, appCfgFileName))
	if err != nil {
		t.Fatal(err)
	}
	var cfg manifest.AppConfiguration
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}

	if cfg.APIVersion != appAPIVersion {
		t.Fatalf("apiVersion = %q, want %q", cfg.APIVersion, appAPIVersion)
	}
	if cfg.ConfigVersion != configVersion {
		t.Fatalf("olaresManifest.version = %q, want %q", cfg.ConfigVersion, configVersion)
	}
	if cfg.WorkloadReplicas == nil || (*cfg.WorkloadReplicas)["testapp"] != 1 {
		t.Fatalf("workloadReplicas = %#v, want testapp: 1", cfg.WorkloadReplicas)
	}

	var systemDependency *manifest.Dependency
	for i := range cfg.Options.Dependencies {
		dependency := &cfg.Options.Dependencies[i]
		if dependency.Name == olaresSystemDepName && dependency.Type == "system" {
			systemDependency = dependency
			break
		}
	}
	if systemDependency == nil {
		t.Fatal("missing olares system dependency")
	}
	if systemDependency.Version != olaresSystemDepVersion {
		t.Fatalf("olares dependency version = %q, want %q", systemDependency.Version, olaresSystemDepVersion)
	}

	if err := oac.Lint(chartDir, oac.WithAutoOwnerScenarios()); err != nil {
		t.Fatalf("freshly scaffolded chart must pass lint: %v", err)
	}
}
