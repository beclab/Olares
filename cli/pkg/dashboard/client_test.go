package dashboard

import (
	"strings"
	"testing"
)

func TestClassifyStatusFormats459WithoutLeakingBody(t *testing.T) {
	err := classifyStatus(
		459,
		"https://dashboard.example/api",
		[]byte(`{"session_id":"secret-edge-jwt"}`),
		"alice@olares.com",
	)
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
