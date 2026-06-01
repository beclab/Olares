package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUDetailFullCommand exposes the legacy `gpu detail
// <uuid>` cobra surface. Hidden + deprecated since the SPA-aligned
// refactor — `gpu graphics <uuid>` is the new canonical command.
// Kept functional for back-compat; emits the same
// `dashboard.overview.gpu.detail.full` envelope.
func newOverviewGPUDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "detail <uuid>",
		Short:         "Per-GPU detail page (info + gauges + trends; SPA Overview2/GPU/GPUsDetails)",
		Hidden:        true,
		Deprecated:    "use 'olares-cli dashboard overview gpu graphics <uuid>' instead",
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
