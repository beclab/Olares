package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMaterializeNoOpsWhenStaticInputIsMissing(t *testing.T) {
	root := t.TempDir()
	err := Materialize(filepath.Join(root, "installer"), filepath.Join(root, "base"), ProfileSelections{})
	if err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "base", RuntimeRelativeDir)); !os.IsNotExist(err) {
		t.Fatalf("runtime directory stat error = %v", err)
	}
}

func TestMaterializeMissingInputPreservesExistingOverlay(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "base", RuntimeRelativeDir)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(target, "managed")
	if err := os.WriteFile(marker, []byte("preserved"), 0o444); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(filepath.Join(root, "installer"), filepath.Join(root, "base"), ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "preserved" {
		t.Fatalf("existing overlay changed: content=%q error=%v", got, err)
	}
}

func TestMaterializeRejectsOversizedBundle(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	bundlePath := filepath.Join(installerDir, StaticRelativeDir, BundleFileName)
	if err := os.Truncate(bundlePath, MaxBundleJSONBytes+1); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil ||
		!strings.Contains(err.Error(), "bundle.json exceeds") {
		t.Fatalf("Materialize() error = %v", err)
	}
}

func TestMaterializeRejectsOversizedChart(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	chart := filepath.Join(installerDir, StaticRelativeDir, "charts", "app-a-1.0.0.tgz")
	if err := os.Truncate(chart, MaxChartBytes+1); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil ||
		!strings.Contains(err.Error(), "chart exceeds") {
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
		app := BundleAppV1{
			AppID:        fmt.Sprintf("app-%d", i),
			AppName:      fmt.Sprintf("app-%d", i),
			Version:      "1.0.0",
			InstallScope: InstallScopeShared,
			Chart:        name,
			ChartSHA256:  strings.Repeat("0", 64),
			AppEntry:     json.RawMessage(`{}`),
		}
		bundle.Apps = append(bundle.Apps, app)
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

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil ||
		!strings.Contains(err.Error(), "total chart size exceeds") {
		t.Fatalf("Materialize() error = %v", err)
	}
}

func TestMaterializeCopiesValidatedBundleAndBuildsReadOnlyProfile(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	selections := ProfileSelections{
		HardwareProfile: "nvidia-cuda",
		Apps: map[string]AppSelection{
			"app-a": {
				SelectedGPUType: "nvidia",
				Envs:            map[string]string{"WORKER_COUNT": "2"},
			},
		},
	}

	if err := Materialize(installerDir, baseDir, selections); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	contract, err := LoadDirectory(target)
	if err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
	if contract.Profile.HardwareProfile != "nvidia-cuda" ||
		contract.Profile.Apps[0].SelectedGPUType != "nvidia" ||
		contract.Profile.Apps[0].Envs["WORKER_COUNT"] != "2" {
		t.Fatalf("profile = %#v", contract.Profile)
	}
	assertMode(t, target, expectedRootMode())
	assertMode(t, filepath.Join(target, "charts"), 0o555)
	assertMode(t, filepath.Join(target, BundleFileName), 0o444)
	assertMode(t, filepath.Join(target, ProfileFileName), 0o444)
	assertMode(t, filepath.Join(target, "charts", "app-a-1.0.0.tgz"), 0o444)
}

func TestBuildProfileMergesDefaultsWithoutMutatingInputs(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedEnvs = []string{"DEFAULT_ONLY", "MODEL_PATH", "WORKER_COUNT"}
	bundle.Apps[0].DefaultEnvs = map[string]string{
		"DEFAULT_ONLY": "enabled",
		"MODEL_PATH":   "/models/default",
	}
	defaultsBefore := map[string]string{
		"DEFAULT_ONLY": "enabled",
		"MODEL_PATH":   "/models/default",
	}
	runtimeEnvs := map[string]string{
		"MODEL_PATH":   "/models/runtime",
		"WORKER_COUNT": "2",
	}
	runtimeBefore := map[string]string{
		"MODEL_PATH":   "/models/runtime",
		"WORKER_COUNT": "2",
	}
	selections := ProfileSelections{Apps: map[string]AppSelection{
		"app-a": {SelectedGPUType: "nvidia", Envs: runtimeEnvs},
	}}

	profile := buildProfile(bundle, selections)

	if len(profile.Apps) != 1 || profile.Apps[0].AppID != "app-a" {
		t.Fatalf("buildProfile() apps = %#v", profile.Apps)
	}
	if got := profile.Apps[0].Envs; got["DEFAULT_ONLY"] != "enabled" || got["MODEL_PATH"] != "/models/runtime" || got["WORKER_COUNT"] != "2" {
		t.Fatalf("buildProfile() envs = %#v", got)
	}
	if profile.Apps[0].SelectedGPUType != "nvidia" {
		t.Fatalf("buildProfile() selectedGpuType = %q", profile.Apps[0].SelectedGPUType)
	}
	if !mapsEqual(bundle.Apps[0].DefaultEnvs, defaultsBefore) {
		t.Fatalf("bundle defaults mutated: got %#v want %#v", bundle.Apps[0].DefaultEnvs, defaultsBefore)
	}
	if !mapsEqual(runtimeEnvs, runtimeBefore) {
		t.Fatalf("runtime envs mutated: got %#v want %#v", runtimeEnvs, runtimeBefore)
	}
	profile.Apps[0].Envs["MODEL_PATH"] = "changed"
	if bundle.Apps[0].DefaultEnvs["MODEL_PATH"] != "/models/default" || runtimeEnvs["MODEL_PATH"] != "/models/runtime" {
		t.Fatalf("profile envs alias input maps")
	}
}

func TestBuildProfileIncludesDefaultsOnlyAppsInDeterministicOrder(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AppID = "app-b"
	bundle.Apps[0].DefaultEnvs = map[string]string{"WORKER_COUNT": "1"}
	first := bundle.Apps[0]
	first.AppID = "app-a"
	first.AppName = "app-a"
	first.Chart = "charts/app-first.tgz"
	first.DefaultEnvs = map[string]string{"WORKER_COUNT": ""}
	bundle.Apps = append(bundle.Apps, first)

	profile := buildProfile(bundle, ProfileSelections{})

	if len(profile.Apps) != 2 || profile.Apps[0].AppID != "app-a" || profile.Apps[1].AppID != "app-b" {
		t.Fatalf("buildProfile() apps = %#v", profile.Apps)
	}
	if got := profile.Apps[0].Envs["WORKER_COUNT"]; got != "" {
		t.Fatalf("buildProfile() empty default = %q", got)
	}
}

func TestBuildProfileAppliesDetectedGPUOnlyToAllowedApps(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedGPUTypes = []string{"nvidia", "cpu"}
	unsupported := bundle.Apps[0]
	unsupported.AppID = "app-b"
	unsupported.AppName = "app-b"
	unsupported.AllowedGPUTypes = []string{"cpu"}
	unsupported.DefaultEnvs = map[string]string{"WORKER_COUNT": "1"}
	bundle.Apps = append(bundle.Apps, unsupported)

	profile := buildProfile(bundle, ProfileSelections{
		HardwareProfile: "nvidia",
		DetectedGPUType: "nvidia",
	})

	if profile.HardwareProfile != "nvidia" {
		t.Fatalf("buildProfile() hardwareProfile = %q", profile.HardwareProfile)
	}
	if len(profile.Apps) != 2 {
		t.Fatalf("buildProfile() apps = %#v", profile.Apps)
	}
	if profile.Apps[0].AppID != "app-a" || profile.Apps[0].SelectedGPUType != "nvidia" {
		t.Fatalf("buildProfile() allowed app = %#v", profile.Apps[0])
	}
	if profile.Apps[1].AppID != "app-b" || profile.Apps[1].SelectedGPUType != "" ||
		profile.Apps[1].Envs["WORKER_COUNT"] != "1" {
		t.Fatalf("buildProfile() unsupported app = %#v", profile.Apps[1])
	}
}

func TestBuildProfileExplicitGPUOverridesDetectedGPU(t *testing.T) {
	bundle := decodeBundle(t, validBundleJSON)
	bundle.Apps[0].AllowedGPUTypes = []string{"nvidia", "cpu"}

	profile := buildProfile(bundle, ProfileSelections{
		DetectedGPUType: "nvidia",
		Apps: map[string]AppSelection{
			"app-a": {SelectedGPUType: "cpu"},
		},
	})

	if len(profile.Apps) != 1 || profile.Apps[0].SelectedGPUType != "cpu" {
		t.Fatalf("buildProfile() apps = %#v", profile.Apps)
	}
}

func TestMaterializePublishesArtifactManifestWithoutPayload(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	artifact, manifestData := addMaterializeArtifactFixture(t, installerDir)

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}

	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	if _, err := LoadDirectory(target); err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
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
	assertMode(t, filepath.Join(target, "manifests"), 0o555)
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

			err := Materialize(installerDir, baseDir, ProfileSelections{})

			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("Materialize() error = %v, want symlink rejection", err)
			}
		})
	}
}

func TestMaterializeRejectsOversizedArtifactManifest(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	artifact, _ := addMaterializeArtifactFixture(t, installerDir)
	if err := os.Truncate(
		filepath.Join(installerDir, StaticRelativeDir, filepath.FromSlash(artifact.Manifest)),
		MaxArtifactManifestBytes+1,
	); err != nil {
		t.Fatal(err)
	}

	err := Materialize(installerDir, baseDir, ProfileSelections{})

	if err == nil || !strings.Contains(err.Error(), "manifest exceeds") {
		t.Fatalf("Materialize() error = %v, want oversized manifest rejection", err)
	}
}

func TestMaterializeRejectsArtifactManifestDigestMismatch(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	artifact, manifestData := addMaterializeArtifactFixture(t, installerDir)
	if err := os.WriteFile(
		filepath.Join(installerDir, StaticRelativeDir, filepath.FromSlash(artifact.Manifest)),
		append(manifestData, ' '),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	err := Materialize(installerDir, baseDir, ProfileSelections{})

	if err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("Materialize() error = %v, want manifest digest mismatch", err)
	}
}

func TestCopyArtifactManifestRejectsReplacementBetweenLstatAndOpen(t *testing.T) {
	installerDir, _ := writeStaticBundle(t)
	artifact, manifestData := addMaterializeArtifactFixture(t, installerDir)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	sourceRoot, err := os.OpenRoot(staticDir)
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	stagingRoot, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer stagingRoot.Close()
	if err := stagingRoot.MkdirAll(path.Dir(artifact.Manifest), 0o755); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest))

	err = copyArtifactManifest(sourceRoot, stagingRoot, artifact, artifactManifestCopyHooks{
		afterLstat: func() error {
			if err := os.Rename(manifestPath, manifestPath+".replaced"); err != nil {
				return err
			}
			return os.WriteFile(manifestPath, manifestData, 0o644)
		},
	})

	if err == nil || !strings.Contains(err.Error(), "changed while opening") {
		t.Fatalf("copyArtifactManifest() error = %v, want inode replacement rejection", err)
	}
}

func TestCopyArtifactManifestRejectsSourceChangedAfterOpen(t *testing.T) {
	installerDir, _ := writeStaticBundle(t)
	artifact, manifestData := addMaterializeArtifactFixture(t, installerDir)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	sourceRoot, err := os.OpenRoot(staticDir)
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	stagingPath := t.TempDir()
	stagingRoot, err := os.OpenRoot(stagingPath)
	if err != nil {
		t.Fatal(err)
	}
	defer stagingRoot.Close()
	if err := stagingRoot.MkdirAll(path.Dir(artifact.Manifest), 0o755); err != nil {
		t.Fatal(err)
	}

	err = copyArtifactManifest(sourceRoot, stagingRoot, artifact, artifactManifestCopyHooks{
		beforeCopy: func() error {
			return os.WriteFile(
				filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest)),
				append(manifestData, ' '),
				0o644,
			)
		},
	})

	if err == nil || !strings.Contains(err.Error(), "changed while copying") {
		t.Fatalf("copyArtifactManifest() error = %v, want source change rejection", err)
	}
}

func TestMaterializeRejectsSymlinkWithoutReplacingExistingTarget(t *testing.T) {
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

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want symlink rejection")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "existing" {
		t.Fatalf("existing target changed: content=%q error=%v", got, err)
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

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil {
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
		if err := Materialize(link, baseDir, ProfileSelections{}); err == nil {
			t.Fatal("Materialize() error = nil, want installer symlink rejection")
		}
	})

	t.Run("base", func(t *testing.T) {
		installerDir, realBase := writeStaticBundle(t)
		link := filepath.Join(t.TempDir(), "base")
		if err := os.Symlink(realBase, link); err != nil {
			t.Fatal(err)
		}
		if err := Materialize(installerDir, link, ProfileSelections{}); err == nil {
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

		if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
			t.Fatalf("Materialize() through symlinked installer ancestor error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(baseDir, RuntimeRelativeDir, BundleFileName)); err != nil {
			t.Fatalf("materialized bundle missing: %v", err)
		}
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

		if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
			t.Fatalf("Materialize() through symlinked base ancestor error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(realParent, "base", RuntimeRelativeDir, BundleFileName)); err != nil {
			t.Fatalf("materialized bundle missing under real base: %v", err)
		}
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

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want runtime parent symlink rejection")
	}
}

func TestMaterializeRejectsDigestMismatch(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	chart := filepath.Join(installerDir, StaticRelativeDir, "charts", "app-a-1.0.0.tgz")
	if err := os.WriteFile(chart, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err == nil {
		t.Fatal("Materialize() error = nil, want digest mismatch")
	}
}

func TestMaterializeAcceptsUppercaseHexDigest(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	bundlePath := filepath.Join(installerDir, StaticRelativeDir, BundleFileName)
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	var bundle BundleV1
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	bundle.Apps[0].ChartSHA256 = strings.ToUpper(bundle.Apps[0].ChartSHA256)
	data, err = json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundlePath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
}

func TestMaterializeReplacesExistingReadOnlyTarget(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("first Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("second Materialize() error = %v", err)
	}
	if _, err := LoadDirectory(target); err != nil {
		t.Fatalf("LoadDirectory() error = %v", err)
	}
}

// Published is what turns the market deployment's importer on, so it must
// answer for the tree Materialize leaves behind and for nothing else.
func TestPublishedFollowsWhatMaterializeLeavesBehind(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	if Published(baseDir) {
		t.Fatal("Published() is true before anything was published")
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })

	if !Published(baseDir) {
		t.Fatal("Published() is false for a bundle this run published")
	}
	if err := makeWritable(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(target, BundleFileName)); err != nil {
		t.Fatal(err)
	}
	if Published(baseDir) {
		t.Fatal("Published() is true for a directory with no bundle in it")
	}
}

// Every interrupted run leaves a staging directory holding a full copy of the
// bundle. Nothing else writes that name, so the next run clears them; a
// directory that only looks similar is left where it is.
func TestMaterializeRemovesStagingLeftBehindByAnInterruptedRun(t *testing.T) {
	installerDir, baseDir := writeStaticBundle(t)
	parent := filepath.Join(baseDir, filepath.FromSlash(path.Dir(RuntimeRelativeDir)))
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(parent, stagingPrefix+strings.Repeat("a", 16))
	unrelated := filepath.Join(parent, stagingPrefix+"not-a-staging-token")
	for _, dir := range []string{stale, unrelated} {
		if err := os.MkdirAll(filepath.Join(dir, "charts"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	t.Cleanup(func() { _ = makeWritable(filepath.Join(baseDir, RuntimeRelativeDir)) })

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

	if err := Materialize(installerDir, baseDir, ProfileSelections{}); err != nil {
		t.Fatalf("Materialize() error = %v", err)
	}
	target := filepath.Join(baseDir, RuntimeRelativeDir)
	t.Cleanup(func() { _ = makeWritable(target) })
	assertMode(t, filepath.Join(target, filepath.FromSlash(bundle.Apps[0].Chart)), 0o444)
	assertMode(t, filepath.Join(target, "charts", "nested"), 0o555)
}

func TestPopulateStagingSealsTreeBeforePublish(t *testing.T) {
	installerDir, _ := writeStaticBundle(t)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	sourceRoot, err := os.OpenRoot(staticDir)
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	bundleData, err := os.ReadFile(filepath.Join(staticDir, BundleFileName))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		t.Fatal(err)
	}
	staging := filepath.Join(t.TempDir(), "staging")
	if err := os.Mkdir(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	stagingRoot, err := os.OpenRoot(staging)
	if err != nil {
		t.Fatal(err)
	}
	defer stagingRoot.Close()

	if err := populateStaging(sourceRoot, stagingRoot, bundleData, bundle, buildProfile(bundle, ProfileSelections{})); err != nil {
		t.Fatalf("populateStaging() error = %v", err)
	}

	t.Cleanup(func() { _ = makeWritable(staging) })
	assertMode(t, staging, expectedRootMode())
	assertMode(t, filepath.Join(staging, "charts"), 0o555)
	assertMode(t, filepath.Join(staging, BundleFileName), 0o444)
}

func TestMaterializedRootModeByPlatform(t *testing.T) {
	tests := map[string]os.FileMode{
		"darwin":  0o755,
		"linux":   0o555,
		"windows": 0o555,
	}
	for goos, want := range tests {
		if got := materializedRootModeFor(goos); got != want {
			t.Errorf("materializedRootModeFor(%q) = %o, want %o", goos, got, want)
		}
	}
}

func TestRecoverPreviousRestoresMissingTarget(t *testing.T) {
	parent := canonicalTempDir(t)
	target := filepath.Join(parent, "market-preinstall")
	backup := target + ".previous"
	if err := os.Mkdir(backup, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(backup, "marker"), []byte("old"), 0o444); err != nil {
		t.Fatal(err)
	}
	if err := recoverPrevious(target); err != nil {
		t.Fatalf("recoverPrevious() error = %v", err)
	}

	t.Cleanup(func() { _ = makeWritable(target) })
	if got, err := os.ReadFile(filepath.Join(target, "marker")); err != nil || string(got) != "old" {
		t.Fatalf("recovered marker = %q, error = %v", got, err)
	}
	if _, err := os.Lstat(backup); !os.IsNotExist(err) {
		t.Fatalf("backup still exists: %v", err)
	}
}

func TestRecoverPreviousRemovesStaleBackupBesideTarget(t *testing.T) {
	parent := canonicalTempDir(t)
	target := filepath.Join(parent, "market-preinstall")
	backup := target + ".previous"
	for _, directory := range []string{target, backup} {
		if err := os.Mkdir(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(target, "marker"), []byte("current"), 0o444); err != nil {
		t.Fatal(err)
	}

	if err := recoverPrevious(target); err != nil {
		t.Fatalf("recoverPrevious() error = %v", err)
	}

	if got, err := os.ReadFile(filepath.Join(target, "marker")); err != nil || string(got) != "current" {
		t.Fatalf("current marker = %q, error = %v", got, err)
	}
	if _, err := os.Lstat(backup); !os.IsNotExist(err) {
		t.Fatalf("stale backup still exists: %v", err)
	}
}

func TestReplaceDirectoryRollsBackWhenStagingIsMissing(t *testing.T) {
	parentPath := t.TempDir()
	target := filepath.Join(parentPath, "market-preinstall")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "marker"), []byte("old"), 0o444); err != nil {
		t.Fatal(err)
	}
	parent, err := os.OpenRoot(parentPath)
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()

	err = replaceDirectoryRoot(parent, "market-preinstall", "missing-staging")

	if err == nil || !strings.Contains(err.Error(), "activate preinstall directory") {
		t.Fatalf("replaceDirectoryRoot() error = %v", err)
	}
	if got, readErr := os.ReadFile(filepath.Join(target, "marker")); readErr != nil || string(got) != "old" {
		t.Fatalf("rolled back marker = %q, error = %v", got, readErr)
	}
	if _, statErr := os.Lstat(target + ".previous"); !os.IsNotExist(statErr) {
		t.Fatalf("backup still exists: %v", statErr)
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

func expectedRootMode() os.FileMode {
	if runtime.GOOS == "darwin" {
		return 0o755
	}
	return 0o555
}
