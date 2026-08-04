package preinstall

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

//go:embed ensure-apps.json
var ensureAppsData []byte

const ensureStagingPrefix = ".market-ensure-stage-"

// PublishEnsureApps atomically publishes the catalog-following app declaration.
func PublishEnsureApps(rootDir string) error {
	olaresRoot, err := openOrCreateDirectoryNoSymlink(rootDir)
	if err != nil {
		return fmt.Errorf("open olares root: %w", err)
	}
	defer olaresRoot.Close()
	parentRelative := path.Dir(EnsureRuntimeRelativeDir)
	targetName := path.Base(EnsureRuntimeRelativeDir)
	if err := olaresRoot.MkdirAll(parentRelative, 0o755); err != nil {
		return fmt.Errorf("create ensure apps parent: %w", err)
	}
	if err := rejectRootSymlinkComponents(olaresRoot, parentRelative); err != nil {
		return err
	}
	parentRoot, err := openDirectoryNoSymlink(filepath.Join(rootDir, filepath.FromSlash(parentRelative)))
	if err != nil {
		return fmt.Errorf("open ensure apps parent: %w", err)
	}
	defer parentRoot.Close()
	if _, err := directoryStateRoot(parentRoot, targetName); err != nil {
		return err
	}
	stagingName, _, stagingRoot, _, err := createTrustedStaging(parentRoot, ensureStagingPrefix, 8, 0o700)
	if err != nil {
		return err
	}
	defer func() {
		if stagingRoot != nil {
			_ = stagingRoot.Close()
		}
		_ = removeRootTree(parentRoot, stagingName)
	}()
	if err := writeSealedRootFile(stagingRoot, EnsureAppsFileName, ensureAppsData); err != nil {
		return err
	}
	if err := sealRootDirectories(stagingRoot); err != nil {
		return err
	}
	if err := stagingRoot.Close(); err != nil {
		return fmt.Errorf("close ensure apps staging root: %w", err)
	}
	stagingRoot = nil
	return replaceDirectoryRoot(parentRoot, targetName, stagingName)
}

// EnsureAppsPublished reports whether the declaration is available to Market.
func EnsureAppsPublished(rootDir string) bool {
	info, err := os.Lstat(filepath.Join(rootDir, filepath.FromSlash(EnsureRuntimeRelativeDir), EnsureAppsFileName))
	return err == nil && info.Mode().IsRegular()
}
