package credential

import (
	"net/http"
	"strings"
	"testing"
)

func TestFormatHTTPAuthErrorDistinguishesPermissionFromAuthentication(t *testing.T) {
	const sensitiveBody = `{"message":"forbidden","session_id":"secret-edge-jwt"}`
	forbidden := FormatHTTPAuthError(http.StatusForbidden, []byte(sensitiveBody), "alice@olares.com")
	if got := forbidden.Error(); !strings.Contains(got, "profile whoami --refresh") || strings.Contains(got, "profile login") {
		t.Fatalf("403 should explain permission recovery without login: %q", got)
	}

	unauthorized := FormatHTTPAuthError(http.StatusUnauthorized, []byte(sensitiveBody), "alice@olares.com")
	if got := unauthorized.Error(); !strings.Contains(got, "profile login --olares-id alice@olares.com") {
		t.Fatalf("401 should point to login: %q", got)
	}

	edgeAuthFailure := FormatHTTPAuthError(459, []byte(sensitiveBody), "alice@olares.com")
	if got := edgeAuthFailure.Error(); !strings.Contains(got, "profile login --olares-id alice@olares.com") {
		t.Fatalf("459 should point to login: %q", got)
	}

	for name, err := range map[string]error{
		"403": forbidden,
		"401": unauthorized,
		"459": edgeAuthFailure,
	} {
		if got := err.Error(); strings.Contains(got, "secret-edge-jwt") || strings.Contains(got, "session_id") {
			t.Fatalf("%s auth error leaked response body: %q", name, got)
		}
	}
}
