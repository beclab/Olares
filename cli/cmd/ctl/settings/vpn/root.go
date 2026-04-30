// Package vpn implements the `olares-cli settings vpn` subtree (Settings ->
// VPN). Backed by user-service's headscale/headscale.controller.ts +
// bfl/acl.controller.ts + the public-domain-policy slice of
// bfl/network.controller.ts.
package vpn

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewVPNCommand returns the `settings vpn` parent: Headscale device
// + route management, SSH / sub-routes ACL toggles, the per-app ACL
// editor, and public-domain-policy.
func NewVPNCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN / Headscale (devices, routes, ACL, public-domain-policy)",
		Long: `Manage the per-Olares Headscale mesh: devices, routes, SSH/sub-routes ACLs,
and public-domain-policy.

Subcommands:
  devices list
  devices routes <device-id>
  devices rename <device-id> <new-name>
  devices delete <device-id>
  devices tags set <device-id> --tag <name>...
  routes enable | disable <route-id>
  ssh status | enable | disable
  subroutes status | enable | disable
  acl get <app>
  acl set <app> [--tcp PORT...] [--udp PORT...]
  acl add <app> [--tcp PORT...] [--udp PORT...]
  acl remove <app> [--tcp PORT...] [--udp PORT...]
  acl clear <app>
  public-domain-policy get
  public-domain-policy set --deny-all | --allow-all
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
