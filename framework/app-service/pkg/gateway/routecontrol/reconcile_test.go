package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func mustHTTPRouteSectionName(t *testing.T, route *unstructured.Unstructured) string {
	t.Helper()
	parentRefs, found, err := unstructured.NestedSlice(route.Object, "spec", "parentRefs")
	if err != nil || !found || len(parentRefs) == 0 {
		t.Fatalf("spec.parentRefs missing: found=%v err=%v", found, err)
	}
	parentRef, ok := parentRefs[0].(map[string]any)
	if !ok {
		t.Fatalf("parentRefs[0] type = %T, want map[string]any", parentRefs[0])
	}
	section, ok := parentRef["sectionName"].(string)
	if !ok {
		t.Fatalf("parentRefs[0].sectionName type = %T, want string", parentRef["sectionName"])
	}
	return section
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := srrv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := networkingv1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	gw := schema.GroupVersion{Group: "gateway.networking.k8s.io", Version: "v1"}
	s.AddKnownTypeWithName(gw.WithKind("HTTPRoute"), &unstructured.Unstructured{})
	s.AddKnownTypeWithName(gw.WithKind("HTTPRouteList"), &unstructured.UnstructuredList{})
	return s
}

func TestResolveServicePort(t *testing.T) {
	svc := &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{
		{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP},
		{Name: "metrics", Port: 9090, Protocol: corev1.ProtocolTCP},
	}}}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{Port: 9090}); err != nil || p != 9090 {
		t.Errorf("by port = %d, %v", p, err)
	}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{PortName: "http"}); err != nil || p != 80 {
		t.Errorf("by name = %d, %v", p, err)
	}
	if p, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{}); err != nil || p != 80 {
		t.Errorf("default = %d, %v", p, err)
	}
	if _, err := resolveServicePort(svc, srrv1alpha1.UpstreamRef{Port: 1234}); err == nil {
		t.Error("missing port should error")
	}
}

func TestReconcileSharedRouteGatewayMode(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-web", Namespace: "demo-shared", UID: "uid-1"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"ab12cd34.shared.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.HTTPRouteName != "shared-demo-web" {
		t.Fatalf("result = %+v", res)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-web"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	if got := mustHTTPRouteSectionName(t, route); got != "http" {
		t.Fatalf("HTTPRoute sectionName = %q, want http", got)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: NetworkPolicyName}, np); err != nil {
		t.Fatalf("NetworkPolicy not created: %v", err)
	}
}

func TestReconcileSharedRouteGatewayModeApplicationSection(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-app", Namespace: "demo-shared", UID: "uid-app"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.viewer.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-app"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	if got := mustHTTPRouteSectionName(t, route); got != "https" {
		t.Fatalf("HTTPRoute sectionName = %q, want https", got)
	}
}

func TestReconcileSharedRouteGatewayModeEmptyEntranceClassDefaultsToShared(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-empty", Namespace: "demo-shared", UID: "uid-empty"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:    srrv1alpha1.RouteModeGateway,
			HostPatterns: []string{"empty.shared.olares.com"},
			Upstream:     srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-shared", Name: "shared-demo-empty"}, route); err != nil {
		t.Fatalf("HTTPRoute not created: %v", err)
	}
	if got := mustHTTPRouteSectionName(t, route); got != "http" {
		t.Fatalf("HTTPRoute sectionName = %q, want http for empty EntranceClass", got)
	}
}

func TestReconcileSharedRouteDirectMode(t *testing.T) {
	s := testScheme(t)
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-demo-web", Namespace: "demo-shared", UID: "uid-2"},
		Spec:       srrv1alpha1.SharedRouteRegistrySpec{RouteMode: srrv1alpha1.RouteModeDirect},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(srr).Build()
	res, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr)
	if err != nil {
		t.Fatalf("ReconcileSharedRoute direct: %v", err)
	}
	if res.Status != metav1.ConditionTrue || res.Reason != ReasonDirectMode {
		t.Fatalf("direct result = %+v", res)
	}
}
