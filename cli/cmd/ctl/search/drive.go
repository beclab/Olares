package search

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
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

Drive search is session-based: the CLI bootstraps /api/search/init, then
pages deeper results via /api/search/more using the same session id.

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

	doer, err := newDoer(ctx, f)
	if err != nil {
		return err
	}

	reqid := uuid.NewString()
	defer func() {
		_ = doEnvelope(ctx, doer, "POST", "/api/search/cancel",
			map[string]interface{}{"reqid": reqid}, nil)
	}()

	initBody := map[string]interface{}{
		"reqid":   reqid,
		"keyword": keyword,
		"type":    searchType,
		"app":     appFilesV2,
		"offset":  0,
		"limit":   o.limit,
	}
	var rawRows []json.RawMessage
	if err := doEnvelope(ctx, doer, "POST", "/api/search/init", initBody, &rawRows); err != nil {
		return err
	}
	if o.offset > 0 {
		moreBody := map[string]interface{}{
			"reqid":  reqid,
			"offset": o.offset,
			"limit":  o.limit,
		}
		if err := doEnvelope(ctx, doer, "POST", "/api/search/more", moreBody, &rawRows); err != nil {
			return err
		}
	}

	items, err := decodeResultRows(rawRows)
	if err != nil {
		return err
	}
	return printSearchResults(format, items)
}
