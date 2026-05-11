// Package advanced implements the `olares-cli settings advanced` subtree —
// the docs call it "Advanced", the SPA's Vue page directory is "Developer/".
// Backed by user-service's terminusd.controller.ts (containerd, logs,
// hosts-file, upgrade) + bfl/env.controller.ts (system / user env). The
// hardware/restart-class verbs (reboot, shutdown, ssh-password) and the OS
// upgrade flow are owner-only and require JWS-signed bodies; they are
// out of scope until a JWS-key sourcing path exists.
package advanced

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAdvancedCommand returns the `settings advanced` parent: containerd
// registries / images inspection plus system / user env management.
// JWS-gated writes (registries mutations, OS upgrade, reboot / shutdown,
// log collection) are not in scope until a JWS-key sourcing path lands.
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
  status
  registries list
  images list [--registry <name>]
  env (system|user) list / set --var KEY=VAL

Out of scope until a JWS key sourcing path exists:
  registries mirrors put/delete, registries prune,
  images delete / prune,
  upgrade state / start / cancel,
  reboot / shutdown / ssh-password,
  collect-logs (terminusd-signed)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewRegistriesCommand(f))
	cmd.AddCommand(NewImagesCommand(f))
	cmd.AddCommand(NewEnvCommand(f))
	return cmd
}
