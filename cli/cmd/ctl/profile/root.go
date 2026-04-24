// Package profile implements the `olares-cli profile` command tree.
//
// Phase 1 surface (5 subcommands, no separate `auth` namespace):
//
//	profile list                # list all profiles + login status, mark current
//	profile use <name|->        # switch current profile (`-` reverts to previous)
//	profile remove <name>       # delete profile + its stored token
//	profile login --olares-id <id> ...     # password-based login (mode A)
//	profile import --olares-id <id> ...    # refresh-token bootstrap (mode B)
//
// See docs/notes/olares-cli-auth-profile-config.md for the full design.
package profile

import "github.com/spf13/cobra"

// NewProfileCommand returns the `profile` parent command, ready to be added
// to the olares-cli root.
func NewProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "manage olares-cli profiles (one profile = one Olares instance + one user identity)",
		Long: `Manage olares-cli profiles. A profile bundles a target Olares instance
(identified by an olaresId such as "alice@olares.com") with the local
credentials used to talk to it.

Phase 1 stores tokens in plaintext at ~/.olares-cli/tokens.json with 0600
permissions; OS keychain support arrives in Phase 2.`,
	}
	for _, sub := range []*cobra.Command{
		NewListCommand(),
		NewUseCommand(),
		NewRemoveCommand(),
		NewLoginCommand(),
		NewImportCommand(),
	} {
		// Don't dump cobra usage on every runtime error — those are user
		// errors (bad creds, network, already-authenticated) whose message
		// is already actionable. SilenceUsage is per-command (no inheritance
		// from the parent), so we set it on every subcommand explicitly.
		sub.SilenceUsage = true
		cmd.AddCommand(sub)
	}
	return cmd
}
