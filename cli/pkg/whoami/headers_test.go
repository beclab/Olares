package whoami

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSONSendsContextRequestHeaders(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	client := NewHTTPClient(srv.Client(), srv.URL, "alice@olares.com")
	ctx := ContextWithRequestHeaders(context.Background(), map[string]string{
		"Idempotency-Key": "abc-123",
	})
	var out map[string]any
	if err := client.DoJSON(ctx, http.MethodPost, "/api/download", map[string]string{"url": "https://x"}, &out); err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
	if got != "abc-123" {
		t.Fatalf("Idempotency-Key = %q, want abc-123", got)
	}
}
