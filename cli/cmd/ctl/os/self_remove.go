package os

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// prepareSelfRemoval captures the running binary before uninstall starts, so
// cleanup cannot remove a replacement installed while the pipeline was running.
func prepareSelfRemoval() (func() error, error) {
	if runtime.GOOS == "windows" {
		// Windows does not allow unlinking a running executable.
		return nil, nil
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, err
	}
	runningPath := executable
	if runtime.GOOS == "linux" {
		runningPath = "/proc/self/exe"
	}
	running, err := os.Stat(runningPath)
	if err != nil {
		return nil, err
	}
	// os.Executable can already resolve the launch symlink. Check argv[0]
	// separately, plus the installer-managed entry point for direct launches
	// of the target binary. Only capture links to this running executable.
	return prepareExecutableRemoval(executable, running, os.Args[0], "/usr/local/bin/olares-cli"), nil
}

func prepareExecutableRemoval(executable string, running os.FileInfo, entryPaths ...string) func() error {
	type removalFile struct {
		path string
		info os.FileInfo
	}
	var files []removalFile
	seen := map[string]bool{executable: true}
	for _, entry := range entryPaths {
		if !strings.ContainsRune(entry, os.PathSeparator) {
			var err error
			entry, err = exec.LookPath(entry)
			if err != nil {
				continue
			}
		}
		entry, err := filepath.Abs(entry)
		if err != nil || seen[entry] {
			continue
		}
		link, err := os.Lstat(entry)
		if err != nil || link.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Stat(entry)
		if err != nil || !os.SameFile(running, target) {
			continue
		}
		files = append(files, removalFile{entry, link})
		seen[entry] = true
	}
	files = append(files, removalFile{executable, running})
	return func() error {
		// Validate all paths before removing anything: a replaced entry point
		// might still need the original target. Remove entry links first.
		for _, file := range files {
			if err := checkUnchangedExecutable(file.path, file.info); err != nil {
				return err
			}
		}
		for _, file := range files {
			if err := removeUnchangedExecutable(file.path, file.info); err != nil {
				return err
			}
		}
		return nil
	}
}

func checkUnchangedExecutable(path string, running os.FileInfo) error {
	current, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !os.SameFile(running, current) {
		return fmt.Errorf("%s was replaced during uninstall; leaving the replacement in place", path)
	}
	return nil
}

func removeUnchangedExecutable(path string, running os.FileInfo) error {
	if err := checkUnchangedExecutable(path, running); err != nil {
		return err
	}
	if err := os.Remove(path); !os.IsNotExist(err) {
		return err
	}
	return nil
}
