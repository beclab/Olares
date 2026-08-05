package upgrade

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

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

func TestAssignCookieDepConditionsIndependent(t *testing.T) {
	conds := map[string]bool{"AppGatewayDataReady": true}
	assignCookieDepConditions(conds, false)
	if conds["EntranceCookieCovered"] {
		t.Fatal("missing Cookie EEP must record Covered=false")
	}
	if !conds["AppGatewayDataReady"] {
		t.Fatal("cookie assign must not clear AppGatewayDataReady")
	}
}

func TestProbeEntranceCookieCoveredMissing(t *testing.T) {
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
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		deenvySRR_GVR:                   "SharedRouteRegistryList",
		deenvyEnvoyExtensionPolicyGVR: "EnvoyExtensionPolicyList",
	}
	dc := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, srr)
	ok, err := probeEntranceCookieCovered(context.Background(), dc)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("missing Cookie EEP must be false")
	}
}

func TestAssignProbeBypassDepConditions(t *testing.T) {
	conds := map[string]bool{}
	assignProbeBypassDepConditions(conds, false)
	if conds["EntranceProbeBypassReady"] {
		t.Fatal("expected false")
	}
	assignProbeBypassDepConditions(conds, true)
	if !conds["EntranceProbeBypassReady"] {
		t.Fatal("expected true")
	}
}

func TestAssignRouteModeDepConditions(t *testing.T) {
	conds := map[string]bool{}
	assignRouteModeDepConditions(conds, false)
	if conds["RouteModeGateway"] {
		t.Fatal("expected false")
	}
	assignRouteModeDepConditions(conds, true)
	if !conds["RouteModeGateway"] {
		t.Fatal("expected true")
	}
}

func TestEvaluateEntranceRouteModeGatewayReadyPartial(t *testing.T) {
	targets := []entranceRouteModeTarget{
		{Namespace: "ns", AppName: "a", Mode: "gateway"},
		{Namespace: "ns", AppName: "b", Mode: ""},
	}
	if evaluateEntranceRouteModeGatewayReady(targets) {
		t.Fatal("partial gateway must be false")
	}
	if !hasPartialEntranceRouteModeGateway(targets) {
		t.Fatal("expected partial")
	}
	if !evaluateEntranceRouteModeGatewayReady(nil) {
		t.Fatal("vacuous must be true")
	}
}

func TestCanEnableOesFreeGate(t *testing.T) {
	good := map[string]bool{
		"EntranceExtAuthCovered":   true,
		"EntranceCookieCovered":    true,
		"EntranceProbeBypassReady": true,
		"EntranceAuxCovered":       true,
		"RouteModeGateway":         true,
		"L4ProxyReady":             true,
	}
	if !canEnableOesFreeGate(good) {
		t.Fatal("full coverage must allow oes-free gate")
	}
	bad := map[string]bool{}
	for k, v := range good {
		bad[k] = v
	}
	bad["EntranceExtAuthCovered"] = false
	if canEnableOesFreeGate(bad) {
		t.Fatal("missing ExtAuth must refuse oes-free gate")
	}
	if canEnableOesFreeGate(nil) {
		t.Fatal("nil must refuse")
	}
}

func TestEvaluateAcceptSuitePassed(t *testing.T) {
	good := map[string]bool{
		"ZeroOesInventory":         true,
		"EntranceExtAuthCovered":   true,
		"EntranceCookieCovered":    true,
		"EntranceProbeBypassReady": true,
		"EntranceAuxCovered":       true,
		"RouteModeGateway":         true,
	}
	if !evaluateAcceptSuitePassed(good) {
		t.Fatal("full coverage must pass")
	}
	bad := map[string]bool{}
	for k, v := range good {
		bad[k] = v
	}
	bad["EntranceCookieCovered"] = false
	if evaluateAcceptSuitePassed(bad) {
		t.Fatal("missing Cookie must fail AcceptSuite")
	}
	if evaluateAcceptSuitePassed(nil) {
		t.Fatal("nil must fail")
	}
}

func TestAssignAuxDepConditions(t *testing.T) {
	conds := map[string]bool{}
	assignAuxDepConditions(conds, false)
	if conds["EntranceAuxCovered"] {
		t.Fatal("expected false")
	}
	assignAuxDepConditions(conds, true)
	if !conds["EntranceAuxCovered"] {
		t.Fatal("expected true")
	}
}

func TestEvaluateEntranceAuxCoveredVacuousAndWS(t *testing.T) {
	if !evaluateEntranceAuxCovered([]entranceAuxTarget{{
		Namespace: "ns", SRRName: "a", HTTPRoute: "a", Needs: entranceAuxNeeds{},
	}}, func(entranceAuxTarget) bool { return false }) {
		t.Fatal("vacuous must be true")
	}
	if evaluateEntranceAuxCovered([]entranceAuxTarget{{
		Namespace: "ns", SRRName: "a", HTTPRoute: "a", Needs: entranceAuxNeeds{WS: true},
	}}, func(entranceAuxTarget) bool { return false }) {
		t.Fatal("ws without ready must be false")
	}
}

func TestImagesRouteBypassesExtAuthCLI(t *testing.T) {
	good := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"labels":      map[string]any{"gateway.olares.io/auth-kind": deenvyAuthKindImages},
			"annotations": map[string]any{deenvyAuxExtAuthBypassAnn: "true"},
		},
		"spec": map[string]any{
			"rules": []any{
				map[string]any{
					"matches": []any{
						map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/images/upload"}},
					},
					"backendRefs": []any{
						map[string]any{"name": deenvyAuxImagesSvc, "port": int64(8080)},
					},
				},
			},
		},
	}}
	if !imagesRouteBypassesExtAuth(good) {
		t.Fatal("expected bypass ready")
	}
	bad := good.DeepCopy()
	bad.SetAnnotations(nil)
	if imagesRouteBypassesExtAuth(bad) {
		t.Fatal("missing bypass ann must fail")
	}
}

func TestProbeBypassObjectsReadyRequiresUALabel(t *testing.T) {
	route := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"annotations": map[string]any{deenvyProbePathsAnn: "/healthz"}},
	}}
	pol := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"labels": map[string]any{}},
		"spec":     map[string]any{"lua": []any{map[string]any{"inline": "x"}}},
	}}
	if probeBypassObjectsReady(route, pol) {
		t.Fatal("missing UA label must fail")
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


