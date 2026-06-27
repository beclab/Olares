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

	doer, err := newDoer(ctx, f)
	if err != nil {
		return err
	}

	reqid := uuid.NewString()
	defer func() {
		_ = doEnvelope(ctx, doer, "POST", "/api/search/cancel",
			map[string]interface{}{"reqid": reqid}, nil)
	}()

	// init runs the search, caches the full (server-capped) result set, and
	// returns only the first initPageSize hits. It ignores offset/limit, so we
	// don't send them.
	initBody := map[string]interface{}{
		"reqid":   reqid,
		"keyword": keyword,
		"type":    searchType,
		"app":     appFilesV2,
	}
	var initRows []json.RawMessage
	if err := doEnvelope(ctx, doer, "POST", "/api/search/init", initBody, &initRows); err != nil {
		return err
	}

	// Honor --offset/--limit client-side. If the requested window already lies
	// within what init returned -- or init returned a short final page (fewer
	// than initPageSize hits means the cache holds no more) -- serve it
	// directly. Otherwise page the exact window via /search/more, whose limit
	// must stay within the backend's 1-100 range; a past-the-end offset comes
	// back as codeNoMoreResults, which we treat as an empty result set.
	var window []json.RawMessage
	if needsMorePage(o.offset, o.limit, len(initRows)) {
		moreBody := map[string]interface{}{
			"reqid":  reqid,
			"offset": o.offset,
			"limit":  clampMoreLimit(o.limit),
		}
		if err := doEnvelopeAllowing(ctx, doer, "POST", "/api/search/more", moreBody, &window, codeNoMoreResults); err != nil {
			return err
		}
	} else {
		window = paginateRaw(initRows, o.offset, o.limit)
	}

	items, err := decodeResultRows(window)
	if err != nil {
		return err
	}
	return printSearchResults(format, items)
}
