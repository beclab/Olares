package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// ----------------------------------------------------------------------------
// HAMI list / detail endpoints
// ----------------------------------------------------------------------------

// GraphicsListBody mirrors the SPA's GraphicsListParams. The fields are
// emitted UNCONDITIONALLY (no `omitempty`) — HAMI's WebUI rejects a body
// missing the `filters` key with a 500 "unknown request error" because
// downstream code dereferences the (would-be) Filters struct without a
// nil guard. The SPA always sends `"filters": {}` (see
// `Overview2/GPU/GPUsTable.vue:195-201`); we match that wire shape.
//
// History: an earlier revision used `omitempty` on both fields. With a
// nil-input filter map, encoding/json emits `{"pageRequest":{...}}` —
// HAMI then panics, the gin recovery middleware returns a generic 5xx,
// and `olares-cli dashboard overview gpu` lights up `vgpu_unavailable`
// while the SPA in the same browser tab continues to render data.
// `TestGraphicsListBody_AlwaysIncludesFiltersKey` is the regression net
// for this.
type GraphicsListBody struct {
	Filters     map[string]string `json:"filters"`
	PageRequest map[string]string `json:"pageRequest"`
}

// FetchGraphicsList posts to /hami/api/vgpu/v1/gpus and returns the
// (HAMI-flat) `list` array. Caller is responsible for the 404 (no
// HAMI integration) and 5xx (HAMI unhealthy) branches via
// VgpuUnavailableFromError / IsHTTPError.
func FetchGraphicsList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	if filters == nil {
		filters = map[string]string{}
	}
	body := GraphicsListBody{
		Filters: filters,
		PageRequest: map[string]string{
			"sort":      "DESC",
			"sortField": "id",
		},
	}
	// HAMI returns the list at the TOP LEVEL: `{"list": [...]}` — there
	// is no `data` envelope around it. The SPA's `GraphicsListResponse`
	// type confirms this (`{ list: Graphics[] }`). Wrapping in a `data`
	// struct here used to silently produce "0 GPUs" even on machines
	// where the SPA shows devices.
	var raw struct {
		List []map[string]any `json:"list"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/gpus", nil, body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/gpus", ErrorKind: "http_4xx"}
	}
	if status >= 400 {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/gpus", Body: string(payload), ErrorKind: ClassifyKind(status)}
	}
	if err := jsonUnmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw.List, nil
}

// FetchTaskList posts to /hami/api/vgpu/v1/containers and returns the
// flat `items` array. Same 404 / 5xx semantics as FetchGraphicsList.
func FetchTaskList(ctx context.Context, c *Client, filters map[string]string) ([]map[string]any, error) {
	if filters == nil {
		filters = map[string]string{}
	}
	body := GraphicsListBody{
		Filters: filters,
		PageRequest: map[string]string{
			"sort":      "DESC",
			"sortField": "id",
		},
	}
	// HAMI returns `{"items": [...]}` at the top level (matches
	// `TaskListResponse`). No `data` envelope.
	var raw struct {
		Items []map[string]any `json:"items"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/containers", nil, body)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/containers", ErrorKind: "http_4xx"}
	}
	if status >= 400 {
		return nil, &HTTPError{Status: status, URL: c.BaseURL() + "/hami/api/vgpu/v1/containers", Body: string(payload), ErrorKind: ClassifyKind(status)}
	}
	if err := jsonUnmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Items, nil
}

// FetchGraphicsDetail returns HAMI's `/v1/gpu` payload directly — the
// SPA's `GraphicsDetailsResponse` is a flat object, no `data` envelope.
func FetchGraphicsDetail(ctx context.Context, c *Client, uuid string) (map[string]any, error) {
	q := url.Values{"uuid": []string{uuid}}
	var raw map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/hami/api/vgpu/v1/gpu", q, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// FetchTaskDetail returns HAMI's `/v1/container` payload directly —
// like the GPU detail, the response is a flat object.
func FetchTaskDetail(ctx context.Context, c *Client, name, podUID, sharemode string) (map[string]any, error) {
	q := url.Values{"name": []string{name}, "podUid": []string{podUID}}
	if sharemode != "" {
		q.Set("sharemode", sharemode)
	}
	var raw map[string]any
	if err := c.DoJSON(ctx, http.MethodGet, "/hami/api/vgpu/v1/container", q, nil, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// ----------------------------------------------------------------------------
// HAMI monitor query endpoints (instant-vector / range-vector)
// ----------------------------------------------------------------------------
//
// IMPORTANT — wire-shape gotcha:
//
// The "list/tasks/detail" HAMI endpoints return their payload at the TOP
// LEVEL (no `data` envelope; see FetchGraphicsList et al. above). The
// **monitor query endpoints** are the exception — they DO wrap the result
// in a single-level `data` field, matching the SPA's
// `InstantVectorResponse { data: InstantVector[] }` and
// `RangeVectorResponse { data: RangeVector[] }` types in
// src/apps/dashboard/types/gpu.ts. Do NOT "normalise" the wrapper away;
// `TestFetchInstantVector_ParsesDataEnvelope` /
// `TestFetchRangeVector_ParsesDataEnvelope` enforce this contract.

type instantVectorBody struct {
	Query string `json:"query"`
}

// InstantVectorSample mirrors HAMI's `data[i]` row. Value is a number on
// the wire but float64 covers HAMI's full range (it caps at 1e308).
type InstantVectorSample struct {
	Metric    map[string]string `json:"metric"`
	Value     float64           `json:"value"`
	Timestamp string            `json:"timestamp"`
}

// FetchInstantVector posts `query` to /hami/api/vgpu/v1/monitor/query/instant-vector
// and returns the decoded `data` array. HAMI returns one element per
// matching series; most CLI gauges just read `data[0]`, but a query
// can theoretically expand into >1 series so we hand back the slice.
func FetchInstantVector(ctx context.Context, c *Client, query string) ([]InstantVectorSample, error) {
	body := instantVectorBody{Query: query}
	var raw struct {
		Data []InstantVectorSample `json:"data"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/monitor/query/instant-vector", nil, body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, &HTTPError{
			Status:    status,
			URL:       c.BaseURL() + "/hami/api/vgpu/v1/monitor/query/instant-vector",
			Body:      string(payload),
			ErrorKind: ClassifyKind(status),
		}
	}
	if err := jsonUnmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// RangeVectorRange mirrors the SPA's `RangeVectorParams.range` object.
type RangeVectorRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Step  string `json:"step"`
}

type rangeVectorBody struct {
	Query string           `json:"query"`
	Range RangeVectorRange `json:"range"`
}

// RangeVectorPoint is one (timestamp, value) pair inside a series. Both
// fields are strings on the wire (per the SPA's `RangeVector.values`
// type definition); the CLI parses them lazily on render.
type RangeVectorPoint struct {
	Value     any    `json:"value"`
	Timestamp string `json:"timestamp"`
}

// RangeVectorSeries mirrors HAMI's `data[i]` series row.
type RangeVectorSeries struct {
	Metric map[string]string  `json:"metric"`
	Values []RangeVectorPoint `json:"values"`
}

// FetchRangeVector posts a range query (start/end/step) to HAMI's
// /v1/monitor/query/range-vector. SPA's `getStepWithTimeRange` builds
// `step` (a string like "30m"); ISO-formatted start/end are computed by
// the caller via `GPUTrendTimestampISO`.
func FetchRangeVector(ctx context.Context, c *Client, query, start, end, step string) ([]RangeVectorSeries, error) {
	body := rangeVectorBody{
		Query: query,
		Range: RangeVectorRange{Start: start, End: end, Step: step},
	}
	var raw struct {
		Data []RangeVectorSeries `json:"data"`
	}
	status, payload, err := c.DoRaw(ctx, http.MethodPost, "/hami/api/vgpu/v1/monitor/query/range-vector", nil, body)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, &HTTPError{
			Status:    status,
			URL:       c.BaseURL() + "/hami/api/vgpu/v1/monitor/query/range-vector",
			Body:      string(payload),
			ErrorKind: ClassifyKind(status),
		}
	}
	if err := jsonUnmarshal(payload, &raw); err != nil {
		return nil, err
	}
	return raw.Data, nil
}

// ----------------------------------------------------------------------------
// GPU trend window helpers (1:1 port of the SPA's timeRangeFormate)
// ----------------------------------------------------------------------------

// GPUTrendStep is a 1:1 port of `timeRangeFormate(diff_s, 16)` from
// packages/app/src/apps/controlPanelCommon/containers/Monitoring/utils.js.
// Algorithm:
//
//  1. Convert (end-start) to whole minutes (rounded down).
//  2. If the minutes count matches one of the SPA's preset windows
//     (10/20/30, 60, 120, 180, 300, 480, 720, 1440, 4320, 10080), use
//     the matching preset step.
//  3. Otherwise compute `floor(minutes/16)m` (same as `getStep(value, 16)`),
//     then enforce a [1m..60m] range.
func GPUTrendStep(start, end time.Time) string {
	totalMinutes := int(end.Sub(start) / time.Minute)
	if totalMinutes <= 0 {
		// Defensive — empty / inverted ranges: pick the smallest sane
		// step so the caller doesn't divide by zero downstream.
		return "1m"
	}
	if step, ok := GPUStepPreset(totalMinutes); ok {
		return step
	}
	step := totalMinutes / 16
	if step < 1 {
		// SPA fallback: bump to a 10-bucket window when 16-bucket
		// rounds to 0m. (See the `if (stepNum < 1) { times = 10 }`
		// branch in utils.js.)
		step = totalMinutes / 10
		if step < 1 {
			step = 1
		}
	}
	if step > 60 {
		step = 60
	}
	return fmt.Sprintf("%dm", step)
}

// GPUStepPreset reproduces the `timeReflection` table in utils.js.
// Returns the preset step + true when minutes match a known window.
func GPUStepPreset(minutes int) (string, bool) {
	switch minutes {
	case 10, 20, 30:
		return "1m", true
	case 60:
		return "10m", true
	case 120:
		return "20m", true
	case 180, 300:
		return "10m", true
	case 480, 720:
		return "30m", true
	case 1440, 4320, 10080:
		return "60m", true
	default:
		return "", false
	}
}

// GPUTrendTimestampISO formats a time.Time the way the SPA's
// `timeParse(date)` does for monitor queries: `YYYY-MM-DD HH:mm:ss` in
// the caller's timezone (no offset suffix). HAMI's WebUI accepts
// either Unix-seconds or this human-readable form; the SPA exclusively
// sends the latter, so we match.
func GPUTrendTimestampISO(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// ----------------------------------------------------------------------------
// GPU rendering helpers (cmd-side leaves use these to format display strings)
// ----------------------------------------------------------------------------

// ClassifyKind maps an HTTP status code to the error_kind enum surfaced via
// Meta.ErrorKind / HTTPError.ErrorKind. Centralised so leaves don't drift.
func ClassifyKind(status int) string {
	switch {
	case status >= 500:
		return "http_5xx"
	case status >= 400:
		return "http_4xx"
	default:
		return ""
	}
}

// RenderTemperature picks the right unit suffix for ConvertTemperature.
// Used by overview cpu / overview fan live.
func RenderTemperature(celsius float64, target format.TempUnit) string {
	v := format.ConvertTemperature(celsius, target)
	suffix := "°C"
	switch target {
	case format.TempF:
		suffix = "°F"
	case format.TempK:
		suffix = "K"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.1f", v), "0"), ".") + suffix
}

// PercentString formats a 0..1 ratio as "N.NN%" (SPA style — utilisation
// metrics are percent of unit interval).
func PercentString(ratio float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", ratio*100), "0"), ".") + "%"
}

// PercentDirect formats a value already expressed as a percent (e.g. HAMI
// returns `coreUtilizedPercent: 25.5`, NOT 0.255). The SPA renders these
// with `round(val, 2) + '%'`; we match that and trim trailing zeros for
// readability ("25%" instead of "25.00%").
func PercentDirect(pct float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", pct), "0"), ".") + "%"
}

// GPUModeLabel translates HAMI's `shareMode` string into the SPA-rendered
// label. SPA mapping (constant/index.ts:VRAMModeLabel):
//
//	"0" → "App exclusive"
//	"1" → "Memory slicing"
//	"2" → "Time slicing"
//
// Real fixtures sometimes return "3" (observed on `olarestest005` —
// HAMI's WebUI silently falls back to showing the raw value). To avoid
// surfacing an empty cell in the CLI table we pass unknown values
// through unchanged, prefixed with `mode=` so a human can tell that we
// preserved the wire byte instead of mistranslating.
func GPUModeLabel(raw any) string {
	s := fmt.Sprintf("%v", raw)
	switch s {
	case "0":
		return "App exclusive"
	case "1":
		return "Memory slicing"
	case "2":
		return "Time slicing"
	case "":
		return "-"
	default:
		return "mode=" + s
	}
}

// GPUHealthLabel turns HAMI's boolean `health` into a human-readable
// status. The SPA leaves it as "true"/"false"; we surface the friendlier
// "healthy"/"unhealthy" pair (raw envelope still carries the original
// bool for agents that prefer the wire shape).
func GPUHealthLabel(raw any) string {
	if b, ok := raw.(bool); ok {
		if b {
			return "healthy"
		}
		return "unhealthy"
	}
	return fmt.Sprintf("%v", raw)
}

// FirstAnyInArray returns the first element of a slice-shaped value
// (e.g. `[]any` or `[]string`) decoded from JSON. HAMI returns
// per-device fields like `devicesCoreUtilizedPercent` as arrays — the
// SPA uses `val[0]` because tasks observed in the wild only ever bind
// a single device. We mirror that decision here, while preserving the
// full slice in `Raw` for multi-GPU consumers down the line.
func FirstAnyInArray(v any) any {
	switch x := v.(type) {
	case []any:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	case []string:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	case []float64:
		if len(x) == 0 {
			return nil
		}
		return x[0]
	default:
		return nil
	}
}

// GPUVRAMHuman formats a MiB count (HAMI's `memoryTotal` / `memoryUsed`
// units) as a SPA-style "1.5GiB"-shaped string. Mirrors the SPA's
// `getDiskSize(val * 1024 * 1024)` call; treats 0 as "-" so the table
// doesn't show a misleading "0B" for honest "no allocation" cases.
func GPUVRAMHuman(mibVal any) string {
	mib := ToFloat(mibVal)
	if mib <= 0 {
		return "-"
	}
	bytes := mib * 1024.0 * 1024.0
	return format.GetDiskSize(strconv.FormatFloat(bytes, 'f', -1, 64))
}

// ToFloat coerces an arbitrary JSON-decoded scalar into a float64.
// Returns 0 for nil / empty / unparsable inputs. Used by the cmd-side
// leaves' display building plus GPUVRAMHuman.
func ToFloat(v any) float64 {
	switch x := v.(type) {
	case nil:
		return 0
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		if x == "" {
			return 0
		}
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}
