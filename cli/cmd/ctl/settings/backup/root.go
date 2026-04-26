// Package backup implements the `olares-cli settings backup` subtree
// (Settings -> Backup). Unlike most areas, backup rides a different ingress
// path prefix on the same desktop origin: `/apis/backup/v1/plans/backup/...`
// served by BFL's backup-server. Only the password endpoint goes through
// user-service at `/api/backup/password/:name` (backup.new.controller.ts).
//
// Phase 6 lands the read + write verbs once the simpler areas are stable —
// backup is intentionally last because of its deeper workflows and distinct
// ingress prefix, not because of any technical blocker. See plan.md's
// "Phase 6 — backup + restore" for the porting plan.
package backup

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewBackupCommand returns the `settings backup` parent. Phase 1
// ships the read-only verbs that exercise the BFL backup-server
// ingress prefix end-to-end; Phase 6 lands the write verbs (CRUD
// + pause/resume + password set).
func NewBackupCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup plans + snapshots + repository password (Settings -> Backup)",
		Long: `Manage backup plans (BFL backup-server, /apis/backup/v1/plans/backup/...)
and the repository password (user-service, /api/backup/password/:name).

Subcommands:
  plans list                                              (Phase 1)
  plans delete <id>  | pause <id> | resume <id>           (Phase 6)
  snapshots list   <backup-id>                            (Phase 1)
  snapshots run    <backup-id>                            (Phase 6)
  snapshots cancel <backup-id> <snapshot-id>              (Phase 6)
  password set     <name>                                 (Phase 6)

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
