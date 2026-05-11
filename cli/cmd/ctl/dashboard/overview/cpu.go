package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewCPUCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:           "cpu",
		Short:         "Per-node CPU details (model / freq / cores / utilisation breakdown / temp / load avg)",
		Example:       `  olares-cli dashboard overview cpu -o json`,
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
			return pkgoverview.RunCPU(c.Context(), cli, common)
		},
	}
}
