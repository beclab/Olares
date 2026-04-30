package disk

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdisk "github.com/beclab/Olares/cli/pkg/dashboard/overview/disk"
)

func newOverviewDiskMainCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "main",
		Short:         "Per-physical-disk table (device / type / health / total / used / avail / temp / model / serial / firmware ...)",
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
			return pkgdisk.RunMain(c.Context(), cli, common)
		},
	}
}
