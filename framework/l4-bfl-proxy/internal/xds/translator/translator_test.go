package translator

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	ppupstreamv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func asListener(t *testing.T, r interface{}) *listenerv3.Listener {
	t.Helper()
	l, ok := r.(*listenerv3.Listener)
	require.True(t, ok, "expected *listenerv3.Listener")
	return l
}

func asCluster(t *testing.T, r interface{}) *clusterv3.Cluster {
	t.Helper()
	c, ok := r.(*clusterv3.Cluster)
	require.True(t, ok, "expected *clusterv3.Cluster")
	return c
}

func asRouteConfig(t *testing.T, r interface{}) *routev3.RouteConfiguration {
	t.Helper()
	rc, ok := r.(*routev3.RouteConfiguration)
	require.True(t, ok, "expected *routev3.RouteConfiguration")
	return rc
}

func socketAddr(l *listenerv3.Listener) *corev3.SocketAddress {
	return l.GetAddress().GetSocketAddress()
}

// ---------------------------------------------------------------------------
// buildHTTPRedirectListener
// ---------------------------------------------------------------------------

func TestBuildHTTPRedirectListener(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "http_redirect_81",
		Address:  "0.0.0.0",
		Port:     81,
		Protocol: ir.ProtocolHTTP,
		HTTPRedirect: &ir.HTTPRedirectIR{
			Scheme: "https",
			Code:   301,
		},
	}

	l := buildHTTPRedirectListener(irListener)

	require.NotNil(t, l)
	assert.Equal(t, "http_redirect_81", l.Name)
	assert.Equal(t, "0.0.0.0", socketAddr(l).Address)
	assert.Equal(t, uint32(81), socketAddr(l).GetPortValue())
	require.Len(t, l.FilterChains, 1)
	require.Len(t, l.FilterChains[0].Filters, 1)
	assert.Equal(t, "envoy.filters.network.http_connection_manager", l.FilterChains[0].Filters[0].Name)
}

func TestBuildHTTPRedirectListener_NilRedirect(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "no_redirect",
		Protocol: ir.ProtocolHTTP,
	}

	assert.Nil(t, buildHTTPRedirectListener(irListener))
}

func TestBuildHTTPRedirectListener_StreamIdleTimeout(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "http_redirect_81",
		Address:  "0.0.0.0",
		Port:     81,
		Protocol: ir.ProtocolHTTP,
		HTTPRedirect: &ir.HTTPRedirectIR{
			Scheme: "https",
			Code:   301,
		},
	}

	l := buildHTTPRedirectListener(irListener)
	require.NotNil(t, l)

	require.Len(t, l.FilterChains, 1)
	require.Len(t, l.FilterChains[0].Filters, 1)
	filter := l.FilterChains[0].Filters[0]
	hcm := &hcmv3.HttpConnectionManager{}
	require.NoError(t, filter.GetTypedConfig().UnmarshalTo(hcm))

	require.NotNil(t, hcm.GetStreamIdleTimeout())
	assert.EqualValues(t, 1800, hcm.GetStreamIdleTimeout().GetSeconds())
}

// ---------------------------------------------------------------------------
// buildTLSListener
// ---------------------------------------------------------------------------

func TestBuildTLSListener_Basic(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:         "tls_443",
		Address:      "0.0.0.0",
		Port:         443,
		Protocol:     ir.ProtocolTLS,
		TLSInspector: true,
		Routes: []*ir.RouteIR{
			{
				Name:       "route_alice",
				SNIMatches: []string{"alice.example.com", "*.alice.example.com"},
				Destination: &ir.DestinationIR{
					Name: "user_alice",
					Host: "bfl.user-space-alice",
					Port: 444,
				},
				ProxyProtocolUpstream: true,
			},
		},
	}

	clusterSet := make(map[string]bool)
	l, clusters := buildTLSListener(irListener, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "tls_443", l.Name)
	assert.Equal(t, uint32(443), socketAddr(l).GetPortValue())

	require.Len(t, l.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.listener.tls_inspector", l.ListenerFilters[0].Name)

	require.Len(t, l.FilterChains, 1)
	fc := l.FilterChains[0]
	assert.Equal(t, "route_alice", fc.Name)
	assert.Equal(t, []string{"alice.example.com", "*.alice.example.com"}, fc.FilterChainMatch.ServerNames)

	require.Len(t, clusters, 1)
	c := asCluster(t, clusters[0])
	assert.Equal(t, "user_alice", c.Name)
	assert.True(t, clusterSet["user_alice"])
}

func TestBuildTLSListener_ProxyProtocol(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:          "tls_444",
		Address:       "0.0.0.0",
		Port:          444,
		Protocol:      ir.ProtocolTLS,
		ProxyProtocol: true,
		TLSInspector:  true,
		Routes: []*ir.RouteIR{
			{
				Name:       "route_alice",
				SNIMatches: []string{"alice.example.com"},
				Destination: &ir.DestinationIR{
					Name: "user_alice",
					Host: "bfl.user-space-alice",
					Port: 444,
				},
			},
		},
	}

	clusterSet := make(map[string]bool)
	l, _ := buildTLSListener(irListener, clusterSet)

	require.Len(t, l.ListenerFilters, 2)
	assert.Equal(t, "envoy.filters.listener.proxy_protocol", l.ListenerFilters[0].Name)
	assert.Equal(t, "envoy.filters.listener.tls_inspector", l.ListenerFilters[1].Name)
}

func TestBuildTLSListener_DenyAll_SourceCIDR(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:         "tls_443",
		Address:      "0.0.0.0",
		Port:         443,
		Protocol:     ir.ProtocolTLS,
		TLSInspector: true,
		Routes: []*ir.RouteIR{
			{
				Name:       "route_bob_allowed",
				SNIMatches: []string{"app.bob.example.com"},
				Destination: &ir.DestinationIR{
					Name: "user_bob",
					Host: "bfl.user-space-bob",
					Port: 444,
				},
			},
			{
				Name:               "route_bob_restricted",
				SNIMatches:         []string{"bob.example.com", "*.bob.example.com"},
				SourcePrefixRanges: []string{"100.64.0.0/24", "192.168.1.100/32"},
				Destination: &ir.DestinationIR{
					Name: "user_bob",
					Host: "bfl.user-space-bob",
					Port: 444,
				},
			},
		},
	}

	clusterSet := make(map[string]bool)
	l, clusters := buildTLSListener(irListener, clusterSet)

	require.Len(t, l.FilterChains, 2)

	fcAllowed := l.FilterChains[0]
	assert.Equal(t, []string{"app.bob.example.com"}, fcAllowed.FilterChainMatch.ServerNames)
	assert.Empty(t, fcAllowed.FilterChainMatch.SourcePrefixRanges)

	fcRestricted := l.FilterChains[1]
	assert.Equal(t, []string{"bob.example.com", "*.bob.example.com"}, fcRestricted.FilterChainMatch.ServerNames)
	require.Len(t, fcRestricted.FilterChainMatch.SourcePrefixRanges, 2)
	assert.Equal(t, "100.64.0.0", fcRestricted.FilterChainMatch.SourcePrefixRanges[0].AddressPrefix)
	assert.Equal(t, uint32(24), fcRestricted.FilterChainMatch.SourcePrefixRanges[0].PrefixLen.GetValue())
	assert.Equal(t, "192.168.1.100", fcRestricted.FilterChainMatch.SourcePrefixRanges[1].AddressPrefix)
	assert.Equal(t, uint32(32), fcRestricted.FilterChainMatch.SourcePrefixRanges[1].PrefixLen.GetValue())

	require.Len(t, clusters, 1)
}

func TestBuildTLSListener_ClusterDedup(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "tls_443",
		Address:  "0.0.0.0",
		Port:     443,
		Protocol: ir.ProtocolTLS,
		Routes: []*ir.RouteIR{
			{
				Name:       "route_a",
				SNIMatches: []string{"a.example.com"},
				Destination: &ir.DestinationIR{
					Name: "same_cluster",
					Host: "10.0.0.1",
					Port: 443,
				},
			},
			{
				Name:       "route_b",
				SNIMatches: []string{"b.example.com"},
				Destination: &ir.DestinationIR{
					Name: "same_cluster",
					Host: "10.0.0.1",
					Port: 443,
				},
			},
		},
	}

	clusterSet := make(map[string]bool)
	_, clusters := buildTLSListener(irListener, clusterSet)
	assert.Len(t, clusters, 1)
}

// ---------------------------------------------------------------------------
// buildL4TCPListener / buildL4UDPListener
// ---------------------------------------------------------------------------

func TestBuildL4TCPListener(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "stream_tcp_48126",
		Address:  "0.0.0.0",
		Port:     48126,
		Protocol: ir.ProtocolTCP,
		Routes: []*ir.RouteIR{
			{
				Name: "direct_48126",
				Destination: &ir.DestinationIR{
					Name: "direct_alice_48126",
					Host: "bfl.user-space-alice",
					Port: 48126,
				},
			},
		},
	}

	clusterSet := make(map[string]bool)
	l, clusters := buildL4TCPListener(irListener, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "stream_tcp_48126", l.Name)
	assert.Equal(t, uint32(48126), socketAddr(l).GetPortValue())
	require.Len(t, l.FilterChains, 1)
	assert.Equal(t, "direct_48126", l.FilterChains[0].Name)
	assert.Empty(t, l.ListenerFilters)

	require.Len(t, clusters, 1)
	c := asCluster(t, clusters[0])
	assert.Equal(t, "direct_alice_48126", c.Name)
	assert.Equal(t, clusterv3.Cluster_STRICT_DNS, c.GetType())
}

func TestBuildL4UDPListener(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "stream_udp_51820",
		Address:  "0.0.0.0",
		Port:     51820,
		Protocol: ir.ProtocolUDP,
		Routes: []*ir.RouteIR{
			{
				Name: "direct_51820",
				Destination: &ir.DestinationIR{
					Name: "direct_alice_51820",
					Host: "bfl.user-space-alice",
					Port: 51820,
				},
			},
		},
	}

	clusterSet := make(map[string]bool)
	l, clusters := buildL4UDPListener(irListener, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "stream_udp_51820", l.Name)
	sa := socketAddr(l)
	assert.Equal(t, uint32(51820), sa.GetPortValue())
	assert.Equal(t, corev3.SocketAddress_UDP, sa.Protocol)
	assert.NotNil(t, l.UdpListenerConfig)

	require.Len(t, l.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.udp_listener.udp_proxy", l.ListenerFilters[0].Name)

	require.Len(t, clusters, 1)
}

func TestBuildL4UDPListener_NoRoutes(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "empty_udp",
		Protocol: ir.ProtocolUDP,
	}
	l, clusters := buildL4UDPListener(irListener, make(map[string]bool))
	assert.Nil(t, l)
	assert.Nil(t, clusters)
}

// ---------------------------------------------------------------------------
// buildL4Cluster / buildL7Cluster
// ---------------------------------------------------------------------------

func TestBuildL4Cluster_Basic(t *testing.T) {
	dest := &ir.DestinationIR{
		Name: "user_alice",
		Host: "bfl.user-space-alice",
		Port: 444,
	}

	c := buildL4Cluster(dest, false)

	assert.Equal(t, "user_alice", c.Name)
	assert.Equal(t, clusterv3.Cluster_STRICT_DNS, c.GetType())
	assert.Nil(t, c.TransportSocket, "no transport socket when proxyProtocolUpstream=false")

	endpoints := c.LoadAssignment.Endpoints
	require.Len(t, endpoints, 1)
	require.Len(t, endpoints[0].LbEndpoints, 1)
	ep := endpoints[0].LbEndpoints[0].GetEndpoint()
	assert.Equal(t, "bfl.user-space-alice", ep.Address.GetSocketAddress().Address)
	assert.Equal(t, uint32(444), ep.Address.GetSocketAddress().GetPortValue())
}

func TestBuildL4Cluster_ProxyProtocolUpstream(t *testing.T) {
	dest := &ir.DestinationIR{
		Name: "user_alice",
		Host: "bfl.user-space-alice",
		Port: 444,
	}

	c := buildL4Cluster(dest, true)

	require.NotNil(t, c.TransportSocket)
	assert.Equal(t, "envoy.transport_sockets.upstream_proxy_protocol", c.TransportSocket.Name)

	ppUpstream := &ppupstreamv3.ProxyProtocolUpstreamTransport{}
	require.NoError(t, c.TransportSocket.GetTypedConfig().UnmarshalTo(ppUpstream))
	assert.Equal(t, corev3.ProxyProtocolConfig_V1, ppUpstream.Config.Version)
	assert.Equal(t, "envoy.transport_sockets.raw_buffer", ppUpstream.TransportSocket.Name)
}

func TestBuildL7Cluster(t *testing.T) {
	c := buildL7Cluster(&ir.ClusterIR{
		Name:   "app_alice_vault_main",
		Host:   "vault-svc.vault-alice.svc.cluster.local",
		Port:   8080,
		UseDNS: true,
	})

	assert.Equal(t, "app_alice_vault_main", c.Name)
	assert.Equal(t, clusterv3.Cluster_STRICT_DNS, c.GetType())

	ep := c.LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint()
	assert.Equal(t, "vault-svc.vault-alice.svc.cluster.local", ep.Address.GetSocketAddress().Address)
	assert.Equal(t, uint32(8080), ep.Address.GetSocketAddress().GetPortValue())
}

// ---------------------------------------------------------------------------
// buildMultiUserHTTPSListener
// ---------------------------------------------------------------------------

func TestBuildMultiUserHTTPSListener(t *testing.T) {
	clusterMap := map[string]*ir.ClusterIR{
		"app_alice_vault_main": {
			Name: "app_alice_vault_main", Host: "vault-svc.vault-alice.svc.cluster.local", Port: 8080, UseDNS: true,
		},
		"profile_service_alice": {
			Name: "profile_service_alice", Host: "profile-service.user-space-alice.svc.cluster.local", Port: 3000, UseDNS: true,
		},
		"authelia_backend_alice": {
			Name: "authelia_backend_alice", Host: "authelia-backend.user-system-alice.svc.cluster.local", Port: 9091, UseDNS: true,
		},
	}

	httpListeners := []*ir.HTTPListenerIR{
		{
			Name:          "https_443_alice",
			Address:       "0.0.0.0",
			Port:          443,
			TLS:           true,
			ProxyProtocol: false,
			SNIMatches:    []string{"alice.example.com", "*.alice.example.com"},
			TLSCert:       &ir.SecretIR{Name: "main-tls-alice", CertData: "cert", KeyData: "key"},
			UserName:      "alice",
			VirtualHosts: []*ir.VirtualHostIR{
				{
					Name:    "profile_alice",
					Domains: []string{"alice.example.com"},
					Routes: []*ir.HTTPRouteIR{{
						Name:       "profile_root_alice",
						PathPrefix: "/",
						Cluster:    "profile_service_alice",
					}},
				},
				{
					Name:    "app_alice_vault_main",
					Domains: []string{"vault.alice.example.com"},
					Routes: []*ir.HTTPRouteIR{{
						Name:             "default_alice_vault_main",
						PathPrefix:       "/",
						Cluster:          "app_alice_vault_main",
						WebSocketUpgrade: true,
					}},
				},
			},
			DefaultResponse: &ir.DirectResponseIR{Status: 421, Body: "not found", ContentType: "text/html"},
			PingRoute:       true,
		},
	}

	clusterSet := make(map[string]bool)
	listener, routeConfigs, clusters := buildMultiUserHTTPSListener(443, false, httpListeners, clusterMap, clusterSet)

	require.NotNil(t, listener)
	assert.Equal(t, "https_443", listener.Name)
	assert.Equal(t, uint32(443), socketAddr(listener).GetPortValue())

	// Should have tls_inspector listener filter
	require.Len(t, listener.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.listener.tls_inspector", listener.ListenerFilters[0].Name)

	// One filter chain for alice
	require.Len(t, listener.FilterChains, 1)
	fc := listener.FilterChains[0]
	assert.Equal(t, "https_443_alice", fc.Name)
	assert.Equal(t, []string{"alice.example.com", "*.alice.example.com"}, fc.FilterChainMatch.ServerNames)
	require.NotNil(t, fc.TransportSocket)

	// Route config
	require.Len(t, routeConfigs, 1)
	rc := asRouteConfig(t, routeConfigs[0])
	assert.Equal(t, "https_443_alice_routes", rc.Name)
	require.GreaterOrEqual(t, len(rc.VirtualHosts), 2)

	// Clusters created
	assert.True(t, len(clusters) >= 2)
}

// ---------------------------------------------------------------------------
// parseCIDR
// ---------------------------------------------------------------------------

func TestParseCIDR(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantPrefix string
		wantLen    uint32
		wantErr    bool
	}{
		{
			name:       "valid CIDR /24",
			input:      "100.64.0.0/24",
			wantPrefix: "100.64.0.0",
			wantLen:    24,
		},
		{
			name:       "valid CIDR /32",
			input:      "192.168.1.100/32",
			wantPrefix: "192.168.1.100",
			wantLen:    32,
		},
		{
			name:       "bare IP (no mask)",
			input:      "10.0.0.1",
			wantPrefix: "10.0.0.1",
			wantLen:    32,
		},
		{
			name:    "invalid",
			input:   "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCIDR(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantPrefix, got.AddressPrefix)
			assert.Equal(t, tt.wantLen, got.PrefixLen.GetValue())
		})
	}
}

// ---------------------------------------------------------------------------
// Translate (end-to-end)
// ---------------------------------------------------------------------------

func TestTranslate_AllProtocols(t *testing.T) {
	xdsIR := &ir.Xds{
		Listeners: []*ir.ListenerIR{
			{
				Name:     "http_redirect_81",
				Address:  "0.0.0.0",
				Port:     81,
				Protocol: ir.ProtocolHTTP,
				HTTPRedirect: &ir.HTTPRedirectIR{
					Scheme: "https",
					Code:   301,
				},
			},
			{
				Name:         "tls_443",
				Address:      "0.0.0.0",
				Port:         443,
				Protocol:     ir.ProtocolTLS,
				TLSInspector: true,
				Routes: []*ir.RouteIR{
					{
						Name:       "route_alice",
						SNIMatches: []string{"alice.example.com"},
						Destination: &ir.DestinationIR{
							Name: "user_alice",
							Host: "bfl.user-space-alice",
							Port: 444,
						},
						ProxyProtocolUpstream: true,
					},
				},
			},
			{
				Name:     "stream_tcp_30000",
				Address:  "0.0.0.0",
				Port:     30000,
				Protocol: ir.ProtocolTCP,
				Routes: []*ir.RouteIR{
					{
						Name: "direct_30000",
						Destination: &ir.DestinationIR{
							Name: "direct_alice_30000",
							Host: "bfl.user-space-alice",
							Port: 30000,
						},
					},
				},
			},
			{
				Name:     "stream_udp_51820",
				Address:  "0.0.0.0",
				Port:     51820,
				Protocol: ir.ProtocolUDP,
				Routes: []*ir.RouteIR{
					{
						Name: "direct_51820",
						Destination: &ir.DestinationIR{
							Name: "direct_alice_51820",
							Host: "bfl.user-space-alice",
							Port: 51820,
						},
					},
				},
			},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	// 4 listeners: HTTP redirect, TLS, TCP, UDP
	require.Len(t, snap.Listeners, 4)

	httpL := asListener(t, snap.Listeners[0])
	assert.Equal(t, "http_redirect_81", httpL.Name)

	tlsL := asListener(t, snap.Listeners[1])
	assert.Equal(t, "tls_443", tlsL.Name)
	require.Len(t, tlsL.FilterChains, 1)

	tcpL := asListener(t, snap.Listeners[2])
	assert.Equal(t, "stream_tcp_30000", tcpL.Name)

	udpL := asListener(t, snap.Listeners[3])
	assert.Equal(t, "stream_udp_51820", udpL.Name)
	assert.NotNil(t, udpL.UdpListenerConfig)

	// 3 distinct clusters: user_alice, direct_alice_30000, direct_alice_51820
	require.Len(t, snap.Clusters, 3)
}

func TestTranslate_HTTPSListeners(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{
			{
				Name:       "https_443_alice",
				Address:    "0.0.0.0",
				Port:       443,
				TLS:        true,
				SNIMatches: []string{"alice.example.com"},
				TLSCert:    &ir.SecretIR{Name: "main-tls-alice", CertData: "cert", KeyData: "key"},
				UserName:   "alice",
				VirtualHosts: []*ir.VirtualHostIR{
					{
						Name:    "profile_alice",
						Domains: []string{"alice.example.com"},
						Routes: []*ir.HTTPRouteIR{{
							Name:       "profile_root",
							PathPrefix: "/",
							Cluster:    "profile_service_alice",
						}},
					},
				},
				DefaultResponse: &ir.DirectResponseIR{Status: 421, Body: "err"},
				PingRoute:       true,
			},
		},
		Clusters: []*ir.ClusterIR{
			{Name: "profile_service_alice", Host: "profile-service.user-space-alice.svc.cluster.local", Port: 3000, UseDNS: true},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	// 1 HTTPS listener
	require.Len(t, snap.Listeners, 1)
	l := asListener(t, snap.Listeners[0])
	assert.Equal(t, "https_443", l.Name)
	require.Len(t, l.FilterChains, 1)

	// 1 route config
	require.Len(t, snap.Routes, 1)
	rc := asRouteConfig(t, snap.Routes[0])
	assert.Equal(t, "https_443_alice_routes", rc.Name)

	// 1 cluster
	require.Len(t, snap.Clusters, 1)
}

func TestTranslate_EmptyIR(t *testing.T) {
	xt := &XdsTranslator{}
	snap := xt.Translate(&ir.Xds{})

	assert.Empty(t, snap.Listeners)
	assert.Empty(t, snap.Clusters)
}

// ---------------------------------------------------------------------------
// suppress unused imports
// ---------------------------------------------------------------------------

var _ = []cachetypes.Resource(nil)
