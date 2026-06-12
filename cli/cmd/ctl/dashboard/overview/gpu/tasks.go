package gpu

import (
	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	pkggpu "github.com/beclab/Olares/cli/pkg/dashboard/overview/gpu"
)

// newOverviewGPUTasksCommand wires the SPA-aligned `gpu tasks
// [ref]` parent command. Mirrors the SPA's "Task management" tab —
// bare invocation lists every running vGPU task, passing a `<ref>`
// drills into that task's detail page (info + 2 gauges + 2 trends,
// the SPA's TasksDetails route).
//
// `<ref>` accepts either the row's `name` (TASK column) or its
// `podUid` (POD_UID column) — both are surfaced in the bare `gpu`
// / `gpu tasks` listings, and copy-paste from either column should
// "just work". The pkg layer's RunTaskByRef reverse-resolves
// pod-uid + sharemode from the task list (same path the SPA takes
// when the user clicks "View details" on a TasksTable row).
//
// The legacy `gpu task-detail <name> <pod-uid>` two-arg form
// remains hidden + deprecated for agent scripts that already
// cached the pod-uid.
func newOverviewGPUTasksCommand(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "tasks [name|pod-uid]",
		Short: "Task management — list vGPU tasks (no arg) or detail page (name or pod-uid; SPA Overview2/GPU/TasksDetails)",
		Long: `Task management — SPA's "Task management" tab.

  olares-cli dashboard overview gpu tasks                 # list every vGPU task (HAMI /v1/containers)
  olares-cli dashboard overview gpu tasks <ref>           # info + gauges + trends for one task
                                                          # <ref> = TASK (name) or POD_UID column from
                                                          #  the listing; auto-resolves the missing half
                                                          #  + sharemode (no kubectl needed).

When two tasks share the same name, the CLI errors out with the candidate pod-uids;
re-run with one of those pod-uids (still ` + "`gpu tasks <pod-uid>`" + `) to disambiguate.`,
		Example: `  olares-cli dashboard overview gpu tasks
  olares-cli dashboard overview gpu tasks comfyai -o json
  olares-cli dashboard overview gpu tasks d2a8ea32-8e56-49f9-b876-21b2fa0c5a83`,
		Args:          cobra.MaximumNArgs(1),
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
			if len(args) == 0 {
				return pkggpu.RunTasks(c.Context(), cli, common)
			}
			return pkggpu.RunTaskByRef(c.Context(), cli, common, args[0])
		},
	}
}
