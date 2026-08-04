package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newListenersServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/listeners" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

// useAdminAddr points the readiness check at addr for the duration of the test.
func useAdminAddr(t *testing.T, addr string) {
	t.Helper()
	old := envoyAdminAddr
	envoyAdminAddr = addr
	t.Cleanup(func() { envoyAdminAddr = old })
}

func adminHostPort(ts *httptest.Server) string {
	return strings.TrimPrefix(ts.URL, "http://")
}

const sampleListeners = `http_redirect_81::0.0.0.0:81
https_443::0.0.0.0:443
https_pp_444::0.0.0.0:444
stream_udp_olarestest004_53::0.0.0.0:53
`

// All core ingress listeners are active → ready.
func TestReadyCheck_AllActive_Ready(t *testing.T) {
	ts := newListenersServer(t, http.StatusOK, sampleListeners)
	defer ts.Close()
	useAdminAddr(t, adminHostPort(ts))

	s := &XdsServer{}
	if err := s.ReadyCheck(nil); err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
}

// A core ingress listener missing from Envoy → not ready, and the error names
// the missing listener.
func TestReadyCheck_MissingCritical_NotReady(t *testing.T) {
	body := "http_redirect_81::0.0.0.0:81\nhttps_443::0.0.0.0:443\n" // no 444
	ts := newListenersServer(t, http.StatusOK, body)
	defer ts.Close()
	useAdminAddr(t, adminHostPort(ts))

	s := &XdsServer{}
	err := s.ReadyCheck(nil)
	if err == nil {
		t.Fatal("expected not ready when https_pp_444 is missing")
	}
	if !strings.Contains(err.Error(), "https_pp_444") {
		t.Fatalf("error should name the missing listener, got %v", err)
	}
}

// A per-app stream listener failing to bind (absent from /listeners) does not
// affect readiness, since it is not a critical listener.
func TestReadyCheck_AppStreamListenerIrrelevant(t *testing.T) {
	body := "https_443::0.0.0.0:443\nhttps_pp_444::0.0.0.0:444\n" // 53 never bound
	ts := newListenersServer(t, http.StatusOK, body)
	defer ts.Close()
	useAdminAddr(t, adminHostPort(ts))

	s := &XdsServer{}
	if err := s.ReadyCheck(nil); err != nil {
		t.Fatalf("app stream listener absence must not affect readiness, got %v", err)
	}
}

// Envoy admin unreachable → not ready.
func TestReadyCheck_AdminUnreachable_NotReady(t *testing.T) {
	ts := newListenersServer(t, http.StatusOK, sampleListeners)
	addr := adminHostPort(ts)
	ts.Close() // now unreachable
	useAdminAddr(t, addr)

	s := &XdsServer{}
	if err := s.ReadyCheck(nil); err == nil {
		t.Fatal("expected not ready when admin is unreachable")
	}
}

func TestEvaluateReadiness(t *testing.T) {
	active := map[string]bool{"https_443": true, "https_pp_444": true}
	if err := evaluateReadiness([]string{"https_443", "https_pp_444"}, active); err != nil {
		t.Fatalf("expected ready, got %v", err)
	}
	err := evaluateReadiness([]string{"https_443", "https_pp_444"}, map[string]bool{"https_443": true})
	if err == nil || !strings.Contains(err.Error(), "https_pp_444") {
		t.Fatalf("expected error naming https_pp_444, got %v", err)
	}
}

func TestFetchActiveListeners_Parsing(t *testing.T) {
	ts := newListenersServer(t, http.StatusOK, sampleListeners)
	defer ts.Close()

	active, err := fetchActiveListeners(context.Background(), adminHostPort(ts))
	if err != nil {
		t.Fatalf("fetchActiveListeners: %v", err)
	}
	for _, name := range []string{"http_redirect_81", "https_443", "https_pp_444", "stream_udp_olarestest004_53"} {
		if !active[name] {
			t.Errorf("expected listener %q to be parsed as active", name)
		}
	}
	if len(active) != 4 {
		t.Errorf("expected 4 active listeners, got %d", len(active))
	}
}
