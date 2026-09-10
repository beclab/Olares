package dashboard

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// The Intel and AMD exporters publish through KubeSphere's monitoring API
// rather than HAMI, so their numbers arrive as plain Prometheus series with
// no server-side label pushdown: KubeSphere filters on the metric NAME only.
// That shapes everything below — one request carries every metric a view
// needs, and all the label selection, aggregation and per-card grouping
// happens here on the client.
//
// This file is the Go counterpart of the SPA's utils/gpuMetrics.ts. The two
// have to agree on aggregation and clamping or the CLI and the page will
// print different numbers off the same series.

// GPUMetricLevel selects the KubeSphere endpoint and the metric-name prefix
// that goes with it (`cluster_intel_hw_power_watts` against
// `node_intel_hw_power_watts`).
type GPUMetricLevel string

const (
	GPULevelCluster GPUMetricLevel = "cluster"
	GPULevelNode    GPUMetricLevel = "node"
)

// Aggregation collapses the series that survive label matching into one
// number. Which one is right is a property of the metric: power sums across
// cards, a ratio averages, and a "busiest engine" reading takes the max.
type Aggregation string

const (
	AggSum Aggregation = "sum"
	AggAvg Aggregation = "avg"
	AggMax Aggregation = "max"
)

// MetricRead is one reading off the monitoring API. `Metric` is the name
// without the level prefix, which the caller chooses. `Match` is applied
// client-side; listing several values for a label matches any one of them.
type MetricRead struct {
	Metric string
	Match  map[string][]string
	Agg    Aggregation
	Scale  float64
	// Clamp caps a reading the exporter may push a hair past its own range —
	// Intel's utilisation ratios report 1.00014 at saturation, which would
	// otherwise render as 100.01%.
	Clamp float64
}

// GPUPoint is one (timestamp, value) sample. Timestamps are unix seconds,
// matching what the monitoring API returns.
type GPUPoint struct {
	TS    float64
	Value float64
}

// MetricSeries is one Prometheus series: its labels plus its samples. An
// instant read carries a single point, a range read the whole window.
type MetricSeries struct {
	Labels map[string]string
	Points []GPUPoint
}

// MetricTable groups series by metric name with the level prefix stripped,
// so callers name metrics the same way regardless of which endpoint answered.
type MetricTable map[string][]MetricSeries

// GPURangeWindow describes a range query. Step and Times mirror the
// monitoring API's own parameters.
type GPURangeWindow struct {
	Start int64
	End   int64
	Step  string
	Times int
}

// GPUMetricOptions selects the endpoint and, for a node-level read, the one
// node worth asking about.
type GPUMetricOptions struct {
	Level GPUMetricLevel
	Node  string
	Range *GPURangeWindow
}

// DeviceFilter pins a query to one card. Both fields are optional; an empty
// filter matches every card.
type DeviceFilter struct {
	ID   string
	Node string
}

// FetchGPUMetrics issues a single monitoring request for every named metric
// and returns the series grouped by name.
//
// Every name is anchored: the server treats `metrics_filter` as a substring
// match, so an unanchored name would let a longer metric answer for a shorter
// one — `hw_frequency_hertz` would also collect `hw_frequency_limit_hertz`
// and quietly average a limit into a reading.
func FetchGPUMetrics(ctx context.Context, c *Client, metrics []string, opts GPUMetricOptions) (MetricTable, error) {
	level := opts.Level
	if level == "" {
		level = GPULevelCluster
	}
	names := make([]string, 0, len(metrics))
	seen := map[string]bool{}
	for _, m := range metrics {
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		names = append(names, string(level)+"_"+m)
	}
	if len(names) == 0 {
		return MetricTable{}, nil
	}

	q := url.Values{}
	q.Set("metrics_filter", AnchorMetricsFilter(names))
	if r := opts.Range; r != nil {
		q.Set("start", strconv.FormatInt(r.Start, 10))
		q.Set("end", strconv.FormatInt(r.End, 10))
		q.Set("step", r.Step)
		q.Set("times", strconv.Itoa(r.Times))
	}
	path := "/kapis/monitoring.kubesphere.io/v1alpha3/cluster"
	if level == GPULevelNode {
		path = "/kapis/monitoring.kubesphere.io/v1alpha3/nodes"
		if opts.Node != "" {
			q.Set("resources_filter", opts.Node+"$")
		}
	}

	results, err := DoMonitoring(ctx, c, path, q)
	if err != nil {
		return nil, err
	}
	prefix := string(level) + "_"
	table := MetricTable{}
	for name, res := range results {
		short := name
		if len(short) > len(prefix) && short[:len(prefix)] == prefix {
			short = short[len(prefix):]
		}
		table[short] = seriesOfResult(res)
	}
	return table, nil
}

// seriesOfResult flattens one metric's monitoring payload. The API returns
// range samples under `values` and instant samples under `value`; both
// collapse to the same point list here so readers don't branch.
func seriesOfResult(res format.MonitoringResult) []MetricSeries {
	out := make([]MetricSeries, 0, len(res.Data.Result))
	for _, r := range res.Data.Result {
		pairs := r.Values
		if len(pairs) == 0 && len(r.Value) > 0 {
			pairs = [][]any{r.Value}
		}
		s := MetricSeries{Labels: r.Metric}
		for _, pair := range pairs {
			if len(pair) < 2 {
				continue
			}
			ts, okTS := anyToFloat(pair[0])
			v, okV := anyToFloat(pair[1])
			if !okTS || !okV {
				continue
			}
			s.Points = append(s.Points, GPUPoint{TS: ts, Value: v})
		}
		out = append(out, s)
	}
	return out
}

// anyToFloat coerces a JSON-decoded sample field. The monitoring API is
// inconsistent about whether it sends numbers or numeric strings, and a
// non-finite value ("NaN" for a gap) has to be rejected rather than folded
// into an average.
func anyToFloat(v any) (float64, bool) {
	var f float64
	switch x := v.(type) {
	case float64:
		f = x
	case int:
		f = float64(x)
	case int64:
		f = float64(x)
	case string:
		parsed, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, false
		}
		f = parsed
	default:
		return 0, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

// labelsMatch reports whether a series satisfies every constraint. A label
// listing several values matches any one of them.
func labelsMatch(labels map[string]string, match map[string][]string) bool {
	for label, want := range match {
		got, ok := labels[label]
		if !ok {
			return false
		}
		hit := false
		for _, w := range want {
			if got == w {
				hit = true
				break
			}
		}
		if !hit {
			return false
		}
	}
	return true
}

// SelectSeries returns the series of one metric that satisfy both the read's
// own label constraints and any extra ones the caller adds (a detail view
// pinning one card, for instance).
func SelectSeries(table MetricTable, read MetricRead, extra map[string][]string) []MetricSeries {
	var out []MetricSeries
	for _, s := range table[read.Metric] {
		if labelsMatch(s.Labels, read.Match) && labelsMatch(s.Labels, extra) {
			out = append(out, s)
		}
	}
	return out
}

// combine collapses values under the read's aggregation. Returns ok=false
// for an empty input so callers can tell "no series matched" apart from a
// genuine zero — the distinction matters for a health or utilisation cell.
func combine(values []float64, agg Aggregation) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}
	switch agg {
	case AggAvg:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values)), true
	case AggMax:
		out := values[0]
		for _, v := range values[1:] {
			if v > out {
				out = v
			}
		}
		return out, true
	default:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum, true
	}
}

// scaled applies the read's unit conversion and ceiling.
func scaled(v float64, read MetricRead) float64 {
	if read.Scale != 0 {
		v *= read.Scale
	}
	if read.Clamp != 0 && v > read.Clamp {
		return read.Clamp
	}
	return v
}

// ReadInstant takes the newest point of every matching series and collapses
// them into one number. ok=false means nothing matched.
func ReadInstant(table MetricTable, read MetricRead, extra map[string][]string) (float64, bool) {
	var values []float64
	for _, s := range SelectSeries(table, read, extra) {
		if len(s.Points) == 0 {
			continue
		}
		values = append(values, s.Points[len(s.Points)-1].Value)
	}
	v, ok := combine(values, read.Agg)
	if !ok {
		return 0, false
	}
	return scaled(v, read), true
}

// ReadRange returns one aggregated point per timestamp present in the
// matching series, in ascending time order.
func ReadRange(table MetricTable, read MetricRead, extra map[string][]string) []GPUPoint {
	byTime := map[float64][]float64{}
	for _, s := range SelectSeries(table, read, extra) {
		for _, p := range s.Points {
			byTime[p.TS] = append(byTime[p.TS], p.Value)
		}
	}
	stamps := make([]float64, 0, len(byTime))
	for ts := range byTime {
		stamps = append(stamps, ts)
	}
	sort.Float64s(stamps)
	out := make([]GPUPoint, 0, len(stamps))
	for _, ts := range stamps {
		v, ok := combine(byTime[ts], read.Agg)
		if !ok {
			continue
		}
		out = append(out, GPUPoint{TS: ts, Value: scaled(v, read)})
	}
	return out
}

// RatioRange expresses one series as a percentage of another. The two
// metrics come back as separate results, so points are paired by timestamp
// and any timestamp missing either side is dropped rather than filled with a
// guess — a fabricated point would show up as a real dip in the chart.
func RatioRange(numerator, denominator []GPUPoint) []GPUPoint {
	divisor := make(map[float64]float64, len(denominator))
	for _, p := range denominator {
		divisor[p.TS] = p.Value
	}
	out := make([]GPUPoint, 0, len(numerator))
	for _, p := range numerator {
		d, ok := divisor[p.TS]
		if !ok || d == 0 {
			continue
		}
		out = append(out, GPUPoint{TS: p.TS, Value: p.Value / d * 100})
	}
	return out
}

// DeviceRow is one card's worth of table columns. Values are `any` because
// a row mixes identity strings with numeric readings and the set differs per
// vendor; the vendor spec names every key it fills.
type DeviceRow map[string]any

// deviceRowKey identifies a card across metrics. Node is part of the key
// because two nodes can report cards with the same index, and vendor because
// a node can carry cards from two vendors at once.
func deviceRowKey(vendor GPUVendor, node, id string) string {
	return fmt.Sprintf("%s|%s|%s", vendor, node, id)
}

// JoinDeviceRows builds one row per card. The identity metric supplies the
// labels that name a card; every other read is grouped back onto its card by
// the same key, which is what keeps the request count flat in the number of
// cards.
func JoinDeviceRows(vendor GPUVendor, spec KsVendorSpec, table MetricTable) []DeviceRow {
	keyOf := func(labels map[string]string) string {
		return deviceRowKey(vendor, labels[spec.DeviceLabels.Node], labels[spec.DeviceLabels.ID])
	}

	rows := map[string]DeviceRow{}
	// Insertion order is tracked explicitly: Go map iteration is randomised,
	// and a card list that reshuffles between invocations would break both
	// the table's stable column-index contract and any agent diffing output.
	var order []string
	for _, s := range table[spec.Identity.Metric] {
		row := DeviceRow{"vendor": string(vendor), "isExternal": false}
		for field, label := range spec.Identity.Fields {
			row[field] = s.Labels[label]
		}
		row["deviceId"] = s.Labels[spec.DeviceLabels.ID]
		row["deviceNode"] = s.Labels[spec.DeviceLabels.Node]
		key := keyOf(s.Labels)
		row["rowKey"] = key
		if _, seen := rows[key]; !seen {
			order = append(order, key)
		}
		rows[key] = row
	}

	for _, col := range spec.DeviceMetrics {
		grouped := map[string][]float64{}
		for _, s := range SelectSeries(table, col.Value, nil) {
			if len(s.Points) == 0 {
				continue
			}
			key := keyOf(s.Labels)
			grouped[key] = append(grouped[key], s.Points[len(s.Points)-1].Value)
		}
		for key, values := range grouped {
			row, ok := rows[key]
			if !ok {
				continue
			}
			if v, ok := combine(values, col.Value.Agg); ok {
				row[col.Field] = scaled(v, col.Value)
			}
		}
	}

	out := make([]DeviceRow, 0, len(order))
	for _, key := range order {
		row := rows[key]
		for field, derive := range spec.Derived {
			row[field] = derive(row)
		}
		row["health"] = spec.IsHealthy(row)
		if spec.HealthReason != nil {
			if reason := spec.HealthReason(row); reason != "" {
				row["healthReason"] = reason
			}
		}
		out = append(out, row)
	}
	return out
}

// DeviceRowFor picks the single row a detail view is pinned to.
func DeviceRowFor(vendor GPUVendor, spec KsVendorSpec, table MetricTable, device DeviceFilter) (DeviceRow, bool) {
	for _, row := range JoinDeviceRows(vendor, spec, table) {
		if rowString(row, "deviceId") != device.ID {
			continue
		}
		if device.Node != "" && rowString(row, "deviceNode") != device.Node {
			continue
		}
		return row, true
	}
	return nil, false
}

// DeviceMatch turns a detail view's device filter into label constraints for
// this vendor's own label names.
func DeviceMatch(spec KsVendorSpec, device DeviceFilter) map[string][]string {
	match := map[string][]string{}
	if device.Node != "" {
		match[spec.DeviceLabels.Node] = []string{device.Node}
	}
	if device.ID != "" {
		match[spec.DeviceLabels.ID] = []string{device.ID}
	}
	if len(match) == 0 {
		return nil
	}
	return match
}

// rowString reads a row field as a string, tolerating an absent key.
func rowString(row DeviceRow, field string) string {
	v, ok := row[field]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// RowFloat reads a row field as a number, returning ok=false when the column
// was never filled — an unreported metric has to stay distinguishable from a
// reported zero.
func RowFloat(row DeviceRow, field string) (float64, bool) {
	v, ok := row[field]
	if !ok || v == nil {
		return 0, false
	}
	f, ok := anyToFloat(v)
	return f, ok
}

// DeviceMetricNames lists every metric the device table needs, for a single
// batched request.
func DeviceMetricNames(spec KsVendorSpec) []string {
	names := []string{spec.Identity.Metric}
	for _, col := range spec.DeviceMetrics {
		names = append(names, col.Value.Metric)
	}
	return names
}

// PanelMetricNames lists every metric the detail panels need, likewise for
// one batched request.
func PanelMetricNames(spec KsVendorSpec) []string {
	var names []string
	for _, p := range append(append([]KsPanelSpec{}, spec.Gauges...), spec.Lines...) {
		for _, read := range p.reads() {
			names = append(names, read.Metric)
		}
	}
	return names
}

// GPUTrendTimes is how many samples a window of `step` holds. The monitoring
// API wants the sample count alongside start/end/step, and a count that
// disagrees with the window makes it resample to a different resolution than
// the caller asked for.
func GPUTrendTimes(start, end time.Time, step string) int {
	d, err := time.ParseDuration(step)
	if err != nil || d <= 0 {
		return 1
	}
	n := int(end.Sub(start) / d)
	if n < 1 {
		return 1
	}
	return n
}
