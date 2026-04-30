package gpu

import (
	"context"
	"fmt"
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
			Lines: []trendLine{
				{Label: "core", Query: allocCorePct},
				{Label: "memory", Query: allocMemPct},
			},
		},
		{
			Key:   "usage_trend",
			Title: "Resource usage trend",
			Unit:  "%",
			Lines: []trendLine{
				{Label: "core", Query: utilCorePct},
				{Label: "memory", Query: utilMemPct},
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
	if total > 0 && (spec.TotalQuery != "" || spec.TotalLiteral > 0) {
		disp["percent"] = percentDirect(percent)
		disp["used_total"] = fmt.Sprintf("%g/%g", used, total)
	} else {
		disp["percent"] = "—"
		disp["used_total"] = "—"
	}
	disp["value"] = formatFloat(used)
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
func runTrend(ctx context.Context, c *pkgdashboard.Client, spec trendSpec, repl *strings.Replacer, start, end, step string, addWarning func(string)) (pkgdashboard.Item, map[string]string) {
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
				pointsRaw = append(pointsRaw, map[string]any{
					"timestamp": p.Timestamp,
					"value":     pf,
				})
				pointsDisp = append(pointsDisp, fmt.Sprintf("%s=%g", p.Timestamp, pf))
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
	addWarning func(string),
	captureLabelsForTrendKey string,
) ([]pkgdashboard.Item, []pkgdashboard.Item, map[string]string) {
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
			item, labels := runTrend(gctx, c, spec, repl, start, end, step, addWarning)
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
// human-friendly key/value map. Only the SPA-rendered fields end up
// here; the full HAMI document is preserved on `Raw` for agents.
func gpuDetailDisplayCopy(d map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range []string{
		"uuid", "type", "nodeName", "nodeUid", "shareMode",
		"vgpuUsed", "vgpuTotal", "coreUsed", "coreTotal",
		"memoryUsed", "memoryTotal",
		"power", "powerLimit", "temperature",
		"device_no", "driver_version",
	} {
		if v, ok := d[k]; ok {
			out[k] = fmt.Sprintf("%v", v)
		}
	}
	if v, ok := d["health"]; ok {
		out["health"] = gpuHealthLabel(v)
	}
	if v, ok := d["shareMode"]; ok {
		out["mode"] = gpuModeLabel(v)
	}
	return out
}

func gpuTaskDetailDisplayCopy(d map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range []string{
		"name", "status", "podUid", "nodeName", "nodeUid",
		"type", "appName", "namespace", "createTime", "startTime", "endTime",
		"allocatedDevices", "allocatedCores", "allocatedMem",
		"resourcePool", "flavor", "priority",
	} {
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
		}
	}
	if v, ok := d["deviceShareModes"]; ok {
		first := firstAnyInArray(v)
		out["mode"] = gpuModeLabel(first)
	}
	return out
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
