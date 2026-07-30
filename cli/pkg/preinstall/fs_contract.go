package preinstall

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type verifiedCopy struct {
	Source, Target string
	Size, MaxSize  int64
	SHA256         string
	OutputMode     os.FileMode
	AfterLstat     func() error
	BeforeCopy     func() error
}

type rootFileWrite struct {
	Mode          os.FileMode
	BeforeSeal    func(*os.File) error
	SyncDirectory bool
	RemoveOnError bool
}

func createTrustedStaging(parent *os.Root, prefix string, tokenBytes int, mode os.FileMode) (string, string, *os.Root, os.FileInfo, error) {
	for range 100 {
		random := make([]byte, tokenBytes)
		if _, err := rand.Read(random); err != nil {
			return "", "", nil, nil, fmt.Errorf("generate staging name: %w", err)
		}
		token := hex.EncodeToString(random)
		name := prefix + token
		if err := parent.Mkdir(name, mode); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", nil, nil, fmt.Errorf("create staging %q: %w", name, err)
		}
		root, info, trusted, err := openTrustedStaging(parent, name, uint32(os.Geteuid()), mode)
		if err != nil || !trusted {
			_ = parent.RemoveAll(name)
			return "", "", nil, nil, errors.Join(fmt.Errorf("open created staging %q", name), err)
		}
		return name, token, root, info, nil
	}
	return "", "", nil, nil, fmt.Errorf("create unique staging directory")
}

func openTrustedStaging(parent *os.Root, name string, trustedUID uint32, modes ...os.FileMode) (*os.Root, os.FileInfo, bool, error) {
	info, err := parent.Lstat(name)
	if err != nil {
		return nil, nil, false, err
	}
	if !trustedStagingInfo(info, trustedUID, modes, fileOwnerUID) {
		return nil, info, false, nil
	}
	root, err := parent.OpenRoot(name)
	if err != nil {
		return nil, info, false, err
	}
	openedInfo, err := root.Stat(".")
	if err != nil || !os.SameFile(info, openedInfo) {
		_ = root.Close()
		return nil, info, false, errors.Join(fmt.Errorf("staging %q changed while opening", name), err)
	}
	return root, info, true, nil
}

func trustedStagingInfo(info os.FileInfo, trustedUID uint32, modes []os.FileMode, owner func(os.FileInfo) (uint32, bool)) bool {
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return false
	}
	uid, known := owner(info)
	if !known || uid != trustedUID {
		return false
	}
	for _, mode := range modes {
		if info.Mode().Perm() == mode {
			return true
		}
	}
	return false
}

func writeRootFile(root *os.Root, name string, data []byte, options rootFileWrite) error {
	file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, options.Mode)
	if err != nil {
		return fmt.Errorf("create %q: %w", name, err)
	}
	_, writeErr := file.Write(data)
	var beforeSealErr error
	if writeErr == nil && options.BeforeSeal != nil {
		beforeSealErr = options.BeforeSeal(file)
	}
	err = errors.Join(writeErr, beforeSealErr, sealFile(file, name, options.Mode))
	if err == nil && options.SyncDirectory {
		err = syncRootDirectory(root, ".")
	}
	if err != nil && options.RemoveOnError {
		err = errors.Join(err, root.Remove(name), syncRootDirectory(root, "."))
	}
	if err != nil {
		return fmt.Errorf("write %q: %w", name, err)
	}
	return nil
}

func copyVerifiedRegularFile(sourceRoot, targetRoot *os.Root, spec verifiedCopy) (int64, error) {
	if err := rejectRootSymlinkComponents(sourceRoot, spec.Source); err != nil {
		return 0, err
	}
	lstatInfo, err := sourceRoot.Lstat(spec.Source)
	if err != nil {
		return 0, fmt.Errorf("inspect %q: %w", spec.Source, err)
	}
	if !lstatInfo.Mode().IsRegular() {
		return 0, fmt.Errorf("%q must be a regular file", spec.Source)
	}
	if hasMultipleLinks(lstatInfo) {
		return 0, fmt.Errorf("%q must not be a hardlink", spec.Source)
	}
	if spec.Size < 0 || spec.Size > spec.MaxSize {
		return 0, fmt.Errorf("%q exceeds %d bytes", spec.Source, spec.MaxSize)
	}
	if spec.AfterLstat != nil {
		if err := spec.AfterLstat(); err != nil {
			return 0, err
		}
	}
	input, err := sourceRoot.Open(spec.Source)
	if err != nil {
		return 0, fmt.Errorf("open %q: %w", spec.Source, err)
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return 0, fmt.Errorf("fstat %q: %w", spec.Source, err)
	}
	if !os.SameFile(lstatInfo, info) || !info.Mode().IsRegular() || hasMultipleLinks(info) {
		return 0, fmt.Errorf("%q changed while opening", spec.Source)
	}
	if info.Size() != spec.Size {
		return 0, fmt.Errorf("%q size mismatch: got %d, want %d", spec.Source, info.Size(), spec.Size)
	}
	if spec.BeforeCopy != nil {
		if err := spec.BeforeCopy(); err != nil {
			return 0, err
		}
	}
	output, err := targetRoot.OpenFile(spec.Target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, spec.OutputMode)
	if err != nil {
		return 0, fmt.Errorf("create %q: %w", spec.Target, err)
	}
	hasher := sha256.New()
	copied, copyErr := io.Copy(io.MultiWriter(output, hasher), io.LimitReader(input, spec.Size+1))
	if copied != spec.Size {
		copyErr = errors.Join(copyErr, fmt.Errorf("%q size changed while copying", spec.Source))
	}
	afterInfo, statErr := input.Stat()
	if statErr != nil {
		copyErr = errors.Join(copyErr, fmt.Errorf("fstat %q after copy: %w", spec.Source, statErr))
	} else if fileMetadataChanged(info, afterInfo) {
		copyErr = errors.Join(copyErr, fmt.Errorf("%q changed while copying", spec.Source))
	}
	currentInfo, lstatErr := sourceRoot.Lstat(spec.Source)
	if lstatErr != nil {
		copyErr = errors.Join(copyErr, fmt.Errorf("inspect %q after copy: %w", spec.Source, lstatErr))
	} else if fileMetadataChanged(info, currentInfo) {
		copyErr = errors.Join(copyErr, fmt.Errorf("%q was replaced while copying", spec.Source))
	}
	if spec.SHA256 != "" && !strings.EqualFold(hex.EncodeToString(hasher.Sum(nil)), spec.SHA256) {
		copyErr = errors.Join(copyErr, fmt.Errorf("%q digest mismatch", spec.Source))
	}
	closeErr := sealFile(output, spec.Target, spec.OutputMode)
	if err := errors.Join(copyErr, closeErr); err != nil {
		return 0, err
	}
	return copied, nil
}

func fileMetadataChanged(before, after os.FileInfo) bool {
	return !os.SameFile(before, after) ||
		!after.Mode().IsRegular() ||
		before.Mode() != after.Mode() ||
		before.Size() != after.Size() ||
		!before.ModTime().Equal(after.ModTime())
}

func sealFile(file *os.File, name string, mode os.FileMode) error {
	err := errors.Join(file.Chmod(mode), file.Sync(), file.Close())
	if err != nil {
		return fmt.Errorf("seal %q: %w", name, err)
	}
	return nil
}

func validateSingleEntry(name string) error {
	if name == "" || name == "." || name == ".." ||
		filepath.Clean(name) != name || filepath.Base(name) != name {
		return fmt.Errorf("path %q must be a clean single entry", name)
	}
	return nil
}

func readRootFileLimited(root *os.Root, name string, limit int64) ([]byte, error) {
	file, err := openRootRegularFile(root, name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%s exceeds %d bytes", name, limit)
	}
	return data, nil
}

func readMarker(root *os.Root, name string, out any) (bool, error) {
	data, err := readRootFileLimited(root, name, 4096)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := strictDecode(data, out); err != nil {
		return false, nil
	}
	return true, nil
}

func openStaticBundle(installerDir string) (*os.Root, []byte, BundleV1, bool, error) {
	staticPath := filepath.Join(installerDir, filepath.FromSlash(StaticRelativeDir))
	if _, err := os.Lstat(staticPath); os.IsNotExist(err) {
		return nil, nil, BundleV1{}, false, nil
	} else if err != nil {
		return nil, nil, BundleV1{}, false, fmt.Errorf("inspect preinstall source: %w", err)
	}
	installerRoot, err := openDirectoryNoSymlink(installerDir)
	if err != nil {
		return nil, nil, BundleV1{}, false, fmt.Errorf("open installer root: %w", err)
	}
	defer installerRoot.Close()
	if err := rejectRootSymlinkComponents(installerRoot, StaticRelativeDir); err != nil {
		return nil, nil, BundleV1{}, false, err
	}
	info, err := installerRoot.Lstat(StaticRelativeDir)
	if err != nil {
		return nil, nil, BundleV1{}, false, fmt.Errorf("inspect preinstall source: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, BundleV1{}, false, fmt.Errorf("preinstall source must be a directory")
	}
	root, err := installerRoot.OpenRoot(StaticRelativeDir)
	if err != nil {
		return nil, nil, BundleV1{}, false, fmt.Errorf("open preinstall source: %w", err)
	}
	data, err := readRootFileLimited(root, BundleFileName, MaxBundleJSONBytes)
	if err != nil {
		_ = root.Close()
		return nil, nil, BundleV1{}, false, err
	}
	bundle, err := DecodeBundle(data)
	if err != nil {
		_ = root.Close()
		return nil, nil, BundleV1{}, false, err
	}
	return root, data, bundle, true, nil
}

func openRootRegularFile(root *os.Root, name string) (*os.File, error) {
	if err := rejectRootSymlinkComponents(root, name); err != nil {
		return nil, err
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", name, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("fstat %q: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, fmt.Errorf("%q must be a regular file", name)
	}
	return file, nil
}

func rejectRootSymlinkComponents(root *os.Root, name string) error {
	current := ""
	for _, component := range strings.Split(filepath.ToSlash(name), "/") {
		if current == "" {
			current = component
		} else {
			current += "/" + component
		}
		info, err := root.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect preinstall path %q: %w", name, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("preinstall path contains symlink %q", name)
		}
	}
	return nil
}

func openDirectoryNoSymlink(name string) (*os.Root, error) {
	return openDirectoryPath(name, false)
}

// openDirectoryPath opens name as a trusted entry root. Ancestors of name may
// legitimately be symlinks (e.g. macOS /var -> /private/var) and are
// canonicalized once via filepath.EvalSymlinks; name's final component and
// everything below it still go through os.Root's no-symlink checks.
func openDirectoryPath(name string, create bool) (*os.Root, error) {
	absolute, err := canonicalDirectoryPath(name)
	if err != nil {
		return nil, err
	}
	parent := filepath.Dir(absolute)
	leaf := filepath.Base(absolute)
	if parent == absolute {
		return nil, fmt.Errorf("directory path %q has no parent", name)
	}
	anchor, components, err := existingAncestor(parent)
	if err != nil {
		return nil, err
	}
	components = append(components, leaf)
	current, err := os.OpenRoot(anchor)
	if err != nil {
		return nil, fmt.Errorf("open path anchor %q: %w", anchor, err)
	}
	for _, component := range components {
		info, err := current.Lstat(component)
		if os.IsNotExist(err) && create {
			if err := current.Mkdir(component, 0o755); err != nil && !os.IsExist(err) {
				_ = current.Close()
				return nil, fmt.Errorf("create directory component %q: %w", component, err)
			}
			info, err = current.Lstat(component)
		}
		if err != nil {
			_ = current.Close()
			return nil, fmt.Errorf("inspect directory component %q: %w", component, err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			_ = current.Close()
			return nil, fmt.Errorf("directory path contains symlink or non-directory %q", component)
		}
		next, err := current.OpenRoot(component)
		if err != nil {
			_ = current.Close()
			return nil, fmt.Errorf("open directory component %q: %w", component, err)
		}
		openedInfo, err := next.Stat(".")
		if err != nil {
			_ = next.Close()
			_ = current.Close()
			return nil, fmt.Errorf("stat opened directory component %q: %w", component, err)
		}
		if !os.SameFile(info, openedInfo) {
			_ = next.Close()
			_ = current.Close()
			return nil, fmt.Errorf("directory component %q changed while opening", component)
		}
		if err := current.Close(); err != nil {
			_ = next.Close()
			return nil, fmt.Errorf("close parent directory component: %w", err)
		}
		current = next
	}
	absoluteRoot, err := os.OpenRoot(absolute)
	if err != nil {
		_ = current.Close()
		return nil, fmt.Errorf("open verified directory path %q: %w", absolute, err)
	}
	currentInfo, currentErr := current.Stat(".")
	absoluteInfo, absoluteErr := absoluteRoot.Stat(".")
	if currentErr != nil || absoluteErr != nil || !os.SameFile(currentInfo, absoluteInfo) {
		_ = absoluteRoot.Close()
		_ = current.Close()
		return nil, errors.Join(
			fmt.Errorf("directory path %q changed while opening", absolute),
			currentErr,
			absoluteErr,
		)
	}
	if err := current.Close(); err != nil {
		_ = absoluteRoot.Close()
		return nil, fmt.Errorf("close verified directory path: %w", err)
	}
	return absoluteRoot, nil
}

func canonicalDirectoryPath(name string) (string, error) {
	if name == "" || filepath.Clean(name) != name {
		return "", fmt.Errorf("directory path must be non-empty and clean")
	}
	for _, component := range strings.Split(filepath.ToSlash(name), "/") {
		if component == "." || component == ".." {
			return "", fmt.Errorf("directory path must not contain %q", component)
		}
	}
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", fmt.Errorf("resolve directory path %q: %w", name, err)
	}
	return absolute, nil
}

// existingAncestor walks upward until it finds an existing path, canonicalizes
// it via filepath.EvalSymlinks, and returns it plus the not-yet-existing
// components below it (still subject to os.Root's symlink checks).
func existingAncestor(absolute string) (string, []string, error) {
	current := absolute
	var pending []string
	for {
		if _, err := os.Lstat(current); err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", nil, fmt.Errorf("resolve directory path %q: %w", current, err)
			}
			return resolved, pending, nil
		} else if !os.IsNotExist(err) {
			return "", nil, fmt.Errorf("inspect directory path %q: %w", current, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", nil, fmt.Errorf("directory path %q has no existing ancestor", absolute)
		}
		pending = append([]string{filepath.Base(current)}, pending...)
		current = parent
	}
}
