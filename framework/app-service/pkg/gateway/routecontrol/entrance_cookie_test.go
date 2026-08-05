package routecontrol

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/sandbox/sidecar"
)

func TestDesiredEntranceCookiePolicy(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-demo-web", Namespace: "demo-alice"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"demo.alice.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo"},
		},
	}
	pol := desiredEntranceCookiePolicy(srr)
	if pol.GetName() != mesh.EntranceCookiePolicyName(srr.Name) {
		t.Fatalf("name = %q", pol.GetName())
	}
	if pol.GetLabels()["gateway.olares.io/auth-kind"] != mesh.AuthKindEntranceCookie {
		t.Fatalf("labels = %#v", pol.GetLabels())
	}
	spec, _, _ := unstructuredNestedMap(pol.Object, "spec")
	target, _ := spec["targetRef"].(map[string]any)
	if name, _ := target["name"].(string); name != "app-demo-web" {
		t.Fatalf("targetRef.name = %q", name)
	}
	luaList, ok := spec["lua"].([]any)
	if !ok || len(luaList) != 1 {
		t.Fatalf("lua = %#v", spec["lua"])
	}
	lua0, _ := luaList[0].(map[string]any)
	if typ, _ := lua0["type"].(string); typ != "Inline" {
		t.Fatalf("lua.type = %q", typ)
	}
	inline, _ := lua0["inline"].(string)
	want := strings.TrimSpace(sidecar.EnvoySetCookieLua())
	if inline != want {
		t.Fatalf("lua inline must match set_cookie_tpl (len got=%d want=%d)", len(inline), len(want))
	}
	if !strings.Contains(inline, "reset_cookie_domain") || !strings.Contains(inline, "Domain=") {
		t.Fatal("lua missing Domain clear semantics")
	}
}

func TestNeedsEntranceCookieMatchesExtAuth(t *testing.T) {
	app := &srrv1alpha1.SharedRouteRegistry{
		Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassApplication},
	}
	shared := &srrv1alpha1.SharedRouteRegistry{
		Spec: srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassShared},
	}
	if !needsEntranceCookie(app) || needsEntranceCookie(shared) || needsEntranceCookie(nil) {
		t.Fatal("cookie needs must match ExtAuth set")
	}
}

func TestReconcileCreatesEntranceCookieEEP(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-demo-web", Namespace: "demo-user", UID: "uid-cookie"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.demo.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr).Build()
	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(envoyExtensionPolicyGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: mesh.EntranceCookiePolicyName(srr.Name)}, got); err != nil {
		t.Fatalf("Cookie EEP not created: %v", err)
	}
}

func TestReconcileDeletesCookieEEPInDirectMode(t *testing.T) {
	s := testScheme(t)
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-demo-web", Namespace: "demo-user", UID: "uid-cookie-del"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeDirect,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.demo.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	eep := desiredEntranceCookiePolicy(srr)
	eep.SetGroupVersionKind(envoyExtensionPolicyGVK)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(srr, eep).Build()
	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	got := &unstructured.Unstructured{}
	got.SetGroupVersionKind(envoyExtensionPolicyGVK)
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: mesh.EntranceCookiePolicyName(srr.Name)}, got)
	if err == nil {
		t.Fatal("expected Cookie EEP deleted in direct mode")
	}
}
