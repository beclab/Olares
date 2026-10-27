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

	// Intel and AMD cards never appear in HAMI's list — it only enumerates
	// CUDA devices — so they are collected separately from KubeSphere and
	// appended. Their failures are warnings rather than errors: a broken
	// Intel read should not blank out the NVIDIA cards sitting beside it.
	ksItems, ksWarnings := ksListItems(ctx, c, cf, now)
	if len(ksWarnings) > 0 {
		env.Meta.Warnings = append(env.Meta.Warnings, ksWarnings...)
	}

	if err != nil && len(ksItems) > 0 {
		// HAMI is down or absent but another vendor answered. Report what we
		// have and keep the HAMI failure visible as a warning instead of
		// discarding a working vendor's cards.
		env.Meta.Warnings = append(env.Meta.Warnings, fmt.Sprintf("hami: %v", err))
		env.Items = ksItems
		return env
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
	if len(list) == 0 && len(ksItems) == 0 {
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
	env.Items = append(env.Items, ksItems...)
	return env
}

// ksListItems collects the device rows of every KubeSphere-sourced vendor
// the cluster carries. Returns the items plus one warning per vendor that
// failed, so a partial answer still reaches the user.
func ksListItems(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) ([]pkgdashboard.Item, []string) {
	inv, err := pkgdashboard.DetectGPUVendors(ctx, c)
	if err != nil {
		return nil, []string{fmt.Sprintf("gpu vendor detection: %v", err)}
	}
	var (
		items    []pkgdashboard.Item
		warnings []string
	)
	for _, vendor := range inv.Vendors {
		spec, ok := pkgdashboard.KsSpecFor(vendor)
		if !ok {
			continue
		}
		table, err := pkgdashboard.FetchGPUMetrics(ctx, c, pkgdashboard.DeviceMetricNames(spec),
			pkgdashboard.GPUMetricOptions{Level: pkgdashboard.GPULevelCluster})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", vendor, err))
			continue
		}
		for _, row := range pkgdashboard.JoinDeviceRows(vendor, spec, table) {
			items = append(items, ksRowItem(vendor, spec, row, cf))
		}
	}
	return items, warnings
}

// ksRowItem renders one card into the shared list shape. Every vendor fills
// the same columns so the table stays one table; a column this vendor cannot
// report shows "-" rather than a zero that would read as a real measurement.
func ksRowItem(vendor pkgdashboard.GPUVendor, spec pkgdashboard.KsVendorSpec, row pkgdashboard.DeviceRow, cf *pkgdashboard.CommonFlags) pkgdashboard.Item {
	raw := map[string]any{}
	for k, v := range row {
		raw[k] = v
	}
	hidden := map[string]bool{}
	for _, col := range spec.HiddenColumns {
		hidden[col] = true
	}

	disp := map[string]any{
		"gpu_id":    rowText(row, "uuid"),
		"model":     rowText(row, "type"),
		"host_node": rowText(row, "nodeName"),
		"vendor":    string(vendor),
		"mode":      "-",
		"health":    gpuHealthLabel(row["health"]),
	}
	disp["core_util"] = percentOrDash(row, "coreUtilizedPercent")
	disp["vram_usage"] = percentOrDash(row, "memoryUtilizedPercent")
	if v, ok := pkgdashboard.RowFloat(row, "memoryTotal"); ok {
		disp["vram_total"] = gpuVRAMHuman(v)
	} else {
		disp["vram_total"] = "-"
	}
	if v, ok := pkgdashboard.RowFloat(row, "memoryUsed"); ok {
		disp["vram_used"] = gpuVRAMHuman(v)
	} else {
		disp["vram_used"] = "-"
	}
	if v, ok := pkgdashboard.RowFloat(row, "power"); ok {
		disp["power"] = fmt.Sprintf("%.2f W", v)
	} else {
		disp["power"] = "-"
	}
	if v, ok := pkgdashboard.RowFloat(row, "temperature"); ok && !hidden["temperature"] {
		disp["temperature"] = renderTemperature(v, cf.TempUnit)
	} else {
		disp["temperature"] = "-"
	}
	return pkgdashboard.Item{Raw: raw, Display: disp}
}

// rowText reads an identity field, falling back to "-" so a label the
// exporter omitted does not render as an empty cell.
func rowText(row pkgdashboard.DeviceRow, field string) string {
	v, ok := row[field]
	if !ok || v == nil || v == "" {
		return "-"
	}
	return fmt.Sprintf("%v", v)
}

// percentOrDash formats a percentage column, distinguishing "not reported"
// from a genuine zero.
func percentOrDash(row pkgdashboard.DeviceRow, field string) string {
	v, ok := pkgdashboard.RowFloat(row, field)
	if !ok {
		return "-"
	}
	return percentDirect(v)
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
