package e2e

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/beclab/l4-bfl-proxy/internal/translator"
	xdstranslator "github.com/beclab/l4-bfl-proxy/internal/xds/translator"
)

var update = flag.Bool("update", false, "update golden files")

// ---------------------------------------------------------------------------
// test cases
// ---------------------------------------------------------------------------

func TestE2E_SingleAdminUser(t *testing.T) {
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL: &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Zone:      "snowinning.com",
				DenyAll:   false,
				Apps: []*message.AppInfo{
					{
						Name:      "vault",
						Appid:     "vault",
						Namespace: "vault-alice",
						Owner:     "alice",
						Entrances: []*message.EntranceInfo{
							{Name: "vault-frontend", Host: "vault-svc", Port: 8080},
						},
						Ports: []*message.PortInfo{
							{Name: "tcp", Host: "vault-svc", Port: 8080, ExposePort: 48126, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}
	runE2E(t, "single-admin-user", resources)
}

func TestE2E_DenyAllWithAllowedDomains(t *testing.T) {
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL:           &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		LocalDomainIP: "192.168.1.100",
		Users: []*message.UserInfo{
			{
				Name:           "bob",
				Namespace:      "user-space-bob",
				Zone:           "snowinning.com",
				DenyAll:        true,
				AllowedDomains: []string{"vault-bob.snowinning.com"},
				Apps:           []*message.AppInfo{},
			},
		},
	}
	runE2E(t, "deny-all-with-allowed-domains", resources)
}

func TestE2E_MultipleUsersAndApps(t *testing.T) {
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL:           &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		LocalDomainIP: "10.0.0.5",
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Zone:      "snowinning.com",
				DenyAll:   false,
			},
			{
				Name:           "bob",
				Namespace:      "user-space-bob",
				Zone:           "snowinning.com",
				DenyAll:        true,
				AllowedDomains: []string{"desktop-bob.snowinning.com"},
				Apps: []*message.AppInfo{
					{
						Name:      "vault",
						Appid:     "vault",
						Namespace: "vault-bob",
						Owner:     "bob",
						Entrances: []*message.EntranceInfo{
							{Name: "vault-frontend", Host: "vault-svc", Port: 8080},
						},
					},
					{
						Name:      "desktop",
						Appid:     "desktop",
						Namespace: "desktop-bob",
						Owner:     "bob",
						Entrances: []*message.EntranceInfo{
							{Name: "desktop-frontend", Host: "desktop-svc", Port: 80},
						},
						Ports: []*message.PortInfo{
							{Name: "rdp", Host: "desktop-svc", Port: 3389, ExposePort: 33890, Protocol: "tcp"},
						},
					},
				},
			},
		},
	}
	runE2E(t, "multiple-users-and-apps", resources)
}

func TestE2E_UDPPort(t *testing.T) {
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL: &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		Users: []*message.UserInfo{
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Zone:      "snowinning.com",
				DenyAll:   false,
				Apps: []*message.AppInfo{
					{
						Name:      "wireguard",
						Appid:     "wireguard",
						Namespace: "wireguard-alice",
						Owner:     "alice",
						Entrances: []*message.EntranceInfo{
							{Name: "wg", Host: "wireguard-svc", Port: 51820},
						},
						Ports: []*message.PortInfo{
							{Name: "wg", Host: "wireguard-svc", Port: 51820, ExposePort: 51820, Protocol: "udp"},
						},
					},
				},
			},
		},
	}
	runE2E(t, "udp-port", resources)
}

// TestE2E_SharedAppOpenToAllUsers exercises the v3 / shared-app routing
// end-to-end: every user — admin and normal — reaches the same shared
// upstream cluster. Shared apps no longer carry a 403 gate.
func TestE2E_SharedAppOpenToAllUsers(t *testing.T) {
	sharedApp := func(owner string) *message.AppInfo {
		return &message.AppInfo{
			Name:      "shareme",
			Appid:     "shareme",
			Namespace: "shareme-shared",
			Owner:     owner,
			IsShared:  true,
			Entrances: []*message.EntranceInfo{
				{Name: "web", Host: "shareme-svc", Port: 8080},
			},
		}
	}
	resources := &message.Resources{
		DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local",
		SSL: &message.SSLConfig{CertData: "machine-cert", KeyData: "machine-key"},
		Users: []*message.UserInfo{
			{
				Name:      "admin",
				Namespace: "user-space-admin",
				Zone:      "snowinning.com",
				DenyAll:   false,
				Apps:      []*message.AppInfo{sharedApp("admin")},
			},
			{
				Name:      "alice",
				Namespace: "user-space-alice",
				Zone:      "snowinning.com",
				DenyAll:   false,
				Apps:      []*message.AppInfo{sharedApp("admin")},
			},
			{
				Name:      "bob",
				Namespace: "user-space-bob",
				Zone:      "snowinning.com",
				DenyAll:   false,
				Apps:      []*message.AppInfo{sharedApp("admin")},
			},
		},
	}
	runE2E(t, "shared-app-open", resources)
}

// ---------------------------------------------------------------------------
// pipeline runner + golden file comparison
// ---------------------------------------------------------------------------

func runE2E(t *testing.T, name string, resources *message.Resources) {
	t.Helper()

	// Stage 1: Resources -> IR
	tr := translator.New(nil, nil, &translator.Config{
		SSLServerPort:      443,
		SSLProxyServerPort: 444,
	})
	irResult := tr.Translate(resources)

	// Stage 2: IR -> xDS protobuf
	xt := xdstranslator.New(nil, nil)
	snap := xt.Translate(irResult)

	// Serialize to JSON
	gotListeners := marshalResources(t, snap.Listeners)
	gotClusters := marshalResources(t, snap.Clusters)
	gotRoutes := marshalResources(t, snap.Routes)

	got := map[string]interface{}{
		"listeners": gotListeners,
		"clusters":  gotClusters,
		"routes":    gotRoutes,
	}
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", name+".json")

	if *update {
		err := os.WriteFile(goldenPath, append(gotJSON, '\n'), 0644)
		require.NoError(t, err)
		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	goldenData, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "golden file not found, run with -update to generate: %s", goldenPath)

	assert.JSONEq(t, string(goldenData), string(gotJSON),
		"envoy config mismatch for test case %q, run with -update to regenerate golden file", name)
}

// Shared machine filter chains do not change when an additional user is created.
func TestE2E_AddUserNoExistingDrain(t *testing.T) {
	tr := translator.New(nil, nil, &translator.Config{SSLServerPort: 443, SSLProxyServerPort: 444})
	xt := xdstranslator.New(nil, nil)
	resources := &message.Resources{DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local", SSL: &message.SSLConfig{CertData: "cert", KeyData: "key"}, Users: []*message.UserInfo{{Name: "alice"}}}
	before := xt.Translate(tr.Translate(resources))
	resources.Users = append(resources.Users, &message.UserInfo{Name: "bob"})
	after := xt.Translate(tr.Translate(resources))
	for _, chainName := range []string{"https_443_default", "https_444_default"} {
		port := "https_443"
		if chainName == "https_444_default" {
			port = "https_pp_444"
		}
		original := extractFilterChainJSON(before.Listeners, port, chainName)
		require.NotEmpty(t, original)
		assert.JSONEq(t, original, extractFilterChainJSON(after.Listeners, port, chainName))
	}
}

// Adding an application-specific certificate does not replace the machine chain.
func TestE2E_AddCustomDomainNoExistingDrain(t *testing.T) {
	tr := translator.New(nil, nil, &translator.Config{SSLServerPort: 443, SSLProxyServerPort: 444})
	xt := xdstranslator.New(nil, nil)
	user := &message.UserInfo{Name: "alice"}
	resources := &message.Resources{DefaultDomain: "lab.example.com", LocalZone: "lab.olares.local", SSL: &message.SSLConfig{CertData: "cert", KeyData: "key"}, Users: []*message.UserInfo{user}}
	before := xt.Translate(tr.Translate(resources))
	user.CustomDomainCerts = []*message.CertInfo{{Domain: "app.custom.io", CertData: "custom-cert", KeyData: "custom-key"}}
	after := xt.Translate(tr.Translate(resources))
	original := extractFilterChainJSON(before.Listeners, "https_443", "https_443_default")
	require.NotEmpty(t, original)
	assert.JSONEq(t, original, extractFilterChainJSON(after.Listeners, "https_443", "https_443_default"))
	assert.Empty(t, extractFilterChainJSON(before.Listeners, "https_443", "https_443_alice_custom_app_custom_io"))
	assert.NotEmpty(t, extractFilterChainJSON(after.Listeners, "https_443", "https_443_alice_custom_app_custom_io"))
}

// extractFilterChainJSON returns the JSON of the named filter chain inside the
// named listener, or "" if not found.  Uses protojson for deterministic output.
func extractFilterChainJSON(listeners []cachetypes.Resource, listenerName, chainName string) string {
	marshaler := protojson.MarshalOptions{}
	for _, r := range listeners {
		msg, ok := r.(proto.Message)
		if !ok {
			continue
		}
		b, err := marshaler.Marshal(msg)
		if err != nil {
			continue
		}
		var generic map[string]json.RawMessage
		if err := json.Unmarshal(b, &generic); err != nil {
			continue
		}
		var name string
		if err := json.Unmarshal(generic["name"], &name); err != nil || name != listenerName {
			continue
		}
		var chains []json.RawMessage
		if err := json.Unmarshal(generic["filterChains"], &chains); err != nil {
			continue
		}
		for _, chainRaw := range chains {
			var chain map[string]json.RawMessage
			if err := json.Unmarshal(chainRaw, &chain); err != nil {
				continue
			}
			var cname string
			if err := json.Unmarshal(chain["name"], &cname); err != nil || cname != chainName {
				continue
			}
			return string(chainRaw)
		}
	}
	return ""
}

func marshalResources(t *testing.T, resources []cachetypes.Resource) []json.RawMessage {
	t.Helper()
	marshaler := protojson.MarshalOptions{
		Indent: "  ",
	}
	var result []json.RawMessage
	for _, r := range resources {
		data, err := marshaler.Marshal(r.(proto.Message))
		require.NoError(t, err)
		result = append(result, json.RawMessage(data))
	}
	return result
}
