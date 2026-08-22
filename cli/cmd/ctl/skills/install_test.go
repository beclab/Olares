package skills

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// sandboxHome points home-directory lookups at a temporary tree. Both names
// are set because os.UserHomeDir reads HOME on Unix and USERPROFILE on
// Windows, and a test that only set one would quietly install into the
// developer's own home on the other platform.
func sandboxHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func mkdirAll(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	return path
}

func quiet() *outputOptions { return &outputOptions{Output: "table", Quiet: true} }

// The reason this command places files itself instead of delegating to the
// skills CLI: `npx skills add --agent '*'` writes to every agent it knows
// about, which on a clean machine means creating fifty-six directories for
// tools that are not installed, and without that flag it writes to no agent
// directory at all. Neither is something to do to somebody's home directory.
func TestInstallOnlyTouchesAgentsThatAreHere(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	cursor := mkdirAll(t, filepath.Join(home, ".cursor", "skills"))

	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("install: %v", err)
	}

	for _, dir := range []string{claude, cursor} {
		link := filepath.Join(dir, "olares-shared")
		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("stat %s: %v", link, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a link into the store", link)
		}
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatalf("resolve %s: %v", link, err)
		}
		if !strings.Contains(resolved, filepath.Join(".agents", "skills")) {
			t.Errorf("%s resolves to %s, not into the store", link, resolved)
		}
	}

	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatalf("read home: %v", err)
	}
	var created []string
	for _, entry := range entries {
		created = append(created, entry.Name())
	}
	sort.Strings(created)
	want := []string{".agents", ".claude", ".cursor"}
	if strings.Join(created, " ") != strings.Join(want, " ") {
		t.Errorf("install left %v in the home directory; expected only %v", created, want)
	}
}

// An agent that has never run has no directory, and several read the shared
// store directly. Writing the store and saying nothing was linked is the
// honest outcome, not a failure.
func TestInstallOnACleanMachineWritesOnlyTheStore(t *testing.T) {
	home := sandboxHome(t)
	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("install: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "olares-shared", "SKILL.md")); err != nil {
		t.Errorf("the store is not populated: %v", err)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatalf("read home: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("install created %d directories on a clean machine; it should create one", len(entries))
	}
}

// The local development layout links straight at the checkout, and an install
// that repointed it would end live editing silently. In an agent's own
// directory that is one agent's arrangement: leave it, install for everyone
// else. Failing the whole command would leave every other agent without the
// update, which is the same silence arrived at from the other side.
func TestInstallSkipsOnlyTheAgentLinkedByHand(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	cursor := mkdirAll(t, filepath.Join(home, ".cursor", "skills"))
	checkout := t.TempDir()
	byHand := filepath.Join(claude, "olares-shared")
	if err := os.Symlink(checkout, byHand); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("install: %v", err)
	}

	if target, err := os.Readlink(byHand); err != nil || target != checkout {
		t.Errorf("the hand-made link was disturbed: %v -> %v", err, target)
	}
	// That directory is skipped whole, not partially converted.
	if _, err := os.Lstat(filepath.Join(claude, "olares-market")); !os.IsNotExist(err) {
		t.Error("install linked some skills into the directory it had to skip")
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "olares-shared", "SKILL.md")); err != nil {
		t.Errorf("the store is not populated: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(cursor, "olares-shared")); err != nil {
		t.Errorf("a clean agent directory was held back by another agent's link: %v", err)
	}
}

// The skip is the only sign that agent did not get the update, so it carries
// both the reason and which flag would take it over.
func TestTheSkippedAgentIsReportedAsLinkedByHand(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	store := mkdirAll(t, filepath.Join(home, ".agents", "skills"))
	checkout := t.TempDir()
	if err := os.Symlink(checkout, filepath.Join(claude, "olares-shared")); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	result := link(claude, store, []string{"olares-shared"}, false)
	if result.kind != skipped || !result.handMade {
		t.Fatalf("got kind %v, handMade %v; want a hand-made skip", result.kind, result.handMade)
	}
	for _, expected := range []string{"olares-shared", checkout} {
		if !strings.Contains(result.reason, expected) {
			t.Errorf("the reason does not mention %q: %s", expected, result.reason)
		}
	}
}

// The store is the one copy every agent directory links to, so a hand-made
// link there decides the whole install: there is nothing to link to until it
// is dealt with.
func TestInstallRefusesAHandMadeLinkInTheStore(t *testing.T) {
	home := sandboxHome(t)
	mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	store := mkdirAll(t, filepath.Join(home, ".agents", "skills"))
	checkout, marker := fakeCheckout(t)
	if err := os.Symlink(checkout, filepath.Join(store, "olares-shared")); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	err := runInstall(quiet(), false)
	if err == nil {
		t.Fatal("install proceeded over a hand-made link in the store")
	}
	for _, expected := range []string{"olares-shared", checkout, "--force"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("the refusal does not mention %q:\n%s", expected, err)
		}
	}
	target, err := os.Readlink(filepath.Join(store, "olares-shared"))
	if err != nil || target != checkout {
		t.Errorf("the hand-made link was disturbed: %v -> %v", err, target)
	}
	assertUntouched(t, marker)
}

// --force is what that refusal names, so it has to reach the store: skipping
// the check and then failing inside Export — whose own refusal says nothing
// about --force — is what this covers.
func TestForceTakesOverAHandMadeLinkInTheStore(t *testing.T) {
	home := sandboxHome(t)
	store := mkdirAll(t, filepath.Join(home, ".agents", "skills"))
	checkout, marker := fakeCheckout(t)
	installed := filepath.Join(store, "olares-shared")
	if err := os.Symlink(checkout, installed); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	if err := runInstall(quiet(), true); err != nil {
		t.Fatalf("forced install: %v", err)
	}
	info, err := os.Lstat(installed)
	if err != nil {
		t.Fatalf("stat %s: %v", installed, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("--force left the hand-made link in place")
	}
	if !info.IsDir() {
		t.Errorf("%s is not the installed copy", installed)
	}
	// The link was unlinked rather than written through.
	assertUntouched(t, marker)
}

// fakeCheckout stands in for a git checkout of this repository's skills, and
// returns the directory a link would name plus a file that must survive.
func fakeCheckout(t *testing.T) (string, string) {
	t.Helper()
	dir := mkdirAll(t, filepath.Join(t.TempDir(), "olares-shared"))
	marker := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(marker, []byte("the working tree\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", marker, err)
	}
	return dir, marker
}

func assertUntouched(t *testing.T, marker string) {
	t.Helper()
	source, err := os.ReadFile(marker)
	if err != nil || !strings.Contains(string(source), "the working tree") {
		t.Errorf("the checkout was written into: %v", err)
	}
}

func TestForceTakesOverLinksMadeByHand(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	checkout := t.TempDir()
	link := filepath.Join(claude, "olares-shared")
	if err := os.Symlink(checkout, link); err != nil {
		t.Skipf("cannot create symlinks here: %v", err)
	}

	if err := runInstall(quiet(), true); err != nil {
		t.Fatalf("forced install: %v", err)
	}
	resolved, err := filepath.EvalSymlinks(link)
	if err != nil {
		t.Fatalf("resolve %s: %v", link, err)
	}
	if !strings.Contains(resolved, filepath.Join(".agents", "skills")) {
		t.Errorf("--force did not repoint the link; it resolves to %s", resolved)
	}
}

// Installing is how an upgrade brings the skills along, so it runs on a
// machine that is already installed far more often than on a fresh one.
func TestInstallingTwiceChangesNothing(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))

	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("first install: %v", err)
	}
	first := listing(t, claude)
	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("second install: %v", err)
	}
	if second := listing(t, claude); second != first {
		t.Errorf("the second install changed the layout:\n%s\nbecame\n%s", first, second)
	}
}

// Somebody else's directory under our name is not ours to delete. Reporting
// the skip is what makes an agent that did not get the update visible.
func TestInstallLeavesAForeignDirectoryAloneAndSaysSo(t *testing.T) {
	home := sandboxHome(t)
	claude := mkdirAll(t, filepath.Join(home, ".claude", "skills"))
	theirs := mkdirAll(t, filepath.Join(claude, "olares-shared"))
	marker := filepath.Join(theirs, "SKILL.md")
	if err := os.WriteFile(marker, []byte("somebody else's copy\n"), 0o644); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("install: %v", err)
	}
	source, err := os.ReadFile(marker)
	if err != nil || !strings.Contains(string(source), "somebody else") {
		t.Errorf("the existing directory was overwritten: %v", err)
	}
	// Nothing else in that directory was linked either: it is reported as one
	// skipped directory, not partially converted.
	if _, err := os.Lstat(filepath.Join(claude, "olares-market")); !os.IsNotExist(err) {
		t.Error("install linked some skills into a directory it had to skip")
	}
}

// A few agents keep their skills one level deeper. Missing those would mean
// silently not installing for them, which is why the scan looks two levels
// down rather than matching a list of names.
func TestNestedAgentDirectoriesAreFound(t *testing.T) {
	home := sandboxHome(t)
	nested := []string{
		filepath.Join(home, ".codeium", "windsurf", "skills"),
		filepath.Join(home, ".config", "crush", "skills"),
		filepath.Join(home, ".pi", "agent", "skills"),
	}
	for _, dir := range nested {
		mkdirAll(t, dir)
	}
	if err := runInstall(quiet(), false); err != nil {
		t.Fatalf("install: %v", err)
	}
	for _, dir := range nested {
		if _, err := os.Lstat(filepath.Join(dir, "olares-shared")); err != nil {
			t.Errorf("%s was not installed to: %v", dir, err)
		}
	}
}

func listing(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var lines []string
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join(dir, entry.Name()))
		if err != nil {
			target = "(not a link)"
		}
		lines = append(lines, entry.Name()+" -> "+target)
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}
