package routecontrol

import (
	"context"
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func TestRouteEgressDoesNotGrantSystemDependenciesToOtherSources(t *testing.T) {
	for _, source := range []meshTransportPeer{
		{"user-space-alice", map[string]string{"app": "systemserver", "linkerd.io/control-plane-ns": "os-mesh"}},
		{"user-system-", map[string]string{"app": "systemserver", "linkerd.io/control-plane-ns": "os-mesh"}},
		{"user-system-alice", map[string]string{"app": "systemserver"}},
		{"user-system-alice", map[string]string{"app": "other", "linkerd.io/control-plane-ns": "os-mesh"}},
	} {
		backend := meshTransportBackend{
			peer:        meshTransportPeer{"os-framework", map[string]string{"app": "authelia-backend"}},
			targetPorts: []netv1.NetworkPolicyPort{meshTransportPort(9091)},
		}
		policies, err := authTransportPair(t.Context(), nil, transportLabels(client.ObjectKey{Namespace: source.namespace, Name: "test"}), "auth-server", source, backend)
		if err != nil || len(policies[0].Spec.Egress) != 1 {
			t.Fatalf("non-system-server source received an extra egress rule: %+v", source)
		}
	}
}

func TestSystemLoginStartsBackendTransportBeforeFrontend(t *testing.T) {
	ctx := context.Background()
	authBackends := autheliaMeshServiceObjects("user-system-alice")
	for _, object := range authBackends {
		if svc, ok := object.(*corev1.Service); ok {
			svc.Spec.Type = corev1.ServiceTypeClusterIP
			if svc.Namespace == "user-system-alice" {
				svc.Spec.Selector["app"] = "systemserver"
			}
		}
	}
	dns := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "kube-system", Name: "coredns"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.10",
			Selector: map[string]string{"k8s-app": "kube-dns"},
			Ports: []corev1.ServicePort{
				{Name: "dns", Port: 53, Protocol: corev1.ProtocolUDP},
				{Name: "dns-tcp", Port: 53, Protocol: corev1.ProtocolTCP},
				{Name: "metrics", Port: 9153, Protocol: corev1.ProtocolTCP},
			},
		},
	}
	infisical := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-protected", Name: "infisical-service"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.11",
			Selector: map[string]string{"app": "infisical"},
			Ports: []corev1.ServicePort{
				{Port: 80, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(8088)},
				{Port: 8080, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(8080)},
			},
		},
	}
	bfl := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-space-alice", Name: "bfl"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.12",
			Selector: map[string]string{"tier": "bfl"},
			Ports:    []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(8080)}},
		},
	}
	appService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-framework", Name: "app-service"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.13",
			Selector: map[string]string{"tier": "app-service"},
			Ports: []corev1.ServicePort{
				{Port: 6755, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(6755)},
				{Port: 8433, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt32(8433)},
				{Port: 8444, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromString("jwks")},
			},
		},
	}
	authBackends = append(authBackends, dns, infisical, bfl, appService)
	c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(authBackends...).Build()
	if err := gateway.EnsureSystemPlaneRoutes(ctx, c, "alice", "alice.olares.com", "user-space-alice", false); err != nil {
		t.Fatal(err)
	}
	key := client.ObjectKey{Namespace: "user-space-alice", Name: "sys-auth"}
	srr := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(ctx, key, srr); err != nil {
		t.Fatal(err)
	}
	selector := client.MatchingLabels(transportLabels(key))
	policies := &netv1.NetworkPolicyList{}
	verify := func(count int) {
		t.Helper()
		if err := c.List(ctx, policies, selector); err != nil || len(policies.Items) != count {
			t.Fatalf("want %d login transport policies, got %d: %v", count, len(policies.Items), err)
		}
	}
	checkRoute := func(want string) {
		t.Helper()
		result, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, srr)
		if err != nil || result.Reason != want {
			t.Fatalf("login reconcile = %+v, %v; want %q", result, err, want)
		}
	}
	checkRoute(ReasonAwaitingDataPlane)
	verify(2)
	var systemEgress *netv1.NetworkPolicy
	for _, p := range policies.Items {
		if p.Namespace == "user-system-alice" {
			systemEgress = p.DeepCopy()
			if !reflect.DeepEqual(p.Spec.PodSelector.MatchLabels, map[string]string{
				"app": "systemserver", "linkerd.io/control-plane-ns": "os-mesh",
			}) {
				t.Fatalf("unexpected bootstrap egress source: %+v", p.Spec.PodSelector)
			}
			if !reflect.DeepEqual(p.Spec.Egress, []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{(meshTransportPeer{"os-framework", map[string]string{
						"app": "authelia-backend", "linkerd.io/control-plane-ns": "os-mesh",
					}}).networkPolicyPeer()},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(4143), meshTransportPort(9091)},
				},
				{
					To: []netv1.NetworkPolicyPeer{(meshTransportPeer{"os-platform", map[string]string{
						"app": "citus", "app.kubernetes.io/name": "citus", "statefulset.kubernetes.io/pod-name": "citus-0",
					}}).networkPolicyPeer()},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(5432)},
				},
				{
					To: []netv1.NetworkPolicyPeer{
						(meshTransportPeer{"kube-system", map[string]string{"k8s-app": "kube-dns"}}).networkPolicyPeer(),
						{IPBlock: &netv1.IPBlock{CIDR: "10.96.0.10/32"}},
					},
					Ports: []netv1.NetworkPolicyPort{
						meshTransportPort(53),
						{Protocol: ptr.To(corev1.ProtocolUDP), Port: ptr.To(intstr.FromInt32(53))},
					},
				},
				{
					To:    []netv1.NetworkPolicyPeer{(meshTransportPeer{"os-protected", map[string]string{"app": "infisical"}}).networkPolicyPeer()},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(8080)},
				},
				{
					To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.96.0.11/32"}}},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(8080)},
				},
				{
					To: []netv1.NetworkPolicyPeer{(meshTransportPeer{"user-space-alice", map[string]string{
						"tier": "bfl", "linkerd.io/control-plane-ns": "os-mesh", "traffic.olares.io/bfl-api-mtls": "required",
					}}).networkPolicyPeer()},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(8080), meshTransportPort(4143)},
				},
				{
					To: []netv1.NetworkPolicyPeer{(meshTransportPeer{"os-framework", map[string]string{
						"tier": "app-service", "linkerd.io/control-plane-ns": "os-mesh",
						"traffic.olares.io/app-service-api-mtls": "required",
					}}).networkPolicyPeer()},
					Ports: []netv1.NetworkPolicyPort{meshTransportPort(6755), meshTransportPort(4143)},
				},
			}) {
				t.Fatalf("unexpected system-server bootstrap dependencies: %+v", p.Spec.Egress)
			}
		}
		if p.Namespace == "os-framework" &&
			(p.Spec.PodSelector.MatchLabels["app"] != "authelia-backend" ||
				p.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "user-system-alice" ||
				p.Spec.Ingress[0].From[0].PodSelector.MatchLabels["app"] != "systemserver") {
			t.Fatalf("bootstrap ingress grants an unexpected identity: %+v", p.Spec)
		}
	}
	if systemEgress == nil {
		t.Fatal("login bootstrap did not install the system-server egress policy")
	}
	wantEgress := systemEgress.DeepCopy().Spec.Egress
	systemEgress.Spec.Egress = systemEgress.Spec.Egress[:3]
	if err := c.Update(ctx, systemEgress); err != nil {
		t.Fatal(err)
	}
	checkRoute(ReasonAwaitingDataPlane)
	if err := c.Get(ctx, client.ObjectKeyFromObject(systemEgress), systemEgress); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(systemEgress.Spec.Egress, wantEgress) {
		t.Fatalf("old login egress policy was not repaired in place: %+v", systemEgress.Spec.Egress)
	}
	// Removing the Infisical provider port and changing the BFL target must
	// withdraw both grants while preserving the original dependencies.
	if err := c.Get(ctx, client.ObjectKeyFromObject(infisical), infisical); err != nil {
		t.Fatal(err)
	}
	infisical.Spec.Ports = infisical.Spec.Ports[:1]
	if err := c.Update(ctx, infisical); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(ctx, client.ObjectKeyFromObject(bfl), bfl); err != nil {
		t.Fatal(err)
	}
	bfl.Spec.Ports[0].TargetPort = intstr.FromInt32(8081)
	if err := c.Update(ctx, bfl); err != nil {
		t.Fatal(err)
	}
	if err := c.Delete(ctx, appService); err != nil {
		t.Fatal(err)
	}
	checkRoute(ReasonAwaitingDataPlane)
	if err := c.Get(ctx, client.ObjectKeyFromObject(systemEgress), systemEgress); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(systemEgress.Spec.Egress, wantEgress[:3]) {
		t.Fatalf("stale provider grants were retained: %+v", systemEgress.Spec.Egress)
	}
	frontend := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: "authelia-svc"},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolTCP}}}}
	for _, object := range meshedServiceObjects(frontend) {
		if err := c.Create(ctx, object); err != nil {
			t.Fatal(err)
		}
	}
	checkRoute(ReasonReconciled)
	verify(4)

	desktop := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: key.Namespace, Name: "sys-desktop"}, desktop); err != nil {
		t.Fatal(err)
	}
	desktopService := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: "edge-desktop"},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(91)}}}}
	desktopBackend := meshedServiceObjects(desktopService)
	desktopBackend[2].(*discoveryv1.EndpointSlice).Ports[0].Port = ptr.To(int32(91))
	for _, object := range desktopBackend {
		if err := c.Create(ctx, object); err != nil {
			t.Fatal(err)
		}
	}
	result, err := ReconcileSharedRoute(ctx, c, GatewayRef{}, desktop)
	if err != nil || result.Reason != ReasonReconciled {
		t.Fatalf("desktop reconcile = %+v, %v", result, err)
	}
	for _, name := range []string{"sys-auth", "sys-desktop"} {
		httpRoute := &unstructured.Unstructured{}
		httpRoute.SetGroupVersionKind(l4AuthzRoute(desktop).GroupVersionKind())
		if err := c.Get(ctx, client.ObjectKey{Namespace: key.Namespace, Name: name}, httpRoute); err != nil {
			t.Fatal(err)
		}
		mustIncludeParentSection(t, httpRoute, l4MTLSGatewaySectN)
		parents, _, _ := unstructured.NestedSlice(httpRoute.Object, "spec", "parentRefs")
		if len(parents) != 1 {
			t.Fatalf("%s is also attached outside Entry mTLS: %v", name, parents)
		}
	}
	authRoute := l4AuthzRoute(desktop)
	if err := c.Get(ctx, client.ObjectKeyFromObject(authRoute), authRoute); err != nil {
		t.Fatal(err)
	}
	mustIncludeParentSection(t, authRoute, l4AuthzSection)
	hosts, _, _ := unstructured.NestedStringSlice(authRoute.Object, "spec", "hostnames")
	rules, _, _ := unstructured.NestedSlice(authRoute.Object, "spec", "rules")
	refs, _, _ := unstructured.NestedSlice(rules[0].(map[string]any), "backendRefs")
	ref := refs[0].(map[string]any)
	if !reflect.DeepEqual(hosts, []string{"desktop.alice.olares.com", "desktop.alice.olares.local"}) ||
		ref["namespace"] != "user-system-alice" || ref["name"] != AutheliaBackendServiceName {
		t.Fatalf("desktop authorizer is not bound to its canonical tenant: hosts=%v ref=%v", hosts, ref)
	}

	backend := &corev1.Service{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: "os-framework", Name: AutheliaBackendServiceName}, backend); err != nil {
		t.Fatal(err)
	}
	backend.Spec.Selector["app"] = "unexpected"
	if err := c.Update(ctx, backend); err != nil {
		t.Fatal(err)
	}
	checkRoute(ReasonAwaitingDataPlane)
	verify(0)
	backend.Spec.Selector["app"] = "authelia-backend"
	if err := c.Update(ctx, backend); err != nil {
		t.Fatal(err)
	}
	checkRoute(ReasonReconciled)
	verify(4)
}

func TestSystemLoginBootstrapDoesNotTrustForeignTenant(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{ObjectMeta: metav1.ObjectMeta{
		Name: "sys-auth", Namespace: "user-space-alice",
		Labels: map[string]string{
			ManagedByLabel: ManagedByValue, InstanceLabel: gateway.SystemPlaneInstance,
			gateway.LabelSystemPlane: gateway.LabelSystemPlaneValue,
		},
		Annotations: map[string]string{gateway.AnnotationExtAuthUser: "bob", gateway.AnnotationAuthMode: gateway.AuthModeNone},
	}, Spec: srrv1alpha1.SharedRouteRegistrySpec{
		EntranceClass: srrv1alpha1.EntranceClassApplication,
		Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "authelia-svc", Port: 80},
	}}
	policies, err := systemLoginTransport(context.Background(), nil, srr)
	if err != nil || len(policies) != 0 {
		t.Fatalf("foreign tenant received login transport: %+v %v", policies, err)
	}
}
