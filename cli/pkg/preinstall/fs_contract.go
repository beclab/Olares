package preinstall

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
