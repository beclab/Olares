package translator

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	hm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/header_mutation/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

// Export generated edge configuration for app-service's real Envoy fixture.
// The normal suite always checks the edge identity and client-IP contract.
func TestEntryEdgeIdentityContract(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name: "desktop", PathPrefix: "/", Cluster: "fixture_agw",
		RequestHeaders: map[string]string{"X-Original-URL": "%REQ(:SCHEME)%://%REQ(:AUTHORITY)%%REQ(:PATH)%", "X-Original-Method": "%REQ(:METHOD)%"},
		ExtAuth:        &ir.ExtAuthConfigIR{Cluster: "authelia_backend_alice", PathPrefix: "/api/verify/"},
	}
	input := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{{Name: "entry_fixture_edge", Address: "0.0.0.0", Port: 10003, UserName: "alice", VirtualHosts: []*ir.VirtualHostIR{{Name: "desktop", Domains: []string{"desktop.alice.example.org"}, Routes: []*ir.HTTPRouteIR{routeIR}}}}},
		Clusters:      []*ir.ClusterIR{{Name: "fixture_agw", Host: "127.0.0.1", Port: 10000}, {Name: "authelia_backend_alice", Host: "127.0.0.1", Port: 9091}},
	}
	snap := (&XdsTranslator{}).Translate(input)
	listener := asListener(t, snap.Listeners[0])
	h := hcmFromListener(t, listener)
	strip := &hm.HeaderMutation{}
	require.NoError(t, h.HttpFilters[0].GetTypedConfig().UnmarshalTo(strip))
	route := asRouteConfig(t, snap.Routes[0])
	for _, name := range []string{"remote-user", "remote-groups", "x-bfl-user", "x-caller-user", "forwarded", "x-forwarded-for", "x-real-ip", "x-original-forwarded-for", "x-envoy-external-address", "x-original-url"} {
		require.Contains(t, headerMutationRemoves(t, strip.GetMutations().GetRequestMutations()), name)
	}
	for _, name := range []string{"X-Forwarded-For", "X-Real-IP"} {
		found := false
		for _, v := range route.RequestHeadersToAdd {
			if v.Header.Key == name {
				found = true
				require.Equal(t, "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%", v.Header.Value)
			}
		}
		require.True(t, found)
		found = false
		for _, v := range autheliaAuthorizationRequest("alice").HeadersToAdd {
			if v.Key == name {
				found = true
				require.Equal(t, "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%", v.Value)
			}
		}
		require.True(t, found)
	}
	path := os.Getenv("ENTRY_EDGE_FIXTURE")
	if path == "" {
		return
	}
	h.RouteSpecifier = &hcmv3.HttpConnectionManager_RouteConfig{RouteConfig: route}
	listener.FilterChains[0].Filters[0].GetTypedConfig().Value = mustAny(h).Value
	encodedListener, err := protojson.Marshal(listener)
	require.NoError(t, err)
	clusters := []json.RawMessage{}
	for _, c := range snap.Clusters {
		encoded, err := protojson.Marshal(asCluster(t, c))
		require.NoError(t, err)
		clusters = append(clusters, encoded)
	}
	out, err := json.Marshal(map[string]any{"listener": json.RawMessage(encodedListener), "clusters": clusters})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, out, 0600))
}
