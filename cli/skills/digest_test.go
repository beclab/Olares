package skills

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestTheDigestIsTheSameEveryTime(t *testing.T) {
	first, err := Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	second, err := Digest()
	if err != nil {
		t.Fatalf("digest again: %v", err)
	}
	if first != second {
		t.Errorf("two reads of one binary disagree: %s and %s", first, second)
	}
	if len(first) != 64 {
		t.Errorf("digest %q is not a sha256 in hex", first)
	}
}

// The whole point: a skill edited on main and shipped by a daily build under
// the version somebody bumped weeks ago has to be visible as a change.
func TestOneEditedByteChangesTheDigest(t *testing.T) {
	before := digest(t, fstest.MapFS{
		"olares-shared/SKILL.md": &fstest.MapFile{Data: []byte("read this first")},
	})
	after := digest(t, fstest.MapFS{
		"olares-shared/SKILL.md": &fstest.MapFile{Data: []byte("read this second")},
	})
	if before == after {
		t.Error("editing a skill did not change the digest")
	}
}

// Deleting a reference and adding one of the same size is a change to the
// suite even though the bytes weigh the same.
func TestRenamingAFileChangesTheDigest(t *testing.T) {
	before := digest(t, fstest.MapFS{
		"olares-router/references/calling.md": &fstest.MapFile{Data: []byte("x")},
	})
	after := digest(t, fstest.MapFS{
		"olares-router/references/routing.md": &fstest.MapFile{Data: []byte("x")},
	})
	if before == after {
		t.Error("moving a reference did not change the digest")
	}
}

// Without a length between the path and the bytes these two hash the same:
// the first tree's single file ends in a newline and continues with what
// reads exactly like the second tree's next path. Both shapes — a markdown
// file ending in a newline, and a skill with two files — are what this suite
// is made of, so the collision is not a contrivance to guard against.
func TestATreeCannotBeReshapedIntoTheSameDigest(t *testing.T) {
	oneFile := digest(t, fstest.MapFS{
		"a": &fstest.MapFile{Data: []byte("x\nb\ny")},
	})
	twoFiles := digest(t, fstest.MapFS{
		"a": &fstest.MapFile{Data: []byte("x\n")},
		"b": &fstest.MapFile{Data: []byte("y")},
	})
	if oneFile == twoFiles {
		t.Error("two different trees produced one digest")
	}
}

func TestTheSuiteVersionIsWhatTheSkillsDeclare(t *testing.T) {
	version, err := SuiteVersion()
	if err != nil {
		t.Fatalf("suite version: %v", err)
	}
	metas, err := List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, meta := range metas {
		if meta.Version != version {
			t.Errorf("%s declares %s but the suite reports %s", meta.Name, meta.Version, version)
		}
	}
}

// The marker is how a tree on disk answers "did this binary write you?", so
// what Export records has to be what a later run reads back.
func TestTheMarkerRoundTrips(t *testing.T) {
	dir := t.TempDir()
	if err := writeIdentity(dir); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	identity, ok := ReadIdentity(dir)
	if !ok {
		t.Fatal("the marker just written could not be read")
	}
	digest, err := Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	version, err := SuiteVersion()
	if err != nil {
		t.Fatalf("suite version: %v", err)
	}
	if identity.Digest != digest || identity.Version != version {
		t.Errorf("read back %+v; wrote digest %s version %s", identity, digest, version)
	}
	// Nothing is left behind: a temporary file beside the marker would be
	// listed by whatever reads this directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	if len(entries) != 1 || entries[0].Name() != SuiteMarker {
		t.Errorf("writing the marker left %d entries in the directory", len(entries))
	}
}

// Every uncertainty reads as "no marker", because the caller's alternative
// to knowing is interrupting somebody's command with a guess.
func TestAnUnreadableMarkerIsNoMarker(t *testing.T) {
	for name, source := range map[string]string{
		"not json":      "digest=abc\n",
		"no digest":     `{"version":"1.12.7-cli.5"}`,
		"empty digest":  `{"digest":"","version":"1.12.7-cli.5"}`,
		"half a object": `{"digest":`,
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, SuiteMarker), []byte(source), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		if _, ok := ReadIdentity(dir); ok {
			t.Errorf("%s was read as a marker", name)
		}
	}
	if _, ok := ReadIdentity(t.TempDir()); ok {
		t.Error("a directory with no marker reported one")
	}
}

func digest(t *testing.T, fsys fstest.MapFS) string {
	t.Helper()
	sum, err := digestOf(fsys)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	return sum
}
