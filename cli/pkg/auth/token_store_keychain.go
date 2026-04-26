package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/beclab/Olares/cli/internal/keychain"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// listWarnSink receives one warning per profile whose token couldn't be
// decoded during List(). Defaults to os.Stderr; tests overwrite it to capture
// output without printing during `go test`.
var listWarnSink io.Writer = os.Stderr

// ProfileLister enumerates the olaresIds the CLI knows about. The keychain
// backend can't enumerate its own contents (that's a deliberate property of
// every OS-keychain API: no globbing across accounts), so List() needs an
// external index. cliconfig.MultiProfileConfig already serves that purpose.
//
// The interface is deliberately tiny so tests can swap it out without
// pulling in the whole config layer, and so we avoid widening pkg/auth's
// coupling to cliconfig beyond the single function it actually needs.
type ProfileLister interface {
	ListOlaresIDs() ([]string, error)
}

// cliconfigProfileLister is the production ProfileLister backed by
// cliconfig.LoadMultiProfileConfig. A missing config file is treated as
// "no profiles" (no error) so List() works on a clean machine.
type cliconfigProfileLister struct{}

func (cliconfigProfileLister) ListOlaresIDs() ([]string, error) {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(cfg.Profiles))
	for _, p := range cfg.Profiles {
		out = append(out, p.OlaresID)
	}
	return out, nil
}

// keychainStore implements TokenStore on top of an OS keychain. Each profile
// gets exactly one keychain entry (service=OlaresCliService, account=olaresId)
// whose value is the JSON-encoded StoredToken. We chose JSON-blob-per-account
// (rather than splitting access/refresh into separate entries) so:
//   - the entry is atomic — no partial-write window where access exists but
//     refresh is missing,
//   - rotating any field (SessionID, GrantedAt, InvalidatedAt) is one write,
//   - moving fields around in StoredToken doesn't require re-keying anything,
//   - List() only needs one keychain Get per profile, not two or three.
type keychainStore struct {
	kc       keychain.KeychainAccess
	profiles ProfileLister
}

// NewTokenStore returns the production TokenStore: keychain.Default()
// backend + cliconfig-driven profile enumeration.
func NewTokenStore() TokenStore {
	return &keychainStore{
		kc:       keychain.Default(),
		profiles: cliconfigProfileLister{},
	}
}

// NewTokenStoreWith is the test seam: pass any KeychainAccess + ProfileLister.
// Production code should call NewTokenStore.
func NewTokenStoreWith(kc keychain.KeychainAccess, profiles ProfileLister) TokenStore {
	return &keychainStore{kc: kc, profiles: profiles}
}

// Get returns the StoredToken for olaresID. The keychain backend signals
// "not present" by returning ("", nil); we promote that to ErrTokenNotFound
// so callers can use errors.Is uniformly across the interface.
func (s *keychainStore) Get(olaresID string) (*StoredToken, error) {
	if olaresID == "" {
		return nil, errors.New("olaresID is required")
	}
	raw, err := s.kc.Get(keychain.OlaresCliService, olaresID)
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, ErrTokenNotFound
	}
	var tok StoredToken
	if err := json.Unmarshal([]byte(raw), &tok); err != nil {
		return nil, fmt.Errorf("decode stored token for %s: %w", olaresID, err)
	}
	// Defense in depth: a corrupted blob that decodes but lost its OlaresID
	// shouldn't masquerade as a different account's grant.
	if tok.OlaresID == "" {
		tok.OlaresID = olaresID
	}
	return &tok, nil
}

// Set persists a fresh grant for token.OlaresID, overwriting any previous
// value. As with the historical fileStore, we defensively zero InvalidatedAt
// so a fresh write clears any prior invalidation stamp even if the caller
// forgot to.
func (s *keychainStore) Set(token StoredToken) error {
	if token.OlaresID == "" {
		return errors.New("StoredToken.OlaresID is required")
	}
	token.InvalidatedAt = 0
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("encode stored token: %w", err)
	}
	return s.kc.Set(keychain.OlaresCliService, token.OlaresID, string(data))
}

// MarkInvalidated stamps an existing entry as unusable. Read-modify-write,
// not atomic across processes; that's acceptable because (a) keychain writes
// are racy by nature and (b) the worst-case is a stale stamp that another
// process immediately re-clears with a successful Set.
func (s *keychainStore) MarkInvalidated(olaresID string, at time.Time) error {
	tok, err := s.Get(olaresID)
	if err != nil {
		return err
	}
	tok.InvalidatedAt = at.UnixMilli()
	data, err := json.Marshal(tok)
	if err != nil {
		return fmt.Errorf("encode stored token: %w", err)
	}
	return s.kc.Set(keychain.OlaresCliService, olaresID, string(data))
}

// Delete removes the keychain entry for olaresID.
//
// Behavior contract (callers MUST handle this):
//
//   - If no entry exists for olaresID, Delete returns ErrTokenNotFound. This
//     is intentionally NOT a no-op at our layer — it gives callers a typed
//     signal so flows like `profile remove` can distinguish "deleted" from
//     "wasn't there to begin with" and decide whether to print a warning.
//     Callers that prefer no-op semantics should filter via errors.Is(err,
//     ErrTokenNotFound) (see cmd/ctl/profile/remove.go).
//   - The underlying keychain.Remove IS a no-op on missing entries, but we
//     gate on a prior Get so that a successful return guarantees an entry
//     was actually present and is now gone.
//   - Any backend error from Get / Remove (corrupted blob, locked keychain,
//     transient permission failure) is surfaced verbatim — Delete does NOT
//     swallow them.
func (s *keychainStore) Delete(olaresID string) error {
	if olaresID == "" {
		return errors.New("olaresID is required")
	}
	if _, err := s.Get(olaresID); err != nil {
		return err
	}
	return s.kc.Remove(keychain.OlaresCliService, olaresID)
}

// List returns every StoredToken whose olaresId appears in the config-side
// profile index. Profiles without a stored token are silently skipped — that
// matches the user-visible model (a profile can exist before it has been
// authenticated).
//
// Per-entry read failures (corrupted blob, decode error, transient keychain
// access denial) are degraded to a single stderr warning and the offending
// entry is skipped. Aborting the entire list on one bad blob would make
// `profile list` useless across all profiles whenever a single keychain
// record gets damaged — surfacing the rest is strictly more useful.
//
// Only an error from the ProfileLister itself (i.e. we couldn't even find
// out which olaresIds to look up) is returned to the caller.
func (s *keychainStore) List() ([]StoredToken, error) {
	ids, err := s.profiles.ListOlaresIDs()
	if err != nil {
		return nil, fmt.Errorf("enumerate profiles: %w", err)
	}
	out := make([]StoredToken, 0, len(ids))
	for _, id := range ids {
		tok, err := s.Get(id)
		if err != nil {
			if errors.Is(err, ErrTokenNotFound) {
				continue
			}
			fmt.Fprintf(listWarnSink,
				"warning: skipping stored token for %s: %v\n", id, err)
			continue
		}
		out = append(out, *tok)
	}
	return out, nil
}
