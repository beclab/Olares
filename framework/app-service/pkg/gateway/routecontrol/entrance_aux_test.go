package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

func TestDesiredEntranceImagesHTTPRouteBypassesExtAuth(t *testing.T) {
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "app-web", Namespace: "demo-user"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			EntranceClass: srrv1alpha1.EntranceClassApplication,
			HostPatterns:  []string{"app.demo.olares.com"},
		},
	}
	route := desiredEntranceImagesHTTPRoute(GatewayRef{}, srr, "user-system-alice")
	if route.GetLabels()["gateway.olares.io/auth-kind"] != mesh.AuthKindEntranceImages {
		t.Fatal("auth-kind")
	}
	if route.GetAnnotations()[mesh.AuxExtAuthBypassAnnotation] != "true" {
		t.Fatal("/images/upload route must annotate ExtAuth bypass")
	}
	if !mesh.ImagesRouteBypassesExtAuth(route) {
		t.Fatal("desired images route must pass ImagesRouteBypassesExtAuth")
	}
	rules, _, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	rm := rules[0].(map[string]any)
	backends := rm["backendRefs"].([]any)
	bm := backends[0].(map[string]any)
	if bm["name"] != "tapr-images-svc" || bm["namespace"] != "user-system-alice" || bm["port"].(int64) != 8080 {
		t.Fatalf("backend = %#v", bm)
	}
}

func TestReconcileCreatesAuxWSAndImages(t *testing.T) {
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
			Containers: []corev1.Container{
				{Name: "app"},
				{Name: constants.WsContainerName},
				{Name: constants.UploadContainerName},
			},
		},
	}
	srr := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-demo-web", Namespace: "demo-user", UID: "uid-aux",
			Labels: map[string]string{"applications.app.bytetrade.io/owner": "alice"},
		},
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

	helper := &corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: "demo-svc-deenvy-aux"}, helper); err != nil {
		t.Fatalf("aux helper Service: %v", err)
	}
	ports := map[string]int32{}
	for _, p := range helper.Spec.Ports {
		ports[p.Name] = p.Port
	}
	if ports["ws"] != 40010 || ports["upload"] != 40030 {
		t.Fatalf("helper ports = %#v", ports)
	}

	main := &unstructured.Unstructured{}
	main.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: srr.Name}, main); err != nil {
		t.Fatalf("main HTTPRoute: %v", err)
	}
	ann := main.GetAnnotations()[mesh.AuxCapabilitiesAnnotation]
	if ann != "ws,upload,images" {
		t.Fatalf("aux caps ann = %q", ann)
	}
	if !mesh.HTTPRouteHasAuxPathRule(main, "/ws", "-deenvy-aux", 40010) {
		t.Fatal("main route missing /ws")
	}
	if !mesh.HTTPRouteHasAuxPathRule(main, "/upload", "-deenvy-aux", 40030) {
		t.Fatal("main route missing /upload")
	}

	images := &unstructured.Unstructured{}
	images.SetGroupVersionKind(schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"})
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: mesh.EntranceImagesRouteName(srr.Name)}, images); err != nil {
		t.Fatalf("images HTTPRoute: %v", err)
	}
	if !mesh.ImagesRouteBypassesExtAuth(images) {
		t.Fatal("images route must bypass ExtAuth")
	}
	grant := &unstructured.Unstructured{}
	grant.SetGroupVersionKind(referenceGrantGVK)
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "user-system-alice",
		Name:      "allow-httproute-images-demo-user",
	}, grant); err != nil {
		t.Fatalf("images ReferenceGrant: %v", err)
	}
}
