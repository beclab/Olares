package fake

import (
	"context"
	"sync"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// Identity is an in-memory ports.Identity. Programmed via three
// fields on the fake — defaults are conservative (nobody is admin,
// no admin known, no cluster owner). Safe for concurrent read.
type Identity struct {
	mu sync.RWMutex

	// Admins is the set of usernames that IsAdmin returns true for.
	Admins map[string]bool
	// AdminName is what AdminUsername returns.
	AdminName string
	// Owner is what ClusterOwner returns.
	Owner string
}

// SetAdmin toggles admin membership for a single user. Convenient
// for setup lines in table-driven tests.
func (i *Identity) SetAdmin(user string, admin bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.Admins == nil {
		i.Admins = make(map[string]bool)
	}
	if admin {
		i.Admins[user] = true
	} else {
		delete(i.Admins, user)
	}
}

func (i *Identity) IsAdmin(_ context.Context, user string) (bool, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.Admins[user], nil
}

func (i *Identity) AdminUsername(_ context.Context) (string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.AdminName, nil
}

func (i *Identity) ClusterOwner(_ context.Context) (string, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.Owner, nil
}

// Interface conformance guard.
var _ ports.Identity = (*Identity)(nil)
