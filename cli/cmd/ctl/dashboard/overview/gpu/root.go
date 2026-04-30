package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// NewGPUCommand assembles `dashboard overview gpu` (root + 6
// leaves). cf is the shared *pkgdashboard.CommonFlags pointer
// stored in the area's `common` package var; cobra's persistent-
// flag inheritance from the dashboard root mutates the pointed-at
// struct before any leaf RunE fires. The default RunE (no
// subverb) forwards to `list`, mirroring the SPA's "GPU view
// opens to the device list" behaviour.
func NewGPUCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:           "gpu",
		Short:         "vGPU views: list / tasks / get / task / detail / task-detail",
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
			return pkggpu.RunList(c.Context(), cli, common)
		},
	}
	cmd.AddCommand(newOverviewGPUListCommand(f))
	cmd.AddCommand(newOverviewGPUTasksCommand(f))
	cmd.AddCommand(newOverviewGPUGetCommand(f))
	cmd.AddCommand(newOverviewGPUTaskCommand(f))
	cmd.AddCommand(newOverviewGPUDetailFullCommand(f))
	cmd.AddCommand(newOverviewGPUTaskDetailFullCommand(f))
	return cmd
}
