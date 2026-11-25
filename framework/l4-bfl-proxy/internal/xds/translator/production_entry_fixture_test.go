package translator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	production "github.com/beclab/l4-bfl-proxy/internal/translator"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Export the real terminal-mode translator output, not a hand-written IR.
// Only ADS delivery and cluster IPs are replaced for a single isolated Envoy;
// both authorization (9444) and business (8443) retain production mTLS/SNI.
func TestProductionEntryFixture(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "isolated-entry-fixture"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, DNSNames: []string{"*.alice.example.org", "*.alice.olares.local", production.DefaultAppGatewayHost, "l4-bfl-proxy.entry.client"}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	require.NoError(t, err)
	private, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private}))
	tr := production.New(nil, nil, &production.Config{AppGatewayMTLS: true, SSLServerPort: 10005, SSLProxyServerPort: 10006})
	input := tr.Translate(&message.Resources{Users: []*message.UserInfo{{Name: "alice", ServerNameDomains: []string{"alice.example.org", "alice.olares.local"}, Namespace: "user-space-alice", SSL: &message.SSLConfig{Zone: "alice.example.org", CertData: certPEM, KeyData: keyPEM}}}, EntryMTLS: &message.EntryMTLSMaterial{ClientCert: certPEM, ClientKey: keyPEM, CA: certPEM}})
	for _, c := range input.Clusters {
		require.NotNil(t, c.UpstreamMTLS, "terminal mode must not retain a cleartext business backend")
		require.Contains(t, []uint32{8443, 9444}, c.Port)
	}
	snap := (&XdsTranslator{}).Translate(input)
	asMap := func(p proto.Message) map[string]any {
		raw, e := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(p)
		require.NoError(t, e)
		var out map[string]any
		require.NoError(t, json.Unmarshal(raw, &out))
		return out
	}
	routes := map[string]map[string]any{}
	for _, p := range snap.Routes {
		m := asMap(p)
		routes[m["name"].(string)] = m
	}
	var inline func(any)
	inline = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			// Static secrets use the same generated material; no SDS server needed.
			delete(x, "sds_config")
			if r, ok := x["rds"].(map[string]any); ok {
				name := r["route_config_name"].(string)
				require.NotNil(t, routes[name])
				x["route_config"] = routes[name]
				delete(x, "rds")
			}
			for _, y := range x {
				inline(y)
			}
		case []any:
			for _, y := range x {
				inline(y)
			}
		}
	}
	var listener map[string]any
	for _, p := range snap.Listeners {
		m := asMap(p)
		if m["address"].(map[string]any)["socket_address"].(map[string]any)["port_value"] == float64(10005) {
			listener = m
			break
		}
	}
	require.NotNil(t, listener)
	inline(listener)
	clusters := []any{}
	for _, p := range snap.Clusters {
		m := asMap(p)
		inline(m)
		assignment := m["load_assignment"].(map[string]any)
		for _, ep := range assignment["endpoints"].([]any) {
			for _, lb := range ep.(map[string]any)["lb_endpoints"].([]any) {
				lb.(map[string]any)["endpoint"].(map[string]any)["address"].(map[string]any)["socket_address"].(map[string]any)["address"] = "127.0.0.1"
			}
		}
		clusters = append(clusters, m)
	}
	secrets := []any{}
	for _, p := range snap.Secrets {
		secrets = append(secrets, asMap(p))
	}
	if path := os.Getenv("ENTRY_EDGE_FIXTURE"); path != "" {
		raw, e := json.Marshal(map[string]any{"listener": listener, "clusters": clusters, "secrets": secrets, "ca": certPEM})
		require.NoError(t, e)
		require.NoError(t, os.WriteFile(path+".production", raw, 0600))
	}
}
