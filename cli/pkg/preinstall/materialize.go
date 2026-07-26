package preinstall

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	source := filepath.Join(installerDir, filepath.FromSlash(StaticRelativeDir))
	sourceInfo, err := os.Lstat(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect preinstall source: %w", err)
	}
	installerRoot, err := openDirectoryNoSymlink(installerDir)
	if err != nil {
		return fmt.Errorf("open installer root: %w", err)
	}
	defer installerRoot.Close()
	if err := rejectRootSymlinkComponents(installerRoot, StaticRelativeDir); err != nil {
		return err
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("preinstall source must be a directory")
	}

	sourceRoot, err := installerRoot.OpenRoot(StaticRelativeDir)
	if err != nil {
		return fmt.Errorf("open preinstall source root: %w", err)
	}
	defer sourceRoot.Close()

	bundleData, err := readRootFileLimited(sourceRoot, BundleFileName, MaxBundleJSONBytes)
	if err != nil {
		return err
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		return err
	}
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
			if err := copyArtifactManifest(sourceRoot, stagingRoot, artifact, artifactManifestCopyHooks{}); err != nil {
				return err
			}
		}
	}
	return sealRootDirectories(stagingRoot)
}

type artifactManifestCopyHooks struct {
	afterLstat func() error
	beforeCopy func() error
}

func copyArtifactManifest(
	sourceRoot, stagingRoot *os.Root,
	artifact BundleArtifactV1,
	hooks artifactManifestCopyHooks,
) error {
	if err := rejectRootSymlinkComponents(sourceRoot, artifact.Manifest); err != nil {
		return err
	}
	lstatInfo, err := sourceRoot.Lstat(artifact.Manifest)
	if err != nil {
		return fmt.Errorf("lstat artifact manifest %q: %w", artifact.Manifest, err)
	}
	if !lstatInfo.Mode().IsRegular() {
		return fmt.Errorf("artifact manifest %q must be a regular file", artifact.Manifest)
	}
	if hooks.afterLstat != nil {
		if err := hooks.afterLstat(); err != nil {
			return fmt.Errorf("after artifact manifest lstat: %w", err)
		}
	}
	input, err := sourceRoot.Open(artifact.Manifest)
	if err != nil {
		return fmt.Errorf("open artifact manifest %q: %w", artifact.Manifest, err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("fstat artifact manifest %q: %w", artifact.Manifest, err)
	}
	if !os.SameFile(lstatInfo, info) || !info.Mode().IsRegular() {
		return fmt.Errorf("artifact manifest %q changed while opening", artifact.Manifest)
	}
	if info.Size() > MaxArtifactManifestBytes {
		return fmt.Errorf("artifact manifest exceeds %d bytes: %q", MaxArtifactManifestBytes, artifact.Manifest)
	}
	if hooks.beforeCopy != nil {
		if err := hooks.beforeCopy(); err != nil {
			return fmt.Errorf("before artifact manifest copy: %w", err)
		}
	}
	output, err := stagingRoot.OpenFile(artifact.Manifest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create artifact manifest %q: %w", artifact.Manifest, err)
	}
	hasher := sha256.New()
	copied, copyErr := io.Copy(
		io.MultiWriter(output, hasher),
		io.LimitReader(input, info.Size()+1),
	)
	if copied != info.Size() {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact manifest %q changed while copying", artifact.Manifest))
	}
	afterInfo, statErr := input.Stat()
	if statErr != nil {
		copyErr = errors.Join(copyErr, fmt.Errorf("fstat artifact manifest %q after copy: %w", artifact.Manifest, statErr))
	} else if artifactManifestMetadataChanged(info, afterInfo) {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact manifest %q changed while copying", artifact.Manifest))
	}
	currentInfo, lstatErr := sourceRoot.Lstat(artifact.Manifest)
	if lstatErr != nil {
		copyErr = errors.Join(copyErr, fmt.Errorf("lstat artifact manifest %q after copy: %w", artifact.Manifest, lstatErr))
	} else if artifactManifestMetadataChanged(info, currentInfo) {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact manifest %q was replaced while copying", artifact.Manifest))
	}
	if !strings.EqualFold(hex.EncodeToString(hasher.Sum(nil)), artifact.ManifestSHA256) {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact manifest %q digest mismatch", artifact.Manifest))
	}
	sealErr := sealOpenFile(output, artifact.Manifest)
	if err := errors.Join(copyErr, sealErr); err != nil {
		return err
	}
	return nil
}

func artifactManifestMetadataChanged(before, after fs.FileInfo) bool {
	return !os.SameFile(before, after) ||
		!after.Mode().IsRegular() ||
		before.Mode() != after.Mode() ||
		before.Size() != after.Size() ||
		!before.ModTime().Equal(after.ModTime())
}

func copyChart(sourceRoot, stagingRoot *os.Root, app BundleAppV1, totalRemaining int64) (int64, error) {
	if err := rejectRootSymlinkComponents(sourceRoot, app.Chart); err != nil {
		return 0, err
	}
	input, err := openRootRegularFile(sourceRoot, app.Chart)
	if err != nil {
		return 0, err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return 0, fmt.Errorf("fstat chart %q: %w", app.Chart, err)
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("chart %q must be a regular file", app.Chart)
	}
	if info.Size() > MaxChartBytes {
		return 0, fmt.Errorf("chart exceeds %d bytes: %q", MaxChartBytes, app.Chart)
	}
	if info.Size() > totalRemaining {
		return 0, fmt.Errorf("total chart size exceeds %d bytes", MaxTotalChartBytes)
	}

	output, err := stagingRoot.OpenFile(app.Chart, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return 0, fmt.Errorf("create %q: %w", app.Chart, err)
	}
	hasher := sha256.New()
	readLimit := min(MaxChartBytes, totalRemaining) + 1
	copied, copyErr := io.Copy(io.MultiWriter(output, hasher), io.LimitReader(input, readLimit))
	if copied > MaxChartBytes {
		copyErr = errors.Join(copyErr, fmt.Errorf("chart exceeds %d bytes: %q", MaxChartBytes, app.Chart))
	}
	if copied > totalRemaining {
		copyErr = errors.Join(copyErr, fmt.Errorf("total chart size exceeds %d bytes", MaxTotalChartBytes))
	}
	if copied != info.Size() {
		copyErr = errors.Join(copyErr, fmt.Errorf("chart %q size changed while copying", app.Chart))
	}
	if !strings.EqualFold(hex.EncodeToString(hasher.Sum(nil)), app.ChartSHA256) {
		copyErr = errors.Join(copyErr, fmt.Errorf("chart %q digest mismatch", app.Chart))
	}
	sealErr := sealOpenFile(output, app.Chart)
	if err := errors.Join(copyErr, sealErr); err != nil {
		return 0, err
	}
	return copied, nil
}

func writeSealedRootFile(root *os.Root, name string, data []byte) error {
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write %q: %w", name, err)
	}
	return sealOpenFile(file, name)
}

func sealOpenFile(file *os.File, path string) error {
	chmodErr := file.Chmod(0o444)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(chmodErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("seal %q: %w", path, err)
	}
	return nil
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
		if name == "" || name == "." || name == ".." ||
			filepath.Clean(name) != name || filepath.Base(name) != name {
			return fmt.Errorf("rename path %q must be a clean single entry", name)
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

func createStagingRoot(parent *os.Root) (string, *os.Root, error) {
	for range 100 {
		random := make([]byte, 8)
		if _, err := rand.Read(random); err != nil {
			return "", nil, fmt.Errorf("generate staging name: %w", err)
		}
		name := ".market-preinstall-stage-" + hex.EncodeToString(random)
		if err := parent.Mkdir(name, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", nil, fmt.Errorf("create preinstall staging: %w", err)
		}
		root, err := parent.OpenRoot(name)
		if err != nil {
			_ = parent.Remove(name)
			return "", nil, fmt.Errorf("open preinstall staging: %w", err)
		}
		return name, root, nil
	}
	return "", nil, fmt.Errorf("create unique preinstall staging directory")
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
