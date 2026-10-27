package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

const (
	vpnCIDR                  = "100.64.0.0/16"
	autheliaClusterPrefix    = "authelia_backend"
	autheliaBackendHost      = "authelia-backend.os-framework.svc.cluster.local"
	autheliaPort             = uint32(9091)
	autheliaVerifyPathPrefix = "/api/verify/"

	// probeUARegex matches webhook.getProbeUA shape: {uuid}/{md5hex}.
	probeUARegex = `^[0-9a-fA-F-]+/[0-9a-fA-F]+$`

	settingsCustomDomain                 = "customDomain"
	settingsCustomDomainThirdLevelDomain = "third_level_domain"
	settingsCustomDomainThirdPartyDomain = "third_party_domain"

	// entranceTypeDev marks a development entrance, which gets an extra
	// `<appid>-<port>.<zone>` alias routed to the same upstream.
	entranceTypeDev = "dev"

	// systemServicePriority makes shared system-service virtual hosts outrank
	// apps and custom domains when a domain is claimed by more
	// than one virtual host in the same route config.
	systemServicePriority = 100
)

var nodeLocationPrefixes = []string{
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
	"/api/archive/",
}

var masterLocationPrefixes = []string{
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

var nodeRegexSharePrefixes = []string{
	"/api/resources/share/",
	"/api/preview/share/",
	"/api/tree/share/",
	"/api/raw/share/",
}

type systemServiceDef struct {
	Name      string
	SvcFormat string
	// Namespace contains the shared frontend service.
	Namespace string
}

var systemServices = []systemServiceDef{
	{Name: "auth", SvcFormat: "authelia-svc.%s.svc.cluster.local", Namespace: "os-frontend"},
	{Name: "desktop", SvcFormat: "edge-desktop.%s.svc.cluster.local", Namespace: "os-frontend"},
}

var fileserverPatchApps = map[string]bool{
	"files":    true,
	"settings": true,
	"share":    true,
}

type Config struct {
	SSLServerPort      int
	SSLProxyServerPort int
	LocalHTTPProxyPort int
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
	klog.Info("translator: starting...")
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

func (t *Translator) Translate(resources *message.Resources) *ir.Xds {
	xds := &ir.Xds{}
	if resources == nil {
		resources = &message.Resources{}
	}
	resources.Sort()
	clusterSet := make(map[string]*ir.ClusterIR)
	xds.Listeners = append(xds.Listeners, t.buildHTTPRedirectListener())
	clusterSet[autheliaClusterPrefix] = &ir.ClusterIR{Name: autheliaClusterPrefix, Host: autheliaBackendHost, Port: autheliaPort, UseDNS: true}
	vhosts, customVhosts := t.machineVirtualHosts(resources, clusterSet)
	if local := t.buildSharedLocalHTTPListener(filterLocalDomains(vhosts)); local != nil {
		local.ExtAuthCluster = autheliaClusterPrefix
		localZone := resources.LocalZone
		if localZone == "" {
			localZone = "olares.local"
		}
		for _, host := range local.VirtualHosts {
			host.UserZone = localZone
		}
		xds.HTTPListeners = append(xds.HTTPListeners, local)
	}
	ports := []struct {
		port  uint32
		proxy bool
	}{{uint32(t.cfg.SSLServerPort), false}, {uint32(t.cfg.SSLProxyServerPort), true}}
	if resources.DefaultDomain != "" && resources.SSL != nil && resources.SSL.CertData != "" {
		cert := &ir.SecretIR{Name: "default-domain-tls", CertData: resources.SSL.CertData, KeyData: resources.SSL.KeyData}
		xds.Secrets = append(xds.Secrets, cert)
		for _, port := range ports {
			xds.HTTPListeners = append(xds.HTTPListeners, t.machineTLSListener(fmt.Sprintf("https_%d_default", port.port), port.port, port.proxy, vhosts, cert, []string{resources.DefaultDomain, "*." + resources.DefaultDomain}))
		}
	}
	for _, user := range resources.Users {
		for _, cert := range user.CustomDomainCerts {
			secret := &ir.SecretIR{Name: fmt.Sprintf("custom-tls-%s-%s", user.Name, cert.Domain), CertData: cert.CertData, KeyData: cert.KeyData}
			xds.Secrets = append(xds.Secrets, secret)
			// Keep the custom domain's own user context and the shared system
			// reservations, without inheriting another user's custom routes.
			custom := uniqueMachineVirtualHosts(append(append([]*ir.VirtualHostIR{}, vhosts...), customVhosts[user.Name]...))
			for _, port := range ports {
				listener := t.machineTLSListener(fmt.Sprintf("https_%d_%s_custom_%s", port.port, user.Name, sanitizeName(cert.Domain)), port.port, port.proxy, custom, secret, []string{cert.Domain})
				listener.CreatedAt = cert.CreatedAt
				xds.HTTPListeners = append(xds.HTTPListeners, listener)
			}
		}
	}
	xds.StreamListeners = t.buildStreamListeners(resources, clusterSet)
	for _, cluster := range clusterSet {
		xds.Clusters = append(xds.Clusters, cluster)
	}
	sort.Slice(xds.Clusters, func(i, j int) bool { return xds.Clusters[i].Name < xds.Clusters[j].Name })
	return xds
}

// buildHTTPRedirectListener creates the HTTP->HTTPS redirect on port 81.
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

// applyDenyAllRestrictions stamps SourceCIDRs onto restricted VHs so the xDS
// translator can generate per-VH RBAC rules in the RDS route config.
// extraVHosts are ephemeral wizard VHs that should always be publicly accessible.
func (t *Translator) applyDenyAllRestrictions(user *message.UserInfo, vhosts []*ir.VirtualHostIR, extraVHosts []*ir.VirtualHostIR, localDomainIP string) {
	if !user.DenyAll {
		return
	}

	allowedSet := make(map[string]bool, len(user.AllowedDomains))
	for _, d := range user.AllowedDomains {
		allowedSet[d] = true
	}
	ephemeralDomainSet := make(map[string]bool)
	for _, vh := range extraVHosts {
		for _, d := range vh.Domains {
			ephemeralDomainSet[d] = true
		}
	}

	restrictCIDRs := []string{vpnCIDR}
	if ip := net.ParseIP(localDomainIP); ip != nil {
		bits := 128
		if ip.To4() != nil {
			bits = 32
		}
		restrictCIDRs = append(restrictCIDRs, fmt.Sprintf("%s/%d", ip.String(), bits))
	}
	if user.MasterNodeCIDR != "" {
		restrictCIDRs = append(restrictCIDRs, user.MasterNodeCIDR)
	}
	//restrictCIDRs = append(restrictCIDRs, user.AllowCIDRs...)

	for _, vh := range vhosts {
		isAllowed := false
		for _, d := range vh.Domains {
			if allowedSet[d] || ephemeralDomainSet[d] {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			vh.SourceCIDRs = restrictCIDRs
		}
	}
}

func (t *Translator) localHTTPPort() uint32 {
	if t.cfg != nil && t.cfg.LocalHTTPProxyPort > 0 {
		return uint32(t.cfg.LocalHTTPProxyPort)
	}
	return 82
}

// buildSharedLocalHTTPListener is the single plain-HTTP listener on port 82.
// Every user's .olares.local virtual host shares one filter chain. Host
// matching is done by the route config. Each ExtAuth route carries its own
// Authelia cluster. This listener has no user of its own, so the chain does
// not point at anyone's Authelia. Public zone names stay on 443/444.
func (t *Translator) buildSharedLocalHTTPListener(vhosts []*ir.VirtualHostIR) *ir.HTTPListenerIR {
	localVhosts := scopeLocalExtAuth(vhosts)
	if len(localVhosts) == 0 {
		return nil
	}
	port := t.localHTTPPort()
	return &ir.HTTPListenerIR{
		Name:          fmt.Sprintf("http_%d", port),
		Address:       "0.0.0.0",
		Port:          port,
		ProxyProtocol: true,
		VirtualHosts:  localVhosts,
		DefaultResponse: &ir.DirectResponseIR{
			Status:      421,
			Body:        "<h1><a href='https://www.olares.com/'>Olares</a></h1>",
			ContentType: "text/html",
		},
		PingRoute: true,
	}
}

// scopeLocalExtAuth deep-copies virtual hosts and marks each ExtAuth route so
// the xDS layer emits that user's Authelia cluster. The copy keeps the HTTPS
// chains on 443/444 sharing the filter-level Authelia default.
func scopeLocalExtAuth(vhosts []*ir.VirtualHostIR) []*ir.VirtualHostIR {
	out := make([]*ir.VirtualHostIR, 0, len(vhosts))
	for _, vh := range vhosts {
		if vh == nil {
			continue
		}
		cp := vh.DeepCopy()
		for _, route := range cp.Routes {
			if route.ExtAuth == nil || route.ExtAuth.Disabled {
				continue
			}
			route.ExtAuth.ServiceOverride = true
			if route.ExtAuth.User == "" {
				route.ExtAuth.User = cp.UserName
			}
		}
		out = append(out, cp)
	}
	return out
}

func filterLocalDomains(vhosts []*ir.VirtualHostIR) []*ir.VirtualHostIR {
	var out []*ir.VirtualHostIR
	for _, vh := range vhosts {
		if vh == nil {
			continue
		}
		var domains []string
		for _, d := range vh.Domains {
			if strings.HasSuffix(d, ".olares.local") {
				domains = append(domains, d)
			}
		}
		if len(domains) == 0 {
			continue
		}
		cp := *vh
		cp.Domains = domains
		out = append(out, &cp)
	}
	return out
}

func (t *Translator) buildAppVirtualHosts(user *message.UserInfo, app *message.AppInfo, zone string, clusterSet map[string]*ir.ClusterIR, localZones ...string) []*ir.VirtualHostIR {
	localZone := "olares.local"
	if len(localZones) > 0 && localZones[0] != "" {
		localZone = localZones[0]
	}
	var vhosts []*ir.VirtualHostIR

	var appDomainConfigs []appv1alpha1.DefaultThirdLevelDomainConfig
	if raw, ok := app.Settings["defaultThirdLevelDomainConfig"]; ok && raw != "" {
		err := json.Unmarshal([]byte(raw), &appDomainConfigs)
		if err != nil {
			klog.Warningf("buildAppVirtualHosts: defaultThirdLevelDomainConfig unmarshal error: %v", err)
		}
	}

	customDomainMap := parseSettingsJSON(app.Settings, settingsCustomDomain)

	for index, entrance := range app.Entrances {
		if entrance.Host == "" {
			klog.Warningf("buildAppVirtualHosts: app: %s,entrance: %s with empty host, skipping...", app.Name, entrance.Name)
			continue
		}

		prefix := appv1alpha1.ResolveEntranceIDWithDefaultThirdLevelDomainOverride(app.Entrances, index, app.Appid, appDomainConfigs)

		hostname := fmt.Sprintf("%s.%s", prefix, zone)

		localHost := prefix + "." + localZone
		domains := []string{hostname}
		if localHost != hostname {
			domains = append(domains, localHost)
		}

		if entranceMap, ok := customDomainMap[entrance.Name]; ok {
			if thirdLevel := entranceMap[settingsCustomDomainThirdLevelDomain]; thirdLevel != "" {
				extHostname := fmt.Sprintf("%s.%s", thirdLevel, zone)
				domains = append(domains, extHostname, thirdLevel+"."+localZone)
			}
		}

		// A "dev" entrance additionally answers on `<appid>-<port>.<zone>`
		// (plus its .olares.local alias), routed to the same upstream as the
		// entrance's primary hostname.
		if entrance.Type == entranceTypeDev {
			devPrefix := fmt.Sprintf("%s-%d", app.Appid, entrance.Port)
			devHostname := fmt.Sprintf("%s.%s", devPrefix, zone)

			domains = append(domains, devHostname)
			if devLocal := devPrefix + "." + localZone; devLocal != devHostname {
				domains = append(domains, devLocal)
			}
		}

		clusterName := fmt.Sprintf("app_%s_%s_%s", user.Name, app.Name, entrance.Name)
		cluster := &ir.ClusterIR{
			Name:   clusterName,
			Host:   fmt.Sprintf("%s.%s.svc.cluster.local", entrance.Host, app.Namespace),
			Port:   uint32(entrance.Port),
			UseDNS: true,
		}
		// A "dev" entrance's Host is already the backing pod IP (set by
		// proxylistener), so route straight to it as a STATIC cluster rather
		// than a Kubernetes service DNS name. Dev apps often have no
		// stable/ready service endpoint, which makes a STRICT_DNS cluster
		// resolve to nothing and return 503 no_healthy_upstream; hitting the
		// pod IP directly avoids that.
		if entrance.Type == entranceTypeDev {
			cluster.Host = entrance.Host
			cluster.UseDNS = false
		}
		clusterSet[clusterName] = cluster

		_, enableOIDC := app.Settings["oidc.client.id"]

		vhost := &ir.VirtualHostIR{
			Name:                  fmt.Sprintf("app_%s_%s_%s", user.Name, app.Name, entrance.Name),
			Domains:               domains,
			EnableOIDC:            enableOIDC,
			EnableWindowPushState: entrance.WindowPushState,
			Language:              user.Language,
			UserZone:              zone,
			UserName:              user.Name,
			// Files, Settings and Share file routes need Envoy-level CORS
			IsFileserver: fileserverPatchApps[prefix],
			Priority:     systemAppPriority(app.Name),
		}

		var routes []*ir.HTTPRouteIR

		if prefix == "share" {
			routes = append(routes, t.buildShareFileserverRoutes(user, clusterSet)...)
		} else if fileserverPatchApps[prefix] {
			routes = append(routes, t.buildFileserverRoutes(user, clusterSet)...)
		}

		if user.DenyAll &&
			app.IsSysApp &&
			app.Appid == "olares-app" &&
			entrance.Name == "headscale" {
			publicRoute := func(name, path, method, target string, upgrade bool) *ir.HTTPRouteIR {
				return &ir.HTTPRouteIR{
					Name:             fmt.Sprintf("headscale_%s_%s", user.Name, name),
					PathExact:        path,
					Cluster:          target,
					HeaderMatches:    []ir.HeaderMatchIR{{Name: ":method", Exact: method}},
					ExtAuth:          buildAppExtAuthConfig(),
					WebSocketUpgrade: upgrade,
					BypassSourceCIDR: true,
				}
			}
			routes = append(routes,
				publicRoute("key_get", "/key", "GET", clusterName, false),
				publicRoute("ts2021_get", "/ts2021", "GET", clusterName, true),
				publicRoute("ts2021_post", "/ts2021", "POST", clusterName, true),
				publicRoute("refresh_post", "/api/refresh", "POST", clusterName, false),
				publicRoute("refresh_options", "/api/refresh", "OPTIONS", clusterName, false),
			)
		}

		defaultRoute := &ir.HTTPRouteIR{
			Name:             fmt.Sprintf("default_%s_%s_%s", user.Name, app.Name, entrance.Name),
			PathPrefix:       "/",
			Cluster:          clusterName,
			WebSocketUpgrade: true,
		}

		defaultRoute.ExtAuth = buildAppExtAuthConfig()

		// F3: Exact probe paths with signed UA bypass ExtAuth (oes parity).
		// Emit before the catch-all so Envoy matches them first.
		routes = append(routes, buildProbeBypassRoutes(
			fmt.Sprintf("%s_%s_%s", user.Name, app.Name, entrance.Name),
			clusterName,
			app.EntranceProbePaths[entrance.Name],
		)...)
		routes = append(routes, defaultRoute)

		vhost.Routes = routes
		vhosts = append(vhosts, vhost)
	}

	return vhosts
}

// buildProbeBypassRoutes emits Exact path routes that skip ExtAuth only when
// User-Agent matches the webhook probe signature format. Requests to the same
// path without a valid UA fall through to the default authenticated route.
func buildProbeBypassRoutes(namePrefix, cluster string, paths []string) []*ir.HTTPRouteIR {
	if len(paths) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var routes []*ir.HTTPRouteIR
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" || path == "/" {
			continue
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		routes = append(routes, &ir.HTTPRouteIR{
			Name:      fmt.Sprintf("probe_%s_%s", namePrefix, sanitizeProbePath(path)),
			PathExact: path,
			Cluster:   cluster,
			HeaderMatches: []ir.HeaderMatchIR{{
				Name:      "user-agent",
				SafeRegex: probeUARegex,
			}},
			// ExtAuth nil → keep VH-level ext_authz Disabled (auth bypass).
			// No outbound zone-owner X-BFL-USER stamp (Authelia owns identity headers).
		})
	}
	return routes
}

func sanitizeProbePath(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, ".", "_")
	if path == "" {
		return "root"
	}
	return path
}

// buildAppExtAuthConfig returns Authelia ext_auth for app routes.
// All non-public routes use Legacy /api/verify/ (filter default; no path split).
func buildAppExtAuthConfig() *ir.ExtAuthConfigIR {
	return &ir.ExtAuthConfigIR{
		Cluster:    autheliaClusterPrefix,
		PathPrefix: autheliaVerifyPathPrefix,
		RequestHeaders: []string{
			"X-Original-URL",
			"X-Original-Method",
			"X-Forwarded-For",
			"X-Authorization",
			"Cookie",
		},
	}
}

func (t *Translator) buildFileserverRoutes(user *message.UserInfo, clusterSet map[string]*ir.ClusterIR) []*ir.HTTPRouteIR {
	var routes []*ir.HTTPRouteIR

	extAuthCfg := &ir.ExtAuthConfigIR{
		Cluster:    autheliaClusterPrefix,
		PathPrefix: autheliaVerifyPathPrefix,
		RequestHeaders: []string{
			"X-Original-URL",
			"X-Original-Method",
			"X-Forwarded-For",
			"X-Authorization",
			"Cookie",
		},
	}

	// Pass 1: node-scoped routes for every node. These carry the node name in
	// the path (e.g. /api/resources/external/<node>/ or the regex share form
	// ^/api/resources/share/<node>_.*) and are the MORE SPECIFIC matches.
	//
	// They MUST all be emitted before the master's generic catch-all prefixes
	// (pass 2), because Envoy evaluates routes top-to-bottom and stops at the
	// first match. A generic master prefix like /api/resources/external/ is a
	// strict prefix of /api/resources/external/<node>/, so if it came first it
	// would swallow every non-master node's request and misroute it to the
	// master node. The old single-pass loop emitted the master routes right
	// after the master node's own scoped routes, which — since FileserverNodes
	// is sorted by name and the master usually sorts first — placed them ahead
	// of later nodes' scoped routes and caused exactly that misrouting.
	for _, node := range user.FileserverNodes {
		proxyCluster := fmt.Sprintf("files_%s_%s", user.Name, node.NodeName)
		// L4 selects the Files pod on this node. Target its HTTP listener directly;
		// neither a shared Service nor the Files backend should dispatch it again.
		clusterSet[proxyCluster] = &ir.ClusterIR{
			Name: proxyCluster, Host: node.PodIP, Port: 80, UseDNS: false,
		}

		for _, pfx := range nodeLocationPrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:       fmt.Sprintf("files_node_%s_%s_%s", user.Name, node.NodeName, sanitizeName(pfx)),
				PathPrefix: fmt.Sprintf("%s%s/", pfx, node.NodeName),
				Cluster:    proxyCluster,
				RequestHeaders: map[string]string{
					"X-Terminus-Node": node.NodeName,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}

		for _, pfx := range nodeRegexSharePrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:      fmt.Sprintf("files_node_%s_%s_share_%s", user.Name, node.NodeName, sanitizeName(pfx)),
				PathRegex: fmt.Sprintf("^%s%s_.*", pfx, node.NodeName),
				Cluster:   proxyCluster,
				RequestHeaders: map[string]string{
					"X-Terminus-Node": node.NodeName,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}
	}

	// Pass 2: master node generic catch-all prefixes (no node name), emitted
	// last so the node-scoped routes above always take precedence.
	for _, node := range user.FileserverNodes {
		if !node.IsMaster {
			continue
		}
		proxyCluster := fmt.Sprintf("files_%s_%s", user.Name, node.NodeName)
		for _, pfx := range masterLocationPrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:       fmt.Sprintf("files_master_%s_%s", user.Name, sanitizeName(pfx)),
				PathPrefix: pfx,
				Cluster:    proxyCluster,
				RequestHeaders: map[string]string{
					"X-Terminus-Node": node.NodeName,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}
	}

	return routes
}

// Share uses the same entrance-specific Authelia verification as its existing
// catch-all, not user-service's mandatory-login identity subrequest. Files still
// checks share tokens/passwords. Only share and transfer paths get node routes;
// private file-management paths must not be added to the public entrance.
func (t *Translator) buildShareFileserverRoutes(user *message.UserInfo, clusterSet map[string]*ir.ClusterIR) []*ir.HTTPRouteIR {
	var routes []*ir.HTTPRouteIR
	for _, route := range t.buildFileserverRoutes(user, clusterSet) {
		shared := route.PathRegex != ""
		for _, prefix := range []string{"/api/resources/share/", "/api/preview/share/", "/api/tree/share/", "/api/raw/share/", "/upload/upload-link/", "/upload/file-uploaded-bytes/", "/videos/"} {
			shared = shared || strings.HasPrefix(route.PathPrefix, prefix)
		}
		if !shared {
			continue
		}
		route.Name = "share_" + route.Name
		route.RequestHeaders["X-Forwarded-Host"] = "%REQ(:AUTHORITY)%"
		route.RequestHeaders["X-Provider-Proxy"] = ""
		route.RequestHeaders["X-Internal-Share-Token"] = ""
		routes = append(routes, route)
	}
	return routes
}

func (t *Translator) buildCustomDomainVirtualHosts(user *message.UserInfo, app *message.AppInfo, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	var vhosts []*ir.VirtualHostIR

	customDomainMap := parseSettingsJSON(app.Settings, settingsCustomDomain)
	if len(customDomainMap) == 0 {
		return nil
	}

	for _, entrance := range app.Entrances {
		if entrance.Host == "" {
			klog.Warningf("buildCustomDomainVirtualHosts:app: %s,entrance: %s with empty host, skipping...", app.Name, entrance.Name)
			continue
		}

		entranceCustomDomain, ok := customDomainMap[entrance.Name]
		if !ok {
			continue
		}
		customDomainName := entranceCustomDomain[settingsCustomDomainThirdPartyDomain]
		if customDomainName == "" {
			continue
		}

		clusterName := fmt.Sprintf("custom_%s_%s_%s", user.Name, app.Name, entrance.Name)
		upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", entrance.Host, app.Namespace)
		clusterSet[clusterName] = &ir.ClusterIR{
			Name: clusterName, Host: upstreamHost, Port: uint32(entrance.Port), UseDNS: true,
		}

		customRoute := &ir.HTTPRouteIR{
			Name:             fmt.Sprintf("custom_%s_%s_%s_root", user.Name, app.Name, entrance.Name),
			PathPrefix:       "/",
			Cluster:          clusterName,
			WebSocketUpgrade: true,
		}

		customRoute.ExtAuth = buildAppExtAuthConfig()

		routes := buildProbeBypassRoutes(
			fmt.Sprintf("custom_%s_%s_%s", user.Name, app.Name, entrance.Name),
			clusterName,
			app.EntranceProbePaths[entrance.Name],
		)
		routes = append(routes, customRoute)

		vhost := &ir.VirtualHostIR{
			Name:     fmt.Sprintf("custom_%s_%s_%s", user.Name, app.Name, entrance.Name),
			Domains:  []string{customDomainName},
			Language: user.Language,
			UserZone: user.Zone,
			UserName: user.Name,
			Routes:   routes,
		}
		vhosts = append(vhosts, vhost)
	}

	return vhosts
}

func (t *Translator) buildStreamListeners(resources *message.Resources, clusterSet map[string]*ir.ClusterIR) []*ir.StreamListenerIR {
	if resources == nil {
		return nil
	}
	resources.Sort()
	var listeners []*ir.StreamListenerIR
	seen := make(map[string]bool)
	// First binding wins for a given (protocol, exposePort): only one listener can
	// bind to 0.0.0.0:port. Later declarations are ignored and must not create
	// or overwrite another user's cluster for that port.
	portOwnerUser := make(map[string]string)

	for _, user := range resources.Users {
		for _, app := range user.Apps {
			for _, port := range app.Ports {
				if port.Host == "" || port.ExposePort < 1 || port.ExposePort > 65535 {
					klog.Warningf("buildStreamListeners: user: %s, app: %s, invalid port (host=%q exposePort=%d), skipping", user.Name, app.Name, port.Host, port.ExposePort)
					continue
				}

				proto := strings.ToLower(port.Protocol)
				if proto == "" {
					proto = "tcp"
				}

				portKey := fmt.Sprintf("%s:%d", proto, port.ExposePort)
				if seen[portKey] {
					klog.V(2).Infof(
						"buildStreamListeners: duplicate exposed %s from user %s (app %s), skipping; port already bound for user %s",
						portKey, user.Name, app.Name, portOwnerUser[portKey],
					)
					continue
				}
				seen[portKey] = true
				portOwnerUser[portKey] = user.Name

				clusterName := fmt.Sprintf("stream_%s_%s_%d", proto, sanitizeName(user.Name), port.ExposePort)
				upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", port.Host, app.Namespace)
				clusterSet[clusterName] = &ir.ClusterIR{
					Name: clusterName, Host: upstreamHost, Port: uint32(port.Port), UseDNS: true,
				}

				listeners = append(listeners, &ir.StreamListenerIR{
					Name:     clusterName,
					Address:  "0.0.0.0",
					Port:     uint32(port.ExposePort),
					Protocol: proto,
					Cluster:  clusterName,
				})
			}
		}
	}

	return listeners
}

func parseSettingsJSON(settings map[string]string, key string) map[string]map[string]string {
	r := make(map[string]map[string]string)
	data, ok := settings[key]
	if !ok || data == "" {
		return r
	}
	err := json.Unmarshal([]byte(data), &r)
	if err != nil {
		klog.Warningf("parseSettingsJSON: unmarshal error for key %q: %v", key, err)
	}
	return r
}

var sanitizeReplacer = strings.NewReplacer("/", "_", ".", "_")

func sanitizeName(s string) string {
	s = sanitizeReplacer.Replace(s)
	s = strings.TrimPrefix(s, "_")
	s = strings.TrimSuffix(s, "_")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func systemAppPriority(appName string) int {
	if appName == "olares-app" {
		return systemServicePriority
	}
	return 0
}
