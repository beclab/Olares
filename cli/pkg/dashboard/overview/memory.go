package overview

import (
	"context"
	"fmt"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// RunMemory is the cmd-side entry point for
// `dashboard overview memory --mode <physical|swap>`. The two modes
// share the per-node scaffold but diverge on metric set + column
// list + row builder; mode validation is owned here so the cmd-side
// leaf stays a one-liner.
func RunMemory(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, mode string) error {
	switch mode {
	case "", "physical":
		return RunPerNodeMetric(ctx, c, cf, pkgdashboard.KindOverviewMemory,
			memoryPhysicalMetricSet(), memoryPhysicalColumns(), memoryPhysicalDisplay)
	case "swap":
		return RunPerNodeMetric(ctx, c, cf, pkgdashboard.KindOverviewMemory,
			memorySwapMetricSet(), memorySwapColumns(), memorySwapDisplay)
	default:
		return fmt.Errorf("--mode: %q must be physical or swap", mode)
	}
}

// memoryPhysicalMetricSet pins the 6 series the SPA's memory panel
// reads — kept package-private so the leaf RunE never has to spell
// them out.
func memoryPhysicalMetricSet() []string {
	return []string{
		"node_memory_total", "node_memory_usage_wo_cache", "node_memory_available",
		"node_memory_utilisation", "node_memory_cached", "node_memory_buffers",
	}
}

func memoryPhysicalColumns() []pkgdashboard.TableColumn {
	return []pkgdashboard.TableColumn{
		{Header: "NODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "total") }},
		{Header: "USED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "used") }},
		{Header: "AVAIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "avail") }},
		{Header: "BUFFERS", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "buffers") }},
		{Header: "CACHED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cached") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "util") }},
	}
}

func memoryPhysicalDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	total := sampleFloat(last["node_memory_total"])
	used := sampleFloat(last["node_memory_usage_wo_cache"])
	avail := sampleFloat(last["node_memory_available"])
	util := sampleFloat(last["node_memory_utilisation"])
	cached := sampleFloat(last["node_memory_cached"])
	buffers := sampleFloat(last["node_memory_buffers"])
	raw := map[string]any{
		"node": node, "total": total, "used": used, "avail": avail,
		"util": util, "cached": cached, "buffers": buffers, "mode": "physical",
	}
	disp := map[string]any{
		"node":    node,
		"total":   format.GetDiskSize(formatFloat(total)),
		"used":    format.GetDiskSize(formatFloat(used)),
		"avail":   format.GetDiskSize(formatFloat(avail)),
		"buffers": format.GetDiskSize(formatFloat(buffers)),
		"cached":  format.GetDiskSize(formatFloat(cached)),
		"util":    percentString(util),
	}
	return raw, disp
}

func memorySwapMetricSet() []string {
	return []string{
		"node_memory_swap_total", "node_memory_swap_used",
		"node_memory_pgpgin_rate", "node_memory_pgpgout_rate",
	}
}

func memorySwapColumns() []pkgdashboard.TableColumn {
	return []pkgdashboard.TableColumn{
		{Header: "NODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "total") }},
		{Header: "USED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "used") }},
		{Header: "PG_IN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "pg_in") }},
		{Header: "PG_OUT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "pg_out") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "util") }},
	}
}

func memorySwapDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	total := sampleFloat(last["node_memory_swap_total"])
	used := sampleFloat(last["node_memory_swap_used"])
	pgIn := sampleFloat(last["node_memory_pgpgin_rate"])
	pgOut := sampleFloat(last["node_memory_pgpgout_rate"])
	util := safeRatio(used, total)
	raw := map[string]any{
		"node": node, "total": total, "used": used, "pg_in": pgIn, "pg_out": pgOut,
		"util": util, "mode": "swap",
	}
	disp := map[string]any{
		"node":   node,
		"total":  format.GetDiskSize(formatFloat(total)),
		"used":   format.GetDiskSize(formatFloat(used)),
		"pg_in":  format.WorthValue(formatFloat(pgIn)),
		"pg_out": format.WorthValue(formatFloat(pgOut)),
		"util":   percentString(util),
	}
	return raw, disp
}
