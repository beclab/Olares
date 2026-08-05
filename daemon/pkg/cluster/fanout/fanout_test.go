package fanout

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// localDispatcher points the fan-out at a test server instead of the fixed
// olaresd port, which is the only way to exercise it off a real cluster.
func localDispatcher(t *testing.T, srv *httptest.Server) (*Dispatcher, NodeTarget) {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("test server port: %v", err)
	}
	return &Dispatcher{PeerPath: "/command/power-node", Port: port},
		NodeTarget{Name: "worker-1", IP: u.Hostname()}
}

// A node-local endpoint may require more than the access token. Whatever the
// caller presented has to reach the second hop, or the node refuses work the
// master was authorized to ask for.
func TestDispatcherForwardsExtraHeaders(t *testing.T) {
	var (
		mu   sync.Mutex
		got  http.Header
		body []byte
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = r.Header.Clone()
		body, _ = io.ReadAll(r.Body)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":200}`))
	}))
	defer srv.Close()

	d, target := localDispatcher(t, srv)
	d.AuthToken = "token-1"
	d.Headers = map[string]string{"X-Signature": "jws-1"}

	results := d.Run(context.Background(), []NodeTarget{target}, func(NodeTarget) any {
		return map[string]string{"type": "reboot"}
	})

	if len(results) != 1 || results[0].Status != StatusOK {
		t.Fatalf("results = %+v", results)
	}

	mu.Lock()
	defer mu.Unlock()
	if got.Get("X-Authorization") != "token-1" {
		t.Errorf("access token not forwarded: %v", got)
	}
	if got.Get("X-Signature") != "jws-1" {
		t.Errorf("extra header not forwarded: %v", got)
	}
	if !strings.Contains(string(body), "reboot") {
		t.Errorf("payload = %s", body)
	}
}

func TestDispatcherReportsANodeThatRefuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":403,"message":"not the owner of this Olares"}`))
	}))
	defer srv.Close()

	d, target := localDispatcher(t, srv)
	results := d.Run(context.Background(), []NodeTarget{target}, func(NodeTarget) any { return nil })

	if len(results) != 1 {
		t.Fatalf("results = %+v", results)
	}
	if results[0].Status != StatusError {
		t.Errorf("status = %q, want error", results[0].Status)
	}
	if !strings.Contains(results[0].Err, "403") {
		t.Errorf("err = %q, want the node's own refusal", results[0].Err)
	}
}

func TestNodeURLBracketsIPv6Addresses(t *testing.T) {
	got := nodeURL("fd00::12", 18088, "/system/node-status")
	if got != "http://[fd00::12]:18088/system/node-status" {
		t.Fatalf("nodeURL = %q", got)
	}
}
