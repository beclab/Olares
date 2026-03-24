package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/telepresenceio/watchable"
	"k8s.io/klog/v2"
)

const (
	vpnCIDR               = "100.64.0.0/24"
	autheliaClusterPrefix = "authelia_backend"
	autheliaHostFormat    = "authelia-backend.user-system-%s.svc.cluster.local"
	fileserverHostFormat  = "files-%s.user-system-%s.svc.cluster.local"
	autheliaPort          = uint32(9091)
	autheliaPathPrefix    = "/api/authz/ext-authz/"

	settingsCustomDomain                 = "customDomain"
	settingsCustomDomainThirdLevelDomain = "third_level_domain"
	settingsCustomDomainThirdPartyDomain = "third_party_domain"
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
}

var systemServices = []systemServiceDef{
	{Name: "auth", SvcFormat: "authelia-svc.%s.svc.cluster.local"},
	{Name: "desktop", SvcFormat: "edge-desktop.%s.svc.cluster.local"},
	{Name: "wizard", SvcFormat: "wizard.%s.svc.cluster.local"},
}

var fileserverPatchApps = map[string]bool{
	"files":    true,
	"settings": true,
}

type Config struct {
	SSLServerPort      int
	SSLProxyServerPort int
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
	clusterSet := make(map[string]*ir.ClusterIR)

	// HTTP redirect listener (port 81 -> HTTPS)
	xds.Listeners = append(xds.Listeners, t.buildHTTPRedirectListener())

	// Group ephemeral users by the owner's zone, so we can merge their wizard
	// virtual host into the owner's filter chain instead of creating a
	// separate filter chain that would cause SNI overlap with the owner's
	// zone wildcard (*.alice.zone) and force Envoy to reject the listener.
	ephemeralsByZone := make(map[string][]*message.UserInfo)
	for _, u := range resources.Users {
		if u.IsEphemeral && u.SSL != nil {
			ephemeralsByZone[u.SSL.Zone] = append(ephemeralsByZone[u.SSL.Zone], u)
		}
	}

	for _, user := range resources.Users {
		if user.IsEphemeral {
			// Ephemeral users are merged into the owner's filter chain below.
			// A separate filter chain would conflict with the owner's
			// *.zone wildcard → Envoy "overlapping matching rules" error.
			continue
		}
		if user.SSL == nil || user.SSL.CertData == "" {
			klog.V(4).Infof("translator: user %s has no SSL config, skipping HTTPS listeners", user.Name)
			continue
		}

		// Build wizard virtual hosts for any ephemeral users whose zone
		// matches this owner's zone, to be merged into the owner's HCM.
		var ephemeralVHosts []*ir.VirtualHostIR
		for _, ephUser := range ephemeralsByZone[user.SSL.Zone] {
			if wvh := t.buildEphemeralWizardVHost(ephUser, user.SSL.Zone, clusterSet); wvh != nil {
				ephemeralVHosts = append(ephemeralVHosts, wvh)
			}
		}

		zone := user.SSL.Zone
		vhosts := t.buildUserVirtualHosts(user, zone, false, clusterSet)
		vhosts = append(vhosts, ephemeralVHosts...)
		t.applyDenyAllRestrictions(user, vhosts, ephemeralVHosts)

		// HTTPS listeners on port 443 and 444 (with proxy_protocol)
		for _, portCfg := range []struct {
			port          uint32
			proxyProtocol bool
		}{
			{uint32(t.cfg.SSLServerPort), false},
			{uint32(t.cfg.SSLProxyServerPort), true},
		} {
			listeners := t.buildUserFilterChains(user, vhosts, portCfg.port, portCfg.proxyProtocol)
			xds.HTTPListeners = append(xds.HTTPListeners, listeners...)
		}

		// TLS secrets
		xds.Secrets = append(xds.Secrets, &ir.SecretIR{
			Name:     fmt.Sprintf("main-tls-%s", user.Name),
			CertData: user.SSL.CertData,
			KeyData:  user.SSL.KeyData,
		})
		for _, cert := range user.CustomDomainCerts {
			xds.Secrets = append(xds.Secrets, &ir.SecretIR{
				Name:     fmt.Sprintf("custom-tls-%s-%s", user.Name, cert.Domain),
				CertData: cert.CertData,
				KeyData:  cert.KeyData,
			})
		}

		// Authelia cluster per user
		autheliaClName := fmt.Sprintf("%s_%s", autheliaClusterPrefix, user.Name)
		clusterSet[autheliaClName] = &ir.ClusterIR{
			Name:   autheliaClName,
			Host:   fmt.Sprintf(autheliaHostFormat, user.Name),
			Port:   autheliaPort,
			UseDNS: true,
		}
	}

	// Stream listeners (TCP/UDP) for all users' apps — route directly to services
	xds.StreamListeners = t.buildStreamListeners(resources, clusterSet)

	for _, c := range clusterSet {
		xds.Clusters = append(xds.Clusters, c)
	}
	sort.Slice(xds.Clusters, func(i, j int) bool {
		return xds.Clusters[i].Name < xds.Clusters[j].Name
	})

	return xds
}

// buildEphemeralWizardVHost builds the single wizard virtual host for an
// ephemeral user.  This VH is merged into the owner's filter chain rather
// than creating a separate filter chain — the owner's *.zone wildcard already
// covers wizard-{guest}.zone so no extra filter chain is needed.
func (t *Translator) buildEphemeralWizardVHost(ephUser *message.UserInfo, zone string, clusterSet map[string]*ir.ClusterIR) *ir.VirtualHostIR {
	wizardDef := systemServiceDef{Name: "wizard", SvcFormat: "wizard.%s.svc.cluster.local"}
	return t.buildSystemVirtualHost(ephUser, wizardDef, zone, true, ephUser.Namespace, clusterSet)
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
func (t *Translator) applyDenyAllRestrictions(user *message.UserInfo, vhosts []*ir.VirtualHostIR, extraVHosts []*ir.VirtualHostIR) {
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
	if user.LocalDomainIP != "" {
		restrictCIDRs = append(restrictCIDRs, user.LocalDomainIP+"/32")
	}
	restrictCIDRs = append(restrictCIDRs, user.AllowCIDRs...)

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

// buildUserFilterChains produces one or more HTTPListenerIR entries for a
// single user on a given port, using pre-built virtual hosts.
//
// SNI strategy: one filter chain per ServerNameDomain entry, each matching
// exactly zoneDomain + *.zoneDomain. Custom domains get separate chains with
// their own TLS cert.
func (t *Translator) buildUserFilterChains(user *message.UserInfo, vhosts []*ir.VirtualHostIR, port uint32, proxyProtocol bool) []*ir.HTTPListenerIR {
	defaultResponse := &ir.DirectResponseIR{
		Status:      421,
		Body:        "<h1><a href='https://www.olares.com/'>Olares</a></h1>",
		ContentType: "text/html",
	}

	tlsCert := &ir.SecretIR{
		Name:     fmt.Sprintf("main-tls-%s", user.Name),
		CertData: user.SSL.CertData,
		KeyData:  user.SSL.KeyData,
	}

	makeListener := func(name string, snis []string) *ir.HTTPListenerIR {
		return &ir.HTTPListenerIR{
			Name:            name,
			Address:         "0.0.0.0",
			Port:            port,
			TLS:             true,
			ProxyProtocol:   proxyProtocol,
			VirtualHosts:    vhosts,
			DefaultResponse: defaultResponse,
			PingRoute:       true,
			SNIMatches:      snis,
			TLSCert:         tlsCert,
			UserName:        user.Name,
		}
	}

	customDomainSet := make(map[string]bool, len(user.CustomDomainCerts))
	for _, cert := range user.CustomDomainCerts {
		customDomainSet[cert.Domain] = true
	}

	// One filter chain per zone domain — identical for both deny_all and
	// non-deny_all users.  Each chain matches exactly two SNI patterns:
	//   zoneDomain  +  *.zoneDomain
	//
	// This makes every filter chain stable and independent:
	//   • Adding/removing an app   → RDS update only, filter chain unchanged.
	//   • Rotating a cert          → SDS update only, filter chain unchanged.
	//   • Adding a zone            → NEW filter chain added in-place.
	//   • Changing AllowedDomains  → RDS only (per-VH RBAC changes).
	//   • Adding ephemeral user    → RDS only (wizard VH added).
	//   • Changing AllowCIDRs      → RDS only (per-VH RBAC changes).
	//
	// For deny_all users the access control that was previously implemented
	// via separate allowed/restricted filter chains with SourceCIDRs is now
	// handled by per-VH RBAC rules in the route config (see SourceCIDRs on
	// VirtualHostIR above).
	var result []*ir.HTTPListenerIR
	for _, zoneDomain := range user.ServerNameDomains {
		if customDomainSet[zoneDomain] {
			continue
		}
		snis := []string{"*." + zoneDomain, zoneDomain}
		result = append(result, makeListener(
			fmt.Sprintf("https_%d_%s_%s", port, user.Name, sanitizeName(zoneDomain)),
			snis,
		))
	}

	// Custom domain filter chains (separate TLS certs, same virtual hosts)
	for _, cert := range user.CustomDomainCerts {
		customTLS := &ir.SecretIR{
			Name:     fmt.Sprintf("custom-tls-%s-%s", user.Name, cert.Domain),
			CertData: cert.CertData,
			KeyData:  cert.KeyData,
		}
		result = append(result, &ir.HTTPListenerIR{
			Name:            fmt.Sprintf("https_%d_%s_custom_%s", port, user.Name, sanitizeName(cert.Domain)),
			Address:         "0.0.0.0",
			Port:            port,
			TLS:             true,
			ProxyProtocol:   proxyProtocol,
			VirtualHosts:    vhosts,
			DefaultResponse: defaultResponse,
			PingRoute:       true,
			SNIMatches:      []string{cert.Domain},
			TLSCert:         customTLS,
			UserName:        user.Name,
		})
	}

	return result
}

// buildZoneWildcardSNI returns the two SNI patterns for every domain in
// serverNameDomains: the exact domain and a single-label wildcard (*.domain).
// The output is sorted for deterministic comparison.
// Used by tests; production code now calls this per-zone (one entry at a time).
func buildZoneWildcardSNI(serverNameDomains []string) []string {
	seen := make(map[string]bool, len(serverNameDomains)*2)
	snis := make([]string, 0, len(serverNameDomains)*2)
	for _, d := range serverNameDomains {
		for _, s := range []string{d, "*." + d} {
			if !seen[s] {
				seen[s] = true
				snis = append(snis, s)
			}
		}
	}
	sort.Strings(snis)
	return snis
}

func (t *Translator) buildUserVirtualHosts(user *message.UserInfo, zone string, isEphemeral bool, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	var vhosts []*ir.VirtualHostIR
	namespace := user.Namespace

	// Profile/root server
	profileCluster := fmt.Sprintf("profile_service_%s", user.Name)
	profileHost := fmt.Sprintf("profile-service.%s.svc.cluster.local", namespace)
	clusterSet[profileCluster] = &ir.ClusterIR{
		Name: profileCluster, Host: profileHost, Port: 3000, UseDNS: true,
	}

	var zoneAliases []string
	zoneTokens := strings.Split(zone, ".")
	if len(zoneTokens) > 1 {
		zoneAliases = append(zoneAliases, fmt.Sprintf("%s.%s", zoneTokens[0], "olares.local"))
	}

	profileVHost := &ir.VirtualHostIR{
		Name:     fmt.Sprintf("profile_%s", user.Name),
		Domains:  append([]string{zone}, zoneAliases...),
		Language: user.Language,
		UserZone: zone,
		UserName: user.Name,
		Routes: []*ir.HTTPRouteIR{{
			Name:       fmt.Sprintf("profile_root_%s", user.Name),
			PathPrefix: "/",
			Cluster:    profileCluster,
			RequestHeaders: map[string]string{
				"X-BFL-USER": user.Name,
			},
		}},
	}
	vhosts = append(vhosts, profileVHost)

	// Application servers
	for _, app := range user.Apps {
		appVhosts := t.buildAppVirtualHosts(user, app, zone, isEphemeral, clusterSet)
		vhosts = append(vhosts, appVhosts...)
	}

	// System services (auth, desktop, wizard)
	for _, def := range systemServices {
		vhost := t.buildSystemVirtualHost(user, def, zone, isEphemeral, namespace, clusterSet)
		vhosts = append(vhosts, vhost)
	}

	// Custom domain virtual hosts
	for _, app := range user.Apps {
		cvhosts := t.buildCustomDomainVirtualHosts(user, app, clusterSet)
		vhosts = append(vhosts, cvhosts...)
	}

	return vhosts
}

func (t *Translator) buildAppVirtualHosts(user *message.UserInfo, app *message.AppInfo, zone string, isEphemeral bool, clusterSet map[string]*ir.ClusterIR) []*ir.VirtualHostIR {
	var vhosts []*ir.VirtualHostIR

	var appDomainConfigs []defaultThirdLevelDomainConfig
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

		prefix := resolveEntrancePrefix(app.Entrances, index, app.Appid, appDomainConfigs)

		hostname := fmt.Sprintf("%s.%s", prefix, zone)
		if isEphemeral {
			hostname = fmt.Sprintf("%s-%s.%s", prefix, user.Name, zone)
		}

		localHost := toLocalDomain(hostname)
		domains := []string{hostname}
		if localHost != hostname {
			domains = append(domains, localHost)
		}

		if entranceMap, ok := customDomainMap[entrance.Name]; ok {
			if thirdLevel := entranceMap[settingsCustomDomainThirdLevelDomain]; thirdLevel != "" {
				extHostname := fmt.Sprintf("%s.%s", thirdLevel, zone)
				if isEphemeral {
					extHostname = fmt.Sprintf("%s-%s.%s", thirdLevel, user.Name, zone)
				}
				domains = append(domains, extHostname, toLocalDomain(extHostname))
			}
		}

		clusterName := fmt.Sprintf("app_%s_%s_%s", user.Name, app.Name, entrance.Name)
		upstreamHost := fmt.Sprintf("%s.%s.svc.cluster.local", entrance.Host, app.Namespace)
		clusterSet[clusterName] = &ir.ClusterIR{
			Name: clusterName, Host: upstreamHost, Port: uint32(entrance.Port), UseDNS: true,
		}

		_, enableOIDC := app.Settings["oidc.client.id"]

		vhost := &ir.VirtualHostIR{
			Name:                  fmt.Sprintf("app_%s_%s_%s", user.Name, app.Name, entrance.Name),
			Domains:               domains,
			EnableOIDC:            enableOIDC,
			EnableWindowPushState: entrance.WindowPushState,
			Language:              user.Language,
			UserZone:              zone,
			UserName:              user.Name,
		}

		var routes []*ir.HTTPRouteIR

		if fileserverPatchApps[prefix] {
			routes = append(routes, t.buildFileserverRoutes(user, clusterSet)...)
		}

		routes = append(routes, &ir.HTTPRouteIR{
			Name:       fmt.Sprintf("default_%s_%s_%s", user.Name, app.Name, entrance.Name),
			PathPrefix: "/",
			Cluster:    clusterName,
			RequestHeaders: map[string]string{
				"X-BFL-USER": user.Name,
			},
			WebSocketUpgrade: true,
		})

		vhost.Routes = routes
		vhosts = append(vhosts, vhost)
	}

	return vhosts
}

func (t *Translator) buildFileserverRoutes(user *message.UserInfo, clusterSet map[string]*ir.ClusterIR) []*ir.HTTPRouteIR {
	var routes []*ir.HTTPRouteIR

	autheliaClName := fmt.Sprintf("%s_%s", autheliaClusterPrefix, user.Name)
	extAuthCfg := &ir.ExtAuthConfigIR{
		Cluster:    autheliaClName,
		PathPrefix: autheliaPathPrefix,
		RequestHeaders: []string{
			"X-Original-URL",
			"X-Original-Method",
			"X-Forwarded-For",
			"X-BFL-USER",
			"X-Authorization",
			"Cookie",
		},
	}

	for _, node := range user.FileserverNodes {
		proxyCluster := fmt.Sprintf("files_%s_%s", user.Name, node.NodeName)
		proxyHost := fmt.Sprintf(fileserverHostFormat, node.NodeName, user.Name)
		clusterSet[proxyCluster] = &ir.ClusterIR{
			Name: proxyCluster, Host: proxyHost, Port: 28080, UseDNS: true,
		}

		for _, pfx := range nodeLocationPrefixes {
			routes = append(routes, &ir.HTTPRouteIR{
				Name:       fmt.Sprintf("files_node_%s_%s_%s", user.Name, node.NodeName, sanitizeName(pfx)),
				PathPrefix: fmt.Sprintf("%s%s/", pfx, node.NodeName),
				Cluster:    proxyCluster,
				RequestHeaders: map[string]string{
					"X-BFL-USER":       user.Name,
					"X-Terminus-Node":  node.NodeName,
					"X-Provider-Proxy": proxyHost,
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
					"X-BFL-USER":       user.Name,
					"X-Terminus-Node":  node.NodeName,
					"X-Provider-Proxy": proxyHost,
				},
				ExtAuth:          extAuthCfg,
				WebSocketUpgrade: true,
			})
		}

		if node.IsMaster {
			for _, pfx := range masterLocationPrefixes {
				routes = append(routes, &ir.HTTPRouteIR{
					Name:       fmt.Sprintf("files_master_%s_%s", user.Name, sanitizeName(pfx)),
					PathPrefix: pfx,
					Cluster:    proxyCluster,
					RequestHeaders: map[string]string{
						"X-BFL-USER":       user.Name,
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

func (t *Translator) buildSystemVirtualHost(user *message.UserInfo, def systemServiceDef, zone string, isEphemeral bool, namespace string, clusterSet map[string]*ir.ClusterIR) *ir.VirtualHostIR {
	hostname := fmt.Sprintf("%s.%s", def.Name, zone)
	if isEphemeral {
		hostname = fmt.Sprintf("%s-%s.%s", def.Name, user.Name, zone)
	}

	clusterName := fmt.Sprintf("nonapp_%s_%s", def.Name, user.Name)
	upstreamHost := fmt.Sprintf(def.SvcFormat, namespace)
	clusterSet[clusterName] = &ir.ClusterIR{
		Name: clusterName, Host: upstreamHost, Port: 80, UseDNS: true,
	}
	localHost := toLocalDomain(hostname)
	domains := []string{hostname}
	if localHost != hostname {
		domains = append(domains, localHost)
	}

	return &ir.VirtualHostIR{
		Name:     fmt.Sprintf("nonapp_%s_%s", def.Name, user.Name),
		Domains:  domains,
		Language: user.Language,
		UserZone: zone,
		UserName: user.Name,
		Routes: []*ir.HTTPRouteIR{{
			Name:       fmt.Sprintf("nonapp_%s_root_%s", def.Name, user.Name),
			PathPrefix: "/",
			Cluster:    clusterName,
			RequestHeaders: map[string]string{
				"X-BFL-USER": user.Name,
			},
			WebSocketUpgrade: true,
		}},
	}
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

		vhost := &ir.VirtualHostIR{
			Name:     fmt.Sprintf("custom_%s_%s_%s", user.Name, app.Name, entrance.Name),
			Domains:  []string{customDomainName},
			Language: user.Language,
			UserZone: user.Zone,
			UserName: user.Name,
			Routes: []*ir.HTTPRouteIR{{
				Name:       fmt.Sprintf("custom_%s_%s_%s_root", user.Name, app.Name, entrance.Name),
				PathPrefix: "/",
				Cluster:    clusterName,
				RequestHeaders: map[string]string{
					"X-BFL-USER": user.Name,
				},
				WebSocketUpgrade: true,
			}},
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

type defaultThirdLevelDomainConfig struct {
	EntranceName     string `json:"entranceName"`
	ThirdLevelDomain string `json:"thirdLevelDomain"`
}

func resolveEntrancePrefix(entrances []*message.EntranceInfo, index int, appid string, configs []defaultThirdLevelDomainConfig) string {
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

func toLocalDomain(hostname string) string {
	tokens := strings.Split(hostname, ".")
	if len(tokens) < 2 {
		return hostname
	}
	return strings.Join([]string{tokens[0], tokens[1], "olares", "local"}, ".")
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
