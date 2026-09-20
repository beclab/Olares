package chart

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeMinimalChart lays down the smallest tree loader.LoadDir accepts.
func writeMinimalChart(t *testing.T, dir, name, version string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := "apiVersion: v2\nname: " + name + "\nversion: " + version + "\nappVersion: \"" + version + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "values.yaml"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestPackageRefusesToOverwrite pins the fix for a silent data loss: the
// archive name comes from Chart.yaml, so repackaging after an edit without
// bumping the version replaced the previous archive with no mention of it,
// and an upload run afterwards could send either build.
func TestPackageRefusesToOverwrite(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "probe")
	out := filepath.Join(root, "dist")
	writeMinimalChart(t, src, "probe", "0.0.1")

	first, err := Package(src, out, false)
	if err != nil {
		t.Fatalf("first package: %v", err)
	}
	info, err := os.Stat(first)
	if err != nil {
		t.Fatalf("stat first archive: %v", err)
	}

	_, err = Package(src, out, false)
	if !errors.Is(err, ErrArchiveExists) {
		t.Fatalf("second package error = %v, want ErrArchiveExists", err)
	}
	if after, serr := os.Stat(first); serr != nil {
		t.Fatalf("archive disappeared after a refused package: %v", serr)
	} else if !after.ModTime().Equal(info.ModTime()) {
		t.Error("refused package must leave the existing archive untouched")
	}

	if _, err := Package(src, out, true); err != nil {
		t.Fatalf("package with overwrite: %v", err)
	}
}

// TestPackageOutputDirIsolatesVersions guards the escape hatch the error
// message suggests: a different -o directory is not blocked.
func TestPackageOutputDirIsolatesVersions(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "probe")
	writeMinimalChart(t, src, "probe", "0.0.1")

	if _, err := Package(src, filepath.Join(root, "a"), false); err != nil {
		t.Fatalf("package into a: %v", err)
	}
	if _, err := Package(src, filepath.Join(root, "b"), false); err != nil {
		t.Fatalf("package into b: %v", err)
	}
}
