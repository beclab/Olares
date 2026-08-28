package search

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

type knowledgeOptions struct {
	pagingOptions
}

func newKnowledgeCommand(f *cmdutil.Factory) *cobra.Command {
	o := &knowledgeOptions{}
	cmd := &cobra.Command{
		Use:     "knowledge <keyword>",
		Aliases: []string{"wise"},
		Short:   "Search Wise / Knowledge content (Olares >= 1.12.7)",
		Long: fmt.Sprintf(`Search Wise (Knowledge) content via the search3 index.

Requires Olares >= %s. Uses the same session-based protocol as
`+"`search drive`"+` (/api/search/init + /more + /cancel), with app=knowledge.

Search mode is fixed to aggregate (full-content); there is no --type flag.
Result locations are typically http(s) URLs into the Wise app.

Examples:
  olares-cli search knowledge report
  olares-cli search wise invoice --limit 50
  olares-cli search knowledge "design doc" --offset 20 -o json
`, searchSessionAppMinOlaresVersion),
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			keyword, err := parseKeyword(args)
			if err != nil {
				return err
			}
			return runKnowledgeSearch(c.Context(), f, keyword, o)
		},
	}
	cmd.SilenceUsage = true
	registerPagingFlags(cmd, &o.pagingOptions, initPageSize)
	return cmd
}

func runKnowledgeSearch(ctx context.Context, f *cmdutil.Factory, keyword string, o *knowledgeOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := requireSessionAppBackendVersion(ctx, f, "search knowledge", "search3 app=knowledge index"); err != nil {
		return err
	}
	format, err := o.validate()
	if err != nil {
		return err
	}

	items, err := runSessionSearch(ctx, f, keyword, appKnowledge, searchTypeAggregate, &o.pagingOptions)
	if err != nil {
		return err
	}
	return printSearchResults(format, searchPage{items: items, offset: o.offset, windowed: o.windowed()})
}
