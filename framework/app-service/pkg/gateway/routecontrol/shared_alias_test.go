package routecontrol

import (
	"context"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/callerjwt"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func sharedAliasFixtures() (*srrv1alpha1.SharedRouteRegistry, *srrv1alpha1.SharedRouteRegistry) {
	controller := true
	app := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-f3395cd5-router-frontend", Namespace: "router-shared", UID: "ordinary-uid", Generation: 1,
			Annotations: map[string]string{gateway.AnnotationExtAuthUser: "alice"},
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue, InstanceLabel: "router", "gateway.olares.io/appid": "f3395cd5",
				"gateway.olares.io/entrance": "router-frontend",
			},
			OwnerReferences: []metav1.OwnerReference{{Kind: "Application", UID: "router-app", Controller: &controller}},
		},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"router.alice.olares.com", "f3395cd5.alice.olares.com", "router.alice.olares.local", "f3395cd5.alice.olares.local"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "router-svc", Port: 80},
		},
	}
	shared := app.DeepCopy()
	shared.Name, shared.UID = "shared-f3395cd5-routershare", "shared-uid"
	shared.Spec.EntranceClass = srrv1alpha1.EntranceClassShared
	shared.Spec.HostPatterns = []string{"3a77b15d.shared.olares.com"}
	shared.Spec.Upstream.ServiceName = "sharedentrances-router"
	shared.Status.Conditions = []metav1.Condition{{Type: ConditionReady, Status: metav1.ConditionTrue, ObservedGeneration: 1}}
	return app, shared
}

func TestSharedAliasPublicationAndWithdrawal(t *testing.T) {
	ctx := context.Background()
	app, shared := sharedAliasFixtures()
	scheme := testSchemeWithSecurityPolicy(t)
	c := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(app).WithObjects(app, shared).Build()
	route := aliasObject(app, schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	policy := aliasObject(app, securityPolicyGVK)
	reconcile := func(enabled bool) {
		t.Helper()
		if err := reconcileSharedAlias(ctx, c, GatewayRef{}, app, enabled); err != nil {
			t.Fatal(err)
		}
	}
	read := func(obj client.Object) {
		t.Helper()
		if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
			t.Fatal(err)
		}
	}
	rule := func() map[string]any {
		t.Helper()
		read(route)
		return route.Object["spec"].(map[string]any)["rules"].([]any)[0].(map[string]any)
	}
	reconcile(true)
	parked := rule()
	if parked["backendRefs"] != nil || meta.FindStatusCondition(app.Status.Conditions, sharedAliasCondition) != nil {
		t.Fatal("an unaccepted policy must publish neither a backend nor DNS")
	}
	if matches := parked["matches"]; !reflect.DeepEqual(matches, []any{map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/v1"}}}) {
		t.Fatalf("Router scope changed: %v", matches)
	}
	parents, _, _ := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
	if len(parents) != 1 || parents[0].(map[string]any)["sectionName"] != "http" {
		t.Fatalf("workload route attached to a northbound listener: %v", parents)
	}
	hosts, _, _ := unstructured.NestedStringSlice(route.Object, "spec", "hostnames")
	if !reflect.DeepEqual(hosts, app.Spec.HostPatterns) {
		t.Fatalf("missing alias or hash hosts: %v", hosts)
	}
	read(policy)
	authorization, _, _ := unstructured.NestedMap(policy.Object, "spec", "authorization")
	rules := authorization["rules"].([]any)
	jwtRule := rules[0].(map[string]any)["principal"].(map[string]any)["jwt"].(map[string]any)
	wantClaim := map[string]any{"name": callerjwt.ClaimSharedTargets, "valueType": "StringArray", "values": []any{"alice/router", "alice/@all"}}
	if authorization["defaultAction"] != "Deny" || len(rules) != 1 || jwtRule["provider"] != CallerJWTProviderName ||
		!reflect.DeepEqual(jwtRule["claims"], []any{wantClaim}) {
		t.Fatalf("unexpected tenant/target authorization: %v", authorization)
	}
	accept := func(obj *unstructured.Unstructured, section, refKey string, kinds ...string) {
		t.Helper()
		ref := map[string]any{"namespace": "os-gateway", "name": "app-gateway", "sectionName": "http"}
		var conditions []any
		for _, kind := range kinds {
			conditions = append(conditions, map[string]any{"type": kind, "status": "True", "observedGeneration": obj.GetGeneration()})
		}
		obj.Object["status"] = map[string]any{section: []any{map[string]any{
			refKey: ref, "controllerName": "gateway.envoyproxy.io/gatewayclass-controller", "conditions": conditions,
		}}}
		if err := c.Update(ctx, obj); err != nil {
			t.Fatal(err)
		}
	}
	accept(policy, "ancestors", "ancestorRef", "Accepted")
	reconcile(true)
	if rule()["backendRefs"] != nil {
		t.Fatal("route activated without the inbound identity header filter")
	}
	strip := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.envoyproxy.io/v1alpha1", "kind": "ClientTrafficPolicy",
		"metadata": map[string]any{"namespace": "os-gateway", "name": "app-gateway-strip-identity-headers"},
		"spec": map[string]any{
			"targetRef": map[string]any{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "app-gateway"},
			"headers":   map[string]any{"earlyRequestHeaders": map[string]any{"remove": []any{"x-caller-appid", "x-caller-user"}}},
		},
	}}
	if err := c.Create(ctx, strip); err != nil {
		t.Fatal(err)
	}
	reconcile(true)
	if rule()["backendRefs"] == nil {
		t.Fatal("policy acceptance did not activate the workload backend")
	}
	if meta.FindStatusCondition(app.Status.Conditions, sharedAliasCondition) != nil {
		t.Fatal("DNS became ready before Envoy accepted the active route")
	}
	accept(route, "parents", "parentRef", "Accepted", "ResolvedRefs")
	reconcile(true)
	if !meta.IsStatusConditionTrue(app.Status.Conditions, sharedAliasCondition) {
		t.Fatal("accepted workload route was not published to DNS")
	}
	read(policy)
	if err := c.Delete(ctx, policy); err != nil {
		t.Fatal(err)
	}
	reconcile(true)
	if rule()["backendRefs"] != nil || meta.IsStatusConditionTrue(app.Status.Conditions, sharedAliasCondition) {
		t.Fatal("policy loss left the workload route active")
	}
	reconcile(false)
	if meta.IsStatusConditionTrue(app.Status.Conditions, sharedAliasCondition) {
		t.Fatal("withdrawn workload route remains in DNS")
	}
	read(policy)
	reconcile(false) // Absent route cleanup is idempotent.
}

func TestSharedAliasRequiresReadySiblingWithSameOwner(t *testing.T) {
	app, shared := sharedAliasFixtures()
	for _, tc := range []struct {
		name   string
		change func(*srrv1alpha1.SharedRouteRegistry)
		want   string
	}{
		{name: "known Router contract", want: "/v1"},
		{name: "same backend", change: func(s *srrv1alpha1.SharedRouteRegistry) { s.Spec.Upstream = app.Spec.Upstream }, want: "/"},
		{name: "unknown dedicated backend", change: func(s *srrv1alpha1.SharedRouteRegistry) { s.Spec.Upstream.ServiceName = "another-api" }},
		{name: "another Application owner", change: func(s *srrv1alpha1.SharedRouteRegistry) { s.OwnerReferences[0].UID = "different" }},
		{name: "stale readiness", change: func(s *srrv1alpha1.SharedRouteRegistry) { s.Generation++ }},
		{name: "direct mode", change: func(s *srrv1alpha1.SharedRouteRegistry) { s.Spec.RouteMode = srrv1alpha1.RouteModeDirect }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sibling := shared.DeepCopy()
			if tc.change != nil {
				tc.change(sibling)
			}
			c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(app, sibling).Build()
			prefix, err := sharedAliasPrefix(context.Background(), c, app)
			if err != nil || prefix != tc.want {
				t.Fatalf("prefix=%q, want %q, err=%v", prefix, tc.want, err)
			}
		})
	}
}

func TestWorkloadAliasLocalHostAllowsHTTPForItsOwnTenant(t *testing.T) {
	app, _ := sharedAliasFixtures()
	auth, tls := enumerateHostsForViewer("alice", []srrv1alpha1.SharedRouteRegistry{*app}, "olares.com")
	if len(auth) != 4 || len(tls) != 2 {
		t.Fatalf("unexpected HTTP and TLS hosts: %v, %v", auth, tls)
	}
	auth, _ = enumerateHostsForViewer("bob", []srrv1alpha1.SharedRouteRegistry{*app}, "olares.com")
	if len(auth) != 2 {
		t.Fatalf("another viewer received local tenant names: %v", auth)
	}
}

func TestWorkloadAliasWithdrawalPreservesOtherOwner(t *testing.T) {
	ctx := context.Background()
	app, _ := sharedAliasFixtures()
	previous := app.DeepCopy()
	previous.UID = "previous-owner"
	route := desiredSharedAlias(previous, GatewayRef{}, "/v1")
	c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(route).Build()
	if err := withdrawSharedAlias(ctx, c, app); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(route), route); err != nil {
		t.Fatalf("withdrawal removed a different SRR's route: %v", err)
	}
}
