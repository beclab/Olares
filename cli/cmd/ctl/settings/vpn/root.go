// Package vpn implements the `olares-cli settings vpn` subtree (Settings ->
// VPN). Backed by user-service's headscale/headscale.controller.ts +
// bfl/acl.controller.ts + the public-domain-policy slice of
// bfl/network.controller.ts.
package vpn

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewVPNCommand returns the `settings vpn` parent.
func NewVPNCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpn",
		Short: "VPN / Headscale (devices, routes, ACL, public-domain-policy)",
		Long: `Manage the per-Olares Headscale mesh: devices, routes, SSH/sub-routes ACLs,
and public-domain-policy.

Subcommands will be added in subsequent phases:
  Phase 1: devices list, routes list
  Phase 3: devices rename / delete / tags, routes enable / disable,
           ssh status / enable / disable, subroutes status / enable / disable,
           public-domain-policy get / set
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
