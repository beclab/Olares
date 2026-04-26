// Package search implements the `olares-cli settings search` subtree
// (Settings -> Search). Backed by user-service's search.controller.ts.
// Streaming proxies (chat-style answer streams) are intentionally out of
// scope; CLI exposes only the configuration + rebuild verbs.
package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the `settings search` parent.
func NewSearchCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search index settings (Settings -> Search)",
		Long: `Manage the local search index: status / rebuild / excludes / dirs.

Subcommands will be added in subsequent phases:
  Phase 1: status
  Phase 2: rebuild, excludes (add|rm), dirs (add|rm)

Streaming search proxies stay out of scope — they are interactive,
chat-style flows that don't fit a one-shot CLI verb.
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
