package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewMemoryCommand(f *cmdutil.Factory) *cobra.Command {
	var mode string
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Per-node memory breakdown (--mode physical | swap)",
		Example: `  olares-cli dashboard overview memory --mode physical
  olares-cli dashboard overview memory --mode swap`,
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
			return pkgoverview.RunMemory(c.Context(), cli, common, mode)
		},
	}
	cmd.Flags().StringVar(&mode, "mode", "physical", "memory view: physical | swap")
	return cmd
}
