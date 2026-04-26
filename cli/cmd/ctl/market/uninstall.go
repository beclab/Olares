package market

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketUninstall(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "uninstall {app-name}",
		Aliases: []string{"remove", "rm"},
		Short:   "Uninstall an app",
		Long: `Uninstall an application.

For C/S (client/server) architecture apps with multiple sub-charts,
use --cascade to uninstall both server and client parts.

Use --delete-data to also remove the app's persistent data.

Examples:
  olares-cli market uninstall myapp
  olares-cli market uninstall myapp --cascade
  olares-cli market uninstall myapp --cascade --delete-data`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	opts.addDeleteDataFlag(cmd)
	opts.addWatchFlags(cmd)
	return cmd
}

func runUninstall(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	opts.info("Uninstalling '%s' for user '%s'...", appName, mc.olaresID)
	if opts.Cascade {
		opts.info("  --cascade: will uninstall all sub-charts")
	}
	if opts.DeleteData {
		opts.info("  --delete-data: will delete persistent data")
	}

	ctx := context.Background()
	resp, err := mc.UninstallApp(ctx, appName, opts.Cascade, opts.DeleteData)
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	result := newOperationResult(mc, "uninstall", appName, "", "", "uninstall requested", resp)
	// Uninstall is unique: the row may simply disappear from /market/state
	// once the backend cleans it up, so the watch target opts in to the
	// "absent means success (provided we saw it earlier)" shortcut.
	return runWithWatch(opts, mc, result, newWatchTarget(watchUninstall, appName, opts.Source))
}
