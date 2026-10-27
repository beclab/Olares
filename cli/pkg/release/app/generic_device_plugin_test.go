package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageGenericDevicePlugin(t *testing.T) {
	repoRoot := t.TempDir()
	distPath := t.TempDir()
	source := filepath.Join(repoRoot, "infrastructure", "generic-device-plugin", ".olares", "generic-device-plugin.yaml")
	if err := os.MkdirAll(filepath.Dir(source), 0755); err != nil {
		t.Fatal(err)
	}
	const manifest = "apiVersion: apps/v1\nkind: DaemonSet\n"
	if err := os.WriteFile(source, []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	m := NewManager(repoRoot, distPath)
	if err := m.packageGenericDevicePlugin(); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(distPath, "wizard", "config", "generic-device-plugin", "generic-device-plugin.yaml")
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != manifest {
		t.Fatalf("packaged manifest = %q, want %q", got, manifest)
	}
}
