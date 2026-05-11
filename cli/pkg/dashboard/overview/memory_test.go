package overview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// memorySrv stubs /v1alpha3/nodes returning the union of physical
// + swap metric series. Both modes hit the same endpoint with
// different metricsFilter values; the upstream answers everything,
// the scaffold's metrics filter only requests the relevant subset
// per mode.
func memorySrv(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_memory_total","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"17179869184"]}]}},
          {"metric_name":"node_memory_usage_wo_cache","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"4294967296"]}]}},
          {"metric_name":"node_memory_available","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"12884901888"]}]}},
          {"metric_name":"node_memory_utilisation","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"0.25"]}]}},
          {"metric_name":"node_memory_cached","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"1073741824"]}]}},
          {"metric_name":"node_memory_buffers","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"536870912"]}]}},
          {"metric_name":"node_memory_swap_total","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"4294967296"]}]}},
          {"metric_name":"node_memory_swap_used","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"1073741824"]}]}},
          {"metric_name":"node_memory_pgpgin_rate","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"100"]}]}},
          {"metric_name":"node_memory_pgpgout_rate","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"50"]}]}}
        ]}`))
	}))
}

// TestMemoryPhysicalMode pins the SPA's per-node memory shape:
// 6-key Display (total/used/avail/buffers/cached/util) + Raw.mode
// = "physical" so JSON consumers can demux when both modes get
// piped into one stream.
func TestMemoryPhysicalMode(t *testing.T) {
	srv := memorySrv(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPerNodeEnvelope(context.Background(), c, cf,
		pkgdashboard.KindOverviewMemory, memoryPhysicalMetricSet(),
		memoryPhysicalDisplay, time.Now())
	if err != nil {
		t.Fatalf("BuildPerNodeEnvelope (physical): %v", err)
	}
	if len(env.Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(env.Items))
	}
	row := env.Items[0]
	if row.Raw["mode"] != "physical" {
		t.Errorf("Raw.mode = %v, want physical", row.Raw["mode"])
	}
	for _, key := range []string{"total", "used", "avail", "buffers", "cached", "util"} {
		if v, ok := row.Display[key]; !ok || v == nil {
			t.Errorf("Display missing %q key", key)
		}
	}
	if total, _ := row.Display["total"].(string); !strings.Contains(total, "Gi") {
		t.Errorf("total = %q, want a 'Gi' suffix on 16 GiB", total)
	}
	if row.Display["util"] != "25%" {
		t.Errorf("util = %v, want 25%%", row.Display["util"])
	}
}

// TestMemorySwapMode pins the swap-mode divergence: Raw.mode =
// "swap", PG_IN / PG_OUT columns populated, and util computed
// locally as used/total since swap_utilisation isn't a separate
// upstream series.
func TestMemorySwapMode(t *testing.T) {
	srv := memorySrv(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPerNodeEnvelope(context.Background(), c, cf,
		pkgdashboard.KindOverviewMemory, memorySwapMetricSet(),
		memorySwapDisplay, time.Now())
	if err != nil {
		t.Fatalf("BuildPerNodeEnvelope (swap): %v", err)
	}
	if len(env.Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(env.Items))
	}
	row := env.Items[0]
	if row.Raw["mode"] != "swap" {
		t.Errorf("Raw.mode = %v, want swap", row.Raw["mode"])
	}
	for _, key := range []string{"total", "used", "pg_in", "pg_out", "util"} {
		if v, ok := row.Display[key]; !ok || v == nil {
			t.Errorf("Display missing %q key", key)
		}
	}
	// 1 GiB used / 4 GiB total = 25% — locally computed by
	// safeRatio rather than a dedicated upstream metric.
	if row.Display["util"] != "25%" {
		t.Errorf("util = %v, want 25%% (used/total locally)", row.Display["util"])
	}
}

// TestRunMemory_RejectsBadMode pins the cmd-side error wording 1:1.
func TestRunMemory_RejectsBadMode(t *testing.T) {
	cf := fixtureFlags(t)
	err := RunMemory(context.Background(), nil, cf, "ramdisk")
	if err == nil {
		t.Fatal("expected --mode enum error; got nil")
	}
	if err.Error() != `--mode: "ramdisk" must be physical or swap` {
		t.Errorf("error = %q, want %q", err, `--mode: "ramdisk" must be physical or swap`)
	}
}
