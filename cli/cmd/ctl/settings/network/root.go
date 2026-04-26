// Package network implements the `olares-cli settings network` subtree
// (Settings -> Network). Backed by user-service's bfl/network.controller.ts
// for reverse-proxy / FRP / external-network / SSL, plus
// terminusd.controller.ts for the hosts-file slice.
package network

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewNetworkCommand returns the `settings network` parent.
func NewNetworkCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Network settings (reverse-proxy, FRP, external-network, SSL, hosts-file)",
		Long: `Read and configure network plumbing: reverse-proxy mode, FRP server, the
external-network switch (owner-only), SSL toggles, and the system hosts-file.

Subcommands will be added in subsequent phases:
  Phase 1: reverse-proxy get, frp list, external-network get, hosts-file get
  Phase 4: reverse-proxy set, frp set, external-network set, ssl enable,
           hosts-file set

Note: external-network set and reverse-proxy set (FRP host write) are
owner-only; non-owner callers will hit a 403 from BFL.
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
