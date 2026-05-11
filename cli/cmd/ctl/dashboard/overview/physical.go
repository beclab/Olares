package overview

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

func newOverviewPhysicalCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "physical",
		Short:         "9-row cluster-level resource snapshot (CPU/Memory/Disk/Pods/Net + extras)",
		Example:       `  olares-cli dashboard overview physical -o json`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewPhysical(c.Context(), f)
		},
	}
	return cmd
}

func runOverviewPhysical(ctx context.Context, f *cmdutil.Factory) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildPhysicalEnvelope(ctx, c, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writePhysicalTable(env)
		},
	}
	return r.Run(ctx)
}

// buildPhysicalEnvelope is the shared fetcher used by both `overview
// physical` (standalone) and the `overview` default sections envelope.
func buildPhysicalEnvelope(ctx context.Context, c *Client, now time.Time) (Envelope, error) {
	metrics := []string{
		"cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation",
		"cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_memory_utilisation",
		"cluster_disk_size_usage", "cluster_disk_size_capacity", "cluster_disk_size_utilisation",
		"cluster_pod_running_count", "cluster_pod_quota",
		"cluster_net_bytes_received", "cluster_net_bytes_transmitted",
	}
	res, err := fetchClusterMetrics(ctx, c, metrics, defaultClusterWindow(), now, false)
	if err != nil {
		return Envelope{Kind: KindOverviewPhysical}, err
	}
	last := format.GetLastMonitoringData(res, 0)
	rows := derivePhysicalRows(last)
	items := make([]Item, 0, len(rows))
	for _, r := range rows {
		raw := map[string]any{
			"metric":      r.Key,
			"label":       r.Label,
			"value":       r.Value,
			"total":       r.Total,
			"unit":        r.Unit,
			"utilisation": r.Utilisation,
		}
		if r.Detail != "" {
			raw["detail"] = r.Detail
		}
		display := map[string]any{
			"metric":      r.Label,
			"value":       formatPhysicalValue(r),
			"unit":        r.Unit,
			"utilisation": percentString(r.Utilisation),
			"detail":      r.Detail,
		}
		items = append(items, Item{Raw: raw, Display: display})
	}
	return Envelope{
		Kind:  KindOverviewPhysical,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

// derivePhysicalRows turns the last-sample map into the 9 (or so) rows the
// SPA renders. Each row carries the headline number + the unit + a
// utilisation ratio. Ordering mirrors the SPA's panel.
func derivePhysicalRows(last map[string]format.LastMonitoringSample) []physicalMetric {
	cpuUsage := sampleFloat(last["cluster_cpu_usage"])
	cpuTotal := sampleFloat(last["cluster_cpu_total"])
	cpuUtil := sampleFloat(last["cluster_cpu_utilisation"])

	memUsage := sampleFloat(last["cluster_memory_usage_wo_cache"])
	memTotal := sampleFloat(last["cluster_memory_total"])
	memUtil := sampleFloat(last["cluster_memory_utilisation"])

	diskUsage := sampleFloat(last["cluster_disk_size_usage"])
	diskCap := sampleFloat(last["cluster_disk_size_capacity"])
	diskUtil := sampleFloat(last["cluster_disk_size_utilisation"])

	podsRun := sampleFloat(last["cluster_pod_running_count"])
	podsQuota := sampleFloat(last["cluster_pod_quota"])

	netIn := sampleFloat(last["cluster_net_bytes_received"])
	netOut := sampleFloat(last["cluster_net_bytes_transmitted"])

	rows := []physicalMetric{
		{Key: "cpu", Label: "CPU", Value: cpuUsage, Total: cpuTotal, Unit: "core", Utilisation: cpuUtil},
		{Key: "memory", Label: "Memory", Value: memUsage, Total: memTotal, Unit: format.GetSuitableUnit(memTotal, format.UnitTypeMemory), Utilisation: memUtil},
		{Key: "disk", Label: "Disk", Value: diskUsage, Total: diskCap, Unit: format.GetSuitableUnit(diskCap, format.UnitTypeDisk), Utilisation: diskUtil},
		{Key: "pods", Label: "Pods", Value: podsRun, Total: podsQuota, Unit: "", Utilisation: safeRatio(podsRun, podsQuota)},
		{Key: "net_in", Label: "Net In", Value: netIn, Total: 0, Unit: format.GetSuitableUnit(netIn, format.UnitTypeThroughput), Utilisation: 0, Detail: format.GetThroughput(formatFloat(netIn))},
		{Key: "net_out", Label: "Net Out", Value: netOut, Total: 0, Unit: format.GetSuitableUnit(netOut, format.UnitTypeThroughput), Utilisation: 0, Detail: format.GetThroughput(formatFloat(netOut))},
	}
	return rows
}

// formatPhysicalValue renders the headline value column for a physical
// row, mirroring the SPA's "value / total" + unit formatting.
func formatPhysicalValue(r physicalMetric) string {
	switch r.Key {
	case "cpu":
		return fmt.Sprintf("%.2f / %.2f", r.Value, r.Total)
	case "memory":
		return fmt.Sprintf("%s / %s",
			format.GetDiskSize(formatFloat(r.Value)),
			format.GetDiskSize(formatFloat(r.Total)))
	case "disk":
		return fmt.Sprintf("%s / %s",
			format.GetDiskSize(formatFloat(r.Value)),
			format.GetDiskSize(formatFloat(r.Total)))
	case "pods":
		return fmt.Sprintf("%.0f / %.0f", r.Value, r.Total)
	case "net_in", "net_out":
		return r.Detail
	default:
		return formatFloat(r.Value)
	}
}

func writePhysicalTable(env Envelope) error {
	cols := []TableColumn{
		{Header: "METRIC", Get: func(it Item) string { return DisplayString(it, "metric") }},
		{Header: "VALUE", Get: func(it Item) string { return DisplayString(it, "value") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "utilisation") }},
	}
	return WriteTable(os.Stdout, cols, env.Items)
}

// ----------------------------------------------------------------------------
// overview user — CPU / memory quota
// ----------------------------------------------------------------------------
