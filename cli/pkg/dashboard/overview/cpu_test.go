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

// TestRunCPU_DisplayShapeAndTempUnit pins the per-node CPU row
// shape: every cpuColumns() key is present on Display, the temp
// renders in the cf-selected unit (Fahrenheit here so it's
// distinguishable from the Celsius default), and frequency is the
// SPA's "X.YZ GHz" shape rather than a raw Hz number.
func TestRunCPU_DisplayShapeAndTempUnit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_cpu_total","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"8"]}]}},
          {"metric_name":"node_cpu_utilisation","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"0.42"]}]}},
          {"metric_name":"node_user_cpu_usage","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"0.3"]}]}},
          {"metric_name":"node_system_cpu_usage","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"0.1"]}]}},
          {"metric_name":"node_iowait_cpu_usage","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"0.02"]}]}},
          {"metric_name":"node_load1","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"1.5"]}]}},
          {"metric_name":"node_load5","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"1.2"]}]}},
          {"metric_name":"node_load15","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"1.0"]}]}},
          {"metric_name":"node_cpu_temp_celsius","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"50"]}]}},
          {"metric_name":"node_cpu_base_frequency_hertz_max","data":{"result":[{"metric":{"node":"olares-1"},"value":[1714600000,"3500000000"]}]}}
        ]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.TempUnit = "F"

	disp := cpuDisplayFn(cf)
	env, err := BuildPerNodeEnvelope(context.Background(), c, cf,
		pkgdashboard.KindOverviewCPU, cpuMetricSet(), disp, time.Now())
	if err != nil {
		t.Fatalf("BuildPerNodeEnvelope: %v", err)
	}
	if len(env.Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(env.Items))
	}
	row := env.Items[0]
	for _, key := range []string{"node", "freq", "cores", "cpu_util", "user", "system", "iowait", "load1", "load5", "load15", "temp"} {
		if v, ok := row.Display[key]; !ok || v == nil {
			t.Errorf("Display missing %q key", key)
		}
	}
	if !strings.Contains(row.Display["temp"].(string), "F") {
		t.Errorf("temp = %q, want a Fahrenheit suffix", row.Display["temp"])
	}
	if !strings.Contains(row.Display["freq"].(string), "GHz") {
		t.Errorf("freq = %q, want a 'GHz' suffix on a 3.5 GHz cpu", row.Display["freq"])
	}
	if row.Display["cpu_util"] != "42%" {
		t.Errorf("cpu_util = %v, want 42%% (sampleFloat 0.42 → percentString)", row.Display["cpu_util"])
	}
}
