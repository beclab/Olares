package translator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

const (
	mapKey  = "default"
	vpnCIDR = "100.64.0.0/24"
)

type Config struct {
	SSLServerPort       int
	SSLProxyServerPort  int
	UserNamespacePrefix string
}

type Translator struct {
	providerResources *message.ProviderResources
	xdsIR             *message.XdsIR
	cfg               *Config
}

func New(providerResources *message.ProviderResources, xdsIR *message.XdsIR, cfg *Config) *Translator {
	return &Translator{
		providerResources: providerResources,
		xdsIR:             xdsIR,
		cfg:               cfg,
	}
}

func (t *Translator) Name() string { return "translator" }

func (t *Translator) Start(ctx context.Context) error {
	klog.Info("translator: starting")
	subscription := t.providerResources.Subscribe(ctx)
	t.process(subscription)
	return nil
}

func (t *Translator) process(subscription <-chan watchable.Snapshot[string, *message.Resources]) {
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
				t.xdsIR.Delete(update.Key)
				continue
			}
			if update.Value != nil {
				t.handleUpdate(update.Key, update.Value)
			}
		}
	}
	klog.Info("translator: subscription closed")
}

func (t *Translator) handleUpdate(key string, resources *message.Resources) {
	xds := t.Translate(resources)

	if old, ok := t.xdsIR.Load(key); ok && old.Equal(xds) {
		klog.V(4).Infof("translator: xdsIR unchanged for key %s, skipping", key)
		return
	}

	t.xdsIR.Store(key, xds)
	klog.Infof("translator: published xdsIR with %d listeners", len(xds.Listeners))
}

func (t *Translator) Translate(resources *message.Resources) *ir.Xds {
	xds := &ir.Xds{}

	xds.Listeners = append(xds.Listeners, t.buildHTTPRedirectListener())
	xds.Listeners = append(xds.Listeners, t.buildTLSListener(resources, uint32(t.cfg.SSLServerPort), false))
	xds.Listeners = append(xds.Listeners, t.buildTLSListener(resources, uint32(t.cfg.SSLProxyServerPort), true))
	xds.Listeners = append(xds.Listeners, t.buildStreamListeners(resources)...)

	return xds
}

func (t *Translator) buildHTTPRedirectListener() *ir.ListenerIR {
	return &ir.ListenerIR{
		Name:     "http_redirect_81",
		Address:  "0.0.0.0",
		Port:     81,
		Protocol: ir.ProtocolHTTP,
		HTTPRedirect: &ir.HTTPRedirectIR{
			Scheme: "https",
			Code:   301,
		},
	}
}

func (t *Translator) buildTLSListener(resources *message.Resources, port uint32, proxyProtocol bool) *ir.ListenerIR {
	name := fmt.Sprintf("tls_%d", port)
	listener := &ir.ListenerIR{
		Name:          name,
		Address:       "0.0.0.0",
		Port:          port,
		Protocol:      ir.ProtocolTLS,
		ProxyProtocol: proxyProtocol,
		TLSInspector:  true,
	}

	allAppIDs := collectAppIDs(resources.Apps)

	for _, user := range resources.Users {
		routes := t.buildUserRoutes(user, allAppIDs)
		listener.Routes = append(listener.Routes, routes...)
	}

	return listener
}

// buildUserRoutes creates filter chain routes for a single user.
// If deny_all is set, two filter chains are created:
//  1. allowed_domains -> any source
//  2. all domains -> only allowed source CIDRs (local IP, VPN)
//
// Otherwise a single filter chain matching all user domains.
func (t *Translator) buildUserRoutes(user *message.UserInfo, allAppIDs []string) []*ir.RouteIR {
	clusterName := fmt.Sprintf("user_%s", user.Name)
	dest := &ir.DestinationIR{
		Name: clusterName,
		Host: user.BFLHost,
		Port: uint32(user.BFLPort),
	}

	sniMatches := buildSNIMatches(user, allAppIDs)

	if !user.DenyAll {
		return []*ir.RouteIR{{
			Name:                  fmt.Sprintf("route_%s", user.Name),
			SNIMatches:            sniMatches,
			Destination:           dest,
			ProxyProtocolUpstream: true,
		}}
	}

	var routes []*ir.RouteIR

	if len(user.AllowedDomains) > 0 {
		routes = append(routes, &ir.RouteIR{
			Name:                  fmt.Sprintf("route_%s_allowed", user.Name),
			SNIMatches:            user.AllowedDomains,
			Destination:           dest,
			ProxyProtocolUpstream: true,
		})
	}

	allowedCIDRs := []string{vpnCIDR}
	if user.LocalDomainIP != "" {
		allowedCIDRs = append(allowedCIDRs, user.LocalDomainIP+"/32")
	}
	routes = append(routes, &ir.RouteIR{
		Name:                  fmt.Sprintf("route_%s_restricted", user.Name),
		SNIMatches:            sniMatches,
		SourcePrefixRanges:    allowedCIDRs,
		Destination:           dest,
		ProxyProtocolUpstream: true,
	})

	return routes
}

// buildSNIMatches generates all SNI patterns for a user.
// Admin users use wildcard suffix matching (e.g. "*.zone").
// Ephemeral users enumerate exact "{appid}-{username}.{zone}" for all known appids.
func buildSNIMatches(user *message.UserInfo, allAppIDs []string) []string {
	if user.IsEphemeral {
		return buildEphemeralSNI(user)
	}

	var snis []string
	for _, domain := range user.ServerNameDomains {
		snis = append(snis, domain)
		snis = append(snis, "*."+domain)
	}
	return snis
}

func buildEphemeralSNI(user *message.UserInfo) []string {
	var snis []string
	seen := make(map[string]struct{})
	sni := fmt.Sprintf("wizard-%s.%s", user.Name, user.Zone)
	if _, ok := seen[sni]; !ok {
		snis = append(snis, sni)
		seen[sni] = struct{}{}
	}
	local := WithLocal("local")(user.Zone)
	sni = fmt.Sprintf("wizard-%s.%s", user.Name, local)
	if _, ok := seen[sni]; !ok && local != user.Zone {
		snis = append(snis, sni)
		seen[sni] = struct{}{}
	}
	return snis
}

func WithLocal(local string) func(string) string {
	return func(domain string) string {
		lastDot := strings.LastIndex(domain, ".")
		if lastDot == -1 {
			return domain
		}
		return domain[:lastDot+1] + local
	}
}

func (t *Translator) buildStreamListeners(resources *message.Resources) []*ir.ListenerIR {
	bflHostMap := make(map[string]string)
	for _, u := range resources.Users {
		bflHostMap[u.Name] = u.BFLHost
	}

	// seenPorts tracks "proto:port" keys to skip duplicate expose ports across apps.
	seenPorts := make(map[string]bool)
	var listeners []*ir.ListenerIR

	for _, app := range resources.Apps {
		for _, port := range app.Ports {
			if port.ExposePort < 1 || port.ExposePort > 65535 {
				continue
			}
			bflHost := bflHostMap[app.Owner]
			if bflHost == "" {
				klog.Warningf("translator: no BFL host for app owner %q", app.Owner)
				continue
			}

			proto := strings.ToLower(port.Protocol)
			if proto == "" {
				proto = "tcp"
			}
			portKey := fmt.Sprintf("%s:%d", proto, port.ExposePort)
			if seenPorts[portKey] {
				klog.Warningf("translator: duplicate expose port %s for app %q, skipping", portKey, app.Name)
				continue
			}
			seenPorts[portKey] = true

			protocol := ir.ProtocolTCP
			if proto == "udp" {
				protocol = ir.ProtocolUDP
			}

			listener := &ir.ListenerIR{
				Name:     fmt.Sprintf("stream_%s_%d", proto, port.ExposePort),
				Address:  "0.0.0.0",
				Port:     uint32(port.ExposePort),
				Protocol: protocol,
				Routes: []*ir.RouteIR{{
					Name: fmt.Sprintf("direct_%d", port.ExposePort),
					Destination: &ir.DestinationIR{
						Name: fmt.Sprintf("direct_%s_%d", app.Owner, port.ExposePort),
						Host: bflHost,
						Port: uint32(port.ExposePort),
					},
				}},
			}
			listeners = append(listeners, listener)
		}
	}
	return listeners
}

func collectAppIDs(apps []*message.AppInfo) []string {
	seen := make(map[string]struct{})
	var ids []string
	for _, app := range apps {
		entranceCount := len(app.Entrances)
		for i := range app.Entrances {
			var prefix string
			if entranceCount == 1 {
				prefix = app.Appid
			} else {
				prefix = fmt.Sprintf("%s%d", app.Appid, i)
			}
			if _, ok := seen[prefix]; !ok {
				seen[prefix] = struct{}{}
				ids = append(ids, prefix)
			}
		}
	}
	sort.Strings(ids)
	return ids
}
