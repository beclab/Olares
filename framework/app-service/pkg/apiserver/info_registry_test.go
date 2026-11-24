package apiserver

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeManifest(t *testing.T, dir, name, version string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := "olaresManifest.version: '0.10.0'\nolaresManifest.type: app\nmetadata:\n  name: " + name +
		"\n  title: " + name + "\n  version: '" + version + "'\n"
	if err := os.WriteFile(filepath.Join(dir, "OlaresManifest.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryAppsReadsBothLayouts(t *testing.T) {
	root := t.TempDir()
	// legacy flat layout from an earlier release
	writeManifest(t, filepath.Join(root, "legacy"), "legacy", "1.0.0")
	// per-version cache layout, two versions of one app; the newer write wins
	oldDir := filepath.Join(root, "market.olares", "alice", "notes", "1.0.0-aaaaaaaaaaaa", "notes")
	newDir := filepath.Join(root, "market.olares", "alice", "notes", "2.0.0-bbbbbbbbbbbb", "notes")
	writeManifest(t, oldDir, "notes", "1.0.0")
	writeManifest(t, newDir, "notes", "2.0.0")
	past := time.Now().Add(-time.Hour)
	_ = os.Chtimes(filepath.Join(oldDir, "OlaresManifest.yaml"), past, past)
	// a sub-chart inside a chart root is not an app of its own
	writeManifest(t, filepath.Join(newDir, "server"), "server", "2.0.0")
	// temporary unpacks are ignored
	writeManifest(t, filepath.Join(root, ".tmp", "chart-123", "half"), "half", "0.0.1")

	apps, err := registryApps(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 2 {
		t.Fatalf("found %d apps, want 2: %v", len(apps), keys(apps))
	}
	if apps["legacy"] == nil {
		t.Fatal("legacy flat chart not listed")
	}
	if got := apps["notes"].Metadata.Version; got != "2.0.0" {
		t.Fatalf("notes version = %s, want the newest 2.0.0", got)
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
