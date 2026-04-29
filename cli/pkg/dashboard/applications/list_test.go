package applications

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
	pkgdashboard "github.com/beclab/Olares/cli/pkg/dashboard"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// newTestClient mints a *pkgdashboard.Client pointed at srv. Mirrors
// the cli/pkg/dashboard/dashboard_test.go fixture so this package
// stays self-contained: tests don't reach into the parent test file
// for helpers (which would create a build dependency on _test.go
// files in another package).
func newTestClient(srv *httptest.Server) *pkgdashboard.Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return pkgdashboard.NewClient(srv.Client(), rp)
}

// rankingFixtureServer is the canonical 3-custom-app upstream stub:
// alice owns three apps (jellyfin / nextcloud / pixelfed), none of
// them in user-space-alice so the workload fan-out only hits the
// /namespaces endpoint (the /pods path is exercised separately in
// dashboard_test.go's TestFetchWorkloadsMetrics_DualFetchPaths).
//
// The ns-level metric values are deliberately spread (CPU 2.5 / 1.5
// / 0.8) so SortBy="cpu" / SortDir="desc" produces a deterministic
// jellyfin → nextcloud → pixelfed order the assertion can pin.
func rankingFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case r.URL.Path == "/user-service/api/myapps_v2":
			_, _ = w.Write([]byte(`{
                "code":0,"message":null,"data":[
                  {"name":"jellyfin","title":"Jellyfin","namespace":"jellyfin","deployment":"jellyfin","entrances":[{"name":"web"}]},
                  {"name":"nextcloud","title":"Nextcloud","namespace":"nextcloud","deployment":"nextcloud","entrances":[{"name":"web"}]},
                  {"name":"pixelfed","title":"Pixelfed","namespace":"pixelfed","deployment":"pixelfed","entrances":[{"name":"web"}]}
                ]
            }`))
		case r.URL.Path == "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces":
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
			t.Errorf("unexpected upstream path %q", r.URL.Path)
		}
	}))
}

// fixtureFlags returns a CommonFlags whose Validate() has already run.
// CPU sort + 0 head + JSON output is the assertion-friendly default;
// individual tests override what they need.
func fixtureFlags(t *testing.T) *pkgdashboard.CommonFlags {
	t.Helper()
	cf := &pkgdashboard.CommonFlags{Timezone: format.LocalLocation()}
	if err := cf.Validate(); err != nil {
		t.Fatalf("CommonFlags.Validate: %v", err)
	}
	return cf
}

// TestRunList_SortByCPU pins the workload-grain ranking happy path:
// upstream returns three custom apps with deterministic CPU values,
// BuildListEnvelope re-tags the kind, threads RecommendedPollSeconds
// + Profile, and emits items sorted CPU-desc with the SPA's
// nine-column shape (rank/app/namespace/state/pods/cpu/memory/
// net_in/net_out) populated on every row.
func TestRunList_SortByCPU(t *testing.T) {
	srv := rankingFixtureServer(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildListEnvelope(context.Background(), c, cf, "cpu", "desc", time.Now())
	if err != nil {
		t.Fatalf("BuildListEnvelope: %v", err)
	}

	if env.Kind != pkgdashboard.KindApplicationsList {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindApplicationsList)
	}
	if env.Meta.RecommendedPollSeconds != recommendedPollSeconds {
		t.Errorf("Meta.RecommendedPollSeconds = %d, want %d",
			env.Meta.RecommendedPollSeconds, recommendedPollSeconds)
	}
	if env.Meta.Profile != "alice@olares.com" {
		t.Errorf("Meta.Profile = %q, want alice@olares.com", env.Meta.Profile)
	}
	if len(env.Items) != 3 {
		t.Fatalf("Items len = %d, want 3", len(env.Items))
	}

	wantOrder := []struct {
		app string
		cpu float64
	}{
		{"jellyfin", 2.5},
		{"nextcloud", 1.5},
		{"pixelfed", 0.8},
	}
	for i, want := range wantOrder {
		got := env.Items[i]
		if got.Raw["app"] != want.app {
			t.Errorf("row %d: app = %v, want %q", i, got.Raw["app"], want.app)
		}
		// JSON-decoded numbers come back as float64 from the
		// upstream string-encoded sample, which BuildRankingEnvelope
		// hands through to Raw["cpu"] verbatim.
		if got.Raw["cpu"] != want.cpu {
			t.Errorf("row %d: cpu = %v, want %v", i, got.Raw["cpu"], want.cpu)
		}
		// Display must carry strings for every column the
		// nine-column table renders; missing keys leak as "-" and
		// regress the SPA-aligned wire shape.
		for _, key := range []string{"rank", "app", "namespace", "state", "pods", "cpu", "memory", "net_in", "net_out"} {
			if v, ok := got.Display[key]; !ok || v == nil {
				t.Errorf("row %d: Display missing %q key", i, key)
			}
		}
	}

	// Also exercise WriteListTable so the column-order regression
	// (RANK / APP / NAMESPACE / STATE / PODS / CPU / MEMORY /
	// NET_IN / NET_OUT) is pinned in the same test.
	var buf bytes.Buffer
	if err := WriteListTable(&buf, env); err != nil {
		t.Fatalf("WriteListTable: %v", err)
	}
	out := buf.String()
	for _, header := range []string{"RANK", "APP", "NAMESPACE", "STATE", "PODS", "CPU", "MEMORY", "NET_IN", "NET_OUT"} {
		if !strings.Contains(out, header) {
			t.Errorf("table missing header %q; full output:\n%s", header, out)
		}
	}
	if !strings.Contains(out, "Jellyfin") {
		t.Errorf("table missing Jellyfin row; full output:\n%s", out)
	}
}

// TestRunList_HeadTruncates pins --head's "first-N rows after sort"
// semantic: same upstream fixture, cf.Head=2, output contains only
// the top two CPU consumers in deterministic order. The remaining
// row (pixelfed) MUST be absent — head trims after sort, never
// before, so the assertion doubles as a regression net for the
// (sort, head) ordering.
func TestRunList_HeadTruncates(t *testing.T) {
	srv := rankingFixtureServer(t)
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)
	cf.Head = 2

	env, err := BuildListEnvelope(context.Background(), c, cf, "cpu", "desc", time.Now())
	if err != nil {
		t.Fatalf("BuildListEnvelope: %v", err)
	}

	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2 (cf.Head=2)", len(env.Items))
	}
	if env.Items[0].Raw["app"] != "jellyfin" || env.Items[1].Raw["app"] != "nextcloud" {
		t.Errorf("--head=2 should keep top-2 CPU rows in order; got %v / %v",
			env.Items[0].Raw["app"], env.Items[1].Raw["app"])
	}

	// Double-check: pixelfed (lowest CPU) is fully absent from
	// the rendered table — not silently rendered as "-".
	var buf bytes.Buffer
	if err := WriteListTable(&buf, env); err != nil {
		t.Fatalf("WriteListTable: %v", err)
	}
	if strings.Contains(buf.String(), "pixelfed") || strings.Contains(buf.String(), "Pixelfed") {
		t.Errorf("--head=2 leaked the trimmed row into the table; full output:\n%s", buf.String())
	}
}

// TestValidateListFlags pins the cmd-side error wording 1:1 with the
// pre-refactor cmd RunE — the dashboard test suite asserts these
// messages verbatim, and a wording drift would silently break the
// agent-facing diagnostic surface.
func TestValidateListFlags(t *testing.T) {
	cases := []struct {
		name    string
		sortBy  string
		sortDir string
		wantErr string
	}{
		{"good", "cpu", "desc", ""},
		{"good_asc", "memory", "asc", ""},
		{"bad_sort_dir", "cpu", "sideways", `--sort: "sideways" is not asc/desc`},
		{"bad_sort_by", "uptime", "desc", `--sort-by: "uptime" is not cpu|memory|net_in|net_out`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateListFlags(tc.sortBy, tc.sortDir)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("got error %q, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
