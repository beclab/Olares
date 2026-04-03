package app

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCmdAppStop() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "stop {app-name}",
		Short: "Stop a running app",
		Long: `Stop a running application (suspend it).

For C/S architecture apps, use --cascade to stop all sub-charts.

Examples:
  olares-cli app stop myapp
  olares-cli app stop myapp --cascade`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(opts, args[0])
		},
	}
	opts.addConnectionFlags(cmd)
	opts.addOutputFlags(cmd)
	opts.addCascadeFlag(cmd)
	return cmd
}

func runStop(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	opts.info("Stopping '%s' for user '%s'...", appName, mc.user)

	ctx := context.Background()
	resp, err := mc.StopApp(ctx, appName, opts.Cascade)
	if err != nil {
		return opts.failOp("stop", appName, err)
	}

	result := newOperationResult(mc, "stop", appName, "", "", "stop requested", resp)
	return finishOperation(opts, mc, result)
}
