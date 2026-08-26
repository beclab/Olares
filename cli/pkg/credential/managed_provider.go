package credential

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// ManagedProvider resolves a profile whose grant the platform mounted into
// this container. It replaces the env-var stub that predated the contract:
// the credential arrives as a file, not as environment variables, and it is a
// refresh token rather than a ready-to-use access token.
//
// Structurally it is DefaultProvider plus one recovery: when the keychain has
// no usable access token, it exchanges the mounted refresh token on the spot
// instead of handing back an empty one and letting the request 401. That
// happens whenever the startup import could not reach the network — the
// container came up before the cluster's auth service did, which is the
// normal ordering — and the profile would otherwise stay unusable until the
// pod restarted.
//
// An invalidated grant is not retried here. The marker means the server
// refused this exact refresh token, and the way out is the platform issuing a
// new one, which happens when the application is reinstalled or repaired and
// brings a new pod with it.
type ManagedProvider struct {
	store   auth.TokenStore
	load    func() (*ManagedCredential, bool)
	refresh refreshFunc
}

// NewManagedProvider returns the production provider.
func NewManagedProvider() Provider {
	return &ManagedProvider{
		store:   auth.NewTokenStore(),
		load:    LoadManagedCredential,
		refresh: NewRefresher().RefreshWith,
	}
}

// Name implements Provider.
func (p *ManagedProvider) Name() string { return "managed" }

// Resolve implements Provider.
func (p *ManagedProvider) Resolve(ctx context.Context, profile *cliconfig.ProfileConfig) (*ResolvedProfile, error) {
	if profile == nil || !profile.Managed {
		return nil, nil
	}

	stored, err := p.store.Get(profile.OlaresID)
	switch {
	case err == nil && stored.InvalidatedAt > 0:
		return nil, &ErrTokenInvalidated{
			OlaresID:      profile.OlaresID,
			InvalidatedAt: time.UnixMilli(stored.InvalidatedAt),
		}
	case err == nil && stored.AccessToken != "":
		exp, expErr := auth.ExpiresAt(stored.AccessToken)
		if expErr != nil && !errors.Is(expErr, auth.ErrNoExpClaim) {
			return nil, fmt.Errorf("decode access token: %w", expErr)
		}
		// A stale exp is not rejected: the HTTP transport rotates the
		// token on a 401, and refusing here would deny it that chance.
		return buildResolved(profile, stored.AccessToken, exp)
	case err == nil || errors.Is(err, auth.ErrTokenNotFound):
		// Nothing usable stored — fall through to the exchange below.
	default:
		return nil, fmt.Errorf("read token store: %w", err)
	}

	cred, ok := p.load()
	if !ok || cred.OlaresID != profile.OlaresID {
		// The profile says the platform vouches for this identity but the
		// mount is gone or names somebody else, so there is no token to
		// be had and no local action that would produce one.
		return nil, &ErrNotLoggedIn{OlaresID: profile.OlaresID}
	}
	authURL, err := profile.ResolvedAuthURL()
	if err != nil {
		return nil, fmt.Errorf("derive auth URL: %w", err)
	}
	accessToken, err := p.refresh(ctx, profile.OlaresID, authURL, "", cred.RefreshToken, profile.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}
	exp, expErr := auth.ExpiresAt(accessToken)
	if expErr != nil && !errors.Is(expErr, auth.ErrNoExpClaim) {
		return nil, fmt.Errorf("decode access token: %w", expErr)
	}
	return buildResolved(profile, accessToken, exp)
}
