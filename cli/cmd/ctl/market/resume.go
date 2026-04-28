package market

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketResume(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "resume {app-name}",
		Aliases: []string{"start"},
		Short:   "Resume a stopped app",
		Long: `Resume a stopped (suspended) application.

Examples:
  olares-cli market resume myapp`,
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
	resp, err := mc.ResumeApp(ctx, appName)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	result := newOperationResult(mc, "resume", appName, "", "", "resume requested", resp)
	return runWithWatch(opts, mc, result, newWatchTarget(watchResume, appName, opts.Source))
}
