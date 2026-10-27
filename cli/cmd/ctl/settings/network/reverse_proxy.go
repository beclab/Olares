package network

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
	"github.com/spf13/cobra"
)

const reverseProxyPath = "/api/network/reverse-proxy"

func NewReverseProxyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{Use: "reverse-proxy", Short: "Inspect or configure the machine's public entrance", SilenceUsage: true}
	cmd.AddCommand(newReverseProxyGetCommand(f), newReverseProxySetCommand(f))
	return cmd
}

func newReverseProxyGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use: "get", Short: "Show reverse-proxy configuration and status", Args: cobra.NoArgs,
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

type reverseProxyConfig struct {
	Mode          string `json:"mode"`
	FRPServer     string `json:"frp_server,omitempty"`
	FRPPort       int    `json:"frp_port,omitempty"`
	FRPAuthMethod string `json:"frp_auth_method,omitempty"`
	FRPAuthToken  string `json:"frp_auth_token,omitempty"`
	DNSTarget     string `json:"dns_target,omitempty"`
}

type reverseProxyView struct {
	Config       reverseProxyConfig `json:"config"`
	Username     string             `json:"username,omitempty"`
	HasAuthToken bool               `json:"has_auth_token"`
	Status       struct {
		Phase     string `json:"phase"`
		Message   string `json:"message,omitempty"`
		DNSTarget string `json:"dns_target,omitempty"`
	} `json:"status"`
	Entrance struct {
		LocalIP       string `json:"local_ip,omitempty"`
		PublicIP      string `json:"public_ip,omitempty"`
		PublicIPError string `json:"public_ip_error,omitempty"`
	} `json:"entrance"`
}

func runReverseProxyGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	format, err := parseFormat(outputRaw)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var view reverseProxyView
	if err := doGetEnvelope(ctx, pc.doer, reverseProxyPath, &view); err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, view)
	}
	return renderReverseProxy(os.Stdout, view)
}

func renderReverseProxy(w io.Writer, view reverseProxyView) error {
	rows := [][2]string{
		{"Mode", nonEmpty(view.Config.Mode)},
		{"Status", nonEmpty(view.Status.Phase)},
		{"Message", nonEmpty(view.Status.Message)},
		{"Public IP", nonEmpty(view.Entrance.PublicIP)},
		{"Public IP Error", nonEmpty(view.Entrance.PublicIPError)},
		{"DNS Target", nonEmpty(view.Status.DNSTarget)},
		{"FRP Server", nonEmpty(view.Config.FRPServer)},
		{"FRP Port", fmt.Sprintf("%d", view.Config.FRPPort)},
		{"FRP Auth Method", nonEmpty(view.Config.FRPAuthMethod)},
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "%-22s %s\n", row[0]+":", row[1]); err != nil {
			return err
		}
	}
	return nil
}

func newReverseProxySetCommand(f *cmdutil.Factory) *cobra.Command {
	var flags reverseProxySetFlags
	cmd := &cobra.Command{
		Use: "set", Short: "Configure the machine's public entrance",
		Long: "Set --mode none, public-ip or frp. Public IP mode detects the address automatically. FRP configuration may be saved before a domain is bound.", Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "set reverse-proxy mode"); err != nil {
				return err
			}
			flags.ServerSet = c.Flags().Changed("frp-server")
			flags.FRPPortSet = c.Flags().Changed("frp-port")
			flags.MethodSet = c.Flags().Changed("frp-auth-method")
			flags.TokenSet = c.Flags().Changed("frp-auth-token")
			return preflight.Wrap(ctx, f, runReverseProxySet(ctx, f, flags), "set reverse-proxy mode")
		},
	}
	cmd.Flags().StringVar(&flags.Mode, "mode", "", "entrance mode (none|public-ip|frp)")
	cmd.Flags().StringVar(&flags.FRPServer, "frp-server", "", "FRP server hostname")
	cmd.Flags().IntVar(&flags.FRPPort, "frp-port", 0, "FRP server port")
	cmd.Flags().StringVar(&flags.FRPAuthMethod, "frp-auth-method", "", "FRP auth method (token|jws)")
	cmd.Flags().StringVar(&flags.FRPAuthToken, "frp-auth-token", "", "FRP auth token")
	_ = cmd.MarkFlagRequired("mode")
	return cmd
}

type reverseProxySetFlags struct {
	Mode, FRPServer, FRPAuthMethod, FRPAuthToken string
	FRPPort                                      int
	ServerSet, FRPPortSet, MethodSet, TokenSet   bool
}

func runReverseProxySet(ctx context.Context, f *cmdutil.Factory, flags reverseProxySetFlags) error {
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var current reverseProxyView
	if err := doGetEnvelope(ctx, pc.doer, reverseProxyPath, &current); err != nil {
		return err
	}
	next, err := applyReverseProxyMode(current.Config, flags)
	if err != nil {
		return err
	}
	if err := doMutateEnvelope(ctx, pc.doer, "POST", reverseProxyPath, next, nil); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "reverse-proxy set to mode=%s\n", next.Mode)
	return nil
}

func applyReverseProxyMode(current reverseProxyConfig, flags reverseProxySetFlags) (reverseProxyConfig, error) {
	mode := strings.ToLower(strings.TrimSpace(flags.Mode))
	switch mode {
	case "none", "off":
		return reverseProxyConfig{Mode: "none"}, nil
	case "public-ip":
		return reverseProxyConfig{Mode: "public-ip"}, nil
	case "frp":
		current.Mode = mode
		if flags.ServerSet {
			current.FRPServer = strings.TrimSpace(flags.FRPServer)
			current.DNSTarget = ""
		}
		if flags.FRPPortSet {
			current.FRPPort = flags.FRPPort
		}
		if flags.MethodSet {
			current.FRPAuthMethod = strings.TrimSpace(flags.FRPAuthMethod)
		}
		if flags.TokenSet {
			current.FRPAuthToken = flags.FRPAuthToken
		}
		if current.FRPServer == "" {
			return current, fmt.Errorf("--mode frp requires --frp-server")
		}
		return current, nil
	default:
		return current, fmt.Errorf("unknown --mode %q (allowed: none, public-ip, frp)", flags.Mode)
	}
}
