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
// disable + public-domain-policy set; Phase 3c2 wires the SSH and
// sub-routes ACL toggles; Phase 3c3 finishes the page with the per-app
// ACL editor (`vpn acl <app> get|set|add|remove|clear`).
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
  ssh status | enable | disable                             (Phase 3)
  subroutes status | enable | disable                       (Phase 3)
  acl get <app>                                             (Phase 3)
  acl set <app> [--tcp PORT...] [--udp PORT...]             (Phase 3)
  acl add <app> [--tcp PORT...] [--udp PORT...]             (Phase 3)
  acl remove <app> [--tcp PORT...] [--udp PORT...]          (Phase 3)
  acl clear <app>                                           (Phase 3)
  public-domain-policy get                                  (Phase 1)
  public-domain-policy set --deny-all | --allow-all         (Phase 3)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewDevicesCommand(f))
	cmd.AddCommand(NewRoutesCommand(f))
	cmd.AddCommand(NewSSHCommand(f))
	cmd.AddCommand(NewSubroutesCommand(f))
	cmd.AddCommand(NewACLCommand(f))
	cmd.AddCommand(NewPublicDomainPolicyCommand(f))
	return cmd
}
