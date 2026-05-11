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

func newOverviewGPUListCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List discovered vGPUs (Graphics management tab; 404 = HAMI not installed)",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUList(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewGPUList(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	now := time.Now()
	advisoryNote, advisoryReason := gpuAdvisory(ctx, c)
	list, err := fetchGraphicsList(ctx, c, nil)
	env := Envelope{Kind: KindOverviewGPUList, Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)}
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
			fmt.Fprintln(os.Stdout, "(no vGPUs detected — HAMI integration absent)")
			return nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUList, now); ok {
			if advisoryNote != "" {
				// Stack the advisory ahead of the unavailability
				// note; humans see "GPU sidebar hidden + HAMI down"
				// in one shot. Agents still get both as a single
				// `meta.note` string separated by " | ".
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
		fmt.Fprintln(os.Stdout, "(no GPUs reported — HAMI integration up but no devices)")
		return nil
	}
	_ = advisoryReason // stash; intentionally not surfaced separately, the SPA does not either
	// Field names below match HAMI's actual response shape (see SPA's
	// `Graphics` interface in src/apps/dashboard/types/gpu.ts and the
	// fixture captured from olarestest005). Earlier revisions guessed
	// at field names like "modelName" / "hostname" / "totalMem" — none
	// of which HAMI ever returns; the table silently rendered "<nil>".
	// We expose the entire HAMI object under `Raw` so agents can pull
	// fields the table doesn't surface (vgpuUsed/vgpuTotal, nodeUid,
	// memoryUtilizedPercent, etc.).
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
			"temperature": renderTemperature(toFloat(g["temperature"]), common.TempUnit),
		}
		env.Items = append(env.Items, Item{Raw: raw, Display: disp})
	}
	if common.Output == OutputJSON {
		return WriteJSON(os.Stdout, env)
	}
	cols := []TableColumn{
		{Header: "GPU_ID", Get: func(it Item) string { return DisplayString(it, "gpu_id") }},
		{Header: "MODEL", Get: func(it Item) string { return DisplayString(it, "model") }},
		{Header: "MODE", Get: func(it Item) string { return DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it Item) string { return DisplayString(it, "host_node") }},
		{Header: "HEALTH", Get: func(it Item) string { return DisplayString(it, "health") }},
		{Header: "CORE_UTIL", Get: func(it Item) string { return DisplayString(it, "core_util") }},
		{Header: "VRAM", Get: func(it Item) string { return DisplayString(it, "vram_total") }},
		{Header: "POWER", Get: func(it Item) string { return DisplayString(it, "power") }},
		{Header: "TEMP", Get: func(it Item) string { return DisplayString(it, "temperature") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}
