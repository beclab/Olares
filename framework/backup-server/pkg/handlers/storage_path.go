package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"olares.com/backup-server/pkg/constant"
)

type StorageRoot int

const (
	UserspaceStorageRoot StorageRoot = iota
	AppcacheStorageRoot
	CommonStorageRoot
	ExternalStorageRoot
)

// ParseStoragePath accepts only the virtual paths exposed by Olares and returns
// the root selected by the prefix plus a clean relative path. It intentionally
// rejects traversal rather than cleaning it: callers must not turn an unsafe
// persisted path into a valid one.
func ParseStoragePath(p string) (StorageRoot, string, error) {
	if p == "" {
		return 0, "", fmt.Errorf("storage path is required")
	}
	if strings.ContainsRune(p, '\x00') || strings.Contains(p, "\\") {
		return 0, "", fmt.Errorf("storage path contains an invalid character")
	}
	if strings.Contains(p, "//") {
		return 0, "", fmt.Errorf("storage path contains an empty component")
	}

	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return 0, "", fmt.Errorf("storage path is required")
	}

	root, relative, ok := storagePathRoot(p)
	if !ok {
		return 0, "", fmt.Errorf("storage path must be under an allowed root")
	}
	for _, component := range strings.Split(relative, "/") {
		if component == "." || component == ".." {
			return 0, "", fmt.Errorf("storage path traversal is not allowed")
		}
	}

	return root, relative, nil
}

func storagePathRoot(p string) (StorageRoot, string, bool) {
	external := fmt.Sprintf("/Files/External/%s", constant.NodeName)
	cache := fmt.Sprintf("/Cache/%s", constant.NodeName)
	for _, candidate := range []struct {
		prefix string
		root   StorageRoot
	}{
		{external, ExternalStorageRoot},
		{cache, AppcacheStorageRoot},
		{"/Common", CommonStorageRoot},
		{"/Files", UserspaceStorageRoot},
	} {
		if p == candidate.prefix {
			return candidate.root, "", true
		}
		if strings.HasPrefix(p, candidate.prefix+"/") {
			return candidate.root, strings.TrimPrefix(p, candidate.prefix+"/"), true
		}
	}
	return 0, "", false
}

// ResolveStoragePath re-validates persisted paths in the worker and ensures
// the resulting path, including existing symlinks, remains under its root.
func ResolveStoragePath(p, userspaceRoot, appcacheRoot string) (string, error) {
	storageRoot, relative, err := ParseStoragePath(p)
	if err != nil {
		return "", err
	}

	root := userspaceRoot
	switch storageRoot {
	case AppcacheStorageRoot:
		root = appcacheRoot
	case CommonStorageRoot:
		root = constant.CommonPath
	case ExternalStorageRoot:
		root = constant.ExternalPath
	}

	candidate := filepath.Join(root, filepath.FromSlash(relative))
	if err := ensurePathWithinRoot(root, candidate); err != nil {
		return "", err
	}
	return candidate, nil
}

func ValidateStorageName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return fmt.Errorf("storage name is invalid")
	}
	return nil
}

func ensurePathWithinRoot(root, candidate string) error {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return fmt.Errorf("resolve storage root: %w", err)
	}

	checkPath := candidate
	for {
		if _, err := os.Lstat(checkPath); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("inspect storage path: %w", err)
		}
		parent := filepath.Dir(checkPath)
		if parent == checkPath {
			return fmt.Errorf("storage path has no existing parent")
		}
		checkPath = parent
	}

	realPath, err := filepath.EvalSymlinks(checkPath)
	if err != nil {
		return fmt.Errorf("resolve storage path: %w", err)
	}
	rel, err := filepath.Rel(realRoot, realPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("storage path escapes its allowed root")
	}
	return nil
}
