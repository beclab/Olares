package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

func newOverviewGPUTaskDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task-detail <name> <pod-uid>",
		Short:         "Per-task detail page (info + gauges + trends; SPA Overview2/GPU/TasksDetails)",
		Args:          cobra.ExactArgs(2),
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
			return pkggpu.RunTaskDetail(c.Context(), cli, common, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", `task share mode ("0"=App exclusive, "1"=Memory slicing, "2"=Time slicing). When "2", allocation gauges are skipped to match the SPA.`)
	return cmd
}
