package vpn

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// `olares-cli settings vpn routes ...`
//
// Per-route enable / disable. Backed by user-service's headscale
// controller (headscale.controller.ts:50-65), which forwards to:
//
//   POST /headscale/routes/<route-id>/enable
//   POST /headscale/routes/<route-id>/disable
//
// Body: empty {} (the SPA sends `{}` explicitly; we match it).
//
// Listing routes still lives under `vpn devices routes <device-id>` —
// Headscale exposes per-device-route listing via the machine endpoint,
// not via /routes. The split mirrors the upstream API surface 1:1.

func NewRoutesCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routes",
		Short: "enable or disable advertised Headscale routes",
		Long: `Enable or disable a route a Headscale device is advertising.

To find route IDs, list a device's routes first:
  olares-cli settings vpn devices routes <device-id>

Subcommands:
  enable   <route-id>   permit traffic via this route
  disable  <route-id>   block traffic via this route (route stays advertised)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(newRoutesEnableCommand(f))
	cmd.AddCommand(newRoutesDisableCommand(f))
	return cmd
}

func newRoutesEnableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <route-id>",
		Short: "enable a Headscale route",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteToggle(c.Context(), f, args[0], true)
		},
	}
}

func newRoutesDisableCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <route-id>",
		Short: "disable a Headscale route (route stays advertised)",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runRouteToggle(c.Context(), f, args[0], false)
		},
	}
}

func runRouteToggle(ctx context.Context, f *cmdutil.Factory, routeID string, enable bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	routeID = strings.TrimSpace(routeID)
	if routeID == "" {
		return fmt.Errorf("route-id is required")
	}

	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	if err := doRouteToggleViaDoer(ctx, pc.doer, routeID, enable); err != nil {
		return err
	}
	verb := "disable"
	if enable {
		verb = "enable"
	}
	fmt.Fprintf(os.Stdout, "%sd route %s\n", verb, routeID)
	return nil
}

// doRouteToggleViaDoer is the wire-only core of `vpn routes
// enable|disable`. SPA sends an explicit empty {} — we match it even
// though Headscale doesn't read the body. Some routers in the proxy
// path strip Content-Length when there's no body; sending {} keeps the
// request well-formed for any conservative middleware in front of
// headscale.
func doRouteToggleViaDoer(ctx context.Context, d Doer, routeID string, enable bool) error {
	verb := "disable"
	if enable {
		verb = "enable"
	}
	path := "/headscale/routes/" + url.PathEscape(routeID) + "/" + verb
	return d.DoJSON(ctx, "POST", path, struct{}{}, nil)
}
