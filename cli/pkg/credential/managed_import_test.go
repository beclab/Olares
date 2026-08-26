package credential

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

type importHarness struct {
	*managedImporter
	store  *fakeStore
	stderr *bytes.Buffer

	calls        int
	refreshErr   error
	seenCurrent  string
	seenRefresh  string
	tokenToStore string
}

// newImportHarness gives each test its own config dir and a refresh stub that
// records what it was handed and, on success, persists what RefreshWith would.
func newImportHarness(t *testing.T, cred *ManagedCredential) *importHarness {
	t.Helper()
	t.Setenv("OLARES_CLI_HOME", t.TempDir())

	h := &importHarness{
		store:        newFakeStore(),
		stderr:       &bytes.Buffer{},
		tokenToStore: "AT-fresh",
	}
	h.managedImporter = &managedImporter{
		load: func() (*ManagedCredential, bool) {
			if cred == nil {
				return nil, false
			}
			return cred, true
		},
		store:  h.store,
		stderr: h.stderr,
		refresh: func(_ context.Context, olaresID, _, currentAccessToken, refreshToken string, _ bool) (string, error) {
			h.calls++
			h.seenCurrent = currentAccessToken
			h.seenRefresh = refreshToken
			if h.refreshErr != nil {
				return "", h.refreshErr
			}
			_ = h.store.Set(auth.StoredToken{
				OlaresID:    olaresID,
				AccessToken: h.tokenToStore,
				Managed:     true,
			})
			return h.tokenToStore, nil
		},
	}
	return h
}

func loadProfile(t *testing.T, olaresID string) *cliconfig.ProfileConfig {
	t.Helper()
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return cfg.FindByOlaresID(olaresID)
}

func testCredential() *ManagedCredential {
	return &ManagedCredential{RefreshToken: "mounted-RT", OlaresID: managedID, AppName: "lares"}
}

// A host install has no mount: nothing is created and nothing is called.
func TestImport_NoCredentialIsANoOp(t *testing.T) {
	h := newImportHarness(t, nil)
	h.run(context.Background())

	if h.calls != 0 {
		t.Fatalf("refresh calls = %d, want 0", h.calls)
	}
	if p := loadProfile(t, managedID); p != nil {
		t.Fatalf("profile %+v created out of nothing", p)
	}
}

// The first command in a fresh container creates the entry and exchanges the
// mounted token immediately, so `profile list` has a definite answer.
func TestImport_FirstRunCreatesAndRefreshes(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.run(context.Background())

	if h.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", h.calls)
	}
	if h.seenRefresh != "mounted-RT" {
		t.Errorf("refresh saw token %q, want the mounted one", h.seenRefresh)
	}
	p := loadProfile(t, managedID)
	if p == nil || !p.Managed || p.AppName != "lares" {
		t.Fatalf("profile = %+v, want a managed entry for lares", p)
	}
	stored, err := h.store.Get(managedID)
	if err != nil || stored.AccessToken != "AT-fresh" {
		t.Fatalf("stored = %+v, err = %v", stored, err)
	}
}

// Every command after the first is local only.
func TestImport_SecondRunTouchesNoNetwork(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.run(context.Background())
	h.run(context.Background())

	if h.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1 — only the first import goes to the network", h.calls)
	}
}

// The app an existing managed entry names is updated in place, still without
// a round trip: the pod was reused for a different application.
func TestImport_AppNameIsUpdatedLocally(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.run(context.Background())

	h.load = func() (*ManagedCredential, bool) {
		return &ManagedCredential{RefreshToken: "mounted-RT", OlaresID: managedID, AppName: "wise"}, true
	}
	h.run(context.Background())

	if h.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", h.calls)
	}
	if p := loadProfile(t, managedID); p == nil || p.AppName != "wise" {
		t.Fatalf("profile = %+v, want appName wise", p)
	}
}

// Taking over a hand-made profile keeps how to reach the instance and
// replaces only who vouches for the identity.
func TestImport_TakeoverKeepsConnectionSettings(t *testing.T) {
	h := newImportHarness(t, testCredential())
	seed := &cliconfig.MultiProfileConfig{}
	seed.Upsert(cliconfig.ProfileConfig{
		Name:               "dev",
		OlaresID:           managedID,
		AuthURLOverride:    "https://auth.dev.example",
		LocalURLPrefix:     "dev.",
		InsecureSkipVerify: true,
	})
	if err := cliconfig.SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h.run(context.Background())

	p := loadProfile(t, managedID)
	if p == nil {
		t.Fatal("profile disappeared")
	}
	if !p.Managed || p.AppName != "lares" {
		t.Errorf("profile = %+v, want it flipped to managed", p)
	}
	if p.Name != "dev" || p.AuthURLOverride != "https://auth.dev.example" ||
		p.LocalURLPrefix != "dev." || !p.InsecureSkipVerify {
		t.Errorf("profile = %+v, want the connection settings preserved", p)
	}
}

// The user's refresh token must not survive the takeover: nothing would ever
// read it and `profile remove` refuses to touch a managed entry.
func TestImport_TakeoverDropsTheUsersRefreshToken(t *testing.T) {
	h := newImportHarness(t, testCredential())
	seed := &cliconfig.MultiProfileConfig{}
	seed.Upsert(cliconfig.ProfileConfig{OlaresID: managedID})
	if err := cliconfig.SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "user-AT", RefreshToken: "user-RT"})

	h.run(context.Background())

	if h.seenCurrent != "user-AT" {
		t.Errorf("refresh saw current token %q, want the stored one so it does not short-circuit", h.seenCurrent)
	}
	stored, err := h.store.Get(managedID)
	if err != nil {
		t.Fatalf("nothing stored: %v", err)
	}
	if stored.RefreshToken != "" {
		t.Errorf("stored refresh token = %q, want the user's copy gone", stored.RefreshToken)
	}
	if !stored.Managed {
		t.Error("stored entry should be managed after the takeover")
	}
}

// A transient failure still creates the entry — that is what makes the
// difference between "pending" and "never" visible — and leaves no token.
func TestImport_TransientRefreshFailureStillCreatesTheEntry(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.refreshErr = errors.New("dial tcp: connection refused")

	h.run(context.Background())

	p := loadProfile(t, managedID)
	if p == nil || !p.Managed {
		t.Fatalf("profile = %+v, want a managed entry despite the failure", p)
	}
	if _, err := h.store.Get(managedID); !errors.Is(err, auth.ErrTokenNotFound) {
		t.Errorf("a failed exchange must not leave a token behind (err = %v)", err)
	}
}

// A dead grant keeps its marker so `profile list` can report invalidated
// rather than a state that reads as "you never logged in".
func TestImport_DeadGrantKeepsTheInvalidationMarker(t *testing.T) {
	h := newImportHarness(t, testCredential())
	seed := &cliconfig.MultiProfileConfig{}
	seed.Upsert(cliconfig.ProfileConfig{OlaresID: managedID})
	if err := cliconfig.SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	invalidatedAt := time.Now()
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "user-AT", RefreshToken: "user-RT"})
	h.refreshErr = &ErrTokenInvalidated{OlaresID: managedID, InvalidatedAt: invalidatedAt}
	// The refresher stamps the store before returning this error.
	_ = h.store.MarkInvalidated(managedID, invalidatedAt)

	h.run(context.Background())

	stored, err := h.store.Get(managedID)
	if err != nil {
		t.Fatalf("the marker was dropped: %v", err)
	}
	if stored.InvalidatedAt != invalidatedAt.UnixMilli() {
		t.Errorf("InvalidatedAt = %d, want %d", stored.InvalidatedAt, invalidatedAt.UnixMilli())
	}
	if stored.RefreshToken != "" || !stored.Managed {
		t.Errorf("stored = %+v, want a managed shell with no refresh token", stored)
	}
}

// The same on a fresh container, where the refresher's own stamp has no entry
// to land on. Without a shell to carry it, the refusal would be invisible and
// every command would go out and collect the same 401.
func TestImport_RefusedGrantIsRecordedOnAFreshContainer(t *testing.T) {
	h := newImportHarness(t, testCredential())
	invalidatedAt := time.Now()
	h.refreshErr = &ErrTokenInvalidated{OlaresID: managedID, InvalidatedAt: invalidatedAt}

	h.run(context.Background())

	stored, err := h.store.Get(managedID)
	if err != nil {
		t.Fatalf("nothing recorded the refusal: %v", err)
	}
	if stored.InvalidatedAt != invalidatedAt.UnixMilli() || !stored.Managed {
		t.Errorf("stored = %+v, want a managed shell marked at %d", stored, invalidatedAt.UnixMilli())
	}
}

// A transient failure records nothing, which is what keeps the next command
// retrying rather than reporting a grant nothing has proven dead.
func TestImport_TransientFailureLeavesNoMarker(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.refreshErr = errors.New("dial tcp: connection refused")
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "user-AT", RefreshToken: "user-RT"})

	h.run(context.Background())

	if _, err := h.store.Get(managedID); !errors.Is(err, auth.ErrTokenNotFound) {
		t.Errorf("store holds %v, want the superseded entry gone and no marker in its place", err)
	}
}

// A first-factor grant creates nothing at all, and says so out loud: an entry
// that 401s on every command is worse than no entry.
func TestImport_FirstFactorGrantCreatesNothing(t *testing.T) {
	h := newImportHarness(t, testCredential())
	h.refreshErr = &ErrManagedNotSecondFactor{OlaresID: managedID}

	h.run(context.Background())

	if p := loadProfile(t, managedID); p != nil {
		t.Fatalf("profile %+v should not exist", p)
	}
	if h.stderr.Len() == 0 {
		t.Error("the contract breach should be reported on stderr")
	}
}

// A first-factor grant must also not destroy a login the user made by hand.
func TestImport_FirstFactorGrantLeavesTheUsersTokenAlone(t *testing.T) {
	h := newImportHarness(t, testCredential())
	seed := &cliconfig.MultiProfileConfig{}
	seed.Upsert(cliconfig.ProfileConfig{OlaresID: managedID})
	if err := cliconfig.SaveMultiProfileConfig(seed); err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = h.store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "user-AT", RefreshToken: "user-RT"})
	h.refreshErr = &ErrManagedNotSecondFactor{OlaresID: managedID}

	h.run(context.Background())

	stored, err := h.store.Get(managedID)
	if err != nil {
		t.Fatalf("the user's token was removed: %v", err)
	}
	if stored.RefreshToken != "user-RT" {
		t.Errorf("stored refresh token = %q, want it untouched", stored.RefreshToken)
	}
	if p := loadProfile(t, managedID); p == nil || p.Managed {
		t.Errorf("profile = %+v, want the user's own entry left as it was", p)
	}
}

// An unwritable config dir is a warning, not a failure: every other verb has
// to keep working.
func TestImport_UnwritableConfigWarnsAndContinues(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	h := newImportHarness(t, testCredential())
	// A read-only parent: config.json can be looked for but not created,
	// which is what a container whose HOME is not writable looks like.
	parent := t.TempDir()
	if err := os.Chmod(parent, 0o500); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
	t.Setenv("OLARES_CLI_HOME", filepath.Join(parent, "olares-cli"))

	h.run(context.Background())

	if h.stderr.Len() == 0 {
		t.Error("a failed config write should warn on stderr")
	}
}

// A mount naming something that is not a parseable Olares ID is skipped
// rather than turned into a profile no URL can be derived from.
func TestImport_UnparseableOlaresIDIsSkipped(t *testing.T) {
	const bogus = "alice@olares@com"
	h := newImportHarness(t, &ManagedCredential{
		RefreshToken: "mounted-RT",
		OlaresID:     bogus,
		AppName:      "lares",
	})

	h.run(context.Background())

	if h.calls != 0 {
		t.Errorf("refresh calls = %d, want 0", h.calls)
	}
	if p := loadProfile(t, bogus); p != nil {
		t.Errorf("profile %+v should not exist", p)
	}
}
