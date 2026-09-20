package clusterctx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONFormatsPersistent459WithoutLeakingBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(459)
		_, _ = w.Write([]byte(`{"session_id":"secret-edge-jwt"}`))
	}))
	t.Cleanup(server.Close)

	client := NewHTTPClient(server.Client(), server.URL, "alice@olares.com")
	err := client.DoJSON(context.Background(), http.MethodGet, "/capi/version", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "profile login --olares-id alice@olares.com") {
		t.Fatalf("459 should point to login: %q", got)
	}
	if strings.Contains(got, "secret-edge-jwt") || strings.Contains(got, "session_id") {
		t.Fatalf("459 error leaked response body: %q", got)
	}
}
