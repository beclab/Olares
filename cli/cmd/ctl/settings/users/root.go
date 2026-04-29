// Package users implements the `olares-cli settings users` subtree.
//
// Backed by user-service's bfl/users.controller.ts which itself proxies BFL's
// /api/users surface; see plan.md's section "1. Users" for the authoritative
// API map. Each settings area lives in its own Go package (rather than one
// flat `settings` package) so per-area types, parsers, and printers don't
// collide across the 13 sub-trees.
//
// The umbrella exposes a `me` alias that delegates to the shared whoami
// helper in cmd/ctl/profile, matching plan.md's "two entry points, same
// implementation" rule.
package users

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewUsersCommand returns the `settings users` parent: list and inspect
// Olares users, plus the `me` whoami shortcut. Owner-only CRUD (create /
// delete / set-password / set-limits) is out of scope for now.
func NewUsersCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage Olares users (Settings -> Users)",
		Long: `Manage Olares users — list and inspect.

This corresponds to the Users section of the Olares Settings UI
(apps/packages/app/src/pages/settings/Account/), backed by
user-service's /api/users surface.

Subcommands:
  list                    list users
  get <name>              inspect a single user
  me                      alias for "olares-cli profile whoami"

Out of scope for now (owner-only CRUD):
  create, delete, set-password, set-limits
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewMeCommand(f))
	cmd.AddCommand(NewListCommand(f))
	cmd.AddCommand(NewGetCommand(f))
	return cmd
}
