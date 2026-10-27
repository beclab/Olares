// Package cachedir resolves the base directory olares-cli keeps its derived
// state under — the profile index, the encrypted keychain, and the locks that
// serialize both — inside a container the platform injected a credential into.
//
// app-service mounts an emptyDir and exports $OLARES_CLI_CACHE_DIR at it, and
// both cliconfig.Home and keychain.StorageDir used to take that variable at
// its word. An agent sandbox breaks the promise: dsh confines the shell to the
// session workspace, so the mount is present, world-writable by its mode bits,
// and still refuses every write. What the platform issued then reaches the
// process and evaporates with it, and the next command reports that no profile
// is configured — the one diagnosis that sends a user to `profile login`, which
// is exactly what a platform-issued identity refuses.
//
// So the rung is verified rather than trusted, and a rung that cannot be
// written to is skipped instead of failed on.
package cachedir

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// EnvCacheDir is app-service's contract with the container. The literal is
// duplicated from credential.EnvCacheDir because that package imports
// cliconfig, which imports this one.
const EnvCacheDir = "OLARES_CLI_CACHE_DIR"

// Resolution is memoized per distinct input rather than once per process:
// probing costs a create and an unlink, Home() is called several times per
// command, and keying on the inputs keeps a test that changes them honest
// without a reset hook in the production API.
var (
	mu       sync.Mutex
	cacheKey string
	cached   string
)

// Base returns the writable base directory for platform-injected state, and
// whether there is one at all.
//
// A false second return means this is not a managed container: $OLARES_CLI_CACHE_DIR
// is unset, or nothing derived from it could be written to. Callers keep their
// own lookup chain for that case, so a host install resolves exactly as it did
// before this package existed — including creating no directories, since the
// only way to tell a writable directory from an unwritable one is to write.
//
// When true, the layout under the returned base is the same whether it is the
// platform's directory or the fallback, so a caller joins its own subdirectory
// without asking which one it got.
func Base() (string, bool) {
	configured := os.Getenv(EnvCacheDir)
	key := configured + "\x00" + os.TempDir()

	mu.Lock()
	defer mu.Unlock()
	if key != cacheKey {
		cached = resolve(configured)
		cacheKey = key
	}
	return cached, cached != ""
}

func resolve(configured string) string {
	if configured == "" {
		return ""
	}
	cleaned := filepath.Clean(configured)
	if !filepath.IsAbs(cleaned) {
		// Matches keychain's narrow reading of its own directory variables:
		// a relative path floats with cwd, which is nobody's intent.
		debugf("ignoring relative %s=%q", EnvCacheDir, configured)
		return ""
	}
	if usable(cleaned) {
		return cleaned
	}
	fallback := tempBase()
	if usable(fallback) {
		debugf("%s=%q is not writable; using %q instead", EnvCacheDir, cleaned, fallback)
		return fallback
	}
	debugf("neither %q nor %q is writable", cleaned, fallback)
	return ""
}

// tempBase is stable across invocations on purpose: a random directory would
// make every command re-import the mounted credential and exchange the refresh
// token again. The uid is in the name because /tmp is shared, and a directory
// left by another account would be one this process cannot use.
func tempBase() string {
	name := "olares-cli"
	if uid := os.Getuid(); uid >= 0 {
		name = fmt.Sprintf("olares-cli-%d", uid)
	}
	return filepath.Join(os.TempDir(), name)
}

// usable answers the only question that matters by performing the operation in
// question. A permission probe would not do: Landlock and seccomp-style
// confinement deny the syscall while leaving the mode bits that access(2)
// reports untouched, which is precisely how /olares/cache reads as 0777 and
// rejects mkdir.
func usable(dir string) bool {
	if fi, err := os.Lstat(dir); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return false
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return false
	}
	probe, err := os.CreateTemp(dir, ".writable-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	_ = probe.Close()
	_ = os.Remove(name)
	return true
}

// debugf shares OLARES_CLI_DEBUG with credential.debugManaged and the keychain
// hints: resolution runs before every command, so a rung being skipped is only
// worth saying when somebody is asking.
func debugf(format string, args ...any) {
	if os.Getenv("OLARES_CLI_DEBUG") == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "[olares-cli cachedir] "+format+"\n", args...)
}
