package routecontrol

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const sharedNS = "ollamaserver-shared"

func gatewaySRR(t *testing.T) *srrv1alpha1.SharedRouteRegistry {
	t.Helper()
	srr := logicalSRR("ollamav2-alice", "shared-a5be2268-ollamav2")
	srr.Spec.Upstream.ServiceNamespace = sharedNS
	srr.Spec.Upstream.ServiceName = "sharedentrances-ollama"
	return srr
}

func gatewaySvc() *corev1.Service {
	svc := backendService(sharedNS)
	svc.Name = "sharedentrances-ollama"
	return svc
}

// TestReconcileSharedRoute_GatewayMode_AddsMeshNPAndInject covers the happy path:
// entering gateway mode creates the mesh NP and flips linkerd.io/inject on the
// shared workload namespace.
func TestReconcileSharedRoute_GatewayMode_AddsMeshNPAndInject(t *testing.T) {
	srr := gatewaySRR(t)
	c := newFixture(t, gatewaySvc(), srr)

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	meshNP := getNetworkPolicy(t, c, sharedNS, security.SharedLinkerdMeshIngressNPName)
	if got := meshNP.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]; got != "linkerd" {
		t.Fatalf("expected first ingress peer = linkerd, got %q", got)
	}
	if got := meshNP.Spec.Ingress[0].From[1].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"]; got != "linkerd-viz" {
		t.Fatalf("expected second ingress peer = linkerd-viz, got %q", got)
	}

	ns := getNamespace(t, c, sharedNS)
	if got := ns.Annotations[LinkerdInjectAnnotation]; got != LinkerdInjectEnabled {
		t.Fatalf("linkerd.io/inject = %q, want %q", got, LinkerdInjectEnabled)
	}
}

// TestReconcileSharedRoute_GatewayMode_IngressNPHasNoPortsAndEmptyPodSelector asserts
// NP-minimal v1.0: the app-gateway -> shared ingress NP omits Ports (any TCP
// port is allowed, covering service port, targetPort and linkerd-proxy 4143)
// and uses an empty PodSelector so any pod in the upstream NS is admitted.
func TestReconcileSharedRoute_GatewayMode_IngressNPHasNoPortsAndEmptyPodSelector(t *testing.T) {
	srr := gatewaySRR(t)
	srr.Spec.Upstream.Port = 80
	svc := gatewaySvc()
	svc.Spec.Ports[0].Port = 80
	c := newFixture(t, svc, srr)

	if _, err := ReconcileSharedRoute(context.Background(), c, GatewayRef{}, srr); err != nil {
		t.Fatalf("ReconcileSharedRoute: %v", err)
	}

	np := getNetworkPolicy(t, c, sharedNS, NetworkPolicyName)
	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("expected 1 ingress rule, got %d", len(np.Spec.Ingress))
	}
	if len(np.Spec.Ingress[0].Ports) != 0 {
		t.Fatalf("NP-minimal v1.0: Ports must be empty (allow all TCP), got %#v", np.Spec.Ingress[0].Ports)
	}
	if len(np.Spec.PodSelector.MatchLabels) != 0 || len(np.Spec.PodSelector.MatchExpressions) != 0 {
		t.Fatalf("NP-minimal v1.0: PodSelector must be empty (any pod in upstream NS), got %#v", np.Spec.PodSelector)
	}
	if len(np.OwnerReferences) != 0 {
		t.Fatalf("NP-minimal v1.0: per-NS shared NP must NOT bind ownerRef to a single SRR, got %#v", np.OwnerReferences)
	}
	if _, ok := np.Labels[InstanceLabel]; ok {
		t.Fatalf("NP-minimal v1.0: per-NS shared NP must NOT carry instance label (SRR-scoped), got %#v", np.Labels)
	}
}

// TestReconcileSharedRoute_DirectMode_RemovesMeshNPAndDisablesInject ensures the
// inverse direction: switching to direct cleans up the mesh NP and turns inject off,
// without erasing the operator opt-out annotation.
func TestReconcileSharedRoute_DirectMode_RemovesMeshNPAndDisablesInject(t *testing.T) {
	srr := gatewaySRR(t)
	c := newFixture(t, gatewaySvc(), srr)
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
	ns := getNamespace(t, c, sharedNS)
	if got := ns.Annotations[LinkerdInjectAnnotation]; got != LinkerdInjectDisabled {
		t.Fatalf("linkerd.io/inject = %q, want %q", got, LinkerdInjectDisabled)
	}
}

// TestEnsureSharedNamespaceLinkerdInject_HonorsOperatorOptOut verifies that when an
// operator pins gateway.olares.io/linkerd-inject=disabled, the controller refuses to
// flip linkerd.io/inject in either direction.
func TestEnsureSharedNamespaceLinkerdInject_HonorsOperatorOptOut(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: sharedNS,
			Labels: map[string]string{
				security.NamespaceSharedLabel: "true",
			},
			Annotations: map[string]string{
				AnnotationLinkerdInject: LinkerdInjectDisabled,
			},
		},
	}
	c := plainFixture(t, ns)

	for _, enable := range []bool{true, false} {
		if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, sharedNS, enable); err != nil {
			t.Fatalf("ensure(enable=%v): %v", enable, err)
		}
		got := getNamespace(t, c, sharedNS)
		if _, ok := got.Annotations[LinkerdInjectAnnotation]; ok {
			t.Fatalf("controller wrote linkerd.io/inject despite opt-out (enable=%v): %#v", enable, got.Annotations)
		}
		if v := got.Annotations[AnnotationLinkerdInject]; v != LinkerdInjectDisabled {
			t.Fatalf("opt-out annotation lost: %#v", got.Annotations)
		}
	}
}

// TestEnsureSharedNamespaceLinkerdInject_SkipsNonSharedNamespace guards against an
// accidental SRR pointing at a non-shared namespace (system namespaces such as
// os-framework, or v3 same-namespace shared apps). The controller must leave the
// namespace untouched and not return an error so the parent reconcile keeps going.
func TestEnsureSharedNamespaceLinkerdInject_SkipsNonSharedNamespace(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-framework"}}
	c := plainFixture(t, ns)

	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, "os-framework", true); err != nil {
		t.Fatalf("non-shared namespace should be a soft-skip, got %v", err)
	}
	got := getNamespace(t, c, "os-framework")
	if _, ok := got.Annotations[LinkerdInjectAnnotation]; ok {
		t.Fatalf("controller mutated non-shared namespace: %#v", got.Annotations)
	}
}

// TestEnsureSharedNamespaceLinkerdInject_SkipsMissingNamespace ensures the helper is
// tolerant of fixtures / clusters where the target namespace was deleted out from
// under us. It must not fail the reconcile loop.
func TestEnsureSharedNamespaceLinkerdInject_SkipsMissingNamespace(t *testing.T) {
	c := plainFixture(t)
	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, "does-not-exist", true); err != nil {
		t.Fatalf("missing namespace should be a soft-skip, got %v", err)
	}
}

// TestEnsureSharedNamespaceLinkerdInject_NoOpWhenAlreadyAtDesired ensures the
// controller does not write the namespace when the annotation already matches.
func TestEnsureSharedNamespaceLinkerdInject_NoOpWhenAlreadyAtDesired(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        sharedNS,
			Labels:      map[string]string{security.NamespaceSharedLabel: "true"},
			Annotations: map[string]string{LinkerdInjectAnnotation: LinkerdInjectEnabled},
		},
	}
	c := plainFixture(t, ns)
	rv := getNamespace(t, c, sharedNS).ResourceVersion

	if err := ensureSharedNamespaceLinkerdInject(context.Background(), c, sharedNS, true); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if got := getNamespace(t, c, sharedNS).ResourceVersion; got != rv {
		t.Fatalf("no-op should not bump ResourceVersion: %q -> %q", rv, got)
	}
}

func getNetworkPolicy(t *testing.T, c client.Client, ns, name string) *networkingv1.NetworkPolicy {
	t.Helper()
	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, np); err != nil {
		t.Fatalf("get %s/%s: %v", ns, name, err)
	}
	return np
}

func getNamespace(t *testing.T, c client.Client, name string) *corev1.Namespace {
	t.Helper()
	ns := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: name}, ns); err != nil {
		t.Fatalf("get namespace %s: %v", name, err)
	}
	return ns
}

