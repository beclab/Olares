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

func openDirectoryPath(name string, create bool) (*os.Root, error) {
	absolute, anchor, components, err := rootedPathComponents(name)
	if err != nil {
		return nil, err
	}
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

func rootedPathComponents(name string) (string, string, []string, error) {
	if name == "" || filepath.Clean(name) != name {
		return "", "", nil, fmt.Errorf("directory path must be non-empty and clean")
	}
	for _, component := range strings.Split(filepath.ToSlash(name), "/") {
		if component == "." || component == ".." {
			return "", "", nil, fmt.Errorf("directory path must not contain %q", component)
		}
	}
	absolute, err := filepath.Abs(name)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve directory path %q: %w", name, err)
	}
	volume := filepath.VolumeName(absolute)
	rest := strings.TrimPrefix(absolute, volume)
	separator := string(filepath.Separator)
	if !strings.HasPrefix(rest, separator) {
		return "", "", nil, fmt.Errorf("directory path %q has no absolute volume anchor", name)
	}
	anchor := volume + separator
	rest = strings.TrimLeft(rest, separator)
	if rest == "" {
		return absolute, anchor, nil, nil
	}
	components := strings.Split(rest, separator)
	for _, component := range components {
		if component == "" || component == "." || component == ".." {
			return "", "", nil, fmt.Errorf("directory path %q contains invalid component", name)
		}
	}
	return absolute, anchor, components, nil
}
