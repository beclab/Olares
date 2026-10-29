package search

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type marketOptions struct {
	pagingOptions
	searchType string
}

func newMarketCommand(f *cmdutil.Factory) *cobra.Command {
	o := &marketOptions{}
	cmd := &cobra.Command{
		Use:   "market <keyword>",
		Short: "Search the Market app catalog",
		Long: fmt.Sprintf(`Search the Market app catalog by app name, title, and description.

Requires Olares >= %s and a search3 build that serves the market source.
Uses the same session-based protocol as `+"`search knowledge`"+`
(/api/search/init + /more + /cancel), with app=market. Market answers as
the signed-in user: only sources, NSFW settings, and delisted apps that
user can see are considered.

Unlike `+"`search app`"+`, which matches installed apps and returns their
entrance URLs, this searches the catalog, installed or not. Each result's
location is the Market detail route /app/<source_id>/<app_name>; with -o json
the meta field carries source_id, app_id, icon, version, categories, and
whether this machine's hardware can install the app.

Examples:
  olares-cli search market music
  olares-cli search market "photo editor" --type file_name
  olares-cli search market ollama --limit 50 -o json
`, searchSessionAppMinOlaresVersion),
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runMarketSearch(c.Context(), f, keyword, o)
		},
	}
	cmd.SilenceUsage = true
	cmd.Flags().StringVarP(&o.searchType, "type", "t", searchTypeAggregate,
		"search mode: aggregate, file_name")
	registerPagingFlags(cmd, &o.pagingOptions, initPageSize)
	return cmd
}

func runMarketSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *marketOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireSessionAppBackendVersion(ctx, f, "search market", "search3 app=market source"); err != nil {
		return err
	}
	format, err := o.validate()
	if err != nil {
		return err
	}
	searchType, err := parseSearchType(o.searchType)
	if err != nil {
		return err
	}

	paging := o.sessionPaging()
	items, err := runSessionSearch(ctx, f, keyword, appMarket, searchType, &paging)
	if err != nil {
		return err
	}
	return printSearchResults(format, searchPage{items: items, offset: paging.offset, windowed: paging.windowed()})
}
