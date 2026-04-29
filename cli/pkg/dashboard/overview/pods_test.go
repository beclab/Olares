package overview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// TestRunPods_RowShape pins the 4-column per-node pod-count
// envelope. Running / quota are integer-formatted so a JSON
// consumer that reads Display verbatim doesn't have to re-trim a
// "%.2f" decimal point; util is a percent string.
func TestRunPods_RowShape(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kapis/monitoring.kubesphere.io/v1alpha3/nodes" {
			noUnexpectedPath(t, w, r.URL.Path)
			return
		}
		_, _ = w.Write([]byte(`{"results":[
          {"metric_name":"node_pod_running_count","data":{"result":[
            {"metric":{"node":"olares-1"},"value":[1714600000,"32"]},
            {"metric":{"node":"olares-2"},"value":[1714600000,"7"]}
          ]}},
          {"metric_name":"node_pod_quota","data":{"result":[
            {"metric":{"node":"olares-1"},"value":[1714600000,"110"]},
            {"metric":{"node":"olares-2"},"value":[1714600000,"110"]}
          ]}},
          {"metric_name":"node_pod_utilisation","data":{"result":[
            {"metric":{"node":"olares-1"},"value":[1714600000,"0.291"]},
            {"metric":{"node":"olares-2"},"value":[1714600000,"0.064"]}
          ]}}
        ]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildPerNodeEnvelope(context.Background(), c, cf,
		pkgdashboard.KindOverviewPods, podsMetricSet(), podsDisplay, time.Now())
	if err != nil {
		t.Fatalf("BuildPerNodeEnvelope: %v", err)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2", len(env.Items))
	}
	for _, key := range []string{"node", "running", "quota", "util"} {
		if v, ok := env.Items[0].Display[key]; !ok || v == nil {
			t.Errorf("Display missing %q key", key)
		}
	}
	// Integer-formatted ("32"), not "32.00".
	if env.Items[0].Display["running"] != "32" {
		t.Errorf("running display = %v, want \"32\" (integer-formatted)", env.Items[0].Display["running"])
	}
	if env.Items[0].Display["quota"] != "110" {
		t.Errorf("quota display = %v, want \"110\"", env.Items[0].Display["quota"])
	}
	// percentString trims trailing zeros: 0.291 → "29.1%".
	if env.Items[0].Display["util"] != "29.1%" {
		t.Errorf("util display = %v, want \"29.1%%\"", env.Items[0].Display["util"])
	}
}
