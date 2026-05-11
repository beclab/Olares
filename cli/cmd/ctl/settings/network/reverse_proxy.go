package network

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
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
// `set --mode <mode>` writes the same struct back, with field-level
// overrides for FRP / public-IP modes.
func NewReverseProxyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reverse-proxy",
		Short: "reverse-proxy / FRP / Cloudflare tunnel (Settings -> Network)",
		Long: `Inspect or change the reverse-proxy mode that publishes Olares to the
public internet.

Subcommands:
  get   show the current configuration
  set   change reverse-proxy mode + fields
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newReverseProxyGetCommand(f))
	cmd.AddCommand(newReverseProxySetCommand(f))
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
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "show reverse-proxy config"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runReverseProxyGet(ctx, f, output), "show reverse-proxy config")
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

// `settings network reverse-proxy set` — change the reverse-proxy
// mode (and FRP fields when --mode frp). The SPA sends the FULL
// ReverseProxyConfig back on every save (stores/settings/network.ts:
// updateReverseProxy). To avoid clobbering fields the user didn't pass
// we read-modify-write: GET the current config, overlay --mode +
// per-field flags, POST back.
//
// The four canonical modes follow the SPA's mental model:
//
//	public-ip          enable_cloudflare_tunnel=false, enable_frp=false, ip set
//	frp                enable_cloudflare_tunnel=false, enable_frp=true
//	cloudflare-tunnel  enable_cloudflare_tunnel=true,  enable_frp=false
//	off                enable_cloudflare_tunnel=false, enable_frp=false, ip cleared
//
// `external_network_off` is owner-only and is surfaced read-only here;
// flipping the master switch lives elsewhere (network external-network
// command).

func newReverseProxySetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		mode          string
		ip            string
		frpServer     string
		frpPort       int
		frpAuthMethod string
		frpAuthToken  string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "change reverse-proxy mode + fields",
		Long: `Change the reverse-proxy mode published in BFL.

Modes:
  --mode public-ip          publish via a literal public IP (use --ip)
  --mode frp                publish via an FRP tunnel       (use --frp-* flags)
  --mode cloudflare-tunnel  publish via Cloudflare tunnel
  --mode off                stop publishing (clears IP + tunnel flags)

Examples:
  olares-cli settings network reverse-proxy set --mode frp \
    --frp-server frp.example.com --frp-port 7000 \
    --frp-auth-method token --frp-auth-token "$FRP_TOKEN"

  olares-cli settings network reverse-proxy set --mode public-ip --ip 203.0.113.5

  olares-cli settings network reverse-proxy set --mode cloudflare-tunnel

The CLI does a read-modify-write so unrelated fields the SPA had set
(e.g. an FRP token you don't want to type again when switching modes)
stay intact unless you explicitly override them with --frp-*.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "set reverse-proxy mode"); err != nil {
				return err
			}
			err := runReverseProxySet(ctx, f, reverseProxySetFlags{
				Mode:          mode,
				IP:            ip,
				FRPServer:     frpServer,
				FRPPort:       frpPort,
				FRPAuthMethod: frpAuthMethod,
				FRPAuthToken:  frpAuthToken,
				FRPPortSet:    c.Flags().Changed("frp-port"),
				IPSet:         c.Flags().Changed("ip"),
				ServerSet:     c.Flags().Changed("frp-server"),
				MethodSet:     c.Flags().Changed("frp-auth-method"),
				TokenSet:      c.Flags().Changed("frp-auth-token"),
			})
			return preflight.Wrap(ctx, f, err, "set reverse-proxy mode")
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "", "reverse-proxy mode (public-ip|frp|cloudflare-tunnel|off)")
	cmd.Flags().StringVar(&ip, "ip", "", "public IP (used with --mode public-ip)")
	cmd.Flags().StringVar(&frpServer, "frp-server", "", "FRP server hostname")
	cmd.Flags().IntVar(&frpPort, "frp-port", 0, "FRP server port")
	cmd.Flags().StringVar(&frpAuthMethod, "frp-auth-method", "", "FRP auth method (typically 'token' or 'jws')")
	cmd.Flags().StringVar(&frpAuthToken, "frp-auth-token", "", "FRP auth token (only when --frp-auth-method=token)")
	_ = cmd.MarkFlagRequired("mode")
	return cmd
}

type reverseProxySetFlags struct {
	Mode                                                                string
	IP, FRPServer, FRPAuthMethod, FRPAuthToken                          string
	FRPPort                                                             int
	IPSet, ServerSet, FRPPortSet, MethodSet, TokenSet                   bool
}

func runReverseProxySet(ctx context.Context, f *cmdutil.Factory, flags reverseProxySetFlags) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var current reverseProxyConfig
	if err := doGetEnvelope(ctx, pc.doer, "/api/reverse-proxy", &current); err != nil {
		return err
	}
	next, err := applyReverseProxyMode(current, flags)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", "/api/reverse-proxy", next, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "reverse-proxy set to mode=%s\n", reverseProxyMode(next))
	return nil
}

// applyReverseProxyMode collapses the user's --mode + per-field flags
// onto the existing config. Pure function so it's straightforward to
// unit test the mode transitions.
func applyReverseProxyMode(current reverseProxyConfig, flags reverseProxySetFlags) (reverseProxyConfig, error) {
	out := current
	switch strings.ToLower(strings.TrimSpace(flags.Mode)) {
	case "public-ip", "publicip", "public_ip":
		out.EnableFRP = false
		out.EnableCloudFlareTunnel = false
		if flags.IPSet {
			out.IP = strings.TrimSpace(flags.IP)
		}
		if out.IP == "" {
			return out, fmt.Errorf("--mode public-ip requires --ip <addr> (or a previously-saved IP in the current config)")
		}
	case "frp":
		out.EnableFRP = true
		out.EnableCloudFlareTunnel = false
		if flags.ServerSet {
			out.FRPServer = strings.TrimSpace(flags.FRPServer)
		}
		if flags.FRPPortSet {
			out.FRPPort = flags.FRPPort
		}
		if flags.MethodSet {
			out.FRPAuthMethod = strings.TrimSpace(flags.FRPAuthMethod)
		}
		if flags.TokenSet {
			out.FRPAuthToken = flags.FRPAuthToken
		}
		if out.FRPServer == "" {
			return out, fmt.Errorf("--mode frp requires --frp-server (or a previously-saved server in the current config)")
		}
	case "cloudflare-tunnel", "cloudflare", "cf":
		out.EnableFRP = false
		out.EnableCloudFlareTunnel = true
	case "off", "":
		out.EnableFRP = false
		out.EnableCloudFlareTunnel = false
		out.IP = ""
	default:
		return out, fmt.Errorf("unknown --mode %q (allowed: public-ip, frp, cloudflare-tunnel, off)", flags.Mode)
	}
	return out, nil
}
