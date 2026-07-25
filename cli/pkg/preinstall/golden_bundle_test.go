package preinstall

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestMaterializeMatchesGoldenRuntimeBundle(t *testing.T) {
	source := filepath.Join("testdata", "materialized-bundle", "source")
	golden := filepath.Join("testdata", "materialized-bundle", "golden")
	baseDir := canonicalTempDir(t)

	err := Materialize(source, baseDir, ProfileSelections{
		HardwareProfile: "nvidia-cuda",
		DetectedGPUType: "nvidia",
	})
	if err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	actual := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(actual) })
	assertGoldenTree(t, actual, golden)

	contract, err := LoadDirectory(actual)
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
	if len(contract.Profile.Apps) != 1 {
		t.Fatalf("profile apps = %#v", contract.Profile.Apps)
	}
	app := contract.Profile.Apps[0]
	if app.Envs["HF_HUB_OFFLINE"] != "1" {
		t.Fatalf("profile HF_HUB_OFFLINE = %q", app.Envs["HF_HUB_OFFLINE"])
	}
	if app.SelectedGPUType != "nvidia" {
		t.Fatalf("profile selectedGpuType = %q", app.SelectedGPUType)
	}
}

func assertGoldenTree(t *testing.T, actual, golden string) {
	t.Helper()
	goldenEntries := treeEntries(t, golden)
	actualEntries := treeEntries(t, actual)
	if len(actualEntries) != len(goldenEntries) {
		t.Fatalf("tree entries = %v, want %v", actualEntries, goldenEntries)
	}
	for i := range goldenEntries {
		if actualEntries[i] != goldenEntries[i] {
			t.Fatalf("tree entries = %v, want %v", actualEntries, goldenEntries)
		}
		relative := goldenEntries[i]
		actualPath := filepath.Join(actual, filepath.FromSlash(relative))
		goldenPath := filepath.Join(golden, filepath.FromSlash(relative))
		actualInfo, err := os.Stat(actualPath)
		if err != nil {
			t.Fatal(err)
		}
		if relative == "." {
			assertMode(t, actualPath, expectedRootMode())
		} else if actualInfo.IsDir() {
			assertMode(t, actualPath, 0o555)
		} else {
			assertMode(t, actualPath, 0o444)
			actualData, err := os.ReadFile(actualPath)
			if err != nil {
				t.Fatal(err)
			}
			goldenData, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actualData, goldenData) {
				t.Fatalf("%s bytes differ from golden", relative)
			}
		}
	}
}

func treeEntries(t *testing.T, root string) []string {
	t.Helper()
	var entries []string
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		entries = append(entries, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(entries)
	return entries
}
