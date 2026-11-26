package translator

import (
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
)

func withOptions(t *testing.T, o Options) {
	t.Helper()
	prev := CurrentOptions()
	SetOptions(o)
	t.Cleanup(func() { SetOptions(prev) })
}

func requireKeepaliveSocketOptions(t *testing.T, l *listenerv3.Listener, idle, intvl, cnt int64) {
	t.Helper()
	got := map[string]*corev3.SocketOption{}
	for _, o := range l.GetSocketOptions() {
		got[o.GetDescription()] = o
	}
	require.Len(t, got, 4, "listener %s socket options", l.GetName())

	check := func(desc string, level, name, value int64) {
		o := got[desc]
		require.NotNil(t, o, desc)
		assert.Equal(t, level, o.GetLevel(), desc)
		assert.Equal(t, name, o.GetName(), desc)
		assert.Equal(t, value, o.GetIntValue(), desc)
		assert.Equal(t, corev3.SocketOption_STATE_PREBIND, o.GetState(), desc)
	}
	check("SO_KEEPALIVE", 1, 9, 1)
	check("TCP_KEEPIDLE", 6, 4, idle)
	check("TCP_KEEPINTVL", 6, 5, intvl)
	check("TCP_KEEPCNT", 6, 6, cnt)
}

func requireClusterKeepalive(t *testing.T, c *clusterv3.Cluster, timeSec, intvlSec, probes uint32) {
	t.Helper()
	ka := c.GetUpstreamConnectionOptions().GetTcpKeepalive()
	require.NotNil(t, ka, "cluster %s tcp_keepalive", c.GetName())
	assert.Equal(t, timeSec, ka.GetKeepaliveTime().GetValue())
	assert.Equal(t, intvlSec, ka.GetKeepaliveInterval().GetValue())
	assert.Equal(t, probes, ka.GetKeepaliveProbes().GetValue())
}

func TestDefaultOptions(t *testing.T) {
	o := DefaultOptions()
	assert.Zero(t, o.RouteTimeout, "absolute route timeout must be disabled for slow LLM upstreams")
	assert.Zero(t, o.RouteIdleTimeout, "routes inherit the HCM stream idle timeout by default")
	assert.Equal(t, 30*time.Minute, o.HTTPStreamIdleTimeout)
	assert.Equal(t, 10*time.Second, o.ClusterIdleTimeout)
	assert.Equal(t, 60*time.Second, o.TCPKeepaliveTime)
	assert.Equal(t, 15*time.Second, o.TCPKeepaliveInterval)
	assert.EqualValues(t, 4, o.TCPKeepaliveProbes)
	assert.EqualValues(t, 1<<20, o.HTTPSBufferLimitBytes)
}

func TestTranslateRoute_NoAbsoluteTimeoutInheritsStreamIdle(t *testing.T) {
	withOptions(t, DefaultOptions())

	route := translateRoute(&ir.HTTPRouteIR{Name: "llm", PathPrefix: "/", Cluster: "llm"}, nil)
	ra := route.GetRoute()
	require.NotNil(t, ra)
	require.NotNil(t, ra.GetTimeout(), "timeout must be explicitly 0, otherwise Envoy defaults to 15s")
	assert.Zero(t, ra.GetTimeout().AsDuration())
	assert.Nil(t, ra.GetIdleTimeout(), "an explicit 0 would disable the idle timeout; omit it to inherit stream_idle_timeout")
}

func TestTranslateRoute_CustomTimeouts(t *testing.T) {
	o := DefaultOptions()
	o.RouteTimeout = 2 * time.Minute
	o.RouteIdleTimeout = 45 * time.Second
	withOptions(t, o)

	ra := translateRoute(&ir.HTTPRouteIR{Name: "app", Cluster: "app"}, nil).GetRoute()
	assert.Equal(t, 2*time.Minute, ra.GetTimeout().AsDuration())
	assert.Equal(t, 45*time.Second, ra.GetIdleTimeout().AsDuration())
}

func TestTranslateRoute_DirectResponseHasNoTimeouts(t *testing.T) {
	route := translateRoute(&ir.HTTPRouteIR{Name: "d", DirectResponse: &ir.DirectResponseIR{Status: 403}}, nil)
	assert.Nil(t, route.GetRoute())
}

func TestKeepalive_OnAllTCPListeners(t *testing.T) {
	o := DefaultOptions()
	o.TCPKeepaliveTime = 30 * time.Second
	o.TCPKeepaliveInterval = 5 * time.Second
	o.TCPKeepaliveProbes = 3
	withOptions(t, o)

	xds := &ir.Xds{
		Listeners: []*ir.ListenerIR{
			{
				Name: "tls_443", Address: "0.0.0.0", Port: 443, Protocol: ir.ProtocolTLS, TLSInspector: true,
				Routes: []*ir.RouteIR{{Name: "r1", SNIMatches: []string{"a.example.com"}, Destination: &ir.DestinationIR{Name: "l4_a", Host: "a.svc", Port: 443}}},
			},
			{
				Name: "tcp_9000", Address: "0.0.0.0", Port: 9000, Protocol: ir.ProtocolTCP,
				Routes: []*ir.RouteIR{{Name: "r2", Destination: &ir.DestinationIR{Name: "l4_b", Host: "b.svc", Port: 9000}}},
			},
			{
				Name: "udp_9001", Address: "0.0.0.0", Port: 9001, Protocol: ir.ProtocolUDP,
				Routes: []*ir.RouteIR{{Name: "r3", Destination: &ir.DestinationIR{Name: "l4_c", Host: "c.svc", Port: 9001}}},
			},
		},
		HTTPListeners: []*ir.HTTPListenerIR{{
			Name: "user1", Port: 8443, SNIMatches: []string{"user1.example.com"},
			VirtualHosts: []*ir.VirtualHostIR{{
				Name: "app", Domains: []string{"app.user1.example.com"},
				Routes: []*ir.HTTPRouteIR{{Name: "app", PathPrefix: "/", Cluster: "app_backend"}},
			}},
		}},
		StreamListeners: []*ir.StreamListenerIR{
			{Name: "stream_tcp", Address: "0.0.0.0", Port: 7000, Protocol: "tcp", Cluster: "stream_backend"},
			{Name: "stream_udp", Address: "0.0.0.0", Port: 7001, Protocol: "udp", Cluster: "stream_udp_backend"},
		},
		Clusters: []*ir.ClusterIR{
			{Name: "app_backend", Host: "app.svc", Port: 80, UseDNS: true},
			{Name: "stream_backend", Host: "stream.svc", Port: 7000, UseDNS: true},
			{Name: "stream_udp_backend", Host: "stream.svc", Port: 7001, UseDNS: true},
		},
	}

	snap := New(nil, nil).Translate(xds)

	listeners := map[string]*listenerv3.Listener{}
	for _, r := range snap.Listeners {
		l := asListener(t, r)
		listeners[l.GetName()] = l
	}
	for _, name := range []string{"tls_443", "tcp_9000", "https_8443", "stream_tcp"} {
		require.Contains(t, listeners, name)
		requireKeepaliveSocketOptions(t, listeners[name], 30, 5, 3)
	}
	for _, name := range []string{"udp_9001", "stream_udp"} {
		require.Contains(t, listeners, name)
		assert.Empty(t, listeners[name].GetSocketOptions(), "UDP listener %s must not get TCP socket options", name)
	}

	require.NotEmpty(t, snap.Clusters)
	for _, r := range snap.Clusters {
		requireClusterKeepalive(t, asCluster(t, r), 30, 5, 3)
	}
}

func TestBuildL7Cluster_PoolIdleTimeoutUnchanged(t *testing.T) {
	withOptions(t, DefaultOptions())

	c := buildL7Cluster(&ir.ClusterIR{Name: "app", Host: "app.svc", Port: 80, UseDNS: true})
	assert.Equal(t, 10*time.Second, c.GetCommonHttpProtocolOptions().GetIdleTimeout().AsDuration(),
		"pool idle timeout only applies to connections without active streams and must stay short")
	requireClusterKeepalive(t, c, 60, 15, 4)
}

func TestBuildL4Cluster_Keepalive(t *testing.T) {
	withOptions(t, DefaultOptions())

	c := buildL4Cluster(&ir.DestinationIR{Name: "l4", Host: "x.svc", Port: 443}, true)
	requireClusterKeepalive(t, c, 60, 15, 4)
}

func TestBuildMultiUserHTTPSListener_BufferLimit(t *testing.T) {
	httpListeners := []*ir.HTTPListenerIR{{Name: "u", Port: 443, SNIMatches: []string{"u.example.com"}}}

	withOptions(t, DefaultOptions())
	l, _, _ := buildMultiUserHTTPSListener(443, false, httpListeners, map[string]*ir.ClusterIR{}, map[string]bool{})
	assert.EqualValues(t, 1<<20, l.GetPerConnectionBufferLimitBytes().GetValue())

	o := DefaultOptions()
	o.HTTPSBufferLimitBytes = 65536
	withOptions(t, o)
	l, _, _ = buildMultiUserHTTPSListener(443, false, httpListeners, map[string]*ir.ClusterIR{}, map[string]bool{})
	assert.EqualValues(t, 65536, l.GetPerConnectionBufferLimitBytes().GetValue())
}
