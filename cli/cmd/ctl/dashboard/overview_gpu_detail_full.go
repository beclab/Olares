package dashboard

// `dashboard overview gpu detail <uuid>` and
// `dashboard overview gpu task-detail <name> <pod-uid>` mirror the SPA's
// per-GPU / per-task detail pages (Overview2/GPU/GPUsDetails.vue and
// Overview2/GPU/TasksDetails.vue). Both pages are a three-layer cake:
//
//	1. Static info     — HAMI's /v1/gpu  or /v1/container, flat object
//	2. Top gauges      — N PromQL `instant-vector` queries
//	3. Trend charts    — N PromQL `range-vector`   queries
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
// Soft-failure semantics: a single instant/range query failing does NOT
// abort the envelope. The failed item carries Meta.Error and the parent
// envelope's Meta.Warnings collects a one-line summary so agents can
// branch on `len(meta.warnings) > 0` without scanning every section.

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ----------------------------------------------------------------------------
// Default time windows (mirrors the SPA per-page default)
// ----------------------------------------------------------------------------

const (
	gpuDetailDefaultSince  = 8 * time.Hour
	taskDetailDefaultSince = 1 * time.Hour
)

// ----------------------------------------------------------------------------
// Query-spec types — describe a gauge / trend without actually executing.
// ----------------------------------------------------------------------------

// gpuGaugeSpec models a single instant-vector gauge (the four GPU
// utilisation circles + the two single-value W / °C dials, or the two
// task-level utilisation dials). Field semantics mirror the SPA's
// `useInstantVector` config object:
//
//   - Query       : numerator instant query.
//   - TotalQuery  : denominator instant query (optional). When set,
//     percent = used/total*100. When empty *and* TotalLiteral != 0,
//     percent uses the literal (the SPA's `total: 100` short-circuit
//     for util_core).
//   - TotalLiteral: hard-coded total (e.g. 100 for util_core's "0..100"
//     scale). Mutually exclusive with TotalQuery.
//   - Unit        : human-readable unit appended in the table (Gi / W /
//     ℃ / "" for ratios).
type gpuGaugeSpec struct {
	Key          string
	Title        string
	Unit         string
	Query        string
	TotalQuery   string
	TotalLiteral float64
}

// gpuTrendLine is one line in a trend chart. Multi-line trends (e.g.
// "Resource allocation trend" plots core + memory together) carry one
// gpuTrendLine per series.
type gpuTrendLine struct {
	Label string
	Query string
}

// gpuTrendSpec is one row in the `trends` section.
type gpuTrendSpec struct {
	Key   string
	Title string
	Unit  string
	Lines []gpuTrendLine
}

// ----------------------------------------------------------------------------
// SPA query catalogue (1:1 with GPUsDetails.vue / TasksDetails.vue)
// ----------------------------------------------------------------------------
//
// `$deviceuuid` is replaced with the resolved GPU UUID; `$container`,
// `$pod`, `$namespace` with the resolved task labels. We keep the SPA
// strings verbatim — wrapping `avg(sum(...) by (instance))` around
// otherwise-flat counters is intentional (HAMI's WebUI multiplexes
// instances and the SPA collapses them; CLI must do the same to match
// the rendered numbers).

func gpuDetailGaugeSpecs() []gpuGaugeSpec {
	return []gpuGaugeSpec{
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

func gpuDetailTrendSpecs() []gpuTrendSpec {
	const allocCorePct = `avg(sum(hami_container_vcore_allocated{deviceuuid=~"$deviceuuid"}) by (instance))/avg(sum(hami_core_size{deviceuuid=~"$deviceuuid"}) by (instance)) *100`
	const allocMemPct = `(avg(sum(hami_container_vmemory_allocated{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024 )/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024) *100 `
	const utilCorePct = `avg(sum(hami_core_util_avg{deviceuuid=~"$deviceuuid"}) by (instance))`
	const utilMemPct = `(avg(sum(hami_memory_used{deviceuuid=~"$deviceuuid"}) by (instance)) / 1024)/(avg(sum(hami_memory_size{deviceuuid=~"$deviceuuid"}) by (instance))/1024)*100`
	return []gpuTrendSpec{
		{
			Key:   "alloc_trend",
			Title: "Resource allocation trend",
			Unit:  "%",
			Lines: []gpuTrendLine{
				{Label: "core", Query: allocCorePct},
				{Label: "memory", Query: allocMemPct},
			},
		},
		{
			Key:   "usage_trend",
			Title: "Resource usage trend",
			Unit:  "%",
			Lines: []gpuTrendLine{
				{Label: "core", Query: utilCorePct},
				{Label: "memory", Query: utilMemPct},
			},
		},
		{
			Key:   "power_trend",
			Title: "GPU power",
			Unit:  "W",
			Lines: []gpuTrendLine{{
				Label: "power",
				Query: `avg by (device_no,driver_version) (hami_device_power{deviceuuid=~"$deviceuuid"})`,
			}},
		},
		{
			Key:   "temp_trend",
			Title: "GPU temperature",
			Unit:  "℃",
			Lines: []gpuTrendLine{{
				Label: "temperature",
				Query: `avg(sum(hami_device_temperature{deviceuuid=~"$deviceuuid"}) by (instance))`,
			}},
		},
	}
}

// taskDetailGaugeSpecs / taskDetailTrendSpecs use SPA's
// `$container` / `$pod` / `$namespace` placeholders. The
// allocationGaugesIncluded boolean toggles the two allocation gauges
// (sharemode != "2" / TimeSlicing in the SPA).
func taskDetailGaugeSpecs(allocationGaugesIncluded bool) []gpuGaugeSpec {
	specs := []gpuGaugeSpec{
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
		// drops the allocation gauges; the usage gauges still render.
		// We mirror that by **adding** allocation rows only when the
		// caller asked for them (currently the SPA never does for time-
		// slicing tasks). Today both branches fall through to the same
		// usage gauges; this hook keeps the future allocation rows
		// gated by the same flag for plan-symmetry.
		_ = specs
	}
	return specs
}

func taskDetailTrendSpecs() []gpuTrendSpec {
	return []gpuTrendSpec{
		{
			Key:   "compute_trend",
			Title: "Compute power usage trend",
			Unit:  "%",
			Lines: []gpuTrendLine{{
				Label: "usage",
				Query: `avg(sum(hami_container_core_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
			}},
		},
		{
			Key:   "vram_trend",
			Title: "VRAM usage trend",
			Unit:  "%",
			Lines: []gpuTrendLine{{
				Label: "usage",
				Query: `avg(sum(hami_container_memory_util{container_name=~"$container",pod_name=~"$pod",namespace_name="$namespace"}) by (instance))`,
			}},
		},
	}
}

// ----------------------------------------------------------------------------
// Builders — kick off all queries, fold results into items.
// ----------------------------------------------------------------------------

// buildGPUDetailFullEnvelope produces the full GPU detail sections
// envelope for `dashboard overview gpu detail <uuid>`.
//
// Flow:
//
//  1. Sequentially call HAMI /v1/gpu (basic info). 404/5xx short-circuit
//     to the standard `no_vgpu_integration` / `vgpu_unavailable` empty
//     envelopes via `vgpuUnavailableFromError`.
//  2. Concurrently fan out the gauge + trend queries with errgroup. A
//     per-query failure is captured into the gauge/trend item itself
//     (raw.error / display.error) and added to env.Meta.Warnings; it
//     does NOT abort sibling queries.
//  3. Extract device_no / driver_version from the *power* range query's
//     metric labels (the SPA does the same) and merge them into the
//     detail item so `dashboard overview gpu detail <uuid>` can render
//     the SPA's "Driver version: 590.44.01" cell.
func buildGPUDetailFullEnvelope(ctx context.Context, c *Client, uuid string, start, end time.Time, since time.Duration) (Envelope, error) {
	now := end
	advisoryNote, _ := gpuAdvisory(ctx, c)
	env := Envelope{
		Kind: KindOverviewGPUDetailFull,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	env.Meta.Window = &TimeWindow{
		Since: humanizeSince(since),
		Start: gpuTrendTimestampISO(start),
		End:   gpuTrendTimestampISO(end),
		Step:  gpuTrendStep(start, end),
	}

	// Step 1 — flat detail.
	detail, err := fetchGraphicsDetail(ctx, c, uuid)
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUDetailFull, now); ok {
			if env.Meta.Note != "" {
				unavail.Meta.Note = env.Meta.Note + " | " + unavail.Meta.Note
			}
			unavail.Meta.Window = env.Meta.Window
			return unavail, nil
		}
		return env, err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		return env, nil
	}

	// Step 2/3 — fan-out queries.
	gaugeSpecs := gpuDetailGaugeSpecs()
	trendSpecs := gpuDetailTrendSpecs()
	step := env.Meta.Window.Step
	startISO := env.Meta.Window.Start
	endISO := env.Meta.Window.End
	deviceuuidReplacer := strings.NewReplacer("$deviceuuid", uuid)

	gaugeItems := make([]Item, len(gaugeSpecs))
	trendItems := make([]Item, len(trendSpecs))
	var (
		warnMu       sync.Mutex
		warnings     []string
		labelsFromPw map[string]string // device_no / driver_version from power range query
		labelMu      sync.Mutex
	)
	addWarning := func(msg string) {
		warnMu.Lock()
		warnings = append(warnings, msg)
		warnMu.Unlock()
	}

	g, gctx := errgroup.WithContext(ctx)
	for i, spec := range gaugeSpecs {
		i, spec := i, spec
		g.Go(func() error {
			gaugeItems[i] = runGPUGauge(gctx, c, spec, deviceuuidReplacer, addWarning)
			return nil
		})
	}
	for i, spec := range trendSpecs {
		i, spec := i, spec
		g.Go(func() error {
			item, labels := runGPUTrend(gctx, c, spec, deviceuuidReplacer, startISO, endISO, step, addWarning)
			trendItems[i] = item
			if spec.Key == "power_trend" && len(labels) > 0 {
				labelMu.Lock()
				labelsFromPw = labels
				labelMu.Unlock()
			}
			return nil
		})
	}
	_ = g.Wait() // each goroutine swallows query errors → never returns one

	// Merge device_no / driver_version into detail (SPA's `detail2`).
	for k, v := range labelsFromPw {
		if k == "device_no" || k == "driver_version" {
			if _, present := detail[k]; !present {
				detail[k] = v
			}
		}
	}

	// Compose sections.
	detailEnv := Envelope{
		Kind:  KindOverviewGPUDetail,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: []Item{{Raw: detail, Display: gpuDetailDisplayCopy(detail)}},
	}
	gaugesEnv := Envelope{
		Kind:  KindOverviewGPUGauges,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: gaugeItems,
	}
	trendsEnv := Envelope{
		Kind:  KindOverviewGPUTrends,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: trendItems,
	}
	env.Sections = map[string]Envelope{
		"detail": detailEnv,
		"gauges": gaugesEnv,
		"trends": trendsEnv,
	}
	if len(warnings) > 0 {
		env.Meta.Warnings = warnings
	}
	return env, nil
}

// buildGPUTaskDetailFullEnvelope is the task-flavoured twin of
// buildGPUDetailFullEnvelope. The main difference is the placeholder
// substitution: `$container` / `$pod` / `$namespace` are pulled from
// the task detail itself (the SPA does the same — it can't fan out the
// monitor queries until /v1/container resolves).
func buildGPUTaskDetailFullEnvelope(ctx context.Context, c *Client, name, podUID, sharemode string, start, end time.Time, since time.Duration) (Envelope, error) {
	now := end
	advisoryNote, _ := gpuAdvisory(ctx, c)
	env := Envelope{
		Kind: KindOverviewGPUTaskDetFull,
		Meta: NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
	}
	if advisoryNote != "" {
		env.Meta.Note = advisoryNote
	}
	env.Meta.Window = &TimeWindow{
		Since: humanizeSince(since),
		Start: gpuTrendTimestampISO(start),
		End:   gpuTrendTimestampISO(end),
		Step:  gpuTrendStep(start, end),
	}

	detail, err := fetchTaskDetail(ctx, c, name, podUID, sharemode)
	if err != nil {
		if he, ok := IsHTTPError(err); ok && he.Status == http.StatusNotFound {
			env.Meta.Empty = true
			env.Meta.EmptyReason = "no_vgpu_integration"
			env.Meta.HTTPStatus = he.Status
			return env, nil
		}
		if unavail, ok := vgpuUnavailableFromError(c, err, KindOverviewGPUTaskDetFull, now); ok {
			if env.Meta.Note != "" {
				unavail.Meta.Note = env.Meta.Note + " | " + unavail.Meta.Note
			}
			unavail.Meta.Window = env.Meta.Window
			return unavail, nil
		}
		return env, err
	}
	if len(detail) == 0 {
		env.Meta.Empty = true
		env.Meta.EmptyReason = "no_gpu_detected"
		return env, nil
	}

	// SPA: const displayAllocation = sharemode !== TimeSlicing
	allocationGauges := sharemode != "2"
	gaugeSpecs := taskDetailGaugeSpecs(allocationGauges)
	trendSpecs := taskDetailTrendSpecs()
	step := env.Meta.Window.Step
	startISO := env.Meta.Window.Start
	endISO := env.Meta.Window.End

	repl := strings.NewReplacer(
		"$container", fmt.Sprintf("%v", detail["name"]),
		"$pod", fmt.Sprintf("%v", detail["appName"]),
		"$namespace", fmt.Sprintf("%v", detail["namespace"]),
	)

	gaugeItems := make([]Item, len(gaugeSpecs))
	trendItems := make([]Item, len(trendSpecs))
	var (
		warnMu   sync.Mutex
		warnings []string
	)
	addWarning := func(msg string) {
		warnMu.Lock()
		warnings = append(warnings, msg)
		warnMu.Unlock()
	}

	g, gctx := errgroup.WithContext(ctx)
	for i, spec := range gaugeSpecs {
		i, spec := i, spec
		g.Go(func() error {
			gaugeItems[i] = runGPUGauge(gctx, c, spec, repl, addWarning)
			return nil
		})
	}
	for i, spec := range trendSpecs {
		i, spec := i, spec
		g.Go(func() error {
			item, _ := runGPUTrend(gctx, c, spec, repl, startISO, endISO, step, addWarning)
			trendItems[i] = item
			return nil
		})
	}
	_ = g.Wait()

	detailEnv := Envelope{
		Kind:  KindOverviewGPUTaskDet,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: []Item{{Raw: detail, Display: gpuTaskDetailDisplayCopy(detail)}},
	}
	gaugesEnv := Envelope{
		Kind:  KindOverviewGPUGauges,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: gaugeItems,
	}
	trendsEnv := Envelope{
		Kind:  KindOverviewGPUTrends,
		Meta:  NewMeta(now.In(common.Timezone.Time()), c.OlaresID(), common.User),
		Items: trendItems,
	}
	env.Sections = map[string]Envelope{
		"detail": detailEnv,
		"gauges": gaugesEnv,
		"trends": trendsEnv,
	}
	if len(warnings) > 0 {
		env.Meta.Warnings = warnings
	}
	return env, nil
}

// ----------------------------------------------------------------------------
// Per-query runners
// ----------------------------------------------------------------------------

// runGPUGauge executes one instant + (optional) total query and folds
// the result into a single Item. Soft failure: any HTTP/decode error is
// stored in the item's `raw.error` / `display.error` AND surfaced via
// `addWarning` so the parent envelope can list it under
// `meta.warnings`.
func runGPUGauge(ctx context.Context, c *Client, spec gpuGaugeSpec, repl *strings.Replacer, addWarning func(string)) Item {
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
		samples, err := fetchInstantVector(ctx, c, q)
		if err != nil {
			valErr = err
		} else if len(samples) > 0 {
			used = samples[0].Value
		}
	}
	if totalQ != "" {
		samples, err := fetchInstantVector(ctx, c, totalQ)
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
		"key":    spec.Key,
		"title":  spec.Title,
		"unit":   spec.Unit,
		"query":  q,
		"value":  used,
		"used":   used,
		"total":  total,
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
	return Item{Raw: raw, Display: disp}
}

// runGPUTrend executes the range query for every line in a trend spec.
// Returns the assembled Item plus the metric labels of the FIRST
// non-empty response (used to populate device_no / driver_version on
// the GPU detail page).
func runGPUTrend(ctx context.Context, c *Client, spec gpuTrendSpec, repl *strings.Replacer, start, end, step string, addWarning func(string)) (Item, map[string]string) {
	lineRaws := make([]map[string]any, 0, len(spec.Lines))
	lineDisps := make([]map[string]any, 0, len(spec.Lines))
	var firstLabels map[string]string
	var firstErr error
	for _, line := range spec.Lines {
		q := repl.Replace(line.Query)
		series, err := fetchRangeVector(ctx, c, q, start, end, step)
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
	return Item{Raw: raw, Display: disp}, firstLabels
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
// does: "8h" / "1h" / "30m" — empty when zero so JSON `omitempty` keeps
// the meta block tidy.
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

// resolveGPUDetailWindow picks (start, end, since) based on the user's
// --since / --start / --end / --watch flags. Mirrors the SPA's default
// `start = now - 8h` for GPU detail and `now - 1h` for task detail
// (passed in as `def`). Watch mode always slides the window with each
// tick (since-derived end = now).
func resolveGPUDetailWindow(now time.Time, def time.Duration) (start, end time.Time, since time.Duration) {
	if !common.Start.IsZero() && !common.End.IsZero() {
		return common.Start, common.End, 0
	}
	if common.Since > 0 {
		return now.Add(-common.Since), now, common.Since
	}
	return now.Add(-def), now, def
}

// ----------------------------------------------------------------------------
// Cobra wiring
// ----------------------------------------------------------------------------

func newOverviewGPUDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "detail <uuid>",
		Short:         "Per-GPU detail page (info + gauges + trends; SPA Overview2/GPU/GPUsDetails)",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUDetailFull(c.Context(), f, args[0])
		},
	}
	return cmd
}

func runOverviewGPUDetailFull(ctx context.Context, f *cmdutil.Factory, uuid string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			start, end, since := resolveGPUDetailWindow(now, gpuDetailDefaultSince)
			env, err := buildGPUDetailFullEnvelope(ctx, c, uuid, start, end, since)
			if err != nil {
				return env, err
			}
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeGPUDetailFullTable(env)
		},
	}
	return r.Run(ctx)
}

func newOverviewGPUTaskDetailFullCommand(f *cmdutil.Factory) *cobra.Command {
	var sharemode string
	cmd := &cobra.Command{
		Use:           "task-detail <name> <pod-uid>",
		Short:         "Per-task detail page (info + gauges + trends; SPA Overview2/GPU/TasksDetails)",
		Args:          cobra.ExactArgs(2),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := common.Validate(); err != nil {
				return err
			}
			return runOverviewGPUTaskDetailFull(c.Context(), f, args[0], args[1], sharemode)
		},
	}
	cmd.Flags().StringVar(&sharemode, "sharemode", "", `task share mode ("0"=App exclusive, "1"=Memory slicing, "2"=Time slicing). When "2", allocation gauges are skipped to match the SPA.`)
	return cmd
}

func runOverviewGPUTaskDetailFull(ctx context.Context, f *cmdutil.Factory, name, podUID, sharemode string) error {
	c, err := buildDashboardClient(ctx, f)
	if err != nil {
		return err
	}
	r := &Runner{
		Flags:       &common,
		Recommended: 30 * time.Second,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			start, end, since := resolveGPUDetailWindow(now, taskDetailDefaultSince)
			env, err := buildGPUTaskDetailFullEnvelope(ctx, c, name, podUID, sharemode, start, end, since)
			if err != nil {
				return env, err
			}
			if common.Output == OutputJSON {
				return env, nil
			}
			return env, writeGPUDetailFullTable(env)
		},
	}
	return r.Run(ctx)
}

// ----------------------------------------------------------------------------
// Table renderer (shared by gpu detail / task-detail)
// ----------------------------------------------------------------------------

// writeGPUDetailFullTable emits a three-block table view:
//
//	== Detail ==        — flat key/value list (the basic info card).
//	== Gauges ==        — N rows with title / value / unit / percent / used-total.
//	== Trends ==        — for each trend, a label header + (timestamp, value)
//	                      rows truncated to --head 16 by default.
//
// The output is intentionally text-only (no ANSI), so the bash smoke
// script can grep / awk through it.
func writeGPUDetailFullTable(env Envelope) error {
	out := os.Stdout
	if env.Meta.Empty {
		fmt.Fprintf(out, "(empty: %s", env.Meta.EmptyReason)
		if env.Meta.Note != "" {
			fmt.Fprintf(out, "; note: %s", env.Meta.Note)
		}
		fmt.Fprintln(out, ")")
		return nil
	}
	if env.Meta.Note != "" {
		fmt.Fprintf(os.Stderr, "(advisory) %s\n", env.Meta.Note)
	}
	if env.Meta.Window != nil {
		fmt.Fprintf(out, "Window: start=%s end=%s step=%s",
			env.Meta.Window.Start, env.Meta.Window.End, env.Meta.Window.Step)
		if env.Meta.Window.Since != "" {
			fmt.Fprintf(out, " since=%s", env.Meta.Window.Since)
		}
		fmt.Fprintln(out)
	}

	// Detail section.
	fmt.Fprintln(out, "\n== Detail ==")
	if dEnv, ok := env.Sections["detail"]; ok && len(dEnv.Items) > 0 {
		writeKeyValueTable(out, dEnv.Items[0])
	} else {
		fmt.Fprintln(out, "-")
	}

	// Gauges section.
	fmt.Fprintln(out, "\n== Gauges ==")
	if gEnv, ok := env.Sections["gauges"]; ok {
		cols := []TableColumn{
			{Header: "KEY", Get: func(it Item) string { return DisplayString(it, "key") }},
			{Header: "TITLE", Get: func(it Item) string { return DisplayString(it, "title") }},
			{Header: "VALUE", Get: func(it Item) string { return DisplayString(it, "value") }},
			{Header: "UNIT", Get: func(it Item) string { return DisplayString(it, "unit") }},
			{Header: "PERCENT", Get: func(it Item) string { return DisplayString(it, "percent") }},
			{Header: "USED/TOTAL", Get: func(it Item) string { return DisplayString(it, "used_total") }},
		}
		_ = WriteTable(out, cols, gEnv.Items)
	}

	// Trends section.
	fmt.Fprintln(out, "\n== Trends ==")
	if tEnv, ok := env.Sections["trends"]; ok {
		head := common.Head
		if head <= 0 {
			head = 16 // SPA renders ~16 buckets in the chart
		}
		for _, it := range tEnv.Items {
			title := DisplayString(it, "title")
			unit := DisplayString(it, "unit")
			fmt.Fprintf(out, "\n-- %s (%s) --\n", title, unit)
			if errStr, ok := it.Raw["error"].(string); ok && errStr != "" {
				fmt.Fprintf(out, "(error: %s)\n", errStr)
				continue
			}
			lines, _ := it.Raw["lines"].([]map[string]any)
			if len(lines) == 0 {
				fmt.Fprintln(out, "-")
				continue
			}
			for _, ln := range lines {
				label, _ := ln["label"].(string)
				fmt.Fprintf(out, "  %s:\n", label)
				points, _ := ln["points"].([]map[string]any)
				if len(points) == 0 {
					fmt.Fprintln(out, "    -")
					continue
				}
				rendered := points
				if head > 0 && head < len(points) {
					rendered = points[:head]
				}
				for _, p := range rendered {
					ts, _ := p["timestamp"].(string)
					v := p["value"]
					fmt.Fprintf(out, "    %s\t%v\n", ts, v)
				}
				if len(points) > len(rendered) {
					fmt.Fprintf(out, "    ... (%d more rows; pass --head 0 for full)\n", len(points)-len(rendered))
				}
			}
		}
	}

	if len(env.Meta.Warnings) > 0 {
		fmt.Fprintln(out, "\n== Warnings ==")
		for _, w := range env.Meta.Warnings {
			fmt.Fprintf(out, "- %s\n", w)
		}
	}
	return nil
}

// writeKeyValueTable renders one Item as a vertical key/value table.
// Sorted lexicographically so output is deterministic across runs.
func writeKeyValueTable(out *os.File, it Item) {
	if it.Display == nil {
		fmt.Fprintln(out, "-")
		return
	}
	keys := make([]string, 0, len(it.Display))
	for k := range it.Display {
		keys = append(keys, k)
	}
	sortStrings(keys)
	for _, k := range keys {
		v := it.Display[k]
		fmt.Fprintf(out, "%s\t%v\n", k, v)
	}
}

// sortStrings is a tiny helper to avoid pulling sort into this file's
// imports just for one site.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
