package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"testing"

	apiRuntime "bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"github.com/golang-jwt/jwt/v5"
)

func testHandler(t *testing.T, upstream *httptest.Server, admin bool) *Handler {
	t.Helper()
	target, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{
		next:             http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }),
		maxRequestBytes:  defaultMaxRequestBytes,
		maxResponseBytes: defaultMaxResponseBytes,
		streams:          newStreamLimiter(defaultMaxStreams),
		resolveAdmin: func(context.Context, string) (bool, error) {
			return admin, nil
		},
	}
	h.proxy = &httputil.ReverseProxy{
		Director:       h.director(target),
		Transport:      upstream.Client().Transport,
		ModifyResponse: h.modifyResponse,
		ErrorHandler:   h.proxyError,
	}
	return h
}

func authenticatedRequest(t *testing.T, method, target string, body io.Reader) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(constants.HeaderBflUserKey, "alice")
	req.Header.Set(constants.UserAuthorizationTokenKey, signedToken(t, "alice"))
	return req
}

func signedToken(t *testing.T, username string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &apiRuntime.Claims{Username: username})
	signed, err := token.SignedString(constants.KubeSphereJwtKey)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func withTestUsername(t *testing.T) {
	t.Helper()
	original := constants.Username
	originalKey := constants.KubeSphereJwtKey
	constants.Username = "alice"
	constants.KubeSphereJwtKey = []byte("proxy-test-signing-key")
	t.Cleanup(func() {
		constants.Username = original
		constants.KubeSphereJwtKey = originalKey
	})
}

func TestProxyFiltersNamespaceListAndSanitizesIdentityHeaders(t *testing.T) {
	withTestUsername(t)
	var gotHeaders http.Header
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		gotHeaders = req.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[{"metadata":{"name":"app-alice"}},{"metadata":{"name":"app-bob"}},{"metadata":{"name":"photos-shared"}},{"metadata":{"name":"kube-system"}}],"totalItems":4}`)
	}))
	defer upstream.Close()

	req := authenticatedRequest(t, http.MethodGet, "http://bfl.test"+resourcesNamespacesPath, nil)
	callerToken := req.Header.Get(constants.UserAuthorizationTokenKey)
	req.Header.Set("Authorization", "Bearer untrusted")
	req.Header.Set("X-Forwarded-Host", "attacker.test")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "cookie-token"})
	recorder := httptest.NewRecorder()
	testHandler(t, upstream, false).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var payload struct {
		Items      []map[string]any `json:"items"`
		TotalItems int              `json:"totalItems"`
		Total      int              `json:"total"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) != 1 || nestedString(payload.Items[0], "metadata", "name") != "app-alice" {
		t.Fatalf("items = %#v", payload.Items)
	}
	if payload.TotalItems != 1 || payload.Total != 4 {
		t.Fatalf("totals = %d/%d", payload.TotalItems, payload.Total)
	}
	if got := gotHeaders.Get("Authorization"); got != "Bearer "+callerToken {
		t.Fatalf("authorization = %q", got)
	}
	for _, name := range []string{constants.HeaderBflUserKey, constants.UserAuthorizationTokenKey, "Cookie", "X-Forwarded-Host"} {
		if got := gotHeaders.Get(name); got != "" {
			t.Fatalf("upstream %s = %q", name, got)
		}
	}
}

func TestAdminResponsePassesThrough(t *testing.T) {
	withTestUsername(t)
	const original = `{"items":[{"metadata":{"name":"app-alice"}},{"metadata":{"name":"app-bob"}}],"totalItems":2}`
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, original)
	}))
	defer upstream.Close()

	recorder := httptest.NewRecorder()
	testHandler(t, upstream, true).ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "http://bfl.test"+resourcesNamespacesPath, nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != original {
		t.Fatalf("status/body = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestGzipMonitoringResponseIsFiltered(t *testing.T) {
	withTestUsername(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)
		_, _ = io.WriteString(writer, `[{"metric_name":"m","data":{"result":[{"metric":{"namespace":"app-alice"}},{"metric":{"namespace":"app-bob"}},{"metric":{"namespace":"photos-shared"}}]}}]`)
		_ = writer.Close()
	}))
	defer upstream.Close()

	req := authenticatedRequest(t, http.MethodGet, "http://bfl.test"+monitoringNamespaces, nil)
	req.Header.Set("Accept-Encoding", "gzip")
	recorder := httptest.NewRecorder()
	testHandler(t, upstream, false).ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	reader, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(decoded), "app-bob") || !strings.Contains(string(decoded), "app-alice") || !strings.Contains(string(decoded), "photos-shared") {
		t.Fatalf("filtered response = %s", decoded)
	}
}

func TestProtectedMalformedResponseFailsClosed(t *testing.T) {
	withTestUsername(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `secret malformed response`)
	}))
	defer upstream.Close()

	recorder := httptest.NewRecorder()
	testHandler(t, upstream, false).ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "http://bfl.test"+resourcesNamespacesPath, nil))
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "secret") {
		t.Fatalf("upstream body leaked: %s", recorder.Body.String())
	}
}

func TestMonitoringPostRejectsOversizedJSON(t *testing.T) {
	withTestUsername(t)
	var called bool
	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer upstream.Close()
	h := testHandler(t, upstream, false)
	h.maxRequestBytes = 16
	req := authenticatedRequest(t, http.MethodPost, "http://bfl.test"+monitoringNamespaces, strings.NewReader(`{"metrics_filter":"0123456789"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusRequestEntityTooLarge || called {
		t.Fatalf("status/called = %d/%v", recorder.Code, called)
	}
}

func TestPodsRequestForNormalUserForcesUnlimitedList(t *testing.T) {
	withTestUsername(t)
	var limit string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		limit = req.URL.Query().Get("limit")
		_, _ = io.WriteString(w, `{"items":[],"totalItems":0}`)
	}))
	defer upstream.Close()
	recorder := httptest.NewRecorder()
	testHandler(t, upstream, false).ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "http://bfl.test"+resourcesPodsPath+"?limit=10", nil))
	if recorder.Code != http.StatusOK || limit != "-1" {
		t.Fatalf("status/limit = %d/%q", recorder.Code, limit)
	}
}

func TestStreamLimiter(t *testing.T) {
	limiter := newStreamLimiter(2)
	if !limiter.acquire("alice") || !limiter.acquire("alice") || limiter.acquire("alice") {
		t.Fatal("unexpected limiter acquisition result")
	}
	limiter.release("alice")
	if !limiter.acquire("alice") {
		t.Fatal("released slot was not reusable")
	}

	var wait sync.WaitGroup
	for i := 0; i < 20; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if limiter.acquire("bob") {
				limiter.release("bob")
			}
		}()
	}
	wait.Wait()
}

func TestProxyPathAllowlist(t *testing.T) {
	for _, path := range []string{"/kapis/version", "/api/v1/pods", "/apis/apps/v1/deployments"} {
		if !proxyPath(path) {
			t.Fatalf("expected %q to be proxied", path)
		}
	}
	for _, path := range []string{"/capi/app/detail", "/api/profile/init", "/api/v2", "/kapis-evil"} {
		if proxyPath(path) {
			t.Fatalf("expected %q to be delegated", path)
		}
	}
}

func TestAllowedUserScope(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/api/v1/namespaces/app-alice/pods/p1", want: true},
		{path: "/apis/apps/v1/namespaces/photos-shared/deployments/d1", want: true},
		{path: "/api/v1/namespaces/app-bob/pods/p1", want: false},
		{path: "/api/v1/namespaces/kube-system/pods/p1", want: false},
		{path: "/kapis/monitoring.kubesphere.io/v1alpha3/users/alice", want: true},
		{path: "/kapis/monitoring.kubesphere.io/v1alpha3/users/bob", want: false},
		{path: resourcesNamespacesPath, want: true},
	}
	for _, tt := range tests {
		if got := allowedUserScope(tt.path, "alice"); got != tt.want {
			t.Errorf("allowedUserScope(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestUpgradeIsAuthenticatedAndProxied(t *testing.T) {
	withTestUsername(t)
	callerToken := signedToken(t, "alice")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.Header.Get("Authorization"); got != "Bearer "+callerToken {
			http.Error(w, "missing forwarded token", http.StatusUnauthorized)
			return
		}
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijacking unsupported", http.StatusInternalServerError)
			return
		}
		conn, buffer, err := hijacker.Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = buffer.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\nhello")
		_ = buffer.Flush()
	}))
	defer upstream.Close()

	proxyServer := httptest.NewServer(testHandler(t, upstream, false))
	defer proxyServer.Close()
	proxyURL, err := url.Parse(proxyServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", proxyURL.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err = fmt.Fprintf(conn, "GET /api/v1/watch/pods HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nX-BFL-USER: alice\r\nCookie: auth_token=%s\r\n\r\n", proxyURL.Host, callerToken)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "101 Switching Protocols") {
		t.Fatalf("status = %q", status)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" {
			break
		}
	}
	payload := make([]byte, len("hello"))
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatal(err)
	}
	if string(payload) != "hello" {
		t.Fatalf("upgrade payload = %q", payload)
	}
}
