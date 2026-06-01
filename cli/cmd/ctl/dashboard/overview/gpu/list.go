package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUListCommand exposes the legacy `gpu list` cobra
// surface. Hidden + deprecated since the SPA-aligned refactor —
// `gpu graphics` is the new canonical command. Kept for existing
// agent scripts; emits the same envelope (`dashboard.overview.gpu.list`).
func newOverviewGPUListCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "list",
		Short:         "List discovered vGPUs (Graphics management tab; 404 = HAMI not installed)",
		Hidden:        true,
		Deprecated:    "use 'olares-cli dashboard overview gpu graphics' instead",
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
			return pkggpu.RunList(c.Context(), cli, common)
		},
	}
}
