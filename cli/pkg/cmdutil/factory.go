// Package cmdutil holds the shared "Factory" that command implementations
// reach into instead of constructing their own clients / loading their own
// config / etc. This is the olares-cli analogue of lark-cli's
// cmdutil.Factory.
//
// Phase 1 keeps the Factory deliberately minimal: lazily-resolved credential
// chain + Bearer-injecting HTTP client. Phase 2 will add automatic token
// refresh inside the same HTTPClient call without changing this surface.
package cmdutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// Factory is the dependency-injection seam for olares-cli commands. Build
// one with NewFactory at the root command level and pass it (or a closure
// that closes over it) into command constructors.
//
// All accessors are lazy and memoized — calling HTTPClient(ctx) multiple
// times reuses the same resolved profile + client.
type Factory struct {
	// ProfileOverride, when non-empty, forces ResolveProfile to look up this
	// profile instead of the currently-selected one. Wired from a global
	// `--profile` / `--olares-id` flag at the root command.
	ProfileOverride string

	// IOStreams (Phase 2) will live here so commands can write to a swappable
	// io.Writer pair. For now we just expose a Stderr getter to avoid a hard
	// dependency churn when we add it.
	Stderr io.Writer

	credentialOnce sync.Once
	credentialErr  error
	credential     *credential.CredentialProvider

	resolveOnce sync.Once
	resolveErr  error
	resolved    *credential.ResolvedProfile

	clientOnce sync.Once
	client     *http.Client
}

// NewFactory builds a fresh Factory. Cheap; intended to be called once per
// process from the root command.
func NewFactory() *Factory {
	return &Factory{}
}

// Credential returns the lazily-constructed credential chain. The chain is
// (EnvProvider, DefaultProvider) — env first so future in-cluster builds
// can pre-empt on-disk config.
func (f *Factory) Credential() (*credential.CredentialProvider, error) {
	f.credentialOnce.Do(func() {
		def, err := credential.NewDefaultProvider()
		if err != nil {
			f.credentialErr = fmt.Errorf("init default credential provider: %w", err)
			return
		}
		f.credential = credential.NewCredentialProvider(
			credential.NewEnvProvider(),
			def,
		)
	})
	return f.credential, f.credentialErr
}

// ResolveProfile returns the active profile fully resolved (URLs + token).
// Memoized; subsequent calls return the same ResolvedProfile.
func (f *Factory) ResolveProfile(ctx context.Context) (*credential.ResolvedProfile, error) {
	f.resolveOnce.Do(func() {
		cred, err := f.Credential()
		if err != nil {
			f.resolveErr = err
			return
		}
		rp, err := cred.Resolve(ctx, f.ProfileOverride)
		if err != nil {
			f.resolveErr = err
			return
		}
		f.resolved = rp
	})
	return f.resolved, f.resolveErr
}

// HTTPClient returns an *http.Client whose RoundTripper transparently injects
// `Authorization: Bearer <accessToken>` on every outbound request. The client
// also honors the active profile's InsecureSkipVerify flag.
//
// Phase 1: the token is fetched once at first call and reused until the
// process exits. If it expires mid-run, requests will start returning 401 —
// the user's recourse is to re-run `profile login` / `profile import`.
//
// Phase 2 will refactor this into a refreshing transport without changing
// the signature.
func (f *Factory) HTTPClient(ctx context.Context) (*http.Client, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	f.clientOnce.Do(func() {
		base := http.DefaultTransport.(*http.Transport).Clone()
		if rp.InsecureSkipVerify {
			base.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
		}
		f.client = &http.Client{
			Timeout:   30 * time.Second,
			Transport: &bearerTransport{base: base, token: rp.AccessToken},
		}
	})
	return f.client, nil
}

// bearerTransport injects Authorization: Bearer <token> on outbound requests.
// It clones the request before mutating headers so the caller's *http.Request
// is left untouched (important when callers retry).
type bearerTransport struct {
	base  http.RoundTripper
	token string
}

func (b *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if b.token == "" {
		return b.base.RoundTrip(req)
	}
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+b.token)
	return b.base.RoundTrip(clone)
}
