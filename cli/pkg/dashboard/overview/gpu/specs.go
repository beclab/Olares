package gpu

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// `dashboard overview gpu detail <uuid>` and
// `dashboard overview gpu task-detail <name> <pod-uid>` mirror the
// SPA's per-GPU / per-task detail pages
// (Overview2/GPU/GPUsDetails.vue and Overview2/GPU/TasksDetails.vue).
// Both pages are a three-layer cake:
//
//  1. Static info     — HAMI's /v1/gpu  or /v1/container, flat object
//  2. Top gauges      — N PromQL `instant-vector` queries
//  3. Trend charts    — N PromQL `range-vector`   queries
//
// The CLI surfaces the same data through a `sections` envelope:
//
//	{
//	  "kind": "dashboard.overview.gpu.detail.full",
//	  "meta": {
//	     "fetched_at": "...",
//	     "window":    {"since":"8h","start":"...","end":"...","step":"30m"},
//	     "warnings":  ["gauges[2] (util_core): HAMI returned HTTP 502", ...]
//	  },
//	  "sections": {
//	     "detail":  { ... HAMI's /v1/gpu  body, flat (1 item) ...        },
//	     "gauges":  { ... 6 items (GPU) / 2 items (Task) ...             },
//	     "trends":  { ... 4 items (GPU) / 2 items (Task) ...             }
//	  }
//	}
//
// Soft-failure semantics: a single instant/range query failing does
// NOT abort the envelope. The failed item carries Meta.Error and
// the parent envelope's Meta.Warnings collects a one-line summary
// so agents can branch on `len(meta.warnings) > 0` without scanning
// every section.

// ----------------------------------------------------------------------------
// Default time windows (mirrors the SPA per-page default)
// ----------------------------------------------------------------------------

const (
	GPUDetailDefaultSince  = 8 * time.Hour
	TaskDetailDefaultSince = 1 * time.Hour
)

// ----------------------------------------------------------------------------
// Query-spec types — describe a gauge / trend without actually executing.
// ----------------------------------------------------------------------------

// gaugeSpec models a single instant-vector gauge (the four GPU
// utilisation circles + the two single-value W / °C dials, or the
// two task-level utilisation dials). Field semantics mirror the
// SPA's `useInstantVector` config object:
//
//   - Query       : numerator instant query.
//   - TotalQuery  : denominator instant query (optional). When set,
//     percent = used/total*100. When empty *and* TotalLiteral != 0,
//     percent uses the literal (the SPA's `total: 100` short-circuit
//     for util_core).
//   - TotalLiteral: hard-coded total (e.g. 100 for util_core's
//     "0..100" scale). Mutually exclusive with TotalQuery.
//   - Unit        : human-readable unit appended in the table (Gi /
//     W / ℃ / "" for ratios).
type gaugeSpec struct {
	Key          string
	Title        string
	Unit         string
	Query        string
	TotalQuery   string
	TotalLiteral float64
}

// trendLine is one line in a trend chart. Multi-line trends (e.g.
// "Resource allocation trend" plots core + memory together) carry
// one trendLine per series.
type trendLine struct {
	Label string
	Query string
}

// trendSpec is one row in the `trends` section.
type trendSpec struct {
	Key   string
	Title string
	Unit  string
	Lines []trendLine
}

// ----------------------------------------------------------------------------
// SPA query catalogue (1:1 with GPUsDetails.vue / TasksDetails.vue)
// ----------------------------------------------------------------------------
//
// `$deviceuuid` is replaced with the resolved GPU UUID; `$container`,
// `$pod`, `$namespace` with the resolved task labels. We keep the
// SPA strings verbatim — wrapping `avg(sum(...) by (instance))`
// around otherwise-flat counters is intentional (HAMI's WebUI
// multiplexes instances and the SPA collapses them; CLI must do the
// same to match the rendered numbers).

func gpuDetailGaugeSpecs() []gaugeSpec {
	return []gaugeSpec{
		{
			Key:        "alloc_core",
			Title:      "Calculation power allocation ratio",
			Unit:       "",
			Query:      `avg(sum(hami_container_vcore_allocated{deviceuuid=~"$deviceuuid"}) by (instance))`,
			TotalQuery: `avg(sum(hami_core_size{deviceuuid=~"$deviceuuid"}) by (instance))`,
		},
		{
			Key:        "alloc_mem",
			Title:      "Video memory allocation ratio",
			Unit:       "Gi",
			Query:      `avg(sum(hami_container_vmemory_allocated{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024`,
			TotalQuery: `avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024`,
		},
		{
			// SPA hard-codes total: 100 (it's `hami_core_util` already
			// expressed on a 0..100 scale). totalQuery is empty.
			Key:          "util_core",
			Title:        "GPU utilization",
			Unit:         "",
			Query:        `avg(sum(hami_core_util{deviceuuid=~"$deviceuuid"}) by (instance))`,
			TotalLiteral: 100,
		},
		{
			Key:        "util_mem",
			Title:      "VRAM usage rate",
			Unit:       "Gi",
			Query:      `avg(sum(hami_memory_used{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024`,
			TotalQuery: `avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance))/1024`,
		},
		{
			Key:   "power",
			Title: "GPU power",
			Unit:  "W",
			Query: `avg by (device_no,driver_version) (hami_device_power{deviceuuid=~"$deviceuuid"})`,
		},
		{
			Key:   "temperature",
			Title: "GPU temperature",
			Unit:  "℃",
			Query: `avg(sum(hami_device_temperature{deviceuuid=~"$deviceuuid"}) by (instance))`,
		},
	}
}

func gpuDetailTrendSpecs() []trendSpec {
	const allocCorePct = `avg(sum(hami_container_vcore_allocated{deviceuuid=~"$deviceuuid"}) by (instance))/avg(sum(hami_core_size{deviceuuid=~"$deviceuuid"}) by (instance)) *100`
	const allocMemPct = `(avg(sum(hami_container_vmemory_allocated{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024 )/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024) *100 `
	const utilCorePct = `avg(sum(hami_core_util_avg{deviceuuid=~"$deviceuuid"}) by (instance))`
	const utilMemPct = `(avg(sum(hami_memory_used{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024)/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance))/1024)*100`
	return []trendSpec{
		{
			Key:   "alloc_trend",
			Title: "Resource allocation trend",
			Unit:  "%",
			// Labels mirror the SPA's legend
			// (`legend: [t('GPU_OP.VRAM'), t('GPU_OP.GPU_MEMORY')]`
			// in GPUsDetails.vue:223). Note: despite the names, the
			// FIRST line is fed by `hami_core_*` (compute-power
			// allocation %) and the SECOND by `hami_memory_*` (VRAM
			// allocation %) — that's the SPA's chosen wording, and
			// the CLI must reproduce it 1:1 so agents/users see the
			// same legend tokens across surfaces. Don't "fix" by
			// reverting to `core`/`memory` without a coordinated SPA
			// change.
			Lines: []trendLine{
				{Label: "VRAM", Query: allocCorePct},
				{Label: "GPU_MEMORY", Query: allocMemPct},
			},
		},
		{
			Key:   "usage_trend",
			Title: "Resource usage trend",
			Unit:  "%",
			Lines: []trendLine{
				{Label: "VRAM", Query: utilCorePct},
				{Label: "GPU_MEMORY", Query: utilMemPct},
			},
		},
		{
			Key:   "power_trend",
			Title: "GPU power",
			Unit:  "W",
			Lines: []trendLine{{
				Label: "power",
				Query: `avg by (device_no,driver_version) (hami_device_power{deviceuuid=~"$deviceuuid"})`,
			}},
		},
		{
			Key:   "temp_trend",
			Title: "GPU temperature",
			Unit:  "℃",
			Lines: []trendLine{{
				Label: "temperature",
				Query: `avg(sum(hami_device_temperature{deviceuuid=~"$deviceuuid"}) by (instance))`,
			}},
		},
	}
}

// taskDetailGaugeSpecs / taskDetailTrendSpecs use SPA's
// `$container` / `$pod` / `$namespace` placeholders. The
// allocationGaugesIncluded boolean toggles the two allocation
// gauges (sharemode != "2" / TimeSlicing in the SPA).
func taskDetailGaugeSpecs(allocationGaugesIncluded bool) []gaugeSpec {
	specs := []gaugeSpec{
		{
			Key:        "compute_usage",
			Title:      "Compute usage rate",
			Unit:       "%",
			Query:      `avg(sum(hami_container_core_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
			TotalQuery: `avg(sum(hami_container_vcore_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
		},
		{
			Key:        "vram_usage",
			Title:      "VRAM usage rate",
			Unit:       "GiB",
			Query:      `avg(sum(hami_container_memory_used{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))/ 1024`,
			TotalQuery: `avg(sum(hami_container_vmemory_allocated{container_name="$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))/1024`,
		},
	}
	if !allocationGaugesIncluded {
		// SPA uses these gauges as the "Allocatable …" labels in
		// `displayAllocation`. When sharemode==TimeSlicing the SPA
		// drops the allocation gauges; the usage gauges still
		// render. We mirror that by **adding** allocation rows only
		// when the caller asked for them (currently the SPA never
		// does for time-slicing tasks). Today both branches fall
		// through to the same usage gauges; this hook keeps the
		// future allocation rows gated by the same flag for
		// plan-symmetry.
		_ = specs
	}
	return specs
}

func taskDetailTrendSpecs() []trendSpec {
	return []trendSpec{
		{
			Key:   "compute_trend",
			Title: "Compute power usage trend",
			Unit:  "%",
			Lines: []trendLine{{
				Label: "usage",
				Query: `avg(sum(hami_container_core_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
			}},
		},
		{
			Key:   "vram_trend",
			Title: "VRAM usage trend",
			Unit:  "%",
			Lines: []trendLine{{
				Label: "usage",
				Query: `avg(sum(hami_container_memory_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
			}},
		},
	}
}

// ----------------------------------------------------------------------------
// Per-query runners
// ----------------------------------------------------------------------------

// runGauge executes one instant + (optional) total query and folds
// the result into a single Item. Soft failure: any HTTP/decode
// error is stored in the item's `raw.error` / `display.error` AND
// surfaced via `addWarning` so the parent envelope can list it
// under `meta.warnings`.
func runGauge(ctx context.Context, c *pkgdashboard.Client, spec gaugeSpec, repl *strings.Replacer, addWarning func(string)) pkgdashboard.Item {
	q := repl.Replace(spec.Query)
	totalQ := repl.Replace(spec.TotalQuery)

	var (
		used    float64
		total   = spec.TotalLiteral
		percent float64
		valErr  error
		totErr  error
	)
	if q != "" {
		samples, err := pkgdashboard.FetchInstantVector(ctx, c, q)
		if err != nil {
			valErr = err
		} else if len(samples) > 0 {
			used = samples[0].Value
		}
	}
	if totalQ != "" {
		samples, err := pkgdashboard.FetchInstantVector(ctx, c, totalQ)
		if err != nil {
			totErr = err
		} else if len(samples) > 0 {
			total = samples[0].Value
		}
	}
	if total != 0 {
		percent = used / total * 100
	}

	raw := map[string]any{
		"key":     spec.Key,
		"title":   spec.Title,
		"unit":    spec.Unit,
		"query":   q,
		"value":   used,
		"used":    used,
		"total":   total,
		"percent": percent,
	}
	disp := map[string]any{
		"key":   spec.Key,
		"title": spec.Title,
		"unit":  spec.Unit,
	}
	// Display formatting mirrors GPUsDetails.vue / TasksDetails.vue:
	//   • numbers go through lodash.round(value, 2) — `roundDP(2)` here
	//   • the unit ("Gi" / "W" / "℃" / "%") is rendered next to the
	//     number by the SPA's <MyGaugeChart unit=…> prop. We bake it
	//     directly into `value` and `used_total` so the CLI's table
	//     row reads "23.89 Gi" instead of two columns the agent has
	//     to concat ("VALUE 23.889648 / UNIT Gi").
	// `Raw` keeps the un-rounded float so JSON consumers needing
	// arbitrary precision still get it.
	if total > 0 && (spec.TotalQuery != "" || spec.TotalLiteral > 0) {
		disp["percent"] = percentDirect(percent)
		disp["used_total"] = fmt.Sprintf("%s/%s", roundedNumberString(used), roundedNumberString(total))
	} else {
		disp["percent"] = "—"
		disp["used_total"] = "—"
	}
	disp["value"] = formatGaugeValue(used, spec.Unit)
	if valErr != nil || totErr != nil {
		err := valErr
		if err == nil {
			err = totErr
		}
		raw["error"] = err.Error()
		disp["error"] = err.Error()
		addWarning(fmt.Sprintf("gauges %q: %v", spec.Key, err))
	}
	return pkgdashboard.Item{Raw: raw, Display: disp}
}

// runTrend executes the range query for every line in a trend spec.
// Returns the assembled Item plus the metric labels of the FIRST
// non-empty response (used to populate device_no / driver_version
// on the GPU detail page).
//
// Display formatting (per-point) mirrors the SPA chart axis:
//   - timestamp is HAMI's epoch (seconds with optional sub-second
//     decimals OR raw milliseconds — both shapes observed in the
//     wild from /v1/monitor/query/range-vector). The CLI parses
//     either via `formatTrendTimestamp` and emits the SPA's
//     `YYYY-MM-DD HH:mm:ss` in the caller's --timezone. The raw
//     epoch milliseconds is preserved on `Raw.points[i].timestamp_ms`
//     so JSON consumers needing the wire shape still get it.
//   - value goes through lodash.round(value, 2) — same call the
//     SPA's config.ts:84 does before handing the array to ECharts.
//     The full-precision float is preserved on
//     `Raw.points[i].value_raw` for agents that need it.
func runTrend(ctx context.Context, c *pkgdashboard.Client, spec trendSpec, repl *strings.Replacer, start, end, step string, tz *time.Location, addWarning func(string)) (pkgdashboard.Item, map[string]string) {
	lineRaws := make([]map[string]any, 0, len(spec.Lines))
	lineDisps := make([]map[string]any, 0, len(spec.Lines))
	var firstLabels map[string]string
	var firstErr error
	for _, line := range spec.Lines {
		q := repl.Replace(line.Query)
		series, err := pkgdashboard.FetchRangeVector(ctx, c, q, start, end, step)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			lineRaws = append(lineRaws, map[string]any{
				"label":  line.Label,
				"query":  q,
				"error":  err.Error(),
				"points": []any{},
			})
			lineDisps = append(lineDisps, map[string]any{
				"label": line.Label,
				"error": err.Error(),
			})
			continue
		}
		var (
			pointsRaw  []map[string]any
			pointsDisp []string
			labels     map[string]string
		)
		if len(series) > 0 {
			labels = series[0].Metric
			for _, p := range series[0].Values {
				pf := toFloat(p.Value)
				rounded := roundDP(pf, 2)
				tsHuman := formatTrendTimestamp(p.Timestamp, tz)
				tsMS := trendTimestampMillis(p.Timestamp)
				pointsRaw = append(pointsRaw, map[string]any{
					"timestamp":    tsHuman,
					"timestamp_ms": tsMS,
					"value":        rounded,
					"value_raw":    pf,
				})
				pointsDisp = append(pointsDisp, fmt.Sprintf("%s=%s", tsHuman, roundedNumberString(rounded)))
			}
		}
		if firstLabels == nil && labels != nil {
			firstLabels = labels
		}
		lineRaws = append(lineRaws, map[string]any{
			"label":  line.Label,
			"query":  q,
			"labels": labels,
			"points": pointsRaw,
		})
		lineDisps = append(lineDisps, map[string]any{
			"label":  line.Label,
			"points": strings.Join(pointsDisp, ", "),
		})
	}

	raw := map[string]any{
		"key":   spec.Key,
		"title": spec.Title,
		"unit":  spec.Unit,
		"lines": lineRaws,
	}
	disp := map[string]any{
		"key":   spec.Key,
		"title": spec.Title,
		"unit":  spec.Unit,
		"lines": lineDisps,
	}
	if firstErr != nil {
		raw["error"] = firstErr.Error()
		disp["error"] = firstErr.Error()
		addWarning(fmt.Sprintf("trends %q: %v", spec.Key, firstErr))
	}
	return pkgdashboard.Item{Raw: raw, Display: disp}, firstLabels
}

// fanoutGaugeAndTrend runs every gaugeSpec + trendSpec concurrently.
// captureLabels (when non-nil) is consulted on each completed trend;
// a non-empty labels map for the "power_trend" key (the GPU page's
// driver-version source) is captured under labelMu so the caller
// can merge it back into the detail item. Errors are swallowed by
// per-runner soft-failure; the addWarning closure threads warnings
// to the parent envelope.
func fanoutGaugeAndTrend(
	ctx context.Context, c *pkgdashboard.Client,
	gaugeSpecs []gaugeSpec, trendSpecs []trendSpec,
	repl *strings.Replacer, start, end, step string,
	tz *time.Location,
	addWarning func(string),
	captureLabelsForTrendKey string,
) ([]pkgdashboard.Item, []pkgdashboard.Item, map[string]string) {
	if tz == nil {
		// Defensive: callers (BuildDetailFullEnvelope /
		// BuildTaskDetailFullEnvelope) always pass cf.Timezone, but a
		// nil here would make formatTrendTimestamp return UTC strings
		// silently — we'd rather use the local zone (matches
		// CommonFlags.Validate's empty-flag default).
		tz = time.Local
	}
	gaugeItems := make([]pkgdashboard.Item, len(gaugeSpecs))
	trendItems := make([]pkgdashboard.Item, len(trendSpecs))
	var (
		labelsCaptured map[string]string
		labelMu        sync.Mutex
	)
	g, gctx := errgroup.WithContext(ctx)
	for i, spec := range gaugeSpecs {
		i, spec := i, spec
		g.Go(func() error {
			gaugeItems[i] = runGauge(gctx, c, spec, repl, addWarning)
			return nil
		})
	}
	for i, spec := range trendSpecs {
		i, spec := i, spec
		g.Go(func() error {
			item, labels := runTrend(gctx, c, spec, repl, start, end, step, tz, addWarning)
			trendItems[i] = item
			if captureLabelsForTrendKey != "" && spec.Key == captureLabelsForTrendKey && len(labels) > 0 {
				labelMu.Lock()
				labelsCaptured = labels
				labelMu.Unlock()
			}
			return nil
		})
	}
	_ = g.Wait()
	return gaugeItems, trendItems, labelsCaptured
}

// ----------------------------------------------------------------------------
// Display helpers
// ----------------------------------------------------------------------------

// gpuDetailDisplayCopy converts HAMI's flat /v1/gpu body into a
// human-friendly key/value map for the Detail card. The whitelist
// is the SPA's `columns` array verbatim
// (GPUsDetails.vue:112-137) — six fields, in this order:
//
//	health / uuid / nodeName / type / device_no / driver_version
//
// Other HAMI fields (memoryTotal/Used, power, powerLimit,
// temperature, coreTotal/Used, vgpuTotal/Used, mode, shareMode,
// nodeUid) are deliberately NOT in `Display` — the SPA covers
// those numbers via the gauges + trend lineTools cards, and
// duplicating them here led to:
//
//   - misleading zeros: HAMI's flat /v1/gpu returns
//     `memoryUsed: 0 / power: 0 / temperature: 0` as placeholders
//     because the LIVE values come from instant-vector queries.
//     Surfacing those zeros under "Detail" made the user think the
//     GPU was reporting zero everything.
//   - unit clutter: `memoryTotal: 24463` (raw MiB) without a unit
//     suffix; the SPA never shows this number in the detail card,
//     it shows `23.89 Gi` inside a gauge, so adding a "MiB" suffix
//     in CLI would still mismatch the SPA value.
//
// Agents that need the full HAMI body still get it via `Raw` (the
// envelope's source-of-truth field). Pinned by
// TestGpuDetailDisplayCopy_SPAFieldWhitelist.
func gpuDetailDisplayCopy(d map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range []string{
		"uuid", "nodeName", "type", "device_no", "driver_version",
	} {
		if v, ok := d[k]; ok {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	if v, ok := d["health"]; ok {
		out["health"] = gpuHealthLabel(v)
	}
	return out
}

// gpuTaskDetailDisplayCopy mirrors `TasksDetails.vue:134-174` —
// eight fields (six unconditional + two `displayAllocation`-gated)
// in this order:
//
//	status / deviceIds / nodeName / type /
//	(allocatedCores) / (allocatedMem) / appName / createTime
//
// The two allocation rows are emitted only when present in the
// HAMI body (HAMI omits them for time-slicing tasks); the SPA
// hides the same rows via `displayAllocation = sharemode !==
// TimeSlicing`. Other HAMI fields (podUid, namespace, nodeUid,
// startTime/endTime, flavor, priority, …) live on `Raw` for
// agents but are NOT in `Display`. Pinned by
// TestGpuTaskDetailDisplayCopy_SPAFieldWhitelist.
func gpuTaskDetailDisplayCopy(d map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range []string{"status", "nodeName", "type", "appName", "createTime"} {
		if v, ok := d[k]; ok {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	if v, ok := d["deviceIds"]; ok {
		switch arr := v.(type) {
		case []any:
			parts := make([]string, 0, len(arr))
			for _, x := range arr {
				parts = append(parts, fmt.Sprintf("%v", x))
			}
			out["deviceIds"] = strings.Join(parts, ",")
		case []string:
			out["deviceIds"] = strings.Join(arr, ",")
		default:
			out["deviceIds"] = fmt.Sprintf("%v", v)
		}
	}
	for _, k := range []string{"allocatedCores", "allocatedMem"} {
		if v, ok := d[k]; ok {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	return out
}

// roundDP applies the same lodash.round(value, dp) the SPA uses
// before handing trend points to ECharts (config.ts:84) and gauge
// values to MyGaugeChart (GPUsDetails.vue:33-49). Banker's rounding
// is intentionally NOT used — the SPA's lodash.round is half-away-
// from-zero, and matching that is the whole point of this helper.
func roundDP(v float64, dp int) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return v
	}
	scale := math.Pow(10, float64(dp))
	return math.Round(v*scale) / scale
}

// roundedNumberString renders a float at 2dp the way SPA's
// `String(round(x, 2))` would: trailing zeros stripped, decimal
// point dropped when integer-valued. "23.89" / "100" / "0.29" — not
// "23.890000000000004" / "100.00" / "0.290000".
func roundedNumberString(v float64) string {
	r := roundDP(v, 2)
	if math.IsNaN(r) {
		return "NaN"
	}
	if math.IsInf(r, 0) {
		return "Inf"
	}
	// %g with 6-digit precision after rounding gives the SPA-shaped
	// output for everything in HAMI's observed range. strconv with
	// 'f' / -1 also works but emits "23.89" → "23.89" while %g
	// keeps small / large values readable; we pick %g for parity
	// with the pre-refactor printf path.
	return strconv.FormatFloat(r, 'f', -1, 64)
}

// formatGaugeValue renders a gauge's display VALUE column. Mirrors
// the SPA's `<MyGaugeChart unit=…>` prop: for unit-bearing gauges
// (Gi, W, ℃) the unit is appended to the number; unit-less gauges
// (the four "%" / ratio gauges that pass `unit: ' '`) emit just the
// number. The rounded number always uses roundedNumberString so
// trailing zeros / float noise are stripped.
func formatGaugeValue(v float64, unit string) string {
	num := roundedNumberString(v)
	switch strings.TrimSpace(unit) {
	case "", "-":
		return num
	default:
		return num + " " + strings.TrimSpace(unit)
	}
}

// formatTrendTimestamp converts HAMI's per-point timestamp into the
// SPA's `YYYY-MM-DD HH:mm:ss` shape (utils/gpu.ts::timeParse). HAMI
// is observed to return either:
//   - a 13-character integer string ("1779636713000") = epoch ms
//   - a 10-character integer string ("1779636713")    = epoch s
//   - a float string                ("1779636713.5") = epoch s
//     with sub-second decimals (Prometheus's wire shape)
//
// The detection runs in float64 — a 13-digit integer fits exactly,
// and any value > 1e12 gets demoted from "seconds" to "milliseconds"
// because epoch-seconds in 2026 sit at ~1.78e9. Empty / unparsable
// inputs fall through to the raw string so the user can debug what
// HAMI actually sent (better than silently rendering "1970-01-01").
func formatTrendTimestamp(raw string, tz *time.Location) string {
	if raw == "" {
		return "-"
	}
	if tz == nil {
		tz = time.Local
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw
	}
	var sec int64
	var nsec int64
	if f >= 1e12 {
		// epoch milliseconds.
		ms := int64(math.Round(f))
		sec = ms / 1000
		nsec = (ms % 1000) * int64(time.Millisecond)
	} else {
		sec = int64(f)
		nsec = int64((f - float64(sec)) * float64(time.Second))
	}
	return time.Unix(sec, nsec).In(tz).Format("2006-01-02 15:04:05")
}

// trendTimestampMillis preserves the wire-shape epoch-ms integer on
// the Raw point so JSON consumers needing arbitrary-precision time
// math (or chart libraries that re-derive from epoch) can still
// round-trip without re-parsing the human ISO string. Returns 0 on
// unparsable input — matches HAMI's "no value" sentinel for ints
// and avoids surfacing a fake epoch.
func trendTimestampMillis(raw string) int64 {
	if raw == "" {
		return 0
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0
	}
	if f >= 1e12 {
		return int64(math.Round(f))
	}
	return int64(math.Round(f * 1000))
}

// humanizeSince renders a duration the way the SPA's window picker
// does: "8h" / "1h" / "30m" — empty when zero so JSON `omitempty`
// keeps the meta block tidy.
func humanizeSince(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return d.String()
}

// ----------------------------------------------------------------------------
// Window resolution
// ----------------------------------------------------------------------------

// ResolveDetailWindow picks (start, end, since) based on the user's
// --since / --start / --end flags. Mirrors the SPA's default
// `start = now - 8h` for GPU detail and `now - 1h` for task detail
// (passed in as `def`). Watch mode always slides the window with
// each tick (since-derived end = now).
func ResolveDetailWindow(cf *pkgdashboard.CommonFlags, now time.Time, def time.Duration) (start, end time.Time, since time.Duration) {
	if !cf.Start.IsZero() && !cf.End.IsZero() {
		return cf.Start, cf.End, 0
	}
	if cf.Since > 0 {
		return now.Add(-cf.Since), now, cf.Since
	}
	return now.Add(-def), now, def
}
