package market

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/credential"
)

func TestDoRequestFormatsPersistent459WithoutLeakingBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(459)
		_, _ = w.Write([]byte(`{"fa2":false,"session_id":"secret-edge-jwt","target_url":"/"}`))
	}))
	t.Cleanup(server.Close)

	client := NewMarketClient(server.Client(), server.Client(), &credential.ResolvedProfile{
		MarketURL: server.URL,
		OlaresID:  "alice@olares.com",
	}, "")
	_, err := client.doRequest(context.Background(), http.MethodGet, "/applications", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "HTTP 459") ||
		!strings.Contains(got, "profile login --olares-id alice@olares.com") {
		t.Fatalf("459 should point to login: %q", got)
	}
	if strings.Contains(got, "secret-edge-jwt") || strings.Contains(got, "session_id") {
		t.Fatalf("459 error leaked response body: %q", got)
	}
}

func TestDownloadStreamFormatsPersistent459WithoutLeakingBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(459)
		_, _ = w.Write([]byte(`{"session_id":"secret-edge-jwt"}`))
	}))
	t.Cleanup(server.Close)

	client := NewMarketClient(server.Client(), server.Client(), &credential.ResolvedProfile{
		MarketURL: server.URL,
		OlaresID:  "alice@olares.com",
	}, "")
	_, err := client.downloadStream(context.Background(), "/charts/example", url.Values{})
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
