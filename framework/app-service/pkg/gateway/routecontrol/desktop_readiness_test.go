package routecontrol

import (
	"context"
	"fmt"
	"testing"
	"time"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestDesktopReadyTracksLiveProbeFailureAndRecovery(t *testing.T) {
	ctx, srr := t.Context(), desktopEntrance()
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: srr.Namespace, Name: "desktop"},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "http", Port: 80}}}}
	objects := meshedServiceObjects(svc, srr)
	entry := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: defaultGatewayNS, Name: entryMeshService.ServiceName},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "mtls", Port: 8443}, {Name: "authz", Port: 9444}}}}
	objects = append(objects, meshedServiceObjects(entry)...)
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithStatusSubresource(srr).WithObjects(objects...).Build()
	if res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr); err != nil || res.Status != metav1.ConditionTrue {
		t.Fatalf("fixture route failed: %+v %v", res, err)
	}
	conditions := func(types ...string) []any {
		var out []any
		for _, typ := range types {
			out = append(out, map[string]any{"type": typ, "status": "True", "observedGeneration": int64(0)})
		}
		return out
	}
	ancestor := func(section string) map[string]any {
		return map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": section}
	}
	policyStatus := func(section string) map[string]any {
		return map[string]any{"ancestors": []any{map[string]any{"ancestorRef": ancestor(section), "controllerName": "gateway.envoyproxy.io/gatewayclass-controller", "conditions": conditions("Accepted")}}}
	}
	gateway := &unstructured.Unstructured{Object: map[string]any{"status": map[string]any{
		"conditions": conditions("Programmed"), "listeners": []any{
			map[string]any{"name": l4MTLSGatewaySectN, "conditions": conditions("Programmed", "ResolvedRefs")},
			map[string]any{"name": l4AuthzSection, "conditions": conditions("Programmed", "ResolvedRefs")},
		},
	}}}
	gateway.SetGroupVersionKind(gatewayGVK)
	gateway.SetName(defaultGatewayName)
	gateway.SetNamespace(defaultGatewayNS)
	if err := c.Create(ctx, gateway); err != nil {
		t.Fatal(err)
	}
	for name, section := range map[string]string{srr.Name: l4MTLSGatewaySectN, srr.Name + ".l4-authz": l4AuthzSection} {
		route := &unstructured.Unstructured{}
		route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
		if err := c.Get(ctx, client.ObjectKey{Namespace: srr.Namespace, Name: name}, route); err != nil {
			t.Fatal(err)
		}
		route.Object["status"] = map[string]any{"parents": []any{map[string]any{"parentRef": ancestor(section), "controllerName": "gateway.envoyproxy.io/gatewayclass-controller", "conditions": conditions("Accepted", "ResolvedRefs")}}}
		if err := c.Update(ctx, route); err != nil {
			t.Fatal(err)
		}
	}
	for _, gvk := range []schema.GroupVersionKind{securityPolicyGVK} {
		policy := &unstructured.Unstructured{}
		policy.SetGroupVersionKind(gvk)
		name := securityPolicyName(srr)
		if err := c.Get(ctx, client.ObjectKey{Namespace: srr.Namespace, Name: name}, policy); err != nil {
			t.Fatal(err)
		}
		policy.Object["status"] = policyStatus(l4MTLSGatewaySectN)
		if err := c.Update(ctx, policy); err != nil {
			t.Fatal(err)
		}
	}
	for name, section := range map[string]string{"app-gateway-l4-entry-mtls": l4MTLSGatewaySectN, "app-gateway-l4-authz-mtls": l4AuthzSection} {
		p := &unstructured.Unstructured{Object: map[string]any{"spec": map[string]any{
			"headers":   map[string]any{"earlyRequestHeaders": entryIdentityHeaderSettings()},
			"targetRef": map[string]any{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": defaultGatewayName, "sectionName": section},
			"tls": map[string]any{"clientValidation": map[string]any{"mode": "RequireAndVerify",
				"caCertificateRefs": []any{map[string]any{"kind": "Secret", "name": "olares-l4-agw-mtls"}},
				"subjectAltNames":   map[string]any{"dnsNames": []any{map[string]any{"type": "Exact", "value": "l4-bfl-proxy.entry.client"}}},
			}},
		}, "status": policyStatus(section)}}
		p.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"})
		p.SetName(name)
		p.SetNamespace(defaultGatewayNS)
		if err := c.Create(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	if err := servingReadiness(ctx, c, GatewayRef{}, srr, srr.Name); err != nil {
		t.Fatal("fixture must pass every structural check", err)
	}
	var failure error
	probes := 0
	r := &SharedRouteReconciler{Client: c, desktopProbe: func(context.Context, client.Reader, *srrv1alpha1.SharedRouteRegistry) error { probes++; return failure }}
	for _, code := range []int{400, 0, 503, 0} {
		failure = nil
		if code != 0 {
			failure = fmt.Errorf("Desktop Entry probe returned HTTP %d", code)
		}
		result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(srr)})
		if err != nil {
			t.Fatal(err)
		}
		live := &srrv1alpha1.SharedRouteRegistry{}
		if err := c.Get(ctx, client.ObjectKeyFromObject(srr), live); err != nil {
			t.Fatal(err)
		}
		want, reason, retry := metav1.ConditionTrue, ReasonDesktopEntryAuthorizationReady, 30*time.Second
		if failure != nil {
			want, reason, retry = metav1.ConditionFalse, ReasonAwaitingDataPlane, 5*time.Second
		}
		if len(live.Status.Conditions) != 1 || live.Status.Conditions[0].Status != want || live.Status.Conditions[0].Reason != reason || result.RequeueAfter != retry {
			t.Fatalf("HTTP %d produced incorrect readiness: %+v %+v", code, live.Status, result)
		}
	}
	if probes != 4 {
		t.Fatalf("expected fresh probes on every reconcile, got %d", probes)
	}
	// Production must use the fresh API reader and real probe when no test
	// callback is supplied. Missing credentials/configuration must fail closed.
	r.desktopProbe = nil
	r.ReadinessReader = fake.NewClientBuilder().WithScheme(testScheme(t)).Build()
	if _, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(srr)}); err != nil {
		t.Fatal(err)
	}
	live := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(ctx, client.ObjectKeyFromObject(srr), live); err != nil || live.Status.Conditions[0].Status != metav1.ConditionFalse {
		t.Fatal("missing probe prerequisites did not revoke readiness", err)
	}
	if err := r.servingReadiness(ctx, srr, "missing-route"); err == nil {
		t.Fatal("missing serving route was accepted")
	}
	// Direct mode must not advertise the managed Desktop as ready after it
	// removes the protected Entry route.
	srr.Spec.RouteMode = srrv1alpha1.RouteModeDirect
	res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil || res.Status != metav1.ConditionFalse || res.Reason != ReasonInvalidSpec {
		t.Fatalf("Desktop direct mode was accepted: %+v %v", res, err)
	}
}
