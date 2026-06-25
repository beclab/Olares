package profile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// NewListCommand: `olares-cli profile list`
//
// Output is a TSV-like table: NAME / OLARES-ID / STATUS, with a leading "*"
// marking the current profile. STATUS reflects only what the local token
// store can prove without making a network call:
//
//	logged-in           — token present, JWT exp claim still in the future
//	expired             — token present, exp claim in the past
//	invalidated         — token present but explicitly marked unusable
//	                      (Phase 2 sets this when /api/refresh returns 401/403);
//	                      takes precedence over `expired`
//	never               — no stored token for this profile
//	logged-in           — token present but JWT has no exp claim (we can't
//	                      tell client-side; trust until the server says no)
//
// Per §7.5 of the design doc, we deliberately do NOT print any other JWT
// claims (username / groups / mfa / jid). The OlaresID column is the local
// authoritative identity.
func NewListCommand(f *cmdutil.Factory) *cobra.Command {
	var refresh bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all profiles with login status, current marker, cached location and backend version",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runList(c.Context(), f, refresh, os.Stdout)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "re-detect the CURRENT profile (location, role, backend version) before listing")
	return cmd
}

func runList(ctx context.Context, f *cmdutil.Factory, refresh bool, out *os.File) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// With --refresh, run the unified detect for the ACTIVE profile (re-probe
	// location + refetch role + version) and persist it before we render. This
	// only touches the current profile (the only one we have a token for); the
	// rest still show their last-cached values. Best-effort: any failure
	// degrades to a stderr warning so the listing itself never breaks.
	if refresh && f != nil {
		if rp, rerr := f.ResolveProfile(ctx); rerr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not resolve the current profile to refresh: %v\n", rerr)
		} else if cfg0, cerr := cliconfig.LoadMultiProfileConfig(); cerr == nil {
			if _, derr := whoami.DetectAndCache(ctx, whoami.DetectInput{
				Cfg:             cfg0,
				OlaresID:        rp.OlaresID,
				LocalPrefix:     rp.LocalURLPrefix,
				Insecure:        rp.InsecureSkipVerify,
				AccessToken:     rp.AccessToken,
				AuthURLOverride: rp.AuthURLOverride,
				Now:             time.Now,
			}); derr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not refresh the current profile: %v\n", derr)
			}
		}
	}

	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		fmt.Fprintln(out, "no profiles configured.")
		fmt.Fprintln(out, "run `olares-cli profile login --olares-id <id>` or `olares-cli profile import --olares-id <id> --refresh-token <tok>` to add one.")
		return nil
	}

	store := auth.NewTokenStore()

	current := cfg.Current()
	now := time.Now()

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  \tNAME\tOLARES-ID\tSTATUS\tLOCATION\tVERSION")
	for i := range cfg.Profiles {
		p := &cfg.Profiles[i]
		marker := " "
		if current != nil && current.OlaresID == p.OlaresID {
			marker = "*"
		}
		status := profileStatus(store, p, now)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", marker, p.DisplayName(), p.OlaresID, status, locationCell(p), backendVersionCell(p))
	}
	return w.Flush()
}

// locationCell renders the LOCATION column: the cached network position
// ("external" / "lan" / "host" / "cluster"), or "-" when it hasn't been
// probed yet (pre-existing profile, or a login where probing failed).
func locationCell(p *cliconfig.ProfileConfig) string {
	if p.Location != "" {
		return p.Location
	}
	return "-"
}

// backendVersionCell renders the VERSION column for a profile: the cached
// Olares backend version, or "-" when it hasn't been detected yet (no login
// eager-fetch landed and no version-aware command has run / refreshed it).
func backendVersionCell(p *cliconfig.ProfileConfig) string {
	if v := p.BackendVersion; v != "" {
		return v
	}
	return "-"
}

// profileStatus inspects the token store for `p` and returns a short status
// string. Errors reading the store collapse into an opaque "unknown" rather
// than aborting the whole listing — partial output beats no output here.
func profileStatus(store auth.TokenStore, p *cliconfig.ProfileConfig, now time.Time) string {
	tok, err := store.Get(p.OlaresID)
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return "never"
		}
		return "unknown"
	}
	// Explicit invalidation wins over JWT-exp inspection: a server-side
	// rejection of the refresh leg means the entire grant is dead, even if
	// the access_token JWT happens to still have time left on its `exp`.
	if tok.InvalidatedAt > 0 {
		return "invalidated"
	}
	exp, err := auth.ExpiresAt(tok.AccessToken)
	if err != nil {
		if errors.Is(err, auth.ErrNoExpClaim) {
			return "logged-in"
		}
		return "logged-in (unparseable token)"
	}
	if !now.Before(exp) {
		return "expired"
	}
	return "logged-in"
}

// humanizeDuration prints a coarse "23h59m" / "12m34s" / "5s" representation.
// Days are folded into hours to keep the column width predictable.
func humanizeDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) - m*60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) - h*60
	return fmt.Sprintf("%dh%dm", h, m)
}
