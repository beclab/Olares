package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

func newOverviewGPUTaskCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task <name> <pod-uid>",
		Short:         "Per-task detail (pod-uid from `kubectl get pods -n <ns> -o jsonpath='{.items[*].metadata.uid}'`)",
		Args:          cobra.ExactArgs(2),
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
			return pkggpu.RunTask(c.Context(), cli, common, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", "task share mode (passed to /v1/container?sharemode=)")
	return cmd
}
