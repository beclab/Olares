package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The fallback is the only reader of GET /console/api/model-apps, and what it
// needs off a row is two fields nested differently from every other list the
// CLI reads. So the test pins the path and the shape together: a rename on
// either side that stops them lining up is silent otherwise, because a lookup
// that finds nothing is indistinguishable from an application that is not
// there.
func newModelAppsClient(t *testing.T, body string) (*preparedClient, *[]string) {
	t.Helper()
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path != epModelApps {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": "not_found"}})
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &preparedClient{router: newRouterClient(srv.Client(), srv.URL, "alice@example.com")}, &paths
}

const modelAppsBody = `{"items":[
  {"app_name":"llamacppllmbasev3","install":{"installed":false,"provider_id":""}},
  {"app_name":"vllmgptoss20bv2","install":{"installed":true,"status":"running","provider_id":"p-9"}}
]}`

func TestProviderIDFromModelApps(t *testing.T) {
	pc, paths := newModelAppsClient(t, modelAppsBody)

	if got := providerIDFromModelApps(context.Background(), pc, "vllmgptoss20bv2"); got != "p-9" {
		t.Errorf("provider id: got %q want %q", got, "p-9")
	}
	if len(*paths) != 1 || (*paths)[0] != epModelApps {
		t.Errorf("asked for %v, want one GET %s", *paths, epModelApps)
	}
}

// An application the Market publishes but nobody installed has no provider to
// name, and a template can never have one. Returning its empty id would send
// the caller on to fetch a provider that does not exist.
func TestProviderIDFromModelAppsSkipsRowsWithNoProvider(t *testing.T) {
	pc, _ := newModelAppsClient(t, modelAppsBody)

	if got := providerIDFromModelApps(context.Background(), pc, "llamacppllmbasev3"); got != "" {
		t.Errorf("template provider id: got %q want empty", got)
	}
}

// This runs after the caller's own lookups have already failed, so its failure
// has to read as "not found" rather than replace their error with one about a
// route the user never named. A Router too old for this path answers 404.
func TestProviderIDFromModelAppsIsQuietWhenTheRouteIsGone(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found","message":"no route"}}`))
	}))
	t.Cleanup(srv.Close)
	pc := &preparedClient{router: newRouterClient(srv.Client(), srv.URL, "alice@example.com")}

	if got := providerIDFromModelApps(context.Background(), pc, "vllmgptoss20bv2"); got != "" {
		t.Errorf("got %q, want empty so the caller keeps its own error", got)
	}
}
