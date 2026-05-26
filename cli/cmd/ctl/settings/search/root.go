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
// full-rebuild trigger, and indexed directory editor.
//
// The `excludes` subtree (NewExcludesCommand in excludes.go) is
// intentionally NOT wired here anymore — the Go implementation is
// preserved verbatim so the verb can be restored in one line by
// re-adding `cmd.AddCommand(NewExcludesCommand(f))` below. See
// excludes.go's NewExcludesCommand doc comment for the rationale.
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search index settings (Settings -> Search)",
		Long: `Manage the local search index: status / rebuild / dirs.

Subcommands:
  status
  rebuild
  dirs list

Streaming search proxies stay out of scope — they are interactive,
chat-style flows that don't fit a one-shot CLI verb.
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewStatusCommand(f))
	cmd.AddCommand(NewRebuildCommand(f))
	cmd.AddCommand(NewDirsCommand(f))
	return cmd
}
