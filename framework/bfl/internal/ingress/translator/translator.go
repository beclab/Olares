package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bytetrade.io/web3os/bfl/internal/ingress/ir"
	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

const mapKey = "default"

var autheliaCluster = "authelia_backend"
var autheliaHost = "authelia-backend.os-framework.svc.cluster.local"
var autheliaPort uint32 = 9091
var autheliaPathPrefix = "/api/authz/ext-authz/"

type nonAppServerDef struct {
	Name      string
	SvcFormat string
}

var nonAppServers = []nonAppServerDef{
	{Name: "auth", SvcFormat: "authelia-svc.%s.svc.cluster.local"},
	{Name: "desktop", SvcFormat: "edge-desktop.%s.svc.cluster.local"},
	{Name: "wizard", SvcFormat: "wizard.%s.svc.cluster.local"},
}

var specialPatchApps = map[string]bool{
	"files":    true,
	"settings": true,
}

type Config struct {
	AutheliaURL string
}

type Translator struct {
	providerResources *message.ProviderResources
	xdsIR             *message.XdsIR
	cfg               *Config
}

func New(providerResources *message.ProviderResources, xdsIR *message.XdsIR, cfg *Config) *Translator {
	if cfg != nil && cfg.AutheliaURL != "" {
		if u, err := url.Parse(cfg.AutheliaURL); err == nil {
			host := u.Hostname()
			if host != "" {
				autheliaHost = host
			}
			if p := u.Port(); p != "" {
				if port, err := strconv.ParseUint(p, 10, 32); err == nil {
					autheliaPort = uint32(port)
				}
			}
			if u.Path != "" && u.Path != "/" {
				autheliaPathPrefix = u.Path
			}
			klog.Infof("translator: authelia configured: host=%s port=%d pathPrefix=%s", autheliaHost, autheliaPort, autheliaPathPrefix)
		} else {
			klog.Warningf("translator: failed to parse AutheliaURL %q: %v", cfg.AutheliaURL, err)
		}
	}

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
	klog.Infof("translator: published xdsIR with %d HTTP listeners, %d stream listeners, %d clusters",
		len(xds.HTTPListeners), len(xds.StreamListeners), len(xds.Clusters))
}

func (t *Translator) Translate(res *message.Resources) *ir.Xds {
	xds := &ir.Xds{}
	clusterSet := make(map[string]*ir.ClusterIR)

	// HTTP redirect listener (port 80 -> HTTPS)
	//xds.HTTPListeners = append(xds.HTTPListeners, t.buildHTTPRedirectListener())

	if res.SSL != nil && res.SSL.CertData != "" {
		// Main HTTPS listener (port 443)
		httpsListener := t.buildHTTPSListener(res, 443, false, clusterSet)
		xds.HTTPListeners = append(xds.HTTPListeners, httpsListener)

		// HTTPS listener with proxy_protocol (port 444, from l4-proxy)
		httpsProxyListener := t.buildHTTPSListener(res, 444, true, clusterSet)
		xds.HTTPListeners = append(xds.HTTPListeners, httpsProxyListener)

		// TLS secret
		xds.Secrets = append(xds.Secrets, &ir.SecretIR{
			Name:     "main-tls",
			CertData: res.SSL.CertData,
			KeyData:  res.SSL.KeyData,
		})

		// Custom domain TLS secrets
		for _, cert := range res.CustomDomainCerts {
			xds.Secrets = append(xds.Secrets, &ir.SecretIR{
				Name:     fmt.Sprintf("custom-tls-%s", cert.Domain),
				CertData: cert.CertData,
				KeyData:  cert.KeyData,
			})
		}
	}

	// Authelia cluster (for files/settings ext_authz)
	clusterSet[autheliaCluster] = &ir.ClusterIR{
		Name:   autheliaCluster,
		Host:   autheliaHost,
		Port:   autheliaPort,
		UseDNS: true,
	}

	// Stream listeners (TCP/UDP) for exposed app ports
	xds.StreamListeners = t.buildStreamListeners(res, clusterSet)

	// Collect all clusters
	for _, c := range clusterSet {
		xds.Clusters = append(xds.Clusters, c)
	}

	return xds
}

func (t *Translator) buildHTTPRedirectListener() *ir.HTTPListenerIR {
	return &ir.HTTPListenerIR{
		Name:       "http_redirect_80",
		Address:    "0.0.0.0",
		Port:       80,
		IsRedirect: true,
	}
}

func (t *Translator) buildHTTPSListener(res *message.Resources, port uint32, proxyProtocol bool, clusterSet map[string]*ir.ClusterIR) *ir.HTTPListenerIR {
	name := fmt.Sprintf("https_%d", port)
	listener := &ir.HTTPListenerIR{
		Name:          name,
		Address:       "0.0.0.0",
		Port:          port,
		TLS:           true,
		ProxyProtocol: proxyProtocol,
	}

	zone := res.SSL.Zone
	isEphemeral := res.IsEphemeralUser

	// Profile/root server (zone hostname -> profile service)
	profileCluster := "profile_service"
	profileHost := fmt.Sprintf("profile-service.user-space-%s.svc.cluster.local", res.UserName)
	clusterSet[profileCluster] = &ir.ClusterIR{
		Name: profileCluster, Host: profileHost, Port: 3000, UseDNS: true,
	}

	zoneAliases := []string{}
	zoneTokens := strings.Split(zone, ".")
	if len(zoneTokens) > 1 {
		zoneAliases = append(zoneAliases, fmt.Sprintf("%s.%s", zoneTokens[0], "olares.local"))
	}

	profileVHost := &ir.VirtualHostIR{
		Name:     "profile",
		Domains:  append([]string{zone}, zoneAliases...),
		Language: res.Language,
		UserZone: zone,
		UserName: res.UserName,
		Routes: []*ir.HTTPRouteIR{{
			Name:       "profile_root",
			PathPrefix: "/",
			Cluster:    profileCluster,
			RequestHeaders: map[string]string{
				"X-BFL-USER": res.UserName,
			},
		}},
	}
	listener.VirtualHosts = append(listener.VirtualHosts, profileVHost)

	// Application servers
	for _, app := range res.Apps {
		vhosts := t.buildAppVirtualHosts(res, app, zone, isEphemeral, clusterSet)
		listener.VirtualHosts = append(listener.VirtualHosts, vhosts...)
	}

	// Non-app servers (auth, desktop, wizard)
	for _, def := range nonAppServers {
		vhost := t.buildNonAppVirtualHost(res, def, zone, isEphemeral, clusterSet)
		listener.VirtualHosts = append(listener.VirtualHosts, vhost)
	}

	// Custom domain virtual hosts
	for _, app := range res.Apps {
		cvhosts := t.buildCustomDomainVirtualHosts(res, app, clusterSet)
		listener.VirtualHosts = append(listener.VirtualHosts, cvhosts...)
	}

	// Default response for unmatched hosts
	listener.DefaultResponse = &ir.DirectResponseIR{
		Status:      421,
		Body:        "<h1><a href='https://www.olares.com/'>Olares</a></h1>",
		ContentType: "text/html",
	}
	listener.PingRoute = true

	return listener
}

func (t *Translator) buildAppVirtualHosts(res *message.Resources, app *message.AppInfo, zone string, isEphemeral bool, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	var vhosts []*ir.VirtualHostIR

	var appDomainConfigs []defaultThirdLevelDomainConfig
	if raw, ok := app.Settings["defaultThirdLevelDomainConfig"]; ok && raw != "" {
		_ = json.Unmarshal([]byte(raw), &appDomainConfigs)
	}

	for index, entrance := range app.Entrances {
		if entrance.Host == "" {
			continue
		}

		prefix := getAppEntranceHostName(app.Entrances, index, app.Appid, appDomainConfigs)

		hostname := fmt.Sprintf("%s.%s", prefix, zone)
		if isEphemeral {
			hostname = fmt.Sprintf("%s-%s.%s", prefix, res.UserName, zone)
		}

		localHost := makeLocalHost(hostname)
		domains := []string{hostname}
		if localHost != hostname {
			domains = append(domains, localHost)
		}

		// Custom third-level domain aliases
		customDomainMap := getSettingsMap(app.Settings, constants.ApplicationCustomDomain)
		if entranceMap, ok := customDomainMap[entrance.Name]; ok {
			if thirdLevel := entranceMap[constants.ApplicationThirdLevelDomain]; thirdLevel != "" {
				extHostname := fmt.Sprintf("%s.%s", thirdLevel, zone)
				if isEphemeral {
					extHostname = fmt.Sprintf("%s-%s.%s", thirdLevel, res.UserName, zone)
				}
				domains = append(domains, extHostname, makeLocalHost(extHostname))
			}
		}

		clusterName := fmt.Sprintf("app_%s_%s", app.Name, entrance.Name)
		upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", entrance.Host, app.Namespace)
		clusterSet[clusterName] = &ir.ClusterIR{
			Name: clusterName, Host: upstreamHost, Port: uint32(entrance.Port), UseDNS: true,
		}

		_, enableOIDC := app.Settings["oidc.client.id"]

		vhost := &ir.VirtualHostIR{
			Name:                  fmt.Sprintf("app_%s_%s", app.Name, entrance.Name),
			Domains:               domains,
			EnableOIDC:            enableOIDC,
			EnableWindowPushState: entrance.WindowPushState,
			Language:              res.Language,
			UserZone:              zone,
			UserName:              res.UserName,
		}

		routes := []*ir.HTTPRouteIR{}

		// Apply fileserver patches for files/settings apps
		if specialPatchApps[prefix] {
			routes = append(routes, t.buildFileserverRoutes(res, app, clusterSet)...)
		}

		// Default route for this entrance
		routes = append(routes, &ir.HTTPRouteIR{
			Name:       fmt.Sprintf("default_%s_%s", app.Name, entrance.Name),
			PathPrefix: "/",
			Cluster:    clusterName,
			RequestHeaders: map[string]string{
				"X-BFL-USER": res.UserName,
			},
			WebSocketUpgrade: true,
		})

		vhost.Routes = routes
		vhosts = append(vhosts, vhost)
	}

	return vhosts
}

func (t *Translator) buildFileserverRoutes(res *message.Resources, app *message.AppInfo, clusterSet map[string]*ir.ClusterIR) []*ir.HTTPRouteIR {
	var routes []*ir.HTTPRouteIR

	nodeLocationPrefixes := []string{
		"/api/resources/cache/",
		"/api/preview/cache/",
		"/api/raw/cache/",
		"/api/tree/cache/",
		"/api/resources/external/",
		"/api/preview/external/",
		"/api/raw/external/",
		"/api/tree/external/",
		"/api/mount/",
		"/api/unmount/external/",
		"/api/smb_history/",
		"/upload/upload-link/",
		"/upload/file-uploaded-bytes/",
		"/api/paste/",
		"/videos/",
		"/api/md5/cache/",
		"/api/md5/external/",
		"/api/permission/cache/",
		"/api/permission/external/",
		"/api/task/",
	}

	masterLocationPrefixes := []string{
		"/api/resources/cache/",
		"/api/preview/cache/",
		"/api/resources/external/",
		"/api/preview/external/",
		"/api/paste/",
		"/api/task/",
		"/api/repos/",
		"/seafhttp/",
		"/api/resources/share/",
		"/api/preview/share/",
		"/api/tree/share/",
		"/api/raw/share/",
		"/api/resources/sync/",
		"/api/preview/sync/",
		"/api/tree/sync/",
		"/api/raw/sync/",
		"/api/md5/sync/",
		"/api/sync/account/info/",
		"/api/search/sync_search/",
	}

	nodeRegexSharePrefixes := []string{
		"/api/resources/share/",
		"/api/preview/share/",
		"/api/tree/share/",
		"/api/raw/share/",
	}

	for _, node := range res.FileserverNodes {
		proxyCluster := fmt.Sprintf("files_%s", node.NodeName)
		proxyHost := fmt.Sprintf("files-%s.user-system-%s.svc.cluster.local", node.NodeName, res.UserName)
		clusterSet[proxyCluster] = &ir.ClusterIR{
			Name: proxyCluster, Host: proxyHost, Port: 28080, UseDNS: true,
		}

		extAuthCfg := &ir.ExtAuthConfigIR{
			Cluster:    autheliaCluster,
			PathPrefix: autheliaPathPrefix,
			RequestHeaders: []string{
				"X-Original-URL", "X-Original-Method", "X-Forwarded-For",
				"X-BFL-USER", "X-Authorization", "Cookie",
			},
		}

		for _, pfx := range nodeLocationPrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:       fmt.Sprintf("files_node_%s_%s", node.NodeName, sanitizeName(pfx)),
				PathPrefix: fmt.Sprintf("%s%s/", pfx, node.NodeName),
				Cluster:    proxyCluster,
				RequestHeaders: map[string]string{
					"X-BFL-USER":       res.UserName,
					"X-Terminus-Node":  node.NodeName,
					"X-Provider-Proxy": proxyHost,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}

		// Regex share routes: /api/resources/share/{node}_ etc
		for _, pfx := range nodeRegexSharePrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:      fmt.Sprintf("files_node_%s_share_%s", node.NodeName, sanitizeName(pfx)),
				PathRegex: fmt.Sprintf("^%s%s_.*", pfx, node.NodeName),
				Cluster:   proxyCluster,
				RequestHeaders: map[string]string{
					"X-BFL-USER":       res.UserName,
					"X-Terminus-Node":  node.NodeName,
					"X-Provider-Proxy": proxyHost,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}

		// Master node gets additional routes
		if node.IsMaster {
			for _, pfx := range masterLocationPrefixes {
				routes = append(routes, &ir.HTTPRouteIR{
					Name:       fmt.Sprintf("files_master_%s", sanitizeName(pfx)),
					PathPrefix: pfx,
					Cluster:    proxyCluster,
					RequestHeaders: map[string]string{
						"X-BFL-USER":       res.UserName,
						"X-Terminus-Node":  node.NodeName,
						"X-Provider-Proxy": proxyHost,
					},
					ExtAuth:          extAuthCfg,
					WebSocketUpgrade: true,
				})
			}
		}
	}

	return routes
}

func (t *Translator) buildNonAppVirtualHost(res *message.Resources, def nonAppServerDef, zone string, isEphemeral bool, clusterSet map[string]*ir.ClusterIR) *ir.VirtualHostIR {
	hostname := fmt.Sprintf("%s.%s", def.Name, zone)
	if isEphemeral {
		hostname = fmt.Sprintf("%s-%s.%s", def.Name, res.UserName, zone)
	}

	clusterName := fmt.Sprintf("nonapp_%s", def.Name)
	upstreamHost := fmt.Sprintf(def.SvcFormat, constants.Namespace)
	clusterSet[clusterName] = &ir.ClusterIR{
		Name: clusterName, Host: upstreamHost, Port: 80, UseDNS: true,
	}

	return &ir.VirtualHostIR{
		Name:     fmt.Sprintf("nonapp_%s", def.Name),
		Domains:  []string{hostname, makeLocalHost(hostname)},
		Language: res.Language,
		UserZone: zone,
		UserName: res.UserName,
		Routes: []*ir.HTTPRouteIR{{
			Name:       fmt.Sprintf("nonapp_%s_root", def.Name),
			PathPrefix: "/",
			Cluster:    clusterName,
			RequestHeaders: map[string]string{
				"X-BFL-USER": res.UserName,
			},
			WebSocketUpgrade: true,
		}},
	}
}

func (t *Translator) buildCustomDomainVirtualHosts(res *message.Resources, app *message.AppInfo, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	var vhosts []*ir.VirtualHostIR

	customDomainMap := getSettingsMap(app.Settings, constants.ApplicationCustomDomain)
	if len(customDomainMap) == 0 {
		return nil
	}

	for _, entrance := range app.Entrances {
		if entrance.Host == "" {
			continue
		}

		entranceCustomDomain, ok := customDomainMap[entrance.Name]
		if !ok {
			continue
		}
		customDomainName := entranceCustomDomain[constants.ApplicationThirdPartyDomain]
		if customDomainName == "" {
			continue
		}

		clusterName := fmt.Sprintf("custom_%s_%s", app.Name, entrance.Name)
		upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", entrance.Host, app.Namespace)
		clusterSet[clusterName] = &ir.ClusterIR{
			Name: clusterName, Host: upstreamHost, Port: uint32(entrance.Port), UseDNS: true,
		}

		vhost := &ir.VirtualHostIR{
			Name:     fmt.Sprintf("custom_%s_%s", app.Name, entrance.Name),
			Domains:  []string{customDomainName},
			Language: res.Language,
			UserZone: res.UserZone,
			UserName: res.UserName,
			Routes: []*ir.HTTPRouteIR{{
				Name:       fmt.Sprintf("custom_%s_%s_root", app.Name, entrance.Name),
				PathPrefix: "/",
				Cluster:    clusterName,
				RequestHeaders: map[string]string{
					"X-BFL-USER": res.UserName,
				},
				WebSocketUpgrade: true,
			}},
		}
		vhosts = append(vhosts, vhost)
	}

	return vhosts
}

func (t *Translator) buildStreamListeners(res *message.Resources, clusterSet map[string]*ir.ClusterIR) []*ir.StreamListenerIR {
	var listeners []*ir.StreamListenerIR
	seen := make(map[string]bool)

	for _, app := range res.Apps {
		for _, port := range app.Ports {
			if port.Host == "" || port.ExposePort < 1 || port.ExposePort > 65535 {
				continue
			}

			proto := strings.ToLower(port.Protocol)
			if proto == "" {
				proto = "tcp"
			}

			portKey := fmt.Sprintf("%s:%d", proto, port.ExposePort)
			if seen[portKey] {
				continue
			}
			seen[portKey] = true

			clusterName := fmt.Sprintf("stream_%s_%d", proto, port.ExposePort)
			upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", port.Host, app.Namespace)
			clusterSet[clusterName] = &ir.ClusterIR{
				Name: clusterName, Host: upstreamHost, Port: uint32(port.Port), UseDNS: true,
			}

			listeners = append(listeners, &ir.StreamListenerIR{
				Name:     fmt.Sprintf("stream_%s_%d", proto, port.ExposePort),
				Address:  "0.0.0.0",
				Port:     uint32(port.ExposePort),
				Protocol: proto,
				Cluster:  clusterName,
			})
		}
	}

	return listeners
}

// Helper functions

type defaultThirdLevelDomainConfig struct {
	EntranceName     string `json:"entranceName"`
	ThirdLevelDomain string `json:"thirdLevelDomain"`
}

func getAppEntranceHostName(entrances []*message.EntranceInfo, index int, appid string, configs []defaultThirdLevelDomainConfig) string {
	if len(entrances) == 1 {
		return appid
	}
	for _, cfg := range configs {
		if cfg.EntranceName == entrances[index].Name && cfg.ThirdLevelDomain != "" {
			return cfg.ThirdLevelDomain
		}
	}
	return fmt.Sprintf("%s%d", appid, index)
}

func makeLocalHost(hostname string) string {
	tokens := strings.Split(hostname, ".")
	if len(tokens) < 2 {
		return hostname
	}
	return strings.Join([]string{tokens[0], tokens[1], "olares", "local"}, ".")
}

func getSettingsMap(settings map[string]string, key string) map[string]map[string]string {
	r := make(map[string]map[string]string)
	data, ok := settings[key]
	if !ok || data == "" {
		return r
	}
	_ = json.Unmarshal([]byte(data), &r)
	return r
}

func sanitizeName(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.TrimPrefix(s, "_")
	s = strings.TrimSuffix(s, "_")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}
