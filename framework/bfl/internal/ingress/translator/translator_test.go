package translator

import (
	"testing"

	"bytetrade.io/web3os/bfl/internal/ingress/ir"
	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// buildHTTPRedirectListener
// ---------------------------------------------------------------------------

func TestBuildHTTPRedirectListener(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	l := tr.buildHTTPRedirectListener()

	assert.Equal(t, "http_redirect_80", l.Name)
	assert.Equal(t, "0.0.0.0", l.Address)
	assert.Equal(t, uint32(80), l.Port)
	assert.True(t, l.IsRedirect)
	assert.False(t, l.TLS)
}

// ---------------------------------------------------------------------------
// makeLocalHost
// ---------------------------------------------------------------------------

func TestMakeLocalHost(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"vault.alice.snowinning.com", "vault.alice.olares.local"},
		{"alice.snowinning.com", "alice.snowinning.olares.local"},
		{"localhost", "localhost"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, makeLocalHost(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// getAppEntranceHostName
// ---------------------------------------------------------------------------

func TestGetAppEntranceHostName_SingleEntrance(t *testing.T) {
	entrances := []*message.EntranceInfo{{Name: "main", Host: "svc", Port: 80}}
	got := getAppEntranceHostName(entrances, 0, "vault", nil)
	assert.Equal(t, "vault", got)
}

func TestGetAppEntranceHostName_MultiEntrance(t *testing.T) {
	entrances := []*message.EntranceInfo{
		{Name: "web", Host: "web-svc", Port: 80},
		{Name: "api", Host: "api-svc", Port: 8080},
	}
	assert.Equal(t, "myapp0", getAppEntranceHostName(entrances, 0, "myapp", nil))
	assert.Equal(t, "myapp1", getAppEntranceHostName(entrances, 1, "myapp", nil))
}

func TestGetAppEntranceHostName_CustomThirdLevel(t *testing.T) {
	entrances := []*message.EntranceInfo{
		{Name: "web", Host: "web-svc", Port: 80},
		{Name: "api", Host: "api-svc", Port: 8080},
	}
	configs := []defaultThirdLevelDomainConfig{
		{EntranceName: "api", ThirdLevelDomain: "customapi"},
	}
	assert.Equal(t, "myapp0", getAppEntranceHostName(entrances, 0, "myapp", configs))
	assert.Equal(t, "customapi", getAppEntranceHostName(entrances, 1, "myapp", configs))
}

// ---------------------------------------------------------------------------
// sanitizeName
// ---------------------------------------------------------------------------

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/resources/cache/", "api_resources_cache"},
		{"/simple", "simple"},
		{"no_slash", "no_slash"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeName(tt.input))
		})
	}
}

// ---------------------------------------------------------------------------
// buildStreamListeners
// ---------------------------------------------------------------------------

func TestBuildStreamListeners_TCP(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		Apps: []*message.AppInfo{{
			Name:      "myapp",
			Namespace: "ns-alice",
			Ports: []*message.PortInfo{{
				Host:       "myapp-svc",
				Port:       8080,
				ExposePort: 30000,
				Protocol:   "tcp",
			}},
		}},
	}

	listeners := tr.buildStreamListeners(res, clusterSet)
	require.Len(t, listeners, 1)
	assert.Equal(t, "stream_tcp_30000", listeners[0].Name)
	assert.Equal(t, uint32(30000), listeners[0].Port)
	assert.Equal(t, "tcp", listeners[0].Protocol)
	assert.Contains(t, clusterSet, "stream_tcp_30000")
}

func TestBuildStreamListeners_UDP(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		Apps: []*message.AppInfo{{
			Name:      "wireguard",
			Namespace: "ns-alice",
			Ports: []*message.PortInfo{{
				Host:       "wg-svc",
				Port:       51820,
				ExposePort: 51820,
				Protocol:   "udp",
			}},
		}},
	}

	listeners := tr.buildStreamListeners(res, clusterSet)
	require.Len(t, listeners, 1)
	assert.Equal(t, "udp", listeners[0].Protocol)
}

func TestBuildStreamListeners_SkipsInvalid(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		Apps: []*message.AppInfo{{
			Name:      "bad",
			Namespace: "ns",
			Ports: []*message.PortInfo{
				{Host: "", Port: 80, ExposePort: 30000, Protocol: "tcp"},
				{Host: "svc", Port: 80, ExposePort: 0, Protocol: "tcp"},
				{Host: "svc", Port: 80, ExposePort: 70000, Protocol: "tcp"},
			},
		}},
	}

	listeners := tr.buildStreamListeners(res, clusterSet)
	assert.Empty(t, listeners)
}

func TestBuildStreamListeners_DedupPorts(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		Apps: []*message.AppInfo{
			{Name: "a", Namespace: "ns", Ports: []*message.PortInfo{{Host: "svc-a", Port: 80, ExposePort: 30000, Protocol: "tcp"}}},
			{Name: "b", Namespace: "ns", Ports: []*message.PortInfo{{Host: "svc-b", Port: 80, ExposePort: 30000, Protocol: "tcp"}}},
		},
	}

	listeners := tr.buildStreamListeners(res, clusterSet)
	assert.Len(t, listeners, 1)
}

// ---------------------------------------------------------------------------
// buildNonAppVirtualHost
// ---------------------------------------------------------------------------

func TestBuildNonAppVirtualHost_Normal(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{UserName: "alice"}

	vhost := tr.buildNonAppVirtualHost(res, nonAppServers[1], "alice.snowinning.com", false, clusterSet)

	assert.Equal(t, "nonapp_desktop", vhost.Name)
	assert.Contains(t, vhost.Domains, "desktop.alice.snowinning.com")
	require.Len(t, vhost.Routes, 1)
	assert.Equal(t, "/", vhost.Routes[0].PathPrefix)
}

func TestBuildNonAppVirtualHost_WizardEphemeral(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{UserName: "tempuser"}

	// wizard (index 2) with ephemeral → AuthEnabled: true; filter_pass handled in Envoy Lua
	vhost := tr.buildNonAppVirtualHost(res, nonAppServers[2], "snowinning.com", true, clusterSet)

	assert.Equal(t, "nonapp_wizard", vhost.Name)
	assert.Contains(t, vhost.Domains, "wizard-tempuser.snowinning.com")
}

func TestBuildNonAppVirtualHost_WizardNonEphemeral(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{UserName: "alice"}

	vhost := tr.buildNonAppVirtualHost(res, nonAppServers[2], "alice.snowinning.com", false, clusterSet)

	assert.Equal(t, "nonapp_wizard", vhost.Name)
	assert.Contains(t, vhost.Domains, "wizard.alice.snowinning.com")
}

func TestBuildNonAppVirtualHost_Ephemeral(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{UserName: "tempuser"}

	vhost := tr.buildNonAppVirtualHost(res, nonAppServers[0], "snowinning.com", true, clusterSet)

	assert.Equal(t, "nonapp_auth", vhost.Name)
	assert.Contains(t, vhost.Domains, "auth-tempuser.snowinning.com")
}

// ---------------------------------------------------------------------------
// Translate (full integration)
// ---------------------------------------------------------------------------

func TestTranslate_WithSSL(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	res := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName: "alice",
		UserZone: "alice.snowinning.com",
		Apps: []*message.AppInfo{{
			Name:      "vault",
			Appid:     "vault",
			Namespace: "ns-alice",
			Entrances: []*message.EntranceInfo{{
				Name: "main", Host: "vault-svc", Port: 80,
			}},
		}},
	}

	xds := tr.Translate(res)

	// HTTPS 443 + HTTPS 444 (HTTP redirect is currently disabled)
	require.Len(t, xds.HTTPListeners, 2)
	assert.Equal(t, "https_443", xds.HTTPListeners[0].Name)
	assert.False(t, xds.HTTPListeners[0].ProxyProtocol)
	assert.Equal(t, "https_444", xds.HTTPListeners[1].Name)
	assert.True(t, xds.HTTPListeners[1].ProxyProtocol)

	// TLS secret
	require.Len(t, xds.Secrets, 1)
	assert.Equal(t, "main-tls", xds.Secrets[0].Name)
	assert.Equal(t, "CERT", xds.Secrets[0].CertData)

	// Clusters should include authelia_backend and app clusters
	assert.True(t, len(xds.Clusters) > 0)
	found := false
	for _, c := range xds.Clusters {
		if c.Name == autheliaCluster {
			found = true
		}
	}
	assert.True(t, found, "authelia_backend cluster should be present")
}

func TestTranslate_NoSSL(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	res := &message.Resources{
		UserName: "alice",
	}

	xds := tr.Translate(res)

	// No listeners without SSL (HTTP redirect is currently disabled)
	require.Len(t, xds.HTTPListeners, 0)
	assert.Empty(t, xds.Secrets)
}

func TestTranslate_WithStreamPorts(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	res := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName: "alice",
		Apps: []*message.AppInfo{{
			Name:      "game",
			Namespace: "ns-alice",
			Entrances: []*message.EntranceInfo{{
				Name: "web", Host: "game-svc", Port: 80,
			}},
			Ports: []*message.PortInfo{
				{Host: "game-svc", Port: 25565, ExposePort: 25565, Protocol: "tcp"},
				{Host: "game-svc", Port: 25565, ExposePort: 25566, Protocol: "udp"},
			},
		}},
	}

	xds := tr.Translate(res)

	assert.Len(t, xds.StreamListeners, 2)
	protocols := map[string]bool{}
	for _, sl := range xds.StreamListeners {
		protocols[sl.Protocol] = true
	}
	assert.True(t, protocols["tcp"])
	assert.True(t, protocols["udp"])
}

func TestTranslate_CustomDomainCerts(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	res := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName: "alice",
		CustomDomainCerts: []*message.CertInfo{
			{Domain: "custom.example.com", CertData: "CUSTOM_CERT", KeyData: "CUSTOM_KEY"},
		},
	}

	xds := tr.Translate(res)

	require.Len(t, xds.Secrets, 2)
	assert.Equal(t, "main-tls", xds.Secrets[0].Name)
	assert.Equal(t, "custom-tls-custom.example.com", xds.Secrets[1].Name)
}

// ---------------------------------------------------------------------------
// buildAppVirtualHosts
// ---------------------------------------------------------------------------

func TestBuildAppVirtualHosts_Basic(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		UserName: "alice",
		Language: "en",
		UserZone: "alice.snowinning.com",
	}
	app := &message.AppInfo{
		Name:      "vault",
		Appid:     "vault",
		Namespace: "ns-alice",
		Settings:  map[string]string{},
		Entrances: []*message.EntranceInfo{{
			Name: "main", Host: "vault-svc", Port: 80,
		}},
	}

	vhosts := tr.buildAppVirtualHosts(res, app, "alice.snowinning.com", false, clusterSet)

	require.Len(t, vhosts, 1)
	assert.Equal(t, "app_vault_main", vhosts[0].Name)
	assert.Contains(t, vhosts[0].Domains, "vault.alice.snowinning.com")
	assert.Contains(t, clusterSet, "app_vault_main")
}

func TestBuildAppVirtualHosts_Ephemeral(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		UserName: "tempuser",
		Language: "en",
		UserZone: "snowinning.com",
	}
	app := &message.AppInfo{
		Name:      "vault",
		Appid:     "vault",
		Namespace: "ns-tempuser",
		Settings:  map[string]string{},
		Entrances: []*message.EntranceInfo{{
			Name: "main", Host: "vault-svc", Port: 80,
		}},
	}

	vhosts := tr.buildAppVirtualHosts(res, app, "snowinning.com", true, clusterSet)

	require.Len(t, vhosts, 1)
	assert.Contains(t, vhosts[0].Domains, "vault-tempuser.snowinning.com")
}

func TestBuildAppVirtualHosts_FilesApp(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		UserName: "alice",
		Language: "en",
		UserZone: "alice.snowinning.com",
		FileserverNodes: []*message.FileserverNodeInfo{
			{NodeName: "node1", PodIP: "10.0.0.1", IsMaster: true},
		},
	}
	app := &message.AppInfo{
		Name:      "files",
		Appid:     "files",
		Namespace: "ns-alice",
		Settings:  map[string]string{},
		Entrances: []*message.EntranceInfo{{
			Name: "main", Host: "files-svc", Port: 80,
		}},
	}

	vhosts := tr.buildAppVirtualHosts(res, app, "alice.snowinning.com", false, clusterSet)

	require.Len(t, vhosts, 1)
	// Should have fileserver routes + default route
	assert.True(t, len(vhosts[0].Routes) > 1, "files app should have extra fileserver routes")

	// Check that some routes have ExtAuth configured
	hasExtAuth := false
	for _, route := range vhosts[0].Routes {
		if route.ExtAuth != nil {
			hasExtAuth = true
			break
		}
	}
	assert.True(t, hasExtAuth, "fileserver routes should have ext_authz configured")
}

// ---------------------------------------------------------------------------
// buildFileserverRoutes
// ---------------------------------------------------------------------------

func TestBuildFileserverRoutes_MasterNode(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		UserName: "alice",
		FileserverNodes: []*message.FileserverNodeInfo{
			{NodeName: "node1", PodIP: "10.0.0.1", IsMaster: true},
		},
	}
	app := &message.AppInfo{Name: "files", Namespace: "ns-alice"}

	routes := tr.buildFileserverRoutes(res, app, clusterSet)

	assert.True(t, len(routes) > 0, "should produce fileserver routes")

	// Master node should have both node routes and master routes
	hasNodeRoute := false
	hasMasterRoute := false
	for _, route := range routes {
		if route.Name == "files_node_node1_api_resources_cache" {
			hasNodeRoute = true
		}
		if route.Name == "files_master_api_repos" {
			hasMasterRoute = true
		}
	}
	assert.True(t, hasNodeRoute, "should have node-specific routes")
	assert.True(t, hasMasterRoute, "should have master-only routes")
}

func TestBuildFileserverRoutes_NonMasterNode(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{
		UserName: "alice",
		FileserverNodes: []*message.FileserverNodeInfo{
			{NodeName: "node2", PodIP: "10.0.0.2", IsMaster: false},
		},
	}
	app := &message.AppInfo{Name: "files", Namespace: "ns-alice"}

	routes := tr.buildFileserverRoutes(res, app, clusterSet)

	// Non-master node should only have node routes, no master routes
	for _, route := range routes {
		assert.NotContains(t, route.Name, "master", "non-master node should not have master routes")
	}
}

func TestBuildFileserverRoutes_NoNodes(t *testing.T) {
	tr := &Translator{cfg: &Config{}}
	clusterSet := make(map[string]*ir.ClusterIR)
	res := &message.Resources{UserName: "alice"}
	app := &message.AppInfo{Name: "files", Namespace: "ns-alice"}

	routes := tr.buildFileserverRoutes(res, app, clusterSet)
	assert.Empty(t, routes)
}
