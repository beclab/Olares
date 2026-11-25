package routecontrol

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// Reuse the normal policy fixture as the backend of the actual L4-generated
// TLS graph. This proves transport and policy contracts, not real user login.
func addTerminalEntryFixture(t *testing.T, bootstrap map[string]any) string {
	t.Helper()
	raw, err := os.ReadFile(os.Getenv("ENTRY_EDGE_FIXTURE") + ".production")
	if err != nil {
		t.Fatal("generate TestProductionEntryFixture first", err)
	}
	var prod struct {
		Listener map[string]any `json:"listener"`
		Clusters []any          `json:"clusters"`
		Secrets  []any          `json:"secrets"`
		CA       string         `json:"ca"`
	}
	if err = json.Unmarshal(raw, &prod); err != nil {
		t.Fatal(err)
	}
	// Scope generated cluster names so the independent edge-only contract
	// cannot collide with the terminal-mode authorization cluster.
	var scope func(any)
	scope = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, y := range x {
				if name, ok := y.(string); ok && (name == "authelia_backend_alice" || name == "app_gateway_edge_forward") {
					x[k] = "terminal-" + name
				} else {
					scope(y)
				}
			}
		case []any:
			for _, y := range x {
				scope(y)
			}
		}
	}
	scope(prod.Listener)
	scope(prod.Clusters)
	resources := bootstrap["static_resources"].(map[string]any)
	listeners := resources["listeners"].([]any)
	// Copy the Entry business listener while preserving the generated identity
	// strip and ext_authz policy used by the normal contract fixture.
	encoded, err := json.Marshal(listeners[0])
	if err != nil {
		t.Fatal(err)
	}
	var business map[string]any
	if err = json.Unmarshal(encoded, &business); err != nil {
		t.Fatal(err)
	}
	business["name"] = "terminal-entry-business"
	business["address"] = map[string]any{"socket_address": map[string]any{"address": "127.0.0.1", "port_value": 8443}}
	chain := business["filter_chains"].([]any)[0].(map[string]any)
	tlsSocket := map[string]any{"name": "envoy.transport_sockets.tls", "typed_config": map[string]any{
		"@type":                      "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext",
		"require_client_certificate": true,
		"common_tls_context": map[string]any{
			"alpn_protocols":                     entryALPNProtocols(t),
			"tls_certificate_sds_secret_configs": []any{map[string]any{"name": "main-tls-alice"}},
			"validation_context":                 map[string]any{"trusted_ca": map[string]any{"inline_string": prod.CA}, "match_typed_subject_alt_names": []any{map[string]any{"san_type": "DNS", "matcher": map[string]any{"exact": "l4-bfl-proxy.entry.client"}}}},
		},
	}}
	chain["transport_socket"] = tlsSocket
	h := chain["filters"].([]any)[0].(map[string]any)["typed_config"].(map[string]any)
	vh := h["route_config"].(map[string]any)["virtual_hosts"].([]any)[0].(map[string]any)
	routes := vh["routes"].([]any)
	for _, host := range []string{"auth.alice.example.org", "auth.alice.olares.local"} {
		routes = append([]any{map[string]any{"match": map[string]any{"prefix": "/", "headers": []any{map[string]any{"name": ":authority", "string_match": map[string]any{"exact": host}}}}, "direct_response": map[string]any{"status": 200, "body": map[string]any{"inline_string": "isolated login page"}}, "typed_per_filter_config": map[string]any{"envoy.filters.http.ext_authz": map[string]any{"@type": "type.googleapis.com/envoy.extensions.filters.http.ext_authz.v3.ExtAuthzPerRoute", "disabled": true}}}}, routes...)
	}
	vh["routes"] = routes
	// Compile the production authorization HTTPRoute's tenant binding, rewrite,
	// and authority list into the isolated Entry's dedicated 9444 listener.
	route, err := desiredL4AuthzRoute(GatewayRef{}, desktopEntrance())
	if err != nil {
		t.Fatal(err)
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	filters, _, _ := unstructured.NestedSlice(rules[0].(map[string]any), "filters")
	additions := []any{}
	for _, raw := range filters {
		f := raw.(map[string]any)
		if f["type"] != "RequestHeaderModifier" {
			continue
		}
		headers, _, _ := unstructured.NestedSlice(f, "requestHeaderModifier", "set")
		for _, raw := range headers {
			h := raw.(map[string]any)
			additions = append(additions, map[string]any{"header": map[string]any{"key": h["name"], "value": h["value"]}, "append_action": "OVERWRITE_IF_EXISTS_OR_ADD"})
		}
	}
	hosts, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
	authHCM := map[string]any{"@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager", "stat_prefix": "terminal-authorization", "route_config": map[string]any{"name": "terminal-authorization", "virtual_hosts": []any{map[string]any{"name": "bound-tenant", "domains": hosts, "routes": []any{map[string]any{"match": map[string]any{"prefix": AutheliaVerifyPath}, "route": map[string]any{"cluster": "tenant-auth", "prefix_rewrite": AutheliaVerifyPath}, "request_headers_to_add": additions}}}}}, "http_filters": []any{map[string]any{"name": "envoy.filters.http.router", "typed_config": map[string]any{"@type": "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router"}}}}
	// The same northbound strip applies before either Entry authorization path.
	authHCM["early_header_mutation_extensions"] = h["early_header_mutation_extensions"]
	auth := map[string]any{"name": "terminal-entry-authorization", "address": map[string]any{"socket_address": map[string]any{"address": "127.0.0.1", "port_value": 9444}}, "filter_chains": []any{map[string]any{"transport_socket": tlsSocket, "filters": []any{map[string]any{"name": "envoy.filters.network.http_connection_manager", "typed_config": authHCM}}}}}
	resources["listeners"] = append(listeners, prod.Listener, business, auth)
	resources["clusters"] = append(resources["clusters"].([]any), prod.Clusters...)
	resources["secrets"] = prod.Secrets
	// Keep failure diagnostics free of credentials, request headers and bodies.
	var diagnostics func(any)
	diagnostics = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if x["@type"] == "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager" {
				x["access_log"] = []any{map[string]any{"name": "envoy.access_loggers.stdout", "typed_config": map[string]any{"@type": "type.googleapis.com/envoy.extensions.access_loggers.stream.v3.StdoutAccessLog", "log_format": map[string]any{"text_format_source": map[string]any{"inline_string": "%RESPONSE_CODE% %RESPONSE_FLAGS% %UPSTREAM_TRANSPORT_FAILURE_REASON%\n"}}}}}
			}
			for _, y := range x {
				diagnostics(y)
			}
		case []any:
			for _, y := range x {
				diagnostics(y)
			}
		}
	}
	diagnostics(resources["listeners"])
	return prod.CA
}

// Mirror the shipped Entry ALPN configuration, including Gateway's h2-first
// default when omitted. Otherwise the fixture silently masks codec mismatch.
func entryALPNProtocols(t *testing.T) []string {
	t.Helper()
	raw, err := os.ReadFile("../../../../app-gateway/.olares/config/cluster/deploy/l4-mtls-channel.yaml")
	if err != nil {
		t.Fatal(err)
	}
	dec := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096)
	var protocols []string
	count := 0
	for {
		var obj map[string]any
		err = dec.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if obj["kind"] != "ClientTrafficPolicy" {
			continue
		}
		actual, found, err := unstructured.NestedStringSlice(obj, "spec", "tls", "alpnProtocols")
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			actual = []string{"h2", "http/1.1"}
		}
		if count > 0 && strings.Join(actual, ",") != strings.Join(protocols, ",") {
			t.Fatal("Entry listeners disagree on ALPN")
		}
		protocols = actual
		count++
	}
	if count != 2 {
		t.Fatal("missing dual Entry protocol contract")
	}
	return protocols
}
