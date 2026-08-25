package preinstall

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Publish writes this version's declaration, and the charts and artifact
// manifests it names, under rootDir/RuntimeRelativeDir. rootDir must be the
// Olares root directory the market chart mounts from, not the installer base
// directory.
//
// An empty installerDir, or one carrying no bundle, still publishes: the catalog
// apps this version expects are declared either way, and an upgrade that brought
// no medium is the ordinary case rather than a reason to leave Market with no
// declaration for the release it is now running.
func Publish(installerDir, rootDir, osVersion string, selections ProfileSelections) error {
	trunk := TrunkVersion(osVersion)
	if trunk == "" {
		return fmt.Errorf("publish declaration: no Olares version to name it after")
	}
	source, err := openStaticBundle(installerDir)
	if err != nil {
		return err
	}
	defer source.close()
	if source.present() {
		if err := Validate(source.bundle); err != nil {
			return err
		}
		if err := preflightCharts(source.root, source.bundle); err != nil {
			return err
		}
	}
	declaration, err := BuildDeclaration(source.bundle, selections, catalogDeclarationApps())
	if err != nil {
		return err
	}
	return publishDeclaration(source, rootDir, trunk, declaration)
}

// publishDeclaration writes one trunk's declaration and does nothing if that
// trunk already has one. The declaration is renamed into place last, so its
// presence is what says the payload beside it is complete; a publish interrupted
// halfway leaves verified files and no declaration, and the next attempt
// finishes the job.
func publishDeclaration(source medium, rootDir, trunk string, declaration DeclarationV2) error {
	declarationData, err := json.MarshalIndent(declaration, "", "  ")
	if err != nil {
		return fmt.Errorf("encode declaration: %w", err)
	}
	declarationData = append(declarationData, '\n')
	if len(declarationData) > MaxDeclarationBytes {
		return fmt.Errorf("declaration exceeds %d bytes", MaxDeclarationBytes)
	}

	runtimeRoot, err := openRuntimeDirectory(rootDir)
	if err != nil {
		return err
	}
	defer runtimeRoot.Close()

	name := DeclarationFileName(trunk)
	published, err := regularFileExists(runtimeRoot, name)
	if err != nil {
		return err
	}
	if published {
		// Rewriting it would put a second answer in front of a device that has
		// already acted on the first one, and every build of a release shares
		// this file.
		return nil
	}

	if err := cleanupStagingRoots(runtimeRoot); err != nil {
		return err
	}
	stagingName, stagingRoot, err := createStagingRoot(runtimeRoot)
	if err != nil {
		return err
	}
	defer func() {
		if stagingRoot != nil {
			_ = stagingRoot.Close()
		}
		_ = removeRootTree(runtimeRoot, stagingName)
	}()
	payload, err := stagePayload(source, stagingRoot, declaration)
	if err != nil {
		return err
	}
	if err := writeSealedRootFile(stagingRoot, name, declarationData); err != nil {
		return err
	}
	if err := stagingRoot.Close(); err != nil {
		return fmt.Errorf("close preinstall staging root: %w", err)
	}
	stagingRoot = nil

	for _, relative := range payload {
		if err := promoteStagedFile(runtimeRoot, stagingName, relative); err != nil {
			return err
		}
	}
	return promoteStagedFile(runtimeRoot, stagingName, name)
}

// openRuntimeDirectory opens the published directory, creating it if this is the
// first publish. It stays writable: a device accumulates one declaration per
// release it has run, so the directory has to be added to again. The pod mounts
// it read-only, and every published file is sealed 0o444.
func openRuntimeDirectory(rootDir string) (*os.Root, error) {
	olaresRoot, err := openOrCreateDirectoryNoSymlink(rootDir)
	if err != nil {
		return nil, fmt.Errorf("open olares root: %w", err)
	}
	defer olaresRoot.Close()
	if err := olaresRoot.MkdirAll(RuntimeRelativeDir, 0o755); err != nil {
		return nil, fmt.Errorf("create preinstall directory: %w", err)
	}
	if err := rejectRootSymlinkComponents(olaresRoot, RuntimeRelativeDir); err != nil {
		return nil, err
	}
	runtimePath := filepath.Join(rootDir, filepath.FromSlash(RuntimeRelativeDir))
	runtimeRoot, err := openDirectoryNoSymlink(runtimePath)
	if err != nil {
		return nil, fmt.Errorf("open preinstall directory: %w", err)
	}
	if err := rejectRootSymlinkComponents(runtimeRoot, "."); err != nil {
		_ = runtimeRoot.Close()
		return nil, err
	}
	return runtimeRoot, nil
}

func regularFileExists(root *os.Root, name string) (bool, error) {
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect %q: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("%q must be a regular file", name)
	}
	return true, nil
}

// stagePayload copies every file the declaration names into staging, verifying
// each digest on the way, and returns their paths in the order they should be
// promoted. A declaration with no local entry stages nothing.
func stagePayload(source medium, stagingRoot *os.Root, declaration DeclarationV2) ([]string, error) {
	var (
		files []string
		total int64
	)
	for _, app := range declaration.Apps {
		if !app.local() {
			continue
		}
		if !source.present() {
			return nil, fmt.Errorf("declared app %q needs a chart no medium carries", app.AppID)
		}
		if err := stagingRoot.MkdirAll(path.Dir(app.Chart), 0o755); err != nil {
			return nil, fmt.Errorf("create chart staging directory: %w", err)
		}
		copied, err := copyChart(source.root, stagingRoot, app, MaxTotalChartBytes-total)
		if err != nil {
			return nil, err
		}
		total += copied
		files = append(files, app.Chart)
		for _, artifact := range app.Artifacts {
			if err := stagingRoot.MkdirAll(path.Dir(artifact.Manifest), 0o755); err != nil {
				return nil, fmt.Errorf("create artifact manifest staging directory: %w", err)
			}
			if err := copyArtifactManifest(source.root, stagingRoot, artifact); err != nil {
				return nil, err
			}
			files = append(files, artifact.Manifest)
		}
	}
	sort.Strings(files)
	return files, nil
}

// promoteStagedFile moves one staged file to where Market reads it. Rename over
// an existing name is what makes a republish of the same payload safe: readers
// see either the old file or the new one, and never a half-written one.
func promoteStagedFile(runtimeRoot *os.Root, stagingName, relative string) error {
	directory := path.Dir(relative)
	if directory != "." {
		if err := runtimeRoot.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create %q: %w", directory, err)
		}
	}
	if err := runtimeRoot.Rename(path.Join(stagingName, relative), relative); err != nil {
		return fmt.Errorf("publish %q: %w", relative, err)
	}
	if err := syncRootDirectory(runtimeRoot, directory); err != nil {
		return err
	}
	if directory == "." {
		return nil
	}
	return syncRootDirectory(runtimeRoot, ".")
}

func containsString(values []string, target string) bool {
	if target == "" {
		return false
	}
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func preflightCharts(root *os.Root, bundle BundleV1) error {
	var total int64
	for _, app := range bundle.Apps {
		if err := rejectRootSymlinkComponents(root, app.Chart); err != nil {
			return err
		}
		info, err := root.Stat(app.Chart)
		if err != nil {
			return fmt.Errorf("stat chart %q: %w", app.Chart, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("chart %q must be a regular file", app.Chart)
		}
		if info.Size() > MaxChartBytes {
			return fmt.Errorf("chart exceeds %d bytes: %q", MaxChartBytes, app.Chart)
		}
		if info.Size() > MaxTotalChartBytes-total {
			return fmt.Errorf("total chart size exceeds %d bytes", MaxTotalChartBytes)
		}
		total += info.Size()
	}
	return nil
}

func copyArtifactManifest(
	sourceRoot, stagingRoot *os.Root,
	artifact DeclarationArtifact,
) error {
	if err := rejectRootSymlinkComponents(sourceRoot, artifact.Manifest); err != nil {
		return err
	}
	info, err := sourceRoot.Lstat(artifact.Manifest)
	if err != nil {
		return fmt.Errorf("lstat artifact manifest %q: %w", artifact.Manifest, err)
	}
	if info.Size() > MaxArtifactManifestBytes {
		return fmt.Errorf("artifact manifest exceeds %d bytes: %q", MaxArtifactManifestBytes, artifact.Manifest)
	}
	_, err = copyVerifiedRegularFile(sourceRoot, stagingRoot, verifiedCopy{
		Source:     artifact.Manifest,
		Target:     artifact.Manifest,
		Size:       info.Size(),
		MaxSize:    MaxArtifactManifestBytes,
		SHA256:     artifact.ManifestSHA256,
		OutputMode: 0o444,
	})
	return err
}

func copyChart(sourceRoot, stagingRoot *os.Root, app DeclarationAppV2, totalRemaining int64) (int64, error) {
	spec, err := chartVerifiedCopy(sourceRoot, app.Chart, app.ChartSHA256, totalRemaining)
	if err != nil {
		return 0, err
	}
	return copyVerifiedRegularFile(sourceRoot, stagingRoot, spec)
}

func chartVerifiedCopy(sourceRoot *os.Root, chart, sha256 string, totalRemaining int64) (verifiedCopy, error) {
	info, err := sourceRoot.Lstat(chart)
	if err != nil {
		return verifiedCopy{}, fmt.Errorf("inspect chart %q: %w", chart, err)
	}
	if info.Size() > totalRemaining {
		return verifiedCopy{}, fmt.Errorf("total chart size exceeds %d bytes", MaxTotalChartBytes)
	}
	return verifiedCopy{
		Source:     chart,
		Target:     chart,
		Size:       info.Size(),
		MaxSize:    min(MaxChartBytes, totalRemaining),
		SHA256:     sha256,
		OutputMode: 0o444,
	}, nil
}

func writeSealedRootFile(root *os.Root, name string, data []byte) error {
	return writeRootFile(root, name, data, rootFileWrite{Mode: 0o444})
}

func openOrCreateDirectoryNoSymlink(name string) (*os.Root, error) {
	return openDirectoryPath(name, true)
}

const stagingPrefix = ".market-preinstall-stage-"

// cleanupStagingRoots removes trusted leftovers from interrupted publishes.
func cleanupStagingRoots(parent *os.Root) error {
	entries, err := fs.ReadDir(parent.FS(), ".")
	if err != nil {
		return fmt.Errorf("list preinstall directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if !stagingName(name) {
			continue
		}
		stage, _, trusted, err := openTrustedStaging(parent, name, uint32(os.Geteuid()), 0o700, 0o755, 0o555)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("open stale preinstall staging %q: %w", name, err)
		}
		if !trusted {
			continue
		}
		if err := stage.Close(); err != nil {
			return fmt.Errorf("close stale preinstall staging %q: %w", name, err)
		}
		if err := removeRootTree(parent, name); err != nil {
			return fmt.Errorf("remove stale preinstall staging %q: %w", name, err)
		}
	}
	return nil
}

// stagingName reports whether the name is one createStagingRoot could have
// produced, so cleanup never touches a directory that merely starts the same.
func stagingName(name string) bool {
	suffix, found := strings.CutPrefix(name, stagingPrefix)
	if !found || len(suffix) != 16 {
		return false
	}
	_, err := hex.DecodeString(suffix)
	return err == nil
}

func createStagingRoot(parent *os.Root) (string, *os.Root, error) {
	name, _, root, _, err := createTrustedStaging(parent, stagingPrefix, 8, 0o700)
	return name, root, err
}

func makeWritable(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.Chmod(path, 0o755)
		}
		return nil
	})
}

func removeRootTree(parent *os.Root, name string) error {
	if _, err := parent.Lstat(name); os.IsNotExist(err) {
		return nil
	}
	root, err := parent.OpenRoot(name)
	if err != nil {
		return fmt.Errorf("open tree %q for removal: %w", name, err)
	}
	if err := makeRootWritable(root); err != nil {
		_ = root.Close()
		return err
	}
	if err := root.Close(); err != nil {
		return fmt.Errorf("close tree %q before removal: %w", name, err)
	}
	if err := parent.RemoveAll(name); err != nil {
		return fmt.Errorf("remove tree %q: %w", name, err)
	}
	return nil
}

func makeRootWritable(root *os.Root) error {
	return fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return root.Chmod(name, 0o755)
		}
		return nil
	})
}

func syncRootDirectory(root *os.Root, name string) error {
	dir, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("open directory %q: %w", name, err)
	}
	defer dir.Close()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync directory %q: %w", name, err)
	}
	return nil
}
