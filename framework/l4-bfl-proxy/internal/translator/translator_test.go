package translator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslate_DefaultDomainAndCustomCertificate(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, SSLProxyServerPort: 444}}
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL:   &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		Users: []*message.UserInfo{{Name: "alice", CustomDomainCerts: []*message.CertInfo{{Domain: "myapp.custom.io", CertData: "custom-cert", KeyData: "custom-key"}}, Apps: []*message.AppInfo{{Name: "myapp", Appid: "myapp", Namespace: "myapp-alice", Owner: "alice", Entrances: []*message.EntranceInfo{{Name: "web", Host: "web", Port: 8080}}, Settings: map[string]string{"customDomain": `{"web":{"third_party_domain":"myapp.custom.io"}}`}}}}},
	}
	xds := tr.Translate(resources)
	require.Len(t, xds.Secrets, 2)
	require.Len(t, xds.HTTPListeners, 5)
	for _, listener := range xds.HTTPListeners {
		switch {
		case listener.Name == "http_82":
			assert.False(t, listener.TLS)
		case strings.HasSuffix(listener.Name, "_default"):
			assert.ElementsMatch(t, []string{"lab.example.com", "*.lab.example.com"}, listener.SNIMatches)
			assert.Equal(t, "machine-cert", listener.TLSCert.CertData)
			assert.NotContains(t, listener.SNIMatches, "myapp.custom.io")
		default:
			assert.Equal(t, []string{"myapp.custom.io"}, listener.SNIMatches)
			assert.Equal(t, "custom-cert", listener.TLSCert.CertData)
		}
	}
}

func TestTranslate_DenyAllRestrictsPersonalApps(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, SSLProxyServerPort: 444}}
	resources := &message.Resources{LocalZone: "lab.olares.local", LocalDomainIP: "192.168.1.100", Users: []*message.UserInfo{{Name: "alice", DenyAll: true, AllowedDomains: []string{"allowed.lab.olares.local"}, Apps: []*message.AppInfo{
		{Name: "allowed", Appid: "allowed", Namespace: "allowed-alice", Entrances: []*message.EntranceInfo{{Name: "web", Host: "web", Port: 8080}}},
		{Name: "private", Appid: "private", Namespace: "private-alice", Entrances: []*message.EntranceInfo{{Name: "web", Host: "web", Port: 8080}}},
	}}}}
	xds := tr.Translate(resources)
	require.Len(t, xds.HTTPListeners, 1)
	found := false
	for _, vh := range xds.HTTPListeners[0].VirtualHosts {
		if vh.Name == "app_alice_allowed_web" {
			assert.Empty(t, vh.SourceCIDRs)
		}
		if vh.Name == "app_alice_private_web" {
			found = true
			assert.Contains(t, vh.SourceCIDRs, vpnCIDR)
			assert.Contains(t, vh.SourceCIDRs, "192.168.1.100/32")
		}
	}
	assert.True(t, found)
}

func TestTranslate_Full(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, SSLProxyServerPort: 444}}
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL:   &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		Users: []*message.UserInfo{{Name: "alice", Apps: []*message.AppInfo{{Name: "vault", Appid: "vault", Owner: "alice", Namespace: "vault-alice", Entrances: []*message.EntranceInfo{{Name: "main", Host: "vault-svc", Port: 8080}}, Ports: []*message.PortInfo{{Host: "vault-svc", Port: 8080, ExposePort: 48126, Protocol: "tcp"}}}}}},
	}
	xds := tr.Translate(resources)
	require.Len(t, xds.Listeners, 1)
	assert.Equal(t, "http_redirect_81", xds.Listeners[0].Name)
	require.Len(t, xds.HTTPListeners, 3)
	for _, listener := range xds.HTTPListeners {
		if listener.Name == "http_82" {
			assert.False(t, listener.TLS)
			assert.Nil(t, listener.TLSCert)
		} else {
			assert.True(t, listener.TLS)
			assert.Equal(t, "default-domain-tls", listener.TLSCert.Name)
			assert.Equal(t, listener.Port == 444, listener.ProxyProtocol)
		}
	}
	require.Len(t, xds.StreamListeners, 1)
	assert.Equal(t, "stream_tcp_alice_48126", xds.StreamListeners[0].Name)
}

func TestTranslate_NoZone_LocalHTTPUsesMachineName(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, SSLProxyServerPort: 444}}
	for _, localZone := range []string{"", "lab.olares.local"} {
		t.Run(localZone, func(t *testing.T) {
			xds := tr.Translate(&message.Resources{LocalZone: localZone, Users: []*message.UserInfo{{Name: "alice"}, {Name: "bob"}}})
			require.Len(t, xds.HTTPListeners, 1)
			local := xds.HTTPListeners[0]
			assert.Equal(t, "http_82", local.Name)
			zone := localZone
			if zone == "" {
				zone = "olares.local"
			}
			desktops := 0
			for _, vh := range local.VirtualHosts {
				if vh.Name == "nonapp_desktop" {
					desktops++
					assert.Equal(t, []string{"desktop." + zone}, vh.Domains)
					assert.Empty(t, vh.UserName)
					require.NotNil(t, vh.Routes[0].ExtAuth)
					assert.Empty(t, vh.Routes[0].ExtAuth.User)
				}
			}
			assert.Equal(t, 1, desktops)
		})
	}
}

func TestTranslate_NoCert_OneSharedLocalHTTPListener(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443, SSLProxyServerPort: 444}}
	xds := tr.Translate(&message.Resources{DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local", Users: []*message.UserInfo{{Name: "alice"}, {Name: "bob"}}})
	require.Len(t, xds.HTTPListeners, 1)
	local := xds.HTTPListeners[0]
	assert.Equal(t, "http_82", local.Name)
	assert.False(t, local.TLS)
	assert.Nil(t, local.TLSCert)
	for _, vh := range local.VirtualHosts {
		for _, domain := range vh.Domains {
			assert.True(t, strings.HasSuffix(domain, ".olares.local"))
		}
	}
}

func TestBuildHTTPRedirectListener(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	listener := tr.buildHTTPRedirectListener()

	assert.Equal(t, "http_redirect_81", listener.Name)
	assert.Equal(t, "0.0.0.0", listener.Address)
	assert.Equal(t, uint32(81), listener.Port)
	assert.Equal(t, ir.ProtocolHTTP, listener.Protocol)
	require.NotNil(t, listener.HTTPRedirect)
	assert.Equal(t, "https", listener.HTTPRedirect.Scheme)
	assert.Equal(t, 301, listener.HTTPRedirect.Code)
}

// ---------------------------------------------------------------------------
// buildStreamListeners
// ---------------------------------------------------------------------------

func TestBuildStreamListeners_TCP(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Apps: []*message.AppInfo{
					{
						Owner:     "alice",
						Namespace: "app-ns",
						Ports: []*message.PortInfo{
							{Host: "app-svc", Port: 30000, ExposePort: 30000, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	listeners := tr.buildStreamListeners(resources, clusterSet)
	require.Len(t, listeners, 1)
	assert.Equal(t, "stream_tcp_alice_30000", listeners[0].Name)
	assert.Equal(t, "tcp", listeners[0].Protocol)
	assert.Equal(t, uint32(30000), listeners[0].Port)
}

func TestBuildStreamListeners_UDP(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Apps: []*message.AppInfo{
					{
						Owner:     "alice",
						Namespace: "app-ns",
						Ports: []*message.PortInfo{
							{Host: "app-svc", Port: 51820, ExposePort: 51820, Protocol: "udp"},
						},
					},
				},
			},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	listeners := tr.buildStreamListeners(resources, clusterSet)
	require.Len(t, listeners, 1)
	assert.Equal(t, "stream_udp_alice_51820", listeners[0].Name)
	assert.Equal(t, "udp", listeners[0].Protocol)
}

func TestBuildStreamListeners_DuplicateExposePortSecondUserSkipped(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:      "bob",
				Namespace: "user-space-bob",
				Apps: []*message.AppInfo{
					{
						Name:      "app1",
						Namespace: "ns-b",
						Owner:     "bob",
						Ports: []*message.PortInfo{
							{Host: "svc-b", Port: 80, ExposePort: 10000, Protocol: "tcp"},
						},
					},
				},
			},
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Apps: []*message.AppInfo{
					{
						Name:      "app2",
						Namespace: "ns-a",
						Owner:     "alice",
						Ports: []*message.PortInfo{
							{Host: "svc-a", Port: 80, ExposePort: 10000, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}
	clusterSet := make(map[string]*ir.ClusterIR)
	listeners := tr.buildStreamListeners(resources, clusterSet)
	require.Len(t, listeners, 1)
	assert.Equal(t, "stream_tcp_alice_10000", listeners[0].Name)
	assert.Equal(t, "stream_tcp_alice_10000", listeners[0].Cluster)
	c := clusterSet["stream_tcp_alice_10000"]
	require.NotNil(t, c)
	assert.Contains(t, c.Host, "svc-a")
	assert.NotContains(t, c.Host, "svc-b")
	_, hasBob := clusterSet["stream_tcp_bob_10000"]
	assert.False(t, hasBob)
}

func TestBuildStreamListeners_SkipsInvalidPort(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Apps: []*message.AppInfo{
					{
						Owner:     "alice",
						Namespace: "app-ns",
						Ports: []*message.PortInfo{
							{Host: "svc", Port: 80, ExposePort: 0, Protocol: "tcp"},
							{Host: "svc", Port: 80, ExposePort: -1, Protocol: "tcp"},
							{Host: "svc", Port: 80, ExposePort: 70000, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	listeners := tr.buildStreamListeners(resources, clusterSet)
	assert.Empty(t, listeners)
}

// ---------------------------------------------------------------------------
// helper functions
// ---------------------------------------------------------------------------

func TestSanitizeName(t *testing.T) {
	assert.Equal(t, "api_resources_cache", sanitizeName("/api/resources/cache/"))
	assert.Equal(t, "example_com", sanitizeName("example.com"))
}

// ---------------------------------------------------------------------------
// v3 / shared app routing — open to all users (no gating)
// ---------------------------------------------------------------------------

// makeSharedApp returns a v3 / shared AppInfo with one entrance.
func makeSharedApp() *message.AppInfo {
	return &message.AppInfo{
		Name:      "shareme",
		Appid:     "shareme",
		Namespace: "shareme-shared",
		Owner:     "admin",
		IsShared:  true,
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "shareme-svc", Port: 8080},
		},
	}
}

// makeSharedAppWithCustom returns a v3 / shared AppInfo with a custom domain.
func makeSharedAppWithCustom() *message.AppInfo {
	a := makeSharedApp()
	a.Settings = map[string]string{
		"customDomain": `{"web":{"third_party_domain":"shareme.example.io"}}`,
	}
	return a
}

func TestBuildAppVirtualHosts_HeadscalePublicRoutes(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en", DenyAll: true}
	app := &message.AppInfo{
		Name:      "olares-app",
		Appid:     "olares-app",
		IsSysApp:  true,
		Namespace: "user-space-alice",
		Entrances: []*message.EntranceInfo{{Name: "headscale", Host: "headscale-svc", Port: 80}},
	}

	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", make(map[string]*ir.ClusterIR))
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 6)

	want := []struct {
		path    string
		method  string
		cluster string
	}{
		{"/key", "GET", "app_alice_olares-app_headscale"},
		{"/ts2021", "GET", "app_alice_olares-app_headscale"},
		{"/ts2021", "POST", "app_alice_olares-app_headscale"},
		{"/api/refresh", "POST", "app_alice_olares-app_headscale"},
		{"/api/refresh", "OPTIONS", "app_alice_olares-app_headscale"},
	}
	assert.Equal(t, "headscale_alice_key_get", vhosts[0].Routes[0].Name)
	for i, expected := range want {
		route := vhosts[0].Routes[i]
		assert.Equal(t, expected.path, route.PathExact)
		assert.Equal(t, expected.method, route.HeaderMatches[0].Exact)
		assert.Equal(t, expected.cluster, route.Cluster)
		assert.True(t, route.BypassSourceCIDR)
		require.NotNil(t, route.ExtAuth)
		assert.Equal(t, "authelia_backend_alice", route.ExtAuth.Cluster)
		assert.Equal(t, autheliaVerifyPathPrefix, route.ExtAuth.PathPrefix)
	}
	assert.Equal(t, "/", vhosts[0].Routes[5].PathPrefix)
	assert.False(t, vhosts[0].Routes[5].BypassSourceCIDR)
	require.NotNil(t, vhosts[0].Routes[5].ExtAuth)

	user.DenyAll = false
	vhosts = tr.buildAppVirtualHosts(user, app, "alice.example.com", make(map[string]*ir.ClusterIR))
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	assert.False(t, vhosts[0].Routes[0].BypassSourceCIDR)
	require.NotNil(t, vhosts[0].Routes[0].ExtAuth)
}

func TestBuildAppVirtualHosts_HeadscalePrefixDoesNotGrantPublicRoutes(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en", DenyAll: true}
	app := &message.AppInfo{
		Name:      "third-party-app",
		Appid:     "third-party-app",
		Namespace: "user-space-alice",
		Entrances: []*message.EntranceInfo{{Name: "web", Host: "third-party-svc", Port: 8080}},
		Settings: map[string]string{
			"defaultThirdLevelDomainConfig": `[{"entranceName":"web","thirdLevelDomain":"headscale"}]`,
		},
	}

	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", make(map[string]*ir.ClusterIR))
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	assert.Equal(t, "headscale.alice.example.com", vhosts[0].Domains[0])
	assert.False(t, vhosts[0].Routes[0].BypassSourceCIDR)
	require.NotNil(t, vhosts[0].Routes[0].ExtAuth)

	tr.applyDenyAllRestrictions(user, vhosts, nil, "192.168.1.100")
	assert.Contains(t, vhosts[0].SourceCIDRs, vpnCIDR)
	assert.Contains(t, vhosts[0].SourceCIDRs, "192.168.1.100/32")
}

// Any user (admin or normal) reaches the real upstream of a v3 / shared app.
func TestBuildAppVirtualHosts_SharedApp_OpenToAllUsers(t *testing.T) {
	for _, name := range []string{"admin", "alice"} {
		t.Run(name, func(t *testing.T) {
			tr := &Translator{cfg: &Config{}}
			user := &message.UserInfo{Name: name, Language: "en"}
			clusterSet := make(map[string]*ir.ClusterIR)

			vhosts := tr.buildAppVirtualHosts(user, makeSharedApp(), name+".example.com", clusterSet)
			require.Len(t, vhosts, 1)
			require.Len(t, vhosts[0].Routes, 1)
			r := vhosts[0].Routes[0]
			assert.Nil(t, r.DirectResponse, "no 403 gate is emitted; shared apps are open")
			assert.NotEmpty(t, r.Cluster, "every user must reach the upstream cluster")
			assert.NotEmpty(t, clusterSet, "an upstream cluster must be registered")

			require.NotNil(t, r.ExtAuth, "shared apps must be gated behind Authelia ext_auth")
			assert.Equal(t, fmt.Sprintf("authelia_backend_%s", name), r.ExtAuth.Cluster)
			assert.Equal(t, autheliaVerifyPathPrefix, r.ExtAuth.PathPrefix)
			assert.False(t, r.ExtAuth.Disabled)
		})
	}
}

// Shared apps reached via a custom domain are also gated behind Authelia.
func TestBuildCustomDomainVirtualHosts_SharedApp_ExtAuth(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en", Zone: "alice.example.com"}
	clusterSet := make(map[string]*ir.ClusterIR)

	vhosts := tr.buildCustomDomainVirtualHosts(user, makeSharedAppWithCustom(), clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	r := vhosts[0].Routes[0]
	require.NotNil(t, r.ExtAuth, "shared apps must stay gated even on a custom domain")
	assert.Equal(t, "authelia_backend_alice", r.ExtAuth.Cluster)
	assert.Equal(t, autheliaVerifyPathPrefix, r.ExtAuth.PathPrefix)
}

// v1/v2 (non-shared) apps reach their upstream cluster and are gated behind
// the viewing user's Authelia instance just like shared apps.
func TestBuildAppVirtualHosts_NonSharedApp(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := &message.AppInfo{
		Name:      "v1app",
		Appid:     "v1app",
		Namespace: "v1app-alice",
		Owner:     "alice",
		IsShared:  false,
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "v1app-svc", Port: 8080},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	r := vhosts[0].Routes[0]
	assert.Nil(t, r.DirectResponse)
	require.NotNil(t, r.ExtAuth, "non-shared apps must also be gated behind Authelia ext_auth")
	assert.Equal(t, "authelia_backend_alice", r.ExtAuth.Cluster)
	assert.Equal(t, autheliaVerifyPathPrefix, r.ExtAuth.PathPrefix)
	assert.False(t, r.ExtAuth.Disabled)
	assert.NotEmpty(t, clusterSet)
}

func TestBuildAppVirtualHosts_PublicKeepsExtAuthOmitsOutboundXBFLUser(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)
	app := &message.AppInfo{
		Name:      "pub",
		Appid:     "pub",
		Namespace: "pub-alice",
		Owner:     "alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "pub-svc", Port: 8080, AuthLevel: "public"},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	r := vhosts[0].Routes[0]
	require.NotNil(t, r.ExtAuth, "public must still mount ExtAuth (Bypass at Authelia)")
	assert.Equal(t, "authelia_backend_alice", r.ExtAuth.Cluster)
	assert.Nil(t, r.RequestHeaders, "outbound zone-owner X-BFL-USER stamp must be omitted")
}

func TestBuildProbeBypassRoutes_ExactPathAndUA(t *testing.T) {
	routes := buildProbeBypassRoutes("alice_app_web", "cluster", []string{"/healthz", "/ready", "/", "healthz", "/healthz"})
	require.Len(t, routes, 2, "dedupe and drop root path")
	assert.Equal(t, "/healthz", routes[0].PathExact)
	assert.Equal(t, "/ready", routes[1].PathExact)
	for _, r := range routes {
		assert.Nil(t, r.ExtAuth, "probe bypass must not re-enable ExtAuth")
		assert.Nil(t, r.RequestHeaders, "probe must not stamp outbound X-BFL-USER")
		require.Len(t, r.HeaderMatches, 1)
		assert.Equal(t, "user-agent", r.HeaderMatches[0].Name)
		assert.Equal(t, probeUARegex, r.HeaderMatches[0].SafeRegex)
	}
}

func TestBuildAppVirtualHosts_ProbeBypassBeforeDefault(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)
	app := &message.AppInfo{
		Name:      "v1app",
		Appid:     "v1app",
		Namespace: "v1app-alice",
		Owner:     "alice",
		EntranceProbePaths: map[string][]string{
			"web": {"/healthz"},
		},
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "v1app-svc", Port: 8080},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)
	require.GreaterOrEqual(t, len(vhosts[0].Routes), 2)
	assert.Equal(t, "/healthz", vhosts[0].Routes[0].PathExact)
	assert.Nil(t, vhosts[0].Routes[0].ExtAuth)
	last := vhosts[0].Routes[len(vhosts[0].Routes)-1]
	assert.Equal(t, "/", last.PathPrefix)
	require.NotNil(t, last.ExtAuth)
	assert.Equal(t, autheliaVerifyPathPrefix, last.ExtAuth.PathPrefix)
}

// A "dev" entrance answers on both its normal hostname and a
// `<appid>-<port>.<zone>` alias, both routed to the same upstream cluster.
func TestBuildAppVirtualHosts_DevEntranceAlias(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := &message.AppInfo{
		Name:      "devapp",
		Appid:     "28f57292",
		Namespace: "devapp-alice",
		Owner:     "alice",
		Entrances: []*message.EntranceInfo{
			// A dev entrance's Host is the backing pod IP (set by proxylistener).
			{Name: "web", Host: "10.244.1.23", Port: 9999, Type: "dev"},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)

	assert.Contains(t, vhosts[0].Domains, "28f57292.alice.example.com", "primary hostname must remain")
	assert.Contains(t, vhosts[0].Domains, "28f57292-9999.alice.example.com", "dev alias must be added")

	// The alias shares the single upstream cluster of the entrance, which points
	// straight at the pod IP as a STATIC cluster.
	require.Len(t, vhosts[0].Routes, 1)
	cl := clusterSet[vhosts[0].Routes[0].Cluster]
	require.NotNil(t, cl)
	assert.Equal(t, "10.244.1.23", cl.Host, "dev entrance must route straight to the pod IP in Host")
	assert.False(t, cl.UseDNS, "dev entrance must be a STATIC cluster")
	assert.Equal(t, uint32(9999), cl.Port)
}

// A non-dev entrance uses the Kubernetes service DNS host (STRICT_DNS).
func TestBuildAppVirtualHosts_NonDevEntranceUsesDNS(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := &message.AppInfo{
		Name:      "app",
		Appid:     "28f57292",
		Namespace: "app-alice",
		Owner:     "alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "app-svc", Port: 9999},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)

	cl := clusterSet[vhosts[0].Routes[0].Cluster]
	require.NotNil(t, cl)
	assert.Equal(t, "app-svc.app-alice.svc.cluster.local", cl.Host)
	assert.True(t, cl.UseDNS)
}

// An entrance with an empty Host has no usable upstream and is skipped.
func TestBuildAppVirtualHosts_EmptyHostSkipped(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := &message.AppInfo{
		Name:      "devapp",
		Appid:     "1f47cd9b",
		Namespace: "devapp-alice",
		Owner:     "alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "", Port: 8080, Type: "dev"},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	assert.Empty(t, vhosts, "entrance with empty host must be skipped")
	assert.Empty(t, clusterSet)
}

// A non-dev entrance must NOT get the `<appid>-<port>.<zone>` alias.
func TestBuildAppVirtualHosts_NonDevEntranceNoAlias(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := &message.AppInfo{
		Name:      "app",
		Appid:     "28f57292",
		Namespace: "app-alice",
		Owner:     "alice",
		Entrances: []*message.EntranceInfo{
			{Name: "web", Host: "app-svc", Port: 9999},
		},
	}
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", clusterSet)
	require.Len(t, vhosts, 1)
	assert.NotContains(t, vhosts[0].Domains, "28f57292-9999.alice.example.com")
}

// Custom domains for shared apps are also open to every user.
func TestBuildCustomDomainVirtualHosts_SharedApp_OpenToAllUsers(t *testing.T) {
	for _, name := range []string{"admin", "alice"} {
		t.Run(name, func(t *testing.T) {
			tr := &Translator{cfg: &Config{}}
			user := &message.UserInfo{Name: name, Language: "en", Zone: name + ".example.com"}
			clusterSet := make(map[string]*ir.ClusterIR)

			vhosts := tr.buildCustomDomainVirtualHosts(user, makeSharedAppWithCustom(), clusterSet)
			require.Len(t, vhosts, 1)
			require.Len(t, vhosts[0].Routes, 1)
			assert.Nil(t, vhosts[0].Routes[0].DirectResponse, "no 403 gate on custom domain")
			assert.NotEmpty(t, vhosts[0].Routes[0].Cluster)
			assert.NotEmpty(t, clusterSet)
			assert.Equal(t, []string{"shareme.example.io"}, vhosts[0].Domains)
		})
	}
}

// ---------------------------------------------------------------------------
// buildFileserverRoutes
// ---------------------------------------------------------------------------

// Regression: a non-master node's scoped prefix (e.g.
// /api/resources/external/olares-work/) overlaps with the master node's
// generic catch-all prefix (/api/resources/external/). Envoy matches routes
// top-to-bottom and stops at the first match, so every node-scoped route MUST
// be emitted before the generic master routes — otherwise the master prefix
// swallows the request and misroutes it to the master node.
func TestBuildFileserverRoutes_NodeScopedBeforeMasterGeneric(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	// "olares" sorts first and is the master; "olares-work" is a worker node.
	// (buildFileserverRoutes relies on FileserverNodes being pre-sorted, which
	// is what triggered the original bug — the master's generic routes landed
	// ahead of the worker's scoped routes.)
	user := &message.UserInfo{
		Name: "testip162",
		FileserverNodes: []*message.FileserverNodeInfo{
			{NodeName: "olares", PodIP: "10.233.1.1", IsMaster: true},
			{NodeName: "olares-work", PodIP: "10.233.1.2", IsMaster: false},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	routes := tr.buildFileserverRoutes(user, clusterSet)

	indexOfPrefix := func(prefix string) int {
		for i, r := range routes {
			if r.PathPrefix == prefix {
				return i
			}
		}
		return -1
	}

	nodeExternal := indexOfPrefix("/api/resources/external/olares-work/")
	masterExternal := indexOfPrefix("/api/resources/external/")
	require.GreaterOrEqual(t, nodeExternal, 0, "olares-work node route must exist")
	require.GreaterOrEqual(t, masterExternal, 0, "master generic route must exist")
	assert.Less(t, nodeExternal, masterExternal,
		"node-scoped /api/resources/external/olares-work/ must be ordered before the generic /api/resources/external/")

	// The node-scoped route must target the worker node's cluster, and the
	// generic master route the master node's cluster.
	assert.Equal(t, "files_testip162_olares-work", routes[nodeExternal].Cluster)
	assert.Equal(t, "files_testip162_olares", routes[masterExternal].Cluster)
	assert.Equal(t, "olares-work", routes[nodeExternal].RequestHeaders["X-Terminus-Node"])
	_, hasOutboundBFL := routes[nodeExternal].RequestHeaders["X-BFL-USER"]
	assert.False(t, hasOutboundBFL, "fileserver must not stamp outbound zone-owner X-BFL-USER")
	require.NotNil(t, routes[nodeExternal].ExtAuth)
	assert.Equal(t, autheliaVerifyPathPrefix, routes[nodeExternal].ExtAuth.PathPrefix)
	require.NotNil(t, routes[masterExternal].ExtAuth)
	assert.Equal(t, autheliaVerifyPathPrefix, routes[masterExternal].ExtAuth.PathPrefix)
	_, hasMasterBFL := routes[masterExternal].RequestHeaders["X-BFL-USER"]
	assert.False(t, hasMasterBFL, "fileserver master route must not stamp outbound X-BFL-USER")

	// Every node-scoped prefix that overlaps a master prefix must precede ALL
	// master generic routes. Compute the first master-route index and assert
	// no node-scoped route appears after it.
	firstMaster := len(routes)
	for i, r := range routes {
		if strings.HasPrefix(r.Name, "files_master_") {
			firstMaster = i
			break
		}
	}
	for i, r := range routes {
		if strings.HasPrefix(r.Name, "files_node_") {
			assert.Less(t, i, firstMaster,
				"node-scoped route %q must come before any master generic route", r.Name)
		}
	}
}

func TestBuildMachineSystemVirtualHosts_SharedLogin(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusters := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildMachineSystemVirtualHosts("lab.example.com", "lab.olares.local", clusters)
	byName := map[string]*ir.VirtualHostIR{}
	for _, vh := range vhosts {
		byName[vh.Name] = vh
	}
	require.Contains(t, byName, "nonapp_auth")
	auth := byName["nonapp_auth"]
	assert.ElementsMatch(t, []string{"auth.lab.example.com", "auth.lab.olares.local"}, auth.Domains)
	assert.Nil(t, auth.Routes[0].ExtAuth)
	assert.Empty(t, auth.Routes[0].RequestHeaders["X-BFL-USER"])
	assert.Equal(t, "authelia-svc.os-frontend.svc.cluster.local", clusters[auth.Routes[0].Cluster].Host)
	require.Contains(t, byName, "nonapp_desktop")
	desktop := byName["nonapp_desktop"]
	require.NotNil(t, desktop.Routes[0].ExtAuth)
	assert.Empty(t, desktop.Routes[0].ExtAuth.User)
	assert.Empty(t, desktop.Routes[0].RequestHeaders["X-BFL-USER"])
	assert.Equal(t, "edge-desktop.os-frontend.svc.cluster.local", clusters[desktop.Routes[0].Cluster].Host)
	require.Contains(t, byName, "profile")
	profile := byName["profile"]
	require.NotNil(t, profile.Routes[0].ExtAuth)
	assert.Equal(t, "profile-service.os-frontend.svc.cluster.local", clusters[profile.Routes[0].Cluster].Host)
	assert.NotContains(t, byName, "nonapp_wizard")
}

// L4 selects the Files pod itself, without stamping the URL owner over the
// identity supplied by ExtAuth or asking the backend to dispatch by node header.
func TestBuildFileserverRoutesUseNodePodAndVerifiedIdentity(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "owner", FileserverNodes: []*message.FileserverNodeInfo{
		{NodeName: "olares-worker", PodIP: "10.233.1.1", IsMaster: true},
		{NodeName: "second-node", PodIP: "10.233.1.2"},
	}}
	clusters := make(map[string]*ir.ClusterIR)
	routes := tr.buildFileserverRoutes(user, clusters)
	for _, route := range routes {
		require.NotNil(t, route.ExtAuth, route.Name)
		assert.False(t, route.ExtAuth.Disabled, route.Name)
		assert.Equal(t, autheliaVerifyPathPrefix, route.ExtAuth.PathPrefix, route.Name)
		node := route.RequestHeaders["X-Terminus-Node"]
		assert.NotEmpty(t, node)
		assert.NotContains(t, route.RequestHeaders, "X-Provider-Proxy", route.Name)
		cluster := clusters[route.Cluster]
		require.NotNil(t, cluster, route.Name)
		assert.Equal(t, map[string]string{"olares-worker": "10.233.1.1", "second-node": "10.233.1.2"}[node], cluster.Host)
		assert.EqualValues(t, 80, cluster.Port)
		assert.False(t, cluster.UseDNS, "pod IP endpoints must use static discovery")
		assert.NotContains(t, route.RequestHeaders, "X-BFL-USER", "must not replace the authenticated user with the URL owner")
	}
	for _, prefix := range []string{"/api/task/olares-worker/", "/api/repos/"} {
		var found bool
		for _, route := range routes {
			if route.PathPrefix == prefix {
				found = true
				assert.Equal(t, "files_owner_olares-worker", route.Cluster)
				assert.NotContains(t, route.RequestHeaders, "X-Provider-Proxy")
			}
		}
		assert.True(t, found, prefix)
	}
}

func TestShareFileserverRoutesUseNodePodsAndExistingEntranceAuth(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "owner", FileserverNodes: []*message.FileserverNodeInfo{
		{NodeName: "master", PodIP: "10.233.1.1", IsMaster: true},
		{NodeName: "worker", PodIP: "10.233.1.2"},
	}}
	clusters := make(map[string]*ir.ClusterIR)
	routes := tr.buildShareFileserverRoutes(user, clusters)
	require.NotEmpty(t, routes)
	prefixes := make(map[string]*ir.HTTPRouteIR)
	for _, route := range routes {
		prefixes[route.PathPrefix] = route
		require.NotNil(t, route.ExtAuth)
		assert.Equal(t, buildAppExtAuthConfig(user), route.ExtAuth, "preserve the share entrance's existing verification rules")
		assert.NotContains(t, route.RequestHeaders, "X-BFL-USER", "do not replace the verified caller with the zone owner")
		assert.Equal(t, "%REQ(:AUTHORITY)%", route.RequestHeaders["X-Forwarded-Host"])
		assert.Equal(t, "", route.RequestHeaders["X-Internal-Share-Token"])
	}
	worker := prefixes["/upload/upload-link/worker/"]
	require.NotNil(t, worker)
	assert.Equal(t, "10.233.1.2", clusters[worker.Cluster].Host)
	require.NotNil(t, prefixes["/api/resources/share/"])
	assert.Equal(t, "10.233.1.1", clusters[prefixes["/api/resources/share/"].Cluster].Host)
	for _, path := range []string{"/api/resources/cache/worker/", "/api/resources/external/worker/", "/api/task/worker/", "/api/repos/", "/api/mount/worker/"} {
		assert.NotContains(t, prefixes, path, "private API must not become a share-entrance node route")
	}
}

func TestShareEntranceInstallsFileserverRoutes(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "owner", FileserverNodes: []*message.FileserverNodeInfo{{NodeName: "worker", PodIP: "10.233.1.2"}}}
	app := &message.AppInfo{Name: "share", Appid: "share", Namespace: "os-frontend", Owner: "owner", Entrances: []*message.EntranceInfo{{Name: "web", Host: "share", Port: 80}}}
	vhosts := tr.buildAppVirtualHosts(user, app, "owner.example.com", make(map[string]*ir.ClusterIR))
	require.NotEmpty(t, vhosts)
	found := false
	for _, vh := range vhosts {
		for _, route := range vh.Routes {
			if route.PathPrefix == "/upload/upload-link/worker/" {
				found = true
			}
		}
	}
	assert.True(t, found, "share entrance must install node routes before the frontend catch-all")
}
