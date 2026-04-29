package fan

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgfan "github.com/beclab/Olares/cli/pkg/dashboard/overview/fan"
)

func newOverviewFanCurveCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "curve",
		Short:         "10-row hardcoded fan-curve specification (RPM \u2194 temperature range)",
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
			return pkgfan.RunCurve(c.Context(), cli, common)
		},
	}
}
