package credential

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// DefaultProvider resolves a profile using the local config + plaintext token
// store. It implements the standard "ran `profile login` on the same machine"
// scenario.
//
// Resolve checks failure modes in this fixed priority order, each producing
// a distinct error type but the same user-facing CTA ("run profile login"):
//
//  1. profile nil                    → return (nil, nil); orchestrator surfaces ErrNoProfile
//  2. no token stored                → ErrNotLoggedIn
//  3. stored.InvalidatedAt > 0       → ErrTokenInvalidated  (Phase 2 marks this on /api/refresh 401/403)
//  4. JWT exp claim in the past      → ErrTokenExpired
//  5. otherwise                      → ResolvedProfile
//
// Note (3) takes precedence over (4): an explicitly-invalidated grant is
// unusable even if the access_token JWT happens to still have time left,
// because the refresh leg is dead and we cannot get a new one. Phase 1 does
// NOT auto-refresh. Phase 2 will inject token-refresh logic here.
type DefaultProvider struct {
	store auth.TokenStore
	now   func() time.Time
}

// NewDefaultProvider opens the on-disk token store and returns a Provider
// suitable for normal CLI invocations. Returns an error only if the token
// store path itself can't be resolved (which usually means $HOME is broken).
func NewDefaultProvider() (Provider, error) {
	store, err := auth.NewFileStore()
	if err != nil {
		return nil, err
	}
	return &DefaultProvider{store: store, now: time.Now}, nil
}

// Name implements Provider.
func (d *DefaultProvider) Name() string { return "default" }

// ErrNotLoggedIn is returned when a profile exists but has no stored token.
type ErrNotLoggedIn struct {
	OlaresID string
}

func (e *ErrNotLoggedIn) Error() string {
	return fmt.Sprintf("no access token for %s; run: olares-cli profile login --olares-id %s  (or profile import --refresh-token <tok>)", e.OlaresID, e.OlaresID)
}

// ErrTokenExpired is returned when a stored token's JWT `exp` is in the past.
// Phase 1 does not auto-refresh; Phase 2 will catch this internally.
type ErrTokenExpired struct {
	OlaresID  string
	ExpiredAt time.Time
}

func (e *ErrTokenExpired) Error() string {
	return fmt.Sprintf("access token for %s expired at %s; please run: olares-cli profile login --olares-id %s  (or profile import --olares-id %s --refresh-token <tok>)",
		e.OlaresID, e.ExpiredAt.Format(time.RFC3339), e.OlaresID, e.OlaresID)
}

// ErrTokenInvalidated is returned when a stored token has been explicitly
// marked unusable via TokenStore.MarkInvalidated. The grant cannot be
// recovered locally — the user must re-authenticate.
//
// Phase 1 has no code path that writes InvalidatedAt; the only way to hit
// this in Phase 1 is by hand-editing tokens.json. Phase 2's refreshWithLock
// will write InvalidatedAt when /api/refresh returns 401/403.
type ErrTokenInvalidated struct {
	OlaresID      string
	InvalidatedAt time.Time
}

func (e *ErrTokenInvalidated) Error() string {
	return fmt.Sprintf("refresh token for %s became invalid at %s; please run: olares-cli profile login --olares-id %s  (or profile import --olares-id %s --refresh-token <tok>)",
		e.OlaresID, e.InvalidatedAt.Format(time.RFC3339), e.OlaresID, e.OlaresID)
}

// Resolve implements Provider.
func (d *DefaultProvider) Resolve(_ context.Context, profile *cliconfig.ProfileConfig) (*ResolvedProfile, error) {
	if profile == nil {
		return nil, nil
	}

	stored, err := d.store.Get(profile.OlaresID)
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return nil, &ErrNotLoggedIn{OlaresID: profile.OlaresID}
		}
		return nil, fmt.Errorf("read token store: %w", err)
	}

	// Priority 1: explicit invalidation overrides any local heuristic.
	// We can't talk to the server with this grant even if the JWT looks fresh.
	if stored.InvalidatedAt > 0 {
		return nil, &ErrTokenInvalidated{
			OlaresID:      profile.OlaresID,
			InvalidatedAt: time.UnixMilli(stored.InvalidatedAt),
		}
	}

	exp, expErr := auth.ExpiresAt(stored.AccessToken)
	if expErr != nil && !errors.Is(expErr, auth.ErrNoExpClaim) {
		return nil, fmt.Errorf("decode access token: %w", expErr)
	}
	if !exp.IsZero() && !d.now().Before(exp) {
		return nil, &ErrTokenExpired{OlaresID: profile.OlaresID, ExpiredAt: exp}
	}

	return buildResolved(profile, stored.AccessToken, exp)
}

// buildResolved is shared between DefaultProvider and any future provider that
// needs to turn (ProfileConfig, accessToken) into a ResolvedProfile.
func buildResolved(profile *cliconfig.ProfileConfig, accessToken string, exp time.Time) (*ResolvedProfile, error) {
	authURL, err := profile.ResolvedAuthURL()
	if err != nil {
		return nil, fmt.Errorf("derive auth URL: %w", err)
	}
	id, err := olares.ParseID(profile.OlaresID)
	if err != nil {
		return nil, err
	}
	rp := &ResolvedProfile{
		Name:               profile.DisplayName(),
		OlaresID:           profile.OlaresID,
		UserUID:            profile.UserUID,
		AuthURL:            authURL,
		VaultURL:           id.VaultURL(profile.LocalURLPrefix),
		DesktopURL:         id.DesktopURL(profile.LocalURLPrefix),
		AccessToken:        accessToken,
		InsecureSkipVerify: profile.InsecureSkipVerify,
	}
	if !exp.IsZero() {
		rp.ExpiresAt = exp.Unix()
	}
	return rp, nil
}
