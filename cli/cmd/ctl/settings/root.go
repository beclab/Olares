// Package settings implements the `olares-cli settings` umbrella command
// tree. Identity (which Olares user) and transport (which cluster) are
// resolved from the global --profile flag via cmdutil.Factory, exactly like
// `olares-cli files` and `olares-cli market`.
//
// The subtree mirrors the 12 canonical sections of the Olares Settings UI
// documented at https://docs.olares.com/manual/olares/settings/, plus a
// 13th, non-canonical "me" sub-tree that hosts the SPA's avatar/Person
// dropdown self-service items (whoami / version / login-history / SSO /
// password). The 13th sub-tree is intentionally *outside* the 12 docs
// sections and is documented as such in its own package — it ships under
// `settings` for discoverability, not because it's a docs section.
//
// Phase 0a (this file) ships the umbrella with all area parents in place
// but no real verbs yet. Phase 0b adds role caching + soft preflight; Phases
// 1-6 progressively port the actual verbs.
package settings

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/advanced"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/appearance"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/apps"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/backup"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/gpu"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/integration"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/me"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/network"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/restore"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/search"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/users"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/video"
	"github.com/beclab/Olares/cli/cmd/ctl/settings/vpn"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSettingsCommand assembles the `olares-cli settings` subtree. Every
// area is wired here so the umbrella's --help is the directory of available
// areas from day one — even when individual verbs are still pending.
//
// Authentication and transport are inherited from the shared cmdutil.Factory
// (set up in cli/cmd/ctl/root.go) so the persistent --profile flag flows
// through unchanged. No per-command auth flags.
func NewSettingsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage Olares Settings via the per-user profile (mirror of the SPA Settings UI)",
		Long: `Manage Olares Settings — the same surface the desktop SPA exposes under
"Settings" (https://docs.olares.com/manual/olares/settings/).

This umbrella mirrors the 12 documented sections:

  users         appearance   apps          integration   vpn         network
  gpu           video        search        backup        restore     advanced

Plus a 13th, non-canonical "me" sub-tree for the SPA's avatar/Person
dropdown self-service items (whoami / version / login-history / sso /
password) — folded in here for CLI discoverability.

Identity and transport come from the active profile (--profile or the
currently-selected one), so authentication uses the same access token as
"olares-cli profile login" and the same edge auth chain the Olares web app
uses (Authelia + l4-bfl-proxy). Most APIs ride https://desktop.<terminus>;
backup / restore use the same origin with a /apis/backup/v1/* prefix
served by BFL's backup-server.
`,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.PersistentPreRun = func(c *cobra.Command, args []string) {
		c.SilenceErrors = true
		c.SilenceUsage = true
	}

	for _, sub := range []*cobra.Command{
		users.NewUsersCommand(f),
		appearance.NewAppearanceCommand(f),
		apps.NewAppsCommand(f),
		integration.NewIntegrationCommand(f),
		vpn.NewVPNCommand(f),
		network.NewNetworkCommand(f),
		gpu.NewGPUCommand(f),
		video.NewVideoCommand(f),
		search.NewSearchCommand(f),
		backup.NewBackupCommand(f),
		restore.NewRestoreCommand(f),
		advanced.NewAdvancedCommand(f),
		me.NewMeCommand(f),
	} {
		cmd.AddCommand(sub)
	}
	return cmd
}
