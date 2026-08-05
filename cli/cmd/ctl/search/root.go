package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the top-level `search` command group. Each data
// source is a subcommand so flags are scoped to what the backend supports
// (e.g. sync has no --type; knowledge has no --type; gdrive/dropbox/knowledge
// require Olares >= 1.12.7).
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Drive, Sync, cloud drives, Wise, or installed applications",
		Long: `Run the same searches the Olares Desktop global search dialog uses.

Subcommands:
  drive       Full-content search of user Drive files (search3 index)
  sync        Search Seafile/Sync libraries
  gdrive      Search Google Drive files (Olares >= 1.12.7; aliases: google-drive, google)
  dropbox     Search Dropbox files (Olares >= 1.12.7)
  knowledge   Search Wise / Knowledge content (Olares >= 1.12.7; alias: wise)
  app         Search installed applications by title

Examples:
  olares-cli search drive report
  olares-cli search drive "design doc" --type file_name
  olares-cli search sync notes --offset 20 -o json
  olares-cli search gdrive invoice --limit 50
  olares-cli search dropbox report -o json
  olares-cli search knowledge design
  olares-cli search app wise
`,
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.AddCommand(newDriveCommand(f))
	cmd.AddCommand(newSyncCommand(f))
	cmd.AddCommand(newGDriveCommand(f))
	cmd.AddCommand(newDropboxCommand(f))
	cmd.AddCommand(newKnowledgeCommand(f))
	cmd.AddCommand(newAppCommand(f))
	return cmd
}
