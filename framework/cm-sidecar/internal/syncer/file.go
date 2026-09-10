package syncer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

const (
	fileMode = 0o644
	dirMode  = 0o755

	// tempPrefix is kept short and independent of the file name: ConfigMap keys
	// may be up to 253 characters, which combined with a longer prefix and the
	// random suffix would exceed the 255 byte file name limit.
	tempPrefix = ".cm-sidecar-tmp-"
)

// syncDir makes dir contain exactly the files described by desired, where every
// map entry is a file name and its content. Files in dir that are absent from
// desired are removed, so dir must be dedicated to this process.
func syncDir(dir string, desired map[string][]byte) error {
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	keep := make(map[string]struct{}, len(desired))
	for name, content := range desired {
		if !validFileName(name) {
			klog.Warningf("skipping key %q: not usable as a file name", name)
			continue
		}
		if err := writeFile(dir, name, content); err != nil {
			return err
		}
		keep[name] = struct{}{}
	}

	return prune(dir, keep)
}

// writeFile writes content through a temporary file in the same directory
// followed by a rename, so a reader never observes a partially written file.
// Unchanged content is left alone to avoid touching the modification time,
// which consumers may be watching.
func writeFile(dir, name string, content []byte) error {
	path := filepath.Join(dir, name)

	if current, err := os.ReadFile(path); err == nil && bytes.Equal(current, content) {
		return nil
	}

	tmp, err := os.CreateTemp(dir, tempPrefix)
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmp.Name(), err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", tmp.Name(), err)
	}
	// os.CreateTemp creates the file with 0600.
	if err := os.Chmod(tmp.Name(), fileMode); err != nil {
		return fmt.Errorf("chmod %s: %w", tmp.Name(), err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmp.Name(), path, err)
	}

	klog.Infof("wrote %s (%d bytes)", path, len(content))
	return nil
}

// prune removes the top level regular files of dir that are not in keep. That
// covers keys dropped from a ConfigMap, deleted ConfigMaps, files left over
// from a previous run and temporary files of a run that was killed mid-write.
// Subdirectories and anything that is not a regular file are left untouched.
func prune(dir string, keep map[string]struct{}) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		if _, ok := keep[entry.Name()]; ok {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", path, err)
		}
		klog.Infof("removed %s", path)
	}

	return nil
}

// validFileName rejects keys that would escape dir or address it as a whole.
// The Kubernetes API already restricts ConfigMap keys to a safe character set,
// so this only guards against unexpected input.
func validFileName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	return filepath.Base(name) == name
}
