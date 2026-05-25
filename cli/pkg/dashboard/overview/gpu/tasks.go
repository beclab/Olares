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
//
// Implementation: BuildTasksEnvelope produces the envelope (no
// stdout); RunTasks adds the legacy "(no vGPU tasks)" prose lines
// and the table writer. RunDefault (sections envelope) reuses
// BuildTasksEnvelope directly so a tasks-section transport error
// doesn't print prose ahead of the graphics section.
func RunTasks(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	env := BuildTasksEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	if env.Meta.Empty {
		switch env.Meta.EmptyReason {
		case "no_vgpu_integration", "no_gpu_detected":
			fmt.Fprintln(os.Stdout, "(no vGPU tasks)")
		case "vgpu_unavailable":
			// stderr advisory already printed by VgpuUnavailableFromError.
		}
		return nil
	}
	return WriteTasksTable(os.Stdout, env)
}

// BuildTasksEnvelope assembles the gpu tasks envelope without any
// stdout side effects. Same 3-state empty-data taxonomy as
// BuildListEnvelope. Used by RunTasks (Shape A leaf) and by
// RunDefault as the `tasks` section (Shape B parent).
func BuildTasksEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) pkgdashboard.Envelope {
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
			return env
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUTasks, now, os.Stderr); ok {
			if advisoryNote != "" {
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			return unavail
		}
		env.Meta.Error = err.Error()
		env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		return env
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		return env
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
	return env
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
