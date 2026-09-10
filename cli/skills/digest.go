package skills

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// SuiteMarker names the file Export leaves beside the skills it wrote, so a
// copy on disk can be told apart from the copy a binary carries without
// reading either tree.
//
// Dotted, like the staging directories, because it is not a skill and
// whatever lists this directory looking for skills should not find it.
const SuiteMarker = ".olares-cli-suite"

// Identity is what a written copy of the suite records about itself.
//
// Digest decides whether a reinstall is needed; Version is what a person is
// told. One field cannot do both jobs, because a version does not identify
// a suite: the daily channel ships whatever is on main under the release
// being worked towards, so a label covers as many different suites as there
// are days in that release.
type Identity struct {
	Digest  string `json:"digest"`
	Version string `json:"version"`
}

// Digest is the identity of the embedded suite: every path and every byte of
// it, in sorted order, hashed.
func Digest() (string, error) { return digestOf(content) }

// digestOf hashes any tree, which is what lets digest_test.go state its
// properties against trees written for the purpose rather than against the
// one tree this binary happens to carry.
//
// Each file contributes its path, its length, and then its bytes. The length
// is what keeps two different trees from hashing the same: a file ending in a
// newline followed by a sibling produces the same stream as one file holding
// both, and both of those are ordinary things for a suite of markdown to be.
func digestOf(fsys fs.FS) (string, error) {
	var paths []string
	err := fs.WalkDir(fsys, ".", func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			paths = append(paths, p)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk suite: %w", err)
	}
	sort.Strings(paths)

	sum := sha256.New()
	for _, p := range paths {
		source, err := fs.ReadFile(fsys, p)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", p, err)
		}
		fmt.Fprintf(sum, "%s\n%d\n", p, len(source))
		sum.Write(source)
	}
	return hex.EncodeToString(sum.Sum(nil)), nil
}

// SuiteVersion is the release the embedded skills name. It reads the first
// skill because TestTheSuiteShipsOneVersion pins all of them to one value:
// they are stamped together by the release that builds them.
func SuiteVersion() (string, error) {
	metas, err := List()
	if err != nil {
		return "", err
	}
	if len(metas) == 0 {
		return "", fmt.Errorf("the embedded suite is empty")
	}
	return metas[0].Version, nil
}

// ReadIdentity returns what the copy of the suite in dir says it is.
//
// The boolean is false for every uncertainty — no marker, unreadable, not
// the JSON expected, no digest recorded — because the only caller decides
// whether to interrupt somebody's command, and a guess is worse there than
// saying nothing.
func ReadIdentity(dir string) (Identity, bool) {
	source, err := os.ReadFile(filepath.Join(dir, SuiteMarker))
	if err != nil {
		return Identity{}, false
	}
	var identity Identity
	if err := json.Unmarshal(source, &identity); err != nil || identity.Digest == "" {
		return Identity{}, false
	}
	return identity, true
}

// writeIdentity records what was just written into dir.
//
// Not fatal to an export that has already placed every skill: the files an
// agent reads are there, and losing the marker costs a notice that fires
// once too often, which is a great deal less than a failed install.
func writeIdentity(dir string) error {
	digest, err := Digest()
	if err != nil {
		return err
	}
	version, err := SuiteVersion()
	if err != nil {
		return err
	}
	source, err := json.Marshal(Identity{Digest: digest, Version: version})
	if err != nil {
		return fmt.Errorf("encode suite marker: %w", err)
	}
	staged, err := os.CreateTemp(dir, SuiteMarker+".*")
	if err != nil {
		return fmt.Errorf("create suite marker in %s: %w", dir, err)
	}
	defer os.Remove(staged.Name())
	if _, err := staged.Write(append(source, '\n')); err != nil {
		staged.Close()
		return fmt.Errorf("write suite marker: %w", err)
	}
	if err := staged.Close(); err != nil {
		return fmt.Errorf("write suite marker: %w", err)
	}
	if err := os.Chmod(staged.Name(), fileMode); err != nil {
		return fmt.Errorf("set mode on suite marker: %w", err)
	}
	return os.Rename(staged.Name(), filepath.Join(dir, SuiteMarker))
}
