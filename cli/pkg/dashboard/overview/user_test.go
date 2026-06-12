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

// userFixtureServer stubs the active-profile probe (/capi/app/detail)
// + the user-grain monitoring fetch. role / username / quota values
// are knobs the tests vary to exercise admin / non-admin branches.
//
// The cluster monitoring endpoint is stubbed too — the admin-total
// fallback path (BuildUserEnvelope, see Bug 3 in the dashboard
// area) issues a parallel `cluster_cpu_total` /
// `cluster_memory_total` query when the resolved user is admin.
// We respond with a value the admin tests can pin against
// (24 cores / 96 GiB) but the legacy "user has its own quota"
// tests mask it via the user-grain numbers.
func userFixtureServer(t *testing.T, role, username string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"` + username + `","globalrole":"` + role + `"}}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/users/" + username:
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"user_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"4"]}]}},
              {"metric_name":"user_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"1.5"]}]}},
              {"metric_name":"user_cpu_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.375"]}]}},
              {"metric_name":"user_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"4294967296"]}]}},
              {"metric_name":"user_memory_usage_wo_cache","data":{"result":[{"metric":{},"value":[1714600000,"1073741824"]}]}},
              {"metric_name":"user_memory_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.25"]}]}}
            ]}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"cluster_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"24"]}]}},
              {"metric_name":"cluster_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"103079215104"]}]}}
            ]}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
}

// userFixtureServerNoUserQuota stubs the user-grain endpoint
// returning all-zero `user_cpu_total` / `user_memory_total` (the
// real-world admin-without-ResourceQuota wire shape — `kube_user_*`
// PromQL has no series, monitoring fills with 0). Cluster endpoint
// still returns 24 cores / 96 GiB so the fallback path can pin
// against those.
func userFixtureServerNoUserQuota(t *testing.T, role, username string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"` + username + `","globalrole":"` + role + `"}}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/users/" + username:
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"user_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"0"]}]}},
              {"metric_name":"user_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"1.5"]}]}},
              {"metric_name":"user_cpu_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0"]}]}},
              {"metric_name":"user_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"0"]}]}},
              {"metric_name":"user_memory_usage_wo_cache","data":{"result":[{"metric":{},"value":[1714600000,"1073741824"]}]}},
              {"metric_name":"user_memory_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0"]}]}}
            ]}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"cluster_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"24"]}]}},
              {"metric_name":"cluster_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"103079215104"]}]}}
            ]}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
}

// TestBuildUserEnvelope_SelfTargetingHappyPath: empty target falls
// back to the active profile. The 2-row CPU/Memory envelope must
// carry the resolved username on every row's Raw["user"] (so a JSON
// consumer can join across users without re-querying).
func TestBuildUserEnvelope_SelfTargetingHappyPath(t *testing.T) {
	srv := userFixtureServer(t, "platform-admin", "alice")
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildUserEnvelope(context.Background(), c, cf, "", time.Now())
	if err != nil {
		t.Fatalf("BuildUserEnvelope: %v", err)
	}
	if env.Kind != pkgdashboard.KindOverviewUser {
		t.Errorf("Kind = %q, want %q", env.Kind, pkgdashboard.KindOverviewUser)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2 (CPU/Memory)", len(env.Items))
	}
	if env.Items[0].Raw["metric"] != "CPU" || env.Items[1].Raw["metric"] != "Memory" {
		t.Errorf("row order = %v / %v, want CPU / Memory",
			env.Items[0].Raw["metric"], env.Items[1].Raw["metric"])
	}
	if env.Items[0].Raw["user"] != "alice" {
		t.Errorf("Raw.user = %v, want alice", env.Items[0].Raw["user"])
	}

	// CPU row's Display["used"] should render as "1.50" (always
	// 2-decimal, never the cluster's K-suffix path).
	if env.Items[0].Display["used"] != "1.50" {
		t.Errorf("CPU used display = %v, want 1.50", env.Items[0].Display["used"])
	}
	// Memory row's Display["used"] should infer GiB (used =
	// 1073741824 bytes = 1 GiB).
	memUsed, _ := env.Items[1].Display["used"].(string)
	if !strings.Contains(memUsed, "Gi") {
		t.Errorf("memory used display = %q, want a 'Gi' suffix at 1 GiB", memUsed)
	}
}

// TestBuildUserEnvelope_NonAdminTargetingPeerRejected: a
// workspaces-manager (non-admin) targeting another user MUST get a
// typed admin-required error rather than silently rendering peer
// data. ResolveTargetUser owns the gate; the test pins that we
// surface it 1:1 (error wording matters — agent diagnostic
// stability).
func TestBuildUserEnvelope_NonAdminTargetingPeerRejected(t *testing.T) {
	srv := userFixtureServer(t, "workspaces-manager", "bob")
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	_, err := BuildUserEnvelope(context.Background(), c, cf, "carol", time.Now())
	if err == nil {
		t.Fatal("expected admin-required error for non-admin targeting a peer; got nil")
	}
	if !strings.Contains(err.Error(), "platform-admin") {
		t.Errorf("error %q does not mention 'platform-admin'", err)
	}
}

// TestBuildUserEnvelope_AdminWithoutQuotaFallsBackToCluster pins
// the SPA-aligned admin-total fallback: when `user_cpu_total` /
// `user_memory_total` are zero (typical for platform admins
// without an explicit ResourceQuota), the envelope total must
// switch to `cluster_cpu_total` / `cluster_memory_total` and
// `Raw["total_source"]` must flip to "cluster_total" so agents
// can detect the source change without re-running the heuristic.
//
// Regression net for the production bug where
// `olares-cli dashboard overview user` rendered "1.5 / 0" CPU
// for admins because user-grain totals were zero and the CLI
// did not mirror the SPA's
// (Overview2/IndexPage.vue:cluster_cpu_total) fallback.
func TestBuildUserEnvelope_AdminWithoutQuotaFallsBackToCluster(t *testing.T) {
	srv := userFixtureServerNoUserQuota(t, "platform-admin", "alice")
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildUserEnvelope(context.Background(), c, cf, "", time.Now())
	if err != nil {
		t.Fatalf("BuildUserEnvelope: %v", err)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2", len(env.Items))
	}

	cpuRow := env.Items[0]
	if cpuRow.Display["total"] != "24.00" {
		t.Errorf("CPU total display = %v, want 24.00 (cluster fallback)", cpuRow.Display["total"])
	}
	if cpuRow.Raw["total_source"] != "cluster_total" {
		t.Errorf("CPU Raw.total_source = %v, want cluster_total", cpuRow.Raw["total_source"])
	}
	// 1.5 used / 24 total ≈ 6.25%
	if cpuRow.Display["utilisation"] != "6.25%" {
		t.Errorf("CPU utilisation = %v, want 6.25%% (recomputed against cluster total)", cpuRow.Display["utilisation"])
	}

	memRow := env.Items[1]
	memTotal, _ := memRow.Display["total"].(string)
	if !strings.Contains(memTotal, "Gi") {
		t.Errorf("Memory total display = %q, want a 'Gi' suffix at 96 GiB cluster total", memTotal)
	}
	if memRow.Raw["total_source"] != "cluster_total" {
		t.Errorf("Memory Raw.total_source = %v, want cluster_total", memRow.Raw["total_source"])
	}
}

// TestBuildUserEnvelope_AdminTargetingPeerKeepsPeerQuota pins the
// cross-user variant of the admin-total fallback. When an admin
// (alice) targets a different user (`--user bob`),
// `ResolveTargetUser` returns a synthetic record carrying bob's
// name + `GlobalRole="<admin-target>"` — i.e. `user.IsAdmin()` is
// false. The admin-total fallback MUST stay off in that case:
// "alice is allowed to look at bob's stats" never implies "bob's
// quota should be replaced by the cluster total". Otherwise the
// CLI silently inflates bob's totals to whatever the cluster owns,
// hiding the real ResourceQuota and making `--user bob` useless
// for capacity planning.
//
// Regression net for the bug where the EnsureUser re-probe was
// firing on the active profile (alice→admin) and bumping
// `isAdmin=true` for bob; the fix gates the re-probe on
// `active.Name == user.Name` so the bump only fires for the
// self-target path (where ResolveTargetUser already returned the
// real record with a real role anyway — the re-probe is now pure
// defense-in-depth).
func TestBuildUserEnvelope_AdminTargetingPeerKeepsPeerQuota(t *testing.T) {
	// `bob` has an explicit ResourceQuota (8 CPU / 16 GiB). The
	// cluster total is 24 CPU / 96 GiB — picking the bug would
	// inflate bob's CPU to 24 and Memory to 96 GiB. The test pins
	// the non-inflated values + `total_source=user_quota`.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/users/bob":
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"user_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"8"]}]}},
              {"metric_name":"user_cpu_usage","data":{"result":[{"metric":{},"value":[1714600000,"2"]}]}},
              {"metric_name":"user_cpu_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.25"]}]}},
              {"metric_name":"user_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"17179869184"]}]}},
              {"metric_name":"user_memory_usage_wo_cache","data":{"result":[{"metric":{},"value":[1714600000,"4294967296"]}]}},
              {"metric_name":"user_memory_utilisation","data":{"result":[{"metric":{},"value":[1714600000,"0.25"]}]}}
            ]}`))
		case "/kapis/monitoring.kubesphere.io/v1alpha3/cluster":
			// If the bug fires this gets read and clobbers bob's totals.
			_, _ = w.Write([]byte(`{"results":[
              {"metric_name":"cluster_cpu_total","data":{"result":[{"metric":{},"value":[1714600000,"24"]}]}},
              {"metric_name":"cluster_memory_total","data":{"result":[{"metric":{},"value":[1714600000,"103079215104"]}]}}
            ]}`))
		default:
			noUnexpectedPath(t, w, r.URL.Path)
		}
	}))
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildUserEnvelope(context.Background(), c, cf, "bob", time.Now())
	if err != nil {
		t.Fatalf("BuildUserEnvelope: %v", err)
	}
	if len(env.Items) != 2 {
		t.Fatalf("Items len = %d, want 2", len(env.Items))
	}

	cpuRow := env.Items[0]
	if cpuRow.Raw["user"] != "bob" {
		t.Errorf("CPU Raw.user = %v, want bob (admin-target path)", cpuRow.Raw["user"])
	}
	if cpuRow.Raw["total_source"] != "user_quota" {
		t.Errorf("CPU Raw.total_source = %v, want user_quota (admin targeting peer must NOT activate cluster fallback)", cpuRow.Raw["total_source"])
	}
	if cpuRow.Display["total"] != "8.00" {
		t.Errorf("CPU total display = %v, want 8.00 (bob's quota, NOT cluster's 24)", cpuRow.Display["total"])
	}
	// 2/8 = 0.25 → percentString → "25%" (integer branch of
	// formatNumberJSLike: trunc(v)==v so 0-decimal). If the bug
	// fires the utilisation gets recomputed against cluster 24 →
	// "8.33%".
	if cpuRow.Display["utilisation"] != "25%" {
		t.Errorf("CPU utilisation = %v, want 25%% (2/8, NOT recomputed against cluster 24)", cpuRow.Display["utilisation"])
	}

	memRow := env.Items[1]
	if memRow.Raw["total_source"] != "user_quota" {
		t.Errorf("Memory Raw.total_source = %v, want user_quota", memRow.Raw["total_source"])
	}
	memTotal, _ := memRow.Display["total"].(string)
	// 16 GiB exact (17179869184 bytes / 2^30 = 16) — GetDiskSize
	// renders integer values without decimals via formatNumberJSLike
	// (`v == Trunc(v)` branch). Pin the exact rendered string so a
	// loose substring check doesn't accidentally pass against the
	// cluster's "96Gi" or "9.00Gi" partial overlap.
	if memTotal != "16Gi" {
		t.Errorf("Memory total display = %q, want %q (bob's quota, NOT cluster's 96 GiB)", memTotal, "16Gi")
	}
}

// TestBuildUserEnvelope_NonAdminPrefersUserQuota pins the inverse
// of the admin fallback: a non-admin (`workspaces-manager`)
// querying themselves keeps `total_source = user_quota` even when
// cluster_cpu_total would also be available — non-admins must NOT
// see cluster totals (that's the SPA's IndexPage.vue gate
// `appDetail.isAdmin ? ... : undefined`).
func TestBuildUserEnvelope_NonAdminPrefersUserQuota(t *testing.T) {
	srv := userFixtureServer(t, "workspaces-manager", "bob")
	defer srv.Close()
	c := newTestClient(srv)
	cf := fixtureFlags(t)

	env, err := BuildUserEnvelope(context.Background(), c, cf, "", time.Now())
	if err != nil {
		t.Fatalf("BuildUserEnvelope: %v", err)
	}
	if env.Items[0].Raw["total_source"] != "user_quota" {
		t.Errorf("CPU Raw.total_source = %v, want user_quota for non-admin",
			env.Items[0].Raw["total_source"])
	}
	if env.Items[0].Display["total"] != "4.00" {
		t.Errorf("CPU total = %v, want 4.00 (user quota, NOT cluster total)",
			env.Items[0].Display["total"])
	}
}
