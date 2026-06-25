package cliconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// config.json is mutated from several call sites (location switch / outage
// stamps on the network-error path, role + version caching during detect,
// and the user-facing profile add/remove/use verbs). Without serialization,
// two concurrent writers each do load -> mutate -> save and the second save
// silently clobbers the first writer's change (last-writer-wins).
//
// UpdateLocked / updateLockedInto close that race for the writers that route
// through them: they take a cross-process advisory flock, RE-READ config.json
// from disk under the lock, apply the mutation to that fresh copy, and save.
// Re-reading is the crux — it means a writer always builds on the latest
// on-disk state instead of a snapshot captured before the lock was held.
//
// NOTE: the lock only serializes writers that opt in. The high-frequency,
// background-prone writers introduced by location detection use it; the
// low-frequency user-initiated verbs (profile add/remove/use) still write
// directly, matching their pre-existing behavior.

// configLockTimeout bounds how long a single locked update waits to acquire
// the flock before giving up, so a stuck peer can't hang the CLI forever.
const configLockTimeout = 10 * time.Second

// configLockPoll is how often we retry TryLock while waiting. Matches
// lockfile.Acquire's cadence (flock's blocking Lock isn't ctx-cancelable).
const configLockPoll = 100 * time.Millisecond

// errNoConfigChange lets a mutator signal "nothing actually changed, skip the
// disk write". updateLockedInto / UpdateLocked treat it as success with no
// save (keeping no-op clears cheap). It must never escape to callers.
var errNoConfigChange = errors.New("cliconfig: no change")

// ConfigLockPath returns the advisory lock guarding config.json writes, at
// Home()/locks/config.lock. It sits alongside the per-olaresId refresh locks
// under the same locks/ directory.
func ConfigLockPath() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "locks", "config.lock"), nil
}

// withConfigLock runs fn while holding the exclusive config flock, releasing
// it (and closing the fd) on return.
func withConfigLock(fn func() error) error {
	lockPath, err := ConfigLockPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return fmt.Errorf("create lock dir: %w", err)
	}
	fl := flock.New(lockPath)
	ctx, cancel := context.WithTimeout(context.Background(), configLockTimeout)
	defer cancel()
	for {
		ok, err := fl.TryLock()
		if err != nil {
			return fmt.Errorf("config lock %s: %w", lockPath, err)
		}
		if ok {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("acquire config lock: %w", ctx.Err())
		case <-time.After(configLockPoll):
		}
	}
	defer func() { _ = fl.Unlock() }()
	return fn()
}

// UpdateLocked performs a cross-process-safe read-modify-write of config.json:
// it takes the config flock, re-reads the file from disk, applies mutate to
// that fresh copy, and saves. The fresh *MultiProfileConfig is what mutate
// must operate on — any config the caller already holds is intentionally NOT
// used, so concurrent writers don't clobber each other.
//
// A mutate that returns errNoConfigChange (via the apply* helpers) is treated
// as a successful no-op with no disk write.
func UpdateLocked(mutate func(*MultiProfileConfig) error) error {
	return withConfigLock(func() error {
		cfg, err := LoadMultiProfileConfig()
		if err != nil {
			return err
		}
		if err := mutate(cfg); err != nil {
			if errors.Is(err, errNoConfigChange) {
				return nil
			}
			return err
		}
		return SaveMultiProfileConfig(cfg)
	})
}

// updateLockedInto is UpdateLocked for the method-form Set* helpers: same
// lock + re-read + mutate + save, but it also copies the freshly-persisted
// state back into the receiver so a caller that keeps using m sees the latest
// on-disk truth (including this change and any concurrent ones). On a hard
// error the receiver is left untouched.
func (m *MultiProfileConfig) updateLockedInto(mutate func(*MultiProfileConfig) error) error {
	return withConfigLock(func() error {
		fresh, err := LoadMultiProfileConfig()
		if err != nil {
			return err
		}
		if err := mutate(fresh); err != nil {
			if errors.Is(err, errNoConfigChange) {
				*m = *fresh
				return nil
			}
			return err
		}
		if err := SaveMultiProfileConfig(fresh); err != nil {
			return err
		}
		*m = *fresh
		return nil
	})
}
