package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the top-level `search` command group. Each data
// source is a subcommand so flags are scoped to what the backend supports
// (e.g. sync has no --type).
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Drive files, Sync libraries, or installed applications",
		Long: `Run the same searches the Olares Desktop global search dialog uses.

Subcommands:
  drive   Full-content search of user Drive files (search3 index)
  sync    Search Seafile/Sync libraries
  app     Search installed applications by title

Examples:
  olares-cli search drive report
  olares-cli search drive "design doc" --type file_name
  olares-cli search sync notes --offset 20 -o json
  olares-cli search app wise
`,
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.AddCommand(newDriveCommand(f))
	cmd.AddCommand(newSyncCommand(f))
	cmd.AddCommand(newAppCommand(f))
	return cmd
}
