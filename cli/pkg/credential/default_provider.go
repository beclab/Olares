package credential

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/clierr"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// DefaultProvider resolves a profile using the local config + keychain token
// store. It implements the standard "ran `profile login` on the same machine"
// scenario.
//
// Resolve checks failure modes in this fixed priority order, each producing
// a distinct error type but the same user-facing CTA ("run profile login"):
//
//  1. profile nil                    → return (nil, nil); orchestrator surfaces ErrNoProfile
//  2. no token stored                → ErrNotLoggedIn
//  3. stored.InvalidatedAt > 0       → ErrTokenInvalidated  (refresher writes this on /api/refresh 401/403)
//  4. otherwise                      → ResolvedProfile, even if the JWT exp is in the past
//
// We deliberately do NOT short-circuit on a stale JWT exp claim: cli/pkg/cmdutil's
// refreshingTransport will trigger /api/refresh on a 401 and retry the
// request transparently. Failing here would deny it the chance to recover.
// The ErrTokenExpired type is kept for backward compatibility (no callers
// today) in case a future flow wants to assert on it.
type DefaultProvider struct {
	store auth.TokenStore
	now   func() time.Time
}

// NewDefaultProvider returns a Provider backed by the keychain-backed token
// store. The error return is preserved from the Phase 1 file-store signature
// so future backends with non-trivial init (e.g. a remote sidecar) can opt
// in without re-touching every caller.
func NewDefaultProvider() (Provider, error) {
	return &DefaultProvider{store: auth.NewTokenStore(), now: time.Now}, nil
}

// Name implements Provider.
func (d *DefaultProvider) Name() string { return "default" }

// ErrNotLoggedIn is returned when a profile exists but has no stored token.
type ErrNotLoggedIn struct {
	OlaresID string

	// Managed and AppName redirect the call to action. `profile login` is
	// the wrong instruction for a platform-issued credential: it is
	// refused, and following it would not help if it were not, because
	// nothing local can mint the grant.
	Managed bool
	AppName string
}

func (e *ErrNotLoggedIn) Error() string {
	if e.Managed {
		return fmt.Sprintf("no access token for %s: %s", e.OlaresID, managedRemedy(e.AppName))
	}
	return fmt.Sprintf("no access token for %s; run: olares-cli profile login --olares-id %s  (or profile import --refresh-token <tok>)", e.OlaresID, e.OlaresID)
}

// The three methods below put this error in the `-o json` envelope
// under a code instead of the message it would otherwise be read out
// of. See the note on ErrNoProfile for why the credential errors are
// the ones classified first.

func (e *ErrNotLoggedIn) ErrorCode() string { return clierr.CodeAuthNotLoggedIn }

// Retryable is false rather than unknown: nothing about running the
// same command again produces a token.
func (e *ErrNotLoggedIn) Retryable() *bool { return &no }

func (e *ErrNotLoggedIn) RecoveryAction() string {
	// A managed credential is minted by the platform, so `profile
	// login` is refused for it and would not help if it were not.
	if e.Managed {
		return managedRecovery(e.AppName)
	}
	return fmt.Sprintf("olares-cli profile login --olares-id %s", e.OlaresID)
}

// managedRemedy is the tail shared by every managed-credential failure. The
// application is named when we know it, because a user who finds an account
// they never logged into has no other way to tell which install to repair.
func managedRemedy(appName string) string {
	if appName == "" {
		return "this credential is issued by the platform; reinstall or repair the application that requested it to have it re-issued"
	}
	return fmt.Sprintf("this credential is issued by the platform for application %q; reinstall or repair that application to have it re-issued", appName)
}

// ErrTokenExpired is retained for backward compatibility but is no longer
// produced by Resolve — refreshingTransport handles expiry transparently
// via /api/refresh. Kept so any future flow (or external consumer of the
// credential package) can still assert on it without API churn.
type ErrTokenExpired struct {
	OlaresID  string
	ExpiredAt time.Time
}

func (e *ErrTokenExpired) Error() string {
	return fmt.Sprintf("access token for %s expired at %s; please run: olares-cli profile login --olares-id %s  (or profile import --olares-id %s --refresh-token <tok>)",
		e.OlaresID, e.ExpiredAt.Format(time.RFC3339), e.OlaresID, e.OlaresID)
}

func (e *ErrTokenExpired) ErrorCode() string { return clierr.CodeAuthTokenExpired }
func (e *ErrTokenExpired) Retryable() *bool  { return &no }
func (e *ErrTokenExpired) RecoveryAction() string {
	return fmt.Sprintf("olares-cli profile login --olares-id %s", e.OlaresID)
}

// ErrTokenInvalidated is returned when a stored token has been explicitly
// marked unusable via TokenStore.MarkInvalidated. The grant cannot be
// recovered locally — the user must re-authenticate.
//
// Refresher.Refresh stamps InvalidatedAt when /api/refresh returns 401/403,
// so subsequent commands skip the network round-trip and surface this CTA
// directly.
type ErrTokenInvalidated struct {
	OlaresID      string
	InvalidatedAt time.Time

	// Managed and AppName redirect the call to action; see ErrNotLoggedIn.
	Managed bool
	AppName string
}

func (e *ErrTokenInvalidated) Error() string {
	if e.Managed {
		return fmt.Sprintf("the credential for %s was rejected at %s: %s",
			e.OlaresID, e.InvalidatedAt.Format(time.RFC3339), managedRemedy(e.AppName))
	}
	return fmt.Sprintf("refresh token for %s became invalid at %s; please run: olares-cli profile login --olares-id %s  (or profile import --olares-id %s --refresh-token <tok>)",
		e.OlaresID, e.InvalidatedAt.Format(time.RFC3339), e.OlaresID, e.OlaresID)
}

func (e *ErrTokenInvalidated) ErrorCode() string { return clierr.CodeAuthTokenInvalidated }

// Retryable is false, and this is the case where saying so is worth
// most: the failure looks transient -- a token that worked an hour ago
// -- and no amount of retrying or refreshing recovers a rejected grant.
func (e *ErrTokenInvalidated) Retryable() *bool { return &no }

func (e *ErrTokenInvalidated) RecoveryAction() string {
	if e.Managed {
		return managedRecovery(e.AppName)
	}
	return fmt.Sprintf("olares-cli profile login --olares-id %s", e.OlaresID)
}

// managedRecovery is the action half of managedRemedy. It is prose
// rather than a command because there is no command: what has to happen
// is a reinstall or repair of the application, which is done where that
// application is managed.
func managedRecovery(appName string) string {
	if appName == "" {
		return "reinstall or repair the application that requested this platform-issued credential"
	}
	return fmt.Sprintf("reinstall or repair application %q, which requested this platform-issued credential", appName)
}

// no is addressable so it can be returned as *bool, which is what keeps
// "we do not know" distinct from "no" in the envelope.
var no = false

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
	// Stale JWTs are NOT rejected here — refreshingTransport will swap them
	// out on a 401. We surface exp purely as a hint on ResolvedProfile.

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
		SettingsURL:        id.SettingsURL(profile.LocalURLPrefix),
		FilesURL:           id.FilesURL(profile.LocalURLPrefix),
		MarketURL:          id.MarketURL(profile.LocalURLPrefix),
		DashboardURL:       id.DashboardURL(profile.LocalURLPrefix),
		ControlHubURL:      id.ControlHubURL(profile.LocalURLPrefix),
		AccessToken:        accessToken,
		InsecureSkipVerify: profile.InsecureSkipVerify,
		Managed:            profile.Managed,
		AppName:            profile.AppName,
	}
	if !exp.IsZero() {
		rp.ExpiresAt = exp.Unix()
	}
	return rp, nil
}
