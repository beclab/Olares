package search

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type syncOptions struct {
	pagingOptions
}

func newSyncCommand(f *cmdutil.Factory) *cobra.Command {
	o := &syncOptions{}
	cmd := &cobra.Command{
		Use:   "sync <keyword>",
		Short: "Search Seafile/Sync libraries",
		Long: `Search the user's Sync (Seafile) libraries via /api/search/sync.

Examples:
  olares-cli search sync notes
  olares-cli search sync invoice --limit 50
  olares-cli search sync report --offset 20 -o json
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runSyncSearch(c.Context(), f, keyword, o)
		},
	}
	cmd.SilenceUsage = true
	registerPagingFlags(cmd, &o.pagingOptions)
	return cmd
}

func runSyncSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *syncOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := o.validate()
	if err != nil {
		return err
	}

	doer, err := newDoer(ctx, f)
	if err != nil {
		return err
	}

	// The sync proxy (user-service -> files /api/search/sync_search/) only
	// forwards the query and ignores offset/limit, so paginate client-side to
	// keep --offset/--limit honest.
	body := map[string]interface{}{
		"query": keyword,
	}
	var rawRows []json.RawMessage
	if err := doEnvelope(ctx, doer, "POST", "/api/search/sync", body, &rawRows); err != nil {
		return err
	}
	rawRows = paginateRaw(rawRows, o.offset, o.limit)

	items, err := decodeResultRows(rawRows)
	if err != nil {
		return err
	}
	return printSearchResults(format, items)
}
