// Package search implements the `olares-cli settings search` subtree
// (Settings -> Search). Backed by user-service's search.controller.ts.
// Streaming proxies (chat-style answer streams) are intentionally out of
// scope; CLI exposes only the configuration + rebuild verbs.
package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the `settings search` parent: index status,
// full-rebuild trigger, and exclude pattern / indexed directory editors.
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search index settings (Settings -> Search)",
		Long: `Manage the local search index: status / rebuild / excludes / dirs.

Subcommands:
  status
  rebuild
  excludes list
  excludes add <pattern>...
  excludes rm  <pattern>...
  dirs list
  dirs add  <path>...
  dirs rm   <path>...

Streaming search proxies stay out of scope — they are interactive,
chat-style flows that don't fit a one-shot CLI verb.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewRebuildCommand(f))
	cmd.AddCommand(NewExcludesCommand(f))
	cmd.AddCommand(NewDirsCommand(f))
	return cmd
}
