package translator

import (
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// buildZoneWildcardSNI
// ---------------------------------------------------------------------------

func TestBuildZoneWildcardSNI(t *testing.T) {
	got := buildZoneWildcardSNI([]string{"alice.snowinning.com", "alice.olares.local"})
	assert.Equal(t, []string{
		"*.alice.olares.local",
		"*.alice.snowinning.com",
		"alice.olares.local",
		"alice.snowinning.com",
	}, got)
}

func TestBuildZoneWildcardSNI_Single(t *testing.T) {
	got := buildZoneWildcardSNI([]string{"alice.example.com"})
	assert.Equal(t, []string{"*.alice.example.com", "alice.example.com"}, got)
}

// ---------------------------------------------------------------------------
// collectAppIDs
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// buildUserHTTPSListeners
// ---------------------------------------------------------------------------

func TestBuildUserFilterChains_NoDeny_WithCustomDomainCert(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	// "myapp.custom.io" has its own TLS cert entry. It must NOT appear in
	// the main filter chain's SNI list because it gets its own dedicated
	// filter chain — if it appeared in both, Envoy would reject the listener
	// for "multiple filter chains with overlapping matching rules".
	user := &message.UserInfo{
		Name:              "alice",
		Zone:              "example.com",
		Namespace:         "user-space-alice",
		IsEphemeral:       false,
		DenyAll:           false,
		ServerNameDomains: []string{"alice.example.com"},
		SSL:               &message.SSLConfig{Zone: "alice.example.com", CertData: "cert", KeyData: "key"},
		CustomDomainCerts: []*message.CertInfo{
			{Domain: "myapp.custom.io", CertData: "custom-cert", KeyData: "custom-key"},
		},
		Apps: []*message.AppInfo{
			{
				Name:      "myapp",
				Appid:     "myapp",
				Namespace: "myapp-alice",
				Owner:     "alice",
				Entrances: []*message.EntranceInfo{
					{Name: "web", Host: "myapp-svc", Port: 8080},
				},
				Settings: map[string]string{
					"customDomain": `{"web":{"third_party_domain":"myapp.custom.io"}}`,
				},
			},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	listeners := tr.buildUserFilterChains(user, vhosts, 443, false)

	// One per-zone main chain + one custom-cert chain.
	// ServerNameDomains = ["alice.example.com"], so one zone chain.
	require.Len(t, listeners, 2)

	main := listeners[0]
	// Zone-domain-scoped name: https_{port}_{user}_{sanitized-zone}
	assert.Equal(t, "https_443_alice_alice_example_com", main.Name)
	// The custom domain must NOT be in the zone chain's SNI list.
	assert.NotContains(t, main.SNIMatches, "myapp.custom.io")

	custom := listeners[1]
	assert.Equal(t, []string{"myapp.custom.io"}, custom.SNIMatches)
}

func TestBuildUserFilterChains_NoDeny(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	user := &message.UserInfo{
		Name:              "alice",
		Zone:              "example.com",
		Namespace:         "user-space-alice",
		IsEphemeral:       false,
		DenyAll:           false,
		ServerNameDomains: []string{"alice.example.com"},
		SSL:               &message.SSLConfig{Zone: "alice.example.com", CertData: "cert", KeyData: "key"},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	listeners := tr.buildUserFilterChains(user, vhosts, 443, false)

	// One filter chain per zone domain.  ServerNameDomains = ["alice.example.com"],
	// so exactly one zone chain.
	require.Len(t, listeners, 1)
	// Name encodes the zone domain so adding a second zone creates a NEW filter
	// chain rather than modifying this one.
	assert.Equal(t, "https_443_alice_alice_example_com", listeners[0].Name)
	// SNI covers exactly the two patterns for this zone: zone + *.zone
	assert.Contains(t, listeners[0].SNIMatches, "alice.example.com")
	assert.Contains(t, listeners[0].SNIMatches, "*.alice.example.com")
	assert.Len(t, listeners[0].SNIMatches, 2)
	// Per-app domains are not enumerated (routes live in RDS, not filter chain SNI).
	assert.NotContains(t, listeners[0].SNIMatches, "auth.alice.example.com")
	assert.NotContains(t, listeners[0].SNIMatches, "desktop.alice.example.com")
	assert.Empty(t, listeners[0].SourceCIDRs)
	assert.Equal(t, "alice", listeners[0].UserName)
	assert.NotNil(t, listeners[0].TLSCert)
}

func TestBuildUserFilterChains_NoDeny_MultiZone(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	// User with two zone domains (primary + local).
	user := &message.UserInfo{
		Name:              "alice",
		Zone:              "snowinning.com",
		Namespace:         "user-space-alice",
		IsEphemeral:       false,
		DenyAll:           false,
		ServerNameDomains: []string{"alice.snowinning.com", "alice.olares.local"},
		SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert", KeyData: "key"},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	listeners := tr.buildUserFilterChains(user, vhosts, 443, false)

	// One filter chain per zone → two chains.
	require.Len(t, listeners, 2)

	// Each chain has exactly 2 SNI entries (zone + *.zone) and a stable name.
	names := []string{listeners[0].Name, listeners[1].Name}
	assert.Contains(t, names, "https_443_alice_alice_snowinning_com")
	assert.Contains(t, names, "https_443_alice_alice_olares_local")

	for _, l := range listeners {
		assert.Len(t, l.SNIMatches, 2, "each zone chain must have exactly 2 SNI entries")
		assert.NotNil(t, l.TLSCert)
		assert.Equal(t, "https_443_alice_routes", l.Name[:len("https_443_alice")]+"_routes",
			"all zone chains must share the same route-config name prefix")
	}
}

// TestBuildUserHTTPSListeners_NoDeny_AddZoneNoExistingDrain verifies the core
// invariant: adding a third zone domain to a user leaves the first two filter
// chains byte-for-byte identical (Envoy in-place update, no drain).
func TestBuildUserFilterChains_NoDeny_AddZoneNoExistingDrain(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	makeUser := func(domains []string) *message.UserInfo {
		return &message.UserInfo{
			Name:              "alice",
			Zone:              "snowinning.com",
			Namespace:         "user-space-alice",
			IsEphemeral:       false,
			DenyAll:           false,
			ServerNameDomains: domains,
			SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert", KeyData: "key"},
		}
	}

	buildChains := func(user *message.UserInfo) []*ir.HTTPListenerIR {
		clusterSet := make(map[string]*ir.ClusterIR)
		vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
		tr.applyDenyAllRestrictions(user, vhosts, nil)
		return tr.buildUserFilterChains(user, vhosts, 443, false)
	}

	before := buildChains(makeUser([]string{"alice.snowinning.com", "alice.olares.local"}))
	after := buildChains(makeUser([]string{"alice.snowinning.com", "alice.olares.local", "alice.newzone.com"}))

	// After: one extra filter chain added.
	require.Len(t, before, 2)
	require.Len(t, after, 3)

	// Find each existing chain by name and compare.
	findByName := func(ls []*ir.HTTPListenerIR, name string) *ir.HTTPListenerIR {
		for _, l := range ls {
			if l.Name == name {
				return l
			}
		}
		return nil
	}

	for _, name := range []string{"https_443_alice_alice_snowinning_com", "https_443_alice_alice_olares_local"} {
		b := findByName(before, name)
		a := findByName(after, name)
		require.NotNil(t, b, "filter chain %q must exist before zone addition", name)
		require.NotNil(t, a, "filter chain %q must exist after zone addition", name)
		assert.Equal(t, b.SNIMatches, a.SNIMatches, "SNI for %q must not change", name)
		assert.Equal(t, b.TLSCert, a.TLSCert, "TLS cert ref for %q must not change", name)
	}

	// The new zone gets its own chain.
	newChain := findByName(after, "https_443_alice_alice_newzone_com")
	require.NotNil(t, newChain, "new zone chain must be present after addition")
	assert.Equal(t, []string{"*.alice.newzone.com", "alice.newzone.com"}, newChain.SNIMatches)
}

func TestBuildUserFilterChains_DenyAll_WithAllowedDomains(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	// "app.bob.example.com" is in AllowedDomains.  The allowed VH (app)
	// must NOT have SourceCIDRs; the restricted VHs (profile, auth,
	// desktop, wizard) must have SourceCIDRs.  The filter chain itself
	// uses a stable zone wildcard — no per-domain SNI split.
	user := &message.UserInfo{
		Name:              "bob",
		Zone:              "bob.example.com",
		Namespace:         "user-space-bob",
		IsEphemeral:       false,
		DenyAll:           true,
		AllowedDomains:    []string{"app.bob.example.com"},
		LocalDomainIP:     "192.168.1.100",
		ServerNameDomains: []string{"bob.example.com"},
		SSL:               &message.SSLConfig{Zone: "bob.example.com", CertData: "cert", KeyData: "key"},
		Apps: []*message.AppInfo{
			{
				Name:      "app",
				Appid:     "app",
				Namespace: "app-bob",
				Owner:     "bob",
				Entrances: []*message.EntranceInfo{
					{Name: "main", Host: "app-svc", Port: 8080},
				},
			},
		},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	listeners := tr.buildUserFilterChains(user, vhosts, 443, false)

	// One filter chain per zone domain (same as non-deny_all).
	require.Len(t, listeners, 1)
	l := listeners[0]
	assert.Equal(t, "https_443_bob_bob_example_com", l.Name)
	// Stable zone-wildcard SNI — doesn't change when apps are installed.
	assert.Contains(t, l.SNIMatches, "bob.example.com")
	assert.Contains(t, l.SNIMatches, "*.bob.example.com")

	// Access control is on VHs, not on the filter chain.
	assert.Empty(t, l.SourceCIDRs)

	// The VH for the allowed app must NOT have SourceCIDRs.
	var appVH *ir.VirtualHostIR
	var restrictedVHs []*ir.VirtualHostIR
	for _, vh := range l.VirtualHosts {
		hasAllowedDomain := false
		for _, d := range vh.Domains {
			if d == "app.bob.example.com" {
				hasAllowedDomain = true
				break
			}
		}
		if hasAllowedDomain {
			appVH = vh
		} else if len(vh.SourceCIDRs) > 0 {
			restrictedVHs = append(restrictedVHs, vh)
		}
	}

	require.NotNil(t, appVH, "VH for allowed app must exist")
	assert.Empty(t, appVH.SourceCIDRs, "allowed VH must not have SourceCIDRs")

	require.NotEmpty(t, restrictedVHs, "there must be restricted VHs")
	for _, vh := range restrictedVHs {
		assert.Contains(t, vh.SourceCIDRs, vpnCIDR, "restricted VH %q must include VPN CIDR", vh.Name)
		assert.Contains(t, vh.SourceCIDRs, "192.168.1.100/32", "restricted VH %q must include LocalDomainIP", vh.Name)
	}
}

func TestBuildUserFilterChains_DenyAll_NoAllowedDomains(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	user := &message.UserInfo{
		Name:              "carol",
		Zone:              "carol.example.com",
		Namespace:         "user-space-carol",
		IsEphemeral:       false,
		DenyAll:           true,
		AllowedDomains:    nil,
		LocalDomainIP:     "",
		ServerNameDomains: []string{"carol.example.com"},
		SSL:               &message.SSLConfig{Zone: "carol.example.com", CertData: "cert", KeyData: "key"},
	}

	clusterSet := make(map[string]*ir.ClusterIR)
	vhosts := tr.buildUserVirtualHosts(user, user.SSL.Zone, false, clusterSet)
	tr.applyDenyAllRestrictions(user, vhosts, nil)
	listeners := tr.buildUserFilterChains(user, vhosts, 443, false)

	// One filter chain per zone domain.
	require.Len(t, listeners, 1)
	l := listeners[0]
	assert.Equal(t, "https_443_carol_carol_example_com", l.Name)
	// No SourceCIDRs on the filter chain itself.
	assert.Empty(t, l.SourceCIDRs)

	// All VHs should be restricted (no AllowedDomains).
	for _, vh := range l.VirtualHosts {
		assert.Contains(t, vh.SourceCIDRs, vpnCIDR,
			"VH %q must have VPN CIDR restriction when no AllowedDomains", vh.Name)
	}
}

// ---------------------------------------------------------------------------
// Translate (full integration)
// ---------------------------------------------------------------------------

func TestTranslate_Full(t *testing.T) {
	tr := &Translator{cfg: &Config{
		SSLServerPort:      443,
		SSLProxyServerPort: 444,
	}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Zone:              "example.com",
				Namespace:         "user-space-alice",
				IsEphemeral:       false,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.example.com"},
				SSL:               &message.SSLConfig{Zone: "alice.example.com", CertData: "cert", KeyData: "key"},
				Apps: []*message.AppInfo{
					{
						Name:      "vault",
						Appid:     "vault",
						Owner:     "alice",
						Namespace: "vault-alice",
						Entrances: []*message.EntranceInfo{
							{Name: "main", Host: "vault-svc", Port: 8080},
						},
						Ports: []*message.PortInfo{
							{Name: "tcp-8080", Host: "vault-svc", Port: 8080, ExposePort: 48126, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}

	xds := tr.Translate(resources)

	// Expect: HTTP redirect(81) as Listener + HTTPS on 443 and 444 as HTTPListeners + TCP stream as StreamListener
	require.Len(t, xds.Listeners, 1)
	httpListener := xds.Listeners[0]
	assert.Equal(t, "http_redirect_81", httpListener.Name)
	assert.Equal(t, ir.ProtocolHTTP, httpListener.Protocol)

	require.Len(t, xds.HTTPListeners, 2)
	https443 := xds.HTTPListeners[0]
	// Per-zone naming: https_{port}_{user}_{sanitized-zone-domain}
	assert.Equal(t, "https_443_alice_alice_example_com", https443.Name)
	assert.True(t, https443.TLS)
	assert.False(t, https443.ProxyProtocol)

	https444 := xds.HTTPListeners[1]
	assert.True(t, https444.ProxyProtocol)

	require.Len(t, xds.StreamListeners, 1)
	stream := xds.StreamListeners[0]
	assert.Equal(t, "stream_tcp_alice_48126", stream.Name)
	assert.Equal(t, uint32(48126), stream.Port)
}

// ---------------------------------------------------------------------------
// buildHTTPRedirectListener
// ---------------------------------------------------------------------------

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

func TestToLocalDomain(t *testing.T) {
	assert.Equal(t, "vault.alice.olares.local", toLocalDomain("vault.alice.snowinning.com"))
	assert.Equal(t, "simple", toLocalDomain("simple"))
}

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
			// IsShared=true on the EntranceInfo signals that this entrance
			// participates in gateway-mode rewriting (mirrors what
			// provider.buildAppInfos sets for Application.Spec.SharedEntrances).
			{Name: "web", Host: "shareme-svc", Port: 8080, IsShared: true},
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

// Any user (admin or normal) reaches the real upstream of a v3 / shared app.
func TestBuildAppVirtualHosts_SharedApp_OpenToAllUsers(t *testing.T) {
	for _, name := range []string{"admin", "alice"} {
		t.Run(name, func(t *testing.T) {
			tr := &Translator{cfg: &Config{}}
			user := &message.UserInfo{Name: name, Language: "en"}
			clusterSet := make(map[string]*ir.ClusterIR)

			vhosts := tr.buildAppVirtualHosts(user, makeSharedApp(), name+".example.com", false, clusterSet)
			require.Len(t, vhosts, 1)
			require.Len(t, vhosts[0].Routes, 1)
			r := vhosts[0].Routes[0]
			assert.Nil(t, r.DirectResponse, "no 403 gate is emitted; shared apps are open")
			assert.NotEmpty(t, r.Cluster, "every user must reach the upstream cluster")
			assert.NotEmpty(t, clusterSet, "an upstream cluster must be registered")
		})
	}
}

// v1/v2 (non-shared) apps continue to reach their upstream cluster directly.
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
	vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	assert.Nil(t, vhosts[0].Routes[0].DirectResponse)
	assert.NotEmpty(t, clusterSet)
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
// gateway/direct route-mode switching
// ---------------------------------------------------------------------------

// E1: a shared app annotated with route-mode=gateway must rewrite its upstream
// cluster to the shared Envoy Gateway data-plane Service (app-gateway-data:80),
// while non-annotated shared apps keep the direct Service upstream.
func TestBuildAppVirtualHosts_SharedApp_GatewayModeSwitchesUpstream(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en"}

	t.Run("direct mode keeps service upstream", func(t *testing.T) {
		clusterSet := make(map[string]*ir.ClusterIR)
		app := makeSharedApp()
		vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
		require.Len(t, vhosts, 1)
		require.Len(t, vhosts[0].Routes, 1)
		cluster := clusterSet[vhosts[0].Routes[0].Cluster]
		require.NotNil(t, cluster)
		assert.Equal(t, "shareme-svc.shareme-shared.svc.cluster.local", cluster.Host)
		assert.Equal(t, uint32(8080), cluster.Port)
	})

	t.Run("gateway mode rewrites to app-gateway-data", func(t *testing.T) {
		clusterSet := make(map[string]*ir.ClusterIR)
		app := makeSharedApp()
		app.Annotations = map[string]string{"gateway.olares.io/route-mode": "gateway"}
		vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
		require.Len(t, vhosts, 1)
		require.Len(t, vhosts[0].Routes, 1)
		cluster := clusterSet[vhosts[0].Routes[0].Cluster]
		require.NotNil(t, cluster)
		assert.Equal(t, "app-gateway-data.app-gateway.svc.cluster.local", cluster.Host)
		assert.Equal(t, uint32(80), cluster.Port)
	})

	t.Run("non-shared app ignores annotation", func(t *testing.T) {
		clusterSet := make(map[string]*ir.ClusterIR)
		app := &message.AppInfo{
			Name:      "v1app",
			Appid:     "v1app",
			Namespace: "v1app-alice",
			Owner:     "alice",
			IsShared:  false,
			Annotations: map[string]string{
				"gateway.olares.io/route-mode": "gateway",
			},
			Entrances: []*message.EntranceInfo{
				{Name: "web", Host: "v1app-svc", Port: 8080},
			},
		}
		vhosts := tr.buildAppVirtualHosts(user, app, "alice.example.com", false, clusterSet)
		require.Len(t, vhosts, 1)
		cluster := clusterSet[vhosts[0].Routes[0].Cluster]
		require.NotNil(t, cluster)
		assert.Equal(t, "v1app-svc.v1app-alice.svc.cluster.local", cluster.Host)
		assert.Equal(t, uint32(8080), cluster.Port)
	})
}

// E2: shared app with custom domain in gateway mode also flips the upstream
// cluster, but keeps the custom domain in the virtual host (the Host header
// reaches the HTTPRoute unchanged so EG can match it).
func TestBuildCustomDomainVirtualHosts_SharedApp_GatewayMode(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	user := &message.UserInfo{Name: "alice", Language: "en", Zone: "alice.example.com"}
	clusterSet := make(map[string]*ir.ClusterIR)

	app := makeSharedAppWithCustom()
	app.Annotations = map[string]string{"gateway.olares.io/route-mode": "gateway"}

	vhosts := tr.buildCustomDomainVirtualHosts(user, app, clusterSet)
	require.Len(t, vhosts, 1)
	require.Len(t, vhosts[0].Routes, 1)
	cluster := clusterSet[vhosts[0].Routes[0].Cluster]
	require.NotNil(t, cluster)
	assert.Equal(t, "app-gateway-data.app-gateway.svc.cluster.local", cluster.Host)
	assert.Equal(t, uint32(80), cluster.Port)
	assert.Equal(t, []string{"shareme.example.io"}, vhosts[0].Domains, "custom domain stays so EG HTTPRoute matches")
}

// per-viewer hostname helpers (<hash8>.<viewer>.<platformDomain>).

func TestSharedEntranceHostPrefix_StableAndCaseInsensitive(t *testing.T) {
	a := sharedEntranceHostPrefix("a5be2268", "ollamav2")
	b := sharedEntranceHostPrefix("A5BE2268", " OllamaV2 ")
	if a != b {
		t.Fatalf("hash8 not normalized: %q vs %q", a, b)
	}
	if len(a) != 8 {
		t.Fatalf("hash8 wrong length: %q", a)
	}
}

func TestPlatformDomainFromZone(t *testing.T) {
	cases := []struct {
		zone, viewer, want string
	}{
		{"brucedai.olares.com", "brucedai", "olares.com"},
		{"BruceDai.Olares.com", "brucedai", "olares.com"},
		{"alice.olares.com", "bob", ""},
		{"olares.com", "brucedai", ""},
		{"", "brucedai", ""},
		{"brucedai.olares.com", "", ""},
		{"brucedai.", "brucedai", ""},
	}
	for _, tc := range cases {
		if got := platformDomainFromZone(tc.zone, tc.viewer); got != tc.want {
			t.Fatalf("platformDomainFromZone(%q,%q) = %q, want %q", tc.zone, tc.viewer, got, tc.want)
		}
	}
}

func TestGatewayV2EntranceHostname(t *testing.T) {
	app := &message.AppInfo{Name: "ollamaserver", Appid: "a5be2268", IsShared: true,
		Annotations: map[string]string{"gateway.olares.io/route-mode": "gateway"}}
	got := gatewayV2EntranceHostname(app, "ollamav2", "brucedai", "brucedai.olares.com")
	want := sharedEntranceHostPrefix("a5be2268", "ollamav2") + ".brucedai.olares.com"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// Different viewer changes only the middle label.
	got2 := gatewayV2EntranceHostname(app, "ollamav2", "alice", "alice.olares.com")
	want2 := sharedEntranceHostPrefix("a5be2268", "ollamav2") + ".alice.olares.com"
	if got2 != want2 {
		t.Fatalf("alice host: got %q want %q", got2, want2)
	}
	// Missing appid falls back to app.Name.
	app2 := &message.AppInfo{Name: "ollamaserver"}
	if got := gatewayV2EntranceHostname(app2, "ollamav2", "brucedai", "brucedai.olares.com"); got == "" {
		t.Fatalf("expected non-empty fallback hostname, got %q", got)
	}
	// Missing zone returns "".
	if got := gatewayV2EntranceHostname(app, "ollamav2", "brucedai", "alice.olares.com"); got != "" {
		t.Fatalf("zone/viewer mismatch must return empty: %q", got)
	}
}

// E4: in gateway mode, every viewer of a shared app receives a vhost whose
// primary domain is <hash8>.<viewer>.<platformDomain>; the upstream cluster
// already flips to app-gateway-data (E1).
func TestBuildAppVirtualHosts_SharedApp_GatewayMode_PerViewerHost(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	app := makeSharedApp()
	app.Appid = "a5be2268"
	app.Entrances = []*message.EntranceInfo{{Name: "ollamav2", Host: "shareme-svc", Port: 8080, IsShared: true}}
	app.Annotations = map[string]string{"gateway.olares.io/route-mode": "gateway"}

	expectedHash := sharedEntranceHostPrefix("a5be2268", "ollamav2")

	for _, viewer := range []string{"brucedai", "alice", "bob"} {
		t.Run(viewer, func(t *testing.T) {
			user := &message.UserInfo{Name: viewer, Language: "en"}
			zone := viewer + ".olares.com"
			vhosts := tr.buildAppVirtualHosts(user, app, zone, false, map[string]*ir.ClusterIR{})
			require.Len(t, vhosts, 1)
			want := expectedHash + "." + viewer + ".olares.com"
			require.Equal(t, want, vhosts[0].Domains[0])
		})
	}
}

// E5: gateway mode but missing appid AND missing app.Name falls back to the
// legacy hostname instead of producing an empty / malformed virtual host.
func TestBuildAppVirtualHosts_SharedApp_GatewayMode_FallbackHostname(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	app := &message.AppInfo{
		Name: "", Appid: "", IsShared: true, Namespace: "x-shared",
		Annotations: map[string]string{"gateway.olares.io/route-mode": "gateway"},
		Entrances:   []*message.EntranceInfo{{Name: "api", Host: "svc", Port: 8080, IsShared: true}},
	}
	user := &message.UserInfo{Name: "brucedai", Language: "en"}
	vhosts := tr.buildAppVirtualHosts(user, app, "brucedai.olares.com", false, map[string]*ir.ClusterIR{})
	require.Len(t, vhosts, 1)
	// Must NOT begin with hash8 (since we cannot compute one); falls back to
	// the legacy "<resolved-prefix>.<zone>" form.
	require.NotContains(t, vhosts[0].Domains[0], ".*.")
}

// E6: v2 cluster-scoped gateway-mode apps that mix per-user entrances with
// SharedEntrances must keep the per-user entrances on the legacy
// <appid><idx>.<zone> direct path and only rewrite the SharedEntrances to
// <hash8>.<viewer>.<domain> through the EG data plane. This guards against
// the regression that broke ollamav2's management terminal URL when the
// gateway pilot landed.
func TestBuildAppVirtualHosts_GatewayMode_MixedEntrances_PerUserKeptLegacy(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	app := &message.AppInfo{
		Name:      "ollamav2",
		Appid:     "a5be2268",
		Namespace: "ollamav2-brucedai",
		Owner:     "brucedai",
		IsShared:  true,
		Annotations: map[string]string{
			"gateway.olares.io/route-mode": "gateway",
		},
		Entrances: []*message.EntranceInfo{
			// Per-user entrance, must keep legacy <appid><idx>.<zone>:
			{Name: "terminal", Host: "terminalclient", Port: 8081, AuthLevel: "private"},
			{Name: "ollamaclient", Host: "ollamaclient", Port: 8080, AuthLevel: "internal"},
			// SharedEntrance, must use <hash8>.<viewer>.<domain> via EG:
			{Name: "ollamav2", Host: "sharedentrances-ollama", Port: 80, AuthLevel: "internal", IsShared: true},
		},
	}
	user := &message.UserInfo{Name: "brucedai", Language: "en"}
	zone := "brucedai.olares.com"
	clusterSet := map[string]*ir.ClusterIR{}

	vhosts := tr.buildAppVirtualHosts(user, app, zone, false, clusterSet)
	require.Len(t, vhosts, 3)

	byName := map[string]*ir.VirtualHostIR{}
	for _, v := range vhosts {
		byName[v.Name] = v
	}

	// 1. terminal: index 0 of 3 entrances → "a5be22680.brucedai.olares.com",
	//    upstream is the per-user Service (not the EG data plane).
	term := byName["app_brucedai_ollamav2_terminal"]
	require.NotNil(t, term)
	assert.Equal(t, "a5be22680.brucedai.olares.com", term.Domains[0])
	require.NotNil(t, clusterSet["app_brucedai_ollamav2_terminal"])
	assert.Equal(t, "terminalclient.ollamav2-brucedai.svc.cluster.local",
		clusterSet["app_brucedai_ollamav2_terminal"].Host)
	assert.Equal(t, uint32(8081), clusterSet["app_brucedai_ollamav2_terminal"].Port)

	// 2. ollamaclient: index 1 → "a5be22681...", also direct upstream.
	oc := byName["app_brucedai_ollamav2_ollamaclient"]
	require.NotNil(t, oc)
	assert.Equal(t, "a5be22681.brucedai.olares.com", oc.Domains[0])
	assert.Equal(t, "ollamaclient.ollamav2-brucedai.svc.cluster.local",
		clusterSet["app_brucedai_ollamav2_ollamaclient"].Host)

	// 3. ollamav2 (sharedEntrance): hash8 host AND upstream pointing at the
	//    shared EG data plane.
	wantSharedHost := sharedEntranceHostPrefix("a5be2268", "ollamav2") + ".brucedai.olares.com"
	shared := byName["app_brucedai_ollamav2_ollamav2"]
	require.NotNil(t, shared)
	assert.Equal(t, wantSharedHost, shared.Domains[0])
	assert.Equal(t, gatewayDataPlaneHost, clusterSet["app_brucedai_ollamav2_ollamav2"].Host)
	assert.Equal(t, uint32(gatewayDataPlanePort), clusterSet["app_brucedai_ollamav2_ollamav2"].Port)
}

// E7: when an app has only per-user entrances and no SharedEntrances, even
// the route-mode=gateway annotation must be ignored — every entrance keeps
// the legacy URL + direct upstream.
func TestBuildAppVirtualHosts_GatewayMode_OnlyPerUser_NoRewrite(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	app := &message.AppInfo{
		Name: "ollamav2", Appid: "a5be2268", Namespace: "ollamav2-brucedai",
		Owner: "brucedai", IsShared: true,
		Annotations: map[string]string{"gateway.olares.io/route-mode": "gateway"},
		Entrances: []*message.EntranceInfo{
			{Name: "terminal", Host: "terminalclient", Port: 8081, AuthLevel: "private"},
		},
	}
	user := &message.UserInfo{Name: "brucedai", Language: "en"}
	clusterSet := map[string]*ir.ClusterIR{}

	vhosts := tr.buildAppVirtualHosts(user, app, "brucedai.olares.com", false, clusterSet)
	require.Len(t, vhosts, 1)
	assert.Equal(t, "a5be2268.brucedai.olares.com", vhosts[0].Domains[0])
	assert.Equal(t, "terminalclient.ollamav2-brucedai.svc.cluster.local",
		clusterSet["app_brucedai_ollamav2_terminal"].Host)
}

// E3: the helper isGatewayMode is the single source of truth for the
// route-mode decision and must reject malformed / missing annotations.
func TestIsGatewayMode(t *testing.T) {
	cases := []struct {
		name string
		app  *message.AppInfo
		want bool
	}{
		{"nil app", nil, false},
		{"non-shared with annotation", &message.AppInfo{IsShared: false, Annotations: map[string]string{"gateway.olares.io/route-mode": "gateway"}}, false},
		{"shared without annotations", &message.AppInfo{IsShared: true}, false},
		{"shared with direct mode", &message.AppInfo{IsShared: true, Annotations: map[string]string{"gateway.olares.io/route-mode": "direct"}}, false},
		{"shared with empty mode", &message.AppInfo{IsShared: true, Annotations: map[string]string{"gateway.olares.io/route-mode": ""}}, false},
		{"shared with wrong key", &message.AppInfo{IsShared: true, Annotations: map[string]string{"gateway.olares.io/mode": "gateway"}}, false},
		{"shared with gateway mode", &message.AppInfo{IsShared: true, Annotations: map[string]string{"gateway.olares.io/route-mode": "gateway"}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isGatewayMode(tc.app))
		})
	}
}
