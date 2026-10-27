package cachedir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A host install sets nothing, so every caller keeps its own lookup chain and
// no directory is created behind its back.
func TestBase_UnsetIsNotAManagedContainer(t *testing.T) {
	t.Setenv(EnvCacheDir, "")

	if got, ok := Base(); ok {
		t.Fatalf("Base() = %q, true; want no base", got)
	}
}

// The ordinary managed container: the platform's directory is writable and is
// used as it is.
func TestBase_UsesWritableCacheDir(t *testing.T) {
	dir := writableDir(t)
	t.Setenv(EnvCacheDir, dir)

	got, ok := Base()
	if !ok || got != dir {
		t.Fatalf("Base() = %q, %v; want %q, true", got, ok, dir)
	}
}

// The agent-sandbox case this package exists for: the mount is there, its mode
// bits say world-writable, and the write is refused anyway.
func TestBase_UnwritableCacheDirFallsBackToTemp(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	denied := writableDir(t)
	if err := os.Chmod(denied, 0o555); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(denied, 0o700) })
	tmp := writableDir(t)
	t.Setenv("TMPDIR", tmp)
	t.Setenv(EnvCacheDir, denied)

	got, ok := Base()
	if !ok {
		t.Fatal("Base() found nowhere to write, want the temp fallback")
	}
	if !strings.HasPrefix(got, tmp) {
		t.Fatalf("Base() = %q, want a directory under %q", got, tmp)
	}
	if fi, err := os.Stat(got); err != nil || !fi.IsDir() {
		t.Fatalf("fallback %q is not a usable directory: %v", got, err)
	}
}

// The fallback has to be the same directory next time, or every command would
// re-import the mounted credential and exchange the refresh token again.
func TestBase_FallbackIsStableAcrossResolutions(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	denied := writableDir(t)
	if err := os.Chmod(denied, 0o555); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(denied, 0o700) })
	t.Setenv("TMPDIR", writableDir(t))
	t.Setenv(EnvCacheDir, denied)

	first, _ := Base()
	if second := resolve(denied); second != first {
		t.Fatalf("re-resolved to %q, want the same %q", second, first)
	}
}

// A relative path floats with cwd, which is nobody's intent; it is declined
// rather than joined onto wherever the command happened to run.
func TestBase_RelativeCacheDirIsDeclined(t *testing.T) {
	t.Setenv(EnvCacheDir, "relative-cache")

	if got, ok := Base(); ok {
		t.Fatalf("Base() = %q, true; want the relative path declined", got)
	}
}

// writableDir returns a real directory, resolved through symlinks so a
// comparison against what Base() returns is not defeated by /var -> /private/var.
func writableDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}
	return dir
}
