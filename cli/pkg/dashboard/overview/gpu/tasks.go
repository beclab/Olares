package gpu

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// RunTasks is the cmd-side entry point for `dashboard overview gpu
// tasks`. One-shot. The advisory + integration-gate flow mirrors
// RunList — kept open-coded per leaf so each leaf's HTTP-status
// branches stay close to the call site (the SPA's task page does
// the same).
func RunTasks(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	list, err := pkgdashboard.FetchTaskList(ctx, c, nil)

	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUTasks,
		Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, env)
			}
			fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
			return nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUTasks, now, os.Stderr); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			if cf.Output == pkgdashboard.OutputJSON {
				return pkgdashboard.WriteJSON(os.Stdout, unavail)
			}
			return nil
		}
		return err
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		if cf.Output == pkgdashboard.OutputJSON {
			return pkgdashboard.WriteJSON(os.Stdout, env)
		}
		fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
		return nil
	}
	// HAMI's task entries match the SPA's `TaskItem` interface.
	// The "core util / mem used" columns are arrays (one element
	// per allocated device) — SPA uses index 0 too. Raw envelope
	// retains the full array so multi-GPU tasks aren't silently
	// truncated.
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
		env.Items = append(env.Items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteTasksTable(os.Stdout, env)
}

// WriteTasksTable renders the per-task summary table.
func WriteTasksTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "TASK", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "task_name") }},
		{Header: "STATUS", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "status") }},
		{Header: "MODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "host_node") }},
		{Header: "CORE_UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "core_util") }},
		{Header: "MEM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mem_used") }},
		{Header: "POD_UID", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "pod_uid") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
