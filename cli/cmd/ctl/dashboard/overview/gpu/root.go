package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// ----------------------------------------------------------------------------
// overview gpu — list / tasks / get / task
// ----------------------------------------------------------------------------

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
			// Default action: forward to `list`.
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUList(c.Context(), f)
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
