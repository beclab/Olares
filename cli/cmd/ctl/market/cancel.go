package market

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketCancel(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:   "cancel {app-name}",
		Short: "Cancel the current in-progress app operation",
		Long: `Cancel the current in-progress operation for an app.

Examples:
  olares-cli market cancel myapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runCancel(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	opts.info("Canceling in-progress operation for '%s' (user '%s')...", appName, mc.olaresID)

	ctx := context.Background()
	resp, err := mc.CancelOperation(ctx, appName)
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	result := newOperationResult(mc, "cancel", appName, "", "", "cancel requested", resp)
	// Cancel's terminal row carries the *underlying* OpType (install /
	// upgrade / ...), not "cancel", so the watch target opts out of
	// strict OpType matching via matchOpType=false in newWatchTarget.
	return runWithWatch(opts, mc, result, newWatchTarget(watchCancel, appName, opts.Source))
}
