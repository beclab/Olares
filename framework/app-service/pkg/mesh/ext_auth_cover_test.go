package mesh

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestEvaluateEntranceExtAuthCovered(t *testing.T) {
	targets := []EntranceExtAuthTarget{{Namespace: "ns-a", SRRName: "app-web", HTTPRoute: "app-web"}}
	if !EvaluateEntranceExtAuthCovered(nil, func(EntranceExtAuthTarget) bool { return false }) {
		t.Fatal("empty targets must be covered")
	}
	if EvaluateEntranceExtAuthCovered(targets, nil) {
		t.Fatal("nil matcher must fail closed")
	}
	if EvaluateEntranceExtAuthCovered(targets, func(EntranceExtAuthTarget) bool { return false }) {
		t.Fatal("missing policy must not be covered")
	}
	if !EvaluateEntranceExtAuthCovered(targets, func(EntranceExtAuthTarget) bool { return true }) {
		t.Fatal("matching policy must be covered")
	}
}

func TestSecurityPolicyMatchesEntranceExtAuth(t *testing.T) {
	good := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1",
		"kind":       "SecurityPolicy",
		"metadata": map[string]any{
			"name": "app-web-entrance-ext-auth",
			"labels": map[string]any{
				"gateway.olares.io/auth-kind": authKindEntranceExtAuth,
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  "app-web",
			},
		},
	}}
	if !SecurityPolicyMatchesEntranceExtAuth(good, "app-web") {
		t.Fatal("expected match")
	}
	wrongRoute := good.DeepCopy()
	_ = unstructured.SetNestedField(wrongRoute.Object, "other-route", "spec", "targetRef", "name")
	if SecurityPolicyMatchesEntranceExtAuth(wrongRoute, "app-web") {
		t.Fatal("wrong targetRef must not match")
	}
}

func TestProbeEntranceExtAuthCoveredEGReadyDoesNotCount(t *testing.T) {
	// Cluster has application SRR but no SecurityPolicy — covered must be false
	// even if an operator might observe EG Deployment Ready separately.
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
	dc := newExtAuthProbeDynamic(t, srr)
	ok, err := ProbeEntranceExtAuthCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("EG-ready-style absence of SecurityPolicy must yield ExtAuthCovered=false")
	}
}

func TestProbeEntranceExtAuthCoveredWithMatchingPolicy(t *testing.T) {
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
		"kind":       "SecurityPolicy",
		"metadata": map[string]any{
			"name":      EntranceExtAuthPolicyName("app-web"),
			"namespace": "demo-user",
			"labels": map[string]any{
				"gateway.olares.io/auth-kind": authKindEntranceExtAuth,
			},
		},
		"spec": map[string]any{
			"targetRef": map[string]any{
				"group": "gateway.networking.k8s.io",
				"kind":  "HTTPRoute",
				"name":  "app-web",
			},
			"extAuth": map[string]any{"failOpen": false},
		},
	}}
	dc := newExtAuthProbeDynamic(t, srr, pol)
	ok, err := ProbeEntranceExtAuthCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("matching SecurityPolicy must yield ExtAuthCovered=true")
	}
}

func TestProbeEntranceExtAuthCoveredIgnoresSharedClass(t *testing.T) {
	srr := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]any{
			"name":      "shared-app",
			"namespace": "demo-shared",
		},
		"spec": map[string]any{
			"entranceClass": "shared",
			"hostPatterns":  []any{"shared.example.com"},
			"upstream": map[string]any{
				"serviceName": "shared",
				"port":        int64(80),
			},
		},
	}}
	dc := newExtAuthProbeDynamic(t, srr)
	ok, err := ProbeEntranceExtAuthCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("shared-class SRR must not require entrance ExtAuth")
	}
}

func newExtAuthProbeDynamic(t *testing.T, objs ...runtime.Object) *fake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		sharedRouteRegistryGVR: "SharedRouteRegistryList",
		securityPolicyGVR:      "SecurityPolicyList",
	}
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}
