package translator

import (
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/ir"
	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// buildSNIMatches
// ---------------------------------------------------------------------------

func TestBuildSNIMatches_AdminUser(t *testing.T) {
	user := &message.UserInfo{
		Name:              "alice",
		Zone:              "snowinning.com",
		IsEphemeral:       false,
		ServerNameDomains: []string{"alice.snowinning.com"},
	}
	allAppIDs := []string{"app1", "app2"}

	got := buildSNIMatches(user, allAppIDs)

	assert.Equal(t, []string{
		"alice.snowinning.com",
		"*.alice.snowinning.com",
	}, got)
}

func TestBuildSNIMatches_AdminMultipleDomains(t *testing.T) {
	user := &message.UserInfo{
		Name:        "bob",
		Zone:        "example.com",
		IsEphemeral: false,
		ServerNameDomains: []string{
			"bob.example.com",
			"bob.example.com.olares.local",
		},
	}

	got := buildSNIMatches(user, nil)

	assert.Equal(t, []string{
		"bob.example.com",
		"*.bob.example.com",
		"bob.example.com.olares.local",
		"*.bob.example.com.olares.local",
	}, got)
}

func TestBuildSNIMatches_EphemeralUser(t *testing.T) {
	user := &message.UserInfo{
		Name:        "tempuser",
		Zone:        "snowinning.com",
		IsEphemeral: true,
	}
	allAppIDs := []string{"vault", "desktop"}

	got := buildSNIMatches(user, allAppIDs)

	assert.Equal(t, []string{
		"wizard-tempuser.snowinning.com",
		"wizard-tempuser.snowinning.local",
	}, got)
}

func TestBuildSNIMatches_EphemeralNoApps(t *testing.T) {
	user := &message.UserInfo{
		Name:        "tempuser",
		Zone:        "snowinning.com",
		IsEphemeral: true,
	}

	got := buildSNIMatches(user, nil)
	assert.Equal(t, []string{
		"wizard-tempuser.snowinning.com",
		"wizard-tempuser.snowinning.local",
	}, got)
}

// ---------------------------------------------------------------------------
// collectAppIDs
// ---------------------------------------------------------------------------

func TestCollectAppIDs(t *testing.T) {
	tests := []struct {
		name string
		apps []*message.AppInfo
		want []string
	}{
		{
			name: "single entrance per app",
			apps: []*message.AppInfo{
				{Appid: "vault", Entrances: []message.EntranceInfo{{Name: "main"}}},
				{Appid: "desktop", Entrances: []message.EntranceInfo{{Name: "main"}}},
			},
			want: []string{"desktop", "vault"},
		},
		{
			name: "multiple entrances",
			apps: []*message.AppInfo{
				{Appid: "myapp", Entrances: []message.EntranceInfo{
					{Name: "web"},
					{Name: "api"},
				}},
			},
			want: []string{"myapp0", "myapp1"},
		},
		{
			name: "dedup across apps",
			apps: []*message.AppInfo{
				{Appid: "vault", Entrances: []message.EntranceInfo{{Name: "main"}}},
				{Appid: "vault", Entrances: []message.EntranceInfo{{Name: "main"}}},
			},
			want: []string{"vault"},
		},
		{
			name: "no apps",
			apps: nil,
			want: nil,
		},
		{
			name: "app with no entrances",
			apps: []*message.AppInfo{
				{Appid: "empty", Entrances: nil},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectAppIDs(tt.apps)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// buildUserRoutes
// ---------------------------------------------------------------------------

func TestBuildUserRoutes_NoDeny(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	user := &message.UserInfo{
		Name:              "alice",
		Zone:              "example.com",
		IsEphemeral:       false,
		BFLHost:           "bfl.user-space-alice",
		BFLPort:           444,
		DenyAll:           false,
		ServerNameDomains: []string{"alice.example.com"},
	}

	routes := tr.buildUserRoutes(user, []string{"vault"})

	require.Len(t, routes, 1)
	assert.Equal(t, "route_alice", routes[0].Name)
	assert.Contains(t, routes[0].SNIMatches, "alice.example.com")
	assert.Contains(t, routes[0].SNIMatches, "*.alice.example.com")
	assert.Empty(t, routes[0].SourcePrefixRanges)
	assert.Equal(t, "bfl.user-space-alice", routes[0].Destination.Host)
	assert.Equal(t, uint32(444), routes[0].Destination.Port)
	assert.True(t, routes[0].ProxyProtocolUpstream)
}

func TestBuildUserRoutes_DenyAll_WithAllowedDomains(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	user := &message.UserInfo{
		Name:              "bob",
		Zone:              "example.com",
		IsEphemeral:       false,
		BFLHost:           "bfl.user-space-bob",
		BFLPort:           444,
		DenyAll:           true,
		AllowedDomains:    []string{"app.bob.example.com"},
		LocalDomainIP:     "192.168.1.100",
		ServerNameDomains: []string{"bob.example.com"},
	}

	routes := tr.buildUserRoutes(user, nil)

	// Two routes: one for allowed_domains (any source), one for all domains (restricted source).
	require.Len(t, routes, 2)

	allowed := routes[0]
	assert.Equal(t, "route_bob_allowed", allowed.Name)
	assert.Equal(t, []string{"app.bob.example.com"}, allowed.SNIMatches)
	assert.Empty(t, allowed.SourcePrefixRanges)

	restricted := routes[1]
	assert.Equal(t, "route_bob_restricted", restricted.Name)
	assert.Contains(t, restricted.SNIMatches, "bob.example.com")
	assert.Contains(t, restricted.SourcePrefixRanges, vpnCIDR)
	assert.Contains(t, restricted.SourcePrefixRanges, "192.168.1.100/32")
}

func TestBuildUserRoutes_DenyAll_NoAllowedDomains(t *testing.T) {
	tr := &Translator{cfg: &Config{SSLServerPort: 443}}

	user := &message.UserInfo{
		Name:              "carol",
		Zone:              "example.com",
		IsEphemeral:       false,
		BFLHost:           "bfl.user-space-carol",
		BFLPort:           444,
		DenyAll:           true,
		AllowedDomains:    nil,
		LocalDomainIP:     "",
		ServerNameDomains: []string{"carol.example.com"},
	}

	routes := tr.buildUserRoutes(user, nil)

	// Only one restricted route (no allowed_domains route).
	require.Len(t, routes, 1)
	assert.Equal(t, "route_carol_restricted", routes[0].Name)
	assert.Equal(t, []string{vpnCIDR}, routes[0].SourcePrefixRanges)
}

// ---------------------------------------------------------------------------
// translate (full integration)
// ---------------------------------------------------------------------------

func TestTranslate_Full(t *testing.T) {
	tr := &Translator{cfg: &Config{
		SSLServerPort:       443,
		SSLProxyServerPort:  444,
		UserNamespacePrefix: "user-space",
	}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Zone:              "example.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-alice",
				BFLPort:           444,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.example.com"},
			},
		},
		Apps: []*message.AppInfo{
			{
				Name:  "vault",
				Appid: "vault",
				Owner: "alice",
				Entrances: []message.EntranceInfo{
					{Name: "main"},
				},
				Ports: []message.PortInfo{
					{Name: "tcp-8080", ExposePort: 48126, Protocol: "tcp"},
				},
			},
		},
	}

	xds := tr.Translate(resources)

	// Expect: HTTP redirect(81) + TLS(443) + TLS(444) + TCP stream(48126)
	require.Len(t, xds.Listeners, 4)

	httpListener := xds.Listeners[0]
	assert.Equal(t, "http_redirect_81", httpListener.Name)
	assert.Equal(t, ir.ProtocolHTTP, httpListener.Protocol)
	assert.Equal(t, uint32(81), httpListener.Port)
	assert.NotNil(t, httpListener.HTTPRedirect)

	tls443 := xds.Listeners[1]
	assert.Equal(t, "tls_443", tls443.Name)
	assert.Equal(t, ir.ProtocolTLS, tls443.Protocol)
	assert.Equal(t, uint32(443), tls443.Port)
	assert.False(t, tls443.ProxyProtocol)
	assert.True(t, tls443.TLSInspector)
	require.Len(t, tls443.Routes, 1)
	assert.Equal(t, "route_alice", tls443.Routes[0].Name)

	tls444 := xds.Listeners[2]
	assert.Equal(t, "tls_444", tls444.Name)
	assert.True(t, tls444.ProxyProtocol)
	require.Len(t, tls444.Routes, 1)

	stream := xds.Listeners[3]
	assert.Equal(t, "stream_tcp_48126", stream.Name)
	assert.Equal(t, ir.ProtocolTCP, stream.Protocol)
	assert.Equal(t, uint32(48126), stream.Port)
	require.Len(t, stream.Routes, 1)
	assert.Equal(t, "bfl.user-space-alice", stream.Routes[0].Destination.Host)
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
			{Name: "alice", BFLHost: "bfl.user-space-alice"},
		},
		Apps: []*message.AppInfo{
			{
				Owner: "alice",
				Ports: []message.PortInfo{
					{ExposePort: 30000, Protocol: "tcp"},
				},
			},
		},
	}

	listeners := tr.buildStreamListeners(resources)
	require.Len(t, listeners, 1)
	assert.Equal(t, ir.ProtocolTCP, listeners[0].Protocol)
	assert.Equal(t, uint32(30000), listeners[0].Port)
	assert.Equal(t, "bfl.user-space-alice", listeners[0].Routes[0].Destination.Host)
}

func TestBuildStreamListeners_UDP(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{Name: "alice", BFLHost: "bfl.user-space-alice"},
		},
		Apps: []*message.AppInfo{
			{
				Owner: "alice",
				Ports: []message.PortInfo{
					{ExposePort: 51820, Protocol: "udp"},
				},
			},
		},
	}

	listeners := tr.buildStreamListeners(resources)
	require.Len(t, listeners, 1)
	assert.Equal(t, ir.ProtocolUDP, listeners[0].Protocol)
}

func TestBuildStreamListeners_SkipsInvalidPort(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{
			{Name: "alice", BFLHost: "bfl.user-space-alice"},
		},
		Apps: []*message.AppInfo{
			{
				Owner: "alice",
				Ports: []message.PortInfo{
					{ExposePort: 0, Protocol: "tcp"},
					{ExposePort: -1, Protocol: "tcp"},
					{ExposePort: 70000, Protocol: "tcp"},
				},
			},
		},
	}

	listeners := tr.buildStreamListeners(resources)
	assert.Empty(t, listeners)
}

func TestBuildStreamListeners_SkipsMissingOwner(t *testing.T) {
	tr := &Translator{cfg: &Config{}}

	resources := &message.Resources{
		Users: []*message.UserInfo{},
		Apps: []*message.AppInfo{
			{
				Owner: "unknown",
				Ports: []message.PortInfo{
					{ExposePort: 30000, Protocol: "tcp"},
				},
			},
		},
	}

	listeners := tr.buildStreamListeners(resources)
	assert.Empty(t, listeners)
}
