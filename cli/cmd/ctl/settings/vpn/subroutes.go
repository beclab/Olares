package vpn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/settings/internal/preflight"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

// `olares-cli settings vpn subroutes ...`
//
// Per-app sub-route enablement. Mirrors the SPA's Settings -> VPN ->
// "Allow access from sub-domains" toggle (apps/.../stores/settings/acl.ts:248-258).
// Three endpoints behind one cobra parent:
//
//	GET  /api/acl/subroutes/status                        -> BFL envelope wrapping []string
//	POST /api/acl/subroutes/enable     body {}            -> success / failure
//	POST /api/acl/subroutes/disable    body {}            -> success / failure
//
// Wire shape: BFL's handleGetTailScaleSubnet (framework/bfl/pkg/apis/
// settings/v1alpha1/handle_headscale.go:286-288) returns the per-app
// sub-route list via `response.Success(resp, subRoutes)`, i.e. a BFL
// envelope `{code, message, data: []string}`. user-service forwards the
// envelope verbatim; the SPA only "sees" the inner `[]string` because
// boot/axios.ts:163 strips `data.data` globally. CLI callers don't get
// that for free, so decodeSubroutesStatus has to unwrap defensively
// (same back-history as decodeSSHStatus).
//
// Render contract: the SPA's only visible artifact for this endpoint is
// a single toggle bound to `allow_subroutes = data && data.length > 0`
// (apps/.../stores/settings/acl.ts:257 + Vpn/VPNPage.vue:44). Mirror
// that in --output table: print the derived state + the route list. The
// raw inner data goes out under -o json so downstream tools see the
// same body the SPA's axios interceptor exposes to acl.ts.

func NewSubroutesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subroutes",
		Short: "Sub-route ACL toggle (Settings -> VPN -> Allow sub-domains)",
		Long: `Inspect or change whether sub-domain entrances are reachable across
the Headscale mesh. Mirrors the "Allow sub-domains" switch on the SPA's
VPN page.

Subcommands:
  status     show current sub-route ACL state (allow_subroutes + route list)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newSubroutesStatusCommand(f))
	cmd.AddCommand(newSubroutesEnableCommand(f))
	cmd.AddCommand(newSubroutesDisableCommand(f))
	return cmd
}

// subroutesStatus is the rendered shape for `vpn subroutes status` table
// view: the derived `allow_subroutes` flag (true iff the upstream
// returned a non-empty list, matching the SPA's `data && data.length > 0`
// rule) plus the route list itself. Never serialized — `-o json`
// round-trips the upstream's unwrapped data ([]string) verbatim.
type subroutesStatus struct {
	AllowSubRoutes bool
	Routes         []string
}

func newSubroutesStatusCommand(f *cmdutil.Factory) *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show current sub-route ACL state (allow_subroutes + route list)",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "show sub-routes ACL state"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSubroutesStatus(ctx, f, output), "show sub-routes ACL state")
		},
	}
	addOutputFlag(cmd, &output)
	return cmd
}

func runSubroutesStatus(ctx context.Context, f *cmdutil.Factory, outputRaw string) error {
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
	var raw json.RawMessage
	if err := pc.doer.DoJSON(ctx, "GET", "/api/acl/subroutes/status", nil, &raw); err != nil {
		return err
	}
	routes, err := decodeSubroutesStatus(raw)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		// Hand back exactly the inner []string the SPA's interceptor
		// exposes to acl.ts; nil collapses to `[]` for a stable shape
		// downstream jq pipelines can rely on.
		if routes == nil {
			routes = []string{}
		}
		return printJSON(os.Stdout, routes)
	default:
		return renderSubroutesStatus(os.Stdout, subroutesStatus{
			AllowSubRoutes: len(routes) > 0,
			Routes:         routes,
		})
	}
}

// decodeSubroutesStatus is the wire-shape adapter for
// `GET /api/acl/subroutes/status`. Same envelope back-history as
// decodeSSHStatus: BFL emits `{code, message, data: []string}` and
// user-service forwards it verbatim. The SPA "sees" the inner []string
// only because its global axios interceptor strips `data.data`
// (apps/packages/app/src/boot/axios.ts:163).
//
// We tolerate both wrapped and unwrapped shapes defensively so a future
// user-service rewrite that DOES strip the envelope here (the way it
// already does for /api/launcher-public-domain-access-policy) won't
// flip this command into "always empty" mode. `null` and `[]` both
// round-trip cleanly as an empty list — which is what acl.ts treats as
// `allow_subroutes=false`.
func decodeSubroutesStatus(raw json.RawMessage) ([]string, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	var env bflEnvelope
	if err := json.Unmarshal(trimmed, &env); err == nil && envelopeLooksWrapped(trimmed, env) {
		switch env.Code {
		case 0, 200:
		default:
			msg := strings.TrimSpace(env.Message)
			if msg == "" {
				msg = fmt.Sprintf("server returned code=%d", env.Code)
			}
			return nil, fmt.Errorf("GET /api/acl/subroutes/status: %s", msg)
		}
		if len(env.Data) == 0 || string(bytes.TrimSpace(env.Data)) == "null" {
			return nil, nil
		}
		var routes []string
		if err := json.Unmarshal(env.Data, &routes); err != nil {
			return nil, fmt.Errorf("decode acl subroutes status data: %w", err)
		}
		return routes, nil
	}
	var routes []string
	if err := json.Unmarshal(trimmed, &routes); err != nil {
		return nil, fmt.Errorf("decode acl subroutes status body: %w", err)
	}
	return routes, nil
}

func renderSubroutesStatus(w io.Writer, s subroutesStatus) error {
	state := "disabled"
	if s.AllowSubRoutes {
		state = "enabled"
	}
	if _, err := fmt.Fprintf(w, "%-18s %s\n", "Allow sub-routes:", state); err != nil {
		return err
	}
	if len(s.Routes) == 0 {
		_, err := fmt.Fprintf(w, "%-18s %s\n", "Routes:", "(none)")
		return err
	}
	if _, err := fmt.Fprintf(w, "%-18s (%d)\n", "Routes:", len(s.Routes)); err != nil {
		return err
	}
	for _, r := range s.Routes {
		if _, err := fmt.Fprintf(w, "  - %s\n", r); err != nil {
			return err
		}
	}
	return nil
}

func newSubroutesEnableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:    "enable",
		Short:  "permit sub-domain access across the Headscale mesh",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "enable sub-routes"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSubroutesToggle(ctx, f, true), "enable sub-routes")
		},
	}
}

func newSubroutesDisableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:    "disable",
		Short:  "block sub-domain access across the Headscale mesh",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			if err := preflight.Gate(ctx, f, whoami.RoleAdmin, "disable sub-routes"); err != nil {
				return err
			}
			return preflight.Wrap(ctx, f, runSubroutesToggle(ctx, f, false), "disable sub-routes")
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
