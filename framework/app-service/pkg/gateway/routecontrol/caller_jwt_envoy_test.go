package routecontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Use the production issuer and policy/route generators. Only JWKS transport
// is replaced by an inline public key set in the isolated Envoy fixture.
func callerJWTEnvoyContract(t *testing.T) (string, map[string]any, map[string]any) {
	t.Helper()
	key, err := callerjwt.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	ring := callerjwt.KeyRing{Active: key}
	issuer, err := callerjwt.NewIssuer(ring)
	if err != nil {
		t.Fatal(err)
	}
	token, err := issuer.Issue(callerjwt.IssueRequest{Namespace: "user-space-alice", ServiceAccountName: "client", AppRef: "client", Viewer: "alice", ClientAppid: "test-client"})
	if err != nil {
		t.Fatal(err)
	}
	jwks, err := callerjwt.BuildJWKS(ring)
	if err != nil {
		t.Fatal(err)
	}
	public, err := json.Marshal(jwks)
	if err != nil {
		t.Fatal(err)
	}
	srr := &srrv1alpha1.SharedRouteRegistry{ObjectMeta: metav1.ObjectMeta{Name: "shared", Namespace: "shared", UID: "shared"}, Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassShared, HostPatterns: []string{"shared.alice.example.org"}, Upstream: srrv1alpha1.UpstreamRef{ServiceName: "business", Port: 80}}}
	policy := desiredSharedRouteSecurityPolicy(srr)
	providers, _, _ := unstructured.NestedSlice(policy.Object, "spec", "jwt", "providers")
	provider := providers[0].(map[string]any)
	headers, _, _ := unstructured.NestedSlice(provider, "extractFrom", "headers")
	mappings, _, _ := unstructured.NestedSlice(provider, "claimToHeaders")
	claims := []any{}
	for _, v := range mappings {
		m := v.(map[string]any)
		claims = append(claims, map[string]any{"header_name": m["header"], "claim_name": m["claim"]})
	}
	jwtFilter := map[string]any{"name": "envoy.filters.http.jwt_authn", "typed_config": map[string]any{
		"@type":     "type.googleapis.com/envoy.extensions.filters.http.jwt_authn.v3.JwtAuthentication",
		"providers": map[string]any{"caller": map[string]any{"issuer": provider["issuer"], "audiences": provider["audiences"], "from_headers": headers, "local_jwks": map[string]any{"inline_string": string(public)}, "claim_to_headers": claims}},
		"rules":     []any{map[string]any{"match": map[string]any{"prefix": "/shared"}, "requires": map[string]any{"provider_name": "caller"}}},
	}}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).Build()
	if _, err = applyHTTPRoute(context.Background(), c, GatewayRef{}, srr, 80); err != nil {
		t.Fatal(err)
	}
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err = c.Get(context.Background(), types.NamespacedName{Namespace: srr.Namespace, Name: httpRouteName(srr)}, route); err != nil {
		t.Fatal(err)
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	filters, _, _ := unstructured.NestedSlice(rules[0].(map[string]any), "filters")
	remove, _, _ := unstructured.NestedStringSlice(filters[0].(map[string]any), "requestHeaderModifier", "remove")
	return token, jwtFilter, map[string]any{"match": map[string]any{"prefix": "/shared"}, "route": map[string]any{"cluster": "business"}, "request_headers_to_remove": remove}
}

// The gateway-wide east-west policy strips claims, not the injected credential.
// Read the shipped policy so widening the northbound strip to all listeners
// breaks the real JWT positive control instead of silently breaking callers.
func internalIdentityHeaderSettings(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile("../../../../app-gateway/.olares/config/cluster/deploy/app-gateway-strip-identity-headers.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err = k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(raw), 4096).Decode(&obj); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	h, found, err := unstructured.NestedMap(obj, "spec", "headers", "earlyRequestHeaders")
	if err != nil || !found {
		t.Fatal("missing internal identity policy", err)
	}
	return h
}
