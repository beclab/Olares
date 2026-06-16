package server

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"k8s.io/klog/v2"
)

// healthTracker records Envoy's ACK/NACK of the xDS config that this control
// plane pushes. When Envoy rejects a listener update (for example a bad 443/444
// listener config), it replies with a discovery request whose ErrorDetail is
// set — a NACK. In that case the rejected listener is never installed, so the
// port simply does not listen even though both the control plane and Envoy
// processes are healthy. Tracking NACKs lets us surface that condition through
// the Kubernetes readiness probe.
type healthTracker struct {
	mu    sync.RWMutex
	nacks map[string]string // typeURL -> last error message reported by Envoy
}

func newHealthTracker() *healthTracker {
	return &healthTracker{nacks: make(map[string]string)}
}

// callbacks returns go-control-plane callbacks that update the tracker on every
// (delta and state-of-the-world) discovery request. Envoy sends a request with
// ErrorDetail set when it NACKs a config, and an empty-error request carrying a
// non-empty response nonce when it ACKs one.
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
		// A request with a nonce and no error is an ACK; clear any prior NACK.
		if _, ok := h.nacks[typeURL]; ok {
			klog.Infof("xds-server: Envoy ACK recovered for %s", shortType(typeURL))
			delete(h.nacks, typeURL)
		}
	}
}

// err returns a non-nil error if Envoy is currently rejecting any xDS config.
func (h *healthTracker) err() error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.nacks) == 0 {
		return nil
	}
	parts := make([]string, 0, len(h.nacks))
	for t, m := range h.nacks {
		parts = append(parts, fmt.Sprintf("%s (%s)", shortType(t), m))
	}
	sort.Strings(parts)
	return fmt.Errorf("envoy rejected xDS config: %s", strings.Join(parts, "; "))
}

// ReadyCheck is a controller-runtime readiness checker. It fails while Envoy is
// rejecting the pushed xDS config, e.g. when a duplicate/invalid listener leaves
// ports 443/444 not listening. Register it with mgr.AddReadyzCheck.
func (s *XdsServer) ReadyCheck(_ *http.Request) error {
	return s.health.err()
}

func shortType(typeURL string) string {
	if i := strings.LastIndex(typeURL, "."); i >= 0 {
		return typeURL[i+1:]
	}
	return typeURL
}
