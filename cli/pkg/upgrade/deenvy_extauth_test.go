package upgrade

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

// assignExtAuthDepConditions records platform EG readiness separately from ExtAuth coverage.
func assignExtAuthDepConditions(conds map[string]bool, egReady, extAuthCovered bool) {
	if conds == nil {
		return
	}
	conds["AppGatewayDataReady"] = egReady
	conds["EntranceExtAuthCovered"] = extAuthCovered
}

func TestAssignExtAuthDepConditionsDoesNotEquateEGReady(t *testing.T) {
	conds := map[string]bool{}
	assignExtAuthDepConditions(conds, true, false)
	if !conds["AppGatewayDataReady"] {
		t.Fatal("EG ready must be recorded under AppGatewayDataReady")
	}
	if conds["EntranceExtAuthCovered"] {
		t.Fatal("EG Ready must not imply EntranceExtAuthCovered")
	}
	assignExtAuthDepConditions(conds, true, true)
	if !conds["EntranceExtAuthCovered"] {
		t.Fatal("true ExtAuth probe must set EntranceExtAuthCovered")
	}
}

func TestEvaluateEntranceExtAuthCoveredEmptyAndMissing(t *testing.T) {
	if !evaluateEntranceExtAuthCovered(nil, func(entranceExtAuthTarget) bool { return false }) {
		t.Fatal("empty targets must be covered")
	}
	targets := []entranceExtAuthTarget{{Namespace: "ns", SRRName: "app-web", HTTPRoute: "app-web"}}
	if evaluateEntranceExtAuthCovered(targets, func(entranceExtAuthTarget) bool { return false }) {
		t.Fatal("missing policy must not be covered")
	}
}

func TestProbeEntranceExtAuthCoveredEGReadyDoesNotCount(t *testing.T) {
	srr := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]any{
			"name":      "app-web",
			"namespace": "demo-user",
		},
		"spec": map[string]any{
			"entranceClass": deenvyEntranceClassApp,
			"hostPatterns":  []any{"app.example.com"},
			"upstream": map[string]any{
				"serviceName": "app",
				"port":        int64(80),
			},
		},
	}}
	dc := newDeenvyExtAuthDynamic(t, srr)
	ok, err := probeEntranceExtAuthCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("EG Ready without SecurityPolicy must yield ExtAuthCovered=false")
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
			"entranceClass": deenvyEntranceClassApp,
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
			"name":      "app-web" + deenvyEntranceExtAuthSuffix,
			"namespace": "demo-user",
			"labels": map[string]any{
				"gateway.olares.io/auth-kind": deenvyAuthKindExtAuth,
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
	dc := newDeenvyExtAuthDynamic(t, srr, pol)
	ok, err := probeEntranceExtAuthCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("matching SecurityPolicy must yield ExtAuthCovered=true")
	}
}

func newDeenvyExtAuthDynamic(t *testing.T, objs ...runtime.Object) *fake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		deenvySRR_GVR:           "SharedRouteRegistryList",
		deenvySecurityPolicyGVR: "SecurityPolicyList",
	}
	return fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
}
