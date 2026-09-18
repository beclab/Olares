package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// monitoringFixture serves a KubeSphere monitoring payload and records the
// query it was asked, so tests can assert on both the request shape and the
// decoding.
type monitoringFixture struct {
	metricsFilter   string
	resourcesFilter string
	step            string
	path            string
}

func serveMonitoring(t *testing.T, body string) (*Client, *monitoringFixture) {
	t.Helper()
	got := &monitoringFixture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		got.metricsFilter = r.URL.Query().Get("metrics_filter")
		got.resourcesFilter = r.URL.Query().Get("resources_filter")
		got.step = r.URL.Query().Get("step")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return newTestClient(srv), got
}

// A metric name that is a prefix of another must not collect its longer
// sibling: the server matches metrics_filter as a substring, so an
// unanchored name would silently average a frequency limit into a frequency
// reading.
func TestFetchGPUMetrics_AnchorsEveryName(t *testing.T) {
	c, got := serveMonitoring(t, `{"results":[]}`)

	_, err := FetchGPUMetrics(context.Background(), c,
		[]string{"intel_hw_frequency_hertz", "intel_hw_frequency_limit_hertz"},
		GPUMetricOptions{Level: GPULevelCluster})
	if err != nil {
		t.Fatalf("FetchGPUMetrics: %v", err)
	}

	want := "^(cluster_intel_hw_frequency_hertz|cluster_intel_hw_frequency_limit_hertz)$"
	if got.metricsFilter != want {
		t.Errorf("metrics_filter = %q, want %q", got.metricsFilter, want)
	}
}

// A detail view knows which node holds the card, so it asks that node rather
// than every node in the cluster.
func TestFetchGPUMetrics_NodeLevelScopesToOneNode(t *testing.T) {
	c, got := serveMonitoring(t, `{"results":[]}`)

	_, err := FetchGPUMetrics(context.Background(), c,
		[]string{"amd_gpu_health"},
		GPUMetricOptions{Level: GPULevelNode, Node: "dell"})
	if err != nil {
		t.Fatalf("FetchGPUMetrics: %v", err)
	}

	if got.path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
		t.Errorf("path = %q, want the nodes endpoint", got.path)
	}
	if got.resourcesFilter != "dell$" {
		t.Errorf("resources_filter = %q, want dell$", got.resourcesFilter)
	}
	if got.metricsFilter != "^(node_amd_gpu_health)$" {
		t.Errorf("metrics_filter = %q, want the node-prefixed name", got.metricsFilter)
	}
}

// The level prefix is an artefact of which endpoint answered, so callers
// name metrics the same way either way.
func TestFetchGPUMetrics_StripsLevelPrefix(t *testing.T) {
	c, _ := serveMonitoring(t, `{"results":[{"metric_name":"cluster_intel_hw_power_watts",
		"data":{"result":[{"metric":{"hw_id":"a"},"value":[1700000000,"56.5"]}]}}]}`)

	table, err := FetchGPUMetrics(context.Background(), c,
		[]string{"intel_hw_power_watts"}, GPUMetricOptions{Level: GPULevelCluster})
	if err != nil {
		t.Fatalf("FetchGPUMetrics: %v", err)
	}
	series, ok := table["intel_hw_power_watts"]
	if !ok {
		t.Fatalf("table keyed by %v, want the unprefixed name", keysOf(table))
	}
	if len(series) != 1 || len(series[0].Points) != 1 || series[0].Points[0].Value != 56.5 {
		t.Errorf("series = %+v, want one point of 56.5", series)
	}
}

func keysOf(t MetricTable) []string {
	out := make([]string, 0, len(t))
	for k := range t {
		out = append(out, k)
	}
	return out
}

func instantSeries(labels map[string]string, value float64) MetricSeries {
	return MetricSeries{Labels: labels, Points: []GPUPoint{{TS: 1700000000, Value: value}}}
}

// Intel's "GPU utilization" is the busier of the compute and render
// pipelines, so a panel listing both label values must take the max rather
// than matching only the first.
func TestReadInstant_MatchesAnyOfSeveralLabelValues(t *testing.T) {
	table := MetricTable{
		"intel_hw_gpu_utilization_ratio": {
			instantSeries(map[string]string{"hw_gpu_task": "compute-all"}, 0.9),
			instantSeries(map[string]string{"hw_gpu_task": "render-all"}, 0.2),
			instantSeries(map[string]string{"hw_gpu_task": "media-all"}, 1),
		},
	}

	got, ok := ReadInstant(table, intelShaderUtilization, nil)
	if !ok {
		t.Fatal("ReadInstant found no series")
	}
	// media-all is the highest of the three but is not one of the two the
	// read asks for; taking it would report an idle card as saturated.
	if got != 90 {
		t.Errorf("utilization = %v, want 90", got)
	}
}

// The exporter overshoots its own 0-1 range by a rounding error, which would
// otherwise render as 100.01%.
func TestReadInstant_ClampsAtTheStatedCeiling(t *testing.T) {
	table := MetricTable{
		"intel_hw_gpu_utilization_ratio": {
			instantSeries(map[string]string{"hw_gpu_task": "compute-all"}, 1.00014),
		},
	}

	got, _ := ReadInstant(table, intelShaderUtilization, nil)
	if got != 100 {
		t.Errorf("utilization = %v, want it clamped to 100", got)
	}
}

// An unreported metric has to stay distinguishable from a reported zero: one
// means the column is unavailable, the other that the card is idle.
func TestReadInstant_ReportsNothingRatherThanZero(t *testing.T) {
	table := MetricTable{"intel_hw_power_watts": {
		instantSeries(map[string]string{"hw_sensor_location": "package"}, 63),
	}}
	read := MetricRead{
		Metric: "intel_hw_power_watts",
		Match:  map[string][]string{"hw_sensor_location": {"card"}},
		Agg:    AggAvg,
	}

	if _, ok := ReadInstant(table, read, nil); ok {
		t.Error("a non-matching label set should report nothing, not 0")
	}
}

// A timestamp the denominator never reported is dropped rather than filled
// in — a fabricated point shows up as a real dip in the chart.
func TestRatioRange_DropsUnpairedTimestamps(t *testing.T) {
	got := RatioRange(
		[]GPUPoint{{TS: 1, Value: 5}, {TS: 2, Value: 5}},
		[]GPUPoint{{TS: 1, Value: 10}},
	)
	if len(got) != 1 || got[0].TS != 1 || got[0].Value != 50 {
		t.Errorf("ratio = %+v, want a single 50%% point at t=1", got)
	}
}

// hw_id carries no serial, so two identical cards in the same slot on
// different nodes share it. Without the node in the key they would collapse
// into one row and their readings would be averaged together.
func TestJoinDeviceRows_KeepsSameIDOnDifferentNodesApart(t *testing.T) {
	spec := intelKsSpec()
	table := MetricTable{
		"intel_hw_gpu_info": {
			instantSeries(map[string]string{"hw_id": "card0", "node": "a", "hw_model": "Arc Pro B70"}, 1),
			instantSeries(map[string]string{"hw_id": "card0", "node": "b", "hw_model": "Arc Pro B70"}, 1),
		},
		"intel_hw_power_watts": {
			instantSeries(map[string]string{"hw_id": "card0", "node": "a", "hw_sensor_location": "card"}, 100),
			instantSeries(map[string]string{"hw_id": "card0", "node": "b", "hw_sensor_location": "card"}, 50),
		},
	}

	rows := JoinDeviceRows(VendorIntel, spec, table)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want one per node", len(rows))
	}
	byNode := map[string]float64{}
	for _, row := range rows {
		v, _ := RowFloat(row, "power")
		byNode[rowString(row, "deviceNode")] = v
	}
	if byNode["a"] != 100 || byNode["b"] != 50 {
		t.Errorf("power by node = %v, want a=100 b=50", byNode)
	}
}

// The AMD exporter publishes no memory utilisation metric, so the column is
// computed from the two it does publish.
func TestJoinDeviceRows_DerivesAMDMemoryPercent(t *testing.T) {
	spec := amdKsSpec()
	labels := map[string]string{"gpu_uuid": "u1", "node": "dell", "card_model": "Radeon AI PRO R9700"}
	table := MetricTable{
		"amd_gpu_health":     {instantSeries(labels, 1)},
		"amd_gpu_total_vram": {instantSeries(labels, 32624)},
		"amd_gpu_used_vram":  {instantSeries(labels, 8156)},
	}

	rows := JoinDeviceRows(VendorAMD, spec, table)
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	pct, ok := RowFloat(rows[0], "memoryUtilizedPercent")
	if !ok || pct < 24.9 || pct > 25.1 {
		t.Errorf("memoryUtilizedPercent = %v (ok=%v), want ~25", pct, ok)
	}
	if healthy, _ := rows[0]["health"].(bool); !healthy {
		t.Error("a card reporting health=1 should read healthy")
	}
}

// An unhealthy AMD card should say which signal tripped rather than showing
// a bare unhealthy label.
func TestJoinDeviceRows_ExplainsAMDECCFailure(t *testing.T) {
	spec := amdKsSpec()
	labels := map[string]string{"gpu_uuid": "u1", "node": "dell"}
	table := MetricTable{
		"amd_gpu_health":              {instantSeries(labels, 0)},
		"amd_gpu_ecc_uncorrect_total": {instantSeries(labels, 3)},
	}

	rows := JoinDeviceRows(VendorAMD, spec, table)
	if healthy, _ := rows[0]["health"].(bool); healthy {
		t.Error("a card reporting health=0 should read unhealthy")
	}
	if reason := rowString(rows[0], "healthReason"); reason == "" {
		t.Error("an ECC-failed card should carry a reason")
	}
}

// Intel's health is the absence of a reset request. A card with no
// reset_needed series never asked for one, and `ecc_disabled` is a
// configuration rather than a fault.
func TestJoinDeviceRows_IntelHealthyUntilResetRequested(t *testing.T) {
	spec := intelKsSpec()
	labels := map[string]string{"hw_id": "card0", "node": "olares"}
	healthy := JoinDeviceRows(VendorIntel, spec, MetricTable{
		"intel_hw_gpu_info": {instantSeries(labels, 1)},
	})
	if ok, _ := healthy[0]["health"].(bool); !ok {
		t.Error("no reset_needed series should read healthy")
	}

	failed := JoinDeviceRows(VendorIntel, spec, MetricTable{
		"intel_hw_gpu_info": {instantSeries(labels, 1)},
		"intel_hw_status": {instantSeries(map[string]string{
			"hw_id": "card0", "node": "olares", "hw_type": "gpu", "hw_state": "reset_needed",
		}, 1)},
	})
	if ok, _ := failed[0]["health"].(bool); ok {
		t.Error("a card asking for a reset should read unhealthy")
	}
}

// The vendor inventory drives which exporters get queried, so it has to
// recognise both the per-mode labels and the older cuda-supported marker —
// this binary outlives the cluster it is pointed at.
func TestVendorsOfNodeLabels(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   []GPUVendor
	}{
		{"no labels", map[string]string{}, nil},
		{"intel discrete", map[string]string{"gpu.bytetrade.io/intel-gpu": "true"}, []GPUVendor{VendorIntel}},
		{"amd discrete", map[string]string{"gpu.bytetrade.io/amd-gpu": "true"}, []GPUVendor{VendorAMD}},
		{"nvidia", map[string]string{"gpu.bytetrade.io/nvidia": "true"}, []GPUVendor{VendorNVIDIA}},
		{
			// A node labelled by an Olares older than the per-mode labels
			// carries nothing else, and this binary still has to see its card.
			"legacy cuda marker",
			map[string]string{"gpu.bytetrade.io/cuda-supported": "true"},
			[]GPUVendor{VendorNVIDIA},
		},
		{
			"nvidia labelled both ways is still one vendor",
			map[string]string{"gpu.bytetrade.io/nvidia": "true", "gpu.bytetrade.io/cuda-supported": "true"},
			[]GPUVendor{VendorNVIDIA},
		},
		{
			"two vendors on one node",
			map[string]string{"gpu.bytetrade.io/intel-gpu": "true", "gpu.bytetrade.io/amd-gpu": "true"},
			[]GPUVendor{VendorIntel, VendorAMD},
		},
		{
			// An integrated GPU has no series to render: xpumd drops
			// hw_gpu_type=integrated and the AMD exporter is discrete-only.
			"integrated-only mode is not a vendor",
			map[string]string{"gpu.bytetrade.io/intel": "true"},
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := VendorsOfNodeLabels(tc.labels)
			if len(got) != len(tc.want) {
				t.Fatalf("vendors = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("vendors = %v, want %v", got, tc.want)
				}
			}
		})
	}
}
