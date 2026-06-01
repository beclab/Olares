package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUTaskDetailFullCommand exposes the legacy
// `gpu task-detail <name> <pod-uid>` cobra surface. Hidden +
// deprecated since the SPA-aligned refactor — `gpu tasks <name>`
// is the new canonical command (it auto-resolves pod-uid +
// sharemode, no kubectl call needed). Kept functional for
// back-compat scripts that already cached pod-uid; the explicit
// two-arg form remains the only path that takes a manual
// `--sharemode` override (the new `gpu tasks <name>` always
// auto-detects sharemode from HAMI's task list).
func newOverviewGPUTaskDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task-detail <name> <pod-uid>",
		Short:         "Per-task detail page (info + gauges + trends; SPA Overview2/GPU/TasksDetails)",
		Hidden:        true,
		Deprecated:    "use 'olares-cli dashboard overview gpu tasks <name>' instead (auto-resolves pod-uid; no kubectl needed)",
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
			return pkggpu.RunTaskDetail(c.Context(), cli, common, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", `task share mode ("0"=App exclusive, "1"=Memory slicing, "2"=Time slicing). When "2", allocation gauges are skipped to match the SPA.`)
	return cmd
}
