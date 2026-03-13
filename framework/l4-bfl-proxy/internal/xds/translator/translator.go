package translator

import (
	"context"
	"fmt"
	"net"
	"time"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	accesslogfilev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	routerv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/proxy_protocol/v3"
	tlsinspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpproxyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	ppupstreamv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	rawtransportv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/telepresenceio/watchable"
	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"
)

var (
	tcpIdleTimeout        = time.Hour
	httpStreamIdleTimeout = 30 * time.Minute
	connectTimeout        = 5 * time.Second
)

func SetTimeouts(tcpIdle, httpStream, connect time.Duration) {
	tcpIdleTimeout = tcpIdle
	httpStreamIdleTimeout = httpStream
	connectTimeout = connect
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
	listeners, clusters := t.Translate(xdsIR)
	newSnapshot := &message.XdsSnapshot{
		Listeners: listeners,
		Clusters:  clusters,
	}

	if old, ok := t.xdsResources.Load(key); ok && old.Equal(newSnapshot) {
		klog.V(4).Infof("xds-translator: xDS unchanged for key %s, skipping", key)
		return
	}

	t.xdsResources.Store(key, newSnapshot)
	klog.Infof("xds-translator: published %d listeners, %d clusters", len(listeners), len(clusters))
}

func (t *XdsTranslator) Translate(xdsIR *ir.Xds) ([]cachetypes.Resource, []cachetypes.Resource) {
	var listeners []cachetypes.Resource
	var clusters []cachetypes.Resource
	clusterSet := make(map[string]bool)

	for _, listenerIR := range xdsIR.Listeners {
		switch listenerIR.Protocol {
		case ir.ProtocolHTTP:
			l := buildHTTPRedirectListener(listenerIR)
			if l != nil {
				listeners = append(listeners, l)
			}
		case ir.ProtocolTLS:
			l, cls := buildTLSListener(listenerIR, clusterSet)
			if l != nil {
				listeners = append(listeners, l)
			}
			clusters = append(clusters, cls...)
		case ir.ProtocolTCP:
			l, cls := buildTCPListener(listenerIR, clusterSet)
			if l != nil {
				listeners = append(listeners, l)
			}
			clusters = append(clusters, cls...)
		case ir.ProtocolUDP:
			l, cls := buildUDPListener(listenerIR, clusterSet)
			if l != nil {
				listeners = append(listeners, l)
			}
			clusters = append(clusters, cls...)
		}
	}
	return listeners, clusters
}

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
				Match: &routev3.RouteMatch{
					PathSpecifier: &routev3.RouteMatch_Prefix{Prefix: "/"},
				},
				Action: &routev3.Route_Redirect{
					Redirect: &routev3.RedirectAction{
						SchemeRewriteSpecifier: &routev3.RedirectAction_HttpsRedirect{
							HttpsRedirect: true,
						},
						ResponseCode: routev3.RedirectAction_MOVED_PERMANENTLY,
					},
				},
			}},
		}},
	}

	routerAny := mustAny(&routerv3.Router{})

	hcm := &hcmv3.HttpConnectionManager{
		StatPrefix: listenerIR.Name,
		RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		},
		StreamIdleTimeout: durationpb.New(httpStreamIdleTimeout),
		HttpFilters: []*hcmv3.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: routerAny,
			},
		}},
		InternalAddressConfig: &hcmv3.HttpConnectionManager_InternalAddressConfig{},
	}

	hcmAny := mustAny(hcm)

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerIR.Port,
					},
				},
			},
		},
		FilterChains: []*listenerv3.FilterChain{{
			Filters: []*listenerv3.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: hcmAny,
				},
			}},
		}},
	}
}

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
			clusters = append(clusters, buildCluster(route.Destination, route.ProxyProtocolUpstream))
		}

		fc := buildFilterChain(route, clusterName)
		filterChains = append(filterChains, fc)
	}

	var listenerFilters []*listenerv3.ListenerFilter

	if listenerIR.ProxyProtocol {
		ppConfig := &proxyprotocolv3.ProxyProtocol{}
		ppAny := mustAny(ppConfig)
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name: "envoy.filters.listener.proxy_protocol",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{
				TypedConfig: ppAny,
			},
		})
	}

	if listenerIR.TLSInspector {
		tlsConfig := &tlsinspectorv3.TlsInspector{}
		tlsAny := mustAny(tlsConfig)
		listenerFilters = append(listenerFilters, &listenerv3.ListenerFilter{
			Name: "envoy.filters.listener.tls_inspector",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{
				TypedConfig: tlsAny,
			},
		})
	}

	listener := &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerIR.Port,
					},
				},
			},
		},
		FilterChains:    filterChains,
		ListenerFilters: listenerFilters,
	}

	return listener, clusters
}

func buildFilterChain(route *ir.RouteIR, clusterName string) *listenerv3.FilterChain {
	tcpProxy := &tcpproxyv3.TcpProxy{
		StatPrefix: route.Name,
		ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		IdleTimeout: durationpb.New(tcpIdleTimeout),
		AccessLog:   []*accesslogv3.AccessLog{buildAccessLog()},
	}
	tcpProxyAny := mustAny(tcpProxy)

	fc := &listenerv3.FilterChain{
		Name: route.Name,
		Filters: []*listenerv3.Filter{{
			Name: wellknown.TCPProxy,
			ConfigType: &listenerv3.Filter_TypedConfig{
				TypedConfig: tcpProxyAny,
			},
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

func buildTCPListener(listenerIR *ir.ListenerIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var filterChains []*listenerv3.FilterChain
	var clusters []cachetypes.Resource

	for _, route := range listenerIR.Routes {
		if route.Destination == nil {
			continue
		}
		clusterName := route.Destination.Name
		if !clusterSet[clusterName] {
			clusterSet[clusterName] = true
			clusters = append(clusters, buildCluster(route.Destination, false))
		}

		tcpProxy := &tcpproxyv3.TcpProxy{
			StatPrefix: route.Name,
			ClusterSpecifier: &tcpproxyv3.TcpProxy_Cluster{
				Cluster: clusterName,
			},
			IdleTimeout: durationpb.New(tcpIdleTimeout),
		}
		tcpProxyAny := mustAny(tcpProxy)

		filterChains = append(filterChains, &listenerv3.FilterChain{
			Name: route.Name,
			Filters: []*listenerv3.Filter{{
				Name: wellknown.TCPProxy,
				ConfigType: &listenerv3.Filter_TypedConfig{
					TypedConfig: tcpProxyAny,
				},
			}},
		})
	}

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: listenerIR.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerIR.Port,
					},
				},
			},
		},
		FilterChains: filterChains,
	}, clusters
}

func buildUDPListener(listenerIR *ir.ListenerIR, clusterSet map[string]bool) (*listenerv3.Listener, []cachetypes.Resource) {
	var clusters []cachetypes.Resource

	if len(listenerIR.Routes) == 0 || listenerIR.Routes[0].Destination == nil {
		return nil, nil
	}
	route := listenerIR.Routes[0]
	clusterName := route.Destination.Name

	if !clusterSet[clusterName] {
		clusterSet[clusterName] = true
		clusters = append(clusters, buildCluster(route.Destination, false))
	}

	udpProxy := &udpproxyv3.UdpProxyConfig{
		StatPrefix: route.Name,
		RouteSpecifier: &udpproxyv3.UdpProxyConfig_Cluster{
			Cluster: clusterName,
		},
	}
	udpProxyAny := mustAny(udpProxy)

	return &listenerv3.Listener{
		Name: listenerIR.Name,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address:  listenerIR.Address,
					Protocol: corev3.SocketAddress_UDP,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerIR.Port,
					},
				},
			},
		},
		ListenerFilters: []*listenerv3.ListenerFilter{{
			Name: "envoy.filters.udp_listener.udp_proxy",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{
				TypedConfig: udpProxyAny,
			},
		}},
		UdpListenerConfig: &listenerv3.UdpListenerConfig{},
	}, clusters
}

func buildCluster(dest *ir.DestinationIR, proxyProtocolUpstream bool) *clusterv3.Cluster {
	cluster := &clusterv3.Cluster{
		Name: dest.Name,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_STRICT_DNS,
		},
		ConnectTimeout: durationpb.New(connectTimeout),
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: dest.Name,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address: dest.Host,
										PortSpecifier: &corev3.SocketAddress_PortValue{
											PortValue: dest.Port,
										},
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
		rawBuf := &rawtransportv3.RawBuffer{}
		rawBufAny := mustAny(rawBuf)

		ppUpstream := &ppupstreamv3.ProxyProtocolUpstreamTransport{
			Config: &corev3.ProxyProtocolConfig{
				Version: corev3.ProxyProtocolConfig_V1,
			},
			TransportSocket: &corev3.TransportSocket{
				Name: "envoy.transport_sockets.raw_buffer",
				ConfigType: &corev3.TransportSocket_TypedConfig{
					TypedConfig: rawBufAny,
				},
			},
		}
		ppUpstreamAny := mustAny(ppUpstream)

		cluster.TransportSocket = &corev3.TransportSocket{
			Name: "envoy.transport_sockets.upstream_proxy_protocol",
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: ppUpstreamAny,
			},
		}
	}

	return cluster
}

func buildAccessLog() *accesslogv3.AccessLog {
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
	fileLogAny := mustAny(fileLog)
	return &accesslogv3.AccessLog{
		Name: "envoy.access_loggers.file",
		ConfigType: &accesslogv3.AccessLog_TypedConfig{
			TypedConfig: fileLogAny,
		},
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
