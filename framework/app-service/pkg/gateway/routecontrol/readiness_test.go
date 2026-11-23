package routecontrol

import (
	"context"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestEntryRouteReadinessRequiresCurrentEntryListenerStatus(t *testing.T) {
	cond := func(kind string, generation int64) map[string]any {
		return map[string]any{"type": kind, "status": "True", "observedGeneration": generation}
	}
	route := &unstructured.Unstructured{Object: map[string]any{
		"status": map[string]any{"parents": []any{map[string]any{
			"controllerName": "gateway.envoyproxy.io/gatewayclass-controller",
			"parentRef":      map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": l4MTLSGatewaySectN},
			"conditions":     []any{cond("Accepted", 2), cond("ResolvedRefs", 2)},
		}}},
	}}
	route.SetNamespace("user-space-alice")
	route.SetGeneration(2)
	if !entryParentReady(route, GatewayRef{}) {
		t.Fatal("matching current-generation status must be ready")
	}
	route.SetGeneration(3)
	if entryParentReady(route, GatewayRef{}) {
		t.Fatal("status from an earlier generation must block cutover")
	}
	route.SetGeneration(2)
	parent := route.Object["status"].(map[string]any)["parents"].([]any)[0].(map[string]any)
	parent["parentRef"].(map[string]any)["sectionName"] = "https"
	if entryParentReady(route, GatewayRef{}) {
		t.Fatal("status for the regular HTTPS listener must not approve entry mTLS cutover")
	}
}

func TestServingReadinessWaitsForPolicyAndMeshedEndpoints(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	_ = discoveryv1.AddToScheme(scheme)
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-alice", Name: "srr"},
		Spec:       srrv1alpha1.SharedRouteRegistrySpec{Upstream: srrv1alpha1.UpstreamRef{ServiceName: "app"}},
	}
	obj := func(gvk schema.GroupVersionKind, ns, name string, status map[string]any) *unstructured.Unstructured {
		u := &unstructured.Unstructured{Object: map[string]any{"status": status}}
		u.SetGroupVersionKind(gvk)
		u.SetNamespace(ns)
		u.SetName(name)
		u.SetGeneration(1)
		return u
	}
	cond := func(kinds ...string) []any {
		var out []any
		for _, k := range kinds {
			out = append(out, map[string]any{"type": k, "status": "True", "observedGeneration": int64(1)})
		}
		return out
	}
	ref := map[string]any{"name": defaultGatewayName, "namespace": defaultGatewayNS, "sectionName": l4MTLSGatewaySectN}
	owner := "gateway.envoyproxy.io/gatewayclass-controller"
	gw := obj(gatewayGVK, defaultGatewayNS, defaultGatewayName, map[string]any{
		"conditions": cond("Programmed"), "listeners": []any{map[string]any{"name": l4MTLSGatewaySectN, "conditions": cond("Programmed", "ResolvedRefs")}},
	})
	route := obj(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}, srr.Namespace, "route", map[string]any{
		"parents": []any{map[string]any{"parentRef": ref, "controllerName": owner, "conditions": cond("Accepted", "ResolvedRefs")}},
	})
	policyStatus := map[string]any{"ancestors": []any{map[string]any{"ancestorRef": ref, "controllerName": owner, "conditions": cond("Accepted")}}}
	security := obj(securityPolicyGVK, srr.Namespace, securityPolicyName(srr), policyStatus)
	ctp := obj(schema.GroupVersionKind{Group: "gateway.envoyproxy.io", Version: "v1alpha1", Kind: "ClientTrafficPolicy"}, defaultGatewayNS, "app-gateway-l4-entry-mtls", policyStatus)
	objects := []client.Object{gw, route, security, ctp}
	for _, ep := range []struct{ ns, svc string }{{defaultGatewayNS, "app-gateway-l4-mtls"}, {srr.Namespace, "app"}} {
		uid := types.UID(ep.svc + "-uid")
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: ep.ns, Name: ep.svc, UID: uid},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{Name: "linkerd-proxy", Ready: true}}}}
		slice := &discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Namespace: ep.ns, Name: ep.svc,
			Labels: map[string]string{discoveryv1.LabelServiceName: ep.svc}},
			Endpoints: []discoveryv1.Endpoint{{TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: ep.ns, Name: ep.svc, UID: uid},
				Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}}}}
		objects = append(objects, pod, slice)
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); err != nil {
		t.Fatal(err)
	}
	pod := &corev1.Pod{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: srr.Namespace, Name: "app"}, pod); err != nil {
		t.Fatal(err)
	}
	pod.Status.ContainerStatuses[0].Ready = false
	if err := c.Status().Update(ctx, pod); err != nil {
		t.Fatal(err)
	}
	if err := servingReadiness(ctx, c, GatewayRef{}, srr, "route"); err == nil {
		t.Fatal("an unready proxy on a serving endpoint must stop cutover")
	}
}
