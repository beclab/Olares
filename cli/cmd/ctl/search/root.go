package search

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewSearchCommand returns the top-level `search` command group. Related file
// sources share the drive command; other search surfaces keep their own flags
// (e.g. sync and knowledge have no --type).
func NewSearchCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search files, Sync, Wise, or installed applications",
		Long: `Run the same searches the Olares Desktop global search dialog uses.

Subcommands:
  drive       Search Drive, Sync, Google Drive, and Dropbox (alias: files)
  sync        Search Seafile/Sync libraries only
  knowledge   Search Wise / Knowledge content (Olares >= 1.12.7; alias: wise)
  app         Search installed applications by title

On Olares 1.12.7+, drive and sync both use the asynchronous federated search
channel (search3). drive covers files_v2, google_drive, dropbox, and seafile
in one request; sync restricts the same channel to seafile. Older Olares
versions keep the legacy /api/search/init and /api/search/sync APIs.

Examples:
  olares-cli search drive report
  olares-cli search files report
  olares-cli search drive "design doc" --type file_name
  olares-cli search sync notes --offset 20 -o json
  olares-cli search knowledge design
  olares-cli search app wise
`,
	}
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	cmd.AddCommand(newDriveCommand(f))
	cmd.AddCommand(newSyncCommand(f))
	cmd.AddCommand(newKnowledgeCommand(f))
	cmd.AddCommand(newAppCommand(f))
	return cmd
}
