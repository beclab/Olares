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

// TestBuildNetworkEnvelope_TwoIfaces stubs /capi/system/ifs with
// two NICs and pins:
//   - one Item per iface, in upstream order (no internal sort);
//   - synthetic "Port-N" column counts up from 1, regardless of
//     iface name (the SPA's NIC index — agents need it stable);
//   - status = "up" iff InternetConnected, "down" otherwise;
//   - rate columns go through formatRateAny which yields
//     SPA-style "X B/s" suffix on numeric inputs.
func TestBuildNetworkEnvelope_TwoIfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/system/ifs":
			_, _ = w.Write([]byte(`[
              {"iface":"eth0","method":"dhcp","mtu":1500,"hostname":"olares","ip":"10.0.0.5","ipv4Mask":"255.255.255.0","ipv4Gateway":"10.0.0.1","internetConnected":true,"txRate":1024,"rxRate":2048},
              {"iface":"wlan0","method":"manual","mtu":1500,"hostname":"olares","ip":"","internetConnected":false,"txRate":0,"rxRate":0}
            ]`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildNetworkEnvelope(context.Background(), c, cf, true, time.Now())
	if err != nil {
		t.Fatalf("BuildNetworkEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewNetwork {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewNetwork)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2", len(env.Items))
	}
	if env.Items[0].Raw["port"] != "Port-1" || env.Items[1].Raw["port"] != "Port-2" {
		t.Errorf("ports = %v / %v, want Port-1 / Port-2",
			env.Items[0].Raw["port"], env.Items[1].Raw["port"])
	}
	if env.Items[0].Raw["status"] != "up" {
		t.Errorf("eth0 status = %v, want up", env.Items[0].Raw["status"])
	}
	if env.Items[1].Raw["status"] != "down" {
		t.Errorf("wlan0 status = %v, want down (InternetConnected=false)", env.Items[1].Raw["status"])
	}
	tx, _ := env.Items[0].Display["tx"].(string)
	if !strings.Contains(tx, "/s") {
		t.Errorf("eth0 tx = %q, want a 'B/s'-style throughput suffix", tx)
	}
}
