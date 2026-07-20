package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"k8s.io/klog/v2"
)

// envoyAdminAddr is the host:port of the Envoy admin interface queried by the
// readiness check. It is a var (not const) so tests can point it at a stub.
var envoyAdminAddr = "127.0.0.1:19000"

// adminClient is the HTTP client used to talk to the Envoy admin interface.
var adminClient = &http.Client{Timeout: 2 * time.Second}

// criticalListenerNames are the core ingress listeners whose absence means the
// proxy cannot serve user HTTPS traffic. Readiness only cares about these; a
// per-app TCP/UDP stream listener (e.g. a DNS app on host port 53 that fails to
// bind) must not mark the whole proxy NotReady.
var criticalListenerNames = []string{"https_443", "https_pp_444"}

// healthTracker records Envoy's ACK/NACK of the xDS config purely for
// diagnostics. The readiness decision is made by querying the Envoy admin
// /listeners endpoint (see XdsServer.ReadyCheck); this just surfaces *why* a
// listener failed (e.g. "cannot bind '0.0.0.0:53'") in the logs.
type healthTracker struct {
	mu    sync.RWMutex
	nacks map[string]string // typeURL -> last error message reported by Envoy
}

func newHealthTracker() *healthTracker {
	return &healthTracker{nacks: make(map[string]string)}
}

// callbacks returns go-control-plane callbacks that log Envoy's ACK/NACK on
// every (delta and state-of-the-world) discovery request. Envoy sets
// ErrorDetail on a NACK and sends an empty-error request with a non-empty nonce
// on an ACK.
func (h *healthTracker) callbacks() serverv3.CallbackFuncs {
	return serverv3.CallbackFuncs{
		StreamDeltaRequestFunc: func(_ int64, req *discoverygrpc.DeltaDiscoveryRequest) error {
			h.observe(req.GetTypeUrl(), req.GetResponseNonce(), req.GetErrorDetail().GetMessage())
			return nil
		},
		StreamRequestFunc: func(_ int64, req *discoverygrpc.DiscoveryRequest) error {
			h.observe(req.GetTypeUrl(), req.GetResponseNonce(), req.GetErrorDetail().GetMessage())
			return nil
		},
	}
}

func (h *healthTracker) observe(typeURL, nonce, errMsg string) {
	if typeURL == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	switch {
	case errMsg != "":
		if h.nacks[typeURL] != errMsg {
			klog.Errorf("xds-server: Envoy NACK for %s: %s", shortType(typeURL), errMsg)
		}
		h.nacks[typeURL] = errMsg
	case nonce != "":
		if _, ok := h.nacks[typeURL]; ok {
			klog.Infof("xds-server: Envoy ACK recovered for %s", shortType(typeURL))
			delete(h.nacks, typeURL)
		}
	}
}

// ReadyCheck is a controller-runtime readiness checker. It asks the Envoy admin
// interface which listeners are actually active and fails while any core ingress
// listener (443/444) is not listening — e.g. when a bad/duplicate listener
// config was rejected. App stream listeners are ignored.
func (s *XdsServer) ReadyCheck(req *http.Request) error {
	ctx := context.Background()
	if req != nil {
		ctx = req.Context()
	}
	active, err := fetchActiveListeners(ctx, envoyAdminAddr)
	if err != nil {
		return fmt.Errorf("query envoy admin /listeners: %w", err)
	}
	return evaluateReadiness(criticalListenerNames, active)
}

// evaluateReadiness returns an error naming any expected listener that Envoy is
// not currently reporting as active.
func evaluateReadiness(expected []string, active map[string]bool) error {
	var missing []string
	for _, name := range expected {
		if !active[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("envoy ingress listener(s) not active: %s", strings.Join(missing, ", "))
	}
	return nil
}

// fetchActiveListeners returns the set of listener names Envoy reports as active
// via GET /listeners. The default text format is one "<name>::<address>" per
// line.
func fetchActiveListeners(ctx context.Context, adminAddr string) (map[string]bool, error) {
	url := fmt.Sprintf("http://%s/listeners", adminAddr)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := adminClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	active := make(map[string]bool)
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := line
		if i := strings.Index(line, "::"); i >= 0 {
			name = line[:i]
		}
		active[name] = true
	}
	return active, nil
}

func shortType(typeURL string) string {
	if i := strings.LastIndex(typeURL, "."); i >= 0 {
		return typeURL[i+1:]
	}
	return typeURL
}
