package gpu

import (
	"context"
	"errors"
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
//
// Implementation: BuildListEnvelope assembles the envelope (no
// stdout writes); RunList wraps it with the legacy stdout side
// effects (the "(no vGPUs detected …)" prose lines that earlier
// revisions printed for the bare-leaf invocation). RunDefault
// (sections envelope) reuses BuildListEnvelope directly so a
// graphics-section transport error doesn't print prose ahead of
// the tasks section.
func RunList(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	now := time.Now()
	env := BuildListEnvelope(ctx, c, cf, now)
	if cf.Output == pkgdashboard.OutputJSON {
		return pkgdashboard.WriteJSON(os.Stdout, env)
	}
	if env.Meta.Empty {
		switch env.Meta.EmptyReason {
		case "no_vgpu_integration":
			fmt.Fprintln(os.Stdout, "(no vGPUs detected — HAMI integration absent)")
		case "no_gpu_detected":
			fmt.Fprintln(os.Stdout, "(no GPUs reported — HAMI integration up but no devices)")
		case "vgpu_unavailable":
			// stderr advisory already printed by VgpuUnavailableFromError.
		}
		return nil
	}
	// Unclassifiable transport / 4xx error — BuildListEnvelope
	// stashes it on Meta.Error (Meta.Empty stays false). JSON mode
	// returned the envelope as-is above so agents can branch on
	// `meta.error`; table mode used to fall through to
	// WriteListTable and render a header + "-" row with no
	// diagnostic, silently swallowing the failure. Returning the
	// error here surfaces it to cobra (printed to stderr with a
	// non-zero exit code), restoring the legacy pre-envelope-split
	// behaviour. Checked AFTER the Empty switch so the soft empty
	// states (no_vgpu_integration / vgpu_unavailable / no_gpu_detected)
	// keep their non-fatal semantics — vgpu_unavailable in particular
	// sets both Empty AND Error and is intentionally a clean exit.
	if env.Meta.Error != "" {
		return errors.New(env.Meta.Error)
	}
	return WriteListTable(os.Stdout, env)
}

// BuildListEnvelope assembles the gpu list envelope without any
// stdout side effects. Honors the standard 3-state empty-data
// taxonomy (no_vgpu_integration / vgpu_unavailable /
// no_gpu_detected) and surfaces the GPUAdvisory soft-gate as
// Meta.Note. Used by RunList (Shape A leaf) and by RunDefault as
// the `graphics` section (Shape B parent).
func BuildListEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) pkgdashboard.Envelope {
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
			return env
		}
		if unavail, ok := pkgdashboard.VgpuUnavailableFromError(c, cf, err, pkgdashboard.KindOverviewGPUList, now, os.Stderr); ok {
			if advisoryNote != "" {
				// Stack the advisory ahead of the unavailability
				// note; humans see "GPU sidebar hidden + HAMI down"
				// in one shot. Agents still get both as a single
				// `meta.note` string separated by " | ".
				unavail.Meta.Note = advisoryNote + " | " + unavail.Meta.Note
			}
			return unavail
		}
		// Transport error not classifiable as a soft 4xx/5xx —
		// surface it on the envelope so RunDefault keeps the tasks
		// section. RunList's caller branch above maps Output==JSON
		// to a clean payload too.
		env.Meta.Error = err.Error()
		env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		return env
	}
	if len(list) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		return env
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
	// Display fields mirror SPA's GPUsTable.vue column set verbatim:
	// nodeUid is omitted in favour of the more useful uuid (CLI's
	// `gpu graphics <uuid>` keys on uuid; SPA's column header is
	// labelled "GPU ID" but the underlying field IS nodeUid in the
	// SPA — the CLI departs here intentionally so the column value
	// can be copy-pasted into `graphics <uuid>`). Otherwise: model /
	// mode / host / health / core_util / vram_total / vram_usage /
	// power / temperature, in that order.
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
			"vram_usage":  percentDirect(toFloat(g["memoryUtilizedPercent"])),
			"power":       fmt.Sprintf("%.2f W", toFloat(g["power"])),
			"temperature": renderTemperature(toFloat(g["temperature"]), cf.TempUnit),
		}
		env.Items = append(env.Items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	return env
}

// WriteListTable renders the per-GPU summary table. Column order
// is pinned: agent scrapers depend on the index being stable across
// releases. VRAM_USAGE is the SPA's `memoryUtilizedPercent` column
// that earlier revisions of this table omitted; readded so the CLI
// matches Graphics management tab cell-for-cell.
func WriteListTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "GPU_ID", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "gpu_id") }},
		{Header: "MODEL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "model") }},
		{Header: "MODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "mode") }},
		{Header: "HOST", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "host_node") }},
		{Header: "HEALTH", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "health") }},
		{Header: "CORE_UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "core_util") }},
		{Header: "VRAM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "vram_total") }},
		{Header: "VRAM_USAGE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "vram_usage") }},
		{Header: "POWER", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "power") }},
		{Header: "TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "temperature") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
