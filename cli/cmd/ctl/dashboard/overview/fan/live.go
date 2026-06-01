package fan

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgfan "github.com/beclab/Olares/cli/pkg/dashboard/overview/fan"
)

func newOverviewFanLiveCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "live",
		Short:         "1-row real-time fan / temperature / power snapshot (Olares One)",
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
			return pkgfan.RunLive(c.Context(), cli, common)
		},
	}
}
