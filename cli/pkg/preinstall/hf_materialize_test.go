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

const testHFRevision = "0123456789abcdef0123456789abcdef01234567"

func TestMaterializeHFArtifactsCreatesHuggingFaceLayout(t *testing.T) {
	installerDir, targetRoot, artifact, manifest := writeHFArtifactFixture(t)

	if err := materializeHFArtifacts(installerDir, targetRoot, nil); err != nil {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}

	modelRoot := filepath.Join(targetRoot, "models--acme--tiny-model")
	assertFileContent(t, filepath.Join(modelRoot, "refs", "main"), testHFRevision)
	assertFileContent(t, filepath.Join(modelRoot, "blobs", "weights"), "weights")
	link := filepath.Join(modelRoot, "snapshots", testHFRevision, "model.bin")
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("snapshot link info = %#v, error = %v", info, err)
	}
	if got, err := os.Readlink(link); err != nil || got != "../../blobs/weights" {
		t.Fatalf("snapshot link target = %q, error = %v", got, err)
	}
	assertMode(t, modelRoot, 0o755)
	assertMode(t, filepath.Join(modelRoot, "blobs", "weights"), 0o644)

	var marker hfCacheMarker
	data, err := os.ReadFile(filepath.Join(modelRoot, hfCacheMarkerFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &marker); err != nil {
		t.Fatal(err)
	}
	want := markerFor(artifact)
	if marker != want {
		t.Fatalf("marker = %#v, want %#v (manifest=%#v)", marker, want, manifest)
	}
}

func TestMaterializeHFArtifactsRejectsStreamDigestAndSizeMismatch(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(string)
		want   string
	}{
		{
			name: "digest",
			mutate: func(path string) {
				if err := os.WriteFile(path, []byte("WEIGHTS"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "digest mismatch",
		},
		{
			name: "size",
			mutate: func(path string) {
				if err := os.WriteFile(path, []byte("weights-extra"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: "size mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
			tt.mutate(filepath.Join(installerDir, StaticRelativeDir, "artifacts", "tiny", "blobs", "weights"))

			err := materializeHFArtifacts(installerDir, targetRoot, nil)

			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("materializeHFArtifacts() error = %v, want %q", err, tt.want)
			}
			assertNoHFStaging(t, targetRoot)
		})
	}
}

func TestMaterializeHFArtifactsRejectsActualSymlinkMismatchAndEscape(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "target mismatch", target: "../../blobs/other"},
		{name: "escape", target: "../../../../outside"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
			link := filepath.Join(installerDir, StaticRelativeDir, "artifacts", "tiny", "snapshots", testHFRevision, "model.bin")
			if err := os.Remove(link); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(tt.target, link); err != nil {
				t.Fatal(err)
			}

			err := materializeHFArtifacts(installerDir, targetRoot, nil)

			if err == nil || !strings.Contains(err.Error(), "symlink") {
				t.Fatalf("materializeHFArtifacts() error = %v", err)
			}
			assertNoHFStaging(t, targetRoot)
		})
	}
}

func TestMaterializeHFArtifactsRejectsSymlinkedEntryParent(t *testing.T) {
	installerDir, targetRoot, artifact, manifest := writeHFArtifactFixture(t)
	sourceRoot := filepath.Join(installerDir, StaticRelativeDir, "artifacts", "tiny")
	if err := os.Rename(filepath.Join(sourceRoot, "blobs"), filepath.Join(sourceRoot, "real-blobs")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real-blobs", filepath.Join(sourceRoot, "blobs")); err != nil {
		t.Fatal(err)
	}
	manifest.Entries = append(manifest.Entries[:2], manifest.Entries[3:]...)
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest := sha256.Sum256(manifestData)
	artifact.ManifestSHA256 = hex.EncodeToString(manifestDigest[:])
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	if err := os.WriteFile(filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest)), manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	writeBundleFile(t, staticDir, artifact)

	err = materializeHFArtifacts(installerDir, targetRoot, nil)

	if err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	assertNoHFStaging(t, targetRoot)
}

func TestSetHFOwnershipDoesNotOpenExternalSymlinkTargetAfterReplacement(t *testing.T) {
	rootPath := filepath.Join(t.TempDir(), "staging")
	if err := os.Mkdir(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a-trigger", "b-victim"} {
		if err := os.WriteFile(filepath.Join(rootPath, name), []byte("inside"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	external := filepath.Join(t.TempDir(), "external")
	if err := os.WriteFile(external, []byte("external"), 0o644); err != nil {
		t.Fatal(err)
	}
	externalInfo, err := os.Stat(external)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, filepath.Join(rootPath, "outside")); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	var opened int

	err = setHFOwnership(root, func(file *os.File, _, _ int) error {
		info, err := file.Stat()
		if err != nil {
			return err
		}
		if os.SameFile(info, externalInfo) {
			t.Fatal("external symlink target was opened for chown")
		}
		opened++
		if opened == 1 {
			if err := os.Remove(filepath.Join(rootPath, "b-victim")); err != nil {
				return err
			}
			if err := os.Symlink(external, filepath.Join(rootPath, "b-victim")); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		t.Fatalf("setHFOwnership() error = %v", err)
	}
	if opened != 2 {
		t.Fatalf("opened file count = %d, want trigger file and root directory", opened)
	}
}

func TestMaterializeHFArtifactsSkipsMatchingMarker(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	if err := materializeHFArtifacts(installerDir, targetRoot, nil); err != nil {
		t.Fatal(err)
	}
	modelRoot := filepath.Join(targetRoot, "models--acme--tiny-model")
	sentinel := filepath.Join(modelRoot, "sentinel")
	if err := os.WriteFile(sentinel, []byte("preserved"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := materializeHFArtifacts(installerDir, targetRoot, nil); err != nil {
		t.Fatalf("second materializeHFArtifacts() error = %v", err)
	}
	assertFileContent(t, sentinel, "preserved")
}

func TestMaterializeHFArtifactsRejectsExistingTargetWithoutMatchingMarker(t *testing.T) {
	for _, tt := range []struct {
		name       string
		markerData []byte
	}{
		{name: "missing"},
		{name: "different", markerData: []byte(`{"kind":"hf-cache"}`)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
			modelRoot := filepath.Join(targetRoot, "models--acme--tiny-model")
			if err := os.MkdirAll(modelRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			if tt.markerData != nil {
				if err := os.WriteFile(filepath.Join(modelRoot, hfCacheMarkerFileName), tt.markerData, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			err := materializeHFArtifacts(installerDir, targetRoot, nil)

			if err == nil || !strings.Contains(err.Error(), "existing target") {
				t.Fatalf("materializeHFArtifacts() error = %v", err)
			}
		})
	}
}

func TestMaterializeHFArtifactsRejectsDuplicateRepoBeforeCopy(t *testing.T) {
	installerDir, targetRoot, artifact, _ := writeHFArtifactFixture(t)
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	bundlePath := filepath.Join(staticDir, BundleFileName)
	data, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := DecodeBundle(data)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := bundle.Apps[0]
	duplicate.AppID = "app-b"
	duplicate.AppName = "app-b"
	duplicate.Chart = "charts/app-b.tgz"
	duplicate.Artifacts = []BundleArtifactV1{artifact}
	bundle.Apps = append(bundle.Apps, duplicate)
	data, err = json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bundlePath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	err = materializeHFArtifacts(installerDir, targetRoot, nil)

	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	if _, err := os.Lstat(filepath.Join(targetRoot, "models--acme--tiny-model")); !os.IsNotExist(err) {
		t.Fatalf("duplicate artifact copied before rejection: %v", err)
	}
}

func TestMaterializeHFArtifactsNoReplacePreservesEmptyTargetCreatedAtRename(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	target := "models--acme--tiny-model"
	hooks := hfMaterializeHooks{
		beforeRename: func(root *os.Root, target string) error {
			return root.Mkdir(target, 0o755)
		},
	}

	err := materializeHFArtifactsWithHooks(installerDir, targetRoot, nil, hooks)

	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("materializeHFArtifactsWithHooks() error = %v", err)
	}
	entries, readErr := os.ReadDir(filepath.Join(targetRoot, target))
	if readErr != nil || len(entries) != 0 {
		t.Fatalf("competing target entries = %v, error = %v", entries, readErr)
	}
	assertNoHFStaging(t, targetRoot)
}

func TestMaterializeHFArtifactsPreservesForgedUserOwnedStaging(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	target := "models--acme--tiny-model"
	token := strings.Repeat("e", 32)
	stage := filepath.Join(targetRoot, hfStagingPrefix(target)+token)
	if err := os.Mkdir(stage, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := hfStageMarker{Target: target, Repo: "acme/tiny-model", Token: token}
	data, err := json.Marshal(marker)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stage, hfStageMarkerFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}
	if os.Geteuid() == 0 {
		if err := os.Chown(stage, 1000, 1000); err != nil {
			t.Fatal(err)
		}
	}
	trustedUID := uint32(0)
	hooks := hfMaterializeHooks{trustedStageUID: &trustedUID}

	err = materializeHFArtifactsWithHooks(installerDir, targetRoot, nil, hooks)

	if err == nil || !strings.Contains(err.Error(), "untrusted staging owner") {
		t.Fatalf("materializeHFArtifactsWithHooks() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(stage, hfStageMarkerFileName)); err != nil {
		t.Fatalf("forged staging was removed: %v", err)
	}
}

func TestMaterializeHFArtifactsDeferPreservesReplacedStaging(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	var replacement string
	hooks := hfMaterializeHooks{
		afterStageCreated: func(parent *os.Root, name string, stage *os.Root) error {
			info, err := stage.Stat(".")
			if err != nil {
				return err
			}
			if info.Mode().Perm() != 0o700 {
				return fmt.Errorf("staging mode = %o", info.Mode().Perm())
			}
			if err := parent.RemoveAll(name); err != nil {
				return err
			}
			if err := parent.Mkdir(name, 0o700); err != nil {
				return err
			}
			file, err := parent.OpenFile(name+"/replacement", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
			if err != nil {
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
			replacement = name
			return fmt.Errorf("injected failure")
		},
	}

	err := materializeHFArtifactsWithHooks(installerDir, targetRoot, nil, hooks)

	if err == nil || !strings.Contains(err.Error(), "injected failure") {
		t.Fatalf("materializeHFArtifactsWithHooks() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetRoot, replacement, "replacement")); err != nil {
		t.Fatalf("replacement staging was removed: %v", err)
	}
}

func TestMaterializeHFArtifactsWritesCompletionMarkerAfterOwnership(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	ownership := &hfOwnership{
		tree: func(root *os.Root) error {
			if _, err := root.Lstat(hfCacheMarkerFileName); !os.IsNotExist(err) {
				return fmt.Errorf("completion marker existed during ownership: %v", err)
			}
			return fmt.Errorf("injected ownership failure")
		},
		trustedID: uint32(os.Geteuid()),
	}

	err := materializeHFArtifacts(installerDir, targetRoot, ownership)

	if err == nil || !strings.Contains(err.Error(), "injected ownership failure") {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	marker := filepath.Join(targetRoot, "models--acme--tiny-model", hfCacheMarkerFileName)
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		t.Fatalf("completion marker exists after ownership failure: %v", err)
	}
}

func TestMaterializeHFArtifactsRemovesCompletionMarkerWhenMarkerOwnershipFails(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	ownership := &hfOwnership{
		tree: func(*os.Root) error { return nil },
		marker: func(*os.File) error {
			return fmt.Errorf("injected marker ownership failure")
		},
		trustedID: uint32(os.Geteuid()),
	}

	err := materializeHFArtifacts(installerDir, targetRoot, ownership)

	if err == nil || !strings.Contains(err.Error(), "injected marker ownership failure") {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	marker := filepath.Join(targetRoot, "models--acme--tiny-model", hfCacheMarkerFileName)
	if _, err := os.Lstat(marker); !os.IsNotExist(err) {
		t.Fatalf("completion marker remains after failure: %v", err)
	}
}

func TestMaterializeHFArtifactsCleansOwnedStaging(t *testing.T) {
	installerDir, targetRoot, _, _ := writeHFArtifactFixture(t)
	target := "models--acme--tiny-model"
	token := strings.Repeat("a", 32)
	stale := filepath.Join(targetRoot, hfStagingPrefix(target)+token)
	if err := os.MkdirAll(stale, 0o700); err != nil {
		t.Fatal(err)
	}
	marker := `{"target":"models--acme--tiny-model","repo":"acme/tiny-model","token":"` + token + `"}`
	if err := os.WriteFile(filepath.Join(stale, hfStageMarkerFileName), []byte(marker), 0o644); err != nil {
		t.Fatal(err)
	}
	preserved := []string{
		hfStagingPrefix(target) + "not-a-token",
		hfStagingPrefix(target) + strings.Repeat("b", 32),
	}
	for index, name := range preserved {
		dir := filepath.Join(targetRoot, name)
		mode := os.FileMode(0o755)
		if index == 1 {
			mode = 0o700
		}
		if err := os.MkdirAll(dir, mode); err != nil {
			t.Fatal(err)
		}
	}
	wrongMarker := `{"target":"models--acme--tiny-model","repo":"other/repo","token":"` + strings.Repeat("b", 32) + `"}`
	if err := os.WriteFile(filepath.Join(targetRoot, preserved[1], hfStageMarkerFileName), []byte(wrongMarker), 0o644); err != nil {
		t.Fatal(err)
	}
	symlinkStage := hfStagingPrefix(target) + strings.Repeat("c", 32)
	if err := os.Symlink(t.TempDir(), filepath.Join(targetRoot, symlinkStage)); err != nil {
		t.Fatal(err)
	}
	fileStage := hfStagingPrefix(target) + strings.Repeat("d", 32)
	if err := os.WriteFile(filepath.Join(targetRoot, fileStage), []byte("user data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := materializeHFArtifacts(installerDir, targetRoot, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(stale); !os.IsNotExist(err) {
		t.Fatalf("owned stale staging still exists: %v", err)
	}
	for _, name := range append(preserved, symlinkStage, fileStage) {
		if _, err := os.Lstat(filepath.Join(targetRoot, name)); err != nil {
			t.Fatalf("unowned staging entry %q changed: %v", name, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(targetRoot, target, hfStageMarkerFileName)); !os.IsNotExist(err) {
		t.Fatalf("published target contains staging marker: %v", err)
	}
}

func TestMaterializeHFArtifactsNoOpsWithoutArtifacts(t *testing.T) {
	installerDir, targetRoot := writeBundleWithoutHFArtifact(t)

	if err := materializeHFArtifacts(installerDir, targetRoot, nil); err != nil {
		t.Fatalf("materializeHFArtifacts() error = %v", err)
	}
	if entries, err := os.ReadDir(targetRoot); err != nil || len(entries) != 0 {
		t.Fatalf("target entries = %v, error = %v", entries, err)
	}
}

func writeHFArtifactFixture(t *testing.T) (string, string, BundleArtifactV1, ArtifactManifestV1) {
	t.Helper()
	root := canonicalTempDir(t)
	installerDir := filepath.Join(root, "installer")
	targetRoot := filepath.Join(root, "huggingface")
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	sourceDir := filepath.Join(staticDir, "artifacts", "tiny")
	for _, dir := range []string{
		filepath.Join(sourceDir, "refs"),
		filepath.Join(sourceDir, "blobs"),
		filepath.Join(sourceDir, "snapshots", testHFRevision),
		filepath.Join(staticDir, "manifests"),
		targetRoot,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"refs/main":     testHFRevision,
		"blobs/weights": "weights",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sourceDir, filepath.FromSlash(name)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("../../blobs/weights", filepath.Join(sourceDir, "snapshots", testHFRevision, "model.bin")); err != nil {
		t.Fatal(err)
	}
	entries := []ArtifactManifestEntryV1{
		{Path: "refs", Type: "directory"},
		fileManifestEntry("refs/main", files["refs/main"]),
		{Path: "blobs", Type: "directory"},
		fileManifestEntry("blobs/weights", files["blobs/weights"]),
		{Path: "snapshots", Type: "directory"},
		{Path: "snapshots/" + testHFRevision, Type: "directory"},
		{Path: "snapshots/" + testHFRevision + "/model.bin", Type: "symlink", Target: "../../blobs/weights"},
	}
	manifest := ArtifactManifestV1{
		SchemaVersion: 1,
		Repo:          "acme/tiny-model",
		Revision:      testHFRevision,
		Entries:       entries,
	}
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestDigest := sha256.Sum256(manifestData)
	artifact := BundleArtifactV1{
		Kind:           ArtifactKindHFCache,
		Source:         "artifacts/tiny",
		Repo:           manifest.Repo,
		Revision:       manifest.Revision,
		Manifest:       "manifests/tiny.json",
		ManifestSHA256: hex.EncodeToString(manifestDigest[:]),
		TotalSize:      int64(len(files["refs/main"]) + len(files["blobs/weights"])),
	}
	if err := os.WriteFile(filepath.Join(staticDir, filepath.FromSlash(artifact.Manifest)), manifestData, 0o644); err != nil {
		t.Fatal(err)
	}
	writeBundleFile(t, staticDir, artifact)
	return installerDir, targetRoot, artifact, manifest
}

func writeBundleWithoutHFArtifact(t *testing.T) (string, string) {
	t.Helper()
	root := canonicalTempDir(t)
	installerDir := filepath.Join(root, "installer")
	staticDir := filepath.Join(installerDir, StaticRelativeDir)
	targetRoot := filepath.Join(root, "huggingface")
	if err := os.MkdirAll(targetRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeBundleFile(t, staticDir, BundleArtifactV1{})
	return installerDir, targetRoot
}

func writeBundleFile(t *testing.T, staticDir string, artifact BundleArtifactV1) {
	t.Helper()
	if err := os.MkdirAll(staticDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bundle := decodeBundle(t, validBundleJSON)
	if artifact.Kind != "" {
		bundle.Apps[0].Artifacts = []BundleArtifactV1{artifact}
	}
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, BundleFileName), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func fileManifestEntry(path, content string) ArtifactManifestEntryV1 {
	sum := sha256.Sum256([]byte(content))
	return ArtifactManifestEntryV1{
		Path:   path,
		Type:   "file",
		Size:   int64(len(content)),
		SHA256: hex.EncodeToString(sum[:]),
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Fatalf("%s content = %q, want %q", path, data, want)
	}
}

func assertNoHFStaging(t *testing.T, targetRoot string) {
	t.Helper()
	entries, err := os.ReadDir(targetRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".olares-hf-stage-") {
			t.Fatalf("staging directory remains: %s", entry.Name())
		}
	}
}
