package profile

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// NewListCommand: `olares-cli profile list`
//
// Output is a TSV-like table: NAME / OLARES-ID / STATUS, with a leading "*"
// marking the current profile. STATUS reflects only what the local token
// store can prove without making a network call:
//
//	logged-in (23h59m)  — token present, JWT exp claim still in the future
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
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list all profiles with login status and current marker",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runList(os.Stdout)
		},
	}
}

func runList(out *os.File) error {
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
	fmt.Fprintln(w, "  \tNAME\tOLARES-ID\tSTATUS")
	for i := range cfg.Profiles {
		p := &cfg.Profiles[i]
		marker := " "
		if current != nil && current.OlaresID == p.OlaresID {
			marker = "*"
		}
		status := profileStatus(store, p, now)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, p.DisplayName(), p.OlaresID, status)
	}
	return w.Flush()
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
	return fmt.Sprintf("logged-in (%s)", humanizeDuration(exp.Sub(now)))
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
