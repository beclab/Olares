// Package activeusers maintains a process-local cache of "activated"
// users so callers (notably the NATS event fan-out for v3 / shared apps)
// can answer "who is currently activated?" without hitting the kube API
// on every event.
//
// A user is considered activated when BOTH of the following hold:
//   - annotation  bytetrade.io/wizard-status == "completed"
//   - status.state                          == "Created"
//
// The cache is populated and kept in sync by the UserController's
// informer event handler (see controllers/user_controller.go). All
// public functions are safe to call from multiple goroutines.
package activeusers

import (
	"sync"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
)

const (
	// WizardStatusAnnotation marks the wizard activation progress on a User.
	WizardStatusAnnotation = "bytetrade.io/wizard-status"
	// WizardStatusCompleted is the wizard-status value that means the
	// activation flow has finished successfully.
	WizardStatusCompleted = "completed"
	// UserStateCreated is the iam User.Status.State value indicating
	// the user record has been fully created by the user controller.
	UserStateCreated = "Created"
)

// store is the single shared cache instance backing the package-level
// API. It is intentionally not exposed so callers cannot bypass the
// locking discipline.
var store = newCache()

type cache struct {
	// mu guards `users`. RWMutex is used because reads (List) are
	// expected to be far more frequent than writes (a handful of
	// User events per minute at most, vs. potentially many event
	// publishes per second when apps are reconciling).
	mu sync.RWMutex
	// users holds the *names* of currently-activated users. A map is
	// used so Upsert/Delete are O(1) and idempotent. The empty-struct
	// value type costs zero bytes per entry.
	users map[string]struct{}
}

func newCache() *cache {
	return &cache{users: make(map[string]struct{})}
}

// IsActivated reports whether the given User object would be considered
// activated. Safe to call without any lock — it only reads fields on the
// passed-in object.
func IsActivated(u *iamv1alpha2.User) bool {
	if u == nil {
		return false
	}
	if u.Annotations[WizardStatusAnnotation] != WizardStatusCompleted {
		return false
	}
	if string(u.Status.State) != UserStateCreated {
		return false
	}
	return true
}

// Upsert reconciles the cache with the given user's current state.
// If the user is activated it is added to the set; otherwise it is
// removed. The combined add/remove behavior keeps a single call site
// in the controller correct whether the user just became activated or
// just lost activation (e.g. status flipped to "Failed").
func Upsert(u *iamv1alpha2.User) {
	if u == nil {
		return
	}
	name := u.GetName()
	if name == "" {
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if IsActivated(u) {
		store.users[name] = struct{}{}
	} else {
		delete(store.users, name)
	}
}

// Delete removes the named user from the cache. Idempotent — a missing
// entry is a no-op. Use this from the informer DeleteFunc since a fully
// deleted User CR is unconditionally no longer activated, regardless of
// its last-seen status fields.
func Delete(name string) {
	if name == "" {
		return
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.users, name)
}

// List returns a snapshot of currently-activated user names. The
// returned slice is a fresh copy: callers may iterate, sort, or hold on
// to it without keeping the cache locked, and without observing later
// concurrent mutations. Order is undefined.
func List() []string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	out := make([]string, 0, len(store.users))
	for name := range store.users {
		out = append(out, name)
	}
	return out
}

// Len returns the current number of activated users. Useful for
// diagnostics / metrics without paying the allocation cost of List.
func Len() int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return len(store.users)
}
