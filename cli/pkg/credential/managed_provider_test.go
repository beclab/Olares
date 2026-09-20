package credential

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

func managedProfile() *cliconfig.ProfileConfig {
	return &cliconfig.ProfileConfig{OlaresID: managedID, Managed: true, AppName: "lares"}
}

// jwtWithExp builds a token the exp decoder accepts. The signature is not
// checked anywhere in the CLI — only the payload's exp claim is read.
func jwtWithExp(exp time.Time) string {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	return enc(`{"alg":"HS512"}`) + "." + enc(fmt.Sprintf(`{"exp":%d}`, exp.Unix())) + ".sig"
}

var (
	storedToken    = jwtWithExp(time.Now().Add(time.Hour))
	recoveredToken = jwtWithExp(time.Now().Add(2 * time.Hour))
)

type providerHarness struct {
	*ManagedProvider
	store *fakeStore

	calls      int
	refreshErr error
}

func newProviderHarness(t *testing.T, cred *ManagedCredential) *providerHarness {
	t.Helper()
	h := &providerHarness{store: newFakeStore()}
	h.ManagedProvider = &ManagedProvider{
		store: h.store,
		load: func() (*ManagedCredential, bool) {
			if cred == nil {
				return nil, false
			}
			return cred, true
		},
		refresh: func(_ context.Context, olaresID, _, _, _ string, _ bool) (string, error) {
			h.calls++
			if h.refreshErr != nil {
				return "", h.refreshErr
			}
			_ = h.store.Set(auth.StoredToken{OlaresID: olaresID, AccessToken: recoveredToken, Managed: true})
			return recoveredToken, nil
		},
	}
	return h
}

// A profile the user logged into by hand is not this provider's to answer.
func TestManagedProvider_DeclinesNonManagedProfile(t *testing.T) {
	h := newProviderHarness(t, testCredential())
	rp, err := h.Resolve(context.Background(), &cliconfig.ProfileConfig{OlaresID: managedID})
	if err != nil || rp != nil {
		t.Fatalf("Resolve = (%v, %v), want (nil, nil)", rp, err)
	}
}

func TestManagedProvider_UsesTheStoredToken(t *testing.T) {
	h := newProviderHarness(t, testCredential())
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: storedToken, Managed: true})

	rp, err := h.Resolve(context.Background(), managedProfile())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if rp.AccessToken != storedToken {
		t.Errorf("AccessToken = %q, want the stored one", rp.AccessToken)
	}
	if h.calls != 0 {
		t.Errorf("refresh calls = %d, want 0 when a token is already stored", h.calls)
	}
	if !rp.Managed || rp.AppName != "lares" {
		t.Errorf("resolved = %+v, want the managed markers carried through", rp)
	}
}

// The startup import may have been unable to reach the auth service — a
// container routinely comes up before the cluster does. The next command
// exchanges the mounted token rather than sending an empty one.
func TestManagedProvider_ColdStartExchangesTheMountedToken(t *testing.T) {
	h := newProviderHarness(t, testCredential())

	rp, err := h.Resolve(context.Background(), managedProfile())
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if rp.AccessToken != recoveredToken {
		t.Errorf("AccessToken = %q, want the freshly exchanged one", rp.AccessToken)
	}
	if h.calls != 1 {
		t.Errorf("refresh calls = %d, want 1", h.calls)
	}
}

// Without a mount there is no token to be had and nothing local that would
// produce one.
func TestManagedProvider_NoMountReportsNoToken(t *testing.T) {
	h := newProviderHarness(t, nil)

	_, err := h.Resolve(context.Background(), managedProfile())
	var notLoggedIn *ErrNotLoggedIn
	if !errors.As(err, &notLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
	if h.calls != 0 {
		t.Errorf("refresh calls = %d, want 0", h.calls)
	}
}

// A mount naming a different user is as good as no mount: the profile's
// identity is the one that has to be vouched for.
func TestManagedProvider_MountForAnotherUserIsRefused(t *testing.T) {
	h := newProviderHarness(t, &ManagedCredential{
		RefreshToken: "mounted-RT",
		OlaresID:     "bob@olares.com",
		AppName:      "lares",
	})

	_, err := h.Resolve(context.Background(), managedProfile())
	var notLoggedIn *ErrNotLoggedIn
	if !errors.As(err, &notLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
}

// A grant the server refused is not retried on every command. Recovery means
// the platform issuing a new one, which arrives with a new pod.
func TestManagedProvider_InvalidatedGrantIsNotRetried(t *testing.T) {
	h := newProviderHarness(t, testCredential())
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: storedToken, Managed: true})
	_ = h.store.MarkInvalidated(managedID, time.Now())

	_, err := h.Resolve(context.Background(), managedProfile())
	var invalidated *ErrTokenInvalidated
	if !errors.As(err, &invalidated) {
		t.Fatalf("err = %v, want ErrTokenInvalidated", err)
	}
	if h.calls != 0 {
		t.Errorf("refresh calls = %d, want 0", h.calls)
	}
}

// A failed exchange surfaces as itself, so the caller can tell a dead grant
// from a cluster that is not up yet.
func TestManagedProvider_ExchangeFailurePropagates(t *testing.T) {
	h := newProviderHarness(t, testCredential())
	h.refreshErr = errors.New("dial tcp: connection refused")

	if _, err := h.Resolve(context.Background(), managedProfile()); err == nil {
		t.Fatal("want the exchange failure surfaced")
	}
}

// The dispatch is by profile, not by what happens to be mounted: switching to
// an account the user logged into by hand must keep serving that account.
func TestCredentialProvider_DispatchesOnTheProfile(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	cfg := &cliconfig.MultiProfileConfig{}
	cfg.Upsert(cliconfig.ProfileConfig{OlaresID: managedID, Managed: true, AppName: "lares"})
	cfg.Upsert(cliconfig.ProfileConfig{Name: "mine", OlaresID: "bob@olares.com"})
	cfg.CurrentProfile = "mine"
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	managed := &recordingProvider{name: "managed", token: "AT-managed"}
	local := &recordingProvider{name: "default", token: "AT-local"}
	c := NewCredentialProvider(managed, local)

	rp, err := c.Resolve(context.Background(), "")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if rp.Source != "default" || rp.AccessToken != "AT-local" {
		t.Fatalf("resolved from %q, want the user's own profile", rp.Source)
	}
	if managed.calls != 0 {
		t.Errorf("the managed provider was consulted %d times for a profile it does not own", managed.calls)
	}

	rp, err = c.Resolve(context.Background(), managedID)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if rp.Source != "managed" || rp.AccessToken != "AT-managed" {
		t.Fatalf("resolved from %q, want the managed provider", rp.Source)
	}
}

func TestCredentialProvider_NoProfileAtAll(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	c := NewCredentialProvider(
		&recordingProvider{name: "managed"},
		&recordingProvider{name: "default"},
	)
	if _, err := c.Resolve(context.Background(), ""); !errors.Is(err, ErrNoProfile) {
		t.Fatalf("err = %v, want ErrNoProfile", err)
	}
}

type recordingProvider struct {
	name  string
	token string
	calls int
}

func (p *recordingProvider) Name() string { return p.name }

func (p *recordingProvider) Resolve(_ context.Context, profile *cliconfig.ProfileConfig) (*ResolvedProfile, error) {
	p.calls++
	return &ResolvedProfile{OlaresID: profile.OlaresID, AccessToken: p.token}, nil
}
