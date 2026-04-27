// Package lockfile is a tiny wrapper over github.com/gofrs/flock that adds
// context-cancelable acquisition and a single deterministic lock-directory
// layout under cliconfig.Home(). It exists so the rest of the codebase
// doesn't import flock directly — keeping the dependency surface small and
// letting us swap the backend without touching every call site.
//
// The locks are advisory file locks (flock(2) on darwin/linux, LockFileEx on
// windows). Two olares-cli processes contending for the same olaresId's
// refresh slot will serialize through the OS; a single process with multiple
// goroutines should pre-serialize via an in-process mutex BEFORE asking us
// for the file lock (otherwise the file lock is a no-op for them on linux,
// where flock is per-fd).
package lockfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// dirPerm matches cliconfig's 0700 — the lock dir lives next to config.json
// and should not be world-readable.
const dirPerm os.FileMode = 0o700

// pollInterval is how often Acquire retries TryLock while waiting. flock's
// blocking Lock() is not context-cancelable, so we busy-poll TryLock + sleep.
// 100ms matches lark-cli's cadence: short enough that ctx-cancel feels
// instant, long enough not to thrash a contended lock.
const pollInterval = 100 * time.Millisecond

// Acquire takes an exclusive flock on the file at `path`, creating the
// parent directory and the file as needed (mode 0600 for the file, 0700 for
// the directory).
//
// The call blocks until either:
//   - the lock is acquired (returns release, nil), or
//   - ctx is canceled / its deadline expires (returns nil, ctx.Err()).
//
// release() must be called exactly once to drop the lock and close the
// underlying fd. It is safe to call from a defer.
func Acquire(ctx context.Context, path string) (release func() error, err error) {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}
	fl := flock.New(path)
	for {
		ok, err := fl.TryLock()
		if err != nil {
			return nil, fmt.Errorf("lockfile %s: %w", path, err)
		}
		if ok {
			return fl.Unlock, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
			// retry
		}
	}
}

// RefreshLockPath returns the per-olaresId refresh lock path under
// cliconfig.Home()/locks/. The olaresId is sanitized for filesystem use
// (slashes / null / colon / control chars are replaced with '_'); '@' and
// '.' are left intact since they are valid on every supported OS and
// preserve the original olaresId in directory listings, which is useful for
// debugging.
func RefreshLockPath(olaresID string) (string, error) {
	home, err := cliconfig.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "locks", sanitize(olaresID)+".refresh.lock"), nil
}

// sanitize replaces any character that's problematic on at least one of our
// target filesystems with '_'. We deliberately allow '@' and '.' through
// (they are legal on darwin/linux/windows file names) so the resulting file
// name is still a recognizable olaresId.
func sanitize(s string) string {
	if s == "" {
		return "_"
	}
	const bad = "/\\:*?\"<>|\x00"
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 0x20 || strings.ContainsRune(bad, r) {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
