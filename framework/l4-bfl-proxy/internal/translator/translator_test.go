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
