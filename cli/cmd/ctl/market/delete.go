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
		Short:   "Remove an uploaded helm chart from the SPA's Local Sources → Upload bucket",
		Long: `Remove an app chart that was uploaded to a local source.
This does NOT uninstall the app if it is running — use
'olares-cli market uninstall <app>' for that, then 'market delete'
to also remove the chart from local sources.

The chart is always removed from the SPA's "Local Sources → Upload"
bucket (internal id 'upload') — the same bucket 'market upload' writes
to. The CLI used to expose -s/--source here, but a delete that targets
a different bucket from where the upload landed never resolved
correctly. Pinning the source eliminates that mismatch.

If --version is omitted, every uploaded version of the chart in the
'upload' bucket is removed.

Examples:
  olares-cli market delete myapp                    # remove every uploaded version
  olares-cli market delete myapp --version 1.0.0    # one version only
  olares-cli market delete myapp -o json            # structured result
  olares-cli market delete myapp -q                 # silent; exit code only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	return cmd
}

func runDelete(opts *MarketOptions, appName string) error {
	mc, err := opts.prepare()
	if err != nil {
		return opts.failOp("delete", appName, err)
	}

	// Source is hard-coded to match what `market upload` writes to;
	// see chartUploadSource in common.go.
	source := chartUploadSource

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
