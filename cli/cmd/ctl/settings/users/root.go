// Package users implements the `olares-cli settings users` subtree.
//
// Phase 0a is the empty umbrella — no verbs yet. Phase 1 will add `list` /
// `get`, Phase 2 the owner CRUD (create / delete / set-password / set-limits),
// and Phase 0b the `me` alias that delegates to the shared whoami helper
// from cmd/ctl/profile (matching plan.md's "two entry points, same
// implementation" rule).
//
// Backed by user-service's bfl/users.controller.ts which itself proxies BFL's
// /api/users surface; see plan.md's section "1. Users" for the authoritative
// API map. We split each area into its own Go package (rather than one flat
// `settings` package) so per-area types, parsers, and printers don't collide
// across the 13 sub-trees.
package users

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewUsersCommand returns the `settings users` parent. Subcommands land in
// later phases; for now the parent prints its own help (cobra default when
// no Run/RunE is set and no subcommands match), which is enough to confirm
// the umbrella wires through.
func NewUsersCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage Olares users (Settings -> Users)",
		Long: `Manage Olares users — list, inspect, create / delete, change passwords,
adjust per-user resource limits.

This corresponds to the Users section of the Olares Settings UI
(apps/packages/app/src/pages/settings/Account/), backed by
user-service's /api/users surface.

Subcommands will be added in subsequent phases:
  Phase 1: list, get
  Phase 2: create, delete, set-password, set-limits
  Phase 0b: me  (alias for "olares-cli profile whoami")
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewMeCommand(f))
	return cmd
}
