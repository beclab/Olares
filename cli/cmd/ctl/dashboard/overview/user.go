package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewUserCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "user [<username>]",
		Short: "User-grain CPU / memory quota usage (mirrors the SPA's User Resources panel)",
		Example: `  olares-cli dashboard overview user
  olares-cli dashboard overview user alice    # admin only`,
		Args:          cobra.MaximumNArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			target := common.User
			if len(args) == 1 {
				target = args[0]
			}
			cli, err := prepareClient(c.Context(), f)
			if err != nil {
				return err
			}
			return pkgoverview.RunUser(c.Context(), cli, common, target)
		},
	}
}
