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

// RunList is the cmd-side entry point for `dashboard overview gpu`
// (default) and `dashboard overview gpu list`. One-shot — the GPU
// list view is operational ("which devices does HAMI see") and
// doesn't need watch semantics; if a user wants polling they can
// run `watch -n 5 olares-cli dashboard overview gpu list` from
// outside.
func RunList(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	advisoryNote, _ := pkgdashboard.GPUAdvisory(ctx, c, cf, os.Stderr)
	list, err := pkgdashboard.FetchGraphicsList(ctx, c, nil)

	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUList,
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
			fmt.Fprintln(os.Stdout, "(no vGPUs detected — HAMI integration absent)")
			return nil
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUList, now, os.Stderr); ok {
			if advisoryNote != "" {
				// Stack the advisory ahead of the unavailability
				// note; humans see "GPU sidebar hidden + HAMI down"
				// in one shot. Agents still get both as a single
				// `meta.note` string separated by " | ".
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
		fmt.Fprintln(os.Stdout, "(no GPUs reported — HAMI integration up but no devices)")
		return nil
	}
	// Field names below match HAMI's actual response shape (see
	// SPA's `Graphics` interface in src/apps/dashboard/types/gpu.ts
	// and the fixture captured from olarestest005). Earlier
	// revisions guessed at field names like "modelName" / "hostname"
	// / "totalMem" — none of which HAMI ever returns; the table
	// silently rendered "<nil>". We expose the entire HAMI object
	// under `Raw` so agents can pull fields the table doesn't
	// surface (vgpuUsed/vgpuTotal, nodeUid, memoryUtilizedPercent,
	// etc.).
	for _, g := range list {
		raw := map[string]any{}
		for k, v := range g {
			raw[k] = v
		}
		disp := map[string]any{
			"gpu_id":      fmt.Sprintf("%v", g["uuid"]),
			"model":       fmt.Sprintf("%v", g["type"]),
			"mode":        gpuModeLabel(g["shareMode"]),
			"host_node":   fmt.Sprintf("%v", g["nodeName"]),
			"health":      gpuHealthLabel(g["health"]),
			"core_util":   percentDirect(toFloat(g["coreUtilizedPercent"])),
			"vram_total":  gpuVRAMHuman(g["memoryTotal"]),
			"vram_used":   gpuVRAMHuman(g["memoryUsed"]),
			"power":       fmt.Sprintf("%.2f W", toFloat(g["power"])),
			"temperature": renderTemperature(toFloat(g["temperature"]), cf.TempUnit),
		}
		env.Items = append(env.Items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	return WriteListTable(os.Stdout, env)
}

// WriteListTable renders the per-GPU summary table. Column order
// is pinned: agent scrapers depend on the index being stable across
// releases.
func WriteListTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "GPU_ID", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_id") }},
		{Header: "MODEL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "model") }},
		{Header: "MODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "host_node") }},
		{Header: "HEALTH", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "health") }},
		{Header: "CORE_UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "core_util") }},
		{Header: "VRAM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "vram_total") }},
		{Header: "POWER", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "power") }},
		{Header: "TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "temperature") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
