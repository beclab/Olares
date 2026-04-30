// Package backup implements the `olares-cli settings backup` subtree
// (Settings -> Backup). Unlike most areas, backup rides a different ingress
// path prefix on the same desktop origin: `/apis/backup/v1/plans/backup/...`
// served by BFL's backup-server. Only the password endpoint goes through
// user-service at `/api/backup/password/:name` (backup.new.controller.ts).
package backup

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewBackupCommand returns the `settings backup` parent: plan + snapshot
// management against BFL's backup-server, plus the repository password
// helper served by user-service.
func NewBackupCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup plans + snapshots + repository password (Settings -> Backup)",
		Long: `Manage backup plans (BFL backup-server, /apis/backup/v1/plans/backup/...)
and the repository password (user-service, /api/backup/password/:name).

Subcommands:
  plans list
  plans delete <id>  | pause <id> | resume <id>
  snapshots list   <backup-id>
  snapshots run    <backup-id>
  snapshots cancel <backup-id> <snapshot-id>
  password set     <name>

Out of scope until a richer flag/file UX exists:
  plans create / update    (full BackupPolicy + LocationConfig)
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewPlansCommand(f))
	cmd.AddCommand(NewSnapshotsCommand(f))
	cmd.AddCommand(NewPasswordCommand(f))
	return cmd
}
