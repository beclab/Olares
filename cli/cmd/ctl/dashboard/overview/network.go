package overview

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkgoverview "github.com/beclab/Olares/cli/pkg/dashboard/overview"
)

func newOverviewNetworkCommand(f *cmdutil.Factory) *cobra.Command {
	var testConn bool
	cmd := &cobra.Command{
		Use:           "network",
		Short:         "Per-physical-NIC table from capi /system/ifs",
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
			return pkgoverview.RunNetwork(c.Context(), cli, common, testConn)
		},
	}
	cmd.Flags().BoolVar(&testConn, "test-connectivity", true, "ask the BFF to probe internet/IPv6 connectivity per interface")
	return cmd
}
