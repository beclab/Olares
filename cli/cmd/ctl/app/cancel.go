package app

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCmdAppCancel() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:   "cancel {app-name}",
		Short: "Cancel the current in-progress app operation",
		Long: `Cancel the current in-progress operation for an app.

Examples:
  olares-cli app cancel myapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCancel(opts, args[0])
		},
	}
	opts.addConnectionFlags(cmd)
	opts.addOutputFlags(cmd)
	return cmd
}

func runCancel(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	opts.info("Canceling in-progress operation for '%s' (user '%s')...", appName, mc.user)

	ctx := context.Background()
	resp, err := mc.CancelOperation(ctx, appName)
	if err != nil {
		return opts.failOp("cancel", appName, err)
	}

	result := newOperationResult(mc, "cancel", appName, "", "", "cancel requested", resp)
	return finishOperation(opts, mc, result)
}
