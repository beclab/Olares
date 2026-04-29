package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

func newOverviewGPUDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "detail <uuid>",
		Short:         "Per-GPU detail page (info + gauges + trends; SPA Overview2/GPU/GPUsDetails)",
		Args:          cobra.ExactArgs(1),
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
			return pkggpu.RunDetail(c.Context(), cli, common, args[0])
		},
	}
}
