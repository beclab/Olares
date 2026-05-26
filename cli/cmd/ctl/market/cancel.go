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
		Short: "Cancel the current in-progress app operation (install / upgrade / uninstall / ...)",
		Long: `Cancel the current in-progress operation for an app
(DELETE /apps/{name}/install). Source is implicit — cancel acts on
whichever per-user state row matches the app name (no -s flag exposed).

The cancel watcher is the widest in the market tree: any "row stopped
moving" state counts as success, including *Canceled, *Failed (the
underlying op died, cancel "won by default") and the stable resting
states running / stopped / uninstalled (cancel raced and lost, OR
rollback landed). Failure is ONLY surfaced for *CancelFailed (the
cancel request itself was rejected). This avoids the common hang where
a cancel races with downloadFailed / partial-rollback-to-stopped.

The terminal row carries the *underlying* op (install / upgrade / ...)
as its opType, not "cancel" — matchOpType is off, no race-tracking
gate applies.

Examples:
  olares-cli market cancel firefox                         # fire-and-forget; returns once backend accepts
  olares-cli market cancel firefox --watch                 # block until row settles (any terminal state except *CancelFailed)
  olares-cli market cancel firefox --watch -o json         # JSON; finalState surfaces where the row actually landed
  olares-cli market cancel firefox --watch -q              # silent; exit 0 unless *CancelFailed
  olares-cli market cancel firefox --watch --watch-interval 1s --watch-timeout 2m`,
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
