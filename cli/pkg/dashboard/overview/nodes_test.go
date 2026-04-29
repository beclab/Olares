package overview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// TestBuildPerNodeEnvelope_GroupsAndSortsByNode pins the scaffold's
// invariants without committing to a particular leaf's metric set:
// (a) rows are bucketed by the `node` label across all metric_name
// entries, (b) the resulting Items are sorted alphabetically by
// node, (c) entries with no `node` AND no `instance` label are
// dropped (rather than silently bucketed under "").
func TestBuildPerNodeEnvelope_GroupsAndSortsByNode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		// metric `m_a` has rows for nodes node-bravo + node-alpha
		// (out of order on purpose). metric `m_b` adds node-bravo
		// again + a row whose label-set has neither `node` nor
		// `instance` (must be dropped).
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"m_a","data":{"result":[
            {"metric":{"node":"node-bravo"},"value":[1714600000,"7"]},
            {"metric":{"node":"node-alpha"},"value":[1714600000,"3"]}
          ]}},
          {"metric_name":"m_b","data":{"result":[
            {"metric":{"node":"node-bravo"},"value":[1714600000,"42"]},
            {"metric":{},"value":[1714600000,"99"]}
          ]}}
        ]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	disp := func(node string, last map[string]format.LastMonitoringSample) (map[string]any, map[string]any) {
		raw := map[string]any{
			"node": node,
			"a":    sampleFloat(last["m_a"]),
			"b":    sampleFloat(last["m_b"]),
		}
		return raw, map[string]any{"node": node}
	}
	env, err := BuildPerNodeEnvelope(context.Background(), c, cf,
		pkgdashboard.KindOverviewCPU, []string{"m_a", "m_b"}, disp, time.Now())
	if err != nil {
		t.Fatalf("BuildPerNodeEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewCPU {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewCPU)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2 (label-less row dropped)", len(env.Items))
	}
	if env.Items[0].Raw["node"] != "node-alpha" || env.Items[1].Raw["node"] != "node-bravo" {
		t.Errorf("row order = %v / %v, want node-alpha / node-bravo (alphabetical)",
			env.Items[0].Raw["node"], env.Items[1].Raw["node"])
	}
	// node-bravo carries m_b=42; node-alpha has no m_b row at all,
	// so sampleFloat(empty) must yield 0 rather than NaN.
	if env.Items[0].Raw["b"] != 0.0 {
		t.Errorf("node-alpha b = %v, want 0 (missing metric)", env.Items[0].Raw["b"])
	}
	if env.Items[1].Raw["b"] != 42.0 {
		t.Errorf("node-bravo b = %v, want 42", env.Items[1].Raw["b"])
	}
}
