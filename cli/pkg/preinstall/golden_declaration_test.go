package preinstall

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// The published tree is what Market reads, so it is asserted whole and
// byte-for-byte rather than field by field: a change to the declaration's shape
// has to be visible in a diff of this fixture before it reaches a device.
func TestPublishMatchesGoldenRuntimeDeclaration(t *testing.T) {
	source := filepath.Join("testdata", "materialized-declaration", "source")
	golden := filepath.Join("testdata", "materialized-declaration", "golden")
	baseDir := canonicalTempDir(t)
	freezeGeneratedAt(t)

	err := Publish(source, baseDir, testOSVersion, ProfileSelections{
		HardwareProfile: "nvidia-cuda",
		DetectedGPUType: "nvidia",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	actual := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(actual) })
	assertGoldenTree(t, actual, golden)

	declaration := readPublishedDeclaration(t, actual, testOSVersion)
	app := declarationApp(t, declaration, "47ca2f42")
	if app.ChartSource != ChartSourceLocal || app.InstallScope != InstallScopeShared {
		t.Fatalf("bundled app = %#v", app)
	}
	if app.Envs["HF_HUB_OFFLINE"] != "1" {
		t.Fatalf("HF_HUB_OFFLINE = %q", app.Envs["HF_HUB_OFFLINE"])
	}
	if app.SelectedGPUType != "nvidia" {
		t.Fatalf("selectedGpuType = %q", app.SelectedGPUType)
	}
	// The catalog apps this version expects are declared alongside whatever the
	// medium carried, which is the whole point of one file: Market has one list
	// to read rather than two to reconcile.
	router := declarationApp(t, declaration, catalogApps[0].AppID)
	if router.ChartSource != ChartSourceCatalog || router.Version != "" {
		t.Fatalf("catalog app = %#v", router)
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
		if actualInfo.IsDir() {
			// The directory stays writable because the next release adds a file
			// beside these ones; the pod's mount is read-only and every file is
			// sealed.
			assertMode(t, actualPath, 0o755)
			continue
		}
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
			t.Fatalf("%s bytes differ from golden:\ngot:\n%s\nwant:\n%s", relative, actualData, goldenData)
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
