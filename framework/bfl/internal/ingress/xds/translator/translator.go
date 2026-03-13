package translator

import (
	"context"
	"fmt"
	"sort"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	accesslogfilev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	extauthzv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	luav3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	tlsinspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"bytetrade.io/web3os/bfl/internal/ingress/ir"
	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

func mustAny(m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	if err != nil {
		panic(fmt.Sprintf("anypb.New(%T): %v", m, err))
	}
	return a
}

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
    local new_str = string.gsub(body_str, "</html>", PUSH_STATE_SCRIPT .. "\n</html>")
    if new_str ~= body_str then
      body_str = new_str
      modified = true
    end
  end

  local language = dm["language"]
  if language then
    local meta_tag = '<meta name="terminus-language" content="' .. language .. '"/>'
    local new_str
    if pushstate then
      new_str = string.gsub(body_str, "</head>", meta_tag .. "\n</head>")
    else
      new_str = string.gsub(body_str, "</head>", meta_tag .. "\n</head>", 1)
    end
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
	httpStreamIdleTimeout = 30 * time.Minute
	connectTimeout        = 5 * time.Second
	tcpIdleTimeout        = time.Hour
	routeTimeout          = 5 * time.Minute
)

type XdsTranslator struct {
	xdsIR        *message.XdsIR
	xdsResources *message.XdsResources
}

func SetTimeouts(tcpIdle, httpStream, connect, route time.Duration) {
	tcpIdleTimeout = tcpIdle
	httpStreamIdleTimeout = httpStream
	connectTimeout = connect
	routeTimeout = route
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

	// Build secrets map for TLS lookup
	secretMap := make(map[string]*ir.SecretIR)
	for _, s := range xdsIR.Secrets {
		secretMap[s.Name] = s
	}

	// HTTP listeners
	for _, listenerIR := range xdsIR.HTTPListeners {
		if listenerIR.IsRedirect {
			l := buildHTTPRedirectListener(listenerIR)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			continue
		}

		l, rc, cls := buildHTTPSListener(listenerIR, xdsIR.Clusters, secretMap, clusterSet)
		if l != nil {
			snap.Listeners = append(snap.Listeners, l)
		}
		if rc != nil {
			snap.Routes = append(snap.Routes, rc)
		}
		snap.Clusters = append(snap.Clusters, cls...)
	}

	// Stream listeners
	for _, listenerIR := range xdsIR.StreamListeners {
		switch listenerIR.Protocol {
		case "tcp":
			l, cls := buildTCPListener(listenerIR, xdsIR.Clusters, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		case "udp":
			l, cls := buildUDPListener(listenerIR, xdsIR.Clusters, clusterSet)
			if l != nil {
				snap.Listeners = append(snap.Listeners, l)
			}
			snap.Clusters = append(snap.Clusters, cls...)
		}
	}

	return snap
}

func buildHTTPRedirectListener(listenerIR *ir.HTTPListenerIR) *listenerv3.Listener {
	routeConfig := &routev3.RouteConfiguration{
		Name: listenerIR.Name + "_routes",
		VirtualHosts: []*routev3.VirtualHost{{
			Name:    "redirect",
			Domains: []string{"*"},
			Routes: []*routev3.Route{{
				Match: &routev3.RouteMatch{
					PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"},
				},
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

func buildHTTPSListener(listenerIR *ir.HTTPListenerIR, irClusters []*ir.ClusterIR, secretMap map[string]*ir.SecretIR, clusterSet map[string]bool) (*listenerv3.Listener, *routev3.RouteConfiguration, []cachetypes.Resource) {
	var clusters []cachetypes.Resource
	clusterMap := make(map[string]*ir.ClusterIR)
	for _, c := range irClusters {
		clusterMap[c.Name] = c
	}

	// Build virtual hosts
	var virtualHosts []*routev3.VirtualHost
	for _, vhIR := range listenerIR.VirtualHosts {
		vh := translateVirtualHost(vhIR)
		virtualHosts = append(virtualHosts, vh)

		// Collect clusters used by this virtual host
		for _, route := range vhIR.Routes {
			if route.Cluster != "" && !clusterSet[route.Cluster] {
				if c, ok := clusterMap[route.Cluster]; ok {
					clusterSet[route.Cluster] = true
					clusters = append(clusters, buildCluster(c))
				}
			}
			if route.ExtAuth != nil && route.ExtAuth.Cluster != "" && !clusterSet[route.ExtAuth.Cluster] {
				if c, ok := clusterMap[route.ExtAuth.Cluster]; ok {
					clusterSet[route.ExtAuth.Cluster] = true
					clusters = append(clusters, buildCluster(c))
				}
			}
		}
	}

	// Default virtual host for unmatched hosts — no auth needed
	if listenerIR.DefaultResponse != nil {
		defaultRoutes := []*routev3.Route{}

		if listenerIR.PingRoute {
			defaultRoutes = append(defaultRoutes, &routev3.Route{
				Name: "ping",
				Match: &routev3.RouteMatch{
					PathSpecifier: &routev3.RouteMatch_Path{Path: "/ping"},
				},
				Action: &routev3.Route_DirectResponse{
					DirectResponse: &routev3.DirectResponseAction{
						Status: 200,
						Body: &corev3.DataSource{
							Specifier: &corev3.DataSource_InlineString{InlineString: "pong"},
						},
					},
				},
				ResponseHeadersToAdd: []*corev3.HeaderValueOption{{
					Header:       &corev3.HeaderValue{Key: "content-type", Value: "text/plain"},
					AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
				}},
			})
		}

		defaultRoutes = append(defaultRoutes, &routev3.Route{
			Match: &routev3.RouteMatch{
				PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"},
			},
			Action: &routev3.Route_DirectResponse{
				DirectResponse: &routev3.DirectResponseAction{
					Status: listenerIR.DefaultResponse.Status,
					Body: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{
							InlineString: listenerIR.DefaultResponse.Body,
						},
					},
				},
			},
		})

		virtualHosts = append(virtualHosts, &routev3.VirtualHost{
			Name:    "default",
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

	routeConfigName := listenerIR.Name + "_routes"
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

	// Build HTTP filters: ext_authz (Authelia, for files/settings) -> router
	userName := ""
	for _, vh := range listenerIR.VirtualHosts {
		if vh.UserName != "" {
			userName = vh.UserName
			break
		}
	}

	httpFilters := []*hcmv3.HttpFilter{}

	extAuthzFilter := buildExtAuthzFilter(clusterMap, userName)
	if extAuthzFilter != nil {
		httpFilters = append(httpFilters, extAuthzFilter)
	}

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
		StatPrefix: listenerIR.Name,
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
		AccessLog:         []*accesslogv3.AccessLog{buildAccessLog()},
		UseRemoteAddress:  wrapperspb.Bool(true),
		StreamIdleTimeout: durationpb.New(httpStreamIdleTimeout),
		StripPortMode: &hcmv3.HttpConnectionManager_StripAnyHostPort{
			StripAnyHostPort: true,
		},
		CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
			IdleTimeout: durationpb.New(httpStreamIdleTimeout),
		},
	}

	// TLS context
	var tlsContext *tlsv3.DownstreamTlsContext
	if mainSecret, ok := secretMap["main-tls"]; ok {
		tlsContext = &tlsv3.DownstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificates: []*tlsv3.TlsCertificate{{
					CertificateChain: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: mainSecret.CertData},
					},
					PrivateKey: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: mainSecret.KeyData},
					},
				}},
				AlpnProtocols: []string{"http/1.1"},
			},
		}
	}

	hcmFilter := &listenerv3.Filter{
		Name:       wellknown.HTTPConnectionManager,
		ConfigType: &listenerv3.Filter_TypedConfig{TypedConfig: mustAny(hcm)},
	}

	var filterChains []*listenerv3.FilterChain

	// Custom domain filter chains (with their own TLS certs, matched by SNI)
	for _, s := range secretMap {
		if s.Name == "main-tls" {
			continue
		}
		// Extract domain name from secret name "custom-tls-{domain}"
		domain := ""
		if len(s.Name) > len("custom-tls-") {
			domain = s.Name[len("custom-tls-"):]
		}
		if domain == "" {
			continue
		}
		customTLS := &tlsv3.DownstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsCertificates: []*tlsv3.TlsCertificate{{
					CertificateChain: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: s.CertData},
					},
					PrivateKey: &corev3.DataSource{
						Specifier: &corev3.DataSource_InlineString{InlineString: s.KeyData},
					},
				}},
				AlpnProtocols: []string{"http/1.1"},
			},
		}
		filterChains = append(filterChains, &listenerv3.FilterChain{
			FilterChainMatch: &listenerv3.FilterChainMatch{
				ServerNames: []string{domain},
			},
			Filters: []*listenerv3.Filter{hcmFilter},
			TransportSocket: &corev3.TransportSocket{
				Name:       "envoy.transport_sockets.tls",
				ConfigType: &corev3.TransportSocket_TypedConfig{TypedConfig: mustAny(customTLS)},
			},
		})
	}

	// Main filter chain (default, serves the main wildcard cert)
	mainFilterChain := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{hcmFilter},
	}
	if tlsContext != nil {
		mainFilterChain.TransportSocket = &corev3.TransportSocket{
			Name:       "envoy.transport_sockets.tls",
			ConfigType: &corev3.TransportSocket_TypedConfig{TypedConfig: mustAny(tlsContext)},
		}
	}
	filterChains = append(filterChains, mainFilterChain)

	var listenerFilters []*listenerv3.ListenerFilter

	if listenerIR.ProxyProtocol {
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name:       "envoy.filters.listener.proxy_protocol",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{TypedConfig: mustAny(&proxyprotocolv3.ProxyProtocol{})},
		})
	}

	// TLS inspector needed for SNI-based filter chain matching
	if len(filterChains) > 1 {
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
		FilterChains:                  filterChains,
		ListenerFilters:               listenerFilters,
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(32768),
	}

	// Ensure authelia cluster is built (used by ext_authz for files/settings)
	if !clusterSet["authelia_backend"] {
		if c, ok := clusterMap["authelia_backend"]; ok {
			clusterSet["authelia_backend"] = true
			clusters = append(clusters, buildCluster(c))
		}
	}

	return listener, routeConfig, clusters
}

func translateVirtualHost(vhIR *ir.VirtualHostIR) *routev3.VirtualHost {
	vh := &routev3.VirtualHost{
		Name:    vhIR.Name,
		Domains: vhIR.Domains,
	}

	// Build sub_filter metadata for routes if pushstate or language is set
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

	// Response headers
	if vhIR.EnableOIDC && vhIR.UserZone != "" {
		vh.ResponseHeadersToAdd = append(vh.ResponseHeadersToAdd, &corev3.HeaderValueOption{
			Header:       &corev3.HeaderValue{Key: "Set-Cookie", Value: fmt.Sprintf("prev-host=%%REQ(:AUTHORITY)%%;Domain=.%s;Path=/;", vhIR.UserZone)},
			AppendAction: corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		})
	}

	// CORS response headers (matches nginx add_header on every location)
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

	// Disable ext_authz at VHost level; it is re-enabled per-route for files/settings.
	vh.TypedPerFilterConfig = map[string]*anypb.Any{
		"envoy.filters.http.ext_authz": mustAny(&extauthzv3.ExtAuthzPerRoute{
			Override: &extauthzv3.ExtAuthzPerRoute_Disabled{Disabled: true},
		}),
	}

	return vh
}

func translateRoute(routeIR *ir.HTTPRouteIR) *routev3.Route {
	route := &routev3.Route{
		Name: routeIR.Name,
	}

	// Route match
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

	// Direct response
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

	// Route to cluster
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

	// Request headers (sorted for deterministic output)
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

	// Per-route ext_authz config with authelia
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

func buildExtAuthzFilter(clusterMap map[string]*ir.ClusterIR, userName string) *hcmv3.HttpFilter {
	autheliaClusterName := "authelia_backend"
	if _, ok := clusterMap[autheliaClusterName]; !ok {
		return nil
	}

	extAuthz := &extauthzv3.ExtAuthz{
		Services: &extauthzv3.ExtAuthz_HttpService{
			HttpService: &extauthzv3.HttpService{
				ServerUri: &corev3.HttpUri{
					Uri: fmt.Sprintf("http://%s", autheliaClusterName),
					HttpUpstreamType: &corev3.HttpUri_Cluster{
						Cluster: autheliaClusterName,
					},
					Timeout: durationpb.New(15 * time.Second),
				},
				PathPrefix: "/api/authz/ext-authz/",
				AuthorizationRequest: &extauthzv3.AuthorizationRequest{
					AllowedHeaders: &matcherv3.ListStringMatcher{
						Patterns: []*matcherv3.StringMatcher{
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "host"}},
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "authorization"}},
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "proxy-authorization"}},
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "accept"}},
							{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "cookie"}},
						},
					},
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
		TransportApiVersion: corev3.ApiVersion_V3,
		FailureModeAllow:    false,
		ClearRouteCache:     false,
	}

	return &hcmv3.HttpFilter{
		Name:       "envoy.filters.http.ext_authz",
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: mustAny(extAuthz)},
	}
}

func buildCluster(c *ir.ClusterIR) *clusterv3.Cluster {
	discoveryType := clusterv3.Cluster_STRICT_DNS
	if !c.UseDNS {
		discoveryType = clusterv3.Cluster_STATIC
	}

	return &clusterv3.Cluster{
		Name: c.Name,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: discoveryType,
		},
		ConnectTimeout: durationpb.New(connectTimeout),
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

func buildTCPListener(listenerIR *ir.StreamListenerIR, irClusters []*ir.ClusterIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource

	if !clusterSet[listenerIR.Cluster] {
		for _, c := range irClusters {
			if c.Name == listenerIR.Cluster {
				clusterSet[c.Name] = true
				clusters = append(clusters, buildCluster(c))
				break
			}
		}
	}

	tcpProxy := &tcpproxyv3.TcpProxy{
		StatPrefix: listenerIR.Name,
		ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{
			Cluster: listenerIR.Cluster,
		},
		IdleTimeout: durationpb.New(tcpIdleTimeout),
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

func buildUDPListener(listenerIR *ir.StreamListenerIR, irClusters []*ir.ClusterIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource

	if !clusterSet[listenerIR.Cluster] {
		for _, c := range irClusters {
			if c.Name == listenerIR.Cluster {
				clusterSet[c.Name] = true
				clusters = append(clusters, buildCluster(c))
				break
			}
		}
	}

	udpProxy := &udpproxyv3.UdpProxyConfig{
		StatPrefix: listenerIR.Name,
		RouteSpecifier: &udpproxyv3.UdpProxyConfig_Cluster{
			Cluster: listenerIR.Cluster,
		},
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

func buildAccessLog() *accesslogv3.AccessLog {
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

// Suppress unused import warnings
var _ = matcherv3.StringMatcher{}
var _ = wrapperspb.Bool
