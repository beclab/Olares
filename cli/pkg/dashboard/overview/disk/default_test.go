package disk

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

// TestBuildSectionsEnvelope_Smoke pins the disk-default sections
// fan-out: a single SMART row drives one per-device partitions
// fetch, the envelope carries `main` + `partitions` (with the
// device-keyed inner Sections map), and FetchedAt populates on
// each section so a JSON consumer can compute per-section freshness.
func TestBuildSectionsEnvelope_Smoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		// SMART row + matching aux + lsblk for sda all on one
		// /v1alpha3/nodes endpoint — the upstream fold-in lets the
		// stub stay compact. metricsFilter is irrelevant here
		// because we always return the same superset.
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_disk_smartctl_info","data":{"result":[
            {"metric":{"node":"olares-1","device":"sda","name":"sda","rotational":"0","logical_block_size":"4096","physical_block_size":"4096","health_ok":"true","capacity":"1099511627776","model":"WDC","serial":"S1","protocol":"SATA","firmware":"1.0"},"value":[1714600000,"1"]}
          ]}},
          {"metric_name":"node_one_disk_capacity_size","data":{"result":[
            {"metric":{"node":"olares-1","device":"sda"},"value":[1714600000,"1099000000000"]}
          ]}},
          {"metric_name":"node_one_disk_avail_size","data":{"result":[
            {"metric":{"node":"olares-1","device":"sda"},"value":[1714600000,"549000000000"]}
          ]}},
          {"metric_name":"node_disk_temp_celsius","data":{"result":[
            {"metric":{"node":"olares-1","device":"sda"},"value":[1714600000,"35"]}
          ]}},
          {"metric_name":"node_disk_lsblk_info","data":{"result":[
            {"metric":{"node":"olares-1","name":"sda","pkname":"","size":"1T","fstype":"","mountpoint":"","fsused":"","fsuse_percent":""},"value":[1714600000,"1"]},
            {"metric":{"node":"olares-1","name":"sda1","pkname":"sda","size":"100G","fstype":"ext4","mountpoint":"/","fsused":"50G","fsuse_percent":"50%"},"value":[1714600000,"1"]}
          ]}}
        ]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env := BuildSectionsEnvelope(context.Background(), c, cf, time.Now())
	if env.Kind != pkgdashboard.KindOverviewDisk {
		t.Errorf("parent Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewDisk)
	}
	main, ok := env.Sections["main"]
	if !ok {
		t.Fatal("section `main` missing")
	}
	if main.Meta.Error != "" {
		t.Errorf("main.Meta.Error = %q, want empty", main.Meta.Error)
	}
	if len(main.Items) != 1 {
		t.Errorf("main Items len = %d, want 1", len(main.Items))
	}
	parts, ok := env.Sections["partitions"]
	if !ok {
		t.Fatal("section `partitions` missing")
	}
	pSda, ok := parts.Sections["sda"]
	if !ok {
		t.Fatalf("partitions.Sections[\"sda\"] missing; got keys %v", keysOf(parts.Sections))
	}
	if pSda.Meta.FetchedAt == "" {
		t.Errorf("partitions.sda.Meta.FetchedAt is empty")
	}
	if len(pSda.Items) != 2 {
		t.Errorf("partitions.sda Items len = %d, want 2 (sda + sda1)", len(pSda.Items))
	}
}

// TestWriteSectionsTable_BannersAndOrder pins the human-readable
// scrollback layout: a "== MAIN ==" banner followed by per-device
// "== PARTITIONS: <device> ==" banners in alphabetical order.
func TestWriteSectionsTable_BannersAndOrder(t *testing.T) {
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverviewDisk,
		Sections: map[string]pkgdashboard.Envelope{
			"main": {Kind: pkgdashboard.KindOverviewDiskMain, Items: []pkgdashboard.Item{
				{Display: map[string]any{"device": "sda", "node": "olares-1", "type": "SSD", "health": "Normal", "total": "1Ti"}},
			}},
			"partitions": {Kind: pkgdashboard.KindOverviewDiskPart, Sections: map[string]pkgdashboard.Envelope{
				"nvme0n1": {Items: []pkgdashboard.Item{
					{Display: map[string]any{"name": "nvme0n1", "size": "512G"}},
				}},
				"sda": {Items: []pkgdashboard.Item{
					{Display: map[string]any{"name": "sda", "size": "1T"}},
				}},
			}},
		},
	}
	var buf bytes.Buffer
	if err := WriteSectionsTable(&buf, env); err != nil {
		t.Fatalf("WriteSectionsTable: %v", err)
	}
	out := buf.String()
	mainIdx := strings.Index(out, "== MAIN ==")
	nvmeIdx := strings.Index(out, "== PARTITIONS: nvme0n1 ==")
	sdaIdx := strings.Index(out, "== PARTITIONS: sda ==")
	if mainIdx < 0 || nvmeIdx < 0 || sdaIdx < 0 {
		t.Fatalf("missing banner; full output:\n%s", out)
	}
	if !(mainIdx < nvmeIdx && nvmeIdx < sdaIdx) {
		t.Errorf("banners out of order (main → nvme0n1 → sda); full output:\n%s", out)
	}
}

func keysOf(m map[string]pkgdashboard.Envelope) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
