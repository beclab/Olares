package gpu

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// TestBuildDetailFullEnvelope_PartialFailure: when ONE gauge query
// returns 5xx, the envelope is still emitted (Empty=false) with
// the failed gauge carrying meta.error=… and meta.warnings
// populated. Tests the gpu detail page's soft-failure invariant —
// SPA's behaviour is to render the other 5 gauges + 4 trends even
// when one query is broken; the CLI must do the same so an agent
// inspecting `meta.warnings` can decide whether the partial answer
// is enough.
func TestBuildDetailFullEnvelope_PartialFailure(t *testing.T) {
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"uuid":"GPU-1","type":"NVIDIA","health":true,"shareMode":"0"}`))
		},
		instantVector: func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), `hami_core_util{`) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(`{"message":"upstream unavailable"}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":12.5,"timestamp":"1745000000"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{"device_no":"nvidia0","driver_version":"590.44.01"},"values":[{"value":1.0,"timestamp":"1745000000"}]}]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	end := time.Date(2026, 4, 28, 22, 0, 0, 0, time.UTC)
	start := end.Add(-8 * time.Hour)
	env, err := BuildDetailFullEnvelope(context.Background(), c, cf, "GPU-1", start, end, 8*time.Hour)
	if err != nil {
		t.Fatalf("BuildDetailFullEnvelope: %v", err)
	}
	if env.Meta.Empty {
		t.Fatal("envelope empty=true; partial failure should not nuke the whole envelope")
	}
	if len(env.Meta.Warnings) == 0 {
		t.Fatal("expected env.Meta.Warnings to capture the failed gauge")
	}
	foundKey := false
	for _, warn := range env.Meta.Warnings {
		if strings.Contains(warn, `"util_core"`) {
			foundKey = true
			break
		}
	}
	if !foundKey {
		t.Errorf("warnings = %v; expected one mentioning util_core", env.Meta.Warnings)
	}
	gauges := env.Sections["gauges"].Items
	if len(gauges) != 6 {
		t.Fatalf("len(gauges) = %d, want 6", len(gauges))
	}
	if _, ok := gauges[2].Raw["error"]; !ok {
		t.Errorf("gauges[util_core].raw.error missing; got raw=%v", gauges[2].Raw)
	}
	if v, ok := gauges[0].Raw["value"].(float64); !ok || v == 0 {
		t.Errorf("gauges[alloc_core].raw.value = %v; expected non-zero", gauges[0].Raw["value"])
	}
	det := env.Sections["detail"].Items[0].Raw
	if det["device_no"] != "nvidia0" {
		t.Errorf("detail.device_no = %v, want nvidia0", det["device_no"])
	}
	if det["driver_version"] != "590.44.01" {
		t.Errorf("detail.driver_version = %v, want 590.44.01", det["driver_version"])
	}
}

// TestBuildDetailFullEnvelope_WireAndRenderTZsAreSeparate pins the
// post-fix wire/render TZ split. meta.window.start/end is rendered
// in cf.Timezone (user-visible); the body sent to HAMI's
// /monitor/query/range-vector endpoint is rendered in
// HAMIBackendTimezone() (whatever the backend pod parses with).
//
// We exercise the split by choosing DIFFERENT zones for the two:
//   - cf.Timezone = Asia/Shanghai → meta.window in CST
//   - OLARES_HAMI_BACKEND_TZ = America/Los_Angeles (May → PDT, UTC-7)
//     → wire in PDT
//
// If the two paths regress to a single TZ the test fails on either
// the meta.window assertion (CST != PDT) or the wire-body assertion
// (PDT != CST). Naming intentionally drops the old
// "WindowFormattedInTimezone" suffix — that name implied a single
// TZ governs both sides, which is exactly what we're moving away
// from.
func TestBuildDetailFullEnvelope_WireAndRenderTZsAreSeparate(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skipf("Asia/Shanghai zone unavailable: %v", err)
	}
	if _, err := time.LoadLocation("America/Los_Angeles"); err != nil {
		t.Skipf("America/Los_Angeles zone unavailable: %v", err)
	}
	t.Setenv("OLARES_HAMI_BACKEND_TZ", "America/Los_Angeles")

	var (
		mu              sync.Mutex
		rangeWireBodies []map[string]any
	)
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"uuid":"GPU-1","type":"NVIDIA","health":true,"shareMode":"0","nodeName":"olares"}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":1.0,"timestamp":"1745000000"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			_ = json.Unmarshal(body, &req)
			if rng, ok := req["range"].(map[string]any); ok {
				mu.Lock()
				rangeWireBodies = append(rangeWireBodies, rng)
				mu.Unlock()
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Timezone = format.NewLocation(shanghai)
	cf.Output = pkgdashboard.OutputJSON

	// Same absolute instant as the original user-reported bug
	// (2026-05-25 11:39:54 UTC). Conversions:
	//   end   = 2026-05-25 11:39:54 UTC = 19:39:54 CST = 04:39:54 PDT
	//   start = 2026-05-25 03:39:54 UTC = 11:39:54 CST = 2026-05-24 20:39:54 PDT
	end := time.Date(2026, 5, 25, 11, 39, 54, 0, time.UTC)
	start := end.Add(-8 * time.Hour)
	env, err := BuildDetailFullEnvelope(context.Background(), c, cf, "GPU-1", start, end, 8*time.Hour)
	if err != nil {
		t.Fatalf("BuildDetailFullEnvelope: %v", err)
	}
	// meta.window is the USER-VISIBLE side → CST (cf.Timezone).
	wantMetaStart := "2026-05-25 11:39:54"
	wantMetaEnd := "2026-05-25 19:39:54"
	if env.Meta.Window == nil {
		t.Fatal("meta.window is nil; the agent contract requires it for replays")
	}
	if env.Meta.Window.Start != wantMetaStart {
		t.Errorf("meta.window.start = %q, want %q (CST rendering for the user)", env.Meta.Window.Start, wantMetaStart)
	}
	if env.Meta.Window.End != wantMetaEnd {
		t.Errorf("meta.window.end = %q, want %q (CST rendering for the user)", env.Meta.Window.End, wantMetaEnd)
	}

	// Wire is the BACKEND-VISIBLE side → PDT (HAMI backend TZ
	// override). Every range-vector body in the fan-out MUST carry
	// the PDT-shaped strings, NOT the CST strings from meta.window.
	mu.Lock()
	defer mu.Unlock()
	if len(rangeWireBodies) == 0 {
		t.Fatal("no range-vector requests captured; fan-out broken?")
	}
	wantWireStart := "2026-05-24 20:39:54"
	wantWireEnd := "2026-05-25 04:39:54"
	for i, rng := range rangeWireBodies {
		if rng["start"] != wantWireStart {
			t.Errorf("range[%d].start sent to HAMI = %q, want %q (PDT, backend TZ); MUST NOT equal meta.window.start %q",
				i, rng["start"], wantWireStart, wantMetaStart)
		}
		if rng["end"] != wantWireEnd {
			t.Errorf("range[%d].end sent to HAMI = %q, want %q (PDT, backend TZ); MUST NOT equal meta.window.end %q",
				i, rng["end"], wantWireEnd, wantMetaEnd)
		}
		// Belt-and-suspenders: if the two paths ever collapse back
		// into a single TZ, at least one of these equalities will
		// hold and we want a screaming-loud test failure for it.
		if rng["start"] == wantMetaStart && rng["end"] == wantMetaEnd {
			t.Errorf("range[%d] wire body equals meta.window — wire/render TZ split regressed", i)
		}
	}
}

// TestBuildDetailFullEnvelope_UTCHostStillGetsData is the explicit
// regression net for the original symptom: a CLI session on a UTC
// host (cf.Timezone == UTC, no --timezone override) used to send
// offset-less UTC wall clock strings to a HAMI backend running on
// Asia/Shanghai, which then re-parsed them as CST and queried
// Prometheus 8h in the future → `data: []`.
//
// The fix: wire side is always HAMIBackendTimezone() (default
// Asia/Shanghai), independent of cf.Timezone. This test asserts the
// invariant by checking the wire body is CST even though the user
// asked for UTC rendering. If the fix regresses the test catches it
// at PR time, not at customer-deploy time.
func TestBuildDetailFullEnvelope_UTCHostStillGetsData(t *testing.T) {
	if _, err := time.LoadLocation("Asia/Shanghai"); err != nil {
		t.Skipf("Asia/Shanghai zone unavailable: %v", err)
	}
	t.Setenv("OLARES_HAMI_BACKEND_TZ", "") // default → Asia/Shanghai

	var (
		mu              sync.Mutex
		rangeWireBodies []map[string]any
	)
	srv := gpuStubMux{
		graphicsGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"uuid":"GPU-1","type":"NVIDIA","health":true,"shareMode":"0","nodeName":"olares"}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":1.0,"timestamp":"1745000000"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			_ = json.Unmarshal(body, &req)
			if rng, ok := req["range"].(map[string]any); ok {
				mu.Lock()
				rangeWireBodies = append(rangeWireBodies, rng)
				mu.Unlock()
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Timezone = format.NewLocation(time.UTC) // simulate UTC host
	cf.Output = pkgdashboard.OutputJSON

	end := time.Date(2026, 5, 25, 11, 39, 54, 0, time.UTC)
	start := end.Add(-8 * time.Hour)
	env, err := BuildDetailFullEnvelope(context.Background(), c, cf, "GPU-1", start, end, 8*time.Hour)
	if err != nil {
		t.Fatalf("BuildDetailFullEnvelope: %v", err)
	}
	// User asked for UTC rendering → meta.window is UTC.
	if got, want := env.Meta.Window.Start, "2026-05-25 03:39:54"; got != want {
		t.Errorf("meta.window.start = %q, want %q (UTC rendering for the user)", got, want)
	}
	if got, want := env.Meta.Window.End, "2026-05-25 11:39:54"; got != want {
		t.Errorf("meta.window.end = %q, want %q (UTC rendering for the user)", got, want)
	}

	// Wire is CST regardless of cf.Timezone → this is the bugfix.
	mu.Lock()
	defer mu.Unlock()
	if len(rangeWireBodies) == 0 {
		t.Fatal("no range-vector requests captured; fan-out broken?")
	}
	wantWireStart := "2026-05-25 11:39:54"
	wantWireEnd := "2026-05-25 19:39:54"
	for i, rng := range rangeWireBodies {
		if rng["start"] != wantWireStart {
			t.Errorf("range[%d].start sent to HAMI = %q, want %q (CST, backend TZ — UTC host bug regression)", i, rng["start"], wantWireStart)
		}
		if rng["end"] != wantWireEnd {
			t.Errorf("range[%d].end sent to HAMI = %q, want %q (CST, backend TZ — UTC host bug regression)", i, rng["end"], wantWireEnd)
		}
	}
}

// TestBuildTaskDetailFullEnvelope_WireAndRenderTZsAreSeparate is
// the task-flavoured twin of
// TestBuildDetailFullEnvelope_WireAndRenderTZsAreSeparate. Both
// builders share the same fanoutGaugeAndTrend plumbing; we keep
// the twin to lock in that no task-detail caller accidentally
// re-conflates the two TZs (the SPA's task page uses the same
// /monitor/query/range-vector endpoint, so the contract is
// identical).
func TestBuildTaskDetailFullEnvelope_WireAndRenderTZsAreSeparate(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Skipf("Asia/Shanghai zone unavailable: %v", err)
	}
	if _, err := time.LoadLocation("America/Los_Angeles"); err != nil {
		t.Skipf("America/Los_Angeles zone unavailable: %v", err)
	}
	t.Setenv("OLARES_HAMI_BACKEND_TZ", "America/Los_Angeles")

	var (
		mu              sync.Mutex
		rangeWireBodies []map[string]any
	)
	srv := gpuStubMux{
		taskGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"name":"task-A","status":"running","appName":"pod-1","namespace":"ns-1","podUid":"pod-1","deviceShareModes":["2"]}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":1.0,"timestamp":"1745000000"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var req map[string]any
			_ = json.Unmarshal(body, &req)
			if rng, ok := req["range"].(map[string]any); ok {
				mu.Lock()
				rangeWireBodies = append(rangeWireBodies, rng)
				mu.Unlock()
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Timezone = format.NewLocation(shanghai)
	cf.Output = pkgdashboard.OutputJSON

	// end = 2026-05-25 11:40:20 UTC
	//   = 19:40:20 CST (meta.window)
	//   = 04:40:20 PDT (wire)
	// start = end - 1h = 2026-05-25 10:40:20 UTC
	//   = 18:40:20 CST
	//   = 03:40:20 PDT
	end := time.Date(2026, 5, 25, 11, 40, 20, 0, time.UTC)
	start := end.Add(-1 * time.Hour)
	env, err := BuildTaskDetailFullEnvelope(context.Background(), c, cf, "task-A", "pod-1", "2", start, end, time.Hour)
	if err != nil {
		t.Fatalf("BuildTaskDetailFullEnvelope: %v", err)
	}
	if env.Meta.Window == nil {
		t.Fatal("meta.window is nil")
	}
	if got, want := env.Meta.Window.Start, "2026-05-25 18:40:20"; got != want {
		t.Errorf("meta.window.start = %q, want %q (CST for the user)", got, want)
	}
	if got, want := env.Meta.Window.End, "2026-05-25 19:40:20"; got != want {
		t.Errorf("meta.window.end = %q, want %q (CST for the user)", got, want)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(rangeWireBodies) == 0 {
		t.Fatal("no range-vector requests captured; fan-out broken?")
	}
	wantWireStart := "2026-05-25 03:40:20"
	wantWireEnd := "2026-05-25 04:40:20"
	for i, rng := range rangeWireBodies {
		if rng["start"] != wantWireStart {
			t.Errorf("range[%d].start = %q, want %q (PDT, backend TZ)", i, rng["start"], wantWireStart)
		}
		if rng["end"] != wantWireEnd {
			t.Errorf("range[%d].end = %q, want %q (PDT, backend TZ)", i, rng["end"], wantWireEnd)
		}
	}
}

// TestBuildTaskDetailFullEnvelope_TimeSlicingSkipsAllocation: when
// --sharemode is "2" (TimeSlicing) the SPA hides the allocation
// gauges; CLI must do the same. Captures the wiring intent so a
// future "always emit allocation gauges" refactor fails loudly
// here.
func TestBuildTaskDetailFullEnvelope_TimeSlicingSkipsAllocation(t *testing.T) {
	srv := gpuStubMux{
		taskGet: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"name":"task-A","status":"running","appName":"pod-1","namespace":"ns-1","podUid":"pod-1","deviceShareModes":["2"]}`))
		},
		instantVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":10,"timestamp":"1745000000"}]}`))
		},
		rangeVector: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"values":[{"value":5,"timestamp":"1745000000"}]}]}`))
		},
	}.server(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Output = pkgdashboard.OutputJSON

	end := time.Date(2026, 4, 28, 22, 0, 0, 0, time.UTC)
	start := end.Add(-1 * time.Hour)
	env, err := BuildTaskDetailFullEnvelope(context.Background(), c, cf, "task-A", "pod-1", "2", start, end, time.Hour)
	if err != nil {
		t.Fatalf("BuildTaskDetailFullEnvelope: %v", err)
	}
	gauges := env.Sections["gauges"].Items
	if len(gauges) != 2 {
		t.Fatalf("len(gauges) = %d, want 2 (TimeSlicing should never expose allocation gauges)", len(gauges))
	}
	for _, g := range gauges {
		k := g.Raw["key"].(string)
		if strings.HasPrefix(k, "alloc_") {
			t.Errorf("found allocation gauge %q in TimeSlicing mode — should have been skipped", k)
		}
	}
	trends := env.Sections["trends"].Items
	if len(trends) != 2 {
		t.Errorf("len(trends) = %d, want 2", len(trends))
	}
	det := env.Sections["detail"].Items[0].Raw
	if det["name"] != "task-A" {
		t.Errorf("detail.name = %v, want task-A", det["name"])
	}
}
