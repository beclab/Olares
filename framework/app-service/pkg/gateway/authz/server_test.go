package authz

import (
	"bytes"
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

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
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
	recordAuthzMetrics("allow", "-", time.Millisecond)
	rec = httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics: %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	out := string(body)
	for _, want := range []string{
		"app_service_ext_authz_decisions_total",
		"app_service_ext_authz_latency_ms",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("metrics body missing %q: %s", want, out)
		}
	}
}

func discardAuditor() Auditor {
	return Auditor{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
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

func snapshotInCluster(enabled bool) cluster.SnapshotFunc {
	return func(context.Context) (cluster.Snapshot, error) {
		return cluster.Snapshot{InClusterGatewayEnabled: enabled}, nil
	}
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
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(false),
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
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(false),
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
		allow:        false,
		audit:        discardAuditor(),
		hostUser:     HostUserConfig{Enabled: false},
		snapshotFunc: snapshotInCluster(false),
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

func TestAuthzHandler_InCluster_L5dAllow(t *testing.T) {
	l5d := "default.user-space-alice.serviceaccount.identity.linkerd.cluster.local"
	h := &authzHandler{
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(true),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("a1b2c3d4.alice.olares.com",
		map[string]string{"l5d-client-id": l5d}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	ok, isOK := resp.HttpResponse.(*authv3.CheckResponse_OkResponse)
	if !isOK {
		t.Fatalf("expected OkResponse, got %#v", resp.HttpResponse)
	}
	if len(ok.OkResponse.Headers) == 0 || ok.OkResponse.Headers[0].Header.Value != "alice" {
		t.Fatalf("headers = %#v", ok.OkResponse.Headers)
	}
}

func TestAuthzHandler_InCluster_ViewerMismatchDeny(t *testing.T) {
	l5d := "default.user-space-alice.serviceaccount.identity.linkerd.cluster.local"
	h := &authzHandler{
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(true),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("a1b2c3d4.bob.olares.com",
		map[string]string{"l5d-client-id": l5d}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if _, ok := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse); !ok {
		t.Fatalf("expected deny, got %#v", resp.HttpResponse)
	}
}

// requirement: in-cluster ext_authz must remain reachable for callers that send
// HTTPS to the gateway :443 listener; Linkerd transparently tunnels TLS so no
// l5d-client-id reaches ext_authz; the matrix labels this the L1' compat path.
// behavior: cluster-internal Shared host + no l5d-client-id + X-BFL-USER equal
// to the host viewer label resolves to Allow (Phase A in-cluster shared allow)
// with the host-user decision attached.
// test: TC-005 (https compat) — covers the same logical path EG :443 takes
// when chart tls.enabled=true is in effect. No production behaviour changes.
func TestAuthzHandler_Check_InCluster_HttpsCompatPath_TC005_Allow(t *testing.T) {
	h := &authzHandler{
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(true),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("a1b2c3d4.alice.olares.com",
		map[string]string{"x-bfl-user": "alice"}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	ok, isOK := resp.HttpResponse.(*authv3.CheckResponse_OkResponse)
	if !isOK {
		t.Fatalf("expected Allow on weak compat path, got %#v", resp.HttpResponse)
	}
	if len(ok.OkResponse.Headers) == 0 || ok.OkResponse.Headers[0].Header.Value != "alice" {
		t.Fatalf("x-bfl-user header not echoed: %#v", ok.OkResponse.Headers)
	}
}

// requirement: weak path must not let a caller-supplied X-BFL-USER overrule the
// host viewer; mismatch denies with the stable INVALID_HOST_USER code.
// behavior: HostUser decider runs even when l5d is absent, so the host viewer
// label still wins; mismatched X-BFL-USER yields 403 INVALID_HOST_USER.
// test: TC-005 negative — keeps the §2.7.1 layering conclusion intact (the
// compat path is "usable", not "trusted").
func TestAuthzHandler_Check_InCluster_HttpsCompat_DenyMismatch_TC005(t *testing.T) {
	h := &authzHandler{
		allow:        true,
		audit:        discardAuditor(),
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(true),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := client.Check(ctx, makeReq("a1b2c3d4.alice.olares.com",
		map[string]string{"x-bfl-user": "bob"}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	denied, ok := resp.HttpResponse.(*authv3.CheckResponse_DeniedResponse)
	if !ok {
		t.Fatalf("expected Deny on viewer mismatch, got %#v", resp.HttpResponse)
	}
	if denied.DeniedResponse.Status.Code != envoytypev3.StatusCode_Forbidden {
		t.Fatalf("status: %v", denied.DeniedResponse.Status)
	}
	if !strings.Contains(denied.DeniedResponse.Body, CodeInvalidHostUser) {
		t.Fatalf("body lacks %s: %q", CodeInvalidHostUser, denied.DeniedResponse.Body)
	}
}

// requirement: weak compat path must be observable so operators can confirm a
// request did NOT carry mesh identity; the e2e harness greps the audit line
// for l5d_present=false + decision=allow on the compat scenario.
// behavior: when Check accepts a request with no l5d-client-id, the audit
// emits decision=allow, l5d_present=false, via=incluster_shared_allow.
// test: TC-005 observability — pins the audit contract the P4-https e2e
// step relies on. Changing the field name without updating the e2e harness
// would surface here first.
func TestAuthzHandler_Audit_TC005_HttpsCompat_L5dPresentFalse(t *testing.T) {
	var buf bytes.Buffer
	h := &authzHandler{
		allow:        true,
		audit:        Auditor{Logger: slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))},
		hostUser:     DefaultHostUserConfig(),
		snapshotFunc: snapshotInCluster(true),
	}
	conn, cleanup := startBufServer(t, h)
	defer cleanup()
	client := authv3.NewAuthorizationClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := client.Check(ctx, makeReq("a1b2c3d4.alice.olares.com",
		map[string]string{"x-bfl-user": "alice"})); err != nil {
		t.Fatalf("Check: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"decision=allow",
		"l5d_present=false",
		"via=incluster_shared_allow",
		"authority=a1b2c3d4.alice.olares.com",
		"viewer_host=alice",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("audit missing %q; full line: %s", want, out)
		}
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
