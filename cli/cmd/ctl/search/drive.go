package search

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type driveOptions struct {
	pagingOptions
	searchType string
}

func newDriveCommand(f *cmdutil.Factory) *cobra.Command {
	o := &driveOptions{}
	cmd := &cobra.Command{
		Use:   "drive <keyword>",
		Short: "Full-content search of user Drive files",
		Long: `Search the per-user search3 index for Drive files.

Drive search is session-based: the CLI bootstraps /api/search/init and, only
when the requested window runs past the first page, pages deeper via
/api/search/more using the same session id.

Note: a single search resolves at most ~50 hits server-side, so --limit is
effectively capped around 50.

Examples:
  olares-cli search drive report
  olares-cli search drive invoice --type file_name --limit 50
  olares-cli search drive "design doc" --offset 20 -o json
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

	items, err := runSessionSearch(ctx, f, keyword, appFilesV2, searchType, &o.pagingOptions)
	if err != nil {
		return err
	}
	return printSearchResults(format, items)
}
