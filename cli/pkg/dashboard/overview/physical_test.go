package overview

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestBuildPhysicalEnvelope_SixRowsAndUnitInference pins the 6-row
// shape (cpu / memory / disk / pods / net_in / net_out) and the
// SPA-aligned unit-inference behaviour (memory rows over 1 GiB
// render in GiB, large net throughput renders in MB/s, etc.).
//
// Wire shape: GET /kapis/monitoring.kubesphere.io/v1alpha3/cluster
// returning the standard `{ results: [{ metric_name, data: { result:
// [{ value: [ts, "X"] }] } }] }` envelope.
func TestBuildPhysicalEnvelope_SixRowsAndUnitInference(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/cluster" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		// 13 metric_name entries — DerivePhysicalRows reads them
		// individually so missing metrics yield 0 rather than aborting.
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
		t.Fatalf("Items len = %d, want 6 (cpu/memory/disk/pods/net_in/net_out)", len(env.Items))
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
