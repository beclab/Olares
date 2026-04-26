// Package advanced implements the `olares-cli settings advanced` subtree —
// the docs call it "Advanced", the SPA's Vue page directory is "Developer/".
// Backed by user-service's terminusd.controller.ts (containerd, logs,
// hosts-file, upgrade) + bfl/env.controller.ts (system / user env). The
// hardware/restart-class verbs (reboot, shutdown, ssh-password) and the OS
// upgrade flow are owner-only and require JWS-signed bodies — they land in
// Phase 5 alongside the JWS-key sourcing decision (see plan.md "Phase 5").
package advanced

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAdvancedCommand returns the `settings advanced` parent.
func NewAdvancedCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced / Developer (containerd, env, logs, upgrade, hardware)",
		Long: `Advanced system settings:

  - containerd registries / mirrors / images
  - system + user env
  - log collection (terminusd /api/command/collectLogs)
  - OS upgrade lifecycle
  - hardware / restart-class actions (reboot, shutdown, ssh-password)

Subcommands will be added in subsequent phases:
  Phase 1: status, registries list, images list
  Phase 4: env (system|user) list / get / set / delete, collect-logs
  Phase 5: registries / mirrors / images CRUD + prune; upgrade state /
           start / cancel; reboot / shutdown / ssh-password
           (gated on JWS key sourcing — see plan.md "Open questions")
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
