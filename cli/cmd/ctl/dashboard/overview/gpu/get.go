package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUGetCommand exposes the legacy `gpu get <uuid>`
// cobra surface — a single flat HAMI /v1/gpu detail item with no
// gauge / trend fan-out. Hidden + deprecated: this verb has no SPA
// equivalent (the SPA's GPUsDetails page always renders gauges +
// trends), so new callers should switch to `gpu graphics <uuid>`
// for the SPA-aligned view. Kept functional for back-compat.
func newOverviewGPUGetCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "get <uuid>",
		Short:         "Per-GPU detail by UUID (HAMI raw passthrough; no gauges/trends)",
		Hidden:        true,
		Deprecated:    "use 'olares-cli dashboard overview gpu graphics <uuid>' for the SPA-aligned detail page (info + gauges + trends)",
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
			return pkggpu.RunGet(c.Context(), cli, common, args[0])
		},
	}
}
