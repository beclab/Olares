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

// NewBackupCommand returns the `settings backup` parent.
func NewBackupCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup plans + snapshots + repository password (Settings -> Backup)",
		Long: `Manage backup plans (BFL backup-server, /apis/backup/v1/plans/backup/...)
and the repository password (user-service, /api/backup/password/:name).

Subcommands will be added in Phase 6:
  Phase 1: plans list (read-only sanity check across the BFL prefix)
  Phase 6: plans list / get / create / update / delete,
           snapshots list / get / delete,
           password get / set
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
