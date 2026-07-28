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

// TestBuildPartitionsEnvelope_PknameTreeAndPrefix pins the lsblk
// subtree assembly: a node hosting `sda` with two children (`sda1`,
// `sda2`) where `sda2` itself has a `sda2p1` child. The envelope
// must emit 4 items (the root + three descendants) in pre-order,
// with NAME's Display string carrying the ASCII tree prefix
// (`├── ` / `└── `) and `parent` Raw computed from `pkname`.
func TestBuildPartitionsEnvelope_PknameTreeAndPrefix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		// size / fsused are raw byte counts (lsblk -b style), as the real
		// exporter emits — the table converts them to disk units.
		// sda=1Ti, sda1=100Gi (fsused 50Gi), sda2=900Gi,
		// sda2p1=500Gi (fsused 200Gi), nvme0n1=512Gi.
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_disk_lsblk_info","data":{"result":[
            {"metric":{"node":"olares-1","name":"sda","pkname":"","size":"1099511627776","fstype":"","mountpoint":"","fsused":"","fsuse_percent":""},"value":[1714600000,"1"]},
            {"metric":{"node":"olares-1","name":"sda1","pkname":"sda","size":"107374182400","fstype":"ext4","mountpoint":"/boot","fsused":"53687091200","fsuse_percent":"50%"},"value":[1714600000,"1"]},
            {"metric":{"node":"olares-1","name":"sda2","pkname":"sda","size":"966367641600","fstype":"LVM2_member","mountpoint":"","fsused":"","fsuse_percent":""},"value":[1714600000,"1"]},
            {"metric":{"node":"olares-1","name":"sda2p1","pkname":"sda2","size":"536870912000","fstype":"ext4","mountpoint":"/","fsused":"214748364800","fsuse_percent":"40%"},"value":[1714600000,"1"]},
            {"metric":{"node":"olares-2","name":"nvme0n1","pkname":"","size":"549755813888","fstype":"","mountpoint":"","fsused":"","fsuse_percent":""},"value":[1714600000,"1"]}
          ]}}
        ]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPartitionsEnvelope(context.Background(), c, cf, "sda", time.Now())
	if err != nil {
		t.Fatalf("BuildPartitionsEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewDiskPart {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewDiskPart)
	}
	// Node auto-resolves to olares-1 (the node hosting `sda`); the
	// nvme0n1 row on olares-2 must be filtered out.
	if len(env.Items) != 4 {
		t.Fatalf("Items len = %d, want 4 (sda + sda1 + sda2 + sda2p1)", len(env.Items))
	}
	wantOrder := []string{"sda", "sda1", "sda2", "sda2p1"}
	for i, want := range wantOrder {
		if env.Items[i].Raw["name"] != want {
			t.Errorf("row %d: name = %v, want %q", i, env.Items[i].Raw["name"], want)
		}
	}
	// Tree prefix: root has no glyph; sda1 / sda2 are siblings (`├──` /
	// `└──`); sda2p1 sits under sda2 (`└──` after a continuation).
	for i, want := range []struct {
		needle string
		hint   string
	}{
		{"sda", "root carries no glyph but ends with the bare name"},
		{"── sda1", "child of sda — branch glyph + name"},
		{"── sda2", "second child of sda — branch glyph + name"},
		{"── sda2p1", "grandchild — nested branch + name"},
	} {
		got, _ := env.Items[i].Display["name"].(string)
		if !strings.Contains(got, want.needle) {
			t.Errorf("row %d Display.name = %q, missing %q (%s)", i, got, want.needle, want.hint)
		}
	}
	// Per-row parent is computed from pkname: sda2p1's parent is
	// "sda2", sda1/sda2's parent is "sda", sda's parent is "".
	if env.Items[3].Raw["parent"] != "sda2" {
		t.Errorf("sda2p1 parent = %v, want sda2", env.Items[3].Raw["parent"])
	}
	// Empty fields render as "-", populated fields pass through.
	if env.Items[2].Display["mountpoint"] != "-" {
		t.Errorf("sda2 mountpoint = %v, want \"-\"", env.Items[2].Display["mountpoint"])
	}
	if env.Items[1].Display["mountpoint"] != "/boot" {
		t.Errorf("sda1 mountpoint = %v, want /boot", env.Items[1].Display["mountpoint"])
	}
	// size / fsused run through the disk-unit converter (matching the
	// SPA's formatLsblkDiskValue auto mode): raw byte counts become
	// "1 Ti" / "100 Gi" etc., and empty fsused stays "-".
	if env.Items[0].Display["size"] != "1 Ti" {
		t.Errorf("sda size = %v, want \"1 Ti\"", env.Items[0].Display["size"])
	}
	if env.Items[1].Display["size"] != "100 Gi" {
		t.Errorf("sda1 size = %v, want \"100 Gi\"", env.Items[1].Display["size"])
	}
	if env.Items[1].Display["fsused"] != "50 Gi" {
		t.Errorf("sda1 fsused = %v, want \"50 Gi\"", env.Items[1].Display["fsused"])
	}
	if env.Items[2].Display["fsused"] != "-" {
		t.Errorf("sda2 fsused = %v, want \"-\" (empty)", env.Items[2].Display["fsused"])
	}
}
