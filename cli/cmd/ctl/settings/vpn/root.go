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
// Phase 3 lands the mutating verbs.
func NewVPNCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN / Headscale (devices, routes, ACL, public-domain-policy)",
		Long: `Manage the per-Olares Headscale mesh: devices, routes, SSH/sub-routes ACLs,
and public-domain-policy.

Subcommands:
  devices list                          (Phase 1)
  devices routes <device-id>            (Phase 1)
  public-domain-policy get              (Phase 1)

Subcommands landing in subsequent phases:
  Phase 3: devices rename / delete / tags, routes enable / disable,
           ssh status / enable / disable, subroutes status / enable / disable,
           public-domain-policy set
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewDevicesCommand(f))
	cmd.AddCommand(NewPublicDomainPolicyCommand(f))
	return cmd
}
