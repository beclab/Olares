// Package vpn implements the `olares-cli settings vpn` subtree (Settings ->
// VPN). Backed by user-service's headscale/headscale.controller.ts +
// bfl/acl.controller.ts + the public-domain-policy slice of
// bfl/network.controller.ts.
package vpn

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewVPNCommand returns the `settings vpn` parent. Phase 1 ships the
// read-only verbs (devices list / routes / public-domain-policy get);
// Phase 3c1 lands devices rename / delete / tags + route enable /
// disable + public-domain-policy set; Phase 3c2 will add the ACL +
// SSH/subroutes toggles.
func NewVPNCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN / Headscale (devices, routes, ACL, public-domain-policy)",
		Long: `Manage the per-Olares Headscale mesh: devices, routes, SSH/sub-routes ACLs,
and public-domain-policy.

Subcommands:
  devices list                                              (Phase 1)
  devices routes <device-id>                                (Phase 1)
  devices rename <device-id> <new-name>                     (Phase 3)
  devices delete <device-id>                                (Phase 3)
  devices tags set <device-id> --tag <name>...              (Phase 3)
  routes enable | disable <route-id>                        (Phase 3)
  public-domain-policy get                                  (Phase 1)
  public-domain-policy set --deny-all | --allow-all         (Phase 3)

Subcommands landing in a later Phase 3 slice:
  ssh status / enable / disable
  subroutes status / enable / disable
  acl <app> list / set
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewDevicesCommand(f))
	cmd.AddCommand(NewRoutesCommand(f))
	cmd.AddCommand(NewPublicDomainPolicyCommand(f))
	return cmd
}
