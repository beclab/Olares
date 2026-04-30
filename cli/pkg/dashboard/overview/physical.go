package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// PhysicalMetric is one row of the SPA's Physical Resources panel.
// Columns: metric / value / unit / utilisation / detail. Names mirror
// the SPA's rendering conventions; exported so the sections aggregator
// (RunDefault) and the per-leaf table render share one canonical shape.
type PhysicalMetric struct {
	Key         string  // canonical metric key (cpu / memory / disk / pods / net_in / net_out)
	Label       string  // human-friendly metric name shown in column 1
	Value       float64 // headline numeric value (used / running)
	Total       float64 // total / quota
	Unit        string  // SPA unit suffix
	Utilisation float64 // 0..1 ratio
	Detail      string  // free-form detail string (used by net rows)
}

// RunPhysical is the cmd-side entry point. Owns the watch-aware
// Runner so cmd-side never sees Runner; per-iteration body delegates
// to BuildPhysicalEnvelope and (in table mode) WritePhysicalTable.
func RunPhysical(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildPhysicalEnvelope(ctx, c, cf, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 60
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WritePhysicalTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildPhysicalEnvelope is the shared fetcher used by both
// `overview physical` (standalone) and the `overview` default
// sections envelope (RunDefault). cf is threaded through so the
// monitoring fetch honours --since / --start / --end without going
// through a global; otherwise the function stays cobra-agnostic.
func BuildPhysicalEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) (pkgdashboard.Envelope, error) {
	metrics := []string{
		"cluster_cpu_usage", "cluster_cpu_total", "cluster_cpu_utilisation",
		"cluster_memory_usage_wo_cache", "cluster_memory_total", "cluster_memory_utilisation",
		"cluster_disk_size_usage", "cluster_disk_size_capacity", "cluster_disk_size_utilisation",
		"cluster_pod_running_count", "cluster_pod_quota",
		"cluster_net_bytes_received", "cluster_net_bytes_transmitted",
	}
	res, err := pkgdashboard.FetchClusterMetrics(ctx, c, cf, metrics, pkgdashboard.DefaultClusterWindow(), now, false)
	if err != nil {
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewPhysical}, err
	}
	last := format.GetLastMonitoringData(res, 0)
	rows := DerivePhysicalRows(last)
	items := make([]pkgdashboard.Item, 0, len(rows))
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
		items = append(items, pkgdashboard.Item{Raw: raw, Display: display})
	}
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewPhysical,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: items,
	}, nil
}

// DerivePhysicalRows turns the last-sample map into the 6 rows the SPA
// renders (cpu / memory / disk / pods / net_in / net_out). Each row
// carries the headline number + the unit + a utilisation ratio.
// Ordering mirrors the SPA's panel; consumers who scrape the table
// rely on the row index being stable.
func DerivePhysicalRows(last map[string]format.LastMonitoringSample) []PhysicalMetric {
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

	rows := []PhysicalMetric{
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
// row, mirroring the SPA's "value / total" + unit formatting. Private
// because every caller is in this file (Run / Build / Write).
func formatPhysicalValue(r PhysicalMetric) string {
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

// WritePhysicalTable renders env.Items as the SPA-aligned 3-column
// table (METRIC / VALUE / UTIL). Exported so the package's _test.go
// can capture into a buffer without redirecting os.Stdout.
func WritePhysicalTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "METRIC", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "metric") }},
		{Header: "VALUE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "value") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "utilisation") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
