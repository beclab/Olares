package preinstall

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CheckOptions controls how thoroughly CheckStaticBundle inspects a
// static market bundle directory.
type CheckOptions struct {
	// Full also verifies artifact payload trees under each artifact's
	// source directory against the corresponding manifest entries.
	Full bool
}

// CheckStaticBundle validates a static market bundle directory — the
// directory that directly contains bundle.json (typically
// preinstall/market). It is read-only and does not publish runtime
// state or Hugging Face caches.
//
// By default it checks the V1 JSON contract, chart presence/size/SHA-256,
// and artifact manifests. With opts.Full it also verifies each artifact
// source tree matches its manifest (entry types, digests, and no
// undeclared paths).
func CheckStaticBundle(dir string, opts CheckOptions) error {
	dir = filepath.Clean(dir)
	root, err := openDirectoryNoSymlink(dir)
	if err != nil {
		return fmt.Errorf("open static bundle: %w", err)
	}
	defer root.Close()

	bundleData, err := readRootFileLimited(root, BundleFileName, MaxBundleJSONBytes)
	if err != nil {
		return err
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		return err
	}
	profile := buildProfile(bundle, ProfileSelections{})
	if err := Validate(bundle, profile); err != nil {
		return err
	}
	if err := preflightCharts(root, bundle); err != nil {
		return err
	}
	for _, app := range bundle.Apps {
		if err := verifyChartDigest(root, app); err != nil {
			return fmt.Errorf("bundle app %q: %w", app.AppID, err)
		}
	}
	for _, app := range bundle.Apps {
		for _, artifact := range app.Artifacts {
			manifest, err := LoadArtifactManifest(root, artifact)
			if err != nil {
				return fmt.Errorf("bundle app %q artifact manifest: %w", app.AppID, err)
			}
			if !opts.Full {
				continue
			}
			if err := verifyArtifactSource(root, artifact, manifest); err != nil {
				return fmt.Errorf("bundle app %q artifact %q: %w", app.AppID, artifact.Source, err)
			}
		}
	}
	return nil
}

func verifyChartDigest(root *os.Root, app BundleAppV1) error {
	if err := rejectRootSymlinkComponents(root, app.Chart); err != nil {
		return err
	}
	info, err := root.Lstat(app.Chart)
	if err != nil {
		return fmt.Errorf("inspect chart %q: %w", app.Chart, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("chart %q must be a regular file", app.Chart)
	}
	if hasMultipleLinks(info) {
		return fmt.Errorf("chart %q must not be a hardlink", app.Chart)
	}
	if info.Size() > MaxChartBytes {
		return fmt.Errorf("chart exceeds %d bytes: %q", MaxChartBytes, app.Chart)
	}
	return verifyRootFileDigest(root, app.Chart, info.Size(), app.ChartSHA256)
}

func verifyRootFileDigest(root *os.Root, name string, size int64, wantSHA256 string) error {
	file, err := openRootRegularFile(root, name)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("fstat %q: %w", name, err)
	}
	if info.Size() != size {
		return fmt.Errorf("%q size mismatch: got %d, want %d", name, info.Size(), size)
	}
	hasher := sha256.New()
	copied, err := io.Copy(hasher, io.LimitReader(file, size+1))
	if err != nil {
		return fmt.Errorf("hash %q: %w", name, err)
	}
	if copied != size {
		return fmt.Errorf("%q size changed while hashing", name)
	}
	got := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(got, wantSHA256) {
		return fmt.Errorf("%q digest mismatch", name)
	}
	return nil
}

func verifyArtifactSource(staticRoot *os.Root, artifact BundleArtifactV1, manifest ArtifactManifestV1) error {
	if err := rejectRootSymlinkComponents(staticRoot, artifact.Source); err != nil {
		return err
	}
	sourceInfo, err := staticRoot.Lstat(artifact.Source)
	if err != nil {
		return fmt.Errorf("inspect artifact source %q: %w", artifact.Source, err)
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("artifact source %q must be a directory", artifact.Source)
	}
	sourceRoot, err := staticRoot.OpenRoot(artifact.Source)
	if err != nil {
		return fmt.Errorf("open artifact source %q: %w", artifact.Source, err)
	}
	defer sourceRoot.Close()

	declared := make(map[string]struct{}, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if err := verifyArtifactEntry(sourceRoot, entry); err != nil {
			return err
		}
		declared[entry.Path] = struct{}{}
	}
	return rejectUndeclaredArtifactPaths(sourceRoot, declared)
}

func verifyArtifactEntry(sourceRoot *os.Root, entry ArtifactManifestEntryV1) error {
	if err := rejectHFReservedTarget(entry.Path); err != nil {
		return err
	}
	if parent := path.Dir(entry.Path); parent != "." {
		if err := rejectRootSymlinkComponents(sourceRoot, parent); err != nil {
			return err
		}
	}
	info, err := sourceRoot.Lstat(entry.Path)
	if err != nil {
		return fmt.Errorf("inspect artifact entry %q: %w", entry.Path, err)
	}
	switch entry.Type {
	case "directory":
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("artifact entry %q must be a directory", entry.Path)
		}
		return nil
	case "file":
		if !info.Mode().IsRegular() {
			return fmt.Errorf("artifact entry %q must be a regular file", entry.Path)
		}
		if info.Size() != entry.Size {
			return fmt.Errorf("artifact file %q size mismatch: got %d, want %d", entry.Path, info.Size(), entry.Size)
		}
		return verifyRootFileDigest(sourceRoot, entry.Path, entry.Size, entry.SHA256)
	case "symlink":
		if info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("artifact entry %q must be a symlink", entry.Path)
		}
		target, err := sourceRoot.Readlink(entry.Path)
		if err != nil {
			return fmt.Errorf("read artifact symlink %q: %w", entry.Path, err)
		}
		if target != entry.Target {
			return fmt.Errorf("artifact symlink %q target mismatch", entry.Path)
		}
		if err := validateSymlinkTarget(entry.Path, target); err != nil {
			return fmt.Errorf("artifact symlink %q target: %w", entry.Path, err)
		}
		return nil
	default:
		return fmt.Errorf("artifact entry %q has unsupported type %q", entry.Path, entry.Type)
	}
}

func rejectUndeclaredArtifactPaths(sourceRoot *os.Root, declared map[string]struct{}) error {
	return fs.WalkDir(sourceRoot.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		relative := path.Clean(filepath.ToSlash(name))
		if relative == "." {
			return nil
		}
		if _, ok := declared[relative]; !ok {
			return fmt.Errorf("undeclared artifact path %q", relative)
		}
		return nil
	})
}
