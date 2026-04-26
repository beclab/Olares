package vpn

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings vpn public-domain-policy ...`
//
// The "deny all public-domain access" toggle from the SPA's VPN page
// (apps/.../stores/settings/headscale.ts:34-53). One state flag, two
// values:
//
//	deny_all = 1   block public-domain access entirely
//	deny_all = 0   permit public-domain access (default)
//
// Backend handler:
//   user-service/src/bfl/network.controller.ts:41 (@Get) /
//   user-service/src/bfl/network.controller.ts:54 (@Post)
//
// user-service forwards data.data from BFL's
// /bfl/settings/v1alpha1/launcher-public-domain-access-policy, so the
// CLI receives the BFL inner data shape directly:
//
//   { "deny_all": <0|1> }
//
// (NO outer envelope — user-service unwraps it server-side, and the
// SPA's interceptor passes the body through as-is for the same reason.)
//
// Phase 1 ships GET; Phase 3 will add `set --deny-all/--allow-all`.
func NewPublicDomainPolicyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-domain-policy",
		Short: "public-domain access policy (Settings -> VPN)",
		Long: `Inspect or change the public-domain access policy. When deny_all is 1,
the Olares mesh blocks public-domain access for entrances that haven't
been individually whitelisted.

Subcommands:
  get   show the current policy                           (Phase 1)

Subcommands landing in Phase 3:
  set   change the policy (set --deny-all / --allow-all)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPublicDomainPolicyGetCommand(f))
	return cmd
}

func newPublicDomainPolicyGetCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "show the current public-domain access policy",
		Long: `Show the current launcher-public-domain-access-policy as a single row:

  Deny All:  yes / no
  Raw:       0 / 1

Pass --output json to get the raw upstream {"deny_all": 0|1} body for
scripting.
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPublicDomainPolicyGet(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

// publicDomainPolicy is the upstream wire body. We intentionally keep
// `deny_all` typed as int rather than bool because the upstream uses the
// 0/1 numeric convention everywhere and we don't want to double-translate
// for --output json.
type publicDomainPolicy struct {
	DenyAll int `json:"deny_all"`
}

func runPublicDomainPolicyGet(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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

	var resp publicDomainPolicy
	if err := pc.doer.DoJSON(ctx, "GET", "/api/launcher-public-domain-access-policy", nil, &resp); err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return printJSON(os.Stdout, resp)
	default:
		return renderPublicDomainPolicy(os.Stdout, resp)
	}
}

func renderPublicDomainPolicy(w io.Writer, p publicDomainPolicy) error {
	rows := [][2]string{
		{"Deny All", boolStr(p.DenyAll == 1)},
		{"Raw", fmt.Sprintf("%d", p.DenyAll)},
	}
	for _, r := range rows {
		if _, err := fmt.Fprintf(w, "%-10s %s\n", r[0]+":", r[1]); err != nil {
			return err
		}
	}
	return nil
}
