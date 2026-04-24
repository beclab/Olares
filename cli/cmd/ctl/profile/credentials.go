package profile

import (
	"errors"
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/olares"
)

// commonCredFlags captures the flags shared by `profile login` and
// `profile import`. The two commands diverge only in HOW they obtain the
// initial Token; everything else (CLI surface, profile creation rules,
// persistence) is identical.
type commonCredFlags struct {
	olaresID           string
	name               string
	authURLOverride    string
	localURLPrefix     string
	insecureSkipVerify bool
}

// validateAndDeriveAuthURL canonicalizes the user-supplied flags into a
// concrete (terminusName, authURL) pair, applying AuthURLOverride when
// present and otherwise deriving from the parsed olaresId.
func (f *commonCredFlags) validateAndDeriveAuthURL() (id olares.ID, terminusName, authURL string, err error) {
	if f.olaresID == "" {
		return "", "", "", errors.New("--olares-id is required")
	}
	id, err = olares.ParseID(f.olaresID)
	if err != nil {
		return "", "", "", err
	}
	terminusName = id.TerminusName()
	if f.authURLOverride != "" {
		authURL = f.authURLOverride
	} else {
		authURL = id.AuthURL(f.localURLPrefix)
	}
	return id, terminusName, authURL, nil
}

// ensureProfileWritable enforces the "auto-create-or-reuse, reject if valid
// token exists" rule shared by login and import.
//
// Returns the (possibly newly-allocated) ProfileConfig that the caller should
// upsert AFTER it has successfully obtained a fresh Token. If a profile
// already exists for olaresID, its URL-override fields are preserved and only
// the alias (Name) is updated when the caller passed a non-empty --name. If a
// VALID token is already present in the store for this olaresId, the function
// returns an error instructing the user to `profile remove` first — this
// matches the design doc's "refuse duplicate logins" rule.
func ensureProfileWritable(
	cfg *cliconfig.MultiProfileConfig,
	store auth.TokenStore,
	flags commonCredFlags,
	now time.Time,
) (cliconfig.ProfileConfig, error) {
	// 1. Reject if a still-valid token exists for this olaresId. An
	// explicitly invalidated token (Phase 2 marks this on /api/refresh
	// failure) is always treated as "needs re-login" and falls through —
	// users should be able to recover with a single `profile login`, no
	// `profile remove` required.
	stored, err := store.Get(flags.olaresID)
	if err != nil && !errors.Is(err, auth.ErrTokenNotFound) {
		return cliconfig.ProfileConfig{}, fmt.Errorf("read token store: %w", err)
	}
	if err == nil && stored.InvalidatedAt == 0 {
		exp, expErr := auth.ExpiresAt(stored.AccessToken)
		// "Valid" = exp is parseable AND in the future. A token with no exp
		// claim is treated as "unknown / could still be valid" → also reject,
		// because we can't prove otherwise client-side.
		if expErr == nil && now.Before(exp) {
			return cliconfig.ProfileConfig{}, fmt.Errorf(
				"already authenticated for %s (expires in %s).\nto re-authenticate, run: olares-cli profile remove %s",
				flags.olaresID, humanizeDuration(exp.Sub(now)), flags.olaresID,
			)
		}
		if errors.Is(expErr, auth.ErrNoExpClaim) {
			return cliconfig.ProfileConfig{}, fmt.Errorf(
				"a token is already stored for %s but its expiry can't be determined client-side.\nto re-authenticate, run: olares-cli profile remove %s",
				flags.olaresID, flags.olaresID,
			)
		}
		// Otherwise the token is expired or unparseable → fall through and
		// overwrite it.
	}

	// 2. Build the ProfileConfig we're about to upsert. If one already
	// exists, preserve its overrides unless the caller explicitly passed a
	// new value.
	if existing := cfg.FindByOlaresID(flags.olaresID); existing != nil {
		out := *existing
		if flags.name != "" {
			out.Name = flags.name
		}
		if flags.authURLOverride != "" {
			out.AuthURLOverride = flags.authURLOverride
		}
		if flags.localURLPrefix != "" {
			out.LocalURLPrefix = flags.localURLPrefix
		}
		if flags.insecureSkipVerify {
			out.InsecureSkipVerify = true
		}
		return out, nil
	}
	return cliconfig.ProfileConfig{
		Name:               flags.name,
		OlaresID:           flags.olaresID,
		AuthURLOverride:    flags.authURLOverride,
		LocalURLPrefix:     flags.localURLPrefix,
		InsecureSkipVerify: flags.insecureSkipVerify,
	}, nil
}

// persistResult reports what happened to the active-profile pointer as a side
// effect of persistTokenAndProfile, so callers can print accurate UX.
//
// Switched is true exactly when CurrentProfile changed during this call. In
// that case PreviousCurrent holds whatever CurrentProfile pointed at before
// the switch (may be empty when the just-persisted profile is the very first
// one).
type persistResult struct {
	Switched        bool
	PreviousCurrent string
}

// persistTokenAndProfile writes the freshly-obtained Token into the token
// store and upserts the corresponding ProfileConfig into config.json.
//
// switchCurrent controls whether the just-persisted profile becomes current:
//   - true  → behave like `profile use <id>`: if the new profile differs from
//     the existing CurrentProfile, the old CurrentProfile is moved into
//     PreviousProfile so users can revert with `profile use -`. Re-persisting
//     the already-current profile is a no-op for current/previous.
//   - false → leave CurrentProfile alone, except when it's empty: in that
//     case fall back to the just-persisted profile so we never end up with
//     "profiles exist but no current" — which would break every command that
//     resolves credentials via the current profile.
func persistTokenAndProfile(
	cfg *cliconfig.MultiProfileConfig,
	store auth.TokenStore,
	profile cliconfig.ProfileConfig,
	tok *auth.Token,
	switchCurrent bool,
) (persistResult, error) {
	stored := auth.StoredToken{
		OlaresID:     profile.OlaresID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		SessionID:    tok.SessionID,
		GrantedAt:    time.Now().UnixMilli(),
	}
	if err := store.Set(stored); err != nil {
		return persistResult{}, fmt.Errorf("save token: %w", err)
	}
	persisted := cfg.Upsert(profile)

	res := persistResult{}
	newName := persisted.DisplayName()
	prevCurrent := cfg.CurrentProfile
	switch {
	case switchCurrent && prevCurrent != newName:
		// SetCurrent handles the empty-current case (no PreviousProfile
		// update), and updates PreviousProfile when current actually moves.
		// The lookup can only fail if the upsert above didn't land — treat
		// that as an internal invariant violation.
		if _, err := cfg.SetCurrent(newName); err != nil {
			return persistResult{}, fmt.Errorf("activate profile %q: %w", newName, err)
		}
		res.Switched = true
		res.PreviousCurrent = prevCurrent
	case !switchCurrent && prevCurrent == "":
		// Bootstrap path when --no-switch was passed but there's literally no
		// current to preserve. Still no PreviousProfile bookkeeping (there
		// was nothing to demote).
		cfg.CurrentProfile = newName
		res.Switched = true
	}

	if err := cliconfig.SaveMultiProfileConfig(cfg); err != nil {
		return persistResult{}, fmt.Errorf("save config: %w", err)
	}
	return res, nil
}

// printSwitchNotice renders the post-login UX line(s) that explain whether
// CurrentProfile moved as a result of the just-finished login/import.
//
// We deliberately stay quiet when nothing changed (re-login on the
// already-current profile, or --no-switch with a non-empty current) so the
// happy-path output keeps a single line of "logged in as ...".
func printSwitchNotice(res persistResult, newDisplayName string) {
	if !res.Switched {
		return
	}
	fmt.Printf("switched current profile to %s\n", newDisplayName)
	if res.PreviousCurrent != "" {
		fmt.Printf("previous profile: %s (use 'olares-cli profile use -' to switch back)\n", res.PreviousCurrent)
	}
}

// printPlaintextWarning is shown after every successful login / import to set
// expectations: Phase 1 stores tokens in clear text. Phase 2 will move them
// into the OS keychain.
func printPlaintextWarning() {
	tokensPath, _ := cliconfig.TokensFile()
	fmt.Printf("warning: token stored in plaintext at %s (mode 0600). OS keychain support is coming in a future release.\n", tokensPath)
}
