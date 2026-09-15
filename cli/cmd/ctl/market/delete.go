package market

import (
	"context"
	"encoding/json"
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

Delete always removes the WHOLE app from the bucket: every uploaded
version and its stored chart, whether or not --version is given. The
backend takes the app name and drops all of its artifacts; --version
only selects which version the request names, never how much is
deleted. There is no way to remove a single version today, so treat
this as "unpublish the app" rather than "unpublish a release".

Examples:
  olares-cli market delete myapp                    # remove the app and every uploaded version
  olares-cli market delete myapp --version 1.0.0    # same result: all versions go
  olares-cli market delete myapp -o json            # structured result
  olares-cli market delete myapp -q                 # silent; exit code only`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(opts, args[0])
		},
	}
	opts.addOutputFlags(cmd)
	opts.addVersionFlag(cmd)
	cmd.Flags().Lookup("version").Usage = "version to name in the request (default: latest available). Does NOT narrow the delete: every version of the app is removed either way"
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
		opts.info("note: --version names the request but does not narrow the delete; every uploaded version of '%s' will be removed", appName)
	} else {
		// --version does not narrow a delete, so a failure here must not
		// suggest it: the backend answers a delete of an app it does not hold
		// with success, which turns this correct failure into a false one.
		v, err := resolveVersionInSource(mc, appName, source, false)
		if err != nil {
			return opts.failOp("delete", appName, err)
		}
		version = v
		opts.info("Using version: %s", version)
	}

	opts.info("Deleting chart '%s' (all versions, request names '%s') from source '%s' for user '%s'...", appName, version, source, mc.olaresID)

	ctx := context.Background()
	resp, err := mc.DeleteLocalApp(ctx, appName, version, source)
	if err != nil {
		return opts.failOp("delete", appName, err)
	}

	// The backend answers a delete of an app it does not hold with HTTP 200
	// and success=true, reporting the work it actually did in
	// data.deleted_rows (market pkg/v2/api/catalog_local.go). Without
	// reading that, `delete some-typo` and `delete --version 9.9.9` both
	// printed "all versions deleted" having removed nothing.
	rows, known := deletedRowCount(resp)
	if known && rows == 0 {
		return opts.failOp("delete", appName, fmt.Errorf(
			"nothing was deleted: source '%s' holds no app named '%s' (run 'olares-cli market list -s %s' to see what is there)",
			source, appName, source))
	}

	message := fmt.Sprintf("all versions deleted from source '%s' (request named %s)", source, version)
	if known {
		message = fmt.Sprintf("%s; %d row(s) removed", message, rows)
	}
	result := OperationResult{
		App:       appName,
		Operation: "delete",
		Status:    "success",
		Message:   message,
		Source:    source,
		Version:   version,
	}
	if !opts.Quiet {
		opts.printResult(result)
	}
	return nil
}

// deletedRowCount pulls data.deleted_rows out of a /local-apps/delete
// response. `known` is false when the field is absent or not a number —
// an older backend, or a shape change — in which case the caller must not
// treat the delete as a no-op on the strength of a missing field.
func deletedRowCount(resp *APIResponse) (count int, known bool) {
	if resp == nil || len(resp.Data) == 0 {
		return 0, false
	}
	var payload struct {
		DeletedRows *float64 `json:"deleted_rows"`
	}
	if err := json.Unmarshal(resp.Data, &payload); err != nil || payload.DeletedRows == nil {
		return 0, false
	}
	return int(*payload.DeletedRows), true
}
