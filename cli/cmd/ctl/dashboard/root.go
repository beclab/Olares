// Package dashboard implements the `olares-cli dashboard` subtree — the
// AI-agent-oriented mirror of the dashboard SPA's overview / applications
// pages. Every leaf command makes the same HTTP calls the SPA does
// (against the per-user `dashboard.<terminus>` BFF, ks-apiserver-proxy)
// and renders either a tabular human view (`-o table`, default) or a
// strict JSON envelope for agent / scripting use (`-o json`).
//
// Architecture (see SKILL.md for the full red-line list):
//
//   - cli/cmd/ctl/dashboard/                       — Cobra wiring (this dir).
//     A thin shell: builds the command tree, binds pflags onto the
//     pkg-side CommonFlags, and forwards a *pkgdashboard.Client into each
//     leaf RunE. Per-area subdirectories mirror the command tree exactly:
//
//	     dashboard/
//	       ├── overview/                            (NewOverviewCommand)
//	       │   ├── disk/                            (NewDiskCommand)
//	       │   ├── fan/                             (NewFanCommand)
//	       │   └── gpu/                             (NewGPUCommand)
//	       ├── applications/                        (NewApplicationsCommand)
//	       └── schema/                              (NewSchemaCommand)
//
//   - cli/pkg/dashboard/                           — Heavy core. Owns
//     Envelope/Item/Meta/CommonFlags/Client/Runner, all fetchers, all
//     aggregators, and the schema bundle. Cmd subpackages call into pkg
//     directly; horizontal imports between cmd subpackages are forbidden
//     (the one legitimate cross-area share, BuildRankingEnvelope, lives
//     in pkg/dashboard/ranking.go).
package dashboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/cmdutil"

	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/applications"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/overview"
	"github.com/beclab/Olares/cli/cmd/ctl/dashboard/schema"
)

// common is the shared CommonFlags singleton every leaf command consumes.
// The per-area subpackages each hold a *pkgdashboard.CommonFlags pointer
// wired to this struct via NewXxxCommand factories, so cobra's persistent
// flag inheritance mutates one canonical struct.
var common pkgdashboard.CommonFlags

// NewDashboardCommand assembles the `olares-cli dashboard` subtree. f is
// the shared cmdutil.Factory the root command builds — every leaf reaches
// into it for an authenticated *http.Client and the resolved profile.
func NewDashboardCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Query the per-user Olares dashboard for AI agents",
		Long: `Query the same data the dashboard SPA renders, but from the CLI.

Every leaf command emits one of two output shapes:
  -o table  (default) — a human-readable, two-column-gutter table.
  -o json             — a strict envelope:
                        { kind, meta, items: [...] }
                        or, for "dashboard overview" (default action),
                        { kind, meta, sections: { ... } }

Authentication, transport and per-user routing are inherited from the
currently-selected profile (switch with "olares-cli profile use <name>").
Token refresh on 401/403 is transparent.

For agent integration, run "olares-cli dashboard schema" to discover the
available commands + their JSON Schemas.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          unknownSubcommandRunE,
	}

	bindPersistentFlags(&common, cmd)

	cmd.AddCommand(overview.NewOverviewCommand(f, &common))
	cmd.AddCommand(applications.NewApplicationsCommand(f, &common))
	cmd.AddCommand(schema.NewSchemaCommand(&common))

	wrapLeafErrors(cmd)
	return cmd
}

// wrapLeafErrors decorates every RunE in the dashboard subtree so a
// non-nil error gets printed to stderr before being returned to cobra.
// cobra would otherwise drop the message because every dashboard cmd
// sets SilenceErrors=true (a contract pinned by TestAllLeafCommandsSilenced
// — runtime errors must not splat usage / help text on stderr; only the
// structured envelope or the typed error reaches the agent).
//
// Errors marked with pkgdashboard.ErrAlreadyReported are passed through
// unchanged: the RunE has already written a user-visible diagnostic
// (e.g. unknownSubcommandRunE's typo hint, or Runner.emitFailure's
// per-iteration warning) and printing again would be redundant.
//
// This wiring is local to the dashboard subtree — other ctl subtrees
// (files / profile / os / …) still rely on cobra's default auto-print
// because they don't set SilenceErrors=true.
func wrapLeafErrors(c *cobra.Command) {
	if orig := c.RunE; orig != nil {
		c.RunE = func(cmd *cobra.Command, args []string) error {
			err := orig(cmd, args)
			if err != nil && !errors.Is(err, pkgdashboard.ErrAlreadyReported) {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
			}
			return err
		}
	}
	for _, sub := range c.Commands() {
		wrapLeafErrors(sub)
	}
}

// unknownSubcommandRunE is the RunE wired onto the dashboard root. With
// no args it falls through to cobra's help; with positional args it
// writes a "Did you mean…" hint to stderr (since SilenceErrors=true
// otherwise swallows cobra's own suggestion) and returns
// pkgdashboard.ErrAlreadyReported so the process exits non-zero AND the
// leaf-error wrapper (wrapLeafErrors) doesn't double-print.
func unknownSubcommandRunE(c *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.Help()
	}
	msg := fmt.Sprintf("Error: unknown subcommand %q for %q", args[0], c.CommandPath())
	if suggestions := c.SuggestionsFor(args[0]); len(suggestions) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(suggestions, "\n\t")
	}
	fmt.Fprintln(c.ErrOrStderr(), msg)
	fmt.Fprintf(c.ErrOrStderr(), "\nRun '%s --help' for usage.\n", c.CommandPath())
	return pkgdashboard.ErrAlreadyReported
}
