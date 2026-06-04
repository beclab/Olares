package market

import (
	"context"

	"github.com/spf13/cobra"

	v1_12_5 "github.com/beclab/Olares/cli/cmd/ctl/market/resume/v1_12_5"
	v1_12_6 "github.com/beclab/Olares/cli/cmd/ctl/market/resume/v1_12_6"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
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

	// Resolve the installed app's source: enforces the "only operate on an
	// installed app" guard (a bugfix applying to 1.12.5 and 1.12.6 alike)
	// and yields the source the 1.12.6 body needs (resume exposes no -s).
	// The resolved source also sharpens the --watch state-row match.
	source, err := resolveInstalledSource(ctx, opts, mc, appName)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	atLeast126, err := opts.factory.OlaresBackendAtLeast(ctx, "1.12.6")
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	// 1.12.6 moved the body to {app_name, source}; 1.12.5 keeps {appName}
	// and ignores source.
	method, path, body := v1_12_5.Resume(appName, source)
	if atLeast126 {
		method, path, body = v1_12_6.Resume(appName, source)
	}
	resp, err := mc.doRequest(ctx, method, path, body)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	result := newOperationResult(mc, "resume", appName, "", "", "resume requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchResume, appName, source))
}
