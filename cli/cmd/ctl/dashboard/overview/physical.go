package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewPhysicalCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "physical",
		Short:         "9-row cluster-level resource snapshot (CPU/Memory/Disk/Pods/Net + extras)",
		Example:       `  olares-cli dashboard overview physical -o json`,
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
			return pkgoverview.RunPhysical(c.Context(), cli, common)
		},
	}
}
