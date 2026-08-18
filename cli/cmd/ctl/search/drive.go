package search

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type driveOptions struct {
	pagingOptions
	searchType string
	watch      bool
}

func newDriveCommand(f *cmdutil.Factory) *cobra.Command {
	o := &driveOptions{}
	cmd := &cobra.Command{
		Use:   "drive <keyword>",
		Short: "Search Drive, Google Drive, and Dropbox files",
		Long: `Search indexed files from Drive and connected cloud drives.

On Olares 1.12.7 and newer, one asynchronous search covers files_v2,
google_drive, and dropbox. Olares 1.12.6 and older keep using the legacy
/api/search/init + /more + /cancel API for local Drive files only.

With --watch, Olares 1.12.7+ prints each result as soon as its asynchronous
batch arrives. JSON output is JSONL (one result object per line). Older
versions have no result stream and print the completed legacy result instead.

Examples:
  olares-cli search drive report
  olares-cli search drive report --watch
  olares-cli search drive invoice --type file_name --limit 50
  olares-cli search drive "design doc" --watch -o json
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runDriveSearch(c.Context(), f, keyword, o)
		},
	}
	cmd.SilenceUsage = true
	cmd.Flags().StringVarP(&o.searchType, "type", "t", searchTypeAggregate,
		"search mode: aggregate, file_name")
	cmd.Flags().BoolVarP(&o.watch, "watch", "w", false,
		"print results as asynchronous search batches arrive (Olares >= 1.12.7)")
	registerPagingFlags(cmd, &o.pagingOptions)
	return cmd
}

func runDriveSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *driveOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	searchType, err := parseSearchType(o.searchType)
	if err != nil {
		return err
	}
	format, err := o.validate()
	if err != nil {
		return err
	}

	var watcher *watchResultPrinter
	var onHit func(asyncIndexedHit) error
	if o.watch {
		watcher = newWatchResultPrinter(os.Stdout, format, o.offset, o.limit)
		onHit = watcher.emit
	}

	items, async, err := runVersionedFileSearch(ctx, f, keyword, searchType, &o.pagingOptions, onHit)
	if err != nil {
		return err
	}
	if o.watch && async {
		return watcher.finish()
	}
	return printSearchResults(format, items)
}
