package cluster

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
	"github.com/beclab/Olares/cli/pkg/clusterctx"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewContextCommand: `olares-cli cluster context [--refresh] [-o table|json]`
//
// Reports the active profile's identity from the ControlHub side —
// who the server says you are (username, globalrole), which workspaces
// and system namespaces you can see, which clusters you've been
// granted, and the overall cluster role. Defaults to the locally
// cached value (config.json's clusterContext / clusterContextRefreshedAt
// fields). Use --refresh to force a server roundtrip against
// /capi/app/detail — the same endpoint the ControlHub SPA's
// AppDetail store hits in
// apps/packages/app/src/apps/controlHub/stores/AppDetail.ts.
//
// This is the moral counterpart of `olares-cli profile whoami`:
// whoami answers "who am I in BFL terms?" (OwnerRole on the desktop
// ingress), context answers "who am I in K8s/KubeSphere terms?"
// (globalrole on the ControlHub ingress). Both default to the
// cached value to keep "what scope do I have right now?" a one-key
// answer; both refresh on demand via --refresh.
//
// IMPORTANT: this cache is for display only. NO `cluster ...` verb
// reads it to decide whether a call is allowed — the server is the
// only authority. If you suspect role drift after the cache was
// written, run with --refresh to reconcile.
//
// Output:
//   - table (default): identity + freshness + workspaces + system
//     namespaces + granted clusters, plus a "globalrole changed:
//     X -> Y" notice when --refresh detected drift.
//   - json: clusterctx.Display verbatim, so scripts can branch on
//     {"globalrole":"platform-admin"} without parsing prose.
func NewContextCommand(f *cmdutil.Factory) *cobra.Command {
	var (
		refresh   bool
		outputRaw string
	)
	cmd := &cobra.Command{
		Use:   "context",
		Short: "show the active profile's identity / role / accessible scope on the ControlHub side",
		Long: `Show the active profile's ControlHub-side identity: username,
KubeSphere global role ("platform-admin" / "platform-self-provisioner"
/ "platform-regular"), accessible workspaces, system namespaces, and
granted clusters.

Defaults to the locally cached snapshot (written on first ` + "`cluster context`" + `
or on a previous --refresh). Pass --refresh to force a fresh GET
against /capi/app/detail and update the cache; if the global role
changed since the last refresh you'll see a "globalrole changed:
X -> Y" notice.

The cache is for display only. Any "cluster ..." verb that needs to
know what you can see asks the server directly; we do not gate on the
locally cached role because that would both add attack surface (a
tampered cache could falsely "permit" something the server then
rejects) and silently get stale after role drift.

For BFL-side identity (OwnerRole on the desktop ingress) see
"olares-cli profile whoami".
`,
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, _ []string) error {
			return runContext(c.Context(), f, refresh, outputRaw)
		},
	}
	cmd.Flags().BoolVar(&refresh, "refresh", false, "force a fresh /capi/app/detail roundtrip and update the cached cluster context")
	cmd.Flags().StringVarP(&outputRaw, "output", "o", "table", "output format: table, json")
	return cmd
}

// runContext is the cobra-side glue: resolve the active profile +
// http.Client, parse the --output flag, then delegate to
// clusterctx.Run for the actual cache/server policy and rendering.
func runContext(ctx context.Context, f *cmdutil.Factory, refresh bool, outputRaw string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return fmt.Errorf("internal error: cluster context not wired with cmdutil.Factory")
	}

	format, err := clusterctx.ParseOutput(outputRaw)
	if err != nil {
		return err
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return err
	}
	cfg, err := cliconfig.LoadMultiProfileConfig()
	if err != nil {
		return err
	}

	client := clusterctx.NewHTTPClient(hc, rp.ControlHubURL, rp.OlaresID)
	return clusterctx.Run(ctx, client, cfg, rp.OlaresID, refresh, format, nil, os.Stdout)
}
