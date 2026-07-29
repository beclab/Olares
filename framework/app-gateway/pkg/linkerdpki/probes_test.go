package linkerdpki

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func probeStatus(h http.HandlerFunc) int {
	rr := httptest.NewRecorder()
	h(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	return rr.Code
}

// TC-PKI-G05: /healthz stays 200 while the heartbeat is refreshed within the
// 180s window (simulating ticker wake-ups during a long interval) and flips to
// 503 once the heartbeat goes stale.
func TestHealthHandlerHeartbeat(t *testing.T) {
	cur := time.Now()
	p := NewProbeState(24 * time.Hour)
	p.now = func() time.Time { return cur }

	p.Heartbeat()
	cur = cur.Add(170 * time.Second)
	if got := probeStatus(p.HealthHandler); got != http.StatusOK {
		t.Fatalf("fresh heartbeat: got %d, want 200", got)
	}

	// A loop wake-up refreshes the heartbeat, keeping liveness healthy.
	p.Heartbeat()
	cur = cur.Add(170 * time.Second)
	if got := probeStatus(p.HealthHandler); got != http.StatusOK {
		t.Fatalf("refreshed heartbeat: got %d, want 200", got)
	}

	cur = cur.Add(200 * time.Second)
	if got := probeStatus(p.HealthHandler); got != http.StatusServiceUnavailable {
		t.Fatalf("stale heartbeat: got %d, want 503", got)
	}
}

// TC-PKI-G06: /readyz is 503 with no success, 200 within 2x interval, and 503
// once the last success is older than 2x interval.
func TestReadyHandlerFreshness(t *testing.T) {
	cur := time.Now()
	p := NewProbeState(24 * time.Hour) // readiness window = 48h
	p.now = func() time.Time { return cur }

	if got := probeStatus(p.ReadyHandler); got != http.StatusServiceUnavailable {
		t.Fatalf("no success yet: got %d, want 503", got)
	}

	p.MarkSuccess()
	cur = cur.Add(47 * time.Hour)
	if got := probeStatus(p.ReadyHandler); got != http.StatusOK {
		t.Fatalf("fresh success: got %d, want 200", got)
	}

	cur = cur.Add(2 * time.Hour) // 49h since last success
	if got := probeStatus(p.ReadyHandler); got != http.StatusServiceUnavailable {
		t.Fatalf("stale success: got %d, want 503", got)
	}
}

// TC-PKI-G07: /startupz is 503 until both the client is ready and one reconcile
// attempt has been made, then 200.
func TestStartupHandlerAttempt(t *testing.T) {
	p := NewProbeState(24 * time.Hour)

	if got := probeStatus(p.StartupHandler); got != http.StatusServiceUnavailable {
		t.Fatalf("before client ready: got %d, want 503", got)
	}

	p.MarkClientReady()
	if got := probeStatus(p.StartupHandler); got != http.StatusServiceUnavailable {
		t.Fatalf("before first attempt: got %d, want 503", got)
	}

	p.MarkAttempted()
	if got := probeStatus(p.StartupHandler); got != http.StatusOK {
		t.Fatalf("after first attempt: got %d, want 200", got)
	}
}
