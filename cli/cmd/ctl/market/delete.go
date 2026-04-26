package market

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func NewCmdMarketDelete(f *cmdutil.Factory) *cobra.Command {
	opts := newMarketOptions(f)
	cmd := &cobra.Command{
		Use:     "delete {app-name}",
		Aliases: []string{"del"},
		Short:   "Delete a local app chart from the market source",
		Long: `Remove an app chart that was uploaded to a local source.
This does not uninstall the app if it is running.

Examples:
  olares-cli market delete myapp
  olares-cli market delete myapp --version 1.0.0`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}
	opts.addSourceFlag(cmd, "local source id to delete the chart from (auto-detected when omitted)")
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	return cmd
}

func runDelete(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("delete", appName, err)
	}

	if s := strings.TrimSpace(opts.Source); s != "" {
		if err := validateLocalSource(s); err != nil {
			return opts.failOp("delete", appName, err)
		}
	}

	source := resolveLocalSource(opts)
	if strings.TrimSpace(opts.Source) == "" {
		opts.info("Using source: %s", source)
	}

	version := strings.TrimSpace(opts.Version)
	if version != "" {
		if err := validateVersion(version); err != nil {
			return opts.failOp("delete", appName, err)
		}
	} else {
		v, err := resolveVersionInSource(mc, appName, source)
		if err != nil {
			return opts.failOp("delete", appName, fmt.Errorf("cannot determine version in source '%s': %w (use --version to specify)", source, err))
		}
		version = v
		opts.info("Using version: %s", version)
	}

	opts.info("Deleting chart '%s' version '%s' from source '%s'...", appName, version, source)

	ctx := context.Background()
	if _, err := mc.DeleteLocalApp(ctx, appName, version, source); err != nil {
		return opts.failOp("delete", appName, err)
	}

	result := OperationResult{
		App:       appName,
		Operation: "delete",
		Status:    "success",
		Message:   fmt.Sprintf("version %s deleted from source '%s'", version, source),
		Source:    source,
		Version:   version,
	}
	if !opts.Quiet {
		opts.printResult(result)
	}
	return nil
}
