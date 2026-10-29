// Package backup implements the `olares-cli settings backup` subtree
// (Settings -> Backup). Unlike most areas, backup rides a different ingress
// path prefix on the same desktop origin: `/apis/backup/v1/plans/backup/...`
// served by BFL's backup-server.
package backup

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewBackupCommand returns the `settings backup` parent: plan + snapshot
// management against BFL's backup-server.
func NewBackupCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup plans + snapshots (Settings -> Backup)",
		Long: `Manage backup plans and snapshots (BFL backup-server, /apis/backup/v1/plans/backup/...).

Subcommands:
  plans list
  snapshots list   <backup-id>

Out of scope until a richer flag/file UX exists:
  plans create / update    (full BackupPolicy + LocationConfig)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewPlansCommand(f))
	cmd.AddCommand(NewSnapshotsCommand(f))
	return cmd
}
