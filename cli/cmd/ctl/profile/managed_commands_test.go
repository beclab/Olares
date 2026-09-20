package profile

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/internal/keychain/keychainfake"
	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/credential"
)

const managedID = "alice@olares.test"

// renderListing runs `profile list` against a throwaway config and returns
// what it printed. runList writes to an *os.File, so the capture is a file.
func renderListing(t *testing.T, cfg *cliconfig.MultiProfileConfig) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("OLARES_CLI_HOME", home)
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}
	path := filepath.Join(home, "listing")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	if err := runList(context.Background(), nil, false, out); err != nil {
		out.Close()
		t.Fatalf("runList: %v", err)
	}
	out.Close()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read listing: %v", err)
	}
	return string(raw)
}

func managedConfig() *cliconfig.MultiProfileConfig {
	return &cliconfig.MultiProfileConfig{
		Profiles: []cliconfig.ProfileConfig{
			{OlaresID: managedID, Managed: true, AppName: "lares"},
			{OlaresID: "bob@olares.test"},
		},
		CurrentProfile: managedID,
	}
}

// A managed identity cannot be logged into or imported over. Both verbs share
// ensureProfileWritable, and each names itself so the message says which one
// was refused.
func TestLoginAndImportRefuseAManagedIdentity(t *testing.T) {
	for _, verb := range []string{"log in as", "import credentials for"} {
		t.Run(verb, func(t *testing.T) {
			t.Setenv("OLARES_CLI_HOME", t.TempDir())
			store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{managedID})

			_, err := ensureProfileWritable(managedConfig(), store,
				commonCredFlags{olaresID: managedID}, verb, time.Now())

			var managed *credential.ErrManagedProfile
			if !errors.As(err, &managed) {
				t.Fatalf("err = %v, want *credential.ErrManagedProfile", err)
			}
			if managed.AppName != "lares" {
				t.Errorf("AppName = %q, want the application that requested the grant", managed.AppName)
			}
			if !strings.Contains(err.Error(), verb) {
				t.Errorf("message %q does not say which command was refused", err)
			}
		})
	}
}

// The refusal is scoped to the managed identity. Somebody working inside a
// container is still free to log into an account of their own.
func TestLoginIsAllowedForEveryOtherIdentity(t *testing.T) {
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{"bob@olares.test"})

	if _, err := ensureProfileWritable(managedConfig(), store,
		commonCredFlags{olaresID: "bob@olares.test"}, "log in as", time.Now()); err != nil {
		t.Fatalf("ensureProfileWritable for a hand-made profile: %v", err)
	}
}

// Removing a managed profile would delete an entry the next startup recreates,
// so it is refused rather than silently undone. Both the alias and the
// olaresId route to the same answer.
func TestRemoveRefusesAManagedProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("OLARES_CLI_HOME", home)
	cfg := managedConfig()
	cfg.Profiles[0].Name = "platform"
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	for _, key := range []string{"platform", managedID} {
		err := runRemove(key)
		var managed *credential.ErrManagedProfile
		if !errors.As(err, &managed) {
			t.Fatalf("runRemove(%q) = %v, want *credential.ErrManagedProfile", key, err)
		}
	}

	// The refusal happens before anything is written.
	raw, err := os.ReadFile(filepath.Join(home, "config.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(raw), managedID) {
		t.Errorf("config no longer holds the managed profile: %s", raw)
	}
}

// A managed profile with nothing in the keychain is `pending`, never `never`.
// The two differ in what they ask of the user: `never` says go and log in,
// which is the one thing this account cannot do.
func TestManagedProfileWithNoTokenIsPendingNotNever(t *testing.T) {
	store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{})

	managed := &cliconfig.ProfileConfig{OlaresID: managedID, Managed: true, AppName: "lares"}
	if got := profileStatus(store, managed, time.Now()); got != "pending" {
		t.Errorf("managed status = %q, want pending", got)
	}
	local := &cliconfig.ProfileConfig{OlaresID: "bob@olares.test"}
	if got := profileStatus(store, local, time.Now()); got != "never" {
		t.Errorf("local status = %q, want never", got)
	}
}

// An invalidated managed grant keeps the shared vocabulary: the word is the
// same as for any other account, only the remedy differs.
func TestInvalidatedManagedProfileKeepsTheSharedWord(t *testing.T) {
	store := auth.NewTokenStoreWith(keychainfake.New(), staticProfileLister{managedID})
	if err := store.Set(auth.StoredToken{OlaresID: managedID, Managed: true}); err != nil {
		t.Fatalf("seed token: %v", err)
	}
	// Set deliberately clears InvalidatedAt — a successful write is what
	// revives a grant — so the marker has to be stamped separately.
	if err := store.MarkInvalidated(managedID, time.Now()); err != nil {
		t.Fatalf("mark invalidated: %v", err)
	}

	p := &cliconfig.ProfileConfig{OlaresID: managedID, Managed: true, AppName: "lares"}
	if got := profileStatus(store, p, time.Now()); got != "invalidated" {
		t.Errorf("status = %q, want invalidated", got)
	}
}

// The SOURCE column carries the application, and only appears when a managed
// profile is present — a host install's listing is unchanged.
func TestSourceColumnAppearsOnlyForManagedProfiles(t *testing.T) {
	withManaged := renderListing(t, managedConfig())
	if !strings.Contains(withManaged, "SOURCE") {
		t.Errorf("listing has no SOURCE column:\n%s", withManaged)
	}
	if !strings.Contains(withManaged, "platform(lares)") {
		t.Errorf("listing does not name the application:\n%s", withManaged)
	}
	if !strings.Contains(withManaged, "local") {
		t.Errorf("listing does not mark the hand-made profile:\n%s", withManaged)
	}

	hostOnly := renderListing(t, &cliconfig.MultiProfileConfig{
		Profiles:       []cliconfig.ProfileConfig{{OlaresID: "bob@olares.test"}},
		CurrentProfile: "bob@olares.test",
	})
	if strings.Contains(hostOnly, "SOURCE") {
		t.Errorf("a host install grew a column that says nothing:\n%s", hostOnly)
	}
}

// SOURCE has to be the last column: the four before it are what existing
// output and any script reading it already expect.
func TestSourceIsTheLastColumn(t *testing.T) {
	listing := renderListing(t, managedConfig())
	header := strings.Fields(strings.SplitN(listing, "\n", 2)[0])
	want := []string{"NAME", "OLARES-ID", "STATUS", "VERSION", "SOURCE"}
	if strings.Join(header, " ") != strings.Join(want, " ") {
		t.Errorf("header = %v, want %v", header, want)
	}
}
