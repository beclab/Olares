package syncer

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func mustSync(t *testing.T, dir string, desired map[string][]byte) {
	t.Helper()
	if err := syncDir(dir, desired); err != nil {
		t.Fatalf("syncDir: %v", err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("%s still present, stat err = %v", path, err)
	}
}

func TestSyncDirCreatesFilesAndMissingDirs(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "target")

	mustSync(t, dir, map[string][]byte{
		"app.conf":  []byte("listen = 8080"),
		"cert.p12":  {0x00, 0x01, 0x02},
		"empty.txt": {},
	})

	if got := readFile(t, filepath.Join(dir, "app.conf")); got != "listen = 8080" {
		t.Errorf("app.conf = %q", got)
	}
	if got := readFile(t, filepath.Join(dir, "cert.p12")); got != "\x00\x01\x02" {
		t.Errorf("cert.p12 = %q", got)
	}
	if got := readFile(t, filepath.Join(dir, "empty.txt")); got != "" {
		t.Errorf("empty.txt = %q", got)
	}

	info, err := os.Stat(filepath.Join(dir, "app.conf"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != fileMode {
		t.Errorf("app.conf mode = %o, want %o", perm, fileMode)
	}
}

func TestSyncDirUpdatesChangedContent(t *testing.T) {
	dir := t.TempDir()

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("old")})
	mustSync(t, dir, map[string][]byte{"app.conf": []byte("new")})

	if got := readFile(t, filepath.Join(dir, "app.conf")); got != "new" {
		t.Errorf("app.conf = %q, want %q", got, "new")
	}
}

func TestSyncDirLeavesUnchangedContentAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.conf")

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("stable")})

	// Backdate the file so an unnecessary rewrite shows up as a bumped
	// modification time instead of relying on timestamp resolution.
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("stable")})

	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("mod time changed from %v to %v, file was rewritten", before.ModTime(), after.ModTime())
	}
}

func TestSyncDirRemovesDroppedKey(t *testing.T) {
	dir := t.TempDir()

	mustSync(t, dir, map[string][]byte{
		"keep.conf": []byte("keep"),
		"drop.conf": []byte("drop"),
	})
	mustSync(t, dir, map[string][]byte{"keep.conf": []byte("keep")})

	assertMissing(t, filepath.Join(dir, "drop.conf"))
	if got := readFile(t, filepath.Join(dir, "keep.conf")); got != "keep" {
		t.Errorf("keep.conf = %q", got)
	}
}

func TestSyncDirEmptiesDirWhenNothingIsDesired(t *testing.T) {
	dir := t.TempDir()

	mustSync(t, dir, map[string][]byte{"a.conf": []byte("a"), "b.conf": []byte("b")})
	mustSync(t, dir, map[string][]byte{})

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("dir still has %d entries, want 0", len(entries))
	}
}

func TestSyncDirRemovesLeftoverTempFile(t *testing.T) {
	dir := t.TempDir()
	leftover := filepath.Join(dir, tempPrefix+"123")
	if err := os.WriteFile(leftover, []byte("partial"), fileMode); err != nil {
		t.Fatalf("write: %v", err)
	}

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("app")})

	assertMissing(t, leftover)
}

func TestSyncDirLeavesNoTempFileBehind(t *testing.T) {
	dir := t.TempDir()

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("app")})

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "app.conf" {
		t.Errorf("dir contains %d entries, want only app.conf", len(entries))
	}
}

func TestSyncDirKeepsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested")
	if err := os.Mkdir(sub, dirMode); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	nested := filepath.Join(sub, "untouched.conf")
	if err := os.WriteFile(nested, []byte("untouched"), fileMode); err != nil {
		t.Fatalf("write: %v", err)
	}

	mustSync(t, dir, map[string][]byte{"app.conf": []byte("app")})

	if got := readFile(t, nested); got != "untouched" {
		t.Errorf("nested file = %q, want %q", got, "untouched")
	}
}

func TestSyncDirSkipsUnusableKeys(t *testing.T) {
	dir := t.TempDir()

	mustSync(t, dir, map[string][]byte{
		"good.conf":     []byte("good"),
		"sub/bad.conf":  []byte("bad"),
		"..":            []byte("bad"),
		"../escape.txt": []byte("bad"),
	})

	if got := readFile(t, filepath.Join(dir, "good.conf")); got != "good" {
		t.Errorf("good.conf = %q", got)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("dir has %d entries, want only good.conf", len(entries))
	}
	assertMissing(t, filepath.Join(filepath.Dir(dir), "escape.txt"))
}

func TestValidFileName(t *testing.T) {
	valid := []string{"app.conf", "..leading-dots", "a-b_c.1", ".hidden"}
	for _, name := range valid {
		if !validFileName(name) {
			t.Errorf("validFileName(%q) = false, want true", name)
		}
	}

	invalid := []string{"", ".", "..", "a/b", "/abs", "../up", "sub/"}
	for _, name := range invalid {
		if validFileName(name) {
			t.Errorf("validFileName(%q) = true, want false", name)
		}
	}
}
