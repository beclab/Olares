package network

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings network reverse-proxy ...`
//
// Backed by user-service's /api/reverse-proxy, which forwards
// /bfl/settings/v1alpha1/reverse-proxy. The shape is BFL's
// ReverseProxyConfig embedding FRPConfig:
//
//   {
//     "frp_server": "...", "frp_port": 7000,
//     "frp_auth_method": "token", "frp_auth_token": "***",
//     "ip": "1.2.3.4",
//     "enable_cloudflare_tunnel": false,
//     "enable_frp": true,
//     "external_network_off": false   // owner-only flag, surface only
//   }
//
// Phase 1 ships GET; Phase 4 will land set verbs (set-public-ip,
// set-frp, enable-cloudflare-tunnel, etc).
func NewReverseProxyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reverse-proxy",
		Short: "reverse-proxy / FRP / Cloudflare tunnel (Settings -> Network)",
		Long: `Inspect or change the reverse-proxy mode that publishes Olares to the
public internet.

Subcommands:
  get   show the current configuration                    (Phase 1)

Subcommands landing in Phase 4:
  set   change reverse-proxy fields (FRP server / Cloudflare / IP)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newReverseProxyGetCommand(f))
	return cmd
}

func newReverseProxyGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the current reverse-proxy configuration",
		Long: `Show the current reverse-proxy configuration as a key/value table:

  Mode:                public-ip / frp / cloudflare / off
  Public IP:           literal IP if set
  FRP Server / Port:   when FRP is enabled
  FRP Auth Method:     "token" or "" (Olares-tunnel managed)
  Cloudflare Tunnel:   yes / no
  External Network:    on / off (owner-only flag, surface only)

The ` + "`frp_auth_token`" + ` field is intentionally NOT printed in the table
view; pass --output json if you really need the raw token.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runReverseProxyGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// reverseProxyConfig mirrors BFL's settings.v1alpha1.ReverseProxyConfig
// (with FRPConfig flattened).
type reverseProxyConfig struct {
	FRPServer              string `json:"frp_server"`
	FRPPort                int    `json:"frp_port"`
	FRPAuthMethod          string `json:"frp_auth_method"`
	FRPAuthToken           string `json:"frp_auth_token"`
	IP                     string `json:"ip"`
	EnableCloudFlareTunnel bool   `json:"enable_cloudflare_tunnel"`
	EnableFRP              bool   `json:"enable_frp"`
	ExternalNetworkOff     bool   `json:"external_network_off"`
}

func runReverseProxyGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}

	var cfg reverseProxyConfig
	if err := doGetEnvelope(ctx, pc.doer, "/api/reverse-proxy", &cfg); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, cfg)
	default:
		return renderReverseProxy(os.Stdout, cfg)
	}
}

// reverseProxyMode collapses the boolean+IP triple into the same modes
// the SPA's Reverse Proxy page surfaces.
func reverseProxyMode(c reverseProxyConfig) string {
	switch {
	case c.ExternalNetworkOff:
		return "off (external network disabled)"
	case c.EnableCloudFlareTunnel:
		return "cloudflare-tunnel"
	case c.EnableFRP:
		return "frp"
	case c.IP != "":
		return "public-ip"
	default:
		return "unset"
	}
}

func renderReverseProxy(w io.Writer, c reverseProxyConfig) error {
	rows := [][2]string{
		{"Mode", reverseProxyMode(c)},
		{"Public IP", nonEmpty(c.IP)},
		{"Cloudflare Tunnel", boolStr(c.EnableCloudFlareTunnel)},
		{"FRP Enabled", boolStr(c.EnableFRP)},
		{"FRP Server", nonEmpty(c.FRPServer)},
		{"FRP Port", fmt.Sprintf("%d", c.FRPPort)},
		{"FRP Auth Method", nonEmpty(c.FRPAuthMethod)},
		{"External Network Off", boolStr(c.ExternalNetworkOff)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-22s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
