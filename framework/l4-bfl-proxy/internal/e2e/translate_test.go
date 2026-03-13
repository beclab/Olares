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
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Namespace:         "user-space-alice",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-alice",
				BFLPort:           444,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
			},
		},
		Apps: []*message.AppInfo{
			{
				Name:  "vault",
				Appid: "vault",
				Owner: "alice",
				Entrances: []message.EntranceInfo{
					{Name: "vault-frontend"},
				},
				Ports: []message.PortInfo{
					{Name: "tcp", ExposePort: 48126, Protocol: "tcp"},
				},
			},
		},
	}
	runE2E(t, "single-admin-user", resources)
}

func TestE2E_DenyAllWithAllowedDomains(t *testing.T) {
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "bob",
				Namespace:         "user-space-bob",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-bob",
				BFLPort:           444,
				DenyAll:           true,
				AllowedDomains:    []string{"vault-bob.snowinning.com"},
				LocalDomainIP:     "192.168.1.100",
				ServerNameDomains: []string{"bob.snowinning.com"},
			},
		},
		Apps: []*message.AppInfo{},
	}
	runE2E(t, "deny-all-with-allowed-domains", resources)
}

func TestE2E_MultipleUsersAndApps(t *testing.T) {
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Namespace:         "user-space-alice",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-alice",
				BFLPort:           444,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
			},
			{
				Name:              "bob",
				Namespace:         "user-space-bob",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-bob",
				BFLPort:           444,
				DenyAll:           true,
				AllowedDomains:    []string{"desktop-bob.snowinning.com"},
				LocalDomainIP:     "10.0.0.5",
				ServerNameDomains: []string{"bob.snowinning.com"},
			},
		},
		Apps: []*message.AppInfo{
			{
				Name:  "vault",
				Appid: "vault",
				Owner: "alice",
				Entrances: []message.EntranceInfo{
					{Name: "vault-frontend"},
				},
			},
			{
				Name:  "desktop",
				Appid: "desktop",
				Owner: "bob",
				Entrances: []message.EntranceInfo{
					{Name: "desktop-frontend"},
				},
				Ports: []message.PortInfo{
					{Name: "rdp", ExposePort: 33890, Protocol: "tcp"},
				},
			},
		},
	}
	runE2E(t, "multiple-users-and-apps", resources)
}

func TestE2E_EphemeralUser(t *testing.T) {
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:        "tempuser",
				Namespace:   "user-space-tempuser",
				Zone:        "snowinning.com",
				IsEphemeral: true,
				BFLHost:     "bfl.user-space-tempuser",
				BFLPort:     444,
				DenyAll:     false,
			},
		},
		Apps: []*message.AppInfo{
			{
				Name:  "vault",
				Appid: "vault",
				Owner: "tempuser",
				Entrances: []message.EntranceInfo{
					{Name: "vault-frontend"},
				},
			},
			{
				Name:  "files",
				Appid: "files",
				Owner: "tempuser",
				Entrances: []message.EntranceInfo{
					{Name: "files-frontend"},
				},
			},
		},
	}
	runE2E(t, "ephemeral-user", resources)
}

func TestE2E_UDPPort(t *testing.T) {
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Namespace:         "user-space-alice",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				BFLHost:           "bfl.user-space-alice",
				BFLPort:           444,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
			},
		},
		Apps: []*message.AppInfo{
			{
				Name:  "wireguard",
				Appid: "wireguard",
				Owner: "alice",
				Entrances: []message.EntranceInfo{
					{Name: "wg"},
				},
				Ports: []message.PortInfo{
					{Name: "wg", ExposePort: 51820, Protocol: "udp"},
				},
			},
		},
	}
	runE2E(t, "udp-port", resources)
}

// ---------------------------------------------------------------------------
// pipeline runner + golden file comparison
// ---------------------------------------------------------------------------

func runE2E(t *testing.T, name string, resources *message.Resources) {
	t.Helper()

	// Stage 1: Resources → IR
	tr := translator.New(nil, nil, &translator.Config{
		SSLServerPort:       443,
		SSLProxyServerPort:  444,
		UserNamespacePrefix: "user-space",
	})
	irResult := tr.Translate(resources)

	// Stage 2: IR → xDS protobuf
	xt := xdstranslator.New(nil, nil)
	listeners, clusters := xt.Translate(irResult)

	// Serialize to JSON
	gotListeners := marshalResources(t, listeners)
	gotClusters := marshalResources(t, clusters)

	got := map[string]interface{}{
		"listeners": gotListeners,
		"clusters":  gotClusters,
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
