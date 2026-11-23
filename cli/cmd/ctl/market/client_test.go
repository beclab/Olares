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

// A 403 Market wrote itself is a business refusal; its sentence is the answer
// and must reach the user, while an edge 403 keeps the identity hint.
func TestDoRequestPassesMarketRefusalThrough(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		notWant string
	}{
		{name: "market envelope", body: `{"success":false,"message":"shared apps can only be operated by administrators"}`, want: "API error (HTTP 403): shared apps can only be operated by administrators", notWant: "whoami"},
		{name: "edge page", body: `<html>Forbidden</html>`, want: "profile whoami --refresh"},
		{name: "edge json without envelope", body: `{"status":"KO","message":"Forbidden"}`, want: "profile whoami --refresh", notWant: "API error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)
			client := NewMarketClient(server.Client(), server.Client(), &credential.ResolvedProfile{
				MarketURL: server.URL,
				OlaresID:  "alice@olares.com",
			}, "")

			_, err := client.doRequest(context.Background(), http.MethodPost, "/apps/demo/install", nil)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not contain %q", err, tt.want)
			}
			if tt.notWant != "" && strings.Contains(err.Error(), tt.notWant) {
				t.Fatalf("error %q should not contain %q", err, tt.notWant)
			}
		})
	}
}
