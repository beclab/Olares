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
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert-alice", KeyData: "key-alice"},
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
		Users: []*message.UserInfo{
			{
				Name:              "bob",
				Namespace:         "user-space-bob",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				DenyAll:           true,
				AllowedDomains:    []string{"vault-bob.snowinning.com"},
				LocalDomainIP:     "192.168.1.100",
				ServerNameDomains: []string{"bob.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "bob.snowinning.com", CertData: "cert-bob", KeyData: "key-bob"},
				Apps:              []*message.AppInfo{},
			},
		},
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
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert-alice", KeyData: "key-alice"},
			},
			{
				Name:              "bob",
				Namespace:         "user-space-bob",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				DenyAll:           true,
				AllowedDomains:    []string{"desktop-bob.snowinning.com"},
				LocalDomainIP:     "10.0.0.5",
				ServerNameDomains: []string{"bob.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "bob.snowinning.com", CertData: "cert-bob", KeyData: "key-bob"},
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

func TestE2E_EphemeralUser(t *testing.T) {
	// Ephemeral users are now merged into the owner's filter chain instead of
	// getting a separate filter chain.  The owner's *.zone wildcard covers the
	// wizard-{guest}.zone domain, so no SNI conflict occurs.  Only the wizard
	// virtual host for the ephemeral user is merged (not their app VHs).
	resources := &message.Resources{
		Users: []*message.UserInfo{
			{
				Name:              "alice",
				Namespace:         "user-space-alice",
				Zone:              "snowinning.com",
				IsEphemeral:       false,
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert-alice", KeyData: "key-alice"},
			},
			{
				// Ephemeral user (wizard setup session for alice's zone).
				// Their wizard VH (wizard-tempuser.alice.snowinning.com) will
				// be merged into alice's filter chain.
				Name:        "tempuser",
				Namespace:   "user-space-tempuser",
				Zone:        "alice.snowinning.com",
				IsEphemeral: true,
				DenyAll:     false,
				SSL:         &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert-temp", KeyData: "key-temp"},
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
				DenyAll:           false,
				ServerNameDomains: []string{"alice.snowinning.com"},
				SSL:               &message.SSLConfig{Zone: "alice.snowinning.com", CertData: "cert-alice", KeyData: "key-alice"},
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

// TestE2E_AddUserNoExistingDrain verifies that adding a new user leaves the
// existing users' filter chains byte-for-byte identical.
//
// Envoy performs an "in-place filter chain update" when:
//   - the listener's non-filter-chain config is unchanged (address, port,
//     listener-level filters, per_connection_buffer_limit_bytes)
//   - filter chains are identified by name
//
// For an in-place update with no drain:
//   - alice's filter chain ("https_443_alice") must be identical byte-for-byte
//   - bob's filter chain ("https_443_bob") must be a new addition
//
// This test proves both conditions, confirming that adding bob does NOT
// disrupt alice's existing connections.
func TestE2E_AddUserNoExistingDrain(t *testing.T) {
	makeUser := func(name string) *message.UserInfo {
		zone := name + ".snowinning.com"
		return &message.UserInfo{
			Name:              name,
			Namespace:         "user-space-" + name,
			Zone:              "snowinning.com",
			IsEphemeral:       false,
			DenyAll:           false,
			ServerNameDomains: []string{zone, name + ".olares.local"},
			SSL:               &message.SSLConfig{Zone: zone, CertData: "cert-" + name, KeyData: "key-" + name},
		}
	}

	tr := translator.New(nil, nil, &translator.Config{
		SSLServerPort:      443,
		SSLProxyServerPort: 444,
	})
	xt := xdstranslator.New(nil, nil)

	// Translate with only alice.
	snapBefore := xt.Translate(tr.Translate(&message.Resources{
		Users: []*message.UserInfo{makeUser("alice")},
	}))

	// Translate with alice + bob (new user).
	snapAfter := xt.Translate(tr.Translate(&message.Resources{
		Users: []*message.UserInfo{makeUser("alice"), makeUser("bob")},
	}))

	// Alice now has one filter chain per zone:
	//   https_443_alice_alice_snowinning_com
	//   https_443_alice_alice_olares_local
	aliceChains := []string{
		"https_443_alice_alice_snowinning_com",
		"https_443_alice_alice_olares_local",
	}
	for _, chainName := range aliceChains {
		before := extractFilterChainJSON(snapBefore.Listeners, "https_443", chainName)
		after := extractFilterChainJSON(snapAfter.Listeners, "https_443", chainName)

		require.NotEmpty(t, before, "alice's filter chain %q not found before bob was added", chainName)
		require.NotEmpty(t, after, "alice's filter chain %q not found after bob was added", chainName)

		// Alice's per-zone filter chains must be byte-for-byte identical.
		// Envoy identifies them by name as "unchanged" → no drain.
		assert.JSONEq(t, before, after,
			"alice's filter chain %q changed after bob was added → Envoy would drain alice's connections", chainName)
	}

	// Bob's filter chains must be NEW (absent before, present after).
	bobChains := []string{
		"https_443_bob_bob_snowinning_com",
		"https_443_bob_bob_olares_local",
	}
	for _, chainName := range bobChains {
		assert.Empty(t, extractFilterChainJSON(snapBefore.Listeners, "https_443", chainName),
			"bob's filter chain %q must not exist before he was added", chainName)
		assert.NotEmpty(t, extractFilterChainJSON(snapAfter.Listeners, "https_443", chainName),
			"bob's filter chain %q must exist after he was added", chainName)
	}
}

// TestE2E_AddZoneNoExistingDrain verifies that adding a new zone domain to an
// existing user leaves all other users' filter chains byte-for-byte identical.
//
// Before: alice has zones [alice.snowinning.com, alice.olares.local]
// After:  alice gains alice.newzone.com as a third zone
//
// Expected:
//   - alice's two existing filter chains are unchanged (no drain)
//   - alice gets a third NEW filter chain for alice.newzone.com
//   - bob's filter chains are completely unaffected
func TestE2E_AddZoneNoExistingDrain(t *testing.T) {
	makeUser := func(name string, extraZones ...string) *message.UserInfo {
		zone := name + ".snowinning.com"
		domains := []string{zone, name + ".olares.local"}
		domains = append(domains, extraZones...)
		return &message.UserInfo{
			Name:              name,
			Namespace:         "user-space-" + name,
			Zone:              "snowinning.com",
			IsEphemeral:       false,
			DenyAll:           false,
			ServerNameDomains: domains,
			SSL:               &message.SSLConfig{Zone: zone, CertData: "cert-" + name, KeyData: "key-" + name},
		}
	}

	tr := translator.New(nil, nil, &translator.Config{
		SSLServerPort:      443,
		SSLProxyServerPort: 444,
	})
	xt := xdstranslator.New(nil, nil)

	// Before: alice has 2 zones, bob has 2 zones.
	snapBefore := xt.Translate(tr.Translate(&message.Resources{
		Users: []*message.UserInfo{makeUser("alice"), makeUser("bob")},
	}))

	// After: alice gains a third zone; bob is untouched.
	snapAfter := xt.Translate(tr.Translate(&message.Resources{
		Users: []*message.UserInfo{makeUser("alice", "alice.newzone.com"), makeUser("bob")},
	}))

	// Alice's original two chains must be byte-for-byte identical.
	for _, chainName := range []string{
		"https_443_alice_alice_snowinning_com",
		"https_443_alice_alice_olares_local",
	} {
		before := extractFilterChainJSON(snapBefore.Listeners, "https_443", chainName)
		after := extractFilterChainJSON(snapAfter.Listeners, "https_443", chainName)
		require.NotEmpty(t, before, "filter chain %q not found before zone addition", chainName)
		require.NotEmpty(t, after, "filter chain %q not found after zone addition", chainName)
		assert.JSONEq(t, before, after,
			"filter chain %q changed after zone was added → existing connections would be drained", chainName)
	}

	// Alice gets a NEW chain for the added zone.
	newChain := extractFilterChainJSON(snapAfter.Listeners, "https_443", "https_443_alice_alice_newzone_com")
	assert.NotEmpty(t, newChain, "new zone filter chain must be present after zone addition")
	assert.Empty(t,
		extractFilterChainJSON(snapBefore.Listeners, "https_443", "https_443_alice_alice_newzone_com"),
		"new zone filter chain must not exist before zone addition")

	// Bob's chains are completely untouched.
	for _, chainName := range []string{
		"https_443_bob_bob_snowinning_com",
		"https_443_bob_bob_olares_local",
	} {
		before := extractFilterChainJSON(snapBefore.Listeners, "https_443", chainName)
		after := extractFilterChainJSON(snapAfter.Listeners, "https_443", chainName)
		require.NotEmpty(t, before, "bob's filter chain %q must exist", chainName)
		assert.JSONEq(t, before, after, "bob's filter chain %q must be unchanged", chainName)
	}
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
