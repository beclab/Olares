package app

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCmdAppUninstall() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "uninstall {app-name}",
		Aliases: []string{"remove", "rm"},
		Short:   "Uninstall an app",
		Long: `Uninstall an application.

For C/S (client/server) architecture apps with multiple sub-charts,
use --cascade to uninstall both server and client parts.

Use --delete-data to also remove the app's persistent data.

Examples:
  olares-cli app uninstall myapp
  olares-cli app uninstall myapp --cascade
  olares-cli app uninstall myapp --cascade --delete-data`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(opts, args[0])
		},
	}
	opts.addConnectionFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	opts.addDeleteDataFlag(cmd)
	return cmd
}

func runUninstall(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("uninstall", appName, err)
	}

	opts.info("Uninstalling '%s' for user '%s'...", appName, mc.user)
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
	return finishOperation(opts, mc, result)
}
