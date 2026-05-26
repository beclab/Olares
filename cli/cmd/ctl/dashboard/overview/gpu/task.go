package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUTaskCommand exposes the legacy `gpu task <name>
// <pod-uid>` cobra surface — a single flat HAMI /v1/container body
// with no gauge / trend fan-out. Hidden + deprecated: this verb
// has no SPA equivalent (the SPA's TasksDetails page always renders
// gauges + trends). New callers should switch to `gpu tasks <name>`
// for the SPA-aligned view (auto-resolves pod-uid). Kept functional
// for back-compat.
func newOverviewGPUTaskCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task <name> <pod-uid>",
		Short:         "Per-task detail (HAMI raw passthrough; no gauges/trends)",
		Hidden:        true,
		Deprecated:    "use 'olares-cli dashboard overview gpu tasks <name>' for the SPA-aligned detail page (info + gauges + trends)",
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
			return pkggpu.RunTask(c.Context(), cli, common, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", "task share mode (passed to /v1/container?sharemode=)")
	return cmd
}
