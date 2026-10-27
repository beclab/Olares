package aptsource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/gofrs/flock"
)

// EnsureDebianComponents prepares the local Debian machine's APT sources.
// Run prepare as root; this helper does not elevate privileges or use SSH.
func EnsureDebianComponents() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("Debian APT preparation requires root; run olares-cli prepare with sudo")
	}
	release, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return fmt.Errorf("read local Debian release: %w", err)
	}
	return ensure("/etc/apt", releaseCodename(string(release)), validateSources)
}

func releaseCodename(release string) string {
	for _, line := range strings.Split(release, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok && key == "VERSION_CODENAME" {
			return strings.Trim(strings.TrimSpace(value), "\"'")
		}
	}
	return ""
}

type snapshot struct {
	files []sourceFile
	state map[string]sourceState
}

type sourceState struct {
	path   string // Resolved target: replace this file, not its symlink aliases.
	entry  os.FileInfo
	target os.FileInfo
}

func readSource(file string) (sourceState, string, error) {
	var s sourceState
	var err error
	if s.entry, err = os.Lstat(file); err != nil {
		return s, "", err
	}
	if s.path, err = filepath.EvalSymlinks(file); err != nil {
		return s, "", err
	}
	if s.target, err = os.Stat(s.path); err != nil {
		return s, "", err
	}
	if !s.target.Mode().IsRegular() {
		return s, "", fmt.Errorf("not a regular APT source: %s", file)
	}
	data, err := os.ReadFile(s.path)
	return s, string(data), err
}

func readSources(root string) (snapshot, error) {
	entries, err := os.ReadDir(filepath.Join(root, "sources.list.d"))
	if err != nil && !os.IsNotExist(err) {
		return snapshot{}, err
	}
	names := []string{"sources.list"}
	for _, entry := range entries {
		name := "sources.list.d/" + entry.Name()
		if validName(name) {
			names = append(names, name)
		}
	}
	return snapshotSources(root, names)
}

func snapshotSources(root string, names []string) (snapshot, error) {
	s := snapshot{state: make(map[string]sourceState)}
	for _, name := range names {
		state, content, err := readSource(filepath.Join(root, name))
		// A file (or symlink target) may disappear after directory discovery.
		// A later snapshot still detects deletion of an already-read source.
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return s, err
		}
		s.files = append(s.files, sourceFile{name, content})
		s.state[name] = state
	}
	return s, nil
}

func (s snapshot) equal(other snapshot) bool {
	if !slices.Equal(s.files, other.files) {
		return false
	}
	for name, before := range s.state {
		after := other.state[name]
		if before.path != after.path || !sameMetadata(before.entry, after.entry) || !sameMetadata(before.target, after.target) {
			return false
		}
	}
	return true
}

func sameMetadata(before, after os.FileInfo) bool {
	if !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		return false
	}
	a, b := before.Sys().(*syscall.Stat_t), after.Sys().(*syscall.Stat_t)
	return a.Uid == b.Uid && a.Gid == b.Gid
}

// ensure keeps filesystem tests independent of the host's /etc/apt and APT
// executable. Production always supplies the native APT validator below.
func ensure(root, suite string, validate func(string) error) error {
	lock := flock.New(filepath.Join(root, ".olares-sources.lock"))
	defer lock.Close()
	locked, err := lock.TryLock()
	if err != nil {
		return fmt.Errorf("lock APT sources: %w", err)
	}
	if !locked {
		return fmt.Errorf("APT sources are being prepared by another process; retry later")
	}
	// flock is released by the kernel on process exit, even without defer.
	// Keep the lock file: removing it could allow writers to lock different inodes.

	before, err := readSources(root)
	if err != nil {
		return fmt.Errorf("read APT sources: %w", err)
	}
	changes, err := plan(before.files, suite)
	if err != nil || len(changes) == 0 {
		return err
	}
	work, err := os.MkdirTemp(root, ".olares-sources-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	if err := os.Mkdir(filepath.Join(work, "sources.list.d"), 0755); err != nil {
		return err
	}
	for _, file := range before.files {
		if err := os.WriteFile(filepath.Join(work, file.name), []byte(file.content), 0644); err != nil {
			return err
		}
	}
	for _, change := range changes {
		if err := os.WriteFile(filepath.Join(work, change.name), []byte(change.after), 0644); err != nil {
			return err
		}
	}
	if err := validate(work); err != nil {
		return err
	}

	// Stage all replacements and backups before publishing any source file.
	var staged []string
	var targets []string
	updates := make(map[string]string)
	defer func() {
		for _, file := range staged {
			os.Remove(file)
		}
	}()
	for _, change := range changes {
		target := filepath.Join(root, change.name)
		state := before.state[change.name]
		if change.exists {
			target = state.path
		} else if _, err := os.Lstat(target); !os.IsNotExist(err) {
			// In particular, never overwrite a dangling managed-file symlink
			// that was skipped while reading sources.
			return fmt.Errorf("new APT source path is not absent: %s (%v)", target, err)
		}
		if after, exists := updates[target]; exists {
			if after != change.after {
				return fmt.Errorf("conflicting updates through APT source aliases: %s", target)
			}
			continue
		}
		updates[target] = change.after
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		info := state.target
		tmp, err := stageFile(target, change.after, info)
		if err != nil {
			return err
		}
		staged = append(staged, tmp)
		targets = append(targets, target)
		if change.exists {
			var original string
			for _, file := range before.files {
				if file.name == change.name {
					original = file.content
					break
				}
			}
			if err := backupFile(target, original, info); err != nil {
				return err
			}
		}
	}
	current, err := readSources(root)
	if err != nil {
		return err
	}
	if !before.equal(current) {
		return fmt.Errorf("APT sources changed during preparation; retry")
	}
	for _, change := range changes {
		if !change.exists {
			if _, err := os.Lstat(filepath.Join(root, change.name)); !os.IsNotExist(err) {
				return fmt.Errorf("new APT source path changed during preparation: %s; retry", change.name)
			}
		}
	}
	for i, target := range targets {
		if err := os.Rename(staged[i], target); err != nil {
			return fmt.Errorf("replace APT source %s: %w", target, err)
		}
	}
	return nil
}

// stageFile writes and syncs a same-directory temporary file, preserving the
// original owner and mode. The caller owns cleanup after a successful return.
func stageFile(target, content string, info os.FileInfo) (name string, err error) {
	f, err := os.CreateTemp(filepath.Dir(target), ".olares-source-")
	if err != nil {
		return "", err
	}
	name = f.Name()
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(name)
		}
	}()
	if _, err = f.WriteString(content); err != nil {
		return name, err
	}
	mode := os.FileMode(0644)
	if info != nil {
		stat := info.Sys().(*syscall.Stat_t)
		if err = f.Chown(int(stat.Uid), int(stat.Gid)); err != nil {
			return name, err
		}
		mode = info.Mode()
	}
	if err = f.Chmod(mode); err != nil {
		return name, err
	}
	if err = f.Sync(); err != nil {
		return name, err
	}
	err = f.Close()
	return name, err
}

func backupFile(target, content string, info os.FileInfo) error {
	backup := target + ".olares.bak"
	if existing, err := os.Lstat(backup); err == nil {
		if !existing.Mode().IsRegular() {
			return fmt.Errorf("invalid APT backup: %s", backup)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	tmp, err := stageFile(target, content, info)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)
	if err := os.Chtimes(tmp, info.ModTime(), info.ModTime()); err != nil {
		return err
	}
	// Link publishes a complete backup atomically without replacing an existing one.
	return os.Link(tmp, backup)
}

func validateSources(root string) error {
	if _, err := os.Stat(filepath.Join(root, managedFile)); err == nil {
		if _, err := os.ReadFile(debianKeyring); err != nil {
			return fmt.Errorf("read Debian archive keyring: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	output, err := exec.CommandContext(ctx, "apt-get", "indextargets",
		"-o", "Dir::Etc::sourcelist="+filepath.Join(root, "sources.list"),
		"-o", "Dir::Etc::sourceparts="+filepath.Join(root, "sources.list.d")).CombinedOutput()
	if err != nil {
		return fmt.Errorf("validate APT sources: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
