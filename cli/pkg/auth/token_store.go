package auth

import (
	"errors"
	"time"
)

// StoredToken is the per-olaresId record persisted by the CLI.
//
// Phase 2 backend: the entire StoredToken is JSON-serialized and stored as a
// single keychain entry (service=keychain.OlaresCliService, account=olaresId).
// The keychain backend is OS-specific — see cli/internal/keychain — and on
// every supported OS the value lands encrypted at rest. Phase 1's plaintext
// ~/.olares-cli/tokens.json is gone.
//
// There is intentionally NO `ExpiresAt` field: AccessToken is a JWT and the
// only authoritative expiry comes from decoding its `exp` claim via
// auth.ExpiresAt. Mirroring the server's `expires_in` here would just create
// a second source of truth that can drift.
//
// RefreshToken is stored verbatim. It is not necessarily a JWT, so we never
// attempt to decode it.
//
// InvalidatedAt encodes server-side grant invalidation discovered by the
// client (e.g. /api/refresh returning 401/403). 0 means valid (or expiry
// has not yet been "discovered"); any value > 0 marks the entire grant
// (access_token + refresh_token) as unusable, even if the JWT's `exp` is
// still in the future. Phase 1 only DEFINES this field — no code path
// writes it. Phase 2's refreshWithLock will write it. The only way to
// clear it back to 0 is a successful `profile login` / `profile import`
// (Set() defensively zeroes it).
type StoredToken struct {
	OlaresID      string `json:"olaresId"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken,omitempty"`
	SessionID     string `json:"sessionId,omitempty"`
	GrantedAt     int64  `json:"grantedAt,omitempty"`     // unix milliseconds, audit-only
	InvalidatedAt int64  `json:"invalidatedAt,omitempty"` // unix milliseconds; 0 = valid
}

// TokenStore abstracts the per-olaresId secret backend. Phase 2's only
// production implementation is keychainStore (cli/pkg/auth/token_store_keychain.go);
// tests can supply their own via NewTokenStoreWith.
//
// MarkInvalidated stamps an existing entry's InvalidatedAt without touching
// other fields. Returns ErrTokenNotFound if no entry exists for olaresID.
// Phase 2's refreshWithLock calls this when /api/refresh returns 401/403.
type TokenStore interface {
	Get(olaresID string) (*StoredToken, error)
	Set(token StoredToken) error
	Delete(olaresID string) error
	List() ([]StoredToken, error)
	MarkInvalidated(olaresID string, at time.Time) error
}

// ErrTokenNotFound is returned when no token is stored for a given olaresId.
var ErrTokenNotFound = errors.New("token not found")
