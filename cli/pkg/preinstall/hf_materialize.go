package preinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beclab/Olares/cli/pkg/core/logger"
)

const (
	hfCacheMarkerFileName = ".olares-hf-cache.json"
	hfStageMarkerFileName = ".olares-hf-stage.json"
)

type hfCacheMarker struct {
	Kind           string `json:"kind"`
	Repo           string `json:"repo"`
	Revision       string `json:"revision"`
	ManifestSHA256 string `json:"manifestSha256"`
	TotalSize      int64  `json:"totalSize"`
}

type hfStageMarker struct {
	Target string `json:"target"`
	Repo   string `json:"repo"`
	Token  string `json:"token"`
}

type ownershipFunc func(*os.Root) error

type fileChownFunc func(*os.File, int, int) error

type hfOwnership struct {
	lockRoot  func(*os.Root) (func() error, error)
	tree      ownershipFunc
	marker    func(*os.File) error
	trustedID uint32
}

type hfMaterializeHooks struct {
	beforeRename        func(*os.Root, string) error
	afterStageCreated   func(*os.Root, string, *os.Root) error
	trustedStageUID     *uint32
	syncStageDirectory  func(*os.Root) error
	syncParentDirectory func(*os.Root) error
}

func markerFor(artifact BundleArtifactV1) hfCacheMarker {
	return hfCacheMarker{
		Kind:           artifact.Kind,
		Repo:           artifact.Repo,
		Revision:       strings.ToLower(artifact.Revision),
		ManifestSHA256: strings.ToLower(artifact.ManifestSHA256),
		TotalSize:      artifact.TotalSize,
	}
}

func materializeHFArtifacts(installerDir, targetRootPath string, ownership *hfOwnership) error {
	return materializeHFArtifactsWithHooks(installerDir, targetRootPath, ownership, hfMaterializeHooks{})
}

func materializeHFArtifactsWithHooks(installerDir, targetRootPath string, ownership *hfOwnership, hooks hfMaterializeHooks) (retErr error) {
	if err := healHFCacheRootMode(targetRootPath); err != nil {
		return err
	}
	staticRoot, _, bundle, found, err := openStaticBundle(installerDir)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	defer staticRoot.Close()
	artifacts := bundleHFArtifacts(bundle)
	if len(artifacts) == 0 {
		return nil
	}
	type hfArtifactTarget struct {
		artifact BundleArtifactV1
		target   string
	}
	declarations := make([]hfArtifactTarget, 0, len(artifacts))
	targets := make(map[string]string, len(artifacts))
	for _, artifact := range artifacts {
		target, err := hfTargetName(artifact.Repo)
		if err != nil {
			return err
		}
		if previous, exists := targets[target]; exists {
			return fmt.Errorf("duplicate Hugging Face cache target %q for repositories %q and %q", target, previous, artifact.Repo)
		}
		targets[target] = artifact.Repo
		declarations = append(declarations, hfArtifactTarget{artifact: artifact, target: target})
	}

	targetRoot, err := openDirectoryNoSymlink(targetRootPath)
	if err != nil {
		return fmt.Errorf("open Hugging Face cache root: %w", err)
	}
	defer targetRoot.Close()
	if ownership != nil && ownership.lockRoot != nil {
		restore, err := ownership.lockRoot(targetRoot)
		if err != nil {
			return err
		}
		defer func() {
			retErr = errors.Join(retErr, restore())
		}()
	}
	trustedStageUID := uint32(os.Geteuid())
	if ownership != nil {
		trustedStageUID = ownership.trustedID
	}
	if hooks.trustedStageUID != nil {
		trustedStageUID = *hooks.trustedStageUID
	}

	for _, declaration := range declarations {
		artifact := declaration.artifact
		manifest, err := LoadArtifactManifest(staticRoot, artifact)
		if err != nil {
			return fmt.Errorf("load artifact %q manifest: %w", artifact.Repo, err)
		}
		if err := materializeOneHFArtifact(staticRoot, targetRoot, declaration.target, artifact, manifest, ownership, trustedStageUID, hooks); err != nil {
			return fmt.Errorf("materialize Hugging Face artifact %q: %w", artifact.Repo, err)
		}
	}
	return nil
}

func bundleHFArtifacts(bundle BundleV1) []BundleArtifactV1 {
	var artifacts []BundleArtifactV1
	for _, app := range bundle.Apps {
		for _, artifact := range app.Artifacts {
			if artifact.Kind == ArtifactKindHFCache {
				artifacts = append(artifacts, artifact)
			}
		}
	}
	return artifacts
}

func hfTargetName(repo string) (string, error) {
	if !repoPattern.MatchString(repo) {
		return "", fmt.Errorf("repo %q must be an owner/repo identifier", repo)
	}
	owner, name, ok := strings.Cut(repo, "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") {
		return "", fmt.Errorf("repo %q must contain exactly one slash", repo)
	}
	target := "models--" + owner + "--" + name
	if filepath.Base(target) != target || target == "." || target == ".." {
		return "", fmt.Errorf("repo %q produces an unsafe cache target", repo)
	}
	return target, nil
}

func materializeOneHFArtifact(staticRoot, targetRoot *os.Root, target string, artifact BundleArtifactV1, manifest ArtifactManifestV1, ownership *hfOwnership, trustedStageUID uint32, hooks hfMaterializeHooks) (retErr error) {
	if err := cleanupHFStaging(targetRoot, target, artifact.Repo, trustedStageUID); err != nil {
		return err
	}
	exists, err := targetRoot.Lstat(target)
	if err == nil {
		if !exists.IsDir() || exists.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("existing target %q is not a directory", target)
		}
		matches, err := hfMarkerMatches(targetRoot, target, markerFor(artifact))
		if err != nil {
			return fmt.Errorf("inspect existing target %q: %w", target, err)
		}
		if matches {
			return nil
		}
		// A tree that still carries this artifact's staging marker and no
		// completion marker is a publish this program was interrupted in: the
		// rename landed, the completion marker did not. Redoing it is the only
		// way the host recovers without a human clearing the directory.
		interrupted, err := hfInterruptedPublish(targetRoot, target, artifact.Repo)
		if err != nil {
			return fmt.Errorf("inspect existing target %q: %w", target, err)
		}
		if !interrupted {
			return fmt.Errorf("existing target %q marker is missing or different", target)
		}
		if err := targetRoot.RemoveAll(target); err != nil {
			return fmt.Errorf("remove interrupted publish %q: %w", target, err)
		}
		if err := syncRootDirectory(targetRoot, "."); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect target %q: %w", target, err)
	}

	syncStageDirectory := hooks.syncStageDirectory
	if syncStageDirectory == nil {
		syncStageDirectory = func(root *os.Root) error {
			return syncRootDirectory(root, ".")
		}
	}
	syncParentDirectory := hooks.syncParentDirectory
	if syncParentDirectory == nil {
		syncParentDirectory = func(root *os.Root) error {
			return syncRootDirectory(root, ".")
		}
	}
	staging, stagingRoot, stagingInfo, err := createHFStaging(
		targetRoot,
		target,
		artifact.Repo,
		syncStageDirectory,
		syncParentDirectory,
	)
	if err != nil {
		return err
	}
	defer func() {
		if stagingRoot != nil {
			_ = stagingRoot.Close()
		}
		retErr = errors.Join(retErr, removeKnownHFStaging(targetRoot, staging, stagingInfo))
	}()
	if hooks.afterStageCreated != nil {
		if err := hooks.afterStageCreated(targetRoot, staging, stagingRoot); err != nil {
			return fmt.Errorf("after Hugging Face staging creation: %w", err)
		}
	}

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

	for _, entry := range manifest.Entries {
		if err := materializeHFEntry(sourceRoot, stagingRoot, entry); err != nil {
			return err
		}
	}
	if err := syncHFTree(stagingRoot, 0o700); err != nil {
		return err
	}
	if err := stagingRoot.Close(); err != nil {
		return fmt.Errorf("close Hugging Face staging: %w", err)
	}
	stagingRoot = nil
	if _, err := targetRoot.Lstat(target); err == nil {
		return fmt.Errorf("existing target %q appeared during materialization", target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("recheck target %q: %w", target, err)
	}
	if hooks.beforeRename != nil {
		if err := hooks.beforeRename(targetRoot, target); err != nil {
			return fmt.Errorf("before Hugging Face rename: %w", err)
		}
	}
	parentFile, err := targetRoot.Open(".")
	if err != nil {
		return fmt.Errorf("open Hugging Face publish directory: %w", err)
	}
	publishErr := renameNoReplace(parentFile, staging, target)
	closeErr := parentFile.Close()
	if err := errors.Join(publishErr, closeErr); err != nil {
		return fmt.Errorf("publish Hugging Face cache %q: %w", target, err)
	}
	if err := syncRootDirectory(targetRoot, "."); err != nil {
		return err
	}
	targetInfo, err := targetRoot.Lstat(target)
	if err != nil {
		return fmt.Errorf("inspect published Hugging Face cache %q: %w", target, err)
	}
	if !targetInfo.IsDir() || !os.SameFile(stagingInfo, targetInfo) {
		return fmt.Errorf("published Hugging Face cache %q changed identity", target)
	}
	publishedRoot, err := targetRoot.OpenRoot(target)
	if err != nil {
		return fmt.Errorf("open published Hugging Face cache %q: %w", target, err)
	}
	if err := syncHFTree(publishedRoot, 0o755); err != nil {
		_ = publishedRoot.Close()
		return err
	}
	if ownership != nil && ownership.tree != nil {
		if err := ownership.tree(publishedRoot); err != nil {
			_ = publishedRoot.Close()
			return err
		}
	}
	markerData, err := json.Marshal(markerFor(artifact))
	if err == nil {
		err = writeHFCompletionMarker(publishedRoot, append(markerData, '\n'), ownership)
	}
	// The staging marker is what tells the next run this tree is a publish
	// that was interrupted, so it is dropped only once the completion marker
	// says the publish finished. A crash before this point is recoverable; a
	// crash after it leaves a complete tree.
	if err == nil {
		err = removePublishedHFStageMarker(publishedRoot)
	}
	closeErr = publishedRoot.Close()
	if err := errors.Join(err, closeErr); err != nil {
		return fmt.Errorf("complete published Hugging Face cache %q: %w", target, err)
	}
	return nil
}

func removePublishedHFStageMarker(publishedRoot *os.Root) error {
	if err := publishedRoot.Remove(hfStageMarkerFileName); err != nil {
		return fmt.Errorf("remove published staging marker: %w", err)
	}
	if err := syncRootDirectory(publishedRoot, "."); err != nil {
		return err
	}
	return nil
}

// hfInterruptedPublish reports whether the published tree is one this program
// renamed into place but never finished. The staging marker survives the
// rename and is removed only after the completion marker lands, so finding it
// beside a missing completion marker names exactly that window.
func hfInterruptedPublish(root *os.Root, target, repo string) (bool, error) {
	publishedRoot, err := root.OpenRoot(target)
	if err != nil {
		return false, err
	}
	defer publishedRoot.Close()
	var got hfStageMarker
	ok, err := readMarker(publishedRoot, hfStageMarkerFileName, &got)
	if err != nil || !ok {
		return false, err
	}
	return got.Target == target && got.Repo == repo, nil
}

func materializeHFEntry(sourceRoot, stagingRoot *os.Root, entry ArtifactManifestEntryV1) error {
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
		if err := stagingRoot.MkdirAll(entry.Path, 0o755); err != nil {
			return fmt.Errorf("create artifact directory %q: %w", entry.Path, err)
		}
		return stagingRoot.Chmod(entry.Path, 0o755)
	case "file":
		if !info.Mode().IsRegular() {
			return fmt.Errorf("artifact entry %q must be a regular file", entry.Path)
		}
		if err := stagingRoot.MkdirAll(path.Dir(entry.Path), 0o755); err != nil {
			return fmt.Errorf("create artifact file parent %q: %w", entry.Path, err)
		}
		return copyHFFile(sourceRoot, stagingRoot, entry, info)
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
		if err := stagingRoot.MkdirAll(path.Dir(entry.Path), 0o755); err != nil {
			return fmt.Errorf("create artifact symlink parent %q: %w", entry.Path, err)
		}
		if err := stagingRoot.Symlink(target, entry.Path); err != nil {
			return fmt.Errorf("create artifact symlink %q: %w", entry.Path, err)
		}
		return nil
	default:
		return fmt.Errorf("artifact entry %q has unsupported type %q", entry.Path, entry.Type)
	}
}

func copyHFFile(sourceRoot, stagingRoot *os.Root, entry ArtifactManifestEntryV1, lstatInfo fs.FileInfo) error {
	if lstatInfo.Size() != entry.Size {
		return fmt.Errorf("artifact file %q size mismatch: got %d, want %d", entry.Path, lstatInfo.Size(), entry.Size)
	}
	_, err := copyVerifiedRegularFile(sourceRoot, stagingRoot, verifiedCopy{
		Source:     entry.Path,
		Target:     entry.Path,
		Size:       entry.Size,
		MaxSize:    entry.Size,
		SHA256:     entry.SHA256,
		OutputMode: 0o644,
	})
	return err
}

func rejectHFReservedTarget(name string) error {
	if name == hfCacheMarkerFileName || name == hfStageMarkerFileName {
		return fmt.Errorf("artifact file %q uses a reserved Hugging Face marker name", name)
	}
	return nil
}

func writeHFFile(root *os.Root, name string, data []byte) error {
	return writeRootFile(root, name, data, rootFileWrite{Mode: 0o644})
}

func writeHFCompletionMarker(root *os.Root, data []byte, ownership *hfOwnership) error {
	var beforeSeal func(*os.File) error
	if ownership != nil {
		beforeSeal = ownership.marker
	}
	return writeRootFile(root, hfCacheMarkerFileName, data, rootFileWrite{
		Mode:          0o644,
		BeforeSeal:    beforeSeal,
		SyncDirectory: true,
		RemoveOnError: true,
	})
}

func hfMarkerMatches(root *os.Root, target string, want hfCacheMarker) (bool, error) {
	targetRoot, err := root.OpenRoot(target)
	if err != nil {
		return false, err
	}
	defer targetRoot.Close()
	var got hfCacheMarker
	ok, err := readMarker(targetRoot, hfCacheMarkerFileName, &got)
	if err != nil || !ok {
		return false, err
	}
	return got == want, nil
}

func hfStagingPrefix(target string) string {
	return "." + target + ".olares-hf-stage-"
}

func createHFStaging(
	root *os.Root,
	target, repo string,
	syncDirectory, syncParentDirectory func(*os.Root) error,
) (string, *os.Root, fs.FileInfo, error) {
	name, token, stagingRoot, info, err := createTrustedStaging(root, hfStagingPrefix(target), 16, 0o700)
	if err != nil {
		return "", nil, nil, fmt.Errorf("create Hugging Face staging: %w", err)
	}
	fail := func(err error) (string, *os.Root, fs.FileInfo, error) {
		_ = stagingRoot.Close()
		_ = root.RemoveAll(name)
		return "", nil, nil, err
	}
	markerData, err := json.Marshal(hfStageMarker{Target: target, Repo: repo, Token: token})
	if err == nil {
		err = writeHFFile(stagingRoot, hfStageMarkerFileName, append(markerData, '\n'))
	}
	if err != nil {
		return fail(fmt.Errorf("write Hugging Face staging marker: %w", err))
	}
	if err := syncDirectory(stagingRoot); err != nil {
		return fail(fmt.Errorf("sync Hugging Face staging marker: %w", err))
	}
	if err := syncParentDirectory(root); err != nil {
		return fail(fmt.Errorf("sync Hugging Face cache root after staging creation: %w", err))
	}
	return name, stagingRoot, info, nil
}

func cleanupHFStaging(root *os.Root, target, repo string, trustedUID uint32) error {
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return fmt.Errorf("list Hugging Face cache root: %w", err)
	}
	prefix := hfStagingPrefix(target)
	for _, entry := range entries {
		token, ok := hfStagingToken(entry.Name(), prefix)
		if !ok {
			continue
		}
		stageRoot, info, trusted, err := openTrustedStaging(root, entry.Name(), trustedUID, 0o700)
		if err != nil {
			return fmt.Errorf("open Hugging Face staging %q: %w", entry.Name(), err)
		}
		if !trusted {
			warnf("leaving untrusted Hugging Face staging %q alone", entry.Name())
			continue
		}
		matches := hfStageMarkerMatches(stageRoot, hfStageMarker{Target: target, Repo: repo, Token: token})
		closeErr := stageRoot.Close()
		if closeErr != nil {
			return fmt.Errorf("close Hugging Face staging %q: %w", entry.Name(), closeErr)
		}
		if !matches {
			continue
		}
		if err := removeKnownHFStaging(root, entry.Name(), info); err != nil {
			return fmt.Errorf("remove owned Hugging Face staging %q: %w", entry.Name(), err)
		}
	}
	return nil
}

func hfStagingToken(name, prefix string) (string, bool) {
	if !strings.HasPrefix(name, prefix) {
		return "", false
	}
	token := strings.TrimPrefix(name, prefix)
	if len(token) != 32 {
		return "", false
	}
	for _, character := range token {
		if !strings.ContainsRune("0123456789abcdef", character) {
			return "", false
		}
	}
	return token, true
}

func hfStageMarkerMatches(root *os.Root, want hfStageMarker) bool {
	var got hfStageMarker
	ok, err := readMarker(root, hfStageMarkerFileName, &got)
	return err == nil && ok && got == want
}

func removeKnownHFStaging(root *os.Root, name string, expected fs.FileInfo) error {
	info, err := root.Lstat(name)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect Hugging Face staging %q: %w", name, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("known Hugging Face staging %q is no longer a directory", name)
	}
	if !os.SameFile(info, expected) {
		return fmt.Errorf("known Hugging Face staging %q was replaced", name)
	}
	if err := root.RemoveAll(name); err != nil {
		return fmt.Errorf("remove Hugging Face staging %q: %w", name, err)
	}
	return nil
}

func setHFOwnership(root *os.Root, chown fileChownFunc) error {
	var files, directories []string
	if err := fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			// Relative cache symlinks do not depend on inode ownership.
			return nil
		}
		if entry.IsDir() {
			directories = append(directories, name)
		} else {
			files = append(files, name)
		}
		return nil
	}); err != nil {
		return err
	}
	for _, name := range files {
		if err := chownHFEntry(root, name, false, chown); err != nil {
			return err
		}
	}
	for i := len(directories) - 1; i >= 0; i-- {
		if err := chownHFEntry(root, directories[i], true, chown); err != nil {
			return err
		}
	}
	return nil
}

func chownHFEntry(root *os.Root, name string, wantDirectory bool, chown fileChownFunc) error {
	info, err := root.Lstat(name)
	if err != nil {
		return fmt.Errorf("inspect ownership entry %q: %w", name, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	if info.IsDir() != wantDirectory || (!wantDirectory && !info.Mode().IsRegular()) {
		return fmt.Errorf("ownership entry %q changed type", name)
	}
	file, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("open ownership entry %q: %w", name, err)
	}
	openedInfo, statErr := file.Stat()
	if statErr == nil && (!os.SameFile(info, openedInfo) ||
		(info.IsDir() != openedInfo.IsDir()) ||
		(info.Mode().IsRegular() != openedInfo.Mode().IsRegular())) {
		statErr = fmt.Errorf("ownership entry changed while opening")
	}
	chownErr := error(nil)
	syncErr := error(nil)
	if statErr == nil {
		chownErr = chown(file, 1000, 1000)
		if chownErr == nil {
			syncErr = file.Sync()
		}
	}
	closeErr := file.Close()
	if err := errors.Join(statErr, chownErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("set ownership for %q: %w", name, err)
	}
	return nil
}

func secureHFMarkerOwnership(file *os.File) error {
	return file.Chown(1000, 1000)
}

type hfRootLockOps struct {
	chmod func(*os.File, os.FileMode) error
	sync  func(*os.File) error
}

// lockHFCacheRoot never chowns the root (normally owned by uid 1000, see
// storage.CreateAppCommonDir): a crash mid-chown could strand it under root
// with no unprivileged process able to reclaim it. It only chmods away write
// bits, so a crash mid-lock leaves it self-healable by uid 1000.
func lockHFCacheRoot(root *os.Root) (func() error, error) {
	return lockHFCacheRootWithOps(root, hfRootLockOps{
		chmod: func(file *os.File, mode os.FileMode) error {
			return file.Chmod(mode)
		},
		sync: func(file *os.File) error {
			return file.Sync()
		},
	})
}

// healHFCacheRootMode restores a stale 0o555 lock left by a crash between
// lockHFCacheRoot's chmod and its deferred restore. It runs unconditionally
// (even with no HF artifact in the bundle) and only touches mode, never
// ownership, and completes before this run's own lock cycle begins.
func healHFCacheRootMode(targetRootPath string) error {
	root, err := openDirectoryNoSymlink(targetRootPath)
	if err != nil {
		// os.IsNotExist doesn't unwrap arbitrary %w chains; errors.Is does.
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("open Hugging Face cache root for self-heal: %w", err)
	}
	defer root.Close()
	directory, err := root.Open(".")
	if err != nil {
		return fmt.Errorf("open Hugging Face cache root handle for self-heal: %w", err)
	}
	defer directory.Close()
	info, err := directory.Stat()
	if err != nil {
		return fmt.Errorf("inspect Hugging Face cache root mode: %w", err)
	}
	if info.Mode().Perm() != 0o555 {
		return nil
	}
	if err := directory.Chmod(0o755); err != nil {
		return fmt.Errorf("self-heal Hugging Face cache root mode: %w", err)
	}
	return directory.Sync()
}

func lockHFCacheRootWithOps(root *os.Root, ops hfRootLockOps) (func() error, error) {
	directory, err := root.Open(".")
	if err != nil {
		return nil, fmt.Errorf("open Hugging Face cache root for locking: %w", err)
	}
	restore := func() error {
		err := errors.Join(
			ops.chmod(directory, 0o755),
			ops.sync(directory),
			directory.Close(),
		)
		if err != nil {
			return fmt.Errorf("restore Hugging Face cache root permissions: %w", err)
		}
		return nil
	}
	lockSteps := []struct {
		name string
		run  func() error
	}{
		{name: "deny writes for every principal", run: func() error { return ops.chmod(directory, 0o555) }},
		{name: "sync metadata", run: func() error { return ops.sync(directory) }},
	}
	for _, step := range lockSteps {
		if err := step.run(); err != nil {
			restoreErr := restore()
			return nil, fmt.Errorf("lock Hugging Face cache root during %s: %w", step.name, errors.Join(err, restoreErr))
		}
	}
	return restore, nil
}

func secureHFOwnership(root *os.Root) error {
	return setHFOwnership(root, func(file *os.File, uid, gid int) error {
		return file.Chown(uid, gid)
	})
}

func productionHFOwnership() *hfOwnership {
	return &hfOwnership{
		lockRoot: lockHFCacheRoot,
		tree:     secureHFOwnership,
		marker:   secureHFMarkerOwnership,
		// Staging is created by this process, so the only owner it can trust
		// its own leftovers to carry is the one it runs as. Hard-coding root
		// makes an installer running as anyone else refuse to clean up after
		// its own interrupted run.
		trustedID: uint32(os.Geteuid()),
	}
}

// warnf logs through the CLI logger when one has been installed. This package
// also runs under tests, which never call logger.InitLog.
func warnf(format string, args ...any) {
	if log := logger.GetLogger(); log != nil {
		log.Warnf(format, args...)
	}
}

func syncHFTree(root *os.Root, rootMode os.FileMode) error {
	var directories []string
	if err := fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			directories = append(directories, name)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk Hugging Face staging: %w", err)
	}
	for i := len(directories) - 1; i >= 0; i-- {
		mode := os.FileMode(0o755)
		if directories[i] == "." {
			mode = rootMode
		}
		if err := root.Chmod(directories[i], mode); err != nil {
			return fmt.Errorf("set Hugging Face directory mode: %w", err)
		}
		if err := syncRootDirectory(root, directories[i]); err != nil {
			return err
		}
	}
	return nil
}
