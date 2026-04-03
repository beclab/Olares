package app

import (
	"context"

	"github.com/spf13/cobra"
)

func NewCmdAppResume() *cobra.Command {
	opts := &AppOptions{Output: "table"}
	cmd := &cobra.Command{
		Use:     "resume {app-name}",
		Aliases: []string{"start"},
		Short:   "Resume a stopped app",
		Long: `Resume a stopped (suspended) application.

Examples:
  olares-cli app resume myapp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runResume(opts, args[0])
		},
	}
	opts.addConnectionFlags(cmd)
	opts.addOutputFlags(cmd)
	return cmd
}

func runResume(opts *AppOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	opts.info("Resuming '%s' for user '%s'...", appName, mc.user)

	ctx := context.Background()
	resp, err := mc.ResumeApp(ctx, appName)
	if err != nil {
		return opts.failOp("resume", appName, err)
	}

	result := newOperationResult(mc, "resume", appName, "", "", "resume requested", resp)
	return finishOperation(opts, mc, result)
}
