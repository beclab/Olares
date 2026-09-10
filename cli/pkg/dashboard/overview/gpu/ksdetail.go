package gpu

import (
	"context"
	"fmt"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// The KubeSphere-sourced detail view (Intel, AMD) produces the same three
// sections as the HAMI one — detail / gauges / trends — so an agent can read
// either without branching on vendor. What differs is the source: HAMI
// answers PromQL, KubeSphere answers metric names and leaves the label work
// to us, which is why the panels here are described structurally in
// pkg/dashboard/gpuvendor.go rather than as query strings.

// vendorsSupportTasks reports whether any vendor present can attribute usage
// to a container.
func vendorsSupportTasks(inv pkgdashboard.GPUInventory) bool {
	for _, v := range inv.Vendors {
		spec, ok := pkgdashboard.KsSpecFor(v)
		if !ok {
			// NVIDIA reads HAMI, which is where task attribution comes from.
			return true
		}
		if spec.SupportsTasks {
			return true
		}
	}
	return false
}

// ksDetailTarget names the card a KubeSphere detail view is pinned to, once
// its uuid has been resolved to a vendor.
type ksDetailTarget struct {
	Vendor pkgdashboard.GPUVendor
	Spec   pkgdashboard.KsVendorSpec
	Row    pkgdashboard.DeviceRow
	Node   string
}

// resolveKsDevice finds which KubeSphere vendor owns `uuid`, if any. Returns
// ok=false when no such card exists, which is the signal for the caller to
// fall through to HAMI — that is where an NVIDIA uuid belongs.
//
// The lookup costs one cluster-wide device read per vendor. That is the same
// request the list view makes, and it is the only way to learn a card's
// vendor from a uuid alone: the detail endpoints are per-vendor, so we have
// to know the vendor before we can ask anything else.
func resolveKsDevice(ctx context.Context, c *pkgdashboard.Client, uuid string) (ksDetailTarget, bool) {
	inv, err := pkgdashboard.DetectGPUVendors(ctx, c)
	if err != nil {
		return ksDetailTarget{}, false
	}
	for _, vendor := range inv.Vendors {
		spec, ok := pkgdashboard.KsSpecFor(vendor)
		if !ok {
			continue
		}
		table, err := pkgdashboard.FetchGPUMetrics(ctx, c, pkgdashboard.DeviceMetricNames(spec),
			pkgdashboard.GPUMetricOptions{Level: pkgdashboard.GPULevelCluster})
		if err != nil {
			continue
		}
		row, found := pkgdashboard.DeviceRowFor(vendor, spec, table, pkgdashboard.DeviceFilter{ID: uuid})
		if !found {
			continue
		}
		return ksDetailTarget{
			Vendor: vendor,
			Spec:   spec,
			Row:    row,
			Node:   fmt.Sprintf("%v", row["deviceNode"]),
		}, true
	}
	return ksDetailTarget{}, false
}

// BuildKsDetailFullEnvelope assembles the detail / gauges / trends sections
// for one Intel or AMD card.
//
// Every panel on the page is answered by a single range request scoped to
// the card's own node: KubeSphere filters on metric name only, so asking for
// all of them at once costs no more than asking for one, and the instant
// values the gauges need are just the newest point of each series.
func BuildKsDetailFullEnvelope(
	ctx context.Context,
	c *pkgdashboard.Client,
	cf *pkgdashboard.CommonFlags,
	target ksDetailTarget,
	uuid string,
	start, end time.Time,
	since time.Duration,
) pkgdashboard.Envelope {
	now := end
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUDetailFull,
		Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
	}
	step := pkgdashboard.GPUTrendStep(start, end)
	env.Meta.Window = &pkgdashboard.TimeWindow{
		Since: humanizeSince(since),
		Start: pkgdashboard.GPUTrendTimestampISO(start, cf.Timezone.Time()),
		End:   pkgdashboard.GPUTrendTimestampISO(end, cf.Timezone.Time()),
		Step:  step,
	}

	spec := target.Spec
	panels := append(append([]pkgdashboard.KsPanelSpec{}, spec.Gauges...), spec.Lines...)
	names := pkgdashboard.PanelMetricNames(spec)
	// The node endpoint answers with that node's cards rather than every
	// card in the cluster, which keeps the payload flat as nodes are added.
	table, err := pkgdashboard.FetchGPUMetrics(ctx, c, names, pkgdashboard.GPUMetricOptions{
		Level: pkgdashboard.GPULevelNode,
		Node:  target.Node,
		Range: &pkgdashboard.GPURangeWindow{
			Start: start.Unix(),
			End:   end.Unix(),
			Step:  step,
			Times: pkgdashboard.GPUTrendTimes(start, end, step),
		},
	})
	if err != nil {
		env.Meta.Error = err.Error()
		env.Meta.ErrorKind = pkgdashboard.ClassifyTransportErr(err)
		return env
	}

	match := pkgdashboard.DeviceMatch(spec, pkgdashboard.DeviceFilter{ID: uuid, Node: target.Node})

	gaugeItems := make([]pkgdashboard.Item, 0, len(panels))
	trendItems := make([]pkgdashboard.Item, 0, len(panels))
	for _, panel := range panels {
		if !panel.HideGauge {
			gaugeItems = append(gaugeItems, ksGaugeItem(table, panel, match))
		}
		trendItems = append(trendItems, ksTrendItem(table, panel, match, cf.Timezone.Time()))
	}

	detailEnv := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewGPUDetail,
		Meta: pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
		Items: []pkgdashboard.Item{{
			Raw:     map[string]any(target.Row),
			Display: ksDetailDisplayCopy(target.Row, spec),
		}},
	}
	env.Sections = map[string]pkgdashboard.Envelope{
		"detail": detailEnv,
		"gauges": {
			Kind:  pkgdashboard.KindOverviewGPUGauges,
			Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
			Items: gaugeItems,
		},
		"trends": {
			Kind:  pkgdashboard.KindOverviewGPUTrends,
			Meta:  pkgdashboard.NewMeta(now.In(cf.Timezone.Time()), c.OlaresID(), cf.User),
			Items: trendItems,
		},
	}
	return env
}

// ksGaugeItem folds one panel's newest readings into the gauge shape the
// HAMI path already emits, so both vendors' detail output reads alike.
//
// A panel with a ready-made ratio metric uses it rather than dividing the
// two reads: the exporter's own ratio accounts for detail the separate
// numerator and denominator cannot.
func ksGaugeItem(table pkgdashboard.MetricTable, panel pkgdashboard.KsPanelSpec, match map[string][]string) pkgdashboard.Item {
	used, hasUsed := pkgdashboard.ReadInstant(table, panel.Value, match)
	total, hasTotal := 0.0, false
	if panel.Total != nil {
		total, hasTotal = pkgdashboard.ReadInstant(table, *panel.Total, match)
	} else if panel.TotalConst != 0 {
		total, hasTotal = panel.TotalConst, true
	}

	percent, hasPercent := 0.0, false
	if panel.Ratio != nil {
		percent, hasPercent = pkgdashboard.ReadInstant(table, *panel.Ratio, match)
	} else if hasUsed && hasTotal && total != 0 {
		percent, hasPercent = used/total*100, true
	}

	raw := map[string]any{
		"key":    panel.Key,
		"title":  panel.Title,
		"unit":   panel.Unit,
		"metric": panel.Value.Metric,
	}
	disp := map[string]any{"key": panel.Key, "title": panel.Title, "unit": panel.Unit}
	if hasUsed {
		raw["value"] = used
		raw["used"] = used
		disp["value"] = formatGaugeValue(used, panel.Unit)
	} else {
		disp["value"] = "—"
	}
	if hasTotal {
		raw["total"] = total
	}
	if hasPercent {
		raw["percent"] = percent
		disp["percent"] = percentDirect(percent)
	} else {
		disp["percent"] = "—"
	}
	if hasUsed && hasTotal {
		disp["used_total"] = fmt.Sprintf("%s/%s", roundedNumberString(used), roundedNumberString(total))
	} else {
		disp["used_total"] = "—"
	}
	return pkgdashboard.Item{Raw: raw, Display: disp}
}

// ksTrendItem plots one panel over the window. A panel carrying both a value
// and a denominator is plotted as a percentage so the line matches the gauge
// above it; everything else is plotted in its own unit.
func ksTrendItem(table pkgdashboard.MetricTable, panel pkgdashboard.KsPanelSpec, match map[string][]string, tz *time.Location) pkgdashboard.Item {
	var points []pkgdashboard.GPUPoint
	unit := panel.Unit
	switch {
	case panel.Ratio != nil:
		points = pkgdashboard.ReadRange(table, *panel.Ratio, match)
		unit = "%"
	case panel.Total != nil:
		points = pkgdashboard.RatioRange(
			pkgdashboard.ReadRange(table, panel.Value, match),
			pkgdashboard.ReadRange(table, *panel.Total, match),
		)
		unit = "%"
	default:
		points = pkgdashboard.ReadRange(table, panel.Value, match)
		if panel.TotalConst != 0 {
			unit = "%"
		}
	}

	pointsRaw := make([]map[string]any, 0, len(points))
	pointsDisp := make([]string, 0, len(points))
	for _, p := range points {
		ts := formatTrendTimestamp(fmt.Sprintf("%.0f", p.TS), tz)
		rounded := roundDP(p.Value, 2)
		pointsRaw = append(pointsRaw, map[string]any{
			"timestamp":    ts,
			"timestamp_ms": int64(p.TS * 1000),
			"value":        rounded,
			"value_raw":    p.Value,
		})
		pointsDisp = append(pointsDisp, fmt.Sprintf("%s=%s", ts, roundedNumberString(rounded)))
	}

	line := map[string]any{
		"label":  panel.Title,
		"metric": panel.Value.Metric,
		"points": pointsRaw,
	}
	return pkgdashboard.Item{
		Raw: map[string]any{
			"key":   panel.Key,
			"title": panel.Title,
			"unit":  unit,
			"lines": []map[string]any{line},
		},
		Display: map[string]any{
			"key":   panel.Key,
			"title": panel.Title,
			"unit":  unit,
			"lines": []map[string]any{{
				"label":  panel.Title,
				"points": joinPoints(pointsDisp),
			}},
		},
	}
}

// ksDetailDisplayCopy renders the identity card. The vendor spec's identity
// fields decide what a card can say about itself, so the copy follows them
// rather than a fixed whitelist — Intel has a PCI address and no driver
// version, AMD the reverse.
func ksDetailDisplayCopy(row pkgdashboard.DeviceRow, spec pkgdashboard.KsVendorSpec) map[string]any {
	out := map[string]any{}
	for field := range spec.Identity.Fields {
		if v, ok := row[field]; ok && v != nil && v != "" {
			out[field] = fmt.Sprintf("%v", v)
		}
	}
	out["health"] = gpuHealthLabel(row["health"])
	if reason, ok := row["healthReason"]; ok {
		out["health_reason"] = fmt.Sprintf("%v", reason)
	}
	return out
}

// joinPoints renders a trend's points as the single comma-separated cell the
// table mode prints, matching the HAMI path's display shape.
func joinPoints(points []string) string {
	out := ""
	for i, p := range points {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
