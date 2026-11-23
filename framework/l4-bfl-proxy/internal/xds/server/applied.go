package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type resourceKey struct {
	kind string
	name string
}

type pendingResponse struct {
	nonce    string
	versions map[resourceKey]string
	removed  []resourceKey
}

type deltaStream struct {
	seen    map[string]bool
	pending map[string][]pendingResponse
	applied map[resourceKey]string
}

func criticalResource(key resourceKey) bool {
	switch key.kind {
	case resourcev3.ListenerType:
		return key.name == "https_443" || key.name == "https_pp_444"
	case resourcev3.RouteType:
		return strings.HasPrefix(key.name, "https_443_") || strings.HasPrefix(key.name, "https_444_")
	case resourcev3.ClusterType:
		return key.name == "app_gateway_edge_forward" || strings.HasPrefix(key.name, "authelia_backend_")
	case resourcev3.SecretType:
		return key.name == "l4-app-gateway-client" || key.name == "l4-app-gateway-ca"
	}
	return false
}

func (h *healthTracker) expect(snapshot *cachev3.Snapshot) {
	expected := make(map[resourceKey]string)
	for kind, versions := range snapshot.VersionMap {
		for name, version := range versions {
			key := resourceKey{kind, name}
			if criticalResource(key) {
				expected[key] = version
			}
		}
	}
	h.mu.Lock()
	h.expected = expected
	h.mu.Unlock()
}

func (h *healthTracker) openDeltaStream(id int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.streams[id] = &deltaStream{
		seen: make(map[string]bool), pending: make(map[string][]pendingResponse), applied: make(map[resourceKey]string),
	}
}

func (h *healthTracker) closeDeltaStream(id int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.streams, id)
}

func (h *healthTracker) sent(id int64, resp *discoverygrpc.DeltaDiscoveryResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	stream := h.streams[id]
	if stream == nil || resp.GetNonce() == "" {
		return
	}
	kind := resp.GetTypeUrl()
	pending := pendingResponse{nonce: resp.GetNonce(), versions: make(map[resourceKey]string)}
	for _, item := range resp.GetResources() {
		key := resourceKey{kind, item.GetName()}
		if criticalResource(key) {
			pending.versions[key] = item.GetVersion()
		}
	}
	for _, name := range resp.GetRemovedResources() {
		key := resourceKey{kind, name}
		if criticalResource(key) {
			pending.removed = append(pending.removed, key)
		}
	}
	const maxPendingResponses = 16
	queue := append(stream.pending[kind], pending)
	if len(queue) > maxPendingResponses {
		queue = queue[len(queue)-maxPendingResponses:]
	}
	stream.pending[kind] = queue
}

func (h *healthTracker) received(id int64, req *discoverygrpc.DeltaDiscoveryRequest) {
	h.mu.Lock()
	defer h.mu.Unlock()
	stream := h.streams[id]
	if stream == nil {
		return
	}
	kind := req.GetTypeUrl()
	if !stream.seen[kind] && req.GetResponseNonce() == "" && req.GetErrorDetail() == nil {
		// Envoy supplies its accepted versions again when an ADS stream reconnects.
		for name, version := range req.GetInitialResourceVersions() {
			key := resourceKey{kind, name}
			if criticalResource(key) {
				stream.applied[key] = version
			}
		}
	}
	stream.seen[kind] = true
	queue := stream.pending[kind]
	index := slices.IndexFunc(queue, func(p pendingResponse) bool { return p.nonce == req.GetResponseNonce() })
	if index < 0 {
		return
	}
	pending := queue[index]
	// A response acknowledged later in the ordered ADS stream supersedes
	// earlier outstanding responses for this type.
	stream.pending[kind] = queue[index+1:]
	for key, version := range pending.versions {
		if req.GetErrorDetail() != nil {
			delete(stream.applied, key)
		} else {
			stream.applied[key] = version
		}
	}
	for _, key := range pending.removed {
		delete(stream.applied, key)
	}
}

func (h *healthTracker) acceptedListenerVersions() (map[string]string, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.expected == nil {
		return nil, fmt.Errorf("xDS snapshot not published")
	}
	if len(h.expected) != 0 {
		for _, key := range []resourceKey{
			{resourcev3.ListenerType, "https_443"},
			{resourcev3.ListenerType, "https_pp_444"},
			{resourcev3.ClusterType, "app_gateway_edge_forward"},
			{resourcev3.SecretType, "l4-app-gateway-client"},
			{resourcev3.SecretType, "l4-app-gateway-ca"},
		} {
			if h.expected[key] == "" {
				return nil, fmt.Errorf("xDS snapshot is missing entry resource %s", key.name)
			}
		}
	}
	for _, stream := range h.streams {
		listeners := make(map[string]string)
		accepted := true
		for key, version := range h.expected {
			if stream.applied[key] != version {
				accepted = false
				break
			}
			if key.kind == resourcev3.ListenerType {
				listeners[key.name] = version
			}
		}
		if accepted {
			return listeners, nil
		}
	}
	return nil, fmt.Errorf("Envoy has not acknowledged the current entry xDS resources")
}

// activeListenerVersions excludes warming and draining listeners, even when an
// earlier version of the same listener is still bound to the port.
func activeListenerVersions(ctx context.Context, addr string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://"+addr+"/config_dump?resource=dynamic_listeners&mask=name,active_state.version_info", nil)
	if err != nil {
		return nil, err
	}
	resp, err := adminClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected admin status %d", resp.StatusCode)
	}
	var dump struct {
		Configs []struct {
			Name   string
			Active struct {
				Version string `json:"version_info"`
			} `json:"active_state"`
		}
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&dump); err != nil {
		return nil, err
	}
	versions := make(map[string]string)
	for _, item := range dump.Configs {
		versions[item.Name] = item.Active.Version
	}
	return versions, nil
}
