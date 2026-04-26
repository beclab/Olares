// Package search implements the `olares-cli settings search` subtree
// (Settings -> Search). Backed by user-service's search.controller.ts.
// Streaming proxies (chat-style answer streams) are intentionally out of
// scope; CLI exposes only the configuration + rebuild verbs.
package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the `settings search` parent. Phase 1 ships
// the read-only verbs (status / excludes list / dirs list); Phase 2
// will add rebuild + add/rm verbs for excludes/dirs.
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search index settings (Settings -> Search)",
		Long: `Manage the local search index: status / rebuild / excludes / dirs.

Subcommands:
  status                                                  (Phase 1)
  excludes list                                           (Phase 1)
  dirs list                                               (Phase 1)

Subcommands landing in Phase 2:
  rebuild
  excludes add <pattern>..., excludes rm <pattern>...
  dirs add <path>..., dirs rm <path>...

Streaming search proxies stay out of scope — they are interactive,
chat-style flows that don't fit a one-shot CLI verb.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewExcludesCommand(f))
	cmd.AddCommand(NewDirsCommand(f))
	return cmd
}
