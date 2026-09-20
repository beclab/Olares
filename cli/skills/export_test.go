package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportWritesTheWholeSuite(t *testing.T) {
	dir := t.TempDir()
	written, err := Export(dir)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	names, err := Names()
	if err != nil {
		t.Fatalf("names: %v", err)
	}
	if len(written) != len(names) {
		t.Fatalf("exported %d skills, suite has %d", len(written), len(names))
	}

	embedded, err := Files()
	if err != nil {
		t.Fatalf("list embedded files: %v", err)
	}
	for _, path := range embedded {
		want, err := Read(strings.SplitN(path, "/", 2)[0], strings.SplitN(path, "/", 2)[1])
		if err != nil {
			t.Fatalf("read embedded %s: %v", path, err)
		}
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read exported %s: %v", path, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs between the binary and the export", path)
		}
	}
}

// The staging directory is an implementation detail of the swap, and one
// left behind is a directory an agent would try to read as a skill. The
// marker is the one exception, and it is dotted for the same reason.
func TestExportLeavesNothingBehind(t *testing.T) {
	dir := t.TempDir()
	if _, err := Export(dir); err != nil {
		t.Fatalf("export: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read export directory: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() == SuiteMarker {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "olares-") {
			t.Errorf("export left %q beside the skills", entry.Name())
		}
	}
}

// An export is a tree that can be identified later, which is what lets the
// staleness notice ask about content instead of about a label somebody
// remembered to change.
func TestExportSaysWhatItWrote(t *testing.T) {
	dir := t.TempDir()
	if _, err := Export(dir); err != nil {
		t.Fatalf("export: %v", err)
	}
	identity, ok := ReadIdentity(dir)
	if !ok {
		t.Fatal("the export left no marker")
	}
	digest, err := Digest()
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if identity.Digest != digest {
		t.Errorf("the marker says %s; this binary carries %s", identity.Digest, digest)
	}
}

// embed.FS reports 0444 for everything, so a mode carried straight through
// would ship generate_icon.py unexecutable while its own reference tells
// the reader to run it.
func TestExportedScriptsAreExecutable(t *testing.T) {
	dir := t.TempDir()
	if _, err := Export(dir); err != nil {
		t.Fatalf("export: %v", err)
	}
	script := filepath.Join(dir, "olares-publish", "scripts", "generate_icon.py")
	info, err := os.Stat(script)
	if err != nil {
		t.Fatalf("stat %s: %v", script, err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("%s is %v; a script a reference invokes has to be executable", script, info.Mode().Perm())
	}

	prose := filepath.Join(dir, "olares-shared", "SKILL.md")
	info, err = os.Stat(prose)
	if err != nil {
		t.Fatalf("stat %s: %v", prose, err)
	}
	if info.Mode().Perm()&0o111 != 0 {
		t.Errorf("%s is %v; markdown is not executable", prose, info.Mode().Perm())
	}
}

// Exporting twice is how a container's boot and an update both work, so the
// second run has to leave the same tree — and has to take away a file the
// suite no longer has, which writing over what exists would not.
func TestReExportReplacesRatherThanMerges(t *testing.T) {
	dir := t.TempDir()
	if _, err := Export(dir); err != nil {
		t.Fatalf("first export: %v", err)
	}
	stale := filepath.Join(dir, "olares-shared", "references", "gone-upstream.md")
	if err := os.WriteFile(stale, []byte("left over from an older release\n"), 0o644); err != nil {
		t.Fatalf("plant stale file: %v", err)
	}
	if _, err := Export(dir); err != nil {
		t.Fatalf("second export: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("%s survived a re-export; a reference deleted upstream stays readable", stale)
	}
}

// The local development layout in cli/README.md links an agent directory
// straight at this checkout. Following such a link would overwrite the
// source with a copy of itself, which is a silent end to live editing.
func TestExportRefusesToWriteThroughASymlink(t *testing.T) {
	dir := t.TempDir()
	checkout := t.TempDir()
	sentinel := filepath.Join(checkout, "SKILL.md")
	if err := os.WriteFile(sentinel, []byte("the developer's working copy\n"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}
	if err := os.Symlink(checkout, filepath.Join(dir, "olares-shared")); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	if _, err := Export(dir); err == nil {
		t.Fatal("export through a symlink succeeded")
	} else if !strings.Contains(err.Error(), "symbolic link") {
		t.Errorf("error does not say what the problem is: %v", err)
	}
	source, err := os.ReadFile(sentinel)
	if err != nil || !strings.Contains(string(source), "developer") {
		t.Errorf("the linked-to copy was touched: %v", err)
	}
	// The refusal is a precondition, not a partial write: nothing else got
	// exported either, so the user's next run starts from a known state.
	if _, err := os.Stat(filepath.Join(dir, "olares-market")); !os.IsNotExist(err) {
		t.Error("export wrote some skills before refusing")
	}
}
