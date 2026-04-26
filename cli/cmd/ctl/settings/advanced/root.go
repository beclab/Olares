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

// NewAdvancedCommand returns the `settings advanced` parent. Phase 1
// ships the read-only verbs that don't require JWS signing; Phase 4 +
// Phase 5 add the env/logs/upgrade/restart verbs.
func NewAdvancedCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced / Developer (containerd, env, logs, upgrade, hardware)",
		Long: `Advanced system settings:

  - containerd registries / mirrors / images
  - system + user env
  - log collection (terminusd /api/command/collectLogs)
  - OS upgrade lifecycle
  - hardware / restart-class actions (reboot, shutdown, ssh-password)

Subcommands:
  status                                                  (Phase 1)
  registries list                                         (Phase 1)
  images list [--registry <name>]                         (Phase 1)

Subcommands landing in Phase 4:
  env (system|user) list / get / set / delete, collect-logs

Subcommands landing in Phase 5 (JWS-signed):
  registries mirrors get/set/delete, registries prune,
  images delete / prune,
  upgrade state / start / cancel,
  reboot / shutdown / ssh-password
  (gated on JWS key sourcing — see plan.md "Open questions")
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewRegistriesCommand(f))
	cmd.AddCommand(NewImagesCommand(f))
	return cmd
}
