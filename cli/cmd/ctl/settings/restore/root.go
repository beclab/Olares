// Package restore implements the `olares-cli settings restore` subtree
// (Settings -> Restore). Sister area to `settings backup` — same BFL
// backup-server ingress prefix (`/apis/backup/v1/plans/restore/...`).
package restore

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRestoreCommand returns the `settings restore` parent: list, probe,
// create from snapshot or URL, and cancel restore plans on BFL's
// backup-server.
func NewRestoreCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore plans + URL pre-flight (Settings -> Restore)",
		Long: `Manage restore plans on the BFL backup-server
(/apis/backup/v1/plans/restore/...).

Subcommands:
  plans list
  plans check-url --backup-url URL [--password ... | --password-stdin]
  plans create-from-snapshot --snapshot-id SID --path PATH
  plans create-from-url --backup-url URL --path PATH [--dir DIR]
                        [--password ... | --password-stdin]
  plans cancel <id> [--yes]
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewPlansCommand(f))
	return cmd
}
