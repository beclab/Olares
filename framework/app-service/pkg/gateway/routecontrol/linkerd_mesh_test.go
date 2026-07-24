package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

const sharedNS = "ollamaserver-shared"

func meshGatewaySRR() *srrv1alpha1.SharedRouteRegistry {
	return &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-ollama", Namespace: "user-alice", UID: "uid-mesh-1"},
		Spec: srrv1alpha1.SharedRouteRegistrySpec{
			RouteMode:     srrv1alpha1.RouteModeGateway,
			EntranceClass: srrv1alpha1.EntranceClassShared,
			HostPatterns:  []string{"ab12cd34.shared.olares.com"},
			Upstream: srrv1alpha1.UpstreamRef{
				ServiceName:      "sharedentrances-ollama",
				ServiceNamespace: sharedNS,
				Port:             8080,
			},
		},
	}
}

func meshGatewaySvc() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedentrances-ollama", Namespace: sharedNS},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ollama"},
			Ports:    []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}},
		},
	}
}

func meshSharedNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   sharedNS,
			Labels: map[string]string{security.NamespaceSharedLabel: "true"},
		},
	}
}

func TestReconcileSharedRoute_GatewayMode_AddsMeshNPAndInject(t *testing.T) {
	srr := meshGatewaySRR()
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).
		WithObjects(meshGatewaySvc(), srr, meshSharedNamespace()).Build()

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	meshNP := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: sharedNS, Name: security.SharedLinkerdMeshIngressNPName,
	}, meshNP); err != nil {
		t.Fatalf("mesh NP missing: %v", err)
	}
	got := meshNP.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]
	if got != "os-mesh" {
		t.Fatalf("expected first ingress peer = os-mesh, got %q", got)
	}

	ns := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: sharedNS}, ns); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if got := ns.Annotations[mesh.LinkerdInjectAnnotation]; got != mesh.LinkerdInjectEnabled {
		t.Fatalf("linkerd.io/inject = %q, want %q", got, mesh.LinkerdInjectEnabled)
	}
}

func TestReconcileSharedRoute_DirectMode_RemovesMeshNPAndDisablesInject(t *testing.T) {
	srr := meshGatewaySRR()
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).
		WithObjects(meshGatewaySvc(), srr, meshSharedNamespace()).Build()
	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("seed gateway mode: %v", err)
	}

	srr.Spec.RouteMode = srrv1alpha1.RouteModeDirect
	if err := c.Update(context.Background(), srr); err != nil {
		t.Fatalf("update srr to direct: %v", err)
	}
	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("reconcile direct: %v", err)
	}

	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: sharedNS, Name: security.SharedLinkerdMeshIngressNPName,
	}, &networkingv1.NetworkPolicy{}); !apierrors.IsNotFound(err) {
		t.Fatalf("mesh NP not deleted: err=%v", err)
	}
	ns := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: sharedNS}, ns); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if got := ns.Annotations[mesh.LinkerdInjectAnnotation]; got != mesh.LinkerdInjectDisabled {
		t.Fatalf("linkerd.io/inject = %q, want %q", got, mesh.LinkerdInjectDisabled)
	}
}

func TestEnsureSharedNamespaceLinkerdInject_SkipsNonSharedNamespace(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-framework"}}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(ns).Build()

	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, "os-framework", true); err != nil {
		t.Fatalf("non-shared namespace should be a soft-skip, got %v", err)
	}
	got := &corev1.Namespace{}
	_ = c.Get(context.Background(), types.NamespacedName{Name: "os-framework"}, got)
	if _, ok := got.Annotations[mesh.LinkerdInjectAnnotation]; ok {
		t.Fatalf("controller mutated non-shared namespace: %#v", got.Annotations)
	}
}

func TestEnsureSharedNamespaceLinkerdInject_SkipsMissingNamespace(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).Build()
	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, "does-not-exist", true); err != nil {
		t.Fatalf("missing namespace should be a soft-skip, got %v", err)
	}
}

func TestEnsureSharedNamespaceLinkerdInject_NoOpWhenAlreadyAtDesired(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sharedNS,
			Labels:      map[string]string{security.NamespaceSharedLabel: "true"},
			Annotations: map[string]string{mesh.LinkerdInjectAnnotation: mesh.LinkerdInjectEnabled},
		},
	}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(ns).Build()
	rv := func() string {
		got := &corev1.Namespace{}
		_ = c.Get(context.Background(), types.NamespacedName{Name: sharedNS}, got)
		return got.ResourceVersion
	}()

	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, sharedNS, true); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	got := &corev1.Namespace{}
	_ = c.Get(context.Background(), types.NamespacedName{Name: sharedNS}, got)
	if got.ResourceVersion != rv {
		t.Fatalf("no-op should not bump ResourceVersion: %q -> %q", rv, got.ResourceVersion)
	}
}

// silence unused client import in case helpers evolve
var _ client.Client
