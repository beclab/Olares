package routecontrol

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

func TestUniqueProbePaths(t *testing.T) {
	got := uniqueProbePaths([]string{"/ready", "", "/healthz", "/ready", "live", "/"})
	if len(got) != 3 || got[0] != "/healthz" || got[1] != "/live" || got[2] != "/ready" {
		t.Fatalf("got %#v", got)
	}
}

func TestDesiredEntranceProbeUAPolicyRequiresUA(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-web", Namespace: "demo"},
		Spec:       srrv1alpha1.SharedRouteRegistrySpec{EntranceClass: srrv1alpha1.EntranceClassApplication},
	}
	pol := desiredEntranceProbeUAPolicy(srr)
	if pol.GetLabels()[probeUARequiredLabel] != "true" {
		t.Fatal("UA required label missing — must not disable ExtAuth without UA guard")
	}
	spec, _, _ := unstructuredNestedMap(pol.Object, "spec")
	luaList, _ := spec["lua"].([]any)
	inline, _ := luaList[0].(map[string]any)["inline"].(string)
	if !strings.Contains(inline, "user-agent") {
		t.Fatal("lua must check user-agent")
	}
}

func TestReconcileCreatesProbeBypassWhenPodsHaveProbes(t *testing.T) {
	s := testScheme(t)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-user"},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "demo"},
			Ports:    []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}},
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-0", Namespace: "demo-user", Labels: map[string]string{"app": "demo"}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "app",
				ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/healthz"},
				}},
			}},
		},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-demo-web", Namespace: "demo-user", UID: "uid-probe"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.demo.olares.com"},
			Upstream:      srrv1alpha1.UpstreamRef{ServiceName: "demo-svc", Port: 8080},
		},
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(svc, srr, pod).Build()
	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: mesh.EntranceProbeBypassRouteName(srr.Name)}, route); err != nil {
		t.Fatalf("probe HTTPRoute missing: %v", err)
	}
	if route.GetAnnotations()[probePathsAnnotation] != "/healthz" {
		t.Fatalf("paths ann = %q", route.GetAnnotations()[probePathsAnnotation])
	}
	pol := &unstructured.Unstructured{}
	pol.SetGroupVersionKind(envoyExtensionPolicyGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: mesh.EntranceProbeBypassPolicyName(srr.Name)}, pol); err != nil {
		t.Fatalf("probe EEP missing: %v", err)
	}
	if pol.GetLabels()[probeUARequiredLabel] != "true" {
		t.Fatal("UA guard not wired")
	}
}
