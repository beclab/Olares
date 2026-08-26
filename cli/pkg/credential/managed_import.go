package credential

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// ImportManagedCredential reconciles config.json with the credential the
// platform mounted into this container. It runs once per invocation, before
// any command, and is a no-op on a host install.
//
// Importing on every start rather than once at install time is what keeps the
// two sides consistent without a reconciliation pass: the mount is recreated
// when the pod restarts, the emptyDir holding config.json is not, and neither
// side has a way to notify the other. Re-deriving the entry from the mount is
// cheaper than detecting that they disagree.
//
// Nothing here can fail a command. A container without a grant, with a
// half-written one, or with a config directory it cannot write to still runs
// every verb that does not need this identity.
func ImportManagedCredential(ctx context.Context) {
	newManagedImporter().run(ctx)
}

// refreshFunc matches Refresher.RefreshWith so tests can supply their own.
type refreshFunc func(ctx context.Context, olaresID, authURL, currentAccessToken, refreshToken string, insecureSkipVerify bool) (string, error)

type managedImporter struct {
	load    func() (*ManagedCredential, bool)
	store   auth.TokenStore
	refresh refreshFunc
	stderr  io.Writer
}

func newManagedImporter() *managedImporter {
	return &managedImporter{
		load:    LoadManagedCredential,
		store:   auth.NewTokenStore(),
		refresh: NewRefresher().RefreshWith,
		stderr:  os.Stderr,
	}
}

func (m *managedImporter) run(ctx context.Context) {
	cred, ok := m.load()
	if !ok {
		return
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		// A config file we cannot parse is the user's to fix; overwriting
		// it to make room for the managed entry would destroy profiles we
		// were never asked to touch.
		debugManaged("cannot read config: %v", err)
		return
	}

	existing := cfg.FindByOlaresID(cred.OlaresID)
	if existing != nil && existing.Managed {
		m.reconcileExisting(cfg, existing, cred)
		return
	}
	m.firstImport(ctx, cfg, existing, cred)
}

// reconcileExisting is the steady state, reached by every command after the
// first one in a container's life. It touches no network: the entry is already
// there and the token, if it needs rotating, is rotated by whoever makes the
// call that gets a 401.
func (m *managedImporter) reconcileExisting(cfg *cliconfig.MultiProfileConfig, existing *cliconfig.ProfileConfig, cred *ManagedCredential) {
	if existing.AppName == cred.AppName {
		return
	}
	existing.AppName = cred.AppName
	m.save(cfg)
}

// firstImport creates the managed entry — or converts a manually created one
// — and immediately exchanges the mounted refresh token for an access token.
//
// The exchange happens here rather than being left to the first command that
// needs a token because otherwise `profile list`, which is what a person runs
// first, would show an account with no token and no way to tell "broken" from
// "not started yet".
func (m *managedImporter) firstImport(ctx context.Context, cfg *cliconfig.MultiProfileConfig, existing *cliconfig.ProfileConfig, cred *ManagedCredential) {
	target := cliconfig.ProfileConfig{OlaresID: cred.OlaresID}
	if existing != nil {
		// Taking over a profile the user made by hand keeps how to reach
		// the instance (alias, URL overrides, TLS) and replaces only who
		// vouches for the identity. A dev profile whose auth URL override
		// were dropped here would simply stop connecting.
		target = *existing
	}
	target.Managed = true
	target.AppName = cred.AppName

	authURL, err := target.ResolvedAuthURL()
	if err != nil {
		debugManaged("cannot derive auth URL for %s: %v", cred.OlaresID, err)
		return
	}

	currentAccessToken := ""
	if stored, err := m.store.Get(cred.OlaresID); err == nil {
		// Passing what is stored is what stops the refresher's
		// compare-after-get from mistaking a pre-existing token for a
		// concurrent refresh and skipping the exchange entirely.
		currentAccessToken = stored.AccessToken
	}

	_, refreshErr := m.refresh(ctx, cred.OlaresID, authURL, currentAccessToken, cred.RefreshToken, target.InsecureSkipVerify)

	var notSecondFactor *ErrManagedNotSecondFactor
	if errors.As(refreshErr, &notSecondFactor) {
		// The platform promises a second-factor grant. Getting a
		// first-factor one back means the grant is unusable against every
		// per-service host, so we create nothing: an entry that 401s on
		// every command is worse than no entry at all.
		fmt.Fprintf(m.stderr, "warning: %v\n", refreshErr)
		return
	}
	if refreshErr != nil {
		// Everything else is either a dead grant (already marked by the
		// refresher) or transient. The entry is still created so `profile
		// list` can say which.
		debugManaged("initial refresh for %s failed: %v", cred.OlaresID, refreshErr)
	}

	m.sealUserToken(cred.OlaresID)
	cfg.Upsert(target)
	m.save(cfg)
}

// sealUserToken removes a keychain record that a manual login left behind for
// an olaresId that is now managed.
//
// A managed entry holds no refresh token — the mount is the platform's only
// copy and the only one it can revoke — so the user's would sit there
// unreadable by any code path and unremovable by `profile remove`, which
// refuses to touch a managed profile. The invalidation marker is the one thing
// worth carrying across, because it is what `profile list` reports.
func (m *managedImporter) sealUserToken(olaresID string) {
	stored, err := m.store.Get(olaresID)
	if err != nil || stored.Managed {
		return
	}
	if err := m.store.Delete(olaresID); err != nil {
		debugManaged("cannot remove the superseded token for %s: %v", olaresID, err)
		return
	}
	if stored.InvalidatedAt > 0 {
		if err := m.store.Set(auth.StoredToken{
			OlaresID:      olaresID,
			Managed:       true,
			InvalidatedAt: stored.InvalidatedAt,
		}); err != nil {
			debugManaged("cannot record the invalidated managed grant for %s: %v", olaresID, err)
		}
	}
}

// save degrades a write failure to a warning. config.json lands under a
// directory the container may not be able to create — HOME is sometimes / and
// sometimes read-only — and refusing to run every other verb over that would
// be a far bigger failure than the managed profile being absent.
func (m *managedImporter) save(cfg *cliconfig.MultiProfileConfig) {
	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		fmt.Fprintf(m.stderr, "warning: cannot persist the platform-issued profile: %v\n", err)
	}
}
