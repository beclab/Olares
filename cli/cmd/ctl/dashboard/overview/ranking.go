package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewRankingCommand(f *cmdutil.Factory) *cobra.Command {
	var sortDir string
	cmd := &cobra.Command{
		Use:   "ranking",
		Short: "Workload-grain (per-application) resource ranking (mirrors the SPA's UsageRanking widget)",
		Example: `  olares-cli dashboard overview ranking
  olares-cli dashboard overview ranking --sort asc --head 5`,
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
			return pkgoverview.RunRanking(c.Context(), cli, common, sortDir)
		},
	}
	cmd.Flags().StringVar(&sortDir, "sort", "desc", "sort direction (asc or desc)")
	return cmd
}
