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

// defaultFixtureServer is the canonical multi-section stub: serves
// /capi/app/detail (admin alice) + /kapis/cluster (physical) +
// /kapis/users/<u> (user) + /myapps_v2 + /kapis/namespaces
// (ranking). Lets BuildSectionsEnvelope's three goroutines all
// succeed so the test exercises the happy-path stitching without
// having to trigger error fan-in.
func defaultFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case r.URL.Path == "/user-service/api/myapps_v2":
			_, _ = w.Write([]byte(`{"code":0,"message":null,"data":[
              {"name":"jellyfin","title":"Jellyfin","namespace":"jellyfin","deployment":"jellyfin","entrances":[{"name":"web"}]}
            ]}`))
		case r.URL.Path == "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"cluster_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"2"]}]}},
              {"metric_name":"cluster_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"8"]}]}},
              {"metric_name":"cluster_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"8589934592"]}]}}
            ]}`))
		case r.URL.Path == "/kapis/monitoring.kubesphere.io/v1alpha3/users/alice":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"user_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"4"]}]}},
              {"metric_name":"user_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"1"]}]}},
              {"metric_name":"user_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"4294967296"]}]}}
            ]}`))
		case r.URL.Path == "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"namespace_cpu_usage","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"1.5"]}
              ]}}
            ]}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
}

// TestBuildSectionsEnvelope_Smoke pins the three-section happy
// path: parent envelope's Kind is KindOverview, the Sections map
// carries physical/user/ranking with the expected per-section
// kinds, and FetchedAt is populated on every section (so a JSON
// consumer can compute end-to-end latency per section).
func TestBuildSectionsEnvelope_Smoke(t *testing.T) {
	srv := defaultFixtureServer(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env := BuildSectionsEnvelope(context.Background(), c, cf, time.Now())
	if env.Kind != pkgdashboard.KindOverview {
		t.Errorf("parent Kind = %q, want %q", env.Kind, pkgdashboard.KindOverview)
	}
	if env.Items != nil {
		t.Errorf("parent Items must be nil (sections envelope); got %v", env.Items)
	}
	for _, key := range []string{"physical", "user", "ranking"} {
		section, ok := env.Sections[key]
		if !ok {
			t.Errorf("section %q missing from envelope", key)
			continue
		}
		if section.Meta.FetchedAt == "" {
			t.Errorf("section %q: Meta.FetchedAt is empty", key)
		}
		if section.Meta.Error != "" {
			t.Errorf("section %q: unexpected error %q", key, section.Meta.Error)
		}
	}
	wantKinds := map[string]string{
		"physical": pkgdashboard.KindOverviewPhysical,
		"user":     pkgdashboard.KindOverviewUser,
		"ranking":  pkgdashboard.KindOverviewRanking,
	}
	for key, kind := range wantKinds {
		if env.Sections[key].Kind != kind {
			t.Errorf("section %q: Kind = %q, want %q", key, env.Sections[key].Kind, kind)
		}
	}
}

// TestWriteSectionsTable_BannersAndOrder pins the human-readable
// scrollback layout: every section gets a "== KEY ==" banner in
// the canonical iteration order, and the rendering survives
// individual sections being absent (defensive: BuildSectionsEnvelope
// always emits all three, but a future failure-mode might not).
func TestWriteSectionsTable_BannersAndOrder(t *testing.T) {
	env := pkgdashboard.Envelope{
		Kind: pkgdashboard.KindOverview,
		Sections: map[string]pkgdashboard.Envelope{
			"physical": {Kind: pkgdashboard.KindOverviewPhysical, Items: []pkgdashboard.Item{
				{Display: map[string]any{"metric": "CPU", "value": "1 / 8", "utilisation": "12.5%"}},
			}},
			"user": {Kind: pkgdashboard.KindOverviewUser, Items: []pkgdashboard.Item{
				{Display: map[string]any{"metric": "CPU", "used": "0.5", "total": "4", "utilisation": "12.5%"}},
			}},
			"ranking": {Kind: pkgdashboard.KindOverviewRanking, Items: []pkgdashboard.Item{
				{Display: map[string]any{"rank": "1", "app": "jellyfin", "namespace": "jellyfin", "cpu": "1.5"}},
			}},
		},
	}
	var buf bytes.Buffer
	if err := WriteSectionsTable(&buf, env); err != nil {
		t.Fatalf("WriteSectionsTable: %v", err)
	}
	out := buf.String()
	for _, banner := range []string{"== PHYSICAL ==", "== USER ==", "== RANKING =="} {
		if !strings.Contains(out, banner) {
			t.Errorf("missing banner %q; full output:\n%s", banner, out)
		}
	}
	pIdx := strings.Index(out, "== PHYSICAL ==")
	uIdx := strings.Index(out, "== USER ==")
	rIdx := strings.Index(out, "== RANKING ==")
	if !(pIdx < uIdx && uIdx < rIdx) {
		t.Errorf("section banners out of order; full output:\n%s", out)
	}
}
