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
		Short: "Stop a running app",
		Long: `Stop a running application (suspend it).

For C/S architecture apps, use --cascade to stop all sub-charts.

Examples:
  olares-cli market stop myapp
  olares-cli market stop myapp --cascade`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runStop(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	opts.info("Stopping '%s' for user '%s'...", appName, mc.olaresID)

	ctx := context.Background()
	resp, err := mc.StopApp(ctx, appName, opts.Cascade)
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	result := newOperationResult(mc, "stop", appName, "", "", "stop requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchStop, appName, opts.Source))
}
