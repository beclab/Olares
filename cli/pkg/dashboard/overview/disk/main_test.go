package disk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// diskMainFixtureSrv stubs /v1alpha3/nodes returning two SMART
// rows + matching auxiliary samples for both. node-bravo on purpose
// sorts AFTER node-alpha so the deterministic-order assertion bites
// (BuildMainEnvelope sorts by (node, device)). One disk is HDD
// (rotational=1), one is SSD (rotational=0); one is healthy, one
// reports a SMART exception — exercises both branches of the
// type/health enum derivation.
func diskMainFixtureSrv(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_disk_smartctl_info","data":{"result":[
            {"metric":{"node":"node-bravo","device":"sda","name":"sda","rotational":"1","logical_block_size":"4096","physical_block_size":"4096","health_ok":"true","capacity":"1099511627776","model":"WDC-XYZ","serial":"S1","protocol":"SATA","firmware":"1.0"},"value":[1714600000,"1"]},
            {"metric":{"node":"node-alpha","device":"nvme0n1","name":"nvme0n1","rotational":"0","logical_block_size":"4096","physical_block_size":"4096","health_ok":"false","capacity":"512000000000","model":"Samsung-SSD","serial":"S2","protocol":"NVMe","firmware":"2.1"},"value":[1714600000,"1"]}
          ]}},
          {"metric_name":"node_one_disk_capacity_size","data":{"result":[
            {"metric":{"node":"node-alpha","device":"nvme0n1"},"value":[1714600000,"500000000000"]},
            {"metric":{"node":"node-bravo","device":"sda"},"value":[1714600000,"1099000000000"]}
          ]}},
          {"metric_name":"node_one_disk_avail_size","data":{"result":[
            {"metric":{"node":"node-alpha","device":"nvme0n1"},"value":[1714600000,"125000000000"]},
            {"metric":{"node":"node-bravo","device":"sda"},"value":[1714600000,"549000000000"]}
          ]}},
          {"metric_name":"node_disk_temp_celsius","data":{"result":[
            {"metric":{"node":"node-alpha","device":"nvme0n1"},"value":[1714600000,"42"]},
            {"metric":{"node":"node-bravo","device":"sda"},"value":[1714600000,"35"]}
          ]}},
          {"metric_name":"node_disk_power_on_hours","data":{"result":[
            {"metric":{"node":"node-alpha","device":"nvme0n1"},"value":[1714600000,"500"]}
          ]}},
          {"metric_name":"node_one_disk_data_bytes_written","data":{"result":[
            {"metric":{"node":"node-alpha","device":"nvme0n1"},"value":[1714600000,"107374182400"]},
            {"metric":{"node":"node-bravo","device":"sda"},"value":[1714600000,"5497558138880"]}
          ]}}
        ]}`))
	}))
}

// TestBuildMainEnvelope_SortsAndJoinsAuxiliaries is the canonical
// happy-path test for the disk-main aggregator:
//   - SMART rows are sorted by (node, device) — assertion pins
//     node-alpha/nvme0n1 ahead of node-bravo/sda;
//   - the (device, node) join with auxiliary metrics works through
//     findAux's `strings.Contains`-style matching;
//   - SSD/HDD enum derives from `rotational`;
//   - health enum surfaces "Exception" for health_ok=false;
//   - a missing power_on_hours sample on node-bravo prints "-"
//     rather than "0h" (Empty=true codepath in
//     renderHoursOrDash).
func TestBuildMainEnvelope_SortsAndJoinsAuxiliaries(t *testing.T) {
	srv := diskMainFixtureSrv(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildMainEnvelope(context.Background(), c, cf, time.Now())
	if err != nil {
		t.Fatalf("BuildMainEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewDiskMain {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewDiskMain)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2", len(env.Items))
	}
	row0, row1 := env.Items[0], env.Items[1]
	if row0.Raw["device"] != "nvme0n1" || row1.Raw["device"] != "sda" {
		t.Errorf("sort order = %v / %v, want nvme0n1 / sda (node-alpha < node-bravo)",
			row0.Raw["device"], row1.Raw["device"])
	}
	if row0.Raw["type"] != "SSD" || row1.Raw["type"] != "HDD" {
		t.Errorf("type = %v / %v, want SSD / HDD", row0.Raw["type"], row1.Raw["type"])
	}
	if row0.Raw["health_ok"] != false || row1.Raw["health_ok"] != true {
		t.Errorf("health_ok = %v / %v, want false / true", row0.Raw["health_ok"], row1.Raw["health_ok"])
	}
	if row0.Display["health"] != "Exception" || row1.Display["health"] != "Normal" {
		t.Errorf("health display = %v / %v, want Exception / Normal",
			row0.Display["health"], row1.Display["health"])
	}
	// nvme0n1: cap=500e9, avail=125e9 → used=375e9; ratio = 0.75.
	// The SPA card does not render utilisation, so it is absent from
	// Display but preserved in Raw for JSON consumers.
	if _, ok := row0.Display["util"]; ok {
		t.Errorf("util should be absent from Display (SPA card does not render it)")
	}
	if r, _ := row0.Raw["used_ratio"].(float64); r != 0.75 {
		t.Errorf("nvme0n1 used_ratio = %v, want 0.75 (used/cap)", row0.Raw["used_ratio"])
	}
	// temperature: dual-unit "<c>°C/<f>°F" like the SPA card.
	if row0.Display["temperature"] != "42°C/107.6°F" {
		t.Errorf("nvme0n1 temperature = %v, want 42°C/107.6°F", row0.Display["temperature"])
	}
	// power_on_hours: nvme0n1 has 500h, sda has no row → "-".
	if poh, _ := row0.Display["power_on_hours"].(string); !strings.Contains(poh, "500") {
		t.Errorf("nvme0n1 power_on = %q, want a '500h'-shaped string", poh)
	}
	if row1.Display["power_on_hours"] != "-" {
		t.Errorf("sda power_on = %v, want \"-\" (Empty sample)", row1.Display["power_on_hours"])
	}
	// is_4k_native: SSD with logical_block=4096 → Yes.
	if row0.Display["is_4k_native"] != "Yes" {
		t.Errorf("nvme0n1 4k_native = %v, want Yes", row0.Display["is_4k_native"])
	}
}

// TestRenderDiskTemperature_DualUnitAndDash pins the disk-area
// dual-unit temperature cell. Empty samples (Celsius == 0) render
// as "-/-" to mirror the SPA's config.ts:218–223 behaviour; non-zero
// values always emit both Celsius and Fahrenheit regardless of
// --temp-unit, matching the SPA card which hardcodes "<c>°C/<f>°F".
func TestRenderDiskTemperature_DualUnitAndDash(t *testing.T) {
	if got := renderDiskTemperature(0); got != "-/-" {
		t.Errorf("zero celsius should print '-/-', got %q", got)
	}
	if got := renderDiskTemperature(40); got != "40°C/104°F" {
		t.Errorf("40C → %q, want 40°C/104°F", got)
	}
	if got := renderDiskTemperature(35); got != "35°C/95°F" {
		t.Errorf("35C → %q, want 35°C/95°F", got)
	}
}
