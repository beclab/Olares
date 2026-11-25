package v2alpha1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Exercise Echo's actual target selection, including its raw WebSocket branch.
// The fixture starts after existing provider authorization selects a service.
func TestProviderProxyCannotReachNonce(t *testing.T) {
	var upstreamCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	for _, tc := range []struct {
		name, targetPath, requestPath string
		websocket                     bool
	}{
		{name: "direct", requestPath: "/permission/v1alpha1/nonce"},
		{name: "query", requestPath: "/permission/v1alpha1/nonce?x=1"},
		{name: "encoded", requestPath: "/permission/v1alpha1/%6eonce"},
		{name: "dot-segment", requestPath: "/other/../permission/v1alpha1/nonce"},
		{name: "duplicate-slash", requestPath: "//permission/v1alpha1/nonce"},
		{name: "target-prefix", targetPath: "/permission/v1alpha1", requestPath: "/nonce"},
		{name: "target-endpoint", targetPath: "/permission/v1alpha1/nonce", requestPath: "/"},
		{name: "websocket", requestPath: "/permission/v1alpha1/nonce", websocket: true},
		{name: "websocket-target", targetPath: "/permission/v1alpha1", requestPath: "/nonce", websocket: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := upstreamCalls.Load()
			s := NewRBACProxyServer(context.Background())
			s.proxy.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c echo.Context) error {
					c.Set(serviceKey, upstream.URL+tc.targetPath)
					return next(c)
				}
			})
			s.proxy.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{Balancer: s}))
			req := httptest.NewRequest(http.MethodGet, tc.requestPath, nil)
			req.Header.Set("Authorization", "Bearer forged")
			req.Header.Set("l5d-client-id", "authelia-backend.os-framework.serviceaccount.identity.linkerd.cluster.local")
			if tc.websocket {
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "websocket")
			}
			res := httptest.NewRecorder()
			s.proxy.ServeHTTP(res, req)
			if res.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", res.Code)
			}
			if res.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing no-store")
			}
			if upstreamCalls.Load() != before {
				t.Fatal("reserved request reached upstream")
			}
		})
	}
	s := NewRBACProxyServer(context.Background())
	s.proxy.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(serviceKey, upstream.URL)
			return next(c)
		}
	})
	s.proxy.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{Balancer: s}))
	res := httptest.NewRecorder()
	s.proxy.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/provider/ordinary", nil))
	if res.Code != http.StatusNoContent || upstreamCalls.Load() != 1 {
		t.Fatal("ordinary provider forwarding changed")
	}
}

func TestProviderTargetUnavailable(t *testing.T) {
	s := NewRBACProxyServer(context.Background())
	s.proxy.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{Balancer: s}))
	res := httptest.NewRecorder()
	s.proxy.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/ordinary", nil))
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", res.Code)
	}
}
