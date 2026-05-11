package overview

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// runPerNodeMetric is the shared workhorse for cpu / memory / pods. It
// fetches the requested metric set against /v1alpha3/nodes, groups by node
// (the `node` label), and renders one row per node with the columns / display
// the caller specifies.
func runPerNodeMetric(ctx context.Context, f *cmdutil.Factory, kind string, metrics []string, cols []TableColumn, disp func(node string, last map[string]format.LastMonitoringSample) (rawCols, dispCols map[string]any)) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       common,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			env, err := buildPerNodeEnvelope(ctx, c, kind, metrics, disp, now)
			if err != nil {
				return env, err
			}
			env.Meta = NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User)
			env.Meta.RecommendedPollSeconds = 60
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, WriteTable(os.Stdout, cols, env.Items)
		},
	}
	return r.Run(ctx)
}

// buildPerNodeEnvelope shells out to /v1alpha3/nodes and groups the results
// by the `node` label. Unlike fetchClusterMetrics (which collapses by
// metric_name), per-node metrics carry one row per node within each metric;
// we transpose into one Item per node.
func buildPerNodeEnvelope(ctx context.Context, c *Client, kind string, metrics []string, disp func(node string, last map[string]format.LastMonitoringSample) (rawCols, dispCols map[string]any), now time.Time) (Envelope, error) {
	q := monitoringQuery(metrics, defaultDetailWindow(), now, false)
	var raw struct {
		Results []struct {
			MetricName string `json:"metric_name"`
			Data       struct {
				Result []struct {
					Metric map[string]string `json:"metric"`
					Values [][]any           `json:"values"`
					Value  []any             `json:"value"`
				} `json:"result"`
			} `json:"data"`
		} `json:"results"`
	}
	if err := c.DoJSON(ctx, http.MethodGet, "/kapis/monitoring.kubesphere.io/v1alpha3/nodes", q, nil, &raw); err != nil {
		return Envelope{Kind: kind}, err
	}
	// Group rows by node label.
	type nodeBucket struct {
		samples map[string]format.LastMonitoringSample
	}
	buckets := map[string]*nodeBucket{}
	order := []string{}
	for _, r := range raw.Results {
		for _, e := range r.Data.Result {
			node := e.Metric["node"]
			if node == "" {
				node = e.Metric["instance"]
			}
			if node == "" {
				continue
			}
			b, ok := buckets[node]
			if !ok {
				b = &nodeBucket{samples: map[string]format.LastMonitoringSample{}}
				buckets[node] = b
				order = append(order, node)
			}
			b.samples[r.MetricName] = lastSampleFromRow(e.Values, e.Value)
		}
	}
	sort.Strings(order)
	items := make([]Item, 0, len(order))
	for _, n := range order {
		raws, disps := disp(n, buckets[n].samples)
		items = append(items, Item{Raw: raws, Display: disps})
	}
	return Envelope{
		Kind:  kind,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: items,
	}, nil
}

// lastSampleFromRow lives in pkgdashboard.LastSampleFromRow (overview
// area binds it via common.go).

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

func cpuColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "FREQ", Get: func(it Item) string { return DisplayString(it, "freq") }},
		{Header: "CORES", Get: func(it Item) string { return DisplayString(it, "cores") }},
		{Header: "CPU_UTIL", Get: func(it Item) string { return DisplayString(it, "cpu_util") }},
		{Header: "USER", Get: func(it Item) string { return DisplayString(it, "user") }},
		{Header: "SYSTEM", Get: func(it Item) string { return DisplayString(it, "system") }},
		{Header: "IOWAIT", Get: func(it Item) string { return DisplayString(it, "iowait") }},
		{Header: "LOAD1", Get: func(it Item) string { return DisplayString(it, "load1") }},
		{Header: "LOAD5", Get: func(it Item) string { return DisplayString(it, "load5") }},
		{Header: "LOAD15", Get: func(it Item) string { return DisplayString(it, "load15") }},
		{Header: "TEMP", Get: func(it Item) string { return DisplayString(it, "temp") }},
	}
}

func cpuDisplay(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
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
		"temp":     renderTemperature(temp, common.TempUnit),
	}
	return raw, disp
}

func memoryPhysicalMetricSet() []string {
	return []string{
		"node_memory_total", "node_memory_usage_wo_cache", "node_memory_available",
		"node_memory_utilisation", "node_memory_cached", "node_memory_buffers",
	}
}

func memoryPhysicalColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "AVAIL", Get: func(it Item) string { return DisplayString(it, "avail") }},
		{Header: "BUFFERS", Get: func(it Item) string { return DisplayString(it, "buffers") }},
		{Header: "CACHED", Get: func(it Item) string { return DisplayString(it, "cached") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
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

func memorySwapColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "TOTAL", Get: func(it Item) string { return DisplayString(it, "total") }},
		{Header: "USED", Get: func(it Item) string { return DisplayString(it, "used") }},
		{Header: "PG_IN", Get: func(it Item) string { return DisplayString(it, "pg_in") }},
		{Header: "PG_OUT", Get: func(it Item) string { return DisplayString(it, "pg_out") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
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

func podsMetricSet() []string {
	return []string{
		"node_pod_running_count", "node_pod_quota", "node_pod_utilisation",
	}
}

func podsColumns() []TableColumn {
	return []TableColumn{
		{Header: "NODE", Get: func(it Item) string { return DisplayString(it, "node") }},
		{Header: "RUNNING", Get: func(it Item) string { return DisplayString(it, "running") }},
		{Header: "QUOTA", Get: func(it Item) string { return DisplayString(it, "quota") }},
		{Header: "UTIL", Get: func(it Item) string { return DisplayString(it, "util") }},
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
