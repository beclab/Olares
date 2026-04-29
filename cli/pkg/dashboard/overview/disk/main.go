package disk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// RunMain is the cmd-side entry point for `dashboard overview disk
// main`. Owns the watch-aware Runner so cmd-side never sees Runner;
// per-iteration body delegates to BuildMainEnvelope and (in table
// mode) WriteMainTable.
func RunMain(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags) error {
	r := &pkgdashboard.Runner{
		Flags:       cf,
		Recommended: 60 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (pkgdashboard.Envelope, error) {
			env, err := BuildMainEnvelope(ctx, c, cf, now)
			if err != nil {
				return env, err
			}
			env.Meta = pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User)
			env.Meta.RecommendedPollSeconds = 60
			if cf.Output == pkgdashboard.OutputJSON {
				return env, nil
			}
			return env, WriteMainTable(os.Stdout, env)
		},
	}
	return r.Run(ctx)
}

// BuildMainEnvelope mirrors SPA Overview2/Disk/IndexPage's "main"
// view (`getDiskOptions` in Overview2/Disk/config.ts:131). The
// driver metric is `node_disk_smartctl_info`; for each smartctl row
// we attach the matching samples from the auxiliary metrics by
// `(device, node)` label match (config.ts:142–151).
//
// Auxiliary metrics:
//
//   - node_disk_temp_celsius          → temperature
//   - node_one_disk_capacity_size     → reported "used capacity" total
//   - node_one_disk_avail_size        → free space (used = capacity - avail)
//   - node_disk_power_on_hours        → power-on hours
//   - node_one_disk_data_bytes_written → lifetime write volume
//
// All queries are instant (`step=0s`) — the SPA passes
// `step:'0s'` in IndexPage.vue:113.
//
// Per the user's policy decision the per-second IOPS / throughput
// columns are intentionally NOT emitted; the table is otherwise a
// 1:1 of the SPA card content.
func BuildMainEnvelope(ctx context.Context, c *pkgdashboard.Client, cf *pkgdashboard.CommonFlags, now time.Time) (pkgdashboard.Envelope, error) {
	metrics := []string{
		"node_disk_smartctl_info",
		"node_disk_temp_celsius",
		"node_one_disk_capacity_size",
		"node_one_disk_avail_size",
		"node_disk_power_on_hours",
		"node_one_disk_data_bytes_written",
	}
	q := pkgdashboard.MonitoringQuery(cf, metrics, pkgdashboard.DefaultDetailWindow(), now, true)
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
		return pkgdashboard.Envelope{Kind: pkgdashboard.KindOverviewDiskMain}, err
	}

	// Build a lookup of all auxiliary samples + the SMART rows.
	type smartRow struct {
		labels map[string]string
	}
	type auxSample struct {
		labels map[string]string
		sample format.LastMonitoringSample
	}
	smarts := []smartRow{}
	aux := map[string][]auxSample{}
	for _, r := range raw.Results {
		if r.MetricName == "node_disk_smartctl_info" {
			for _, e := range r.Data.Result {
				smarts = append(smarts, smartRow{labels: e.Metric})
			}
			continue
		}
		for _, e := range r.Data.Result {
			aux[r.MetricName] = append(aux[r.MetricName], auxSample{
				labels: e.Metric,
				sample: lastSampleFromRow(e.Values, e.Value),
			})
		}
	}

	// findAux mirrors `getLastMonitoringDataWithPath` (utils/
	// monitoring) + the predicate in config.ts:143–151:
	// metric.device | metric.disk_name must contain smart.device,
	// AND metric.node must equal smart.node.
	findAux := func(metricName string, smartDevice, smartNode string) format.LastMonitoringSample {
		samples, ok := aux[metricName]
		if !ok {
			return format.LastMonitoringSample{Empty: true}
		}
		for _, s := range samples {
			dev := s.labels["device"]
			if dev == "" {
				dev = s.labels["disk_name"]
			}
			node := s.labels["node"]
			if dev != "" && smartDevice != "" && strings.Contains(dev, smartDevice) && node == smartNode {
				return s.sample
			}
		}
		return format.LastMonitoringSample{Empty: true}
	}

	// Stable order: first by node, then by device, matching the
	// SPA's implicit order (smartctl rows arrive in the BFF's
	// natural order but we sort to keep table output deterministic
	// in tests).
	sort.SliceStable(smarts, func(i, j int) bool {
		if smarts[i].labels["node"] != smarts[j].labels["node"] {
			return smarts[i].labels["node"] < smarts[j].labels["node"]
		}
		return smarts[i].labels["device"] < smarts[j].labels["device"]
	})

	items := make([]pkgdashboard.Item, 0, len(smarts))
	for _, s := range smarts {
		dev := s.labels["device"]
		node := s.labels["node"]
		name := s.labels["name"]
		rotational := s.labels["rotational"]
		logicalBlk := s.labels["logical_block_size"]
		if logicalBlk == "" {
			logicalBlk = "512"
		}
		physicalBlk := s.labels["physical_block_size"]
		if physicalBlk == "" {
			physicalBlk = "512"
		}
		const fourK = "4096"
		is4K := (rotational == "0" && logicalBlk == fourK) ||
			(rotational == "1" && logicalBlk == fourK && physicalBlk == fourK)
		typeStr := "SSD"
		if rotational == "1" {
			typeStr = "HDD"
		}
		healthOK := s.labels["health_ok"] == "true"
		healthStr := "Exception"
		if healthOK {
			healthStr = "Normal"
		}

		capLabel := s.labels["capacity"]
		capLabelF, _ := strconv.ParseFloat(capLabel, 64)

		capSample := findAux("node_one_disk_capacity_size", dev, node)
		availSample := findAux("node_one_disk_avail_size", dev, node)
		tempSample := findAux("node_disk_temp_celsius", dev, node)
		powerSample := findAux("node_disk_power_on_hours", dev, node)
		writtenSample := findAux("node_one_disk_data_bytes_written", dev, node)

		capUsed := sampleFloat(capSample)
		availUsed := sampleFloat(availSample)
		usedSize := capUsed - availUsed
		ratio := safeRatio(usedSize, capUsed)
		celsius := sampleFloat(tempSample)
		powerHours := sampleFloat(powerSample)
		written := sampleFloat(writtenSample)

		raw := map[string]any{
			"device":             dev,
			"node":               node,
			"name":               name,
			"type":               typeStr,
			"rotational":         rotational,
			"health_ok":          healthOK,
			"capacity_total":     capLabelF,
			"capacity_used":      capUsed,
			"capacity_avail":     availUsed,
			"used":               usedSize,
			"used_ratio":         ratio,
			"model":              s.labels["model"],
			"serial":             s.labels["serial"],
			"protocol":           s.labels["protocol"],
			"firmware":           s.labels["firmware"],
			"logical_block":      logicalBlk,
			"physical_block":     physicalBlk,
			"is_4k_native":       is4K,
			"temperature_c":      celsius,
			"power_on_hours":     powerHours,
			"data_bytes_written": written,
		}
		disp := map[string]any{
			"device":         dispOrDash(dev),
			"node":           dispOrDash(node),
			"type":           typeStr,
			"health":         healthStr,
			"total":          format.GetDiskSize(capLabel),
			"used":           format.GetDiskSize(formatFloat(usedSize)),
			"avail":          format.GetDiskSize(formatFloat(availUsed)),
			"util":           percentString(ratio),
			"temperature":    renderDiskTemperature(celsius, cf.TempUnit),
			"model":          dispOrDash(s.labels["model"]),
			"serial":         dispOrDash(s.labels["serial"]),
			"protocol":       dispOrDash(s.labels["protocol"]),
			"firmware":       dispOrDash(s.labels["firmware"]),
			"is_4k_native":   ifYesNo(is4K),
			"power_on_hours": renderHoursOrDash(powerHours, powerSample.Empty),
			"write_volume":   format.GetDiskSize(formatFloat(written)),
		}
		items = append(items, pkgdashboard.Item{Raw: raw, Display: disp})
	}
	return pkgdashboard.Envelope{
		Kind:  pkgdashboard.KindOverviewDiskMain,
		Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: items,
	}, nil
}

// renderDiskTemperature mirrors `renderTemperature` but only emits
// the active --temp-unit value (no dual celsius/fahrenheit display).
// Empty/zero celsius prints "-" the way the SPA does (config.ts:219).
// Private — only the disk-main row builder calls this.
func renderDiskTemperature(celsius float64, target format.TempUnit) string {
	if celsius == 0 {
		return "-"
	}
	return pkgdashboard.RenderTemperature(celsius, target)
}

// renderHoursOrDash renders power-on hours as "Xh", or "-" when the
// metric was absent (Empty=true) — distinguishes "0 hours of
// uptime" from "no reading at all".
func renderHoursOrDash(hours float64, empty bool) string {
	if empty {
		return "-"
	}
	return fmt.Sprintf("%vh", strconv.FormatFloat(hours, 'f', -1, 64))
}

// dispOrDash trims whitespace; empty falls through to "-". Mirrors
// the SPA's empty-cell rendering for SMART label fields.
func dispOrDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

// ifYesNo formats the 4K-native flag as Yes/No (the SPA's choice).
func ifYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// WriteMainTable renders the per-disk summary. Per the user's
// decision IOPS / throughput columns are dropped; everything that
// fits the SPA card lives here in column form. Column order is
// pinned: agents that scrape stdout rely on the index being stable
// across releases.
func WriteMainTable(w io.Writer, env pkgdashboard.Envelope) error {
	cols := []pkgdashboard.TableColumn{
		{Header: "DEVICE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "device") }},
		{Header: "NODE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "node") }},
		{Header: "TYPE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "type") }},
		{Header: "HEALTH", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "health") }},
		{Header: "TOTAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "total") }},
		{Header: "USED", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "used") }},
		{Header: "AVAIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "avail") }},
		{Header: "UTIL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "util") }},
		{Header: "TEMP", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "temperature") }},
		{Header: "MODEL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "model") }},
		{Header: "SERIAL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "serial") }},
		{Header: "PROTOCOL", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "protocol") }},
		{Header: "FIRMWARE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "firmware") }},
		{Header: "4K_NATIVE", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "is_4k_native") }},
		{Header: "POWER_ON", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "power_on_hours") }},
		{Header: "WRITTEN", Get: func(it pkgdashboard.Item) string { return pkgdashboard.DisplayString(it, "write_volume") }},
	}
	return pkgdashboard.WriteTable(w, cols, env.Items)
}
