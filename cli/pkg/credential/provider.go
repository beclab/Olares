package credential

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// CredentialProvider selects the Provider that owns the active profile.
//
// The selection is by profile, not by trial. An earlier version asked each
// provider in turn and took the first non-nil answer, which reads the wrong
// question: the managed provider's claim would rest on a credential file
// being present in the container rather than on the selected profile being
// the platform's. Somebody who ran `profile use` to switch to an account they
// logged into by hand would still have been served the platform's identity.
//
// This is the olares-cli analogue of lark-cli's credential.CredentialProvider,
// minus the multi-app / token-cache plumbing.
type CredentialProvider struct {
	managed Provider
	local   Provider
}

// NewCredentialProvider pairs the provider for platform-issued profiles with
// the one for profiles the user logged into on this machine.
func NewCredentialProvider(managed, local Provider) *CredentialProvider {
	return &CredentialProvider{managed: managed, local: local}
}

// ErrNoProfile is returned when no Provider could resolve a profile (typically
// because the user hasn't run `profile login` yet AND no in-cluster env vars
// are present).
var ErrNoProfile = errors.New("no Olares profile is configured: run `olares-cli profile login --olares-id <id>` or `olares-cli profile import --olares-id <id> --refresh-token <tok>`")

// Resolve loads the on-disk profile and hands it to whichever provider owns
// that kind of profile. ErrNoProfile is returned when there is no profile to
// resolve, or when the provider declines one.
//
// `profileKey` is an optional override (e.g. the `--olares-id` flag on
// `profile login` / `profile import`). When empty, the currently-selected
// profile from config.json is used. There is intentionally no global
// per-invocation flag that fills this in for normal verbs; identity is
// switched explicitly via `olares-cli profile use <name>`.
func (c *CredentialProvider) Resolve(ctx context.Context, profileKey string) (*ResolvedProfile, error) {
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	var profile *cliconfig.ProfileConfig
	if profileKey != "" {
		profile = cfg.FindProfile(profileKey)
		if profile == nil {
			return nil, fmt.Errorf("profile %q not found in %s", profileKey, configFileForError())
		}
	} else {
		profile = cfg.Current()
	}
	if profile == nil {
		return nil, ErrNoProfile
	}

	p := c.local
	if profile.Managed {
		p = c.managed
	}
	resolved, err := p.Resolve(ctx, profile)
	if err != nil {
		// The typed errors (ErrNotLoggedIn, ErrTokenInvalidated) carry the
		// CTA a caller renders, so they are returned as they are rather
		// than wrapped in the provider's name.
		return nil, err
	}
	if resolved == nil {
		return nil, ErrNoProfile
	}
	if resolved.Source == "" {
		resolved.Source = p.Name()
	}
	return resolved, nil
}

// configFileForError best-effort-resolves the config path for inclusion in
// "not found" error messages. Returns "<unknown>" if resolution itself fails.
func configFileForError() string {
	p, err := cliconfig.ConfigFile()
	if err != nil {
		return "<unknown>"
	}
	return p
}

// ErrManagedProfile is returned by RequireNotManaged. It names the
// application so a user who never created this account can tell where it came
// from, and states what actually revokes it, which is not anything the CLI
// can do: the platform holds the only copy of the refresh token.
type ErrManagedProfile struct {
	OlaresID string
	AppName  string
	Verb     string
}

func (e *ErrManagedProfile) Error() string {
	app := e.AppName
	if app == "" {
		app = "an installed application"
	} else {
		app = fmt.Sprintf("application %q", app)
	}
	return fmt.Sprintf("cannot %s %s: its credential is issued by the platform for %s and is not stored locally; uninstall that application to revoke it",
		e.Verb, e.OlaresID, app)
}

// RequireNotManaged refuses a command that would overwrite or delete a
// platform-issued profile. Nothing local backs such a profile — the refresh
// token lives in a mount the platform owns — so a login would write a second
// grant the next startup discards, and a remove would delete an entry the
// next startup recreates. Both read as the command having failed silently.
//
// A nil profile, or one the user logged into by hand, passes: logging into a
// different identity inside the same container is nobody's business but the
// user's.
func RequireNotManaged(profile *cliconfig.ProfileConfig, verb string) error {
	if profile == nil || !profile.Managed {
		return nil
	}
	return &ErrManagedProfile{
		OlaresID: profile.OlaresID,
		AppName:  profile.AppName,
		Verb:     verb,
	}
}
