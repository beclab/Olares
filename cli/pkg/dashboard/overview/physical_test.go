package overview

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// physicalClusterMetricsHandler writes the canonical 13-metric
// cluster body the physical envelope builder reads. Pulled into a
// helper so both the baseline test and the augmented (GPU + fan)
// test can reuse the same payload without duplicating fixture text.
func physicalClusterMetricsHandler(w http.ResponseWriter) {
	_, _ = w.Write([]byte(`{"results":[
      {"metric_name":"cluster_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"4.2"]}]}},
      {"metric_name":"cluster_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"16"]}]}},
      {"metric_name":"cluster_cpu_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.2625"]}]}},
      {"metric_name":"cluster_memory_usage_wo_cache","data":{"result":[{"metric":{},"value":[1714600000,"4294967296"]}]}},
      {"metric_name":"cluster_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"17179869184"]}]}},
      {"metric_name":"cluster_memory_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.25"]}]}},
      {"metric_name":"cluster_disk_size_usage","data":{"result":[{"metric":{},"value":[1714600000,"107374182400"]}]}},
      {"metric_name":"cluster_disk_size_capacity","data":{"result":[{"metric":{},"value":[1714600000,"536870912000"]}]}},
      {"metric_name":"cluster_disk_size_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.2"]}]}},
      {"metric_name":"cluster_pod_running_count","data":{"result":[{"metric":{},"value":[1714600000,"42"]}]}},
      {"metric_name":"cluster_pod_quota","data":{"result":[{"metric":{},"value":[1714600000,"110"]}]}},
      {"metric_name":"cluster_net_bytes_received","data":{"result":[{"metric":{},"value":[1714600000,"4096"]}]}},
      {"metric_name":"cluster_net_bytes_transmitted","data":{"result":[{"metric":{},"value":[1714600000,"8192"]}]}}
    ]}`))
}

// TestBuildPhysicalEnvelope_SixRowsAndUnitInference pins the 6-row
// baseline shape (cpu / memory / disk / pods / net_in / net_out)
// and the SPA-aligned unit-inference behaviour. The GPU + fan
// optional rows are explicitly stubbed as 404 so the test exercises
// the "no GPU, not Olares One" baseline (every row is from cluster
// monitoring).
//
// Wire shape: GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster
// returning the standard `{ results: [{ metric_name, data: { result:
// [{ value: [ts, "X"] }] } }] }` envelope.
func TestBuildPhysicalEnvelope_SixRowsAndUnitInference(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			physicalClusterMetricsHandler(w)
		// Optional GPU/fan endpoints — return 404 so the rows
		// are skipped (matches "HAMI not installed, not Olares
		// One" production wire shape).
		case "/hami/api/vgpu/v1/monitor/query/instant-vector",
			"/hami/api/vgpu/v1/gpus",
			"/user-service/api/system/status",
			"/user-service/api/mdns/olares-one/cpu-gpu":
			w.WriteHeader(http.StatusNotFound)
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPhysicalEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildPhysicalEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewPhysical {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewPhysical)
	}
	if len(env.Items) != 6 {
		t.Fatalf("Items len = %d, want 6 (cpu/memory/disk/pods/net_in/net_out — GPU/fan should be skipped on 404)", len(env.Items))
	}

	wantOrder := []string{"cpu", "memory", "disk", "pods", "net_in", "net_out"}
	for i, want := range wantOrder {
		if env.Items[i].Raw["metric"] != want {
			t.Errorf("row %d: metric = %v, want %q", i, env.Items[i].Raw["metric"], want)
		}
	}

	// Memory row's `value` Display string should infer the GiB
	// magnitude on a 16 GiB total (SPA's GetDiskSize returns the
	// short "Gi" suffix once the magnitude crosses 1 GiB).
	memValue, _ := env.Items[1].Display["value"].(string)
	if !strings.Contains(memValue, "Gi") {
		t.Errorf("memory display = %q, want a 'Gi' suffix once total > 1 GiB", memValue)
	}

	// Net rows: detail string is a SPA-style "X B/s" / "X KB/s"
	// throughput rather than a raw number.
	netInDetail, _ := env.Items[4].Display["detail"].(string)
	if !strings.Contains(netInDetail, "/s") {
		t.Errorf("net_in detail = %q, want a 'B/s'-style throughput suffix", netInDetail)
	}
}

// gpuInstantVectorHandler routes the two SPA-aligned instant-vector
// queries (`hami_memory_used` / `hami_memory_size` aggregated by
// instance) to MiB values. Returns an empty `data` array for any
// other query so unrelated callers don't accidentally pick up
// these stubs.
//
// Note: the SPA's actual queries divide by 1024 (`/ 1024`) to land
// in GiB; we match the *un-divided* MiB form because that's what
// `gpuSummaryUsedQuery` / `gpuSummaryTotalQuery` send. The CLI then
// multiplies by 1024 * 1024 to land in bytes so format.GetDiskSize
// can pick the right Gi/Ti suffix.
func gpuInstantVectorHandler(w http.ResponseWriter, body string, usedMiB, totalMiB float64) {
	switch {
	case strings.Contains(body, "hami_memory_used"):
		_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":` +
			strconv.FormatFloat(usedMiB, 'f', -1, 64) + `,"timestamp":"2026-05-25T00:00:00Z"}]}`))
	case strings.Contains(body, "hami_memory_size"):
		_, _ = w.Write([]byte(`{"data":[{"metric":{},"value":` +
			strconv.FormatFloat(totalMiB, 'f', -1, 64) + `,"timestamp":"2026-05-25T00:00:00Z"}]}`))
	default:
		_, _ = w.Write([]byte(`{"data":[]}`))
	}
}

// TestBuildPhysicalEnvelope_AugmentsWithGPUAndFanRows pins the
// SPA-aligned augmentation behaviour: when HAMI returns a non-empty
// GPU list AND the device reports as Olares One, the envelope adds
// `gpu`, `fan_cpu`, `fan_gpu` rows after the 6 baseline rows. The
// SPA's `Overview2/ClusterResource.vue` does the same — the GPU
// card is appended only when `hasGPU.value` is truthy and the fan
// card only when `FanStore.isOlaresOneDevice`.
//
// Regression net for the production bug where
// `olares-cli dashboard overview` and
// `olares-cli dashboard overview physical` both omitted GPU + fan
// even though the SPA showed those panels.
//
// The GPU `used` value comes from the SPA-aligned
// `avg(sum(hami_memory_used) by (instance))` instant query (NOT
// the /v1/gpus list aggregation, which only counts vGPU-allocated
// VRAM — see physical.go:buildGPUSummaryRow for the rationale).
func TestBuildPhysicalEnvelope_AugmentsWithGPUAndFanRows(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			physicalClusterMetricsHandler(w)
		case "/hami/api/vgpu/v1/monitor/query/instant-vector":
			// 270 MiB used / 24576 MiB (24 GiB) total —
			// matches what the SPA cluster card shows for
			// olarestest004's RTX 5090.
			body, _ := io.ReadAll(r.Body)
			gpuInstantVectorHandler(w, string(body), 270, 24576)
		case "/hami/api/vgpu/v1/gpus":
			// Fallback path: same total but `memoryUsed=0`
			// (the realistic vGPU-allocation-table number;
			// the prom path is preferred precisely because
			// it covers non-vGPU CUDA processes too).
			_, _ = w.Write([]byte(`{"list":[{
              "uuid":"GPU-e5d26177","type":"NVIDIA RTX 5090","health":true,
              "memoryUsed":0,"memoryTotal":24576,"shareMode":"2",
              "nodeName":"olares","power":7.6,"powerLimit":175,
              "temperature":44,"coreUtilizedPercent":0
            }]}`))
		case "/user-service/api/system/status":
			// `device_name = "Olares One"` triggers
			// `IsOlaresOne()` -> true (the only branch that
			// activates the fan rows).
			_, _ = w.Write([]byte(`{"code":0,"data":{"device_name":"Olares One"}}`))
		case "/user-service/api/mdns/olares-one/cpu-gpu":
			_, _ = w.Write([]byte(`{"data":{
              "cpu_fan_speed":1100,"cpu_temperature":55,
              "gpu_fan_speed":1300,"gpu_temperature":48
            }}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPhysicalEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildPhysicalEnvelope: %v", err)
	}
	if len(env.Items) != 9 {
		t.Fatalf("Items len = %d, want 9 (6 baseline + gpu + fan_cpu + fan_gpu)", len(env.Items))
	}
	wantOrder := []string{"cpu", "memory", "disk", "pods", "net_in", "net_out", "gpu", "fan_cpu", "fan_gpu"}
	for i, want := range wantOrder {
		if env.Items[i].Raw["metric"] != want {
			t.Errorf("row %d: metric = %v, want %q", i, env.Items[i].Raw["metric"], want)
		}
	}

	// GPU row: prom path picked the 270 MiB used / 24 GiB total
	// values — Display "value" should carry the Gi suffix and
	// the used segment must NOT be the "-" empty placeholder
	// (regression for the production bug where the row
	// rendered "- / 23.89Gi" because the list-only path saw
	// memoryUsed=0 from the allocation table).
	gpuValue, _ := env.Items[6].Display["value"].(string)
	if !strings.Contains(gpuValue, "Gi") {
		t.Errorf("gpu value = %q, want 'Gi' suffix at 24 GiB total", gpuValue)
	}
	if strings.HasPrefix(gpuValue, "- /") {
		t.Errorf("gpu value = %q, used segment must not be '-' when prom path returned a non-zero used value", gpuValue)
	}

	// Fan rows: detail uses RPM with a "/" separator like
	// the SPA's "1100 / 2900 RPM" cell.
	cpuFanDetail, _ := env.Items[7].Display["detail"].(string)
	if !strings.Contains(cpuFanDetail, "RPM") || !strings.Contains(cpuFanDetail, "/") {
		t.Errorf("fan_cpu detail = %q, want '<rpm> / <max> RPM'", cpuFanDetail)
	}
	gpuFanDetail, _ := env.Items[8].Display["detail"].(string)
	if !strings.Contains(gpuFanDetail, "RPM") || !strings.Contains(gpuFanDetail, "/") {
		t.Errorf("fan_gpu detail = %q, want '<rpm> / <max> RPM'", gpuFanDetail)
	}
}

// TestBuildPhysicalEnvelope_GPUFallsBackToListWhenPromEmpty pins
// the prom-fallback path: when the SPA-aligned instant-vector
// queries succeed but return an empty `data: []` (HAMI prom is
// reachable but has no series for this cluster yet), the GPU row
// builder falls back to aggregating /v1/gpus. The row still ships
// — just with the allocation-table numbers rather than prom's
// real-VRAM figure.
func TestBuildPhysicalEnvelope_GPUFallsBackToListWhenPromEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			physicalClusterMetricsHandler(w)
		case "/hami/api/vgpu/v1/monitor/query/instant-vector":
			// Prom up, no series — caller should fall back.
			_, _ = w.Write([]byte(`{"data":[]}`))
		case "/hami/api/vgpu/v1/gpus":
			_, _ = w.Write([]byte(`{"list":[{
              "uuid":"GPU-e5d26177","memoryUsed":1024,"memoryTotal":24576
            }]}`))
		case "/user-service/api/system/status",
			"/user-service/api/mdns/olares-one/cpu-gpu":
			w.WriteHeader(http.StatusNotFound)
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPhysicalEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildPhysicalEnvelope: %v", err)
	}
	// Expect 7 rows: 6 baseline + gpu (no fan, off-Olares-One).
	if len(env.Items) != 7 {
		t.Fatalf("Items len = %d, want 7 (6 baseline + gpu via list-fallback)", len(env.Items))
	}
	gpuRow := env.Items[6]
	if gpuRow.Raw["metric"] != "gpu" {
		t.Fatalf("row 6 metric = %v, want gpu", gpuRow.Raw["metric"])
	}
	// 1024 MiB used == 1 GiB; format.GetDiskSize converts and
	// returns a Gi-suffixed string.
	gpuValue, _ := gpuRow.Display["value"].(string)
	if !strings.Contains(gpuValue, "Gi") {
		t.Errorf("gpu value = %q, want 'Gi' suffix from list-fallback aggregation", gpuValue)
	}
}

// TestWritePhysicalTable_HeaderAndRowOrder pins the 3-column table
// (METRIC / VALUE / UTIL) header order — agents that scrape the
// human view rely on the column-position index being stable.
func TestWritePhysicalTable_HeaderAndRowOrder(t *testing.T) {
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewPhysical,
		Items: []pkgdashboard.Item{
			{Display: map[string]any{"metric": "CPU", "value": "1.00 / 16.00", "utilisation": "6.25%"}},
			{Display: map[string]any{"metric": "Memory", "value": "4 GiB / 16 GiB", "utilisation": "25%"}},
		},
	}
	var buf bytes.Buffer
	if err := WritePhysicalTable(&buf, env); err != nil {
		t.Fatalf("WritePhysicalTable: %v", err)
	}
	out := buf.String()
	for _, header := range []string{"METRIC", "VALUE", "UTIL"} {
		if !strings.Contains(out, header) {
			t.Errorf("table missing header %q; full output:\n%s", header, out)
		}
	}
	cpuIdx := strings.Index(out, "CPU")
	memIdx := strings.Index(out, "Memory")
	if cpuIdx < 0 || memIdx < 0 || cpuIdx > memIdx {
		t.Errorf("rows out of order or missing; full output:\n%s", out)
	}
}
