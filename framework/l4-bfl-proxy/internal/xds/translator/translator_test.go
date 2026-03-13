package translator

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	ppupstreamv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
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

	// TLS inspector listener filter should be present.
	require.Len(t, l.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.listener.tls_inspector", l.ListenerFilters[0].Name)

	require.Len(t, l.FilterChains, 1)
	fc := l.FilterChains[0]
	assert.Equal(t, "route_alice", fc.Name)
	assert.Equal(t, []string{"alice.example.com", "*.alice.example.com"}, fc.FilterChainMatch.ServerNames)

	// Cluster should be created.
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

	// Both proxy_protocol and tls_inspector listener filters.
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

	// First chain: allowed domains, no source restriction.
	fcAllowed := l.FilterChains[0]
	assert.Equal(t, []string{"app.bob.example.com"}, fcAllowed.FilterChainMatch.ServerNames)
	assert.Empty(t, fcAllowed.FilterChainMatch.SourcePrefixRanges)

	// Second chain: all domains, restricted by source CIDR.
	fcRestricted := l.FilterChains[1]
	assert.Equal(t, []string{"bob.example.com", "*.bob.example.com"}, fcRestricted.FilterChainMatch.ServerNames)
	require.Len(t, fcRestricted.FilterChainMatch.SourcePrefixRanges, 2)
	assert.Equal(t, "100.64.0.0", fcRestricted.FilterChainMatch.SourcePrefixRanges[0].AddressPrefix)
	assert.Equal(t, uint32(24), fcRestricted.FilterChainMatch.SourcePrefixRanges[0].PrefixLen.GetValue())
	assert.Equal(t, "192.168.1.100", fcRestricted.FilterChainMatch.SourcePrefixRanges[1].AddressPrefix)
	assert.Equal(t, uint32(32), fcRestricted.FilterChainMatch.SourcePrefixRanges[1].PrefixLen.GetValue())

	// Only one cluster (deduped).
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
// buildTCPListener
// ---------------------------------------------------------------------------

func TestBuildTCPListener(t *testing.T) {
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
	l, clusters := buildTCPListener(irListener, clusterSet)

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

// ---------------------------------------------------------------------------
// buildUDPListener
// ---------------------------------------------------------------------------

func TestBuildUDPListener(t *testing.T) {
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
	l, clusters := buildUDPListener(irListener, clusterSet)

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

func TestBuildUDPListener_NoRoutes(t *testing.T) {
	irListener := &ir.ListenerIR{
		Name:     "empty_udp",
		Protocol: ir.ProtocolUDP,
	}
	l, clusters := buildUDPListener(irListener, make(map[string]bool))
	assert.Nil(t, l)
	assert.Nil(t, clusters)
}

// ---------------------------------------------------------------------------
// buildCluster
// ---------------------------------------------------------------------------

func TestBuildCluster_Basic(t *testing.T) {
	dest := &ir.DestinationIR{
		Name: "user_alice",
		Host: "bfl.user-space-alice",
		Port: 444,
	}

	c := buildCluster(dest, false)

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

func TestBuildCluster_ProxyProtocolUpstream(t *testing.T) {
	dest := &ir.DestinationIR{
		Name: "user_alice",
		Host: "bfl.user-space-alice",
		Port: 444,
	}

	c := buildCluster(dest, true)

	require.NotNil(t, c.TransportSocket)
	assert.Equal(t, "envoy.transport_sockets.upstream_proxy_protocol", c.TransportSocket.Name)

	ppUpstream := &ppupstreamv3.ProxyProtocolUpstreamTransport{}
	require.NoError(t, c.TransportSocket.GetTypedConfig().UnmarshalTo(ppUpstream))
	assert.Equal(t, corev3.ProxyProtocolConfig_V1, ppUpstream.Config.Version)
	assert.Equal(t, "envoy.transport_sockets.raw_buffer", ppUpstream.TransportSocket.Name)
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
// translate (end-to-end)
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
	listeners, clusters := xt.Translate(xdsIR)

	// 4 listeners: HTTP redirect, TLS, TCP, UDP
	require.Len(t, listeners, 4)

	httpL := asListener(t, listeners[0])
	assert.Equal(t, "http_redirect_81", httpL.Name)

	tlsL := asListener(t, listeners[1])
	assert.Equal(t, "tls_443", tlsL.Name)
	require.Len(t, tlsL.FilterChains, 1)

	tcpL := asListener(t, listeners[2])
	assert.Equal(t, "stream_tcp_30000", tcpL.Name)

	udpL := asListener(t, listeners[3])
	assert.Equal(t, "stream_udp_51820", udpL.Name)
	assert.NotNil(t, udpL.UdpListenerConfig)

	// 3 distinct clusters: user_alice, direct_alice_30000, direct_alice_51820
	require.Len(t, clusters, 3)
}

func TestTranslate_EmptyIR(t *testing.T) {
	xt := &XdsTranslator{}
	listeners, clusters := xt.Translate(&ir.Xds{})

	assert.Empty(t, listeners)
	assert.Empty(t, clusters)
}
