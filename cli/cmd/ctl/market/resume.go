package market

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/olaresclient"
)

func NewCmdMarketResume(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "resume {app-name}",
		Aliases: []string{"start"},
		Short:   "Resume a stopped app (sends POST /apps/resume)",
		Long: `Resume a stopped (suspended) application.

Source is implicit: resume acts on whichever per-user state row matches
the app name, regardless of source (no -s flag exposed).

--watch blocks until the row settles at 'running' (or one of the
resumeFailed / resumingCanceled / resumingCancelFailed failure states).
The watcher is idempotent: 'resume' against an already-running row
returns immediately with success ({state=running, opType=""}), rather
than hanging until --watch-timeout fires.

Examples:
  olares-cli market resume firefox                         # fire-and-forget; returns once backend accepts
  olares-cli market resume firefox --watch                 # block until row reports running
  olares-cli market resume firefox --watch -o json         # JSON OperationResult with finalState/finalOpType
  olares-cli market resume firefox --watch -q              # silent; exit code only
  olares-cli market resume firefox --watch --watch-interval 1s --watch-timeout 2m`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runResume(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	opts.info("Resuming '%s' for user '%s'...", appName, mc.olaresID)

	ctx := context.Background()

	// 1.12.6 requires source in the resume body; resolve it from the
	// installed app's state row (resume exposes no -s).
	source, err := resolveInstalledSource(ctx, opts, mc, appName)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	var data json.RawMessage
	if err := opts.factory.WithOlaresClient(ctx, func(c olaresclient.OlaresClient) error {
		d, e := c.ResumeApp(ctx, mc, appName, source)
		data = d
		return e
	}); err != nil {
		return opts.failOp("resume", appName, err)
	}

	result := newOperationResult(mc, "resume", appName, "", "", "resume requested", &APIResponse{Success: true, Data: data})
	return runWithWatch(opts, mc, result, newWatchTarget(watchResume, appName, opts.Source))
}
