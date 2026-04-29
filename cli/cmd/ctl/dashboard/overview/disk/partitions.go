package disk

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdisk "github.com/beclab/Olares/cli/pkg/dashboard/overview/disk"
)

func newOverviewDiskPartitionsCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "partitions <device>",
		Short:         "Partition-level table for one physical device",
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
			return pkgdisk.RunPartitions(c.Context(), cli, common, args[0])
		},
	}
}
