package gpu

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func newOverviewGPUTasksCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "tasks",
		Short:         "List vGPU tasks (Task management tab)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTasks(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewGPUTasks(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, _ := gpuAdvisory(ctx, c)
	list, err := fetchTaskList(ctx, c, nil)
	env := Envelope{Kind: KindOverviewGPUTasks, Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUTasks, now); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if common.Output == OutputJSON {
				return WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if common.Output == OutputJSON {
			return WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
		return nil
	}
	// HAMI's task entries match the SPA's `TaskItem` interface. The
	// "core util / mem used" columns are arrays (one element per
	// allocated device) — SPA uses index 0 too. Raw envelope retains
	// the full array so multi-GPU tasks aren't silently truncated.
	for _, t := range list {
		raw := map[string]any{}
		for k, v := range t {
			raw[k] = v
		}
		shareModeFirst := firstAnyInArray(t["deviceShareModes"])
		coreUtilFirst := firstAnyInArray(t["devicesCoreUtilizedPercent"])
		memMiBFirst := firstAnyInArray(t["devicesMemUtilized"])
		disp := map[string]any{
			"task_name": fmt.Sprintf("%v", t["name"]),
			"status":    fmt.Sprintf("%v", t["status"]),
			"mode":      gpuModeLabel(shareModeFirst),
			"host_node": fmt.Sprintf("%v", t["nodeName"]),
			"core_util": percentDirect(toFloat(coreUtilFirst)),
			"mem_used":  gpuVRAMHuman(memMiBFirst),
			"pod_uid":   fmt.Sprintf("%v", t["podUid"]),
		}
		env.Items = append(env.Items, Item{Raw: raw, Display: disp})
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	cols := []TableColumn{
		{Header: "TASK", Get: func(it Item) string { return DisplayString(it, "task_name") }},
		{Header: "STATUS", Get: func(it Item) string { return DisplayString(it, "status") }},
		{Header: "MODE", Get: func(it Item) string { return DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host_node") }},
		{Header: "CORE_UTIL", Get: func(it Item) string { return DisplayString(it, "core_util") }},
		{Header: "MEM", Get: func(it Item) string { return DisplayString(it, "mem_used") }},
		{Header: "POD_UID", Get: func(it Item) string { return DisplayString(it, "pod_uid") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}
