// Package restore implements the `olares-cli settings restore` subtree
// (Settings -> Restore). Sister area to `settings backup` — same BFL
// backup-server ingress prefix (`/apis/backup/v1/plans/restore/...`), same
// Phase 6 timing. See plan.md's "Phase 6 — backup + restore" section.
package restore

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRestoreCommand returns the `settings restore` parent.
func NewRestoreCommand(_ *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore plans + URL pre-flight (Settings -> Restore)",
		Long: `Manage restore plans on the BFL backup-server
(/apis/backup/v1/plans/restore/...).

Subcommands will be added in Phase 6:
  Phase 1: plans list (read-only sanity check)
  Phase 6: plans list / get / create / delete, check-url <url>
`,
	}
	cmd.SilenceUsage = true
	return cmd
}
