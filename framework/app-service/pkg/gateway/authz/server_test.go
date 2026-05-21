package authz

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytypev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestParseSkipViewers(t *testing.T) {
	cases := map[string][]string{
		"":                  nil,
		"   ":               nil,
		"alice":             {"alice"},
		"Alice, Bob,,c-d ":  {"alice", "bob", "c-d"},
	}
	for in, want := range cases {
		got := ParseSkipViewers(in)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ParseSkipViewers(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestProbeServer_ReadyToggle(t *testing.T) {
	ready := &atomic.Bool{}
	ready.Store(true)
	srv := newProbeServer(":0", ready)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("readyz initial: %d", rec.Code)
	}
	ready.Store(false)
	rec = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz draining: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "app_service_ext_authz_phase") {
		t.Fatalf("metrics body missing gauge: %s", body)
	}
}

// startBufServer wires a bufconn-backed gRPC server for in-process Check tests.
func startBufServer(t *testing.T, h *authzHandler) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	authv3.RegisterAuthorizationServer(gs, h)
	go func() {
		if err := gs.Serve(lis); err != nil && !strings.Contains(err.Error(), "closed") {
			t.Logf("buf serve: %v", err)
		}
	}()
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) { return lis.Dial() }),
	)
	if err != nil {
		t.Fatalf("bufnet dial: %v", err)
	}
	return conn, func() { _ = conn.Close(); gs.Stop() }
}

func makeReq(host string, headers map[string]string) *authv3.CheckRequest {
	return &authv3.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Host:    host,
					Path:    "/",
					Method:  "GET",
					Headers: headers,
				},
			},
		},
	}
}

func TestAuthzHandler_Check_Allow(t *testing.T) {
	h := &authzHandler{
		allow:    true,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		hostUser: DefaultHostUserConfig(),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("01234567.alice.olares.com",
		map[string]string{"x-bfl-user": "alice"}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if _, ok := resp.HttpResponse.(*authv3.CheckResponse_OkResponse); !ok {
		t.Fatalf("expected OkResponse, got %#v", resp.HttpResponse)
	}
}

func TestAuthzHandler_Check_DenyHostUser(t *testing.T) {
	h := &authzHandler{
		allow:    true,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		hostUser: DefaultHostUserConfig(),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("01234567.alice.olares.com",
		map[string]string{"x-bfl-user": "bob"}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	denied, ok := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse)
	if !ok {
		t.Fatalf("expected DeniedResponse, got %#v", resp.HttpResponse)
	}
	if denied.DeniedResponse.Status.Code != envoytypev3.StatusCode_Forbidden {
		t.Fatalf("unexpected denied status: %v", denied.DeniedResponse.Status)
	}
	if !strings.Contains(denied.DeniedResponse.Body, "INVALID_HOST_USER") {
		t.Fatalf("body lacks code: %q", denied.DeniedResponse.Body)
	}
}

func TestAuthzHandler_Check_DenyMode_HostUserDisabled(t *testing.T) {
	h := &authzHandler{
		allow:    false,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		hostUser: HostUserConfig{Enabled: false},
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("anything", nil))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	denied, ok := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse)
	if !ok {
		t.Fatalf("expected DeniedResponse, got %#v", resp.HttpResponse)
	}
	if denied.DeniedResponse.Status.Code != envoytypev3.StatusCode_Forbidden {
		t.Fatalf("unexpected denied status: %v", denied.DeniedResponse.Status)
	}
}

func TestServer_StartStop_GraceShutdown(t *testing.T) {
	srv := NewServer(ServerOptions{
		Enabled:   true,
		GRPCAddr:  "127.0.0.1:0",
		HTTPAddr:  "127.0.0.1:0",
		AllowMode: true,
		HostUser:  DefaultHostUserConfig(),
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	time.Sleep(150 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned err: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not exit after context cancel")
	}
}
