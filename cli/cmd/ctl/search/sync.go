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

// syncSearchSources narrows the federated search to the one source `search
// drive` covers for Sync, so the two commands can never disagree about which
// libraries are searched.
var syncSearchSources = []string{appSeafile}

func newSyncCommand(f *cmdutil.Factory) *cobra.Command {
	o := &syncOptions{}
	cmd := &cobra.Command{
		Use:   "sync <keyword>",
		Short: "Search Seafile/Sync libraries",
		Long: `Search the user's Sync (Seafile) libraries.

On Olares 1.12.7 and newer this uses the same asynchronous federated
search channel as search drive, restricted to the seafile source.
Olares 1.12.6 and older keep using /api/search/sync.

--offset/--limit are applied client-side on both paths.

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

	useAsync, err := f.OlaresBackendAtLeast(ctx, asyncSearchMinOlaresVersion)
	if err != nil {
		return err
	}
	if useAsync {
		items, err := runAsyncSearch(ctx, f, keyword, syncSearchSources, searchTypeAggregate, &o.pagingOptions, nil)
		if err != nil {
			return err
		}
		return printSearchResults(format, items)
	}

	return runLegacySyncSearch(ctx, f, keyword, o, format)
}

func runLegacySyncSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *syncOptions, format Format) error {
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
