package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"k8s.io/klog/v2"
)

// envoyAdminAddr is the local Envoy admin endpoint queried for readiness.
const envoyAdminAddr = "127.0.0.1:19000"

// adminClient is the HTTP client used to talk to the Envoy admin interface.
var adminClient = &http.Client{Timeout: 2 * time.Second}

// healthTracker records accepted delta resources and keeps NACK details for diagnostics.
type healthTracker struct {
	mu       sync.RWMutex
	nacks    map[string]string // typeURL -> last error message reported by Envoy
	expected map[resourceKey]string
	streams  map[int64]*deltaStream
}

func newHealthTracker() *healthTracker {
	return &healthTracker{nacks: make(map[string]string), streams: make(map[int64]*deltaStream)}
}

// callbacks returns go-control-plane callbacks that log Envoy's ACK/NACK on
// every (delta and state-of-the-world) discovery request. Envoy sets
// ErrorDetail on a NACK and sends an empty-error request with a non-empty nonce
// on an ACK.
func (h *healthTracker) callbacks() serverv3.CallbackFuncs {
	return serverv3.CallbackFuncs{
		DeltaStreamOpenFunc: func(_ context.Context, id int64, _ string) error {
			h.openDeltaStream(id)
			return nil
		},
		DeltaStreamClosedFunc: func(id int64, _ *corev3.Node) { h.closeDeltaStream(id) },
		StreamDeltaResponseFunc: func(id int64, _ *discoverygrpc.DeltaDiscoveryRequest, resp *discoverygrpc.DeltaDiscoveryResponse) {
			h.sent(id, resp)
		},
		StreamDeltaRequestFunc: func(id int64, req *discoverygrpc.DeltaDiscoveryRequest) error {
			h.received(id, req)
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

// ReadyCheck requires the current critical xDS resources to be acknowledged and
// the exact corresponding HTTPS listener versions to be active in Envoy.
func (s *XdsServer) ReadyCheck(req *http.Request) error {
	ctx := context.Background()
	if req != nil {
		ctx = req.Context()
	}
	expected, err := s.health.acceptedListenerVersions()
	if err != nil {
		return err
	}
	active, err := activeListenerVersions(ctx, envoyAdminAddr)
	if err != nil {
		return fmt.Errorf("query Envoy active listener versions: %w", err)
	}
	for name, version := range expected {
		if active[name] != version {
			return fmt.Errorf("Envoy listener %s is not active at the current xDS version", name)
		}
	}
	return nil
}

func shortType(typeURL string) string {
	if i := strings.LastIndex(typeURL, "."); i >= 0 {
		return typeURL[i+1:]
	}
	return typeURL
}
