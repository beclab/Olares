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
