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
	"strings"
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
	staticPath := filepath.Join(installerDir, filepath.FromSlash(StaticRelativeDir))
	if _, err := os.Lstat(staticPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("inspect preinstall source: %w", err)
	}
	staticRoot, err := openDirectoryNoSymlink(staticPath)
	if err != nil {
		return fmt.Errorf("open preinstall source: %w", err)
	}
	defer staticRoot.Close()

	bundleData, err := readRootFileLimited(staticRoot, BundleFileName, MaxBundleJSONBytes)
	if err != nil {
		return err
	}
	bundle, err := DecodeBundle(bundleData)
	if err != nil {
		return err
	}
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
		return fmt.Errorf("existing target %q marker is missing or different", target)
	}
	if !os.IsNotExist(err) {
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
	if err := publishedRoot.Remove(hfStageMarkerFileName); err != nil {
		_ = publishedRoot.Close()
		return fmt.Errorf("remove published staging marker: %w", err)
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
	closeErr = publishedRoot.Close()
	if err := errors.Join(err, closeErr); err != nil {
		return fmt.Errorf("complete published Hugging Face cache %q: %w", target, err)
	}
	return nil
}

func materializeHFEntry(sourceRoot, stagingRoot *os.Root, entry ArtifactManifestEntryV1) error {
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
	input, err := sourceRoot.Open(entry.Path)
	if err != nil {
		return fmt.Errorf("open artifact file %q: %w", entry.Path, err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("fstat artifact file %q: %w", entry.Path, err)
	}
	if !info.Mode().IsRegular() || !os.SameFile(lstatInfo, info) {
		return fmt.Errorf("artifact entry %q changed or is not a regular file", entry.Path)
	}
	output, err := stagingRoot.OpenFile(entry.Path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create artifact file %q: %w", entry.Path, err)
	}
	hasher := sha256.New()
	copied, copyErr := io.Copy(io.MultiWriter(output, hasher), io.LimitReader(input, entry.Size+1))
	if copied != entry.Size {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact file %q size mismatch: got %d, want %d", entry.Path, copied, entry.Size))
	}
	if got := hex.EncodeToString(hasher.Sum(nil)); !strings.EqualFold(got, entry.SHA256) {
		copyErr = errors.Join(copyErr, fmt.Errorf("artifact file %q digest mismatch", entry.Path))
	}
	chmodErr := output.Chmod(0o644)
	syncErr := output.Sync()
	closeErr := output.Close()
	return errors.Join(copyErr, chmodErr, syncErr, closeErr)
}

func writeHFFile(root *os.Root, name string, data []byte) error {
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	_, writeErr := file.Write(data)
	chmodErr := file.Chmod(0o644)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, chmodErr, syncErr, closeErr); err != nil {
		return fmt.Errorf("write %q: %w", name, err)
	}
	return nil
}

func writeHFCompletionMarker(root *os.Root, data []byte, ownership *hfOwnership) error {
	file, err := root.OpenFile(hfCacheMarkerFileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create completion marker: %w", err)
	}
	_, writeErr := file.Write(data)
	var ownershipErr error
	if ownership != nil && ownership.marker != nil && writeErr == nil {
		ownershipErr = ownership.marker(file)
	}
	chmodErr := file.Chmod(0o644)
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(writeErr, ownershipErr, chmodErr, syncErr, closeErr); err != nil {
		removeErr := root.Remove(hfCacheMarkerFileName)
		return fmt.Errorf("write completion marker: %w", errors.Join(err, removeErr))
	}
	if err := syncRootDirectory(root, "."); err != nil {
		removeErr := root.Remove(hfCacheMarkerFileName)
		syncErr := syncRootDirectory(root, ".")
		return fmt.Errorf("sync completion marker: %w", errors.Join(err, removeErr, syncErr))
	}
	return nil
}

func hfMarkerMatches(root *os.Root, target string, want hfCacheMarker) (bool, error) {
	targetRoot, err := root.OpenRoot(target)
	if err != nil {
		return false, err
	}
	defer targetRoot.Close()
	data, err := readRootFileLimited(targetRoot, hfCacheMarkerFileName, 4096)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var got hfCacheMarker
	if err := strictDecode(data, &got); err != nil {
		return false, nil
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
	for range 100 {
		random := make([]byte, 16)
		if _, err := rand.Read(random); err != nil {
			return "", nil, nil, fmt.Errorf("generate Hugging Face staging name: %w", err)
		}
		token := hex.EncodeToString(random)
		name := hfStagingPrefix(target) + token
		if err := root.Mkdir(name, 0o700); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", nil, nil, fmt.Errorf("create Hugging Face staging: %w", err)
		}
		stagingRoot, err := root.OpenRoot(name)
		if err != nil {
			_ = root.Remove(name)
			return "", nil, nil, fmt.Errorf("open Hugging Face staging: %w", err)
		}
		info, err := stagingRoot.Stat(".")
		if err != nil {
			_ = stagingRoot.Close()
			_ = root.RemoveAll(name)
			return "", nil, nil, fmt.Errorf("stat Hugging Face staging: %w", err)
		}
		markerData, err := json.Marshal(hfStageMarker{Target: target, Repo: repo, Token: token})
		if err == nil {
			err = writeHFFile(stagingRoot, hfStageMarkerFileName, append(markerData, '\n'))
		}
		if err != nil {
			_ = stagingRoot.Close()
			_ = root.RemoveAll(name)
			return "", nil, nil, fmt.Errorf("write Hugging Face staging marker: %w", err)
		}
		if err := syncDirectory(stagingRoot); err != nil {
			_ = stagingRoot.Close()
			_ = root.RemoveAll(name)
			return "", nil, nil, fmt.Errorf("sync Hugging Face staging marker: %w", err)
		}
		if err := syncParentDirectory(root); err != nil {
			_ = stagingRoot.Close()
			_ = root.RemoveAll(name)
			return "", nil, nil, fmt.Errorf("sync Hugging Face cache root after staging creation: %w", err)
		}
		return name, stagingRoot, info, nil
	}
	return "", nil, nil, fmt.Errorf("create unique Hugging Face staging directory")
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
		info, err := root.Lstat(entry.Name())
		if err != nil {
			return fmt.Errorf("inspect Hugging Face staging %q: %w", entry.Name(), err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			continue
		}
		uid, ok := fileOwnerUID(info)
		if !ok || uid != trustedUID {
			return fmt.Errorf("untrusted staging owner for %q", entry.Name())
		}
		if info.Mode().Perm() != 0o700 {
			return fmt.Errorf("untrusted staging mode for %q", entry.Name())
		}
		stageRoot, err := root.OpenRoot(entry.Name())
		if err != nil {
			return fmt.Errorf("open Hugging Face staging %q: %w", entry.Name(), err)
		}
		openedInfo, statErr := stageRoot.Stat(".")
		if statErr != nil || !os.SameFile(info, openedInfo) {
			_ = stageRoot.Close()
			return errors.Join(fmt.Errorf("Hugging Face staging %q changed while opening", entry.Name()), statErr)
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
	data, err := readRootFileLimited(root, hfStageMarkerFileName, 4096)
	if err != nil {
		return false
	}
	var got hfStageMarker
	return strictDecode(data, &got) == nil && got == want
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
		lockRoot:  lockHFCacheRoot,
		tree:      secureHFOwnership,
		marker:    secureHFMarkerOwnership,
		trustedID: 0,
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
