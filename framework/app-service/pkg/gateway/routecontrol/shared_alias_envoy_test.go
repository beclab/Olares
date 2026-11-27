package routecontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Exercise the generated claim contract with Envoy's JWT and list-metadata
// RBAC filters. Envoy Gateway CRD reconciliation is covered separately.
func TestWorkloadAliasRealEnvoy(t *testing.T) {
	if os.Getenv("WORKLOAD_ALIAS_ENVOY_TEST") != "1" {
		t.Skip("WORKLOAD_ALIAS_ENVOY_TEST=1 enables the isolated Envoy contract test")
	}
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	ordinary, _ := sharedAliasFixtures()
	policy := desiredSharedAliasPolicy(ordinary)
	rules, _, _ := unstructured.NestedSlice(policy.Object, "spec", "authorization", "rules")
	principal, _, _ := unstructured.NestedMap(rules[0].(map[string]any), "principal", "jwt")
	claims := principal["claims"].([]any)
	claim := claims[0].(map[string]any)
	providerName := principal["provider"].(string)
	key, err := callerjwt.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	ring := callerjwt.KeyRing{Active: key}
	issuer, err := callerjwt.NewIssuer(ring)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := callerjwt.BuildJWKS(ring)
	if err != nil {
		t.Fatal(err)
	}
	public, err := json.Marshal(keys)
	if err != nil {
		t.Fatal(err)
	}

	const jwtFilterName = "envoy.filters.http.jwt_authn"
	allow := map[string]any{}
	for i, expected := range claim["values"].([]any) {
		allow[fmt.Sprintf("claim-%d", i)] = map[string]any{
			"permissions": []any{map[string]any{"any": true}},
			"principals": []any{map[string]any{"metadata": map[string]any{
				"filter": jwtFilterName,
				"path":   []any{map[string]any{"key": providerName}, map[string]any{"key": claim["name"]}},
				"value": map[string]any{"list_match": map[string]any{"one_of": map[string]any{
					"string_match": map[string]any{"exact": expected},
				}}},
			}}},
		}
	}
	filter := func(name, typ string, cfg map[string]any) map[string]any {
		cfg["@type"] = "type.googleapis.com/envoy.extensions.filters." + typ
		return map[string]any{"name": name, "typed_config": cfg}
	}
	hcm := filter("envoy.filters.network.http_connection_manager", "network.http_connection_manager.v3.HttpConnectionManager", map[string]any{
		"stat_prefix": "alias",
		"route_config": map[string]any{"name": "aliases", "virtual_hosts": []any{map[string]any{
			"name": "router", "domains": ordinary.Spec.HostPatterns,
			"routes": []any{map[string]any{"match": map[string]any{"prefix": "/v1"}, "direct_response": map[string]any{"status": 200}}},
		}}},
		"http_filters": []any{
			filter(jwtFilterName, "http.jwt_authn.v3.JwtAuthentication", map[string]any{
				"providers": map[string]any{providerName: map[string]any{
					"issuer": CallerJWTIssuer, "audiences": []string{CallerJWTAudience},
					"from_headers": []any{map[string]any{"name": callerjwt.CallerJWTHeaderName}},
					"local_jwks":   map[string]any{"inline_string": string(public)}, "payload_in_metadata": providerName,
				}},
				"rules": []any{map[string]any{"match": map[string]any{"prefix": "/v1"}, "requires": map[string]any{"provider_name": providerName}}},
			}),
			filter("envoy.filters.http.rbac", "http.rbac.v3.RBAC", map[string]any{"rules": map[string]any{"action": "ALLOW", "policies": allow}}),
			filter("envoy.filters.http.router", "http.router.v3.Router", map[string]any{}),
		},
	})
	listener := map[string]any{
		"name": "alias", "address": map[string]any{"socket_address": map[string]any{"address": "0.0.0.0", "port_value": 10000}},
		"filter_chains": []any{map[string]any{"filters": []any{hcm}}},
	}
	config, err := json.Marshal(map[string]any{"static_resources": map[string]any{"listeners": []any{listener}}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "envoy.json")
	if err = os.WriteFile(path, config, 0644); err != nil {
		t.Fatal(err)
	}
	docker := func(args ...string) string {
		t.Helper()
		out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("docker %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	id := docker("run", "-d", "-p", "127.0.0.1::10000", "-v", path+":/test/envoy.json:ro", entryEnvoyImage, "-c", "/test/envoy.json", "--log-level", "error")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", id).Run() })
	base := "http://" + docker("port", id, "10000/tcp")
	httpClient := &http.Client{Timeout: 2 * time.Second}
	call := func(host, token string) (int, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/v1/models", nil)
		if err != nil {
			return 0, err
		}
		req.Host = host
		req.Header.Set("X-Caller-Appid", "forged")
		if token != "" {
			req.Header.Set(callerjwt.CallerJWTHeaderName, token)
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()
		return resp.StatusCode, nil
	}
	deadline := time.Now().Add(8 * time.Second)
	for {
		if _, err := call(ordinary.Spec.HostPatterns[0], ""); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("Envoy failed to start: %s", docker("logs", id))
		}
		time.Sleep(100 * time.Millisecond)
	}
	for _, tc := range []struct {
		name, scope string
		status      int
	}{
		{"named grant", "alice/router", 200},
		{"existing platform eligibility", "alice/@all", 200},
		{"another tenant", "bob/router", 403},
		{"another target", "alice/model", 403},
		{"header only", "", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token := ""
			if tc.scope != "" {
				token, err = issuer.Issue(callerjwt.IssueRequest{
					Namespace: "lares-alice", ServiceAccountName: "lares", AppRef: "lares", ClientAppid: "lares",
					SharedTargets: []string{tc.scope},
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, host := range ordinary.Spec.HostPatterns {
				status, err := call(host, token)
				if err != nil || status != tc.status {
					t.Fatalf("%s returned %d, want %d, err=%v", host, status, tc.status, err)
				}
			}
		})
	}
}
