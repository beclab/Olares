package credential

import (
	"context"
	"errors"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// CredentialProvider chains zero or more Providers in priority order: the
// first one that returns a non-nil ResolvedProfile wins. Phase 1 wires
// (EnvProvider, DefaultProvider) — env first so the in-cluster scenario can
// pre-empt the on-disk config when shipped.
//
// This is the Phase-1 analogue of lark-cli's credential.CredentialProvider,
// minus the multi-app / token-cache plumbing (Phase 2).
type CredentialProvider struct {
	providers []Provider
}

// NewCredentialProvider returns a chain that consults each Provider in order.
// Pass them most-specific first.
func NewCredentialProvider(providers ...Provider) *CredentialProvider {
	return &CredentialProvider{providers: providers}
}

// ErrNoProfile is returned when no Provider could resolve a profile (typically
// because the user hasn't run `profile login` yet AND no in-cluster env vars
// are present).
var ErrNoProfile = errors.New("no Olares profile is configured: run `olares-cli profile login --olares-id <id>` or `olares-cli profile import --olares-id <id> --refresh-token <tok>`")

// Resolve walks the provider chain. It is responsible for loading the on-disk
// profile (if any) once and feeding it to each provider. The first
// non-nil ResolvedProfile from the chain is returned. If every provider
// declines, ErrNoProfile is returned (or the most informative error from a
// declining provider, if all returned errors).
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

	var lastErr error
	for _, p := range c.providers {
		resolved, err := p.Resolve(ctx, profile)
		if err != nil {
			lastErr = fmt.Errorf("provider %s: %w", p.Name(), err)
			continue
		}
		if resolved != nil {
			if resolved.Source == "" {
				resolved.Source = p.Name()
			}
			return resolved, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, ErrNoProfile
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

// RequireBuiltinCredentialProvider is a hook for Phase 3: when an env-driven
// (or other "external") provider is in play, mutating commands like
// `profile login` should refuse to run because there's nothing local to
// mutate. Phase 1 always returns nil; the call sites are wired now so future
// activation is mechanical.
func RequireBuiltinCredentialProvider(_ *CredentialProvider) error {
	return nil
}
