package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	statusv1 "google.golang.org/genproto/googleapis/rpc/status"
)

func TestReadyCheckRequiresCurrentACKAndActiveHTTPS(t *testing.T) {
	h := newHealthTracker()
	h.expect(&cachev3.Snapshot{VersionMap: map[string]map[string]string{
		resourcev3.ListenerType: {"https_443": "new", "https_pp_444": "pass", "stream_udp_dns_53": "optional"},
		resourcev3.RouteType:    {"https_443_alice_routes": "route", "https_444_alice_routes": "proxy-route"},
		resourcev3.ClusterType:  {"app_gateway_edge_forward": "entry", "authelia_backend_alice": "auth"},
		resourcev3.SecretType:   {"l4-app-gateway-client": "cert", "l4-app-gateway-ca": "ca"},
	}})
	callbacks := h.callbacks()
	if err := callbacks.OnDeltaStreamOpen(t.Context(), 1, ""); err != nil {
		t.Fatal(err)
	}
	for kind, names := range map[string]map[string]string{
		resourcev3.RouteType:   {"https_443_alice_routes": "route", "https_444_alice_routes": "proxy-route"},
		resourcev3.ClusterType: {"app_gateway_edge_forward": "entry", "authelia_backend_alice": "auth"},
		resourcev3.SecretType:  {"l4-app-gateway-client": "cert", "l4-app-gateway-ca": "ca"},
	} {
		// The local Envoy can reconnect with already applied resources.
		if err := callbacks.OnStreamDeltaRequest(1, &discoverygrpc.DeltaDiscoveryRequest{TypeUrl: kind, InitialResourceVersions: names}); err != nil {
			t.Fatal(err)
		}
	}
	active := "old"
	useAdminHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config_dump" || r.URL.Query().Get("resource") != "dynamic_listeners" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"configs": []any{
			map[string]any{"name": "https_443", "active_state": map[string]string{"version_info": active},
				"warming_state": map[string]string{"version_info": "new"}},
			map[string]any{"name": "https_pp_444", "active_state": map[string]string{"version_info": "pass"}},
		}})
	}))

	s := &XdsServer{health: h}
	check := func(want string) {
		t.Helper()
		err := s.ReadyCheck(nil)
		if want == "" && err != nil || want != "" && (err == nil || !strings.Contains(err.Error(), want)) {
			t.Fatalf("readiness = %v, want %q", err, want)
		}
	}
	check("has not acknowledged")
	send := func(nonce string) {
		t.Helper()
		callbacks.OnStreamDeltaResponse(1, nil, &discoverygrpc.DeltaDiscoveryResponse{
			TypeUrl: resourcev3.ListenerType, Nonce: nonce,
			Resources: []*discoverygrpc.Resource{{Name: "https_443", Version: "new"}, {Name: "https_pp_444", Version: "pass"}},
		})
	}
	ack := func(nonce string, detail *statusv1.Status) {
		t.Helper()
		if err := callbacks.OnStreamDeltaRequest(1, &discoverygrpc.DeltaDiscoveryRequest{
			TypeUrl: resourcev3.ListenerType, ResponseNonce: nonce, ErrorDetail: detail,
		}); err != nil {
			t.Fatal(err)
		}
	}
	send("one")
	ack("stale", nil)
	check("has not acknowledged")
	ack("one", &statusv1.Status{Message: "test rejection"})
	check("has not acknowledged")
	send("two")
	ack("two", nil)
	check("not active at the current xDS version")
	active = "new"
	check("")
	callbacks.OnDeltaStreamClosed(1, nil)
	check("has not acknowledged")
}

func TestEntryReadinessRejectsMissingTLSMaterial(t *testing.T) {
	h := newHealthTracker()
	s := &XdsServer{health: h}
	if err := s.ReadyCheck(nil); err == nil || !strings.Contains(err.Error(), "snapshot not published") {
		t.Fatalf("unexpected initial result: %v", err)
	}
	h.expect(&cachev3.Snapshot{VersionMap: map[string]map[string]string{
		resourcev3.ListenerType: {"https_443": "http", "https_pp_444": "proxy"},
		resourcev3.ClusterType:  {"app_gateway_edge_forward": "entry"},
	}})
	if err := s.ReadyCheck(nil); err == nil || !strings.Contains(err.Error(), "l4-app-gateway-client") {
		t.Fatalf("snapshot without Entry TLS material passed: %v", err)
	}
}

type adminRoundTripper func(*http.Request) (*http.Response, error)

func (f adminRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func useAdminHandler(t *testing.T, handler http.Handler) {
	t.Helper()
	oldClient := adminClient
	adminClient = &http.Client{Transport: adminRoundTripper(func(r *http.Request) (*http.Response, error) {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		return recorder.Result(), nil
	})}
	t.Cleanup(func() { adminClient = oldClient })
}

func TestActiveListenerVersionsRejectsBrokenAdminEndpoint(t *testing.T) {
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusOK} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			useAdminHandler(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte("broken"))
			}))
			if _, err := activeListenerVersions(t.Context(), envoyAdminAddr); err == nil {
				t.Errorf("invalid admin response %d accepted", status)
			}
		})
	}
}
