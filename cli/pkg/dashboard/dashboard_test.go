package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/dashboard/format"
)

// ----------------------------------------------------------------------------
// CommonFlags.Validate
// ----------------------------------------------------------------------------

func TestCommonFlags_Validate_DefaultsAreSane(t *testing.T) {
	cf := &CommonFlags{}
	if err := cf.Validate(); err != nil {
		t.Fatalf("Validate(defaults) returned error: %v", err)
	}
	if cf.Output != OutputTable {
		t.Errorf("default Output = %q, want %q", cf.Output, OutputTable)
	}
	if cf.TempUnit != format.TempC {
		t.Errorf("default TempUnit = %q, want C", cf.TempUnit)
	}
	if cf.Timezone == nil {
		t.Errorf("default Timezone is nil; expected local")
	}
}

func TestCommonFlags_Validate_RejectsBadOutput(t *testing.T) {
	cf := &CommonFlags{}
	cf.OutputRaw = "yaml"
	if err := cf.Validate(); err == nil {
		t.Fatal("Validate(--output yaml) should fail; got nil")
	}
}

func TestCommonFlags_Validate_RejectsBadTempUnit(t *testing.T) {
	cf := &CommonFlags{}
	cf.OutputRaw = "table"
	cf.TempUnitRaw = "Z"
	if err := cf.Validate(); err == nil {
		t.Fatal("Validate(--temp-unit Z) should fail; got nil")
	}
}

func TestCommonFlags_Validate_SinceVsAbsoluteWindow(t *testing.T) {
	cases := []struct {
		name      string
		raw       func(*CommonFlags)
		wantError bool
	}{
		{
			name: "since alone is fine",
			raw: func(cf *CommonFlags) {
				cf.OutputRaw = "table"
				cf.TempUnitRaw = "C"
				cf.SinceRaw = "5m"
			},
		},
		{
			name: "start+end alone is fine",
			raw: func(cf *CommonFlags) {
				cf.OutputRaw = "table"
				cf.TempUnitRaw = "C"
				cf.StartRaw = "2025-04-01T10:00:00Z"
				cf.EndRaw = "2025-04-01T10:30:00Z"
			},
		},
		{
			name: "since + start is rejected",
			raw: func(cf *CommonFlags) {
				cf.OutputRaw = "table"
				cf.TempUnitRaw = "C"
				cf.SinceRaw = "5m"
				cf.StartRaw = "2025-04-01T10:00:00Z"
				cf.EndRaw = "2025-04-01T10:30:00Z"
			},
			wantError: true,
		},
		{
			name: "start without end is rejected",
			raw: func(cf *CommonFlags) {
				cf.OutputRaw = "table"
				cf.TempUnitRaw = "C"
				cf.StartRaw = "2025-04-01T10:00:00Z"
			},
			wantError: true,
		},
		{
			name: "start >= end is rejected",
			raw: func(cf *CommonFlags) {
				cf.OutputRaw = "table"
				cf.TempUnitRaw = "C"
				cf.StartRaw = "2025-04-01T10:30:00Z"
				cf.EndRaw = "2025-04-01T10:00:00Z"
			},
			wantError: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cf := &CommonFlags{}
			tc.raw(cf)
			err := cf.Validate()
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCommonFlags_Validate_WatchFlagsRequireWatch(t *testing.T) {
	cf := &CommonFlags{}
	cf.OutputRaw = "table"
	cf.TempUnitRaw = "C"
	cf.WatchIterations = 5
	if err := cf.Validate(); err == nil {
		t.Fatal("--watch-iterations without --watch should fail; got nil")
	}
}

func TestCommonFlags_ResolveWindow_AbsoluteSticksAcrossIterations(t *testing.T) {
	cf := &CommonFlags{
		Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC),
	}
	for i := 0; i < 5; i++ {
		s, e := cf.ResolveWindow(time.Now(), 30*time.Second)
		if !s.Equal(cf.Start) || !e.Equal(cf.End) {
			t.Fatalf("iteration %d: got [%s, %s], want absolute window", i, s, e)
		}
	}
}

func TestCommonFlags_ResolveWindow_SlidesWithSinceWhenWatch(t *testing.T) {
	cf := &CommonFlags{Since: 5 * time.Minute}
	now1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	now2 := now1.Add(5 * time.Minute)
	s1, e1 := cf.ResolveWindow(now1, 0)
	s2, e2 := cf.ResolveWindow(now2, 0)
	if !s2.After(s1) || !e2.After(e1) {
		t.Fatalf("expected sliding window: [%s,%s] then [%s,%s]", s1, e1, s2, e2)
	}
	if e1.Sub(s1) != 5*time.Minute {
		t.Fatalf("window width = %s, want 5m", e1.Sub(s1))
	}
}

// ----------------------------------------------------------------------------
// EnsureUser caching
// ----------------------------------------------------------------------------

func TestClient_EnsureUser_CallsOnceAndCachesResult(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/capi/app/detail" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"clusterRole":"cluster-admin","user":{"username":"alice","globalrole":"platform-admin"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	for i := 0; i < 3; i++ {
		u, err := c.EnsureUser(context.Background())
		if err != nil {
			t.Fatalf("EnsureUser iter %d: %v", i, err)
		}
		if u.Name != "alice" {
			t.Fatalf("iter %d: name = %q, want alice", i, u.Name)
		}
		if !u.IsAdmin() {
			t.Fatalf("iter %d: IsAdmin() = false, want true", i)
		}
	}
	if hits != 1 {
		t.Fatalf("EnsureUser hit upstream %d times, want 1 (sync.Once)", hits)
	}
}

func TestClient_RequireAdmin_RejectsNonAdmin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"user":{"username":"bob","globalrole":"workspaces-manager"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.RequireAdmin(context.Background())
	if err == nil {
		t.Fatal("RequireAdmin should reject non-admin; got nil")
	}
	if !strings.Contains(err.Error(), "platform-admin") {
		t.Errorf("error %q should mention platform-admin", err)
	}
}

// ----------------------------------------------------------------------------
// Output / NDJSON shape
// ----------------------------------------------------------------------------

func TestWriteJSON_EmitsLeafEnvelope(t *testing.T) {
	env := Envelope{
		Kind: KindOverviewCPU,
		Meta: Meta{FetchedAt: "2025-01-01T00:00:00Z"},
		Items: []Item{
			{Raw: map[string]any{"x": 1}, Display: map[string]any{"x": "1"}},
		},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, env); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := buf.Bytes()
	if !bytes.HasSuffix(out, []byte("\n")) {
		t.Errorf("output should end with newline (NDJSON discipline); got %q", out)
	}
	var parsed Envelope
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.Kind != KindOverviewCPU {
		t.Errorf("kind = %q, want %q", parsed.Kind, KindOverviewCPU)
	}
	if len(parsed.Items) != 1 {
		t.Errorf("items len = %d, want 1", len(parsed.Items))
	}
	if parsed.Sections != nil {
		t.Errorf("Sections should be nil for leaf envelope")
	}
}

func TestWriteJSON_EmitsAggregatedSections(t *testing.T) {
	env := Envelope{
		Kind: KindOverview,
		Meta: Meta{FetchedAt: "2025-01-01T00:00:00Z"},
		Sections: map[string]Envelope{
			"physical": {Kind: KindOverviewPhysical, Items: []Item{{Display: map[string]any{"name": "alice"}}}},
			"ranking":  {Kind: KindOverviewRanking, Meta: Meta{Error: "fake-err", ErrorKind: "http_5xx"}},
		},
	}
	var buf bytes.Buffer
	if err := WriteJSON(&buf, env); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	sections, ok := parsed["sections"].(map[string]any)
	if !ok {
		t.Fatalf("sections should be a map; got %T", parsed["sections"])
	}
	if _, ok := sections["physical"]; !ok {
		t.Errorf("expected physical key")
	}
	rank, ok := sections["ranking"].(map[string]any)
	if !ok {
		t.Fatalf("ranking should be a map; got %T", sections["ranking"])
	}
	rankMeta, _ := rank["meta"].(map[string]any)
	if rankMeta["error"] != "fake-err" {
		t.Errorf("expected ranking.meta.error = fake-err, got %v", rankMeta["error"])
	}
}

func TestHeadItems_TruncatesOrPassesThrough(t *testing.T) {
	items := []Item{{}, {}, {}, {}}
	if got := HeadItems(items, 0); len(got) != 4 {
		t.Errorf("HeadItems(0) should be no-op; got len %d", len(got))
	}
	if got := HeadItems(items, 2); len(got) != 2 {
		t.Errorf("HeadItems(2) should truncate to 2; got len %d", len(got))
	}
	if got := HeadItems(items, 10); len(got) != 4 {
		t.Errorf("HeadItems(10) should not pad; got len %d", len(got))
	}
}

// ----------------------------------------------------------------------------
// Watch error classification
// ----------------------------------------------------------------------------

func TestClassifyTransportErr(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"nil → empty", nil, ""},
		{"context deadline → timeout", context.DeadlineExceeded, "timeout"},
		{"ErrNotLoggedIn → auth", &credential.ErrNotLoggedIn{OlaresID: "alice"}, "auth"},
		{"ErrTokenInvalidated → auth", &credential.ErrTokenInvalidated{OlaresID: "alice"}, "auth"},
		{"HTTPError 4xx", &HTTPError{Status: 400, ErrorKind: "http_4xx"}, "http_4xx"},
		{"HTTPError 5xx", &HTTPError{Status: 500, ErrorKind: "http_5xx"}, "http_5xx"},
		{"plain error → transport", errors.New("dial tcp"), "transport"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ClassifyTransportErr(tc.err)
			if got != tc.want {
				t.Errorf("ClassifyTransportErr(%v) = %q, want %q", tc.err, got, tc.want)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Watch ticker — fake clock + RunOnce stub
// ----------------------------------------------------------------------------

func TestRunner_OneShotPath(t *testing.T) {
	cf := &CommonFlags{Output: OutputJSON, Timezone: format.LocalLocation()}
	calls := 0
	r := &Runner{
		Flags:       cf,
		Recommended: time.Second,
		Now:         time.Now,
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			calls++
			return Envelope{Kind: "test", Meta: Meta{FetchedAt: now.Format(time.RFC3339)}}, nil
		},
	}
	r.Stdout = io.Discard
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if calls != 1 {
		t.Fatalf("one-shot Iter called %d times, want 1", calls)
	}
}

func TestRunner_WatchIterationsCap(t *testing.T) {
	cf := &CommonFlags{Output: OutputJSON, Watch: true, WatchIterations: 3, Timezone: format.LocalLocation()}
	calls := 0
	r := &Runner{
		Flags:       cf,
		Recommended: time.Millisecond,
		Now:         time.Now,
		Sleep: func(ctx context.Context, d time.Duration) error {
			return nil // fake clock — instantaneous
		},
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			calls++
			return Envelope{Kind: "test"}, nil
		},
	}
	r.Stdout = io.Discard
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if calls != 3 {
		t.Fatalf("watch Iter called %d times, want 3 (matching --watch-iterations)", calls)
	}
}

func TestRunner_WatchExitsOnConsecutiveFailures(t *testing.T) {
	cf := &CommonFlags{Output: OutputJSON, Watch: true, Timezone: format.LocalLocation()}
	r := &Runner{
		Flags:            cf,
		Recommended:      time.Millisecond,
		FailureThreshold: 2,
		Now:              time.Now,
		Sleep:            func(ctx context.Context, d time.Duration) error { return nil },
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			return Envelope{}, errors.New("boom")
		},
	}
	r.Stdout = io.Discard
	r.Stderr = io.Discard
	err := r.Run(context.Background())
	if err == nil {
		t.Fatal("expected aborted-after-N error; got nil")
	}
	if !strings.Contains(err.Error(), "consecutive failures") {
		t.Errorf("error should mention consecutive failures; got %v", err)
	}
}

func TestRunner_WatchRefusesUnsupportedCommand(t *testing.T) {
	cf := &CommonFlags{Watch: true, Timezone: format.LocalLocation()}
	r := &Runner{
		Flags:       cf,
		Recommended: 0, // command does not support polling
		Iter: func(ctx context.Context, iter int, now time.Time) (Envelope, error) {
			return Envelope{}, nil
		},
	}
	r.Stdout = io.Discard
	if err := r.Run(context.Background()); err == nil {
		t.Fatal("--watch on unsupported command should fail; got nil")
	}
}

// ----------------------------------------------------------------------------
// Cluster monitoring wire-shape — GET /kapis/.../v1alpha3/cluster
//
// Pins the request shape used by `overview physical` (and indirectly the
// physical section in `overview default`). The SPA's getClusterMonitoring
// posts as GET with metrics_filter / start / end / step / times. Any
// regression that changes verb or breaks the trailing `$` anchor surfaces
// here loudly.
// ----------------------------------------------------------------------------

func TestFetchClusterMetrics_GETWithMetricsFilter(t *testing.T) {
	var got struct {
		method        string
		path          string
		metricsFilter string
		step          string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.metricsFilter = r.URL.Query().Get("metrics_filter")
		got.step = r.URL.Query().Get("step")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	common.Timezone = format.LocalLocation()
	c := newTestClient(srv)
	if _, err := fetchClusterMetrics(context.Background(), c,
		[]string{"cluster_cpu_usage", "cluster_memory_total"},
		defaultClusterWindow(), time.Now(), false); err != nil {
		t.Fatalf("fetchClusterMetrics: %v", err)
	}

	if got.method != http.MethodGet {
		t.Errorf("method = %q, want GET", got.method)
	}
	if got.path != "/kapis/monitoring.kubesphere.io/v1alpha3/cluster" {
		t.Errorf("path = %q, want /kapis/monitoring.kubesphere.io/v1alpha3/cluster", got.path)
	}
	if !strings.HasSuffix(got.metricsFilter, "$") {
		t.Errorf("metrics_filter = %q is missing the trailing `$` anchor required by getParams", got.metricsFilter)
	}
	for _, name := range []string{"cluster_cpu_usage", "cluster_memory_total"} {
		if !strings.Contains(got.metricsFilter, name) {
			t.Errorf("metrics_filter %q is missing required metric %q", got.metricsFilter, name)
		}
	}
	if got.step != "600s" {
		t.Errorf(`step = %q, want "600s" (default cluster window)`, got.step)
	}
}

// ----------------------------------------------------------------------------
// Workload-grain wire-shape — overview ranking / applications.
//
// SPA's fetchWorkloadsMetrics fans out 2 monitoring fetches (per-pod for
// system apps + per-namespace for custom apps) and merges the two responses
// into one rows[] keyed by application. Pin the dual-fetch wire shape so
// future refactors don't accidentally drop one half.
// ----------------------------------------------------------------------------

func TestFetchWorkloadsMetrics_DualFetchPaths(t *testing.T) {
	var podsHits, nsHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/namespaces/") && strings.HasSuffix(r.URL.Path, "/pods"):
			podsHits++
		case r.URL.Path == "/kapis/monitoring.kubesphere.io/v1alpha3/namespaces":
			nsHits++
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	common.Timezone = format.LocalLocation()
	c := newTestClient(srv)
	req := workloadRequest{
		Apps: []workloadApp{
			{Name: "files", Title: "Files", Deployment: "files-deployment", Namespace: "user-space-alice", IsSystem: true},
			{Name: "alice-app", Title: "Alice's App", Namespace: "user-app-alice", IsSystem: false},
		},
		UserNamespace: "user-space-alice",
		Sort:          "desc",
	}
	if _, err := fetchWorkloadsMetrics(context.Background(), c, req, defaultClusterWindow(), time.Now()); err != nil {
		t.Fatalf("fetchWorkloadsMetrics: %v", err)
	}
	if podsHits != 1 {
		t.Errorf("/pods hits = %d, want 1 (system app fan-out)", podsHits)
	}
	if nsHits != 1 {
		t.Errorf("/namespaces hits = %d, want 1 (custom app fan-out)", nsHits)
	}
}

// TestFetchAppsList_FiltersEmptyEntrances pins the SPA's `appsWithNamespace`
// filter — entries with an empty `entrances` array are excluded. Mirrors
// stores/AppList.ts's `appsWithNamespace` getter.
func TestFetchAppsList_FiltersEmptyEntrances(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user-service/api/myapps_v2" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
            "code":0,"message":null,"data":[
              {"name":"jellyfin","title":"Jellyfin","namespace":"jellyfin","deployment":"jellyfin","entrances":[{"name":"web"}]},
              {"name":"hidden","namespace":"hidden","deployment":"hidden","entrances":[]},
              {"name":"olares-apps","title":"Olares Apps","namespace":"user-space-alice","deployment":"system-frontend-deployment","entrances":[{"name":"main"}]}
            ]
        }`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	apps, err := fetchAppsList(context.Background(), c)
	if err != nil {
		t.Fatalf("fetchAppsList: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("apps len = %d, want 2 (`hidden` filtered out)", len(apps))
	}
	if apps[0].Name != "jellyfin" || apps[1].Name != "olares-apps" {
		t.Errorf("order or names off: %+v", apps)
	}
}

// TestMergeWorkloadMetrics_LabelAware feeds the merger a tiny synthetic
// payload with two pods of the same deployment + two namespaces, and checks
// (a) per-deployment aggregation sums the pods, (b) per-namespace match
// keys on `metric.namespace`, (c) the system-frontend specialcase clones
// metrics across all entrance apps sharing the same deployment.
func TestMergeWorkloadMetrics_LabelAware(t *testing.T) {
	pod := func(podName string, value float64) struct {
		Metric map[string]string `json:"metric,omitempty"`
		Values [][]any           `json:"values,omitempty"`
		Value  []any             `json:"value,omitempty"`
	} {
		return struct {
			Metric map[string]string `json:"metric,omitempty"`
			Values [][]any           `json:"values,omitempty"`
			Value  []any             `json:"value,omitempty"`
		}{
			Metric: map[string]string{"pod": podName, "owner_kind": "Deployment"},
			Values: [][]any{{float64(0), value}},
		}
	}
	ns := func(name string, value float64) struct {
		Metric map[string]string `json:"metric,omitempty"`
		Values [][]any           `json:"values,omitempty"`
		Value  []any             `json:"value,omitempty"`
	} {
		return struct {
			Metric map[string]string `json:"metric,omitempty"`
			Values [][]any           `json:"values,omitempty"`
			Value  []any             `json:"value,omitempty"`
		}{
			Metric: map[string]string{"namespace": name},
			Values: [][]any{{float64(0), value}},
		}
	}

	var podCPU format.MonitoringResult
	podCPU.Data.Result = append(podCPU.Data.Result,
		pod("system-frontend-deployment-abcd-aa11", 0.10),
		pod("system-frontend-deployment-abcd-bb22", 0.05),
	)
	podData := map[string]format.MonitoringResult{"pod_cpu_usage": podCPU}

	var nsCPU format.MonitoringResult
	nsCPU.Data.Result = append(nsCPU.Data.Result,
		ns("jellyfin", 0.42),
		ns("home-assistant", 0.07),
	)
	var nsPodCount format.MonitoringResult
	nsPodCount.Data.Result = append(nsPodCount.Data.Result,
		ns("jellyfin", 3),
		ns("home-assistant", 1),
	)
	nsData := map[string]format.MonitoringResult{"namespace_cpu_usage": nsCPU, "namespace_pod_count": nsPodCount}

	apps := []workloadApp{
		// Two entrance apps share `system-frontend-deployment` (the SPA
		// specialcase). Both must end up with the same per-deployment
		// metric.
		{Name: "olares-apps", Title: "Olares Apps", Namespace: "user-space-alice", Deployment: "system-frontend-deployment", IsSystem: true},
		{Name: "windows", Title: "Windows", Namespace: "user-space-alice", Deployment: "system-frontend-deployment", IsSystem: true},
		{Name: "jellyfin", Title: "Jellyfin", Namespace: "jellyfin", IsSystem: false},
		{Name: "home-assistant", Title: "Home Assistant", Namespace: "home-assistant", IsSystem: false},
	}
	rows := mergeWorkloadMetrics(apps, podData, nsData)
	if len(rows) != 4 {
		t.Fatalf("rows = %d, want 4", len(rows))
	}
	// Both system-frontend entrance apps see the same summed cpu (.10+.05≈.15;
	// float64 makes this 0.150000000000…02, so compare with a small epsilon).
	const eps = 1e-9
	delta := func(a, b float64) float64 {
		if a > b {
			return a - b
		}
		return b - a
	}
	if delta(rows[0].CPU, 0.15) > eps || delta(rows[1].CPU, 0.15) > eps {
		t.Errorf("system-frontend entrance cpu = (%v,%v), want both 0.15", rows[0].CPU, rows[1].CPU)
	}
	if rows[0].PodCount != 2 || rows[1].PodCount != 2 {
		t.Errorf("system-frontend entrance pod_count = (%d,%d), want both 2", rows[0].PodCount, rows[1].PodCount)
	}
	// Custom apps key on metric.namespace.
	if rows[2].CPU != 0.42 {
		t.Errorf("jellyfin cpu = %v, want 0.42", rows[2].CPU)
	}
	if rows[2].PodCount != 3 {
		t.Errorf("jellyfin pod_count = %d, want 3", rows[2].PodCount)
	}
	if rows[3].CPU != 0.07 {
		t.Errorf("home-assistant cpu = %v, want 0.07", rows[3].CPU)
	}
}

// TestPodDeploymentName pins the slice-trailing-2 quirk from
// Applications2/config.ts:277. Empty / short pod names degrade gracefully.
func TestPodDeploymentName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"system-frontend-deployment-abcd-1234", "system-frontend-deployment"},
		{"jellyfin-7d9c8b4f8d-x9plk", "jellyfin"},
		{"single", "single"},
		{"two-segments", "two-segments"},
		{"", ""},
	}
	for _, tc := range cases {
		got := podDeploymentName(tc.in)
		if got != tc.want {
			t.Errorf("podDeploymentName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// unknownSubcommandRunE / TestAllLeafCommandsSilenced live in
// cli/cmd/ctl/dashboard/dashboard_test.go because they're inherently
// tied to NewDashboardCommand's cobra binding (an artefact of the cmd
// shell, not the pkg core).

// ----------------------------------------------------------------------------
// overview disk — lsblk subtree resolution + tree prefix
// ----------------------------------------------------------------------------
//
// These pin the SPA's `getDiskPartitionRows` algorithm
// (Overview2/Disk/config.ts:398). Three behaviours matter:
//
//   - subtree extraction has two modes (pkname BFS vs prefix fallback)
//     gated by hasPknameLabels;
//   - parent resolution prefers pkname > longest-name-prefix > root;
//   - the tree-prefix walker emits ASCII connectors that respect the
//     "last sibling at depth i" stack.

func TestHasPknameLabels(t *testing.T) {
	if hasPknameLabels([]lsblkRow{{Name: "sda"}}) {
		t.Fatal("rows w/o pkname should report false")
	}
	if !hasPknameLabels([]lsblkRow{{Name: "sda1", Pkname: "sda"}}) {
		t.Fatal("at least one pkname label should report true")
	}
	if hasPknameLabels([]lsblkRow{{Name: "sda1", Pkname: "  "}}) {
		t.Fatal("whitespace-only pkname should not count")
	}
}

func TestCollectSubtreeByPkname_BFSRoot(t *testing.T) {
	rows := []lsblkRow{
		{Name: "sda"},
		{Name: "sda1", Pkname: "sda"},
		{Name: "sda2", Pkname: "sda"},
		{Name: "sda1_lvm", Pkname: "sda1"},
		{Name: "sdb", Pkname: ""}, // unrelated tree
		{Name: "sdb1", Pkname: "sdb"},
	}
	got := collectSubtreeByPkname(rows, "sda")
	want := []string{"sda", "sda1", "sda2", "sda1_lvm"}
	gotNames := []string{}
	for _, r := range got {
		gotNames = append(gotNames, r.Name)
	}
	if strings.Join(gotNames, ",") != strings.Join(want, ",") {
		t.Errorf("subtree(sda) = %v, want %v", gotNames, want)
	}
}

func TestResolveParent_PrefersPknameThenLongestPrefix(t *testing.T) {
	rows := []lsblkRow{
		{Name: "sda"},
		{Name: "sda1", Pkname: "sda"},
		{Name: "sda1_extra"}, // no pkname; falls back to prefix → sda1 (longer than sda)
	}
	nameSet := map[string]bool{"sda": true, "sda1": true, "sda1_extra": true}
	if got := resolveParent(rows[1], "sda", nameSet); got != "sda" {
		t.Errorf("resolveParent(sda1) pkname path = %q, want sda", got)
	}
	if got := resolveParent(rows[2], "sda", nameSet); got != "sda1" {
		t.Errorf("resolveParent(sda1_extra) prefix path = %q, want sda1 (longest prefix)", got)
	}
	// Root row → no parent regardless of labels.
	if got := resolveParent(rows[0], "sda", nameSet); got != "" {
		t.Errorf("resolveParent(root) = %q, want \"\"", got)
	}
}

func TestBuildLsblkTreePrefix_LastStackBlanksTrunk(t *testing.T) {
	cases := []struct {
		depth int
		stack []bool
		want  string
	}{
		{0, nil, ""},
		{1, []bool{false}, "├── "},
		{1, []bool{true}, "└── "},
		{2, []bool{false, false}, "│   ├── "},
		{2, []bool{false, true}, "│   └── "},
		{2, []bool{true, false}, "    ├── "},
		{2, []bool{true, true}, "    └── "},
		{3, []bool{false, true, false}, "│       ├── "},
	}
	for _, tc := range cases {
		if got := buildLsblkTreePrefix(tc.depth, tc.stack); got != tc.want {
			t.Errorf("buildLsblkTreePrefix(%d, %v) = %q, want %q", tc.depth, tc.stack, got, tc.want)
		}
	}
}

func TestFlattenLsblkHierarchy_Pkname(t *testing.T) {
	rows := []lsblkRow{
		{Name: "sda", Size: "100G"},
		{Name: "sda1", Pkname: "sda", Size: "50G"},
		{Name: "sda2", Pkname: "sda", Size: "50G"},
		{Name: "sda2_lvm", Pkname: "sda2", Size: "30G"},
	}
	flat := flattenLsblkHierarchy(rows, "sda")
	if len(flat) != 4 {
		t.Fatalf("flat rows = %d, want 4", len(flat))
	}
	wantNames := []string{"sda", "sda1", "sda2", "sda2_lvm"}
	for i, fr := range flat {
		if fr.Row.Name != wantNames[i] {
			t.Errorf("row[%d].Name = %q, want %q", i, fr.Row.Name, wantNames[i])
		}
	}
	// sda1 is not the last sibling of root → "├── "
	if flat[1].TreePrefix != "├── " {
		t.Errorf("sda1 prefix = %q, want '├── '", flat[1].TreePrefix)
	}
	// sda2 is the last sibling of root → "└── "
	if flat[2].TreePrefix != "└── " {
		t.Errorf("sda2 prefix = %q, want '└── '", flat[2].TreePrefix)
	}
	// sda2_lvm is the only child of sda2 (sda2 was last sibling) → "    └── "
	if flat[3].TreePrefix != "    └── " {
		t.Errorf("sda2_lvm prefix = %q, want '    └── '", flat[3].TreePrefix)
	}
	// Parent links match the source.
	if flat[1].Parent != "sda" || flat[3].Parent != "sda2" {
		t.Errorf("parent links wrong: %+v / %+v", flat[1], flat[3])
	}
	if flat[0].Depth != 0 || flat[3].Depth != 2 {
		t.Errorf("depth wrong: root=%d (want 0), leaf=%d (want 2)", flat[0].Depth, flat[3].Depth)
	}
}

func TestFlattenLsblkHierarchy_FallbackWhenRootMissing(t *testing.T) {
	rows := []lsblkRow{
		{Name: "sdb1"},
		{Name: "sdb2"},
	}
	flat := flattenLsblkHierarchy(rows, "sda") // root not present
	if len(flat) != 2 {
		t.Fatalf("flat rows = %d, want 2", len(flat))
	}
	for _, fr := range flat {
		if fr.Depth != 0 || fr.TreePrefix != "" {
			t.Errorf("degraded path should drop tree info: %+v", fr)
		}
	}
}

func TestRenderDiskTemperature_HonoursUnitAndDash(t *testing.T) {
	if got := renderDiskTemperature(0, format.TempC); got != "-" {
		t.Errorf("zero celsius should print '-', got %q", got)
	}
	if got := renderDiskTemperature(40, format.TempC); got != "40°C" {
		t.Errorf("40C → %q, want 40°C", got)
	}
	if got := renderDiskTemperature(40, format.TempF); got != "104°F" {
		t.Errorf("40C in F → %q, want 104°F", got)
	}
	if got := renderDiskTemperature(40, format.TempK); got != "313.1K" {
		t.Errorf("40C in K → %q, want 313.1K", got)
	}
}

// ----------------------------------------------------------------------------
// Capability gates — fan / gpu subtree pre-flight checks
// ----------------------------------------------------------------------------
//
// gateOlaresOne reads /user-service/api/system/status; gateGPU additionally
// reads /capi/app/detail (admin) + /kapis/.../nodes (CUDA labels). These
// tests pin both the cache discipline (sync.Once) and the structured empty
// envelope agents rely on.

func TestClient_EnsureSystemStatus_CachesResult(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user-service/api/system/status" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":null,"data":{"device_name":"Olares One","host_name":"box","cpu_info":"i9-14900K","gpu_info":"RTX 4070"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	for i := 0; i < 3; i++ {
		s, err := c.EnsureSystemStatus(context.Background())
		if err != nil {
			t.Fatalf("EnsureSystemStatus iter %d: %v", i, err)
		}
		if s.DeviceName != "Olares One" {
			t.Fatalf("iter %d: DeviceName = %q, want %q", i, s.DeviceName, "Olares One")
		}
		if !s.IsOlaresOne() {
			t.Fatalf("iter %d: IsOlaresOne() = false", i)
		}
	}
	if hits != 1 {
		t.Fatalf("EnsureSystemStatus hit upstream %d times, want 1 (sync.Once)", hits)
	}
}

func TestClient_IsOlaresOne_FalseOnGenericBox(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"device_name":"DIY-PC"}}`))
	}))
	defer srv.Close()

	one, err := newTestClient(srv).IsOlaresOne(context.Background())
	if err != nil {
		t.Fatalf("IsOlaresOne: %v", err)
	}
	if one {
		t.Errorf("DIY-PC should not be Olares One")
	}
}

// TestHasCUDANode_LabelMatch covers both directions of the label scan:
// any node carrying gpu.bytetrade.io/cuda-supported=true wins; otherwise
// the cluster reports false. We also assert the per-Client cache so a
// repeated call does not re-fetch.
func TestHasCUDANode_LabelMatch(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "any node with cuda-supported=true wins",
			body: `{"items":[{"metadata":{"labels":{"node-role.kubernetes.io/master":"true"}}},{"metadata":{"labels":{"gpu.bytetrade.io/cuda-supported":"true"}}}]}`,
			want: true,
		},
		{
			name: "no nodes with cuda-supported",
			body: `{"items":[{"metadata":{"labels":{"node-role.kubernetes.io/master":"true"}}}]}`,
			want: false,
		},
		{
			name: "label present but not 'true' does not count",
			body: `{"items":[{"metadata":{"labels":{"gpu.bytetrade.io/cuda-supported":"false"}}}]}`,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hits := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hits++
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			c := newTestClient(srv)
			defer func() {
				cudaNodeMu.Lock()
				delete(cudaNodeCache, c)
				cudaNodeMu.Unlock()
			}()
			got, err := hasCUDANode(context.Background(), c)
			if err != nil {
				t.Fatalf("hasCUDANode: %v", err)
			}
			if got != tc.want {
				t.Errorf("hasCUDANode = %v, want %v", got, tc.want)
			}
			// Second call must hit the cache.
			_, _ = hasCUDANode(context.Background(), c)
			if hits != 1 {
				t.Errorf("hasCUDANode hit upstream %d times, want 1 (per-Client cache)", hits)
			}
		})
	}
}

// TestGateOlaresOne_BlocksGenericBox asserts the empty envelope shape +
// stderr hint when the device is not Olares One. JSON output mode must
// stay silent on stderr (agents read stdout exclusively).
func TestGateOlaresOne_BlocksGenericBox(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"device_name":"DIY-PC"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	prev := common
	t.Cleanup(func() { common = prev })

	common = &CommonFlags{Output: OutputJSON, Timezone: prev.Timezone}
	if common.Timezone == nil {
		_ = common.Validate() // populate Timezone from defaults
	}

	stderrR, stderrW, _ := os.Pipe()
	prevStderr := os.Stderr
	os.Stderr = stderrW
	defer func() { os.Stderr = prevStderr }()

	env, gated := gateOlaresOne(context.Background(), c, "dashboard.overview.fan", time.Now())
	stderrW.Close()
	captured, _ := io.ReadAll(stderrR)

	if !gated {
		t.Fatal("gateOlaresOne should gate on DIY-PC")
	}
	if env.Meta.EmptyReason != "not_olares_one" {
		t.Errorf("EmptyReason = %q, want not_olares_one", env.Meta.EmptyReason)
	}
	if env.Meta.DeviceName != "DIY-PC" {
		t.Errorf("DeviceName = %q, want DIY-PC", env.Meta.DeviceName)
	}
	if env.Meta.Note == "" {
		t.Errorf("Note should be populated")
	}
	if !env.Meta.Empty {
		t.Errorf("Empty should be true")
	}
	if len(captured) != 0 {
		t.Errorf("JSON output mode must keep stderr silent; got %q", captured)
	}
}

func TestGateOlaresOne_PassesOnOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"device_name":"Olares One"}}`))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	prev := common
	t.Cleanup(func() { common = prev })
	common = &CommonFlags{Output: OutputTable, Timezone: prev.Timezone}
	_ = common.Validate()

	_, gated := gateOlaresOne(context.Background(), c, "dashboard.overview.fan", time.Now())
	if gated {
		t.Fatal("gateOlaresOne should pass through on Olares One")
	}
}

// gpuAdvisory is the soft-gate companion to the (still-strict)
// gateOlaresOne. The SPA's GPU detail pages do NOT hard-block on admin
// role or CUDA labels — only the sidebar entry is hidden — so the CLI
// likewise emits a stderr advisory + meta.note hint and lets the data
// fetch proceed. These tests pin the two soft signals
// ("non-admin" / "no CUDA node") and the all-clear path that returns
// empty strings for both note and reason.

// TestGPUAdvisory_NonAdminEmitsHint: a non-admin profile gets a one-
// liner stderr advisory + a meta.note hint, but no envelope is built
// (the caller is responsible for stitching the note onto its own
// envelope). The function MUST NOT touch the nodes endpoint when the
// caller is non-admin — that endpoint is admin-only and would 403.
func TestGPUAdvisory_NonAdminEmitsHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"bob","globalrole":"workspaces-manager"}}`))
		case "/kapis/resources.kubesphere.io/v1alpha3/nodes":
			t.Errorf("non-admin path must not reach the nodes lookup")
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	prev := common
	t.Cleanup(func() { common = prev })
	common = &CommonFlags{Output: OutputTable, Timezone: prev.Timezone} // stderr hint enabled
	_ = common.Validate()

	stderr := captureStderr(t, func() {
		note, reason := gpuAdvisory(context.Background(), c)
		if note == "" {
			t.Error("gpuAdvisory should emit a non-empty note for non-admin")
		}
		if reason != "gpu_sidebar_hidden_non_admin" {
			t.Errorf("reason = %q, want gpu_sidebar_hidden_non_admin", reason)
		}
	})
	if !strings.Contains(stderr, "non-admin") {
		t.Errorf("stderr did not mention 'non-admin': %q", stderr)
	}
}

// TestGPUAdvisory_AdminWithoutCUDA: admin path reaches the nodes
// lookup, and a label-less node set produces the
// "gpu_sidebar_hidden_no_cuda_node" advisory. Data fetch is the
// caller's job — we just emit the hint.
func TestGPUAdvisory_AdminWithoutCUDA(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case "/kapis/resources.kubesphere.io/v1alpha3/nodes":
			_, _ = w.Write([]byte(`{"items":[{"metadata":{"labels":{}}}]}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer func() {
		cudaNodeMu.Lock()
		delete(cudaNodeCache, c)
		cudaNodeMu.Unlock()
	}()
	prev := common
	t.Cleanup(func() { common = prev })
	common = &CommonFlags{Output: OutputJSON, Timezone: prev.Timezone} // suppress stderr
	_ = common.Validate()

	note, reason := gpuAdvisory(context.Background(), c)
	if reason != "gpu_sidebar_hidden_no_cuda_node" {
		t.Errorf("reason = %q, want gpu_sidebar_hidden_no_cuda_node", reason)
	}
	if !strings.Contains(note, "cuda-supported") {
		t.Errorf("note should mention the CUDA label, got %q", note)
	}
}

// TestGPUAdvisory_AdminWithCUDA_AllClear: admin + at least one CUDA-
// capable node returns empty strings for both note and reason —
// nothing to surface, callers proceed unmodified.
func TestGPUAdvisory_AdminWithCUDA_AllClear(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/capi/app/detail":
			_, _ = w.Write([]byte(`{"user":{"username":"alice","globalrole":"platform-admin"}}`))
		case "/kapis/resources.kubesphere.io/v1alpha3/nodes":
			_, _ = w.Write([]byte(`{"items":[{"metadata":{"labels":{"gpu.bytetrade.io/cuda-supported":"true"}}}]}`))
		}
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defer func() {
		cudaNodeMu.Lock()
		delete(cudaNodeCache, c)
		cudaNodeMu.Unlock()
	}()
	prev := common
	t.Cleanup(func() { common = prev })
	common = &CommonFlags{Output: OutputJSON, Timezone: prev.Timezone}
	_ = common.Validate()

	note, reason := gpuAdvisory(context.Background(), c)
	if note != "" || reason != "" {
		t.Errorf("admin + CUDA should be all-clear, got note=%q reason=%q", note, reason)
	}
}

// TestVGPUUnavailableFromError_5xx: HAMI returning HTTP 5xx classifies
// as the new "vgpu_unavailable" empty_reason, copies the upstream body
// `message` into Meta.Error, and stamps the original status into
// Meta.HTTPStatus. The classification fires for the whole 5xx block,
// not just 500.
func TestVGPUUnavailableFromError_5xx(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"code":1,"message":"unknown request error"}`))
		}))
		c := newTestClient(srv)
		_, err := fetchGraphicsList(context.Background(), c, nil)
		if err == nil {
			t.Fatalf("expected HTTP %d to surface as error", status)
		}
		prev := common
		common = &CommonFlags{Output: OutputJSON, Timezone: prev.Timezone}
		_ = common.Validate()
		env, ok := vgpuUnavailableFromError(c, err, "dashboard.overview.gpu.list", time.Now())
		common = prev
		if !ok {
			t.Errorf("HTTP %d should classify as vgpu_unavailable", status)
		}
		if env.Meta.EmptyReason != "vgpu_unavailable" {
			t.Errorf("HTTP %d: EmptyReason = %q, want vgpu_unavailable", status, env.Meta.EmptyReason)
		}
		if env.Meta.HTTPStatus != status {
			t.Errorf("HTTP %d: HTTPStatus = %d, want %d", status, env.Meta.HTTPStatus, status)
		}
		if env.Meta.Error != "unknown request error" {
			t.Errorf("HTTP %d: Meta.Error = %q, want %q", status, env.Meta.Error, "unknown request error")
		}
		srv.Close()
	}
}

// TestVGPUUnavailableFromError_404Skips: HTTP 404 is the SPA's signal
// for "HAMI not installed" and is handled separately by the caller
// (no_vgpu_integration). Confirm vgpuUnavailableFromError stays out of
// the 4xx lane so we don't double-classify.
func TestVGPUUnavailableFromError_404Skips(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	_, err := fetchGraphicsList(context.Background(), c, nil)
	if err == nil {
		t.Fatal("expected 404 error")
	}
	prev := common
	common = &CommonFlags{Output: OutputJSON, Timezone: prev.Timezone}
	_ = common.Validate()
	defer func() { common = prev }()
	if _, ok := vgpuUnavailableFromError(c, err, "dashboard.overview.gpu.list", time.Now()); ok {
		t.Error("404 should NOT classify as vgpu_unavailable; 4xx is reserved for the no_vgpu_integration path")
	}
}

// TestExtractHAMIMessage_FallbackToBody: bodies that aren't JSON or
// don't carry a `message` field still surface a trimmed-and-capped
// version of the body so the agent gets *something* useful.
func TestExtractHAMIMessage_FallbackToBody(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{`{"code":1,"message":"x"}`, "x"},
		{`{"foo":"bar"}`, `{"foo":"bar"}`},
		{`plain text body`, `plain text body`},
		{strings.Repeat("a", 300), strings.Repeat("a", 256)},
		{`   `, ``},
	}
	for i, tc := range cases {
		got := extractHAMIMessage(tc.body)
		if got != tc.want {
			t.Errorf("case %d: got %q, want %q", i, got, tc.want)
		}
	}
}

// ----------------------------------------------------------------------------
// HAMI request-body wire-shape regressions
// ----------------------------------------------------------------------------
//
// HAMI's WebUI /api/vgpu/v1/{gpus,containers} endpoints reject a body
// missing the `filters` key with a generic 500 ("unknown request error")
// — the SPA always sends `{"filters":{},"pageRequest":{...}}` even
// when no filters apply. These tests pin the CLI to the same wire shape
// so the next person to add `omitempty` doesn't quietly bring back the
// `vgpu_unavailable` regression that masqueraded as HAMI being down.

// TestGraphicsListBody_AlwaysIncludesFiltersKey: the request body the
// CLI POSTs to /hami/api/vgpu/v1/gpus must carry an explicit
// `"filters":{}` (or non-empty filter map) — never an absent key, never
// a `null`. Asserts both nil-input and empty-map-input.
func TestGraphicsListBody_AlwaysIncludesFiltersKey(t *testing.T) {
	for _, tc := range []struct {
		name    string
		filters map[string]string
	}{
		{"nil", nil},
		{"empty", map[string]string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var captured map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %q, want POST", r.Method)
				}
				body, _ := io.ReadAll(r.Body)
				if err := json.Unmarshal(body, &captured); err != nil {
					t.Fatalf("decode body: %v (raw=%q)", err, string(body))
				}
				// Mirror HAMI's actual wire shape (top-level
				// `list`, no `data` envelope) so we don't
				// quietly accept a stub that lies about the
				// upstream contract.
				_, _ = w.Write([]byte(`{"list":[]}`))
			}))
			defer srv.Close()
			c := newTestClient(srv)
			if _, err := fetchGraphicsList(context.Background(), c, tc.filters); err != nil {
				t.Fatalf("fetchGraphicsList: %v", err)
			}
			f, ok := captured["filters"]
			if !ok {
				t.Fatal("body missing `filters` key — HAMI 5xx will return")
			}
			if f == nil {
				t.Fatal("body has `filters: null` — HAMI 5xx will return; want `{}`")
			}
			if _, ok := f.(map[string]any); !ok {
				t.Fatalf("filters wire shape = %T, want object", f)
			}
			pr, ok := captured["pageRequest"].(map[string]any)
			if !ok {
				t.Fatal("body missing `pageRequest` object")
			}
			if pr["sort"] != "DESC" || pr["sortField"] != "id" {
				t.Errorf("pageRequest = %v, want sort=DESC sortField=id", pr)
			}
		})
	}
}

// TestTaskListBody_AlwaysIncludesFiltersKey: same regression net for
// the /v1/containers endpoint (task list). Different response shape —
// top-level `items` instead of top-level `list` — so we have to
// exercise this fetcher independently.
func TestTaskListBody_AlwaysIncludesFiltersKey(t *testing.T) {
	var captured map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		// HAMI returns `{"items":[...]}` at the top level; no
		// `data` envelope (matches `TaskListResponse`).
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	if _, err := fetchTaskList(context.Background(), c, nil); err != nil {
		t.Fatalf("fetchTaskList: %v", err)
	}
	f, ok := captured["filters"]
	if !ok {
		t.Fatal("task body missing `filters` key — HAMI 5xx will return")
	}
	if f == nil {
		t.Fatal("task body has `filters: null` — HAMI 5xx will return; want `{}`")
	}
}

// ----------------------------------------------------------------------------
// HAMI response wire-shape regressions
// ----------------------------------------------------------------------------
//
// HAMI's `/api/vgpu/v1/{gpus,containers,gpu,container}` endpoints all
// return the payload at the TOP LEVEL — there is no `data: { ... }`
// envelope. An earlier revision wrapped each fetcher's decoder struct
// with `data` and silently produced "0 GPUs" on machines where the SPA
// rendered devices fine. These tests pin every fetcher to the real
// wire shape so regressing the wrapper would fail loudly.

// TestFetchGraphicsList_ParsesTopLevelList: HAMI returns
// `{"list":[ {...} ]}` (no `data` envelope). Confirm fetchGraphicsList
// returns the list verbatim — no "data.list" indirection.
func TestFetchGraphicsList_ParsesTopLevelList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Real fixture captured from olarestest005.
		_, _ = w.Write([]byte(`{"list":[{"uuid":"GPU-be013ee7","type":"NVIDIA GeForce RTX 5070","shareMode":"3","nodeName":"olares","health":true,"coreUtilizedPercent":0,"memoryTotal":24463,"memoryUsed":0,"power":7.937,"temperature":47}]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	list, err := fetchGraphicsList(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchGraphicsList: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1 — fetcher likely re-introduced the bogus `data` envelope", len(list))
	}
	g := list[0]
	if g["uuid"] != "GPU-be013ee7" {
		t.Errorf("uuid = %v, want GPU-be013ee7", g["uuid"])
	}
	if g["type"] != "NVIDIA GeForce RTX 5070" {
		t.Errorf("type = %v, want NVIDIA GeForce RTX 5070", g["type"])
	}
}

// TestFetchTaskList_ParsesTopLevelItems: parallel to the GPU list test
// but for `/v1/containers`. Same wire-shape contract: top-level `items`,
// no `data` envelope.
func TestFetchTaskList_ParsesTopLevelItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"name":"task-A","status":"running","podUid":"pod-1","nodeName":"olares","deviceShareModes":["2"],"devicesCoreUtilizedPercent":[12.5],"devicesMemUtilized":[1024]}]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	list, err := fetchTaskList(context.Background(), c, nil)
	if err != nil {
		t.Fatalf("fetchTaskList: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(items) = %d, want 1 — fetcher likely re-introduced the bogus `data` envelope", len(list))
	}
	if list[0]["name"] != "task-A" {
		t.Errorf("name = %v, want task-A", list[0]["name"])
	}
}

// TestFetchGraphicsDetail_ReturnsBodyAsIs: HAMI's `/v1/gpu` returns the
// detail object flat at the top level (matches `GraphicsDetailsResponse`
// in src/apps/dashboard/types/gpu.ts). Confirm the fetcher does not
// hunt for a non-existent `data` key and instead returns the raw body.
func TestFetchGraphicsDetail_ReturnsBodyAsIs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"GPU-be013ee7","type":"NVIDIA","health":true,"memoryTotal":24463}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	d, err := fetchGraphicsDetail(context.Background(), c, "GPU-be013ee7")
	if err != nil {
		t.Fatalf("fetchGraphicsDetail: %v", err)
	}
	if d["uuid"] != "GPU-be013ee7" {
		t.Errorf("uuid = %v, want GPU-be013ee7", d["uuid"])
	}
	if d["health"] != true {
		t.Errorf("health = %v, want true (preserve bool)", d["health"])
	}
}

// TestFetchTaskDetail_ReturnsBodyAsIs: same wire-shape contract for
// `/v1/container` (task detail).
func TestFetchTaskDetail_ReturnsBodyAsIs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"name":"task-A","status":"running","podUid":"pod-1"}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	d, err := fetchTaskDetail(context.Background(), c, "task-A", "pod-1", "")
	if err != nil {
		t.Fatalf("fetchTaskDetail: %v", err)
	}
	if d["name"] != "task-A" || d["status"] != "running" || d["podUid"] != "pod-1" {
		t.Errorf("body = %v; expected name/status/podUid preserved", d)
	}
}

// TestGPUFormatHelpers locks the rendering helpers used by the GPU
// list/tasks tables. These are subtle — the SPA uses inconsistent
// conventions (percent values are pre-multiplied, but VRAM is in MiB
// and gets converted to bytes before formatting), and AI agents rely
// on the JSON `display` map matching what humans see in the table.
func TestGPUFormatHelpers(t *testing.T) {
	t.Run("percentDirect", func(t *testing.T) {
		cases := map[float64]string{
			0:     "0%",
			25.5:  "25.5%",
			0.004: "0%", // round to 2dp drops fractional bytes
			100:   "100%",
		}
		for in, want := range cases {
			if got := percentDirect(in); got != want {
				t.Errorf("percentDirect(%v) = %q, want %q", in, got, want)
			}
		}
	})
	t.Run("gpuModeLabel", func(t *testing.T) {
		cases := map[string]string{
			"0": "App exclusive",
			"1": "Memory slicing",
			"2": "Time slicing",
			"3": "mode=3", // unknown — preserve raw
			"":  "-",
		}
		for in, want := range cases {
			if got := gpuModeLabel(in); got != want {
				t.Errorf("gpuModeLabel(%q) = %q, want %q", in, got, want)
			}
		}
	})
	t.Run("gpuHealthLabel", func(t *testing.T) {
		if got := gpuHealthLabel(true); got != "healthy" {
			t.Errorf("gpuHealthLabel(true) = %q, want healthy", got)
		}
		if got := gpuHealthLabel(false); got != "unhealthy" {
			t.Errorf("gpuHealthLabel(false) = %q, want unhealthy", got)
		}
	})
	t.Run("firstAnyInArray", func(t *testing.T) {
		// JSON-decoded input is always []any.
		var arr any = []any{"first", "second"}
		if got := firstAnyInArray(arr); got != "first" {
			t.Errorf("[]any: got %v, want first", got)
		}
		var empty any = []any{}
		if got := firstAnyInArray(empty); got != nil {
			t.Errorf("empty: got %v, want nil", got)
		}
		if got := firstAnyInArray("not-array"); got != nil {
			t.Errorf("string input: got %v, want nil", got)
		}
	})
}

// ----------------------------------------------------------------------------
// HAMI monitor query endpoints — wire-shape regressions
// ----------------------------------------------------------------------------
//
// The `/v1/monitor/query/{instant-vector,range-vector}` endpoints DO
// wrap the result in a `data` envelope (matches the SPA's
// InstantVectorResponse / RangeVectorResponse types). This is the
// opposite contract from `gpus` / `containers` / `gpu` / `container`
// which return the body flat. These tests pin the fetchers to the
// monitor wire shape so a future refactor doesn't accidentally
// "harmonise" the two contracts.

// TestFetchInstantVector_ParsesDataEnvelope: HAMI's instant-vector
// endpoint returns `{"data": [{metric, value, timestamp}]}`. Confirm
// the fetcher reads the `data` array out of the envelope verbatim and
// preserves metric labels + value as decoded JSON values.
func TestFetchInstantVector_ParsesDataEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if req["query"] != "hami_core_util" {
			t.Errorf("body.query = %v, want hami_core_util", req["query"])
		}
		_, _ = w.Write([]byte(`{"data":[{"metric":{"deviceuuid":"GPU-1","device_no":"nvidia0"},"value":42.5,"timestamp":"1745000000"}]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	samples, err := fetchInstantVector(context.Background(), c, "hami_core_util")
	if err != nil {
		t.Fatalf("fetchInstantVector: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("len(samples) = %d, want 1 — fetcher likely lost the `data` envelope", len(samples))
	}
	if samples[0].Value != 42.5 {
		t.Errorf("value = %v, want 42.5", samples[0].Value)
	}
	if samples[0].Metric["device_no"] != "nvidia0" {
		t.Errorf("metric[device_no] = %v, want nvidia0", samples[0].Metric["device_no"])
	}
}

// TestFetchRangeVector_ParsesDataEnvelope: parallel to instant-vector
// but for `/v1/monitor/query/range-vector`. Same `data` envelope; each
// element carries `values: [{value, timestamp}]`. We keep `value` as
// `any` on the wire (SPA fixtures show string-shaped values for some
// counters, numeric for others) so the parser doesn't choke.
func TestFetchRangeVector_ParsesDataEnvelope(t *testing.T) {
	var capturedRange map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		capturedRange, _ = req["range"].(map[string]any)
		_, _ = w.Write([]byte(`{"data":[{"metric":{"device_no":"nvidia0","driver_version":"590.44.01"},"values":[{"value":1.5,"timestamp":"1745000000"},{"value":2.0,"timestamp":"1745000300"}]}]}`))
	}))
	defer srv.Close()
	c := newTestClient(srv)
	series, err := fetchRangeVector(context.Background(), c, "hami_device_power", "2026-04-28 14:00:00", "2026-04-28 22:00:00", "30m")
	if err != nil {
		t.Fatalf("fetchRangeVector: %v", err)
	}
	if capturedRange == nil {
		t.Fatal("body.range missing — fetcher should always include range start/end/step")
	}
	if capturedRange["step"] != "30m" {
		t.Errorf("range.step = %v, want 30m", capturedRange["step"])
	}
	if len(series) != 1 {
		t.Fatalf("len(series) = %d, want 1", len(series))
	}
	if series[0].Metric["driver_version"] != "590.44.01" {
		t.Errorf("metric[driver_version] = %v, want 590.44.01", series[0].Metric["driver_version"])
	}
	if len(series[0].Values) != 2 {
		t.Fatalf("len(values) = %d, want 2", len(series[0].Values))
	}
	if v := toFloat(series[0].Values[1].Value); v != 2.0 {
		t.Errorf("values[1] = %v, want 2.0", v)
	}
}

// TestGPUTrendStep covers every preset window from utils.js's
// `timeReflection` table plus a handful of off-preset windows that
// must fall back to the algorithm `floor(minutes/16)m, capped 1..60`.
func TestGPUTrendStep(t *testing.T) {
	base := time.Date(2026, 4, 28, 14, 42, 0, 0, time.UTC)
	cases := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"10m_preset", 10 * time.Minute, "1m"},
		{"30m_preset", 30 * time.Minute, "1m"},
		{"1h_preset", 60 * time.Minute, "10m"},
		{"2h_preset", 120 * time.Minute, "20m"},
		{"3h_preset", 180 * time.Minute, "10m"},
		{"5h_preset", 300 * time.Minute, "10m"},
		{"8h_preset", 480 * time.Minute, "30m"},
		{"12h_preset", 720 * time.Minute, "30m"},
		{"1d_preset", 1440 * time.Minute, "60m"},
		{"7d_preset", 10080 * time.Minute, "60m"},
		// Off-preset windows hit the algorithm fallback.
		{"45m_off_preset", 45 * time.Minute, "2m"},  // 45/16 = 2
		{"6h_off_preset", 360 * time.Minute, "22m"}, // 360/16 = 22
		// Above 16h the step would exceed 60 → capped at 60m.
		{"3d_off_preset", 4320 * time.Minute, "60m"}, // also a preset → 60m
		{"1d_plus_one", 1500 * time.Minute, "60m"},   // 1500/16 = 93 → cap 60
		// Tiny windows < 16 min hit the "bump to 10-bucket" fallback.
		{"1m_tiny", 1 * time.Minute, "1m"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := gpuTrendStep(base, base.Add(tc.dur))
			if got != tc.want {
				t.Errorf("gpuTrendStep(+%s) = %q, want %q", tc.dur, got, tc.want)
			}
		})
	}
	// Defensive: zero / inverted ranges fall back to "1m" (avoid
	// dividing by zero downstream).
	if got := gpuTrendStep(base, base); got != "1m" {
		t.Errorf("zero range: got %q, want 1m", got)
	}
	if got := gpuTrendStep(base, base.Add(-time.Hour)); got != "1m" {
		t.Errorf("inverted range: got %q, want 1m", got)
	}
}

// TestBuildGPUDetailFullEnvelope_PartialFailure and
// TestBuildGPUTaskDetailFullEnvelope_TimeSlicingSkipsAllocation moved
// to cli/cmd/ctl/dashboard/overview/gpu/detail_test.go (next to the
// builders they exercise; those builders are cmd-side because they
// orchestrate per-leaf cobra-bound concurrency).

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func newTestClient(srv *httptest.Server) *Client {
	rp := &credential.ResolvedProfile{
		OlaresID:     "alice@olares.com",
		DashboardURL: srv.URL,
	}
	return NewClient(srv.Client(), rp)
}

// captureStderr swaps os.Stderr for the duration of fn, then restores
// it and returns whatever fn wrote. Used by the gpuAdvisory tests to
// assert the SPA-aligned stderr hint is emitted in table mode.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	prev := os.Stderr
	os.Stderr = w
	done := make(chan string, 1)
	go func() {
		buf, _ := io.ReadAll(r)
		done <- string(buf)
	}()
	fn()
	_ = w.Close()
	os.Stderr = prev
	return <-done
}
