package routecontrol

import (
	"context"
	"reflect"
	"testing"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestGatewayTransportScopesAndWithdrawsOnBackendLoss(t *testing.T) {
	ctx := context.Background()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-alice", Name: "demo"},
		Spec: corev1.ServiceSpec{Selector: map[string]string{"app": "demo"}, ClusterIP: "10.96.0.12",
			Ports: []corev1.ServicePort{{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(8080)}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Namespace: svc.Namespace, Name: "demo", UID: "route-uid"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode: srrv1alpha1.RouteModeGateway, EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns: []string{"demo.example.org"}, Upstream: srrv1alpha1.UpstreamRef{ServiceName: svc.Name, Port: 80},
		},
	}
	legacy := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{
		Namespace: svc.Namespace, Name: NetworkPolicyName,
		Labels: map[string]string{ManagedByLabel: ManagedByValue, RouteControlComponentLabel: RouteControlComponentValue},
	}}
	objs := meshedServiceObjects(svc, srr, legacy)
	objs[2].(*discoveryv1.EndpointSlice).Ports[0].Port = ptr.To(int32(8080))
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(objs...).Build()
	res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil || res.Status != metav1.ConditionTrue {
		t.Fatalf("route did not reconcile: %+v %v", res, err)
	}
	var policies netv1.NetworkPolicyList
	selector := client.MatchingLabels(transportLabels(client.ObjectKeyFromObject(srr)))
	if err := c.List(ctx, &policies, selector); err != nil || len(policies.Items) != 2 {
		t.Fatalf("want one transport pair: %+v %v", policies.Items, err)
	}
	for _, policy := range policies.Items {
		for _, rule := range policy.Spec.Ingress {
			if len(rule.From) != 1 || rule.From[0].PodSelector.MatchLabels["gateway.envoyproxy.io/owning-gateway-name"] != "app-gateway" ||
				!reflect.DeepEqual(rule.Ports, []netv1.NetworkPolicyPort{meshTransportPort(4143), meshTransportPort(8080)}) {
				t.Fatalf("backend ingress widened: %+v", rule)
			}
		}
		for _, rule := range policy.Spec.Egress {
			for _, peer := range rule.To {
				if peer.IPBlock != nil {
					if peer.IPBlock.CIDR != "10.96.0.12/32" || !reflect.DeepEqual(rule.Ports, []netv1.NetworkPolicyPort{meshTransportPort(80)}) {
						t.Fatalf("pre-DNAT route widened: %+v", rule)
					}
				} else if peer.PodSelector.MatchLabels["app"] != "demo" ||
					peer.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != svc.Namespace ||
					!reflect.DeepEqual(rule.Ports, []netv1.NetworkPolicyPort{meshTransportPort(4143), meshTransportPort(8080)}) {
					t.Fatalf("post-DNAT route widened: %+v", rule)
				}
			}
		}
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(legacy), &netv1.NetworkPolicy{}); !apierrors.IsNotFound(err) {
		t.Fatalf("legacy wildcard ingress was retained: %v", err)
	}
	if err := c.Delete(ctx, svc); err != nil {
		t.Fatal(err)
	}
	res, err = ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil || res.Reason != ReasonBackendMissing {
		t.Fatalf("missing backend not withdrawn: %+v %v", res, err)
	}
	if err := c.List(ctx, &policies, selector); err != nil || len(policies.Items) != 0 {
		t.Fatalf("stale transport permits remained: %+v %v", policies.Items, err)
	}
}

func TestGatewayTransportRejectsForeignWriterAndPrunesDeletedRegistry(t *testing.T) {
	ctx := context.Background()
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "backend-ns", Name: "target"},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080}}}}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Namespace: svc.Namespace, Name: "route", UID: "route-uid"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{HostPatterns: []string{"app.example.org"},
			Upstream: srrv1alpha1.UpstreamRef{Port: 8080, ServiceName: svc.Name}},
	}
	key := client.ObjectKeyFromObject(srr)
	foreign := &netv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: "os-gateway",
		Name: "olares-route-" + transportLabels(key)[routeTransportLabel] + "-http-egress"}}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(meshedServiceObjects(svc, srr, foreign)...).Build()
	res, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil || res.Reason != ReasonAwaitingDataPlane {
		t.Fatalf("foreign policy was adopted: %+v %v", res, err)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(foreign), foreign); err != nil || len(foreign.Labels) != 0 {
		t.Fatalf("foreign policy was changed or removed: %v", err)
	}
	if err := c.Delete(ctx, foreign); err != nil {
		t.Fatal(err)
	}
	res, err = ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
	if err != nil || res.Status != metav1.ConditionTrue {
		t.Fatalf("route did not recover after collision: %+v %v", res, err)
	}
	if err := c.Delete(ctx, srr); err != nil {
		t.Fatal(err)
	}
	if _, err := (&SharedRouteReconciler{Client: c}).Reconcile(ctx, reconcile.Request{NamespacedName: key}); err != nil {
		t.Fatal(err)
	}
	var policies netv1.NetworkPolicyList
	if err := c.List(ctx, &policies, client.MatchingLabels(transportLabels(key))); err != nil || len(policies.Items) != 0 {
		t.Fatalf("deleted registry retained transport: %v", err)
	}
}

func TestTransportRejectsReservedProxyPortsAndInvalidService(t *testing.T) {
	ctx := context.Background()
	for _, port := range []int32{4140, 4143, 4190, 4191, 3389} {
		if hasInterceptedHTTPPort([]discoveryv1.EndpointPort{{Name: ptr.To("http"), Port: &port}}, "http", 0) {
			t.Fatalf("reserved proxy port %d admitted as a workload backend", port)
		}
	}
	for _, selector := range []map[string]string{nil, {"linkerd.io/control-plane-ns": "untrusted"}, {"app": "valid"}} {
		svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "backend-ns", Name: "target"},
			Spec: corev1.ServiceSpec{Selector: selector, Ports: []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolUDP}}}}
		if _, err := transportBackend(ctx, nil, svc, 80); err == nil {
			t.Fatalf("invalid selector or non-TCP service enabled: %+v", svc.Spec)
		}
	}
}
