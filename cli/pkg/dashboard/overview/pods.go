package overview

import (
	"context"
	"fmt"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// RunPods is the cmd-side entry point for `dashboard overview pods`.
// Forwards to the per-node scaffold with the pod-count metric set
// and the 4-column table schema the SPA's Pods panel renders.
func RunPods(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	return RunPerNodeMetric(ctx, c, cf, pkgdashboard.KindOverviewPods,
		podsMetricSet(), podsColumns(), podsDisplay)
}

func podsMetricSet() []string {
	return []string{
		"node_pod_running_count", "node_pod_quota", "node_pod_utilisation",
	}
}

func podsColumns() []pkgdashboard.TableColumn {
	return []pkgdashboard.TableColumn{
		{Header: "NODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "node") }},
		{Header: "RUNNING", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "running") }},
		{Header: "QUOTA", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "quota") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "util") }},
	}
}

func podsDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
	running := sampleFloat(last["node_pod_running_count"])
	quota := sampleFloat(last["node_pod_quota"])
	util := sampleFloat(last["node_pod_utilisation"])
	raw := map[string]any{"node": node, "running": running, "quota": quota, "util": util}
	disp := map[string]any{
		"node":    node,
		"running": fmt.Sprintf("%.0f", running),
		"quota":   fmt.Sprintf("%.0f", quota),
		"util":    percentString(util),
	}
	return raw, disp
}
