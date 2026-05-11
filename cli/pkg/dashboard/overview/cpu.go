package overview

import (
	"context"
	"fmt"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// RunCPU is the cmd-side entry point for `dashboard overview cpu`.
// Forwards to the per-node scaffold with a CPU-specific metric set,
// 11-column table schema, and the cpuDisplay row builder. cf carries
// --output / --watch / TempUnit etc.
func RunCPU(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	return RunPerNodeMetric(ctx, c, cf, pkgdashboard.KindOverviewCPU, cpuMetricSet(), cpuColumns(), cpuDisplayFn(cf))
}

// cpuMetricSet — column 1:1 with SPA Overview2/CPU/config.ts.
func cpuMetricSet() []string {
	return []string{
		"node_cpu_total", "node_cpu_utilisation",
		"node_user_cpu_usage", "node_system_cpu_usage", "node_iowait_cpu_usage",
		"node_load1", "node_load5", "node_load15",
		"node_cpu_temp_celsius",
		"node_cpu_base_frequency_hertz_max",
	}
}

func cpuColumns() []pkgdashboard.TableColumn {
	return []pkgdashboard.TableColumn{
		{Header: "NODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "node") }},
		{Header: "FREQ", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "freq") }},
		{Header: "CORES", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cores") }},
		{Header: "CPU_UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "cpu_util") }},
		{Header: "USER", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "user") }},
		{Header: "SYSTEM", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "system") }},
		{Header: "IOWAIT", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "iowait") }},
		{Header: "LOAD1", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "load1") }},
		{Header: "LOAD5", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "load5") }},
		{Header: "LOAD15", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "load15") }},
		{Header: "TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "temp") }},
	}
}

// cpuDisplayFn closes over cf so the per-row temperature render can
// honour --temp-unit. The scaffold's PerNodeDisplayFn signature is
// kept cf-free because most leaves don't need it; cpu/memory pull
// in cf via this builder.
func cpuDisplayFn(cf *pkgdashboard.CommonFlags) PerNodeDisplayFn {
	return func(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
		cores := sampleFloat(last["node_cpu_total"])
		cpuUtil := sampleFloat(last["node_cpu_utilisation"])
		userCPU := sampleFloat(last["node_user_cpu_usage"])
		sysCPU := sampleFloat(last["node_system_cpu_usage"])
		iowait := sampleFloat(last["node_iowait_cpu_usage"])
		load1 := sampleFloat(last["node_load1"])
		load5 := sampleFloat(last["node_load5"])
		load15 := sampleFloat(last["node_load15"])
		temp := sampleFloat(last["node_cpu_temp_celsius"])
		freq := sampleFloat(last["node_cpu_base_frequency_hertz_max"])
		raw := map[string]any{
			"node":     node,
			"freq_hz":  freq,
			"cores":    cores,
			"cpu_util": cpuUtil,
			"user":     userCPU,
			"system":   sysCPU,
			"iowait":   iowait,
			"load1":    load1, "load5": load5, "load15": load15,
			"temp_c": temp,
		}
		disp := map[string]any{
			"node":     node,
			"freq":     format.FormatFrequency(freq, "Hz"),
			"cores":    fmt.Sprintf("%.0f", cores),
			"cpu_util": percentString(cpuUtil),
			"user":     percentString(userCPU),
			"system":   percentString(sysCPU),
			"iowait":   percentString(iowait),
			"load1":    fmt.Sprintf("%.2f", load1),
			"load5":    fmt.Sprintf("%.2f", load5),
			"load15":   fmt.Sprintf("%.2f", load15),
			"temp":     renderTemperature(temp, cf.TempUnit),
		}
		return raw, disp
	}
}
