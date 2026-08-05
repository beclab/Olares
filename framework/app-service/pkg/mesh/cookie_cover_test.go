package mesh

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestEntranceCookiePolicyName(t *testing.T) {
	if got := EntranceCookiePolicyName("app-web"); got != "app-web-entrance-cookie" {
		t.Fatalf("got %q", got)
	}
}

func TestEvaluateEntranceCookieCovered(t *testing.T) {
	targets := []EntranceCookieTarget{{Namespace: "ns", SRRName: "app-web", HTTPRoute: "app-web"}}
	if !EvaluateEntranceCookieCovered(nil, func(EntranceCookieTarget) bool { return false }) {
		t.Fatal("empty targets must be covered")
	}
	if EvaluateEntranceCookieCovered(targets, func(EntranceCookieTarget) bool { return false }) {
		t.Fatal("missing policy must not be covered")
	}
}

func TestProbeEntranceCookieCoveredMissingPolicy(t *testing.T) {
	srr := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]any{
			"name":      "app-web",
			"namespace": "demo-user",
		},
		"spec": map[string]any{
			"entranceClass": entranceClassApplication,
			"hostPatterns":  []any{"app.example.com"},
			"upstream": map[string]any{
				"serviceName": "app",
				"port":        int64(80),
			},
		},
	}}
	dc := newCookieProbeDynamic(t, srr)
	ok, err := ProbeEntranceCookieCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("missing Cookie EEP must yield Covered=false")
	}
}

func TestProbeEntranceCookieCoveredWithMatchingPolicy(t *testing.T) {
	srr := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]any{
			"name":      "app-web",
			"namespace": "demo-user",
		},
		"spec": map[string]any{
			"entranceClass": entranceClassApplication,
			"hostPatterns":  []any{"app.example.com"},
			"upstream": map[string]any{
				"serviceName": "app",
				"port":        int64(80),
			},
		},
	}}
	pol := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "EnvoyExtensionPolicy",
		"metadata": map[string]any{
			"name":      EntranceCookiePolicyName("app-web"),
			"namespace": "demo-user",
			"labels": map[string]any{
				"gateway.olares.io/auth-kind": AuthKindEntranceCookie,
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  "app-web",
			},
			"lua": []any{
				map[string]any{"type": "Inline", "inline": "function envoy_on_response(h) end"},
			},
		},
	}}
	dc := newCookieProbeDynamic(t, srr, pol)
	ok, err := ProbeEntranceCookieCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("matching Cookie EEP must yield Covered=true")
	}
}

func newCookieProbeDynamic(t *testing.T, objs ...runtime.Object) *fake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		sharedRouteRegistryGVR:  "SharedRouteRegistryList",
		envoyExtensionPolicyGVR: "EnvoyExtensionPolicyList",
	}
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}
