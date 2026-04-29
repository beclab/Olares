package disk

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	pkgdisk "github.com/beclab/Olares/cli/pkg/dashboard/overview/disk"
)

// NewDiskCommand assembles `dashboard overview disk` (root + main +
// partitions). cf is the shared *pkgdashboard.CommonFlags pointer
// stored in the area's `common` package var; cobra's persistent-
// flag inheritance from the dashboard root mutates the pointed-at
// struct before any leaf RunE fires.
func NewDiskCommand(f *cmdutil.Factory, cf *pkgdashboard.CommonFlags) *cobra.Command {
	common = cf
	cmd := &cobra.Command{
		Use:   "disk",
		Short: "Sections envelope: main = per-disk table; partitions = per-device partition tables",
		Example: `  olares-cli dashboard overview disk -o json
  olares-cli dashboard overview disk main
  olares-cli dashboard overview disk partitions sda`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) > 0 {
				return unknownSubcommandRunE(c, args)
			}
			if err := common.Validate(); err != nil {
				return err
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkgdisk.RunDefault(c.Context(), cli, common)
		},
	}
	cmd.AddCommand(newOverviewDiskMainCommand(f))
	cmd.AddCommand(newOverviewDiskPartitionsCommand(f))
	return cmd
}
