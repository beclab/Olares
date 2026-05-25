// Package advanced implements the `olares-cli settings advanced` subtree —
// the docs call it "Advanced", the SPA's Vue page directory is "Developer/".
// Backed by user-service's terminusd.controller.ts (containerd, hosts-file,
// upgrade) + bfl/env.controller.ts (system / user env). The
// hardware/restart-class verbs (reboot, shutdown, ssh-password) and the OS
// upgrade flow are owner-only and require JWS-signed bodies; they are
// out of scope until a JWS-key sourcing path exists.
package advanced

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewAdvancedCommand returns the `settings advanced` parent: containerd
// registries / images inspection and system / user env management.
// JWS-gated writes (registries mutations, OS upgrade, reboot / shutdown)
// stay out of scope until a JWS-key sourcing path exists.
func NewAdvancedCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced / Developer (containerd, env, upgrade, hardware)",
		Long: `Advanced system settings:

  - containerd registries / mirrors / images
  - system + user env
  - OS upgrade lifecycle
  - hardware / restart-class actions (reboot, shutdown, ssh-password)

For CLI log tarball collection, use top-level olares-cli logs (not under settings).

Subcommands:
  status
  registries list
  images list [--registry <name>]
  env (system|user) list

Out of scope until a JWS key sourcing path exists:
  registries mirrors put/delete, registries prune,
  images delete / prune,
  upgrade state / start / cancel,
  reboot / shutdown / ssh-password,
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewRegistriesCommand(f))
	cmd.AddCommand(NewImagesCommand(f))
	cmd.AddCommand(NewEnvCommand(f))
	return cmd
}
