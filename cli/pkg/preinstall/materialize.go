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
	"runtime"
	"sort"
	"strings"
)

// Materialize publishes the bundle under rootDir/RuntimeRelativeDir. rootDir
// must be the Olares root directory the market chart mounts from, not the
// installer base directory.
func Materialize(installerDir, rootDir string, selections ProfileSelections) error {
	sourceRoot, bundleData, bundle, found, err := openStaticBundle(installerDir)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	defer sourceRoot.Close()
	profile := buildProfile(bundle, selections)
	if err := Validate(bundle, profile); err != nil {
		return err
	}
	if err := preflightCharts(sourceRoot, bundle); err != nil {
		return err
	}

	olaresRoot, err := openOrCreateDirectoryNoSymlink(rootDir)
	if err != nil {
		return fmt.Errorf("open olares root: %w", err)
	}
	defer olaresRoot.Close()
	parentRelative := path.Dir(RuntimeRelativeDir)
	targetName := path.Base(RuntimeRelativeDir)
	if err := olaresRoot.MkdirAll(parentRelative, 0o755); err != nil {
		return fmt.Errorf("create preinstall parent: %w", err)
	}
	if err := rejectRootSymlinkComponents(olaresRoot, parentRelative); err != nil {
		return err
	}
	parentPath := filepath.Join(rootDir, filepath.FromSlash(parentRelative))
	parentRoot, err := openDirectoryNoSymlink(parentPath)
	if err != nil {
		return fmt.Errorf("open preinstall parent: %w", err)
	}
	defer parentRoot.Close()
	if err := rejectRootSymlinkComponents(parentRoot, "."); err != nil {
		return err
	}
	if _, err := directoryStateRoot(parentRoot, targetName); err != nil {
		return err
	}

	if err := cleanupStagingRoots(parentRoot); err != nil {
		return err
	}

	stagingName, stagingRoot, err := createStagingRoot(parentRoot)
	if err != nil {
		return err
	}
	defer func() {
		if stagingRoot != nil {
			_ = stagingRoot.Close()
		}
		_ = removeRootTree(parentRoot, stagingName)
	}()
	if err := populateStaging(sourceRoot, stagingRoot, bundleData, bundle, profile); err != nil {
		return err
	}
	if err := stagingRoot.Close(); err != nil {
		return fmt.Errorf("close preinstall staging root: %w", err)
	}
	stagingRoot = nil
	return replaceDirectoryRoot(parentRoot, targetName, stagingName)
}

// Published reports whether rootDir carries a published bundle, which is what
// decides whether the market deployment is rendered with preinstall turned on.
// It answers on the presence of the bundle file rather than on a full load:
// telling Market a bundle is there and letting it report what is wrong with it
// is more useful than turning the feature off over a detail this program did
// not manage to parse.
func Published(rootDir string) bool {
	info, err := os.Lstat(filepath.Join(rootDir, filepath.FromSlash(RuntimeRelativeDir), BundleFileName))
	return err == nil && info.Mode().IsRegular()
}

func buildProfile(bundle BundleV1, selections ProfileSelections) InstallProfileV1 {
	defaultsByApp := make(map[string]map[string]string, len(bundle.Apps))
	allowedGPUByApp := make(map[string][]string, len(bundle.Apps))
	appIDSet := make(map[string]struct{}, len(bundle.Apps)+len(selections.Apps))
	for _, app := range bundle.Apps {
		allowedGPUByApp[app.AppID] = app.AllowedGPUTypes
		if len(app.DefaultEnvs) > 0 {
			defaultsByApp[app.AppID] = app.DefaultEnvs
			appIDSet[app.AppID] = struct{}{}
		}
		if selections.DetectedGPUType != "" && containsString(app.AllowedGPUTypes, selections.DetectedGPUType) {
			appIDSet[app.AppID] = struct{}{}
		}
	}
	for appID, selection := range selections.Apps {
		if selection.SelectedGPUType != "" || len(selection.Envs) > 0 {
			appIDSet[appID] = struct{}{}
		}
	}
	appIDs := make([]string, 0, len(appIDSet))
	for appID := range appIDSet {
		appIDs = append(appIDs, appID)
	}
	sort.Strings(appIDs)
	apps := make([]InstallProfileAppV1, 0, len(appIDs))
	for _, appID := range appIDs {
		selection := selections.Apps[appID]
		selectedGPUType := selection.SelectedGPUType
		if selectedGPUType == "" && containsString(allowedGPUByApp[appID], selections.DetectedGPUType) {
			selectedGPUType = selections.DetectedGPUType
		}
		envs := cloneStringMap(defaultsByApp[appID])
		for key, value := range selection.Envs {
			if envs == nil {
				envs = make(map[string]string, len(selection.Envs))
			}
			envs[key] = value
		}
		apps = append(apps, InstallProfileAppV1{
			AppID:           appID,
			SelectedGPUType: selectedGPUType,
			Envs:            envs,
		})
	}
	return InstallProfileV1{
		SchemaVersion:   SupportedSchemaVersion,
		HardwareProfile: selections.HardwareProfile,
		Apps:            apps,
	}
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

func populateStaging(sourceRoot, stagingRoot *os.Root, bundleData []byte, bundle BundleV1, profile InstallProfileV1) error {
	if err := writeSealedRootFile(stagingRoot, BundleFileName, bundleData); err != nil {
		return err
	}
	profileData, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("encode install profile: %w", err)
	}
	if err := writeSealedRootFile(stagingRoot, ProfileFileName, append(profileData, '\n')); err != nil {
		return err
	}
	var total int64
	for _, app := range bundle.Apps {
		if err := stagingRoot.MkdirAll(path.Dir(app.Chart), 0o755); err != nil {
			return fmt.Errorf("create chart staging directory: %w", err)
		}
		copied, err := copyChart(sourceRoot, stagingRoot, app, MaxTotalChartBytes-total)
		if err != nil {
			return err
		}
		total += copied
	}
	manifests := make(map[string]struct{})
	for _, app := range bundle.Apps {
		for _, artifact := range app.Artifacts {
			if _, exists := manifests[artifact.Manifest]; exists {
				return fmt.Errorf("duplicate artifact manifest path %q", artifact.Manifest)
			}
			manifests[artifact.Manifest] = struct{}{}
			if err := stagingRoot.MkdirAll(path.Dir(artifact.Manifest), 0o755); err != nil {
				return fmt.Errorf("create artifact manifest staging directory: %w", err)
			}
			if err := copyArtifactManifest(sourceRoot, stagingRoot, artifact); err != nil {
				return err
			}
		}
	}
	return sealRootDirectories(stagingRoot)
}

func copyArtifactManifest(
	sourceRoot, stagingRoot *os.Root,
	artifact BundleArtifactV1,
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

func copyChart(sourceRoot, stagingRoot *os.Root, app BundleAppV1, totalRemaining int64) (int64, error) {
	info, err := sourceRoot.Lstat(app.Chart)
	if err != nil {
		return 0, fmt.Errorf("inspect chart %q: %w", app.Chart, err)
	}
	if info.Size() > totalRemaining {
		return 0, fmt.Errorf("total chart size exceeds %d bytes", MaxTotalChartBytes)
	}
	return copyVerifiedRegularFile(sourceRoot, stagingRoot, verifiedCopy{
		Source:     app.Chart,
		Target:     app.Chart,
		Size:       info.Size(),
		MaxSize:    min(MaxChartBytes, totalRemaining),
		SHA256:     app.ChartSHA256,
		OutputMode: 0o444,
	})
}

func writeSealedRootFile(root *os.Root, name string, data []byte) error {
	return writeRootFile(root, name, data, rootFileWrite{Mode: 0o444})
}

func sealRootDirectories(root *os.Root) error {
	var directories []string
	if err := fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("staging contains symlink %q", name)
		}
		if entry.IsDir() {
			directories = append(directories, name)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk staging directories: %w", err)
	}
	for i := len(directories) - 1; i >= 0; i-- {
		mode := os.FileMode(0o555)
		if directories[i] == "." {
			mode = materializedRootMode()
		}
		if err := root.Chmod(directories[i], mode); err != nil {
			return fmt.Errorf("make staging directory read-only: %w", err)
		}
		if err := syncRootDirectory(root, directories[i]); err != nil {
			return err
		}
	}
	return nil
}

// Darwin requires owner write permission to rename a directory; the pod's
// read-only mount remains the runtime write boundary.
func materializedRootMode() os.FileMode {
	return materializedRootModeFor(runtime.GOOS)
}

func materializedRootModeFor(goos string) os.FileMode {
	if goos == "darwin" {
		return 0o755
	}
	return 0o555
}

func renameRootEntry(parent *os.Root, oldName, newName string) error {
	for _, name := range []string{oldName, newName} {
		if err := validateSingleEntry(name); err != nil {
			return fmt.Errorf("rename %w", err)
		}
	}
	return parent.Rename(oldName, newName)
}

func replaceDirectoryRoot(parent *os.Root, target, staging string) error {
	if err := recoverPreviousRoot(parent, target); err != nil {
		return err
	}
	backup := target + ".previous"
	targetExists, err := directoryStateRoot(parent, target)
	if err != nil {
		return err
	}
	if targetExists {
		if err := renameRootEntry(parent, target, backup); err != nil {
			return fmt.Errorf("backup current preinstall directory: %w", err)
		}
		if err := syncRootDirectory(parent, "."); err != nil {
			rollbackErr := renameRootEntry(parent, backup, target)
			rollbackErr = errors.Join(rollbackErr, syncRootDirectory(parent, "."))
			return errors.Join(err, wrapRollback("restore target after backup sync failure", rollbackErr))
		}
	}
	if err := renameRootEntry(parent, staging, target); err != nil {
		var rollbackErr error
		if targetExists {
			rollbackErr = renameRootEntry(parent, backup, target)
			rollbackErr = errors.Join(rollbackErr, syncRootDirectory(parent, "."))
		}
		return errors.Join(
			fmt.Errorf("activate preinstall directory: %w", err),
			wrapRollback("restore previous target", rollbackErr),
		)
	}
	if err := syncRootDirectory(parent, "."); err != nil {
		rollbackErr := rollbackPublishedTarget(parent, target, staging, backup, targetExists)
		return errors.Join(err, rollbackErr)
	}
	if targetExists {
		if err := removeRootTree(parent, backup); err != nil {
			return fmt.Errorf("remove preinstall backup: %w", err)
		}
		if err := syncRootDirectory(parent, "."); err != nil {
			return err
		}
	}
	return nil
}

func recoverPrevious(target string) error {
	parentPath := filepath.Dir(target)
	parent, err := openDirectoryNoSymlink(parentPath)
	if err != nil {
		return fmt.Errorf("open recovery parent: %w", err)
	}
	defer parent.Close()
	return recoverPreviousRoot(parent, filepath.Base(target))
}

func recoverPreviousRoot(parent *os.Root, target string) error {
	backup := target + ".previous"
	targetExists, err := directoryStateRoot(parent, target)
	if err != nil {
		return err
	}
	backupExists, err := directoryStateRoot(parent, backup)
	if err != nil {
		return err
	}
	switch {
	case targetExists && backupExists:
		if err := removeRootTree(parent, backup); err != nil {
			return fmt.Errorf("remove stale preinstall backup: %w", err)
		}
		return syncRootDirectory(parent, ".")
	case !targetExists && backupExists:
		if err := renameRootEntry(parent, backup, target); err != nil {
			return fmt.Errorf("restore interrupted preinstall backup: %w", err)
		}
		return syncRootDirectory(parent, ".")
	default:
		return nil
	}
}

func rollbackPublishedTarget(parent *os.Root, target, staging, backup string, hadTarget bool) error {
	moveNewErr := renameRootEntry(parent, target, staging)
	if moveNewErr != nil {
		return wrapRollback("move failed published target back to staging", moveNewErr)
	}
	var restoreErr error
	if hadTarget {
		restoreErr = renameRootEntry(parent, backup, target)
	}
	syncErr := syncRootDirectory(parent, ".")
	return wrapRollback("restore target after publish sync failure", errors.Join(restoreErr, syncErr))
}

func wrapRollback(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("rollback %s: %w", operation, err)
}

func directoryStateRoot(root *os.Root, name string) (bool, error) {
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect %q: %w", name, err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("%q must be a directory", name)
	}
	return true, nil
}

func openOrCreateDirectoryNoSymlink(name string) (*os.Root, error) {
	return openDirectoryPath(name, true)
}

const stagingPrefix = ".market-preinstall-stage-"

// cleanupStagingRoots removes trusted leftovers from interrupted publishes.
func cleanupStagingRoots(parent *os.Root) error {
	entries, err := fs.ReadDir(parent.FS(), ".")
	if err != nil {
		return fmt.Errorf("list preinstall parent: %w", err)
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
