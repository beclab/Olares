// Package restore implements the `olares-cli settings restore` subtree
// (Settings -> Restore). Sister area to `settings backup` — same BFL
// backup-server ingress prefix (`/apis/backup/v1/plans/restore/...`), same
// Phase 6 timing. See plan.md's "Phase 6 — backup + restore" section.
package restore

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewRestoreCommand returns the `settings restore` parent. Phase 1
// ships `plans list`; Phase 6 lands the write verbs (create from
// snapshot or URL, cancel, plus the URL pre-flight).
func NewRestoreCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore plans + URL pre-flight (Settings -> Restore)",
		Long: `Manage restore plans on the BFL backup-server
(/apis/backup/v1/plans/restore/...).

Subcommands:
  plans list                                              (Phase 1)

Subcommands landing in Phase 6:
  plans get / create-from-snapshot / create-from-url / cancel,
  plans check-url <url> --password <pw>
`,
	}
	cmd.SilenceUsage = true
	cmd.AddCommand(NewPlansCommand(f))
	return cmd
}
