// Package keychainfake provides a process-local KeychainAccess implementation
// for tests. Putting it in its own subpackage (rather than the main keychain
// package) keeps the production binary free of test-only code while letting
// every consumer of keychain.KeychainAccess share one canonical fake instead
// of redeclaring memKeychain in each test file.
//
// Tests are expected to import this package as keychainfake and call New().
// Because all fields on Fake are exported, individual tests can drive
// failure modes (transient access denial, per-account get errors, set/remove
// failures) without subclassing.
package keychainfake

import (
	"sync"

	"github.com/beclab/Olares/cli/internal/keychain"
)

// Fake is an in-memory KeychainAccess.
//
// Field semantics:
//
//   - GetErr: when non-nil, every Get returns this error (after the
//     PerAccountGetErr table is consulted). Lets a test simulate a fully
//     blocked keychain.
//   - PerAccountGetErr: account → error. Consulted before GetErr so a single
//     test can mark just one account as unreadable while the rest succeed —
//     this is the precise shape of the "List() must tolerate one bad blob"
//     contract in pkg/auth.
//   - SetErr / RmErr: equivalent globals for Set / Remove.
//   - GotKeys: ordered log of accounts queried via Get. Useful for asserting
//     a code path didn't make redundant keychain lookups.
//
// The Mu protects every map / slice so the Fake is safe to share across
// goroutines, mirroring the real OS-keychain backends which are also safe
// for concurrent use.
type Fake struct {
	Mu               sync.Mutex
	Data             map[string]string
	GetErr           error
	PerAccountGetErr map[string]error
	SetErr           error
	RmErr            error
	GotKeys          []string
}

// New returns an empty Fake ready for use as a keychain.KeychainAccess.
func New() *Fake {
	return &Fake{Data: map[string]string{}}
}

// Key composes the storage key from (service, account). Exported so tests
// that want to seed Data directly (e.g. with intentionally corrupt JSON for
// the "corrupted blob" branch) can compute the same composite key the
// production lookups will use.
func (f *Fake) Key(service, account string) string {
	return service + "\x00" + account
}

func (f *Fake) Get(service, account string) (string, error) {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	f.GotKeys = append(f.GotKeys, account)
	if err, ok := f.PerAccountGetErr[account]; ok {
		return "", err
	}
	if f.GetErr != nil {
		return "", f.GetErr
	}
	return f.Data[f.Key(service, account)], nil
}

func (f *Fake) Set(service, account, value string) error {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	if f.SetErr != nil {
		return f.SetErr
	}
	f.Data[f.Key(service, account)] = value
	return nil
}

func (f *Fake) Remove(service, account string) error {
	f.Mu.Lock()
	defer f.Mu.Unlock()
	if f.RmErr != nil {
		return f.RmErr
	}
	delete(f.Data, f.Key(service, account))
	return nil
}

// Compile-time check that Fake satisfies the production interface. If this
// fails to compile after a keychain.KeychainAccess change, every fake-using
// test in the tree breaks at the point of import — exactly what we want.
var _ keychain.KeychainAccess = (*Fake)(nil)
