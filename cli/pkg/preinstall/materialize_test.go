package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testOSVersion    = "1.12.7"
	testGeneratedAt  = "2026-07-25T00:00:00Z"
	testNextOSTrunk  = "1.13.0"
	testBundledAppID = "app-a"
)

// A medium with no bundle of its own still has something to say: every device
// running this version is expected to have the catalog apps, and a declaration
// is the only place that is written down.
func TestMaterializeWithoutAStaticBundlePublishesTheCatalogApps(t *testing.T) {
	root := t.TempDir()
	baseDir := filepath.Join(root, "base")

	if err := Materialize(filepath.Join(root, "installer"), baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	declaration := readPublishedDeclaration(t, target, testOSVersion)
	if len(declaration.Apps) != len(catalogApps) {
		t.Fatalf("declared apps = %#v, want the catalog apps alone", declaration.Apps)
	}
	if _, err := os.Stat(filepath.Join(target, "chart")); !os.IsNotExist(err) {
		t.Fatalf("a declaration with no local entry staged a chart directory: %v", err)
	}
}

// One file per trunk, so publishing one release's declaration must leave every
// other release's alone -- including a device that upgraded and now has two.
func TestMaterializeAddsThisReleaseBesideTheOnesAlreadyThere(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	previous := filepath.Join(target, DeclarationFileName(testNextOSTrunk))
	if err := os.WriteFile(previous, []byte("another release"), 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeWritable(target) })

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	if got, err := os.ReadFile(previous); err != nil || string(got) != "another release" {
		t.Fatalf("another release's declaration changed: content=%q error=%v", got, err)
	}
	readPublishedDeclaration(t, target, testOSVersion)
}

// Every build of a release shares one declaration, and a device that has already
// acted on it must not be handed a second answer.
func TestMaterializeLeavesThisTrunksDeclarationAloneOnceItExists(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("first Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	published := filepath.Join(target, DeclarationFileName(testOSVersion))
	first, err := os.ReadFile(published)
	if err != nil {
		t.Fatal(err)
	}

	err = Materialize(installerDir, baseDir, testOSVersion+"-rc.2", ProfileSelections{
		Apps: map[string]AppSelection{testBundledAppID: {Envs: map[string]string{"WORKER_COUNT": "9"}}},
	})

	if err != nil {
		t.Fatalf("second Materialize() error = %v", err)
	}
	second, err := os.ReadFile(published)
	if err != nil {
		t.Fatal(err)
	}
	if string(second) != string(first) {
		t.Fatalf("declaration was rewritten:\ngot:\n%s\nwant:\n%s", second, first)
	}
}

func TestMaterializeRejectsOversizedBundle(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	bundlePath := filepath.Join(installerDir, StaticRelativeDir, BundleFileName)
	if err := os.Truncate(bundlePath, MaxBundleJSONBytes+1); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err == nil ||
		!strings.Contains(err.Error(), "bundle.json exceeds") {
		t.Fatalf("Materialize() error = %v", err)
	}
}

func TestMaterializeRejectsOversizedChartTotal(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps = nil
	const chartCount = 5
	chartSize := int64(MaxTotalChartBytes/chartCount + 1)
	for i := 0; i < chartCount; i++ {
		name := fmt.Sprintf("charts/app-%d.tgz", i)
		bundle.Apps = append(bundle.Apps, BundleAppV1{
			AppID:        fmt.Sprintf("app-%d", i),
			AppName:      fmt.Sprintf("app-%d", i),
			Version:      "1.0.0",
			InstallScope: InstallScopeShared,
			Chart:        name,
			ChartSHA256:  strings.Repeat("0", 64),
			AppEntry:     json.RawMessage(`{}`),
		})
		chart := filepath.Join(staticDir, filepath.FromSlash(name))
		if err := os.WriteFile(chart, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Truncate(chart, chartSize); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, BundleFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err == nil ||
		!strings.Contains(err.Error(), "total chart size exceeds") {
		t.Fatalf("Materialize() error = %v", err)
	}
}

func TestMaterializeFoldsSelectionsIntoTheDeclaration(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	selections := ProfileSelections{
		HardwareProfile: "nvidia-cuda",
		Apps: map[string]AppSelection{
			testBundledAppID: {
				SelectedGPUType: "nvidia",
				Envs:            map[string]string{"WORKER_COUNT": "2"},
			},
		},
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, selections); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	app := declarationApp(t, readPublishedDeclaration(t, target, testOSVersion), testBundledAppID)
	if app.SelectedGPUType != "nvidia" || app.Envs["WORKER_COUNT"] != "2" {
		t.Fatalf("declared app = %#v", app)
	}
	assertMode(t, target, 0o755)
	assertMode(t, filepath.Join(target, "charts"), 0o755)
	assertMode(t, filepath.Join(target, DeclarationFileName(testOSVersion)), 0o444)
	assertMode(t, filepath.Join(target, "charts", "app-a-1.0.0.tgz"), 0o444)
}

func TestMaterializePublishesArtifactManifestWithoutPayload(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	artifact, manifestData := addMaterializeArtifactFixture(t, installerDir)

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	readPublishedDeclaration(t, target, testOSVersion)
	published, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(artifact.Manifest)))
	if err != nil {
		t.Fatal(err)
	}
	if string(published) != string(manifestData) {
		t.Fatalf("published manifest bytes changed")
	}
	sum := sha256.Sum256(published)
	if hex.EncodeToString(sum[:]) != artifact.ManifestSHA256 {
		t.Fatalf("published manifest digest = %x, want %s", sum, artifact.ManifestSHA256)
	}
	if _, err := os.Lstat(filepath.Join(target, filepath.FromSlash(artifact.Source))); !os.IsNotExist(err) {
		t.Fatalf("artifact payload was published: %v", err)
	}
	assertMode(t, filepath.Join(target, "manifests"), 0o755)
	assertMode(t, filepath.Join(target, filepath.FromSlash(artifact.Manifest)), 0o444)
}

func TestMaterializeRejectsArtifactManifestSymlinks(t *testing.T) {
	tests := []struct {
		name    string
		replace func(*testing.T, string, BundleArtifactV1)
	}{
		{
			name: "leaf",
			replace: func(t *testing.T, staticDir string, artifact BundleArtifactV1) {
				manifestPath := filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest))
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					t.Fatal(err)
				}
				outside := filepath.Join(t.TempDir(), "manifest.json")
				if err := os.WriteFile(outside, data, 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(manifestPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, manifestPath); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "ancestor",
			replace: func(t *testing.T, staticDir string, _ BundleArtifactV1) {
				manifests := filepath.Join(staticDir, "manifests")
				outside := filepath.Join(t.TempDir(), "manifests")
				if err := os.Rename(manifests, outside); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, manifests); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installerDir, baseDir := writeStaticBundle(t)
			artifact, _ := addMaterializeArtifactFixture(t, installerDir)
			tt.replace(t, filepath.Join(installerDir, StaticRelativeDir), artifact)

			err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{})

			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("Materialize() error = %v, want symlink rejection", err)
			}
		})
	}
}

// A publish that fails writes no declaration, so whatever the device was already
// told stays the answer.
func TestMaterializeRejectsSymlinkWithoutPublishingADeclaration(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	if err := os.Remove(filepath.Join(staticDir, "charts", "app-a-1.0.0.tgz")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside.tgz"), filepath.Join(staticDir, "charts", "app-a-1.0.0.tgz")); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(target, "marker")
	if err := os.WriteFile(marker, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want symlink rejection")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "existing" {
		t.Fatalf("existing target changed: content=%q error=%v", got, err)
	}
	if _, err := os.Lstat(filepath.Join(target, DeclarationFileName(testOSVersion))); !os.IsNotExist(err) {
		t.Fatalf("a failed publish left a declaration: %v", err)
	}
}

func TestMaterializeRejectsSymlinkedStaticInputParent(t *testing.T) {
	root := t.TempDir()
	realInstaller, baseDir := writeStaticBundle(t)
	installerDir := filepath.Join(root, "installer")
	if err := os.MkdirAll(installerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(realInstaller, "preinstall"), filepath.Join(installerDir, "preinstall")); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want symlinked parent rejection")
	}
}

func TestMaterializeRejectsSymlinkedInstallerAndBasePaths(t *testing.T) {
	t.Run("installer", func(t *testing.T) {
		realInstaller, baseDir := writeStaticBundle(t)
		link := filepath.Join(t.TempDir(), "installer")
		if err := os.Symlink(realInstaller, link); err != nil {
			t.Fatal(err)
		}
		if err := Materialize(link, baseDir, testOSVersion, ProfileSelections{}); err == nil {
			t.Fatal("Materialize() error = nil, want installer symlink rejection")
		}
	})

	t.Run("base", func(t *testing.T) {
		installerDir, realBase := writeStaticBundle(t)
		link := filepath.Join(t.TempDir(), "base")
		if err := os.Symlink(realBase, link); err != nil {
			t.Fatal(err)
		}
		if err := Materialize(installerDir, link, testOSVersion, ProfileSelections{}); err == nil {
			t.Fatal("Materialize() error = nil, want base symlink rejection")
		}
	})
}

func TestMaterializeAllowsSymlinkedAncestorDirectories(t *testing.T) {
	t.Run("installer ancestor", func(t *testing.T) {
		realInstaller, baseDir := writeStaticBundle(t)
		root := canonicalTempDir(t)
		link := filepath.Join(root, "linked-parent")
		if err := os.Symlink(filepath.Dir(realInstaller), link); err != nil {
			t.Fatal(err)
		}
		installerDir := filepath.Join(link, filepath.Base(realInstaller))
		t.Cleanup(func() { _ = makeWritable(baseDir) })

		if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
			t.Fatalf("Materialize() through symlinked installer ancestor error = %v", err)
		}
		readPublishedDeclaration(t, filepath.Join(baseDir, RuntimeRelativeDir), testOSVersion)
	})

	t.Run("base ancestor", func(t *testing.T) {
		installerDir, _ := writeStaticBundle(t)
		realParent := canonicalTempDir(t)
		root := canonicalTempDir(t)
		link := filepath.Join(root, "linked-parent")
		if err := os.Symlink(realParent, link); err != nil {
			t.Fatal(err)
		}
		baseDir := filepath.Join(link, "base")
		t.Cleanup(func() { _ = makeWritable(realParent) })

		if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
			t.Fatalf("Materialize() through symlinked base ancestor error = %v", err)
		}
		readPublishedDeclaration(t, filepath.Join(realParent, "base", RuntimeRelativeDir), testOSVersion)
	})
}

func TestMaterializeRejectsSymlinkedRuntimeParent(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(baseDir, "userdata")); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want runtime parent symlink rejection")
	}
}

func TestMaterializeRefusesAPublishWithNoVersionToNameItAfter(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)

	err := Materialize(installerDir, baseDir, "  ", ProfileSelections{})

	if err == nil || !strings.Contains(err.Error(), "osVersion") {
		t.Fatalf("Materialize() error = %v, want a missing version rejection", err)
	}
}

func TestCleanupStagingRootsRemovesSealedStaging(t *testing.T) {
	parent := testRoot(t)
	name := stagingPrefix + "0123456789abcdef"
	if err := parent.Mkdir(name, 0o700); err != nil {
		t.Fatal(err)
	}
	stage, err := parent.OpenRoot(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := stage.WriteFile("payload", []byte("sealed"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := stage.Chmod(".", 0o555); err != nil {
		t.Fatal(err)
	}
	if err := stage.Close(); err != nil {
		t.Fatal(err)
	}

	if err := cleanupStagingRoots(parent); err != nil {
		t.Fatalf("cleanupStagingRoots() error = %v", err)
	}
	if _, err := parent.Lstat(name); !os.IsNotExist(err) {
		t.Fatalf("sealed staging still exists: %v", err)
	}
}

// Every interrupted run leaves a staging directory holding a copy of the
// payload. Nothing else writes that name, so the next run clears them; a
// directory that only looks similar is left where it is.
func TestMaterializeRemovesStagingLeftBehindByAnInterruptedRun(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(target, stagingPrefix+strings.Repeat("a", 16))
	unrelated := filepath.Join(target, stagingPrefix+"not-a-staging-token")
	for _, dir := range []string{stale, unrelated} {
		if err := os.MkdirAll(filepath.Join(dir, "charts"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	t.Cleanup(func() { _ = makeWritable(target) })

	if _, err := os.Lstat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale staging survived: %v", err)
	}
	if _, err := os.Lstat(unrelated); err != nil {
		t.Fatalf("directory outside the staging naming scheme was removed: %v", err)
	}
}

func TestMaterializeSupportsNestedChartPaths(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	bundlePath := filepath.Join(staticDir, BundleFileName)
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	var bundle BundleV1
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	oldChart := filepath.Join(staticDir, filepath.FromSlash(bundle.Apps[0].Chart))
	bundle.Apps[0].Chart = "charts/nested/app-a-1.0.0.tgz"
	newChart := filepath.Join(staticDir, filepath.FromSlash(bundle.Apps[0].Chart))
	if err := os.MkdirAll(filepath.Dir(newChart), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldChart, newChart); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundlePath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, testOSVersion, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	assertMode(t, filepath.Join(target, filepath.FromSlash(bundle.Apps[0].Chart)), 0o444)
	assertMode(t, filepath.Join(target, "charts", "nested"), 0o755)
}

// An upgrade declares the catalog apps for the release it brings, and carries no
// payload at all: the charts are on the network.
func TestPublishCatalogDeclarationDeclaresTheCatalogAppsAlone(t *testing.T) {
	baseDir := canonicalTempDir(t)

	if err := PublishCatalogDeclaration(baseDir, testNextOSTrunk+"-20260731"); err != nil {
		t.Fatalf("PublishCatalogDeclaration() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	declaration := readPublishedDeclaration(t, target, testNextOSTrunk)
	if len(declaration.Apps) != len(catalogApps) {
		t.Fatalf("declared apps = %#v", declaration.Apps)
	}
	for _, app := range declaration.Apps {
		if app.ChartSource != ChartSourceCatalog {
			t.Fatalf("declared app %q chartSource = %q", app.AppID, app.ChartSource)
		}
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != DeclarationFileName(testNextOSTrunk) {
		t.Fatalf("published entries = %#v, want the declaration alone", entries)
	}
}

func writeStaticBundle(t *testing.T) (string, string) {
	t.Helper()
	root := canonicalTempDir(t)
	installerDir := filepath.Join(root, "installer")
	baseDir := filepath.Join(root, "base")
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	chartData := []byte("fixture chart")
	digest := sha256.Sum256(chartData)
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].ChartSHA256 = hex.EncodeToString(digest[:])
	bundleData, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(staticDir, "charts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, BundleFileName), bundleData, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "charts", "app-a-1.0.0.tgz"), chartData, 0o644); err != nil {
		t.Fatal(err)
	}
	return installerDir, baseDir
}

func readPublishedDeclaration(t *testing.T, target, osVersion string) DeclarationV2 {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(target, DeclarationFileName(osVersion)))
	if err != nil {
		t.Fatal(err)
	}
	var declaration DeclarationV2
	if err := strictDecode(data, &declaration); err != nil {
		t.Fatalf("decode published declaration: %v", err)
	}
	if err := ValidateDeclaration(&declaration); err != nil {
		t.Fatalf("published declaration is not one Market would read: %v", err)
	}
	if TrunkVersion(declaration.OSVersion) != TrunkVersion(osVersion) {
		t.Fatalf("declaration osVersion = %q, want trunk %q", declaration.OSVersion, TrunkVersion(osVersion))
	}
	return declaration
}

func declarationApp(t *testing.T, declaration DeclarationV2, appID string) DeclarationAppV2 {
	t.Helper()
	for _, app := range declaration.Apps {
		if app.AppID == appID {
			return app
		}
	}
	t.Fatalf("declaration has no app %q: %#v", appID, declaration.Apps)
	return DeclarationAppV2{}
}

func freezeGeneratedAt(t *testing.T) {
	t.Helper()
	original := generatedNow
	generatedNow = func() string { return testGeneratedAt }
	t.Cleanup(func() { generatedNow = original })
}

func mapsEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func addMaterializeArtifactFixture(t *testing.T, installerDir string) (BundleArtifactV1, []byte) {
	t.Helper()
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	const payloadSize = int64(20 << 30)
	revision := strings.Repeat("a", 40)
	manifest := ArtifactManifestV1{
		SchemaVersion: 1,
		Repo:          "acme/large-model",
		Revision:      revision,
		Entries: []ArtifactManifestEntryV1{{
			Path:   "weights.bin",
			Type:   "file",
			Size:   payloadSize,
			SHA256: strings.Repeat("0", 64),
		}},
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	manifestData := append([]byte("\n"), append(encoded, '\n')...)
	manifestDigest := sha256.Sum256(manifestData)
	artifact := BundleArtifactV1{
		Kind:           ArtifactKindHFCache,
		Source:         "artifacts/large-model",
		Repo:           manifest.Repo,
		Revision:       manifest.Revision,
		Manifest:       "manifests/large-model.json",
		ManifestSHA256: hex.EncodeToString(manifestDigest[:]),
		TotalSize:      payloadSize,
	}
	sourceDir := filepath.Join(staticDir, filepath.FromSlash(artifact.Source))
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	payload, err := os.Create(filepath.Join(sourceDir, "weights.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := payload.Truncate(payloadSize); err != nil {
		_ = payload.Close()
		t.Fatal(err)
	}
	if err := payload.Close(); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest))
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(staticDir, BundleFileName)
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Apps[0].Artifacts = []BundleArtifactV1{artifact}
	bundleData, err = json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundlePath, bundleData, 0o644); err != nil {
		t.Fatal(err)
	}
	return artifact, manifestData
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %o, want %o", path, got, want)
	}
}
