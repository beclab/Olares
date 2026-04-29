package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

func newOverviewGPUTasksCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "tasks",
		Short:         "List vGPU tasks (Task management tab)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkggpu.RunTasks(c.Context(), cli, common)
		},
	}
}
