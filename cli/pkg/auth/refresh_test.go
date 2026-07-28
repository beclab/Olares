package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestRefreshUnauthorizedRequiresAppRejection pins the attribution rule behind
// ErrRefreshUnauthorized: only Authelia may condemn the grant.
//
// The caller (credential.Refresher) turns this sentinel into a *persisted*
// InvalidatedAt stamp, and that stamp is read by every client sharing the token
// store — including ones reaching the instance over a route that works. Olares'
// Envoy edge denies requests failing its RBAC policy (reaching the public
// entrance while the instance is VPN-only, say) with a bare text/plain 403 that
// Authelia never saw, so mapping status codes alone would strand a healthy
// grant until the user re-logs in.
func TestRefreshUnauthorizedRequiresAppRejection(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantDead bool
	}{
		{
			name:     "envoy rbac denial",
			status:   http.StatusForbidden,
			body:     "RBAC: access denied",
			wantDead: false,
		},
		{
			name:     "unauthorized with no body",
			status:   http.StatusUnauthorized,
			body:     "",
			wantDead: false,
		},
		{
			name:     "unauthorized with non-object json",
			status:   http.StatusUnauthorized,
			body:     `["access denied"]`,
			wantDead: false,
		},
		{
			name:     "gateway 502",
			status:   http.StatusBadGateway,
			body:     "upstream connect error",
			wantDead: false,
		},
		{
			name:     "authelia rejects the grant",
			status:   http.StatusUnauthorized,
			body:     `{"status":"KO","message":"Authentication failed, incorrect password."}`,
			wantDead: true,
		},
		{
			name:     "authelia forbids the grant",
			status:   http.StatusForbidden,
			body:     `{"status":"KO","message":"forbidden"}`,
			wantDead: true,
		},
		{
			name:     "ext-authz 459",
			status:   459,
			body:     `{"fa2":false,"session_id":"s","target_url":"https://auth.example.olares.cn"}`,
			wantDead: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			_, err := Refresh(context.Background(), RefreshRequest{
				AuthURL:      srv.URL,
				RefreshToken: "stored-refresh-token",
				Timeout:      5 * time.Second,
			})
			if err == nil {
				t.Fatal("Refresh: got nil error, want failure")
			}
			if got := errors.Is(err, ErrRefreshUnauthorized); got != tc.wantDead {
				t.Errorf("errors.Is(err, ErrRefreshUnauthorized) = %v, want %v (err = %v)",
					got, tc.wantDead, err)
			}
		})
	}
}

func TestRefreshReturnsRotatedToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"OK","data":{"access_token":"new-access","refresh_token":"new-refresh"}}`))
	}))
	defer srv.Close()

	tok, err := Refresh(context.Background(), RefreshRequest{
		AuthURL:      srv.URL,
		RefreshToken: "old-refresh",
		Timeout:      5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if tok.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "new-access")
	}
	if tok.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q, want %q", tok.RefreshToken, "new-refresh")
	}
}

// A server with rotation disabled omits refresh_token; the caller must keep
// using the one it already holds rather than storing an empty string.
func TestRefreshKeepsRefreshTokenWhenNotRotated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"OK","data":{"access_token":"new-access"}}`))
	}))
	defer srv.Close()

	tok, err := Refresh(context.Background(), RefreshRequest{
		AuthURL:      srv.URL,
		RefreshToken: "old-refresh",
		Timeout:      5 * time.Second,
	})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if tok.RefreshToken != "old-refresh" {
		t.Errorf("RefreshToken = %q, want the caller-supplied %q", tok.RefreshToken, "old-refresh")
	}
}
