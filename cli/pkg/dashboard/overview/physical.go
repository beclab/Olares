package overview

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
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
//
// Row inventory (mirrors the SPA's `Overview2/ClusterResource.vue`
// card stack — every visible card on the page becomes a row here):
//
//  1. cpu / memory / disk / pods / net_in / net_out — always emitted
//     from the kapis cluster monitoring fetch.
//  2. gpu — only when HAMI vGPU is installed AND the GPU list is
//     non-empty. Aggregates `memoryUsed` / `memoryTotal` (MiB)
//     across every device, presented in bytes so the SPA-aligned
//     GiB/TiB unit inference picks the right suffix. HAMI 404 / 5xx
//     just drops the row (matches the SPA hiding the GPU card when
//     the card-gauge fetch comes back empty); the failure is
//     surfaced via Meta.Warnings so JSON consumers can demux.
//  3. fan_cpu / fan_gpu — only when EnsureSystemStatus reports an
//     Olares One device AND `/user-service/api/mdns/olares-one/cpu-gpu`
//     responds successfully. Each row holds the live RPM reading
//     against the `FanSpeedMaxCPU / FanSpeedMaxGPU` constants the
//     SPA uses; non-Olares-One / fan endpoint 404 just drops
//     these rows. Mirrors the SPA's `FanStore.isOlaresOneDevice`
//     gate.
//
// Soft-failure semantics: a missing GPU or fan card NEVER aborts
// the envelope. The 6 always-emitted rows ship even if every
// optional card fails. Only an upstream cluster-metrics failure
// (`FetchClusterMetrics`) returns an error from this function.
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

	// Optional fan-out: GPU summary + fan readings. Run in
	// parallel so a slow / 404 HAMI doesn't gate the fan probe
	// and vice versa. Each branch writes to its own slot so the
	// final row order stays deterministic regardless of which
	// goroutine completes first (gpu first, then fan_cpu /
	// fan_gpu — matches the SPA's `ClusterResource.vue` card
	// stack order). Warnings accumulate behind a mutex.
	var (
		warnMu        sync.Mutex
		extraWarnings []string
		gpuRow        PhysicalMetric
		gpuRowOK      bool
		fanRows       []PhysicalMetric
	)
	addWarning := func(msg string) {
		warnMu.Lock()
		extraWarnings = append(extraWarnings, msg)
		warnMu.Unlock()
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		gpuRow, gpuRowOK = buildGPUSummaryRow(ctx, c, addWarning)
	}()
	go func() {
		defer wg.Done()
		fanRows = buildFanRows(ctx, c, addWarning)
	}()
	wg.Wait()

	if gpuRowOK {
		rows = append(rows, gpuRow)
	}
	rows = append(rows, fanRows...)

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
	env := pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewPhysical,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: items,
	}
	if len(extraWarnings) > 0 {
		env.Meta.Warnings = extraWarnings
	}
	return env, nil
}

// SPA-aligned PromQL for the cluster-overview GPU card. Lifted
// 1:1 from `Overview2/ClusterResource.vue:cardGaugeConfig` —
// `avg(sum(...) by (instance))` collapses HAMI's per-instance
// metrics down to a single cluster-wide value (HAMI's WebUI
// multiplexes instances and the SPA renders the average; we mirror
// that to keep the CLI number == SPA card number). Result unit is
// MiB (raw `hami_memory_used / hami_memory_size` are MiB on the
// wire); we convert to bytes downstream so format.GetDiskSize can
// pick the right Gi/Ti suffix.
const (
	gpuSummaryUsedQuery  = `avg(sum(hami_memory_used) by (instance))`
	gpuSummaryTotalQuery = `avg(sum(hami_memory_size) by (instance))`
)

// buildGPUSummaryRow mirrors the SPA's cluster-overview GPU card
// 1:1. Two stages:
//
//  1. Primary: HAMI prom queries (cardGaugeConfig in
//     Overview2/ClusterResource.vue). `hami_memory_used` covers
//     ALL VRAM consumption — vGPU containers AND raw CUDA
//     processes the device-level allocation table doesn't track.
//     This is what the SPA renders and so the CLI shows the same
//     number as the dashboard browser tab.
//  2. Fallback: aggregate `g["memoryUsed"] / g["memoryTotal"]`
//     across `/v1/gpus`. Used when prom is missing, misconfigured,
//     or returns no series — keeps a non-zero total visible so the
//     row doesn't disappear just because hami-prometheus is down.
//     Note: the fallback total is always correct, but `used` will
//     undercount when no vGPU is allocated (it's the
//     allocation-table number, not the actual VRAM consumption).
//
// Returns ok=false (no row) when:
//   - HAMI's /v1/gpus returns 404 (no vGPU integration), and
//   - the prom path also has no series / errored.
//   - Or both succeeded but reported total = 0 (no devices).
//
// 5xx from either source emits a `gpu_summary: ...` warning so
// agents can branch on len(meta.warnings)>0 without scanning each
// section.
func buildGPUSummaryRow(ctx context.Context, c *pkgdashboard.Client, addWarning func(string)) (PhysicalMetric, bool) {
	usedMiB, totalMiB, promOK := fetchGPUSummaryFromProm(ctx, c, addWarning)
	if !promOK {
		// Prom unavailable -> fall back to list aggregation.
		// The list endpoint may itself be 404 (no HAMI at all)
		// in which case we just skip the row.
		usedMiB, totalMiB, _ = fetchGPUSummaryFromList(ctx, c, addWarning)
	}
	if totalMiB <= 0 {
		return PhysicalMetric{}, false
	}
	usedBytes := usedMiB * 1024 * 1024
	totalBytes := totalMiB * 1024 * 1024
	return PhysicalMetric{
		Key:         "gpu",
		Label:       "GPU",
		Value:       usedBytes,
		Total:       totalBytes,
		Unit:        format.GetSuitableUnit(totalBytes, format.UnitTypeMemory),
		Utilisation: safeRatio(usedBytes, totalBytes),
	}, true
}

// fetchGPUSummaryFromProm runs the SPA-aligned instant-vector
// queries. Returns (used MiB, total MiB, ok). ok=false when EITHER
// query hard-failed; an empty `data: []` response is NOT a hard
// failure (HAMI prom under "no scrape data" returns an empty
// vector — the caller falls back to list aggregation).
func fetchGPUSummaryFromProm(ctx context.Context, c *pkgdashboard.Client, addWarning func(string)) (used, total float64, ok bool) {
	usedSamples, errU := pkgdashboard.FetchInstantVector(ctx, c, gpuSummaryUsedQuery)
	totalSamples, errT := pkgdashboard.FetchInstantVector(ctx, c, gpuSummaryTotalQuery)
	if errU != nil || errT != nil {
		err := errU
		if err == nil {
			err = errT
		}
		if he, isHTTP := pkgdashboard.IsHTTPError(err); isHTTP && he.Status >= 500 {
			addWarning(fmt.Sprintf("gpu_summary (prom): HAMI %d", he.Status))
		}
		return 0, 0, false
	}
	if len(totalSamples) == 0 {
		// Prom is up but no series — caller falls back.
		return 0, 0, false
	}
	if len(usedSamples) > 0 {
		used = usedSamples[0].Value
	}
	total = totalSamples[0].Value
	return used, total, true
}

// fetchGPUSummaryFromList is the prom-fallback path. Sums
// `memoryUsed` / `memoryTotal` (MiB) across /v1/gpus. Returns
// ok=false when HAMI's list endpoint itself errored or returned
// an empty list.
func fetchGPUSummaryFromList(ctx context.Context, c *pkgdashboard.Client, addWarning func(string)) (used, total float64, ok bool) {
	list, err := pkgdashboard.FetchGraphicsList(ctx, c, nil)
	if err != nil {
		if he, isHTTP := pkgdashboard.IsHTTPError(err); isHTTP && he.Status >= 500 {
			addWarning(fmt.Sprintf("gpu_summary (list): HAMI %d", he.Status))
		}
		return 0, 0, false
	}
	if len(list) == 0 {
		return 0, 0, false
	}
	for _, g := range list {
		used += pkgdashboard.ToFloat(g["memoryUsed"])
		total += pkgdashboard.ToFloat(g["memoryTotal"])
	}
	return used, total, true
}

// buildFanRows returns the two fan rows (cpu / gpu) when running
// on an Olares One device with the cooling endpoint reachable.
// Off-device or fan endpoint 404 → empty slice. Mirrors the SPA's
// `FanStore.isOlaresOneDevice` gate inside ClusterResource.vue
// (the fan card is appended to the cluster-overview options only
// when that flag is true).
func buildFanRows(ctx context.Context, c *pkgdashboard.Client, addWarning func(string)) []PhysicalMetric {
	st, err := c.EnsureSystemStatus(ctx)
	if err != nil || st == nil || !st.IsOlaresOne() {
		return nil
	}
	fan, err := pkgdashboard.FetchSystemFan(ctx, c)
	if err != nil {
		if he, ok := pkgdashboard.IsHTTPError(err); ok && he.Status >= 500 {
			addWarning(fmt.Sprintf("fan_live: cooling endpoint %d", he.Status))
		}
		return nil
	}
	cpuMax := float64(pkgdashboard.FanSpeedMaxCPU)
	gpuMax := float64(pkgdashboard.FanSpeedMaxGPU)
	return []PhysicalMetric{
		{
			Key:         "fan_cpu",
			Label:       "Fan CPU",
			Value:       fan.CPUFanSpeed,
			Total:       cpuMax,
			Unit:        "RPM",
			Utilisation: safeRatio(fan.CPUFanSpeed, cpuMax),
			Detail:      fmt.Sprintf("%.0f / %.0f RPM", fan.CPUFanSpeed, cpuMax),
		},
		{
			Key:         "fan_gpu",
			Label:       "Fan GPU",
			Value:       fan.GPUFanSpeed,
			Total:       gpuMax,
			Unit:        "RPM",
			Utilisation: safeRatio(fan.GPUFanSpeed, gpuMax),
			Detail:      fmt.Sprintf("%.0f / %.0f RPM", fan.GPUFanSpeed, gpuMax),
		},
	}
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
	case "gpu":
		// GPU VRAM rendered the same way as cluster memory —
		// SPA's `getDiskSize(memoryUsed * 1024 * 1024)` -> GiB
		// once the magnitude crosses 1 GiB.
		return fmt.Sprintf("%s / %s",
			format.GetDiskSize(formatFloat(r.Value)),
			format.GetDiskSize(formatFloat(r.Total)))
	case "fan_cpu", "fan_gpu":
		return r.Detail
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
