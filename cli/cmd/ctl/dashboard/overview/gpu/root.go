package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// NewGPUCommand assembles `dashboard overview gpu` — the SPA's
// "GPU overview" page in CLI form. cf is the shared
// *pkgdashboard.CommonFlags pointer stored in the area's `common`
// package var; cobra's persistent-flag inheritance from the
// dashboard root mutates the pointed-at struct before any leaf
// RunE fires.
//
// Public surface (--help):
//
//	gpu graphics [uuid]    # SPA's Graphics management tab + GPUsDetails
//	gpu tasks    [name]    # SPA's Task management tab + TasksDetails
//
// The default RunE (no subverb) emits a sections envelope
// containing both `graphics` and `tasks` — same data the SPA's
// "GPU overview" page renders behind its two tabs (both lists are
// loaded eagerly when the page mounts; the user toggles between
// them with the tabs widget). This puts gpu in the parent-default
// = sections-envelope family alongside `overview`, `overview disk`
// and `overview fan`.
//
// Legacy/hidden surface (kept for back-compat agent scripts; runs
// with a Cobra deprecation hint):
//
//	gpu list / get <uuid> / detail <uuid>          → use `gpu graphics [uuid]`
//	gpu task <name> <pod-uid> / task-detail …      → use `gpu tasks [name]`
func NewGPUCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:   "gpu",
		Short: "GPU overview (SPA Overview2/GPU): graphics + tasks sections",
		Long: `GPU overview — mirrors the SPA's Overview2/GPU page (Graphics management + Task management tabs).

  olares-cli dashboard overview gpu                       # default → sections envelope (graphics + tasks)
  olares-cli dashboard overview gpu graphics              # graphics list only (Graphics management tab)
  olares-cli dashboard overview gpu graphics <uuid>       # GPU details page
  olares-cli dashboard overview gpu tasks                 # task list only (Task management tab)
  olares-cli dashboard overview gpu tasks <name>          # task details page (auto-resolves pod-uid)`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkggpu.RunDefault(c.Context(), cli, common)
		},
	}
	// Visible (SPA-aligned).
	cmd.AddCommand(newOverviewGPUGraphicsCommand(f))
	cmd.AddCommand(newOverviewGPUTasksCommand(f))
	// Hidden (legacy, deprecated). Still functional for existing
	// agent scripts; cobra emits a one-line deprecation hint when
	// invoked. Removing them later is a follow-up release-note item,
	// not part of the current refactor.
	cmd.AddCommand(newOverviewGPUListCommand(f))
	cmd.AddCommand(newOverviewGPUGetCommand(f))
	cmd.AddCommand(newOverviewGPUTaskCommand(f))
	cmd.AddCommand(newOverviewGPUDetailFullCommand(f))
	cmd.AddCommand(newOverviewGPUTaskDetailFullCommand(f))
	return cmd
}
