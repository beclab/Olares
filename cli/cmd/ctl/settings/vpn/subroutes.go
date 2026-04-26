package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings vpn subroutes ...`
//
// Per-app sub-route enablement. Mirrors the SPA's Settings -> VPN ->
// "Allow access from sub-domains" toggle (apps/.../stores/settings/acl.ts:243-280).
// Three endpoints behind one cobra parent:
//
//	GET  /api/acl/subroutes/status                        -> opaque upstream JSON
//	POST /api/acl/subroutes/enable     body {}            -> success / failure
//	POST /api/acl/subroutes/disable    body {}            -> success / failure
//
// The status response shape isn't strongly typed in the SPA (it stuffs
// the raw object into `subroutes`), so we render the opaque JSON
// verbatim. --output table just prints "<JSON body>" — no point
// hand-rolling a key/value table for a shape that shifts under us.

func NewSubroutesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subroutes",
		Short: "Sub-route ACL toggle (Settings -> VPN -> Allow sub-domains)",
		Long: `Inspect or change whether sub-domain entrances are reachable across
the Headscale mesh. Mirrors the "Allow sub-domains" switch on the SPA's
VPN page.

Subcommands:
  status     dump the current sub-route ACL state (raw JSON)
  enable     permit sub-domain access across the mesh
  disable    block sub-domain access across the mesh
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSubroutesStatusCommand(f))
	cmd.AddCommand(newSubroutesEnableCommand(f))
	cmd.AddCommand(newSubroutesDisableCommand(f))
	return cmd
}

func newSubroutesStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "dump current sub-route ACL state (raw upstream JSON)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSubroutesStatus(c.Context(), f, output)
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSubroutesStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := parseFormat(outputRaw); err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	var resp json.RawMessage
	if err := pc.doer.DoJSON(ctx, "GET", "/api/acl/subroutes/status", nil, &resp); err != nil {
		return err
	}
	if len(resp) == 0 {
		resp = json.RawMessage("null")
	}
	var v interface{}
	if err := json.Unmarshal(resp, &v); err != nil {
		return fmt.Errorf("decode subroutes status: %w", err)
	}
	return printJSON(os.Stdout, v)
}

func newSubroutesEnableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "permit sub-domain access across the Headscale mesh",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSubroutesToggle(c.Context(), f, true)
		},
	}
}

func newSubroutesDisableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "block sub-domain access across the Headscale mesh",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runSubroutesToggle(c.Context(), f, false)
		},
	}
}

func runSubroutesToggle(ctx context.Context, f *cmdutil.Factory, enable bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doSubroutesToggleViaDoer(ctx, pc.doer, enable); err != nil {
		return err
	}
	verb := "disable"
	if enable {
		verb = "enable"
	}
	fmt.Fprintf(os.Stdout, "%sd sub-route ACL\n", verb)
	return nil
}

// doSubroutesToggleViaDoer is the wire-only core of `vpn subroutes
// enable|disable`. As with SSH and routes, we send {} explicitly to
// match the SPA's request shape.
func doSubroutesToggleViaDoer(ctx context.Context, d Doer, enable bool) error {
	verb := "disable"
	if enable {
		verb = "enable"
	}
	path := "/api/acl/subroutes/" + verb
	return d.DoJSON(ctx, "POST", path, struct{}{}, nil)
}
