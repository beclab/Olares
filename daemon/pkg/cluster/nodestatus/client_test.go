package nodestatus

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
)

func envelope(t *testing.T, status Status) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{"code": 200, "message": "success", "data": status})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return body
}

func TestFromEnvelopeReadsTheNodeReply(t *testing.T) {
	body := envelope(t, Status{
		NodeName:     "worker-1",
		Role:         inventory.RoleWorker,
		Health:       HealthHealthy,
		Capabilities: map[string]Capability{CapPowerShutdown: {Supported: true}},
	})

	got, err := FromEnvelope(body)
	if err != nil {
		t.Fatalf("FromEnvelope: %v", err)
	}
	if got.NodeName != "worker-1" {
		t.Errorf("node name = %q", got.NodeName)
	}
	if c, ok := got.Capabilities[CapPowerShutdown]; !ok || !c.Supported {
		t.Errorf("capabilities lost in transit: %+v", got.Capabilities)
	}
}

func TestFromEnvelopeRejectsSomethingElse(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"not json", `<html>404 not found</html>`},
		{"no data", `{"code":200,"message":"success"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := FromEnvelope([]byte(tc.body)); err == nil {
				t.Fatal("a reply that is not a node status was accepted")
			}
		})
	}
}

func TestFetchReadsTheNodeLocalEndpoint(t *testing.T) {
	var gotPath, gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotToken = r.URL.Path, r.Header.Get(authHeader)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(envelope(t, Status{NodeName: "worker-1", Health: HealthHealthy}))
	}))
	defer srv.Close()

	got, err := Fetch(context.Background(), srv.URL, "token-1")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if gotPath != Path {
		t.Errorf("path = %q, want %q", gotPath, Path)
	}
	if gotToken != "token-1" {
		t.Errorf("token = %q, want the caller's", gotToken)
	}
	if got.NodeName != "worker-1" || got.Health != HealthHealthy {
		t.Errorf("status = %+v", got)
	}
}

// A node that refuses is not a node that answered with nothing to declare.
func TestFetchRejectsANonOKReply(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":403,"message":"forbidden"}`))
	}))
	defer srv.Close()

	if _, err := Fetch(context.Background(), srv.URL, "token-1"); err == nil {
		t.Fatal("a 403 was read as a node status")
	}
}

// How long a node may take to answer is the caller's decision, because a
// precheck and a page load do not have the same patience.
func TestFetchIsBoundedByTheCallersContext(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-release
	}))
	defer func() {
		close(release)
		srv.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	if _, err := Fetch(ctx, srv.URL, "token-1"); err == nil {
		t.Fatal("a node that never answered was read as a status")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("the call ran for %v past its deadline", elapsed)
	}
}
