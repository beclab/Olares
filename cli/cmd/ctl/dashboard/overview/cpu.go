package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewCPUCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cpu",
		Short:         "Per-node CPU details (model / freq / cores / utilisation breakdown / temp / load avg)",
		Example:       `  olares-cli dashboard overview cpu -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runPerNodeMetric(c.Context(), f, KindOverviewCPU, cpuMetricSet(), cpuColumns(), cpuDisplay)
		},
	}
	return cmd
}
