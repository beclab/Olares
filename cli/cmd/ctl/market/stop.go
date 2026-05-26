package market

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketStop(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "stop {app-name}",
		Short: "Stop (suspend) a running app (POST /apps/stop)",
		Long: `Stop a running application (suspend it). Source is implicit
— stop acts on whichever per-user state row matches the app name
(no -s exposed).

For C/S (client/server) v2 multi-chart apps, --cascade controls whether
the shared sub-charts are stopped alongside the user's own chart (the
JSON payload field is "all"). Default behavior mirrors the Market SPA's
csAppStop() dialog:

  - --cascade NOT passed: auto-decided. When the cluster has a single
    user AND the target app is a v2 multi-chart bundle (isCSV2 — see
    SPA's constants.ts), default to --cascade=true; otherwise default
    to false. A short reason is printed on stderr when the
    auto-decision flips the default to true. Probe errors (user count
    or app info) soft-fail to --cascade=false.
  - --cascade or --cascade=true: force enabled.
  - --cascade=false: force disabled (the canonical override for the
    single-user CS auto-default; matches the SPA where the user
    unchecks the cascade checkbox in the multi-user dialog).

--watch blocks until the row settles at 'stopped' (or one of the
stopFailed / stoppingCanceled / stoppingCancelFailed failure states).
The watcher is idempotent: 'stop' against an already-stopped row
returns immediately with success ({state=stopped, opType=""}), rather
than hanging until --watch-timeout fires.

Examples:
  olares-cli market stop firefox                                # fire-and-forget; returns once backend accepts
  olares-cli market stop firefox --cascade=false                # force no cascade on a CS app
  olares-cli market stop firefox --cascade                      # force cascade (also stop dependents)
  olares-cli market stop firefox --watch                        # block until row reports stopped
  olares-cli market stop firefox --watch -o json | jq -r '.finalState'
  olares-cli market stop firefox --watch --watch-interval 1s --watch-timeout 2m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(opts, cmd, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runStop(opts *MarketOptions, cmd *cobra.Command, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	ctx := context.Background()
	cascade := opts.Cascade
	cascadeExplicit := cmd != nil && cmd.Flags().Changed("cascade")
	if !cascadeExplicit {
		// Auto-cascade only ever flips OFF → ON (never the other way),
		// matching the SPA's "default all:true for single-user CS apps,
		// otherwise default all:false" rule in csAppStop() — same probe
		// the uninstall path uses (single-user + isCSV2). On any
		// detection error we soft-fail to the original false default;
		// the user can always force the decision with --cascade=true.
		if auto, why := shouldAutoCascade(ctx, opts, mc, appName); auto {
			cascade = true
			opts.info("--cascade auto-enabled: %s (pass --cascade=false to override)", why)
		}
	}

	opts.info("Stopping '%s' for user '%s'...", appName, mc.olaresID)
	if cascade {
		opts.info("  --cascade: will stop all sub-charts")
	}

	resp, err := mc.StopApp(ctx, appName, cascade)
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	result := newOperationResult(mc, "stop", appName, "", "", "stop requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchStop, appName, opts.Source))
}
