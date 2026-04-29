package overview

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewMemoryCommand(f *cmdutil.Factory) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Per-node memory breakdown (--mode physical | swap)",
		Example: `  olares-cli dashboard overview memory --mode physical
  olares-cli dashboard overview memory --mode swap`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			switch mode {
			case "", "physical":
				return runPerNodeMetric(c.Context(), f, KindOverviewMemory, memoryPhysicalMetricSet(), memoryPhysicalColumns(), memoryPhysicalDisplay)
			case "swap":
				return runPerNodeMetric(c.Context(), f, KindOverviewMemory, memorySwapMetricSet(), memorySwapColumns(), memorySwapDisplay)
			default:
				return fmt.Errorf("--mode: %q must be physical or swap", mode)
			}
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "physical", "memory view: physical | swap")
	return cmd
}
