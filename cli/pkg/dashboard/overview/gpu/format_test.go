package gpu

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// TestFormatTrendTimestamp_AllShapes pins the three timestamp
// shapes the CLI must accept from HAMI's range-vector endpoint —
// epoch ms (13-digit int string, the shape that triggered the
// user-visible bug "1779636713000"), epoch s (10-digit), and
// float-seconds with sub-second precision (Prometheus's wire
// shape). All three MUST collapse to the SPA's
// `YYYY-MM-DD HH:mm:ss` rendering. Edge cases (empty,
// unparseable) MUST fall through to the raw string instead of
// silently rendering "1970-01-01" — agents debugging HAMI need
// to see what HAMI actually sent.
func TestFormatTrendTimestamp_AllShapes(t *testing.T) {
	utc := time.UTC

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"epoch_ms_13_digits", "1779636713000", "2026-05-24 15:31:53"},
		{"epoch_s_10_digits", "1779636713", "2026-05-24 15:31:53"},
		{"epoch_s_with_subsec", "1779636713.5", "2026-05-24 15:31:53"},
		{"empty_returns_dash", "", "-"},
		{"unparseable_passes_through", "not-a-number", "not-a-number"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := formatTrendTimestamp(tc.raw, utc)
			if got != tc.want {
				t.Errorf("formatTrendTimestamp(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestFormatTrendTimestamp_RespectsTimezone pins that the caller's
// --timezone is actually applied. The same epoch produces
// different strings for UTC vs Asia/Shanghai (UTC+8). Without
// honoring tz the CLI would emit UTC even when the user explicitly
// asked for local — a footgun for ops people parsing logs.
func TestFormatTrendTimestamp_RespectsTimezone(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skipf("Asia/Shanghai zone unavailable: %v", err)
	}
	utc := formatTrendTimestamp("1779636713000", time.UTC)
	cn := formatTrendTimestamp("1779636713000", shanghai)
	if utc == cn {
		t.Fatalf("UTC == Asia/Shanghai for the same epoch (%q == %q); tz argument ignored?", utc, cn)
	}
	if !strings.HasPrefix(utc, "2026-05-24 15:") {
		t.Errorf("UTC rendering = %q, want 2026-05-24 15:…", utc)
	}
	if !strings.HasPrefix(cn, "2026-05-24 23:") {
		t.Errorf("CN rendering = %q, want 2026-05-24 23:… (UTC+8 of 15:31:53)", cn)
	}
}

// TestRoundDP_HalfAwayFromZero pins the SPA's lodash.round behavior
// — half-away-from-zero, NOT banker's. Without this someone could
// "fix" CLI to use math.RoundToEven and silently desync from the
// SPA chart values for any .5-ending input.
func TestRoundDP_HalfAwayFromZero(t *testing.T) {
	cases := []struct {
		in   float64
		dp   int
		want float64
	}{
		{0.125, 2, 0.13},
		{0.135, 2, 0.14},
		{2.5, 0, 3},
		{-2.5, 0, -3},
		{23.889648, 2, 23.89},
		{0.28808594, 2, 0.29},
		{99.99876, 2, 100},
	}
	for _, tc := range cases {
		if got := roundDP(tc.in, tc.dp); got != tc.want {
			t.Errorf("roundDP(%v,%d) = %v, want %v", tc.in, tc.dp, got, tc.want)
		}
	}
}

// TestRoundedNumberString_StripsTrailingZeros — the SPA renders
// `String(round(23.89, 2))` as "23.89", not "23.890000000000004"
// or "23.890000". This is what makes the tabular output legible.
func TestRoundedNumberString_StripsTrailingZeros(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{23.889648, "23.89"},
		{0.28808594, "0.29"},
		{100.0, "100"},
		{0.0, "0"},
		{7.866, "7.87"},
	}
	for _, tc := range cases {
		if got := roundedNumberString(tc.in); got != tc.want {
			t.Errorf("roundedNumberString(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFormatGaugeValue pins the `<value> <unit>` join for SPA
// parity. The four ratio gauges (unit=" " or unit="") must NOT
// emit a trailing space + unit; the unit-bearing gauges (Gi / W /
// ℃ / %) MUST. This is exactly what `<MyGaugeChart unit=…>` does
// — the CLI table column "VALUE" matches the gauge dial label
// 1:1.
func TestFormatGaugeValue(t *testing.T) {
	cases := []struct {
		v    float64
		unit string
		want string
	}{
		{23.889648, "Gi", "23.89 Gi"},
		{7.866, "W", "7.87 W"},
		{46.0, "℃", "46 ℃"},
		{1.21, "%", "1.21 %"},
		{100.0, "", "100"},  // ratio gauges (alloc_core) — no unit
		{100.0, " ", "100"}, // SPA's literal " " sentinel
		{0, "Gi", "0 Gi"},
	}
	for _, tc := range cases {
		if got := formatGaugeValue(tc.v, tc.unit); got != tc.want {
			t.Errorf("formatGaugeValue(%v,%q) = %q, want %q", tc.v, tc.unit, got, tc.want)
		}
	}
}

// TestGpuDetailDisplayCopy_SPAFieldWhitelist freezes the 6-field
// whitelist matching GPUsDetails.vue:112-137. If anyone widens
// the whitelist (e.g. brings memoryTotal/power/temperature back
// into the Detail card) this test fails — a deliberate guardrail
// because those fields are placeholders in HAMI's flat
// /v1/gpu body and re-introducing them was the bug we just fixed.
func TestGpuDetailDisplayCopy_SPAFieldWhitelist(t *testing.T) {
	// Full HAMI body — should be available on Raw, but Display must
	// only surface the SPA's 6 columns.
	src := map[string]any{
		"uuid":           "GPU-A",
		"type":           "NVIDIA GeForce RTX 5090",
		"nodeName":       "olares",
		"nodeUid":        "node-uid-redacted",
		"shareMode":      "0",
		"vgpuUsed":       1,
		"vgpuTotal":      30,
		"coreUsed":       100,
		"coreTotal":      100,
		"memoryUsed":     0,
		"memoryTotal":    24463,
		"power":          0,
		"powerLimit":     0,
		"temperature":    0,
		"device_no":      "nvidia0",
		"driver_version": "595.80",
		"health":         true,
	}
	got := gpuDetailDisplayCopy(src)

	wantKeys := map[string]bool{
		"uuid": true, "type": true, "nodeName": true,
		"device_no": true, "driver_version": true, "health": true,
	}
	for k := range got {
		if !wantKeys[k] {
			t.Errorf("Display has unexpected key %q (SPA columns whitelist disallows it); full = %v", k, got)
		}
	}
	for k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Errorf("Display missing whitelisted key %q; full = %v", k, got)
		}
	}
	// health passes through gpuHealthLabel — pin the label conversion.
	if got["health"] != "healthy" {
		t.Errorf("Display.health = %q, want \"healthy\"", got["health"])
	}
}

// TestGpuTaskDetailDisplayCopy_SPAFieldWhitelist freezes the
// task-detail whitelist matching TasksDetails.vue:134-174.
// Allocations are conditional in SPA (displayAllocation) —
// CLI emits them when present, omits when absent.
func TestGpuTaskDetailDisplayCopy_SPAFieldWhitelist(t *testing.T) {
	src := map[string]any{
		"name":             "comfyui",
		"status":           "running",
		"podUid":           "pod-uid-redacted",
		"nodeName":         "olares",
		"nodeUid":          "node-uid-redacted",
		"type":             "NVIDIA-vGPU",
		"appName":          "comfyui-7d4f4915-7d-zwmgs",
		"namespace":        "ns-1",
		"createTime":       "2026-05-25T15:00:00Z",
		"startTime":        "2026-05-25T15:00:01Z",
		"endTime":          "2026-05-25T16:00:00Z",
		"deviceIds":        []any{"GPU-A"},
		"deviceShareModes": []any{"2"},
		"allocatedDevices": 1,
		"allocatedCores":   25,
		"allocatedMem":     "256MiB",
		"resourcePool":     "pool-1",
		"flavor":           "small",
		"priority":         "high",
	}
	got := gpuTaskDetailDisplayCopy(src)

	wantKeys := map[string]bool{
		"status": true, "deviceIds": true, "nodeName": true, "type": true,
		"allocatedCores": true, "allocatedMem": true,
		"appName": true, "createTime": true,
	}
	for k := range got {
		if !wantKeys[k] {
			t.Errorf("Display has unexpected key %q (SPA columns whitelist disallows it); full = %v", k, got)
		}
	}
	for k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Errorf("Display missing whitelisted key %q; full = %v", k, got)
		}
	}
	if got["deviceIds"] != "GPU-A" {
		t.Errorf("deviceIds join = %q, want \"GPU-A\"", got["deviceIds"])
	}
}

// TestRunDetailFull_TrendPointsAreFormatted is the end-to-end pin
// for the user-visible bug ("graphics 的时间格式是 '2026-05-25 …'，
// 数据要格式化成合适的单位"). Stubs HAMI to return:
//   - a range-vector point whose timestamp is epoch ms
//     ("1779636713000") and value is a noisy 7-digit float
//     (`23.889648`).
//
// The envelope's `trends.points[0].timestamp` MUST be
// `2026-05-24 …` (SPA `YYYY-MM-DD HH:mm:ss`), `value` MUST be
// `23.89` (lodash.round(2)), and the wire-shape epoch MUST be
// preserved on `timestamp_ms`. End-to-end test catches a future
// "I refactored runTrend back to raw passthrough" regression.
func TestRunDetailFull_TrendPointsAreFormatted(t *testing.T) {
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"uuid":"GPU-1","type":"NVIDIA","health":true,"shareMode":"0","nodeName":"olares"}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":7.866,"timestamp":"1779636713"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"values":[
              {"value":23.889648,"timestamp":"1779636713000"},
              {"value":99.99876,"timestamp":"1779636773000"}
            ]}]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	// Pin UTC so the assertion strings stay deterministic regardless
	// of the test runner's host timezone — fixtureFlags otherwise
	// uses LocalLocation() which would render Asia/Shanghai on the
	// dev box and UTC on CI.
	cf.Timezone = format.NewLocation(time.UTC)
	cf.Output = pkgdashboard.OutputJSON

	end := time.Date(2026, 5, 24, 23, 31, 53, 0, time.UTC)
	start := end.Add(-8 * time.Hour)
	envelope, err := BuildDetailFullEnvelope(context.Background(), c, cf, "GPU-1", start, end, 8*time.Hour)
	if err != nil {
		t.Fatalf("BuildDetailFullEnvelope: %v", err)
	}
	trends := envelope.Sections["trends"].Items
	if len(trends) == 0 {
		t.Fatal("expected at least one trend item")
	}
	// Find first trend item with non-empty points.
	var firstPoint map[string]any
	for _, it := range trends {
		lines, _ := it.Raw["lines"].([]map[string]any)
		for _, ln := range lines {
			points, _ := ln["points"].([]map[string]any)
			if len(points) > 0 {
				firstPoint = points[0]
				break
			}
		}
		if firstPoint != nil {
			break
		}
	}
	if firstPoint == nil {
		t.Fatal("no trend points captured — stub returned empty?")
	}
	if got := firstPoint["timestamp"]; got != "2026-05-24 15:31:53" {
		t.Errorf("trend point timestamp = %v, want \"2026-05-24 15:31:53\" (epoch ms 1779636713000 in UTC)", got)
	}
	if got, _ := firstPoint["timestamp_ms"].(int64); got != 1779636713000 {
		t.Errorf("trend point timestamp_ms = %v, want 1779636713000 (wire-shape preserved)", firstPoint["timestamp_ms"])
	}
	if got, _ := firstPoint["value"].(float64); got != 23.89 {
		t.Errorf("trend point value = %v, want 23.89 (lodash.round(2) of 23.889648)", firstPoint["value"])
	}
	if got, _ := firstPoint["value_raw"].(float64); got != 23.889648 {
		t.Errorf("trend point value_raw = %v, want 23.889648 (full precision preserved)", firstPoint["value_raw"])
	}
}

// TestGpuTaskDetailDisplayCopy_OmitsAllocationsWhenAbsent — when
// HAMI doesn't return allocatedCores / allocatedMem (the
// time-slicing branch) the rows MUST be absent from Display, not
// rendered as "0" or "<nil>". Pinned because the SPA's
// `displayAllocation` gate guarantees the same: the user sees a
// 6-row card, not 8 rows with two empty.
func TestGpuTaskDetailDisplayCopy_OmitsAllocationsWhenAbsent(t *testing.T) {
	src := map[string]any{
		"name":      "task-A",
		"status":    "running",
		"appName":   "app-a",
		"nodeName":  "node-1",
		"type":      "NVIDIA-vGPU",
		"deviceIds": []any{"GPU-A"},
	}
	got := gpuTaskDetailDisplayCopy(src)
	if _, ok := got["allocatedCores"]; ok {
		t.Errorf("allocatedCores must be absent when source omits it; got=%v", got)
	}
	if _, ok := got["allocatedMem"]; ok {
		t.Errorf("allocatedMem must be absent when source omits it; got=%v", got)
	}
}
