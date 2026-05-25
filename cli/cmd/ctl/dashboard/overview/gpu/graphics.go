package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUGraphicsCommand wires the SPA-aligned `gpu
// graphics [uuid]` parent command. Mirrors the SPA's "Graphics
// management" tab — bare invocation lists every device (the SPA's
// table view), passing a uuid drills into the per-GPU detail page
// (info + 6 gauges + 4 trends, the SPA's GPUsDetails route).
//
// Internally dispatches to RunList (no positional arg) or
// RunDetail (one positional arg). Implementation note: kept on the
// cmd side because SKILL forbids putting fan-out logic in cmd, and
// dispatching by arg count is neither fan-out nor envelope work —
// it's argument routing, the canonical responsibility of cmd.
func newOverviewGPUGraphicsCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "graphics [uuid]",
		Short: "Graphics management — list GPUs (no arg) or detail page (uuid; SPA Overview2/GPU/GPUsDetails)",
		Long: `Graphics management — SPA's "Graphics management" tab.

  olares-cli dashboard overview gpu graphics              # list every GPU (HAMI /v1/gpus)
  olares-cli dashboard overview gpu graphics <uuid>       # info + gauges + trends for one GPU
                                                          # (HAMI /v1/gpu?uid=<uuid> + 4 instant +
                                                          #  4 range PromQL queries)

The single command replaces the older split of ` + "`list` / `detail`" + `.`,
		Example: `  olares-cli dashboard overview gpu graphics
  olares-cli dashboard overview gpu graphics GPU-e5d26177-beec-64c1-2681-56cc973d9910 -o json`,
		Args:          cobra.MaximumNArgs(1),
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
			if len(args) == 0 {
				return pkggpu.RunList(c.Context(), cli, common)
			}
			return pkggpu.RunDetail(c.Context(), cli, common, args[0])
		},
	}
}
