package translator

import (
	"context"
	"fmt"
	"net"
	"sort"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbac_configv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	accesslogfilev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	extauthzv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	luav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	rbac_filterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	tlsinspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	ppupstreamv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	rawtransportv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

const subFilterLuaScript = `
local PUSH_STATE_SCRIPT = "<script>\n" ..
  "(function () {\n" ..
  "  if (window.top == window) {\n" ..
  "    return;\n" ..
  "  }\n" ..
  "  const originalPushState = history.pushState;\n" ..
  "  const pushStateEvent = new Event(\"pushstate\");\n" ..
  "  history.pushState = function (...args) {\n" ..
  "    originalPushState.apply(this, args);\n" ..
  "    window.dispatchEvent(pushStateEvent);\n" ..
  "  };\n" ..
  "  window.addEventListener(\"pushstate\", () => {\n" ..
  "    window.parent.postMessage(\n" ..
  "      {type: \"locationHref\", message: location.href},\n" ..
  "      \"*\"\n" ..
  "    );\n" ..
  "  });\n" ..
  "})();\n" ..
  "</script>"

function envoy_on_request(request_handle)
  local meta = request_handle:metadata()
  if not meta then return end
  local pushstate = meta:get("pushstate")
  local language = meta:get("language")
  if not pushstate and not language then return end
  local dm = request_handle:streamInfo():dynamicMetadata()
  dm:set("envoy.filters.http.lua", "needs_sub_filter", "1")
  if pushstate then
    dm:set("envoy.filters.http.lua", "pushstate", pushstate)
  end
  if language then
    dm:set("envoy.filters.http.lua", "language", language)
  end
end

function envoy_on_response(response_handle)
  local dm = response_handle:streamInfo():dynamicMetadata():get("envoy.filters.http.lua")
  if not dm or not dm["needs_sub_filter"] then return end

  local ct = response_handle:headers():get("content-type")
  if not ct or not string.find(ct, "text/html") then return end

  local body = response_handle:body()
  if not body or body:length() == 0 then return end
  local body_str = body:getBytes(0, body:length())

  local modified = false
  local pushstate = dm["pushstate"]
  if pushstate then
    local new_str = string.gsub(body_str, "</html>", PUSH_STATE_SCRIPT .. "\n</html>", 1)
    if new_str ~= body_str then
      body_str = new_str
      modified = true
    end
  end

  local language = dm["language"]
  if language then
    local meta_tag = '<meta name="terminus-language" content="' .. language .. '"/>'
    local new_str = string.gsub(body_str, "</head>", meta_tag .. "\n</head>", 1)
    if new_str ~= body_str then
      body_str = new_str
      modified = true
    end
  end

  if modified then
    body:setBytes(body_str)
    response_handle:headers():replace("content-length", tostring(#body_str))
  end
end
`

var (
	tcpIdleTimeout        = time.Hour
	httpStreamIdleTimeout = 30 * time.Minute
	connectTimeout        = 5 * time.Second
	routeTimeout          = 5 * time.Minute
)

func SetTimeouts(tcpIdle, httpStream, connect, route time.Duration) {
	tcpIdleTimeout = tcpIdle
	httpStreamIdleTimeout = httpStream
	connectTimeout = connect
	routeTimeout = route
}

// mustAny marshals a proto.Message into an anypb.Any, panicking on error.
// All callers use well-known Envoy types whose type URLs are always registered,
// so a failure here indicates a programming error rather than a runtime condition.
func mustAny(m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	if err != nil {
		panic(fmt.Sprintf("anypb.New(%T): %v", m, err))
	}
	return a
}

type XdsTranslator struct {
	xdsIR        *message.XdsIR
	xdsResources *message.XdsResources
}

func New(xdsIR *message.XdsIR, xdsResources *message.XdsResources) *XdsTranslator {
	return &XdsTranslator{
		xdsIR:        xdsIR,
		xdsResources: xdsResources,
	}
}

func (t *XdsTranslator) Name() string { return "xds-translator" }

func (t *XdsTranslator) Start(ctx context.Context) error {
	klog.Info("xds-translator: starting")
	subscription := t.xdsIR.Subscribe(ctx)
	t.process(subscription)
	return nil
}

func (t *XdsTranslator) process(subscription <-chan watchable.Snapshot[string, *ir.Xds]) {
	first := true
	for snapshot := range subscription {
		if first {
			first = false
			for key, val := range snapshot.State {
				if val != nil {
					t.handleUpdate(key, val)
				}
			}
		}
		for _, update := range snapshot.Updates {
			if update.Delete {
				t.xdsResources.Delete(update.Key)
				continue
			}
			if update.Value != nil {
				t.handleUpdate(update.Key, update.Value)
			}
		}
	}
	klog.Info("xds-translator: subscription closed")
}

func (t *XdsTranslator) handleUpdate(key string, xdsIR *ir.Xds) {
	snapshot := t.Translate(xdsIR)

	if old, ok := t.xdsResources.Load(key); ok && old.Equal(snapshot) {
		klog.V(4).Infof("xds-translator: xDS unchanged for key %s, skipping", key)
		return
	}

	t.xdsResources.Store(key, snapshot)
	klog.Infof("xds-translator: published %d listeners, %d clusters, %d routes, %d secrets",
		len(snapshot.Listeners), len(snapshot.Clusters), len(snapshot.Routes), len(snapshot.Secrets))
}

func (t *XdsTranslator) Translate(xdsIR *ir.Xds) *message.XdsSnapshot {
	snap := &message.XdsSnapshot{}
	clusterSet := make(map[string]bool)

	// Build cluster map from IR
	clusterMap := make(map[string]*ir.ClusterIR)
	for _, c := range xdsIR.Clusters {
		clusterMap[c.Name] = c
	}

	// Build Secrets (SDS resources).
	// Certs are NOT embedded inline in the listener's TLS context.  Instead
	// they are served as separate Secret resources via the ADS/SDS stream and
	// referenced by name.  This means cert rotation only updates the Secret
	// resource; the Listener itself is unchanged → no Envoy listener drain.
	for _, s := range xdsIR.Secrets {
		snap.Secrets = append(snap.Secrets, &tlsv3.Secret{
			Name: s.Name,
			Type: &tlsv3.Secret_TlsCertificate{
				TlsCertificate: &tlsv3.TlsCertificate{
					CertificateChain: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: s.CertData},
					},
					PrivateKey: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: s.KeyData},
					},
				},
			},
		})
	}

	// L4 listeners (HTTP redirect, etc.)
	for _, listenerIR := range xdsIR.Listeners {
		switch listenerIR.Protocol {
		case ir.ProtocolHTTP:
			l := buildHTTPRedirectListener(listenerIR)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
		case ir.ProtocolTLS:
			l, cls := buildTLSListener(listenerIR, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		case ir.ProtocolTCP:
			l, cls := buildL4TCPListener(listenerIR, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		case ir.ProtocolUDP:
			l, cls := buildL4UDPListener(listenerIR, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		}
	}

	// L7 HTTPS listeners: group by (port, proxyProtocol) and merge into single Envoy listeners
	type listenerGroupKey struct {
		Port          uint32
		ProxyProtocol bool
	}
	groups := make(map[listenerGroupKey][]*ir.HTTPListenerIR)
	for _, l := range xdsIR.HTTPListeners {
		if l.IsRedirect {
			snap.Listeners = append(snap.Listeners, buildHTTPRedirectListenerFromHTTPIR(l))
			continue
		}
		key := listenerGroupKey{Port: l.Port, ProxyProtocol: l.ProxyProtocol}
		groups[key] = append(groups[key], l)
	}

	// Sort keys for deterministic output order.
	groupKeys := make([]listenerGroupKey, 0, len(groups))
	for gk := range groups {
		groupKeys = append(groupKeys, gk)
	}
	sort.Slice(groupKeys, func(i, j int) bool {
		if groupKeys[i].Port != groupKeys[j].Port {
			return groupKeys[i].Port < groupKeys[j].Port
		}
		return !groupKeys[i].ProxyProtocol
	})

	for _, gk := range groupKeys {
		l, rcs, cls := buildMultiUserHTTPSListener(gk.Port, gk.ProxyProtocol, groups[gk], clusterMap, clusterSet)
		if l != nil {
			snap.Listeners = append(snap.Listeners, l)
		}
		for _, rc := range rcs {
			snap.Routes = append(snap.Routes, rc)
		}
		snap.Clusters = append(snap.Clusters, cls...)
	}

	// Stream listeners
	for _, listenerIR := range xdsIR.StreamListeners {
		switch listenerIR.Protocol {
		case "tcp":
			l, cls := buildTCPStreamListener(listenerIR, clusterMap, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		case "udp":
			l, cls := buildUDPStreamListener(listenerIR, clusterMap, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		}
	}

	return snap
}

// buildMultiUserHTTPSListener merges all HTTPListenerIR entries that share the
// same (port, proxyProtocol) pair into a single Envoy listener. Each user gets
// one filter chain per ServerNameDomain, matched by SNI so that Envoy routes
// the connection to the correct per-user HCM (HTTP Connection Manager) and RDS
// route config without requiring separate listeners per user. A TLS inspector
// listener filter peeks at the ClientHello to extract the SNI before any filter
// chain is selected. If proxyProtocol is true, a PROXY protocol listener filter
// is prepended so the real client IP is available downstream.
func buildMultiUserHTTPSListener(port uint32, proxyProtocol bool, httpListeners []*ir.HTTPListenerIR, clusterMap map[string]*ir.ClusterIR, clusterSet map[string]bool) (*listenerv3.Listener, []*routev3.RouteConfiguration, []cachetypes.Resource) {
	var filterChains []*listenerv3.FilterChain
	var routeConfigs []*routev3.RouteConfiguration
	var clusters []cachetypes.Resource

	listenerName := fmt.Sprintf("https_%d", port)
	if proxyProtocol {
		listenerName = fmt.Sprintf("https_pp_%d", port)
	}

	for _, httpIR := range httpListeners {
		routeConfigName := httpIR.Name + "_routes"

		// Build virtual hosts
		var virtualHosts []*routev3.VirtualHost
		for _, vhIR := range httpIR.VirtualHosts {
			vh := translateVirtualHost(vhIR)
			virtualHosts = append(virtualHosts, vh)

			for _, route := range vhIR.Routes {
				if route.Cluster != "" && !clusterSet[route.Cluster] {
					if c, ok := clusterMap[route.Cluster]; ok {
						clusterSet[route.Cluster] = true
						clusters = append(clusters, buildL7Cluster(c))
					}
				}
				if route.ExtAuth != nil && route.ExtAuth.Cluster != "" && !clusterSet[route.ExtAuth.Cluster] {
					if c, ok := clusterMap[route.ExtAuth.Cluster]; ok {
						clusterSet[route.ExtAuth.Cluster] = true
						clusters = append(clusters, buildL7Cluster(c))
					}
				}
			}
		}

		// Default virtual host
		if httpIR.DefaultResponse != nil {
			defaultRoutes := []*routev3.Route{}
			if httpIR.PingRoute {
				defaultRoutes = append(defaultRoutes, &routev3.Route{
					Name:  "ping",
					Match: &routev3.RouteMatch{PathSpecifier: &routev3.RouteMatch_Path{Path: "/ping"}},
					Action: &routev3.Route_DirectResponse{
						DirectResponse: &routev3.DirectResponseAction{
							Status: 200,
							Body:   &corev3.DataSource{Specifier: &corev3.DataSource_InlineString{InlineString: "pong"}},
						},
					},
					ResponseHeadersToAdd: []*corev3.HeaderValueOption{{
						Header:       &corev3.HeaderValue{Key: "content-type", Value: "text/plain"},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					}},
				})
			}

			defaultRoutes = append(defaultRoutes, &routev3.Route{
				Match: &routev3.RouteMatch{PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"}},
				Action: &routev3.Route_DirectResponse{
					DirectResponse: &routev3.DirectResponseAction{
						Status: httpIR.DefaultResponse.Status,
						Body: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineString{InlineString: httpIR.DefaultResponse.Body},
						},
					},
				},
				ResponseHeadersToAdd: []*corev3.HeaderValueOption{{
					Header:       &corev3.HeaderValue{Key: "content-type", Value: "text/html; charset=utf-8"},
					AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
				}},
			})

			virtualHosts = append(virtualHosts, &routev3.VirtualHost{
				Name:    fmt.Sprintf("default_%s", httpIR.Name),
				Domains: []string{"*"},
				Routes:  defaultRoutes,
				TypedPerFilterConfig: map[string]*anypb.Any{
					"envoy.filters.http.ext_authz": mustAny(&extauthzv3.ExtAuthzPerRoute{
						Override: &extauthzv3.ExtAuthzPerRoute_Disabled{Disabled: true},
					}),
				},
				ResponseHeadersToAdd: []*corev3.HeaderValueOption{
					{Header: &corev3.HeaderValue{Key: "access-control-allow-headers", Value: "Accept, Content-Type, Accept-Encoding"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
					{Header: &corev3.HeaderValue{Key: "access-control-allow-methods", Value: "GET, OPTIONS"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
					{Header: &corev3.HeaderValue{Key: "access-control-allow-origin", Value: "*"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
				},
			})
		}

		routeConfig := &routev3.RouteConfiguration{
			Name:         routeConfigName,
			VirtualHosts: virtualHosts,
			RequestHeadersToAdd: []*corev3.HeaderValueOption{
				{Header: &corev3.HeaderValue{Key: "X-Forwarded-Host", Value: "%REQ(:AUTHORITY)%"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
				{Header: &corev3.HeaderValue{Key: "X-Forwarded-Proto", Value: "https"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
				{Header: &corev3.HeaderValue{Key: "X-Real-IP", Value: "%DOWNSTREAM_REMOTE_ADDRESS_WITHOUT_PORT%"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
				{Header: &corev3.HeaderValue{Key: "X-Original-Forwarded-For", Value: "%REQ(X-FORWARDED-FOR)%"}, AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD},
			},
		}
		routeConfigs = append(routeConfigs, routeConfig)

		// Derive the per-user authelia cluster name from the well-known naming
		// convention and verify it actually exists in the cluster map.
		autheliaClusterName := fmt.Sprintf("authelia_backend_%s", httpIR.UserName)
		if _, ok := clusterMap[autheliaClusterName]; !ok {
			autheliaClusterName = ""
		}

		// Build HTTP filters.
		//
		// The RBAC filter is always present (even for non-deny_all users).
		// With no rules it's a no-op (allows everything).  When a user
		// toggles deny_all, only the per-VH RBAC overrides in the RDS route
		// config change — the filter chain itself is byte-identical → no
		// Envoy listener drain on deny_all policy switch.
		httpFilters := []*hcmv3.HttpFilter{}
		extAuthzFilter := buildExtAuthzFilter(autheliaClusterName, clusterMap, httpIR.UserName)
		if extAuthzFilter != nil {
			httpFilters = append(httpFilters, extAuthzFilter)
		}

		httpFilters = append(httpFilters, &hcmv3.HttpFilter{
			Name:       "envoy.filters.http.rbac",
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(&rbac_filterv3.RBAC{})},
		})

		httpFilters = append(httpFilters, &hcmv3.HttpFilter{
			Name: "envoy.filters.http.lua",
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: mustAny(&luav3.Lua{
					DefaultSourceCode: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: subFilterLuaScript},
					},
				}),
			},
		})

		httpFilters = append(httpFilters, &hcmv3.HttpFilter{
			Name:       wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(&routerv3.Router{})},
		})

		hcm := &hcmv3.HttpConnectionManager{
			StatPrefix: httpIR.Name,
			RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
				Rds: &hcmv3.Rds{
					RouteConfigName: routeConfigName,
					ConfigSource: &corev3.ConfigSource{
						ResourceApiVersion: corev3.ApiVersion_V3,
						ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
							Ads: &corev3.AggregatedConfigSource{},
						},
					},
				},
			},
			HttpFilters: httpFilters,
			UpgradeConfigs: []*hcmv3.HttpConnectionManager_UpgradeConfig{
				{UpgradeType: "websocket"},
				{UpgradeType: "tailscale-control-protocol"},
			},
			AccessLog:         []*accesslogv3.AccessLog{buildHTTPAccessLog()},
			UseRemoteAddress:  wrapperspb.Bool(true),
			StreamIdleTimeout: durationpb.New(httpStreamIdleTimeout),
			StripPortMode: &hcmv3.HttpConnectionManager_StripAnyHostPort{
				StripAnyHostPort: true,
			},
			CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
				IdleTimeout: durationpb.New(httpStreamIdleTimeout),
			},
		}

		hcmFilter := &listenerv3.Filter{
			Name:       wellknown.HTTPConnectionManager,
			ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(hcm)},
		}

		// TLS context — reference the cert by SDS name instead of inline bytes.
		// The actual cert is delivered separately as a Secret resource via the
		// ADS stream.  When a cert rotates, only the Secret changes; the Listener
		// filter chain is unmodified → no Envoy listener drain on cert rotation.
		var transportSocket *corev3.TransportSocket
		if httpIR.TLSCert != nil {
			adsSource := &corev3.ConfigSource{
				ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
					Ads: &corev3.AggregatedConfigSource{},
				},
				ResourceApiVersion: corev3.ApiVersion_V3,
			}
			tlsContext := &tlsv3.DownstreamTlsContext{
				CommonTlsContext: &tlsv3.CommonTlsContext{
					TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{{
						Name:      httpIR.TLSCert.Name,
						SdsConfig: adsSource,
					}},
					AlpnProtocols: []string{"http/1.1"},
				},
			}
			transportSocket = &corev3.TransportSocket{
				Name:       "envoy.transport_sockets.tls",
				ConfigType: &corev3.TransportSocket_TypedConfig{TypedConfig: mustAny(tlsContext)},
			}
		}

		fc := &listenerv3.FilterChain{
			Name:            httpIR.Name,
			Filters:         []*listenerv3.Filter{hcmFilter},
			TransportSocket: transportSocket,
		}

		if len(httpIR.SNIMatches) > 0 || len(httpIR.SourceCIDRs) > 0 {
			match := &listenerv3.FilterChainMatch{}
			if len(httpIR.SNIMatches) > 0 {
				match.ServerNames = httpIR.SNIMatches
			}
			if len(httpIR.SourceCIDRs) > 0 {
				for _, cidr := range httpIR.SourceCIDRs {
					prefix, err := parseCIDR(cidr)
					if err != nil {
						klog.Warningf("xds-translator: parse CIDR %q: %v", cidr, err)
						continue
					}
					match.SourcePrefixRanges = append(match.SourcePrefixRanges, prefix)
				}
			}
			fc.FilterChainMatch = match
		}

		filterChains = append(filterChains, fc)

		// Ensure authelia cluster is built
		if autheliaClusterName != "" && !clusterSet[autheliaClusterName] {
			if c, ok := clusterMap[autheliaClusterName]; ok {
				clusterSet[autheliaClusterName] = true
				clusters = append(clusters, buildL7Cluster(c))
			}
		}
	}

	// Listener filters
	var listenerFilters []*listenerv3.ListenerFilter
	if proxyProtocol {
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name:       "envoy.filters.listener.proxy_protocol",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(&proxyprotocolv3.ProxyProtocol{})},
		})
	}
	listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
		Name:       "envoy.filters.listener.tls_inspector",
		ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(&tlsinspectorv3.TlsInspector{})},
	})

	listener := &listenerv3.Listener{
		Name: listenerName,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       "0.0.0.0",
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: port},
				},
			},
		},
		FilterChains:                  filterChains,
		ListenerFilters:               listenerFilters,
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(32768),
	}

	return listener, routeConfigs, clusters
}

// translateVirtualHost converts a VirtualHostIR into an Envoy VirtualHost.
// It attaches per-route Lua filter metadata for window.pushState injection and
// language tag insertion, sets CORS response headers, injects an OIDC
// Set-Cookie header when EnableOIDC is true, and disables ext_authz at the
// virtual-host level (routes that require auth re-enable it via per-route
// TypedPerFilterConfig). If SourceCIDRs is non-empty, an RBAC per-route
// override is added to restrict traffic to those CIDRs (deny_all mode).
func translateVirtualHost(vhIR *ir.VirtualHostIR) *routev3.VirtualHost {
	vh := &routev3.VirtualHost{
		Name:    vhIR.Name,
		Domains: vhIR.Domains,
	}

	var subFilterMeta *corev3.Metadata
	if vhIR.EnableWindowPushState || vhIR.Language != "" {
		fields := map[string]*structpb.Value{}
		if vhIR.EnableWindowPushState {
			fields["pushstate"] = structpb.NewStringValue("true")
		}
		if vhIR.Language != "" {
			fields["language"] = structpb.NewStringValue(vhIR.Language)
		}
		subFilterMeta = &corev3.Metadata{
			FilterMetadata: map[string]*structpb.Struct{
				"envoy.filters.http.lua": {Fields: fields},
			},
		}
	}

	for _, routeIR := range vhIR.Routes {
		route := translateRoute(routeIR)
		if subFilterMeta != nil && route.GetDirectResponse() == nil {
			route.Metadata = subFilterMeta
		}
		vh.Routes = append(vh.Routes, route)
	}

	if vhIR.EnableOIDC && vhIR.UserZone != "" {
		vh.ResponseHeadersToAdd = append(vh.ResponseHeadersToAdd, &corev3.HeaderValueOption{
			Header:       &corev3.HeaderValue{Key: "Set-Cookie", Value: fmt.Sprintf("prev-host=%%REQ(:AUTHORITY)%%;Domain=.%s;Path=/;", vhIR.UserZone)},
			AppendAction: corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		})
	}

	vh.ResponseHeadersToAdd = append(vh.ResponseHeadersToAdd,
		&corev3.HeaderValueOption{
			Header:       &corev3.HeaderValue{Key: "access-control-allow-headers", Value: "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization"},
			AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		},
		&corev3.HeaderValueOption{
			Header:       &corev3.HeaderValue{Key: "access-control-allow-methods", Value: "PUT, GET, DELETE, POST, OPTIONS"},
			AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		},
	)

	vh.TypedPerFilterConfig = map[string]*anypb.Any{
		"envoy.filters.http.ext_authz": mustAny(&extauthzv3.ExtAuthzPerRoute{
			Override: &extauthzv3.ExtAuthzPerRoute_Disabled{Disabled: true},
		}),
	}

	if len(vhIR.SourceCIDRs) > 0 {
		vh.TypedPerFilterConfig["envoy.filters.http.rbac"] = mustAny(buildRBACPerRouteFromCIDRs(vhIR.SourceCIDRs))
	}

	return vh
}

// translateRoute converts an HTTPRouteIR into an Envoy Route.
// Path matching priority: PathExact > PathRegex > PathPrefix (default "/").
// For direct-response routes the action is set immediately and the function
// returns early. For proxy routes a RouteAction with configurable timeout is
// built; WebSocket and tailscale-control-protocol upgrade configs are added
// when WebSocketUpgrade is true. Request headers from the IR map are added in
// sorted key order for deterministic output. If ExtAuth is set and not
// disabled, a per-route TypedPerFilterConfig enables the ext_authz check for
// that route.
func translateRoute(routeIR *ir.HTTPRouteIR) *routev3.Route {
	route := &routev3.Route{
		Name: routeIR.Name,
	}

	match := &routev3.RouteMatch{}
	if routeIR.PathExact != "" {
		match.PathSpecifier = &routev3.RouteMatch_Path{Path: routeIR.PathExact}
	} else if routeIR.PathRegex != "" {
		match.PathSpecifier = &routev3.RouteMatch_SafeRegex{
			SafeRegex: &matcherv3.RegexMatcher{Regex: routeIR.PathRegex},
		}
	} else {
		pfx := routeIR.PathPrefix
		if pfx == "" {
			pfx = "/"
		}
		match.PathSpecifier = &routev3.RouteMatch_Prefix{Prefix: pfx}
	}
	route.Match = match

	if routeIR.DirectResponse != nil {
		route.Action = &routev3.Route_DirectResponse{
			DirectResponse: &routev3.DirectResponseAction{
				Status: routeIR.DirectResponse.Status,
				Body: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineString{InlineString: routeIR.DirectResponse.Body},
				},
			},
		}
		return route
	}

	routeAction := &routev3.RouteAction{
		ClusterSpecifier: &routev3.RouteAction_Cluster{Cluster: routeIR.Cluster},
		Timeout:          durationpb.New(routeTimeout),
	}

	if routeIR.WebSocketUpgrade {
		routeAction.UpgradeConfigs = []*routev3.RouteAction_UpgradeConfig{
			{UpgradeType: "websocket"},
			{UpgradeType: "tailscale-control-protocol"},
		}
	}

	route.Action = &routev3.Route_Route{Route: routeAction}

	if len(routeIR.RequestHeaders) > 0 {
		keys := make([]string, 0, len(routeIR.RequestHeaders))
		for k := range routeIR.RequestHeaders {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			route.RequestHeadersToAdd = append(route.RequestHeadersToAdd, &corev3.HeaderValueOption{
				Header:       &corev3.HeaderValue{Key: k, Value: routeIR.RequestHeaders[k]},
				AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
			})
		}
	}

	if routeIR.ExtAuth != nil && !routeIR.ExtAuth.Disabled {
		route.TypedPerFilterConfig = map[string]*anypb.Any{
			"envoy.filters.http.ext_authz": mustAny(&extauthzv3.ExtAuthzPerRoute{
				Override: &extauthzv3.ExtAuthzPerRoute_CheckSettings{
					CheckSettings: &extauthzv3.CheckSettings{},
				},
			}),
		}
	}

	return route
}

// buildExtAuthzFilter constructs the envoy.filters.http.ext_authz HTTP filter
// that delegates authorization decisions to the per-user Authelia instance.
// Returns nil when autheliaClusterName is empty (no Authelia available for
// this user) so the caller can skip adding the filter entirely.
//
// The filter is configured in HTTP service mode pointing at the Authelia
// cluster. Selected request headers (host, authorization, cookie, etc.) are
// forwarded to Authelia via AllowedHeaders; Authelia's response headers
// prefixed with "remote-" or "authelia-" are propagated upstream, and
// "set-cookie" is forwarded back to the client on both allow and deny.
// FailureModeAllow is false so a broken Authelia causes requests to be denied
// rather than passed through. ClearRouteCache is false because auth decisions
// do not change the selected route.
func buildExtAuthzFilter(autheliaClusterName string, clusterMap map[string]*ir.ClusterIR, userName string) *hcmv3.HttpFilter {
	if autheliaClusterName == "" {
		return nil
	}
	clusterIR, ok := clusterMap[autheliaClusterName]
	if !ok {
		return nil
	}

	extAuthz := &extauthzv3.ExtAuthz{
		Services: &extauthzv3.ExtAuthz_HttpService{
			HttpService: &extauthzv3.HttpService{
				ServerUri: &corev3.HttpUri{
					Uri:              fmt.Sprintf("http://%s", clusterIR.Host),
					HttpUpstreamType: &corev3.HttpUri_Cluster{Cluster: autheliaClusterName},
					Timeout:          durationpb.New(15 * time.Second),
				},
				PathPrefix: "/api/authz/ext-authz/",
				AuthorizationRequest: &extauthzv3.AuthorizationRequest{
					HeadersToAdd: []*corev3.HeaderValue{
						{Key: "X-Forwarded-Proto", Value: "https"},
						{Key: "X-BFL-USER", Value: userName},
					},
				},
				AuthorizationResponse: &extauthzv3.AuthorizationResponse{
					AllowedUpstreamHeaders: &matcherv3.ListStringMatcher{
						Patterns: []*matcherv3.StringMatcher{
							{MatchPattern: &matcherv3.StringMatcher_Prefix{Prefix: "remote-"}},
							{MatchPattern: &matcherv3.StringMatcher_Prefix{Prefix: "authelia-"}},
						},
					},
					AllowedClientHeaders: &matcherv3.ListStringMatcher{
						Patterns: []*matcherv3.StringMatcher{
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "set-cookie"}},
						},
					},
					AllowedClientHeadersOnSuccess: &matcherv3.ListStringMatcher{
						Patterns: []*matcherv3.StringMatcher{
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "set-cookie"}},
						},
					},
				},
			},
		},
		// AllowedHeaders replaces the deprecated AuthorizationRequest.allowed_headers.
		// It controls which request headers are forwarded to the authz server.
		AllowedHeaders: &matcherv3.ListStringMatcher{
			Patterns: []*matcherv3.StringMatcher{
				{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "host"}},
				{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "authorization"}},
				{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "proxy-authorization"}},
				{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "accept"}},
				{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "cookie"}},
			},
		},
		TransportApiVersion: corev3.ApiVersion_V3,
		FailureModeAllow:    false,
		ClearRouteCache:     false,
	}

	return &hcmv3.HttpFilter{
		Name:       "envoy.filters.http.ext_authz",
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(extAuthz)},
	}
}

// buildRBACPerRouteFromCIDRs returns an RBACPerRoute config that only allows
// traffic from the given CIDRs.  This is used for deny_all users: their
// restricted VHs get this per-VH override so that only VPN / local traffic
// can reach them, while allowed VHs inherit the default (no RBAC = allow all).
//
// The RBAC policy uses remote_ip (not direct_remote_ip) so it works correctly
// with proxy protocol — remote_ip reflects the real client IP after Envoy's
// use_remote_address + proxy protocol processing.
func buildRBACPerRouteFromCIDRs(cidrs []string) *rbac_filterv3.RBACPerRoute {
	var principals []*rbac_configv3.Principal
	for _, cidr := range cidrs {
		prefix, err := parseCIDR(cidr)
		if err != nil {
			klog.Warningf("xds-translator: RBAC parseCIDR %q: %v", cidr, err)
			continue
		}
		principals = append(principals, &rbac_configv3.Principal{
			Identifier: &rbac_configv3.Principal_RemoteIp{
				RemoteIp: prefix,
			},
		})
	}

	return &rbac_filterv3.RBACPerRoute{
		Rbac: &rbac_filterv3.RBAC{
			Rules: &rbac_configv3.RBAC{
				Action: rbac_configv3.RBAC_ALLOW,
				Policies: map[string]*rbac_configv3.Policy{
					"allow_cidrs": {
						Permissions: []*rbac_configv3.Permission{{
							Rule: &rbac_configv3.Permission_Any{Any: true},
						}},
						Principals: principals,
					},
				},
			},
		},
	}
}

// buildHTTPRedirectListener creates an Envoy listener (from a ListenerIR) that
// issues a 301 HTTP→HTTPS redirect for every request. It is used for the plain
// HTTP port (typically 81) that the L4 IR layer emits as ProtocolHTTP.
// Returns nil when the IR carries no HTTPRedirect directive.
func buildHTTPRedirectListener(listenerIR *ir.ListenerIR) *listenerv3.Listener {
	if listenerIR.HTTPRedirect == nil {
		return nil
	}

	routeConfig := &routev3.RouteConfiguration{
		Name: listenerIR.Name + "_routes",
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "redirect",
			Domains: []string{"*"},
			Routes: []*routev3.Route{{
				Match: &routev3.RouteMatch{PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"}},
				Action: &routev3.Route_Redirect{
					Redirect: &routev3.RedirectAction{
						SchemeRewriteSpecifier: &routev3.RedirectAction_HttpsRedirect{HttpsRedirect: true},
						ResponseCode:           routev3.RedirectAction_MOVED_PERMANENTLY,
					},
				},
			}},
		}},
	}

	hcm := &hcmv3.HttpConnectionManager{
		StatPrefix: listenerIR.Name,
		RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		},
		StreamIdleTimeout: durationpb.New(httpStreamIdleTimeout),
		HttpFilters: []*hcmv3.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(&routerv3.Router{})},
		}},
		InternalAddressConfig: &hcmv3.HttpConnectionManager_InternalAddressConfig{},
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(hcm)},
			}},
		}},
	}
}

// buildHTTPRedirectListenerFromHTTPIR creates an Envoy listener (from an
// HTTPListenerIR marked IsRedirect) that issues a 301 HTTP→HTTPS redirect for
// every request. This variant accepts the higher-level HTTPListenerIR instead
// of the raw ListenerIR so it can be used directly from the HTTPS listener
// grouping loop.
func buildHTTPRedirectListenerFromHTTPIR(listenerIR *ir.HTTPListenerIR) *listenerv3.Listener {
	routeConfig := &routev3.RouteConfiguration{
		Name: listenerIR.Name + "_routes",
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "redirect",
			Domains: []string{"*"},
			Routes: []*routev3.Route{{
				Match: &routev3.RouteMatch{PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"}},
				Action: &routev3.Route_Redirect{
					Redirect: &routev3.RedirectAction{
						SchemeRewriteSpecifier: &routev3.RedirectAction_HttpsRedirect{HttpsRedirect: true},
						ResponseCode:           routev3.RedirectAction_MOVED_PERMANENTLY,
					},
				},
			}},
		}},
	}

	hcm := &hcmv3.HttpConnectionManager{
		StatPrefix:     listenerIR.Name,
		RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{RouteConfig: routeConfig},
		HttpFilters: []*hcmv3.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(&routerv3.Router{})},
		}},
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(hcm)},
			}},
		}},
	}
}

// buildTLSListener creates an Envoy TLS passthrough listener from a ListenerIR
// with Protocol=TLS. Each RouteIR becomes a separate filter chain; SNI and
// source-prefix ranges on the route are translated into FilterChainMatch
// criteria so Envoy selects the correct chain without terminating TLS itself.
// Optional ProxyProtocol and TLSInspector listener filters are added according
// to the IR flags. L4 STRICT_DNS clusters are built for each unique
// destination and returned alongside the listener.
func buildTLSListener(listenerIR *ir.ListenerIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var filterChains []*listenerv3.FilterChain
	var clusters []cachetypes.Resource

	for _, route := range listenerIR.Routes {
		if route.Destination == nil {
			continue
		}

		clusterName := route.Destination.Name
		if !clusterSet[clusterName] {
			clusterSet[clusterName] = true
			clusters = append(clusters, buildL4Cluster(route.Destination, route.ProxyProtocolUpstream))
		}

		fc := buildFilterChain(route, clusterName)
		filterChains = append(filterChains, fc)
	}

	var listenerFilters []*listenerv3.ListenerFilter
	if listenerIR.ProxyProtocol {
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name:       "envoy.filters.listener.proxy_protocol",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(&proxyprotocolv3.ProxyProtocol{})},
		})
	}
	if listenerIR.TLSInspector {
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name:       "envoy.filters.listener.tls_inspector",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(&tlsinspectorv3.TlsInspector{})},
		})
	}

	listener := &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		FilterChains:    filterChains,
		ListenerFilters: listenerFilters,
	}

	return listener, clusters
}

// buildFilterChain builds a single Envoy FilterChain for a TLS-passthrough
// route. The chain contains a TCP proxy filter forwarding traffic to
// clusterName with an idle timeout and access logging. If the RouteIR carries
// SNI patterns or source CIDR ranges, a FilterChainMatch is set so that Envoy
// only selects this chain when the connection attributes match.
func buildFilterChain(route *ir.RouteIR, clusterName string) *listenerv3.FilterChain {
	tcpProxy := &tcpproxyv3.TcpProxy{
		StatPrefix:       route.Name,
		ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{Cluster: clusterName},
		IdleTimeout:      durationpb.New(tcpIdleTimeout),
		AccessLog:        []*accesslogv3.AccessLog{buildTCPAccessLog()},
	}

	fc := &listenerv3.FilterChain{
		Name: route.Name,
		Filters: []*listenerv3.Filter{{
			Name:       wellknown.TCPProxy,
			ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(tcpProxy)},
		}},
	}

	if len(route.SNIMatches) > 0 || len(route.SourcePrefixRanges) > 0 {
		match := &listenerv3.FilterChainMatch{}
		if len(route.SNIMatches) > 0 {
			match.ServerNames = route.SNIMatches
		}
		if len(route.SourcePrefixRanges) > 0 {
			for _, cidr := range route.SourcePrefixRanges {
				prefix, err := parseCIDR(cidr)
				if err != nil {
					klog.Warningf("xds-translator: parse CIDR %q: %v", cidr, err)
					continue
				}
				match.SourcePrefixRanges = append(match.SourcePrefixRanges, prefix)
			}
		}
		fc.FilterChainMatch = match
	}

	return fc
}

// buildTCPStreamListener creates an Envoy TCP listener for a user-declared
// stream port (protocol="tcp"). Unlike TLS passthrough listeners there is no
// SNI matching — all traffic arriving on the port is forwarded to the single
// upstream cluster via a TCP proxy filter. The cluster is built as an L7
// (STRICT_DNS) cluster so Kubernetes DNS resolution is used. This listener is
// used for raw TCP services such as game servers or custom TCP protocols
// exposed through the proxy.
func buildTCPStreamListener(listenerIR *ir.StreamListenerIR, clusterMap map[string]*ir.ClusterIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource
	if !clusterSet[listenerIR.Cluster] {
		if c, ok := clusterMap[listenerIR.Cluster]; ok {
			clusterSet[c.Name] = true
			clusters = append(clusters, buildL7Cluster(c))
		}
	}

	tcpProxy := &tcpproxyv3.TcpProxy{
		StatPrefix:       listenerIR.Name,
		ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{Cluster: listenerIR.Cluster},
		IdleTimeout:      durationpb.New(tcpIdleTimeout),
		AccessLog:        []*accesslogv3.AccessLog{buildTCPAccessLog()},
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name:       wellknown.TCPProxy,
				ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(tcpProxy)},
			}},
		}},
	}, clusters
}

// buildUDPStreamListener creates an Envoy UDP listener for a user-declared
// stream port (protocol="udp"). Envoy models UDP listeners differently from
// TCP: the udp_proxy filter is placed inside a ListenerFilter (not a
// FilterChain), and the socket address must specify Protocol=UDP. The upstream
// cluster is built as an L7 (STRICT_DNS) cluster identical to the TCP stream
// case.
func buildUDPStreamListener(listenerIR *ir.StreamListenerIR, clusterMap map[string]*ir.ClusterIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource
	if !clusterSet[listenerIR.Cluster] {
		if c, ok := clusterMap[listenerIR.Cluster]; ok {
			clusterSet[c.Name] = true
			clusters = append(clusters, buildL7Cluster(c))
		}
	}

	udpProxy := &udpproxyv3.UdpProxyConfig{
		StatPrefix:     listenerIR.Name,
		RouteSpecifier: &udpproxyv3.UdpProxyConfig_Cluster{Cluster: listenerIR.Cluster},
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					Protocol:      corev3.SocketAddress_UDP,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		ListenerFilters: []*listenerv3.ListenerFilter{{
			Name:       "envoy.filters.udp_listener.udp_proxy",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(udpProxy)},
		}},
		UdpListenerConfig: &listenerv3.UdpListenerConfig{},
	}, clusters
}

// buildL4TCPListener creates an Envoy TCP listener from a ListenerIR with
// Protocol=TCP. Each RouteIR becomes a separate named FilterChain containing a
// TCP proxy filter; the chains are differentiated by the route name rather than
// by SNI or source IP (no FilterChainMatch is set). This is used for plain TCP
// forwarding where TLS inspection is not required, e.g. internal service
// tunnels. L4 STRICT_DNS clusters are built for each unique destination.
func buildL4TCPListener(listenerIR *ir.ListenerIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var filterChains []*listenerv3.FilterChain
	var clusters []cachetypes.Resource

	for _, route := range listenerIR.Routes {
		if route.Destination == nil {
			continue
		}
		clusterName := route.Destination.Name
		if !clusterSet[clusterName] {
			clusterSet[clusterName] = true
			clusters = append(clusters, buildL4Cluster(route.Destination, false))
		}

		tcpProxy := &tcpproxyv3.TcpProxy{
			StatPrefix:       route.Name,
			ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{Cluster: clusterName},
			IdleTimeout:      durationpb.New(tcpIdleTimeout),
			AccessLog:        []*accesslogv3.AccessLog{buildTCPAccessLog()},
		}

		filterChains = append(filterChains, &listenerv3.FilterChain{
			Name: route.Name,
			Filters: []*listenerv3.Filter{{
				Name:       wellknown.TCPProxy,
				ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(tcpProxy)},
			}},
		})
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		FilterChains: filterChains,
	}, clusters
}

// buildL4UDPListener creates an Envoy UDP listener from a ListenerIR with
// Protocol=UDP. Only the first route is used because a UDP listener has a
// single forwarding target (the udp_proxy filter does not support multiple
// clusters via filter chains the way TCP does). Returns nil when no valid
// route with a destination is present. The L4 STRICT_DNS cluster is built for
// the single destination.
func buildL4UDPListener(listenerIR *ir.ListenerIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource

	if len(listenerIR.Routes) == 0 || listenerIR.Routes[0].Destination == nil {
		return nil, nil
	}
	route := listenerIR.Routes[0]
	clusterName := route.Destination.Name

	if !clusterSet[clusterName] {
		clusterSet[clusterName] = true
		clusters = append(clusters, buildL4Cluster(route.Destination, false))
	}

	udpProxy := &udpproxyv3.UdpProxyConfig{
		StatPrefix:     route.Name,
		RouteSpecifier: &udpproxyv3.UdpProxyConfig_Cluster{Cluster: clusterName},
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:       listenerIR.Address,
					Protocol:      corev3.SocketAddress_UDP,
					PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: listenerIR.Port},
				},
			},
		},
		ListenerFilters: []*listenerv3.ListenerFilter{{
			Name:       "envoy.filters.udp_listener.udp_proxy",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(udpProxy)},
		}},
		UdpListenerConfig: &listenerv3.UdpListenerConfig{},
	}, clusters
}

func buildL7Cluster(c *ir.ClusterIR) *clusterv3.Cluster {
	discoveryType := clusterv3.Cluster_STRICT_DNS
	if !c.UseDNS {
		discoveryType = clusterv3.Cluster_STATIC
	}
	return &clusterv3.Cluster{
		Name:                 c.Name,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: discoveryType},
		ConnectTimeout:       durationpb.New(connectTimeout),
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: c.Name,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address:       c.Host,
										PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: c.Port},
									},
								},
							},
						},
					},
				}},
			}},
		},
	}
}

func buildL4Cluster(dest *ir.DestinationIR, proxyProtocolUpstream bool) *clusterv3.Cluster {
	cluster := &clusterv3.Cluster{
		Name:                 dest.Name,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STRICT_DNS},
		ConnectTimeout:       durationpb.New(connectTimeout),
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: dest.Name,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address:       dest.Host,
										PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: dest.Port},
									},
								},
							},
						},
					},
				}},
			}},
		},
	}

	if proxyProtocolUpstream {
		rawBufAny := mustAny(&rawtransportv3.RawBuffer{})
		ppUpstream := &ppupstreamv3.ProxyProtocolUpstreamTransport{
			Config: &corev3.ProxyProtocolConfig{Version: corev3.ProxyProtocolConfig_V1},
			TransportSocket: &corev3.TransportSocket{
				Name:       "envoy.transport_sockets.raw_buffer",
				ConfigType: &corev3.TransportSocket_TypedConfig{TypedConfig: rawBufAny},
			},
		}
		cluster.TransportSocket = &corev3.TransportSocket{
			Name:       "envoy.transport_sockets.upstream_proxy_protocol",
			ConfigType: &corev3.TransportSocket_TypedConfig{TypedConfig: mustAny(ppUpstream)},
		}
	}

	return cluster
}

func buildHTTPAccessLog() *accesslogv3.AccessLog {
	fileLog := &accesslogfilev3.FileAccessLog{
		Path: "/dev/stdout",
		AccessLogFormat: &accesslogfilev3.FileAccessLog_LogFormat{
			LogFormat: &corev3.SubstitutionFormatString{
				Format: &corev3.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{
							InlineString: "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS% -> %UPSTREAM_HOST% %REQ(:AUTHORITY)% %REQ(:PATH)% %RESPONSE_CODE% duration=%DURATION%ms rx=%BYTES_RECEIVED% tx=%BYTES_SENT% flags=%RESPONSE_FLAGS% route=%ROUTE_NAME% cluster=%UPSTREAM_CLUSTER% details=%RESPONSE_CODE_DETAILS% ufail=%UPSTREAM_TRANSPORT_FAILURE_REASON%\n",
						},
					},
				},
			},
		},
	}
	return &accesslogv3.AccessLog{
		Name:       "envoy.access_loggers.file",
		ConfigType: &accesslogv3.AccessLog_TypedConfig{TypedConfig: mustAny(fileLog)},
	}
}

func buildTCPAccessLog() *accesslogv3.AccessLog {
	fileLog := &accesslogfilev3.FileAccessLog{
		Path: "/dev/stdout",
		AccessLogFormat: &accesslogfilev3.FileAccessLog_LogFormat{
			LogFormat: &corev3.SubstitutionFormatString{
				Format: &corev3.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{
							InlineString: "[%START_TIME%] %DOWNSTREAM_REMOTE_ADDRESS% -> %UPSTREAM_HOST% SNI=%REQUESTED_SERVER_NAME% duration=%DURATION%ms rx=%BYTES_RECEIVED% tx=%BYTES_SENT% flags=%RESPONSE_FLAGS%\n",
						},
					},
				},
			},
		},
	}
	return &accesslogv3.AccessLog{
		Name:       "envoy.access_loggers.file",
		ConfigType: &accesslogv3.AccessLog_TypedConfig{TypedConfig: mustAny(fileLog)},
	}
}

func parseCIDR(cidr string) (*corev3.CidrRange, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return nil, fmt.Errorf("invalid CIDR %q", cidr)
		}
		return &corev3.CidrRange{
			AddressPrefix: ip.String(),
			PrefixLen:     wrapperspb.UInt32(32),
		}, nil
	}
	ones, _ := ipNet.Mask.Size()
	return &corev3.CidrRange{
		AddressPrefix: ipNet.IP.String(),
		PrefixLen:     wrapperspb.UInt32(uint32(ones)),
	}, nil
}

// Suppress unused import warnings
var _ = matcherv3.StringMatcher{}
var _ = structpb.Value{}
var _ = tlsv3.DownstreamTlsContext{}
