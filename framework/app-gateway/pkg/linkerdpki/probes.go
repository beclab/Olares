package linkerdpki

import (
	"net/http"
	"sync"
	"time"
)

const (
	// livenessHeartbeatThreshold bounds /healthz freshness. The main loop must
	// refresh the heartbeat well within this window even while idling between
	// reconciles; 3xGUARDIAN_INTERVAL must not be used (see detailed design §2.3).
	livenessHeartbeatThreshold = 180 * time.Second

	// readinessIntervalMultiplier bounds /readyz on the last successful
	// reconcile relative to the reconcile interval (default 2x = 48h).
	readinessIntervalMultiplier = 2
)

// ProbeState tracks controller liveness/readiness for the HTTP probes exposed
// on GUARDIAN_HTTP_ADDR. It is safe for concurrent use.
type ProbeState struct {
	interval time.Duration
	now      func() time.Time

	mu          sync.RWMutex
	clientReady bool
	attempted   bool
	heartbeat   time.Time
	lastSuccess time.Time
}

// NewProbeState returns a ProbeState sized to the reconcile interval.
func NewProbeState(interval time.Duration) *ProbeState {
	return &ProbeState{interval: interval, now: time.Now}
}

// MarkClientReady records that the in-cluster client has been constructed and
// seeds the heartbeat so /healthz is fresh before the first reconcile.
func (p *ProbeState) MarkClientReady() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clientReady = true
	p.heartbeat = p.now()
}

// MarkAttempted records that at least one reconcile attempt has completed
// (success or predictable transient), satisfying /startupz.
func (p *ProbeState) MarkAttempted() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.attempted = true
}

// Heartbeat refreshes the liveness timestamp; called on every loop wake-up.
func (p *ProbeState) Heartbeat() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.heartbeat = p.now()
}

// MarkSuccess records a successful reconcile for /readyz and refreshes liveness.
func (p *ProbeState) MarkSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := p.now()
	p.lastSuccess = now
	p.heartbeat = now
}

// StartupHandler serves /startupz: 200 once the client is ready and one
// reconcile attempt has been made.
func (p *ProbeState) StartupHandler(w http.ResponseWriter, _ *http.Request) {
	p.mu.RLock()
	ready := p.clientReady && p.attempted
	p.mu.RUnlock()
	writeProbe(w, ready, "starting")
}

// HealthHandler serves /healthz: 200 while the heartbeat is younger than
// livenessHeartbeatThreshold.
func (p *ProbeState) HealthHandler(w http.ResponseWriter, _ *http.Request) {
	p.mu.RLock()
	fresh := !p.heartbeat.IsZero() && p.now().Sub(p.heartbeat) < livenessHeartbeatThreshold
	p.mu.RUnlock()
	writeProbe(w, fresh, "stale heartbeat")
}

// ReadyHandler serves /readyz: 200 when the last successful reconcile is within
// readinessIntervalMultiplier x interval.
func (p *ProbeState) ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	p.mu.RLock()
	window := time.Duration(readinessIntervalMultiplier) * p.interval
	ok := !p.lastSuccess.IsZero() && p.now().Sub(p.lastSuccess) < window
	p.mu.RUnlock()
	writeProbe(w, ok, "no recent successful reconcile")
}

func writeProbe(w http.ResponseWriter, ok bool, failMsg string) {
	if ok {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(failMsg))
}
