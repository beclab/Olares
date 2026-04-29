package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewPodsCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pods",
		Short:         "Per-node pod count snapshot (last/avg/max running)",
		Example:       `  olares-cli dashboard overview pods -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runPerNodeMetric(c.Context(), f, KindOverviewPods, podsMetricSet(), podsColumns(), podsDisplay)
		},
	}
	return cmd
}
