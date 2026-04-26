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
// Phase 1 ships GET; Phase 3 lands `set --deny-all/--allow-all`.
func NewPublicDomainPolicyCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-domain-policy",
		Short: "public-domain access policy (Settings -> VPN)",
		Long: `Inspect or change the public-domain access policy. When deny_all is 1,
the Olares mesh blocks public-domain access for entrances that haven't
been individually whitelisted.

Subcommands:
  get   show the current policy                           (Phase 1)
  set   change the policy (--deny-all / --allow-all)      (Phase 3)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newPublicDomainPolicyGetCommand(f))
	cmd.AddCommand(newPublicDomainPolicySetCommand(f))
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

// `vpn public-domain-policy set` — flips deny_all between 0 and 1 via
// POST /api/launcher-public-domain-access-policy. The SPA's
// toggleHeadScaleStatus does the same thing on the VPN page's master
// switch (stores/settings/headscale.ts:43-53).
//
// Mutually-exclusive flags rather than a positional value: --deny-all
// and --allow-all read more clearly than `set 1` / `set 0`, and the
// admin/owner reading the help should never have to remember which
// integer means what. We still accept exactly one of the two so the
// command is unambiguous and round-trips cleanly through JSON output.
func newPublicDomainPolicySetCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		denyAll  bool
		allowAll bool
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "set the public-domain access policy",
		Long: `Set the public-domain access policy. Pass exactly one of:

  --deny-all     block public-domain access for non-whitelisted entrances
  --allow-all    permit public-domain access (default Olares behavior)

The change takes effect immediately for new connections. Existing
sessions stay open until their next renegotiation.

Examples:
  olares-cli settings vpn public-domain-policy set --deny-all
  olares-cli settings vpn public-domain-policy set --allow-all
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runPublicDomainPolicySet(c.Context(), f, denyAll, allowAll)
		},
	}
	cmd.Flags().BoolVar(&denyAll, "deny-all", false, "block public-domain access for non-whitelisted entrances (deny_all=1)")
	cmd.Flags().BoolVar(&allowAll, "allow-all", false, "permit public-domain access for all entrances (deny_all=0)")
	return cmd
}

func runPublicDomainPolicySet(ctx context.Context, f *cmdutil.Factory, denyAll, allowAll bool) error {
	if ctx == nil {
		ctx = context.Background()
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	value, label, err := resolvePolicyFlag(denyAll, allowAll)
	if err != nil {
		return err
	}
	if err := doPolicySetViaDoer(ctx, pc.doer, value); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "public-domain-policy set to %s (deny_all=%d)\n", label, value)
	return nil
}

// resolvePolicyFlag enforces mutual exclusion between --deny-all and
// --allow-all and turns the boolean pair into the wire integer (0 = allow,
// 1 = deny) plus a friendly label for the success message.
func resolvePolicyFlag(denyAll, allowAll bool) (int, string, error) {
	if denyAll == allowAll {
		return 0, "", fmt.Errorf("pass exactly one of --deny-all or --allow-all")
	}
	if denyAll {
		return 1, "deny-all", nil
	}
	return 0, "allow-all", nil
}

// doPolicySetViaDoer is the wire-only core of `vpn public-domain-policy
// set`. The value must already be 0 or 1 (use resolvePolicyFlag);
// callers shouldn't pass arbitrary integers.
func doPolicySetViaDoer(ctx context.Context, d Doer, value int) error {
	body := publicDomainPolicy{DenyAll: value}
	return d.DoJSON(ctx, "POST", "/api/launcher-public-domain-access-policy", body, nil)
}
