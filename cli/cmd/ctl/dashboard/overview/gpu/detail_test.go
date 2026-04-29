package gpu

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestBuildGPUDetailFullEnvelope_PartialFailure: when ONE gauge query
// returns 5xx, the envelope is still emitted (Empty=false) with the
// failed gauge carrying meta.error=… and meta.warnings populated. Tests
// in this file live alongside the cmd-side detail/task_detail builders
// they exercise; pure pkg-level fetchers are tested under cli/pkg/
// dashboard/.
func TestBuildGPUDetailFullEnvelope_PartialFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/capi/app/detail"):
			_, _ = w.Write([]byte(`{"clusterRole":"workspaces-manager","user":{"username":"alice","globalrole":"platform-admin","email":"alice@olares.com"}}`))
		case strings.Contains(r.URL.Path, "/kapis/resources.kubesphere.io"):
			_, _ = w.Write([]byte(`{"items":[{"metadata":{"name":"node-1","labels":{"gpu.bytetrade.io/cuda-supported":"true"}}}]}`))
		case strings.HasSuffix(r.URL.Path, "/v1/gpu"):
			_, _ = w.Write([]byte(`{"uuid":"GPU-1","type":"NVIDIA","health":true,"shareMode":"0"}`))
		case strings.HasSuffix(r.URL.Path, "/instant-vector"):
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), `hami_core_util{`) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte(`{"message":"upstream unavailable"}`))
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":12.5,"timestamp":"1745000000"}]}`))
		case strings.HasSuffix(r.URL.Path, "/range-vector"):
			_, _ = w.Write([]byte(`{"data":[{"metric":{"device_no":"nvidia0","driver_version":"590.44.01"},"values":[{"value":1.0,"timestamp":"1745000000"}]}]}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)

	prev := common
	cf := pkgdashboard.CommonFlags{Output: pkgdashboard.OutputJSON}
	_ = cf.Validate()
	common = &cf
	defer func() { common = prev }()

	end := time.Date(2026, 4, 28, 22, 0, 0, 0, time.UTC)
	start := end.Add(-8 * time.Hour)
	env, err := buildGPUDetailFullEnvelope(context.Background(), c, "GPU-1", start, end, 8*time.Hour)
	if err != nil {
		t.Fatalf("buildGPUDetailFullEnvelope: %v", err)
	}
	if env.Meta.Empty {
		t.Fatal("envelope empty=true; partial failure should not nuke the whole envelope")
	}
	if len(env.Meta.Warnings) == 0 {
		t.Fatal("expected env.Meta.Warnings to capture the failed gauge")
	}
	foundKey := false
	for _, w := range env.Meta.Warnings {
		if strings.Contains(w, `"util_core"`) {
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

// TestBuildGPUTaskDetailFullEnvelope_TimeSlicingSkipsAllocation: when
// --sharemode is "2" (TimeSlicing) the SPA hides the allocation gauges;
// CLI must do the same. Captures the wiring intent so a future
// "always emit allocation gauges" refactor fails loudly here.
func TestBuildGPUTaskDetailFullEnvelope_TimeSlicingSkipsAllocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/capi/app/detail"):
			_, _ = w.Write([]byte(`{"clusterRole":"workspaces-manager","user":{"username":"alice","globalrole":"platform-admin","email":"alice@olares.com"}}`))
		case strings.Contains(r.URL.Path, "/kapis/resources.kubesphere.io"):
			_, _ = w.Write([]byte(`{"items":[{"metadata":{"name":"node-1","labels":{"gpu.bytetrade.io/cuda-supported":"true"}}}]}`))
		case strings.HasSuffix(r.URL.Path, "/v1/container"):
			_, _ = w.Write([]byte(`{"name":"task-A","status":"running","appName":"pod-1","namespace":"ns-1","podUid":"pod-1","deviceShareModes":["2"]}`))
		case strings.HasSuffix(r.URL.Path, "/instant-vector"):
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":10,"timestamp":"1745000000"}]}`))
		case strings.HasSuffix(r.URL.Path, "/range-vector"):
			_, _ = w.Write([]byte(`{"data":[{"metric":{},"values":[{"value":5,"timestamp":"1745000000"}]}]}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)

	prev := common
	cf := pkgdashboard.CommonFlags{Output: pkgdashboard.OutputJSON}
	_ = cf.Validate()
	common = &cf
	defer func() { common = prev }()

	end := time.Date(2026, 4, 28, 22, 0, 0, 0, time.UTC)
	start := end.Add(-1 * time.Hour)
	env, err := buildGPUTaskDetailFullEnvelope(context.Background(), c, "task-A", "pod-1", "2", start, end, time.Hour)
	if err != nil {
		t.Fatalf("buildGPUTaskDetailFullEnvelope: %v", err)
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

func newTestClient(srv *httptest.Server) *Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return pkgdashboard.NewClient(srv.Client(), rp)
}
