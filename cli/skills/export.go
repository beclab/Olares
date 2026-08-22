package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Modes the embedded tree cannot carry. embed.FS reports every file as 0444
// regardless of what git recorded, so a script would arrive unexecutable
// and a reference that says to run it would be wrong. The scripts are
// 100755 in git; everything else is prose.
const (
	dirMode        = 0o755
	fileMode       = 0o644
	executableMode = 0o755
	scriptsDir     = "scripts"
)

// stagingPrefix is deliberately dotted: whatever reads a skills directory
// while an export is in flight should not see a half-written skill as one.
const stagingPrefix = ".olares-cli-export-"

// Export writes the embedded suite into dir, one directory per skill, and
// returns the skills it wrote.
//
// It is the whole of what this package does to a filesystem: no network, no
// npx, no state file, no agent discovery. That is what makes it usable from
// places that have none of those — a container's boot, a CI job, a machine
// with no Node — and it is the primitive `skills install` is built on.
//
// Re-exporting replaces each skill wholesale rather than writing over the
// files that happen to exist, so a reference deleted upstream stops being
// readable here too. Each skill is staged and then swapped into place, so a
// reader either sees the previous copy or the new one, never half of either.
//
// It also leaves a SuiteMarker naming what it wrote, which is how a later
// run tells a copy of this suite from a copy of another one.
//
// A skill that already exists as a symlink is refused rather than followed.
// Writing through it would edit wherever it points — for the local
// development layout documented in cli/README.md, that is the git checkout
// this suite was built from.
func Export(dir string) ([]string, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("no directory named")
	}
	names, err := Names()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, fmt.Errorf("create %s: %w", dir, err)
	}
	for _, name := range names {
		if err := refuseSymlink(filepath.Join(dir, name)); err != nil {
			return nil, err
		}
	}

	staging, err := os.MkdirTemp(dir, stagingPrefix+"*")
	if err != nil {
		return nil, fmt.Errorf("create staging directory in %s: %w", dir, err)
	}
	defer os.RemoveAll(staging)

	if err := writeTree(staging); err != nil {
		return nil, err
	}
	for _, name := range names {
		target := filepath.Join(dir, name)
		if err := os.RemoveAll(target); err != nil {
			return nil, fmt.Errorf("replace %s: %w", target, err)
		}
		if err := os.Rename(filepath.Join(staging, name), target); err != nil {
			return nil, fmt.Errorf("move %s into place: %w", name, err)
		}
	}
	// Last, and not fatal: the skills an agent reads are all in place by
	// now, and a missing marker costs a staleness notice that fires once too
	// often. Failing here would undo an install over a bookkeeping file.
	_ = writeIdentity(dir)
	return names, nil
}

// refuseSymlink reports an error if path exists and is a symbolic link.
// Lstat, not Stat: a link to a directory answers Stat as a directory.
func refuseSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("inspect %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	destination, err := os.Readlink(path)
	if err != nil {
		destination = "somewhere else"
	}
	return fmt.Errorf("%s is a symbolic link to %s; writing there would overwrite that copy, "+
		"which for a local development checkout is the source this binary was built from. "+
		"Remove the link first if the export is what you want", path, destination)
}

// writeTree copies the whole embedded suite under root.
func writeTree(root string) error {
	return fs.WalkDir(content, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(root, filepath.FromSlash(path))
		if entry.IsDir() {
			if path == "." {
				return nil
			}
			return os.MkdirAll(target, dirMode)
		}
		source, err := fs.ReadFile(content, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := os.WriteFile(target, source, modeFor(path)); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
		return nil
	})
}

// modeFor decides whether a file is something to read or something to run.
// A skill's references tell the reader to invoke what is under scripts/;
// everything else is prose.
func modeFor(path string) fs.FileMode {
	for _, element := range strings.Split(path, "/") {
		if element == scriptsDir {
			return executableMode
		}
	}
	return fileMode
}
