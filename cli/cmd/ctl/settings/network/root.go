// Package network implements the `olares-cli settings network` subtree
// (Settings -> Network). Backed by user-service's bfl/network.controller.ts
// for reverse-proxy / FRP / external-network / SSL, plus
// terminusd.controller.ts for the hosts-file slice.
package network

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewNetworkCommand returns the `settings network` parent: read-only
// inspection of every sub-area plus the reverse-proxy mode write. The
// remaining mutating endpoints require a JWS-signed device-id header
// the CLI doesn't yet produce, so they're out of scope today.
func NewNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Network settings (reverse-proxy, FRP, external-network, SSL, hosts-file)",
		Long: `Read and configure network plumbing: reverse-proxy mode, FRP server, the
external-network switch (owner-only), SSL toggles, and the system hosts-file.

Subcommands:
  reverse-proxy get
  reverse-proxy set --mode <public-ip|frp|cloudflare-tunnel|off> [...]
  frp list
  external-network get
  ssl status
  hosts-file get

Out of scope until a JWS key sourcing path exists:
  frp set, external-network set, ssl enable / disable / update, hosts-file set

Note: reverse-proxy set is owner-only; non-owner callers will hit a
403 from BFL.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewReverseProxyCommand(f))
	cmd.AddCommand(NewFRPCommand(f))
	cmd.AddCommand(NewExternalNetworkCommand(f))
	cmd.AddCommand(NewSSLCommand(f))
	cmd.AddCommand(NewHostsFileCommand(f))
	return cmd
}
