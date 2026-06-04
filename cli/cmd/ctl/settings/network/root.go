// Package network implements the `olares-cli settings network` subtree
// (Settings -> Network). Backed by user-service's bfl/network.controller.ts
// for reverse-proxy / FRP / SSL, plus terminusd.controller.ts for the
// hosts-file slice.
//
// The external-network master switch (BFL /api/external-network) has an
// implementation in external_network.go but is currently NOT registered on
// the command tree: the feature has no SPA / TermiPass UI surfacing it yet,
// and the matching write requires a JWS-signed device-id header the CLI
// can't produce, so shipping a read-only verb in isolation only confuses
// operators. Re-add cmd.AddCommand(NewExternalNetworkCommand(f)) below
// once the UI lands or the JWS key sourcing path is wired.
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
		Short: "Network settings (reverse-proxy, FRP, SSL, hosts-file)",
		Long: `Read and configure network plumbing: reverse-proxy mode, FRP server,
SSL toggles, and the system hosts-file.

Subcommands:
  reverse-proxy get
  frp list
  hosts-file get

Out of scope until a JWS key sourcing path exists:
  frp set, ssl enable / disable / update, hosts-file set

Note: reverse-proxy set is owner-only; non-owner callers will hit a
403 from BFL.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewReverseProxyCommand(f))
	cmd.AddCommand(NewFRPCommand(f))
	cmd.AddCommand(NewSSLCommand(f))
	cmd.AddCommand(NewHostsFileCommand(f))

	// Overlay gateway is a 1.12.6+ feature. Hide the whole subtree on older
	// backends (advisory, driven by the locally cached version so --help stays
	// offline); the runtime capability gate still returns a precise
	// "requires Olares >= 1.12.6" error if it's invoked anyway.
	overlay := NewOverlayCommand(f)
	if v, ok := f.CachedOlaresBackendVersion(); ok && !supportsOverlay(v) {
		overlay.Hidden = true
	}
	cmd.AddCommand(overlay)
	return cmd
}
