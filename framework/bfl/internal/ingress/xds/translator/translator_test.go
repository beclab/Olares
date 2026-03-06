package translator

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bytetrade.io/web3os/bfl/internal/ingress/ir"
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
	irListener := &ir.HTTPListenerIR{
		Name:       "http_redirect_80",
		Address:    "0.0.0.0",
		Port:       80,
		IsRedirect: true,
	}

	l := buildHTTPRedirectListener(irListener)

	require.NotNil(t, l)
	assert.Equal(t, "http_redirect_80", l.Name)
	assert.Equal(t, "0.0.0.0", socketAddr(l).Address)
	assert.Equal(t, uint32(80), socketAddr(l).GetPortValue())
	require.Len(t, l.FilterChains, 1)
	require.Len(t, l.FilterChains[0].Filters, 1)
	assert.Equal(t, "envoy.filters.network.http_connection_manager", l.FilterChains[0].Filters[0].Name)
}

// ---------------------------------------------------------------------------
// buildHTTPSListener
// ---------------------------------------------------------------------------

func TestBuildHTTPSListener_Basic(t *testing.T) {
	irListener := &ir.HTTPListenerIR{
		Name:    "https_443",
		Address: "0.0.0.0",
		Port:    443,
		TLS:     true,
		VirtualHosts: []*ir.VirtualHostIR{{
			Name:    "app_vault",
			Domains: []string{"vault.alice.snowinning.com"},
			Routes: []*ir.HTTPRouteIR{{
				Name:       "default_vault",
				PathPrefix: "/",
				Cluster:    "app_vault_main",
			}},
		}},
		DefaultResponse: &ir.DirectResponseIR{
			Status: 421, Body: "not found",
		},
	}

	irClusters := []*ir.ClusterIR{
		{Name: "app_vault_main", Host: "vault-svc.ns.svc.cluster.local", Port: 80, UseDNS: true},
		{Name: "authelia_backend", Host: "authelia-backend.os-framework.svc.cluster.local", Port: 9091, UseDNS: true},
	}
	secretMap := map[string]*ir.SecretIR{
		"main-tls": {Name: "main-tls", CertData: "CERT", KeyData: "KEY"},
	}
	clusterSet := make(map[string]bool)

	l, rc, clusters := buildHTTPSListener(irListener, irClusters, secretMap, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "https_443", l.Name)
	assert.Equal(t, uint32(443), socketAddr(l).GetPortValue())

	// Should have TLS transport socket
	require.Len(t, l.FilterChains, 1)
	require.NotNil(t, l.FilterChains[0].TransportSocket)

	// HCM should use RDS
	hcmFilter := l.FilterChains[0].Filters[0]
	hcm := &hcmv3.HttpConnectionManager{}
	require.NoError(t, hcmFilter.GetTypedConfig().UnmarshalTo(hcm))
	assert.NotNil(t, hcm.GetRds(), "HCM should use RDS")
	assert.Equal(t, "https_443_routes", hcm.GetRds().GetRouteConfigName())
	assert.Equal(t, httpStreamIdleTimeout, hcm.GetStreamIdleTimeout().AsDuration())

	// RouteConfiguration returned separately
	require.NotNil(t, rc)
	assert.Equal(t, "https_443_routes", rc.Name)

	assert.True(t, len(clusters) >= 2)
	assert.True(t, clusterSet["app_vault_main"])
	assert.True(t, clusterSet["authelia_backend"])
}

func TestBuildHTTPSListener_ProxyProtocol(t *testing.T) {
	irListener := &ir.HTTPListenerIR{
		Name:          "https_444",
		Address:       "0.0.0.0",
		Port:          444,
		TLS:           true,
		ProxyProtocol: true,
		VirtualHosts:  []*ir.VirtualHostIR{},
	}

	secretMap := map[string]*ir.SecretIR{
		"main-tls": {Name: "main-tls", CertData: "CERT", KeyData: "KEY"},
	}
	clusterSet := make(map[string]bool)

	l, _, _ := buildHTTPSListener(irListener, nil, secretMap, clusterSet)

	// Should have proxy_protocol listener filter
	require.Len(t, l.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.listener.proxy_protocol", l.ListenerFilters[0].Name)
}

// ---------------------------------------------------------------------------
// translateVirtualHost
// ---------------------------------------------------------------------------

func TestTranslateVirtualHost(t *testing.T) {
	vhIR := &ir.VirtualHostIR{
		Name:    "app_vault",
		Domains: []string{"vault.alice.snowinning.com"},
		Routes: []*ir.HTTPRouteIR{{
			Name:       "default_vault",
			PathPrefix: "/",
			Cluster:    "app_vault_main",
		}},
	}

	vh := translateVirtualHost(vhIR)

	assert.Equal(t, "app_vault", vh.Name)
	assert.Equal(t, []string{"vault.alice.snowinning.com"}, vh.Domains)
	require.NotNil(t, vh.TypedPerFilterConfig)
	assert.Contains(t, vh.TypedPerFilterConfig, "envoy.filters.http.ext_authz")
	assert.NotContains(t, vh.TypedPerFilterConfig, "envoy.filters.http.lua")
}

// ---------------------------------------------------------------------------
// translateRoute
// ---------------------------------------------------------------------------

func TestTranslateRoute_PrefixMatch(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name:             "default_vault",
		PathPrefix:       "/api",
		Cluster:          "app_vault_main",
		WebSocketUpgrade: true,
		RequestHeaders:   map[string]string{"X-BFL-USER": "alice"},
	}

	route := translateRoute(routeIR)

	assert.Equal(t, "default_vault", route.Name)
	assert.Equal(t, "/api", route.Match.GetPrefix())
	assert.Equal(t, "app_vault_main", route.GetRoute().GetCluster())
	assert.Equal(t, routeTimeout, route.GetRoute().GetTimeout().AsDuration())
	assert.True(t, len(route.RequestHeadersToAdd) > 0)

	// WebSocket upgrade configs
	assert.True(t, len(route.GetRoute().UpgradeConfigs) > 0)
}

func TestTranslateRoute_ExactMatch(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name:      "exact_route",
		PathExact: "/health",
		Cluster:   "backend",
	}

	route := translateRoute(routeIR)
	assert.Equal(t, "/health", route.Match.GetPath())
}

func TestTranslateRoute_RegexMatch(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name:      "regex_route",
		PathRegex: "/api/v[0-9]+/.*",
		Cluster:   "backend",
	}

	route := translateRoute(routeIR)
	assert.Equal(t, "/api/v[0-9]+/.*", route.Match.GetSafeRegex().GetRegex())
}

func TestTranslateRoute_DirectResponse(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name: "default_response",
		DirectResponse: &ir.DirectResponseIR{
			Status: 421,
			Body:   "not found",
		},
	}

	route := translateRoute(routeIR)
	assert.Equal(t, uint32(421), route.GetDirectResponse().GetStatus())
	assert.Equal(t, "not found", route.GetDirectResponse().GetBody().GetInlineString())
}

func TestTranslateRoute_ExtAuth(t *testing.T) {
	routeIR := &ir.HTTPRouteIR{
		Name:       "files_node",
		PathPrefix: "/api/resources/cache/node1/",
		Cluster:    "files_node1",
		ExtAuth: &ir.ExtAuthConfigIR{
			Cluster:    "authelia_backend",
			PathPrefix: "/api/authz/auth-request",
		},
	}

	route := translateRoute(routeIR)
	assert.NotNil(t, route.TypedPerFilterConfig)
	assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.ext_authz")
}

// ---------------------------------------------------------------------------
// buildCluster
// ---------------------------------------------------------------------------

func TestBuildCluster_DNS(t *testing.T) {
	c := buildCluster(&ir.ClusterIR{
		Name:   "app_vault_main",
		Host:   "vault-svc.ns.svc.cluster.local",
		Port:   80,
		UseDNS: true,
	})

	assert.Equal(t, "app_vault_main", c.Name)
	assert.Equal(t, clusterv3.Cluster_STRICT_DNS, c.GetType())
	assert.Equal(t, connectTimeout, c.GetConnectTimeout().AsDuration())

	ep := c.LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint()
	assert.Equal(t, "vault-svc.ns.svc.cluster.local", ep.Address.GetSocketAddress().Address)
	assert.Equal(t, uint32(80), ep.Address.GetSocketAddress().GetPortValue())
}

func TestBuildCluster_Static(t *testing.T) {
	c := buildCluster(&ir.ClusterIR{
		Name:   "static_backend",
		Host:   "10.0.0.1",
		Port:   8080,
		UseDNS: false,
	})

	assert.Equal(t, clusterv3.Cluster_STATIC, c.GetType())
}

// ---------------------------------------------------------------------------
// buildTCPListener
// ---------------------------------------------------------------------------

func TestBuildTCPListener(t *testing.T) {
	irListener := &ir.StreamListenerIR{
		Name:     "stream_tcp_30000",
		Address:  "0.0.0.0",
		Port:     30000,
		Protocol: "tcp",
		Cluster:  "stream_tcp_30000",
	}
	irClusters := []*ir.ClusterIR{
		{Name: "stream_tcp_30000", Host: "svc.ns.svc.cluster.local", Port: 30000, UseDNS: true},
	}
	clusterSet := make(map[string]bool)

	l, clusters := buildTCPListener(irListener, irClusters, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "stream_tcp_30000", l.Name)
	assert.Equal(t, uint32(30000), socketAddr(l).GetPortValue())
	require.Len(t, l.FilterChains, 1)
	assert.Equal(t, "envoy.filters.network.tcp_proxy", l.FilterChains[0].Filters[0].Name)
	tcpProxy := &tcpproxyv3.TcpProxy{}
	require.NoError(t, l.FilterChains[0].Filters[0].GetTypedConfig().UnmarshalTo(tcpProxy))
	assert.Equal(t, tcpIdleTimeout, tcpProxy.GetIdleTimeout().AsDuration())

	require.Len(t, clusters, 1)
	assert.True(t, clusterSet["stream_tcp_30000"])
}

// ---------------------------------------------------------------------------
// buildUDPListener
// ---------------------------------------------------------------------------

func TestBuildUDPListener(t *testing.T) {
	irListener := &ir.StreamListenerIR{
		Name:     "stream_udp_51820",
		Address:  "0.0.0.0",
		Port:     51820,
		Protocol: "udp",
		Cluster:  "stream_udp_51820",
	}
	irClusters := []*ir.ClusterIR{
		{Name: "stream_udp_51820", Host: "wg.ns.svc.cluster.local", Port: 51820, UseDNS: true},
	}
	clusterSet := make(map[string]bool)

	l, clusters := buildUDPListener(irListener, irClusters, clusterSet)

	require.NotNil(t, l)
	assert.Equal(t, "stream_udp_51820", l.Name)
	assert.Equal(t, corev3.SocketAddress_UDP, socketAddr(l).Protocol)
	assert.NotNil(t, l.UdpListenerConfig)
	require.Len(t, l.ListenerFilters, 1)
	assert.Equal(t, "envoy.filters.udp_listener.udp_proxy", l.ListenerFilters[0].Name)

	require.Len(t, clusters, 1)
}

// ---------------------------------------------------------------------------
// buildExtAuthzFilter
// ---------------------------------------------------------------------------

func TestBuildExtAuthzFilter_WithAuthelia(t *testing.T) {
	clusterMap := map[string]*ir.ClusterIR{
		"authelia_backend": {Name: "authelia_backend", Host: "authelia", Port: 9091},
	}

	filter := buildExtAuthzFilter(clusterMap, "testuser")

	require.NotNil(t, filter)
	assert.Equal(t, "envoy.filters.http.ext_authz", filter.Name)
}

func TestBuildExtAuthzFilter_NoAuthelia(t *testing.T) {
	clusterMap := map[string]*ir.ClusterIR{}
	filter := buildExtAuthzFilter(clusterMap, "testuser")
	assert.Nil(t, filter)
}

// ---------------------------------------------------------------------------
// Translate (end-to-end)
// ---------------------------------------------------------------------------

func TestTranslate_AllProtocols(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{
			{
				Name:       "http_redirect_80",
				Address:    "0.0.0.0",
				Port:       80,
				IsRedirect: true,
			},
			{
				Name:    "https_443",
				Address: "0.0.0.0",
				Port:    443,
				TLS:     true,
				VirtualHosts: []*ir.VirtualHostIR{{
					Name:    "app_vault",
					Domains: []string{"vault.alice.snowinning.com"},
					Routes: []*ir.HTTPRouteIR{{
						Name:       "default_vault",
						PathPrefix: "/",
						Cluster:    "app_vault_main",
					}},
				}},
			},
		},
		StreamListeners: []*ir.StreamListenerIR{
			{Name: "stream_tcp_30000", Address: "0.0.0.0", Port: 30000, Protocol: "tcp", Cluster: "stream_tcp_30000"},
			{Name: "stream_udp_51820", Address: "0.0.0.0", Port: 51820, Protocol: "udp", Cluster: "stream_udp_51820"},
		},
		Clusters: []*ir.ClusterIR{
			{Name: "app_vault_main", Host: "vault-svc.ns.svc.cluster.local", Port: 80, UseDNS: true},
			{Name: "authelia_backend", Host: "authelia-backend.os-framework.svc.cluster.local", Port: 9091, UseDNS: true},
			{Name: "stream_tcp_30000", Host: "game.ns.svc.cluster.local", Port: 30000, UseDNS: true},
			{Name: "stream_udp_51820", Host: "wg.ns.svc.cluster.local", Port: 51820, UseDNS: true},
		},
		Secrets: []*ir.SecretIR{
			{Name: "main-tls", CertData: "CERT", KeyData: "KEY"},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	// 4 listeners: HTTP redirect + HTTPS + TCP + UDP
	require.Len(t, snap.Listeners, 4)

	httpL := asListener(t, snap.Listeners[0])
	assert.Equal(t, "http_redirect_80", httpL.Name)

	httpsL := asListener(t, snap.Listeners[1])
	assert.Equal(t, "https_443", httpsL.Name)
	require.Len(t, httpsL.FilterChains, 1)

	// Verify HCM uses RDS and has ext_authz + router filters
	hcmFilter := httpsL.FilterChains[0].Filters[0]
	hcm := &hcmv3.HttpConnectionManager{}
	require.NoError(t, hcmFilter.GetTypedConfig().UnmarshalTo(hcm))
	assert.Len(t, hcm.HttpFilters, 3) // ext_authz + lua_sub_filter + router
	assert.NotNil(t, hcm.GetRds(), "HCM should use RDS")

	// 1 RDS route config for the HTTPS listener
	require.Len(t, snap.Routes, 1)

	tcpL := asListener(t, snap.Listeners[2])
	assert.Equal(t, "stream_tcp_30000", tcpL.Name)

	udpL := asListener(t, snap.Listeners[3])
	assert.Equal(t, "stream_udp_51820", udpL.Name)
	assert.NotNil(t, udpL.UdpListenerConfig)

	// Clusters: app_vault_main + authelia + stream_tcp + stream_udp
	require.Len(t, snap.Clusters, 4)
	for _, c := range snap.Clusters {
		cluster := asCluster(t, c)
		assert.Equal(t, connectTimeout, cluster.GetConnectTimeout().AsDuration())
	}
}

func TestTranslate_EmptyIR(t *testing.T) {
	xt := &XdsTranslator{}
	snap := xt.Translate(&ir.Xds{})

	assert.Empty(t, snap.Listeners)
	assert.Empty(t, snap.Clusters)
}

// ---------------------------------------------------------------------------
// buildAccessLog
// ---------------------------------------------------------------------------

func TestBuildAccessLog(t *testing.T) {
	al := buildAccessLog()
	require.NotNil(t, al)
	assert.Equal(t, "envoy.access_loggers.file", al.Name)
}

// ---------------------------------------------------------------------------
// Translate: default response virtual host
// ---------------------------------------------------------------------------

func TestTranslate_DefaultResponseVHost(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{{
			Name:    "https_443",
			Address: "0.0.0.0",
			Port:    443,
			TLS:     true,
			VirtualHosts: []*ir.VirtualHostIR{{
				Name:    "app",
				Domains: []string{"app.example.com"},
				Routes: []*ir.HTTPRouteIR{{
					Name:       "root",
					PathPrefix: "/",
					Cluster:    "backend",
				}},
			}},
			DefaultResponse: &ir.DirectResponseIR{
				Status: 421, Body: "<h1>Olares</h1>",
			},
		}},
		Clusters: []*ir.ClusterIR{
			{Name: "backend", Host: "backend.ns", Port: 80, UseDNS: true},
		},
		Secrets: []*ir.SecretIR{
			{Name: "main-tls", CertData: "CERT", KeyData: "KEY"},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	require.Len(t, snap.Listeners, 1)

	// Route config is now delivered via RDS, not inline
	require.Len(t, snap.Routes, 1)
	routeConfig, ok := snap.Routes[0].(*routev3.RouteConfiguration)
	require.True(t, ok)
	require.NotNil(t, routeConfig)

	// Last virtual host should be "default" with direct response
	lastVH := routeConfig.VirtualHosts[len(routeConfig.VirtualHosts)-1]
	assert.Equal(t, "default", lastVH.Name)
	assert.Equal(t, []string{"*"}, lastVH.Domains)

	defaultRoute := lastVH.Routes[0]
	assert.Equal(t, uint32(421), defaultRoute.GetDirectResponse().GetStatus())
}

// ---------------------------------------------------------------------------
// Translate: cluster dedup
// ---------------------------------------------------------------------------

func TestTranslate_ClusterDedup(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{
			{
				Name: "https_443", Address: "0.0.0.0", Port: 443, TLS: true,
				VirtualHosts: []*ir.VirtualHostIR{
					{Name: "vh1", Domains: []string{"a.com"},
						Routes: []*ir.HTTPRouteIR{{Name: "r1", PathPrefix: "/", Cluster: "shared"}}},
					{Name: "vh2", Domains: []string{"b.com"},
						Routes: []*ir.HTTPRouteIR{{Name: "r2", PathPrefix: "/", Cluster: "shared"}}},
				},
			},
			{
				Name: "https_444", Address: "0.0.0.0", Port: 444, TLS: true, ProxyProtocol: true,
				VirtualHosts: []*ir.VirtualHostIR{
					{Name: "vh3", Domains: []string{"a.com"},
						Routes: []*ir.HTTPRouteIR{{Name: "r3", PathPrefix: "/", Cluster: "shared"}}},
				},
			},
		},
		Clusters: []*ir.ClusterIR{
			{Name: "shared", Host: "svc.ns", Port: 80, UseDNS: true},
			{Name: "authelia_backend", Host: "authelia", Port: 9091, UseDNS: true},
		},
		Secrets: []*ir.SecretIR{
			{Name: "main-tls", CertData: "C", KeyData: "K"},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	// "shared" cluster should appear only once despite being referenced by 3 routes across 2 listeners
	sharedCount := 0
	for _, c := range snap.Clusters {
		cl := asCluster(t, c)
		if cl.Name == "shared" {
			sharedCount++
		}
	}
	assert.Equal(t, 1, sharedCount)
}

// ---------------------------------------------------------------------------
// Translate: virtual host with route matching
// ---------------------------------------------------------------------------

func TestTranslate_VirtualHostRouteDetails(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTPListeners: []*ir.HTTPListenerIR{{
			Name: "https_443", Address: "0.0.0.0", Port: 443, TLS: true,
			VirtualHosts: []*ir.VirtualHostIR{{
				Name:    "complex_app",
				Domains: []string{"app.example.com", "app.example.olares.local"},
				Routes: []*ir.HTTPRouteIR{
					{Name: "api_exact", PathExact: "/healthz", Cluster: "backend"},
					{Name: "api_prefix", PathPrefix: "/api/v1/", Cluster: "backend",
						RequestHeaders: map[string]string{"X-BFL-USER": "alice"}},
					{Name: "catch_all", PathPrefix: "/", Cluster: "backend",
						WebSocketUpgrade: true},
				},
			}},
		}},
		Clusters: []*ir.ClusterIR{
			{Name: "backend", Host: "svc.ns", Port: 80, UseDNS: true},
			{Name: "authelia_backend", Host: "authelia", Port: 9091, UseDNS: true},
		},
		Secrets: []*ir.SecretIR{
			{Name: "main-tls", CertData: "C", KeyData: "K"},
		},
	}

	xt := &XdsTranslator{}
	snap := xt.Translate(xdsIR)

	require.Len(t, snap.Listeners, 1)

	// Route config is now delivered via RDS
	require.Len(t, snap.Routes, 1)
	routeConfig, ok := snap.Routes[0].(*routev3.RouteConfiguration)
	require.True(t, ok)
	require.Len(t, routeConfig.VirtualHosts, 1)

	vh := routeConfig.VirtualHosts[0]
	require.Len(t, vh.Routes, 3)

	// Exact match
	assert.Equal(t, "/healthz", vh.Routes[0].Match.GetPath())

	// Prefix match with headers
	assert.Equal(t, "/api/v1/", vh.Routes[1].Match.GetPrefix())
	assert.True(t, len(vh.Routes[1].RequestHeadersToAdd) > 0)

	// Catch-all with websocket
	catchAll := vh.Routes[2]
	assert.Equal(t, "/", catchAll.Match.GetPrefix())
	ra, ok2 := catchAll.Action.(*routev3.Route_Route)
	require.True(t, ok2)
	assert.True(t, len(ra.Route.UpgradeConfigs) > 0)
}
