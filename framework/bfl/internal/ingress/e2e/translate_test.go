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

	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"bytetrade.io/web3os/bfl/internal/ingress/translator"
	xdstranslator "bytetrade.io/web3os/bfl/internal/ingress/xds/translator"
)

var update = flag.Bool("update", false, "update golden files")

// ---------------------------------------------------------------------------
// test cases
// ---------------------------------------------------------------------------

func TestE2E_SingleUserWithApp(t *testing.T) {
	resources := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "-----BEGIN CERTIFICATE-----\nTEST\n-----END CERTIFICATE-----",
			KeyData:  "-----BEGIN PRIVATE KEY-----\nTEST\n-----END PRIVATE KEY-----",
		},
		UserName: "alice",
		UserZone: "alice.snowinning.com",
		Language: "en",
		Apps: []*message.AppInfo{{
			Name:      "vault",
			Appid:     "vault",
			Namespace: "ns-alice",
			Settings:  map[string]string{},
			Entrances: []*message.EntranceInfo{{
				Name: "main", Host: "vault-svc", Port: 80,
			}},
		}},
	}
	runE2E(t, "single-user-with-app", resources)
}

func TestE2E_UserWithStreamPorts(t *testing.T) {
	resources := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName: "alice",
		UserZone: "alice.snowinning.com",
		Language: "en",
		Apps: []*message.AppInfo{{
			Name:      "game",
			Appid:     "game",
			Namespace: "ns-alice",
			Settings:  map[string]string{},
			Entrances: []*message.EntranceInfo{{
				Name: "web", Host: "game-svc", Port: 80,
			}},
			Ports: []*message.PortInfo{
				{Host: "game-svc", Port: 25565, ExposePort: 25565, Protocol: "tcp"},
			},
		}},
	}
	runE2E(t, "user-with-stream-ports", resources)
}

func TestE2E_EphemeralUser(t *testing.T) {
	resources := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName:        "tempuser",
		UserZone:        "snowinning.com",
		IsEphemeralUser: true,
		Language:        "en",
		Apps: []*message.AppInfo{{
			Name:      "vault",
			Appid:     "vault",
			Namespace: "ns-tempuser",
			Settings:  map[string]string{},
			Entrances: []*message.EntranceInfo{{
				Name: "main", Host: "vault-svc", Port: 80,
			}},
		}},
	}
	runE2E(t, "ephemeral-user", resources)
}

func TestE2E_FilesApp(t *testing.T) {
	resources := &message.Resources{
		SSL: &message.SSLConfig{
			Zone:     "alice.snowinning.com",
			CertData: "CERT",
			KeyData:  "KEY",
		},
		UserName: "alice",
		UserZone: "alice.snowinning.com",
		Language: "en",
		FileserverNodes: []*message.FileserverNodeInfo{
			{NodeName: "node1", PodIP: "10.0.0.1", IsMaster: true},
		},
		Apps: []*message.AppInfo{{
			Name:      "files",
			Appid:     "files",
			Namespace: "ns-alice",
			Settings:  map[string]string{},
			Entrances: []*message.EntranceInfo{{
				Name: "main", Host: "files-svc", Port: 80,
			}},
		}},
	}
	runE2E(t, "files-app", resources)
}

func TestE2E_NoSSL(t *testing.T) {
	resources := &message.Resources{
		UserName: "alice",
		UserZone: "alice.snowinning.com",
		Language: "en",
	}
	runE2E(t, "no-ssl", resources)
}

// ---------------------------------------------------------------------------
// pipeline runner + golden file comparison
// ---------------------------------------------------------------------------

func runE2E(t *testing.T, name string, resources *message.Resources) {
	t.Helper()

	// Stage 1: Resources -> IR
	tr := translator.New(nil, nil, &translator.Config{})
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
	if os.IsNotExist(err) {
		// Auto-generate on first run
		err = os.WriteFile(goldenPath, append(gotJSON, '\n'), 0644)
		require.NoError(t, err)
		t.Logf("generated golden file: %s (re-run to verify)", goldenPath)
		return
	}
	require.NoError(t, err, "golden file error: %s", goldenPath)

	assert.JSONEq(t, string(goldenData), string(gotJSON),
		"envoy config mismatch for test case %q, run `go test ./internal/ingress/e2e -run TestE2E -update` to regenerate golden files", name)
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
