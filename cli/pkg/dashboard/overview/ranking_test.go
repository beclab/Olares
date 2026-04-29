package overview

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
)

// rankingFixtureServer is the canonical 3-custom-app upstream stub
// for the overview-area ranking. Reuses the same wire shape as the
// applications-area fixture; kept here as a parallel copy so the
// two subpackages don't import each other's _test.go files.
func rankingFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case "/user-service/api/myapps_v2":
			_, _ = w.Write([]byte(`{
                "code":0,"message":null,"data":[
                  {"name":"jellyfin","title":"Jellyfin","namespace":"jellyfin","deployment":"jellyfin","entrances":[{"name":"web"}]},
                  {"name":"nextcloud","title":"Nextcloud","namespace":"nextcloud","deployment":"nextcloud","entrances":[{"name":"web"}]},
                  {"name":"pixelfed","title":"Pixelfed","namespace":"pixelfed","deployment":"pixelfed","entrances":[{"name":"web"}]}
                ]
            }`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"namespace_cpu_usage","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"2.5"]},
                {"metric":{"namespace":"nextcloud"},"value":[1714600000,"1.5"]},
                {"metric":{"namespace":"pixelfed"},"value":[1714600000,"0.8"]}
              ]}},
              {"metric_name":"namespace_memory_usage_wo_cache","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"536870912"]},
                {"metric":{"namespace":"nextcloud"},"value":[1714600000,"268435456"]},
                {"metric":{"namespace":"pixelfed"},"value":[1714600000,"134217728"]}
              ]}},
              {"metric_name":"namespace_net_bytes_received","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"4096"]},
                {"metric":{"namespace":"nextcloud"},"value":[1714600000,"2048"]},
                {"metric":{"namespace":"pixelfed"},"value":[1714600000,"1024"]}
              ]}},
              {"metric_name":"namespace_net_bytes_transmitted","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"8192"]},
                {"metric":{"namespace":"nextcloud"},"value":[1714600000,"4096"]},
                {"metric":{"namespace":"pixelfed"},"value":[1714600000,"2048"]}
              ]}},
              {"metric_name":"namespace_pod_count","data":{"result":[
                {"metric":{"namespace":"jellyfin"},"value":[1714600000,"3"]},
                {"metric":{"namespace":"nextcloud"},"value":[1714600000,"2"]},
                {"metric":{"namespace":"pixelfed"},"value":[1714600000,"1"]}
              ]}}
            ]}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
}

// TestBuildRankingEnvelope_DescAndAsc pins both sort directions for
// the cpu hard-coded sortBy. Asc must reverse the row order without
// re-fetching (BuildRankingEnvelope sorts in-process). The Kind is
// the shared KindOverviewRanking — both standalone leaf and the
// default sections envelope reuse it.
func TestBuildRankingEnvelope_DescAndAsc(t *testing.T) {
	srv := rankingFixtureServer(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	for _, tc := range []struct {
		dir       string
		wantOrder []string
	}{
		{"desc", []string{"jellyfin", "nextcloud", "pixelfed"}},
		{"asc", []string{"pixelfed", "nextcloud", "jellyfin"}},
	} {
		t.Run(tc.dir, func(t *testing.T) {
			env, err := BuildRankingEnvelope(context.Background(), c, cf, tc.dir, time.Now())
			if err != nil {
				t.Fatalf("BuildRankingEnvelope: %v", err)
			}
			if env.Kind != pkgdashboard.KindOverviewRanking {
				t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewRanking)
			}
			if len(env.Items) != 3 {
				t.Fatalf("Items len = %d, want 3", len(env.Items))
			}
			for i, want := range tc.wantOrder {
				if env.Items[i].Raw["app"] != want {
					t.Errorf("row %d: app = %v, want %q", i, env.Items[i].Raw["app"], want)
				}
			}
		})
	}
}

// TestRunRanking_RejectsBadSortDir pins the cmd-side error wording
// 1:1 — the test suite asserts this verbatim, and a wording drift
// would silently break the agent-facing diagnostic surface.
func TestRunRanking_RejectsBadSortDir(t *testing.T) {
	cf := fixtureFlags(t)
	err := RunRanking(context.Background(), nil, cf, "sideways")
	if err == nil {
		t.Fatal("expected --sort enum error; got nil")
	}
	if err.Error() != `--sort: "sideways" is not asc/desc` {
		t.Errorf("error = %q, want %q", err, `--sort: "sideways" is not asc/desc`)
	}
}
