package routecontrol

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

func TestIsMeshMandatoryCallerNamespace(t *testing.T) {
	cases := []struct {
		ns   string
		want bool
	}{
		{"user-space-alice", true},
		{"user-system-bob", true},
		{"litellm-alice", true},
		{"os-network", false},
		{"app-gateway", false},
		{"linkerd", false},
		{"kube-system", false},
		{"os-framework", false},
	}
	for _, tc := range cases {
		if got := isMeshMandatoryCallerNamespace(tc.ns); got != tc.want {
			t.Fatalf("isMeshMandatoryCallerNamespace(%q) = %v, want %v", tc.ns, got, tc.want)
		}
	}
}

// TestCallerReconciler_optInInjectsAndWritesGatewayIngress is the NP-minimal
// v1.0 happy path: opt-in caller NS gets linkerd.io/inject=enabled and the
// gateway-side singleton ingress NP appears (NO managed caller egress).
func TestCallerReconciler_optInInjectsAndWritesGatewayIngress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-gateway"}},
			&appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name: "litellm",
					Annotations: map[string]string{
						gateway.AnnotationInCluster: gateway.InClusterGateway,
					},
				},
				Spec: appv1alpha1.ApplicationSpec{
					Name:      "litellm",
					Namespace: ns,
					Settings:  map[string]string{"clusterAppRef": "ollamav2"},
				},
			},
		).Build()

	r := &CallerReconciler{Client: c, GatewayNS: "app-gateway"}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("NP-minimal v1.0: caller egress %q must not be created", name)
		}
	}

	var gwNP networkingv1.NetworkPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &gwNP); err != nil {
		t.Fatalf("gateway caller ingress NP missing: %v", err)
	}
	if len(gwNP.Spec.PodSelector.MatchLabels) != 0 || len(gwNP.Spec.PodSelector.MatchExpressions) != 0 {
		t.Fatalf("gateway caller ingress podSelector must be empty, got %#v", gwNP.Spec.PodSelector)
	}
	if len(gwNP.Spec.Ingress) != 1 || len(gwNP.Spec.Ingress[0].Ports) != 0 {
		t.Fatalf("gateway caller ingress must omit Ports, got %#v", gwNP.Spec.Ingress)
	}

	var nsObj corev1.Namespace
	if err := c.Get(context.Background(), types.NamespacedName{Name: ns}, &nsObj); err != nil {
		t.Fatalf("get ns: %v", err)
	}
	if nsObj.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("inject = %q", nsObj.Annotations[LinkerdInjectAnnotation])
	}
}

func TestCallerReconciler_osNetworkNoOp(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-network"}}).
		Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), "os-network"); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	var np networkingv1.NetworkPolicy
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &np); err == nil {
		t.Fatal("os-network reconcile must not create gateway caller ingress NP")
	}
}

// TestCallerReconciler_optInAlsoGCsLegacyEgress is the upgrade-path safety net:
// clusters that ran pre-v1.0 still carry caller egress NPs in opted-in caller
// namespaces. opt-out cleanup never fires while the caller remains opted-in,
// so the opt-in reconcile branch must GC the legacy NPs every loop.
func TestCallerReconciler_optInAlsoGCsLegacyEgress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "app-gateway"}},
			security.NewCallerMeshEgressNP(ns),
			security.NewCallerToAppGatewayEgressNP(ns, "app-gateway"),
			security.NewCallerDNSEgressNP(ns),
			security.NewCallerMiddlewareEgressNP(ns, "user-system-alice"),
			&appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name: "litellm",
					Annotations: map[string]string{
						gateway.AnnotationInCluster: gateway.InClusterGateway,
					},
				},
				Spec: appv1alpha1.ApplicationSpec{
					Name:      "litellm",
					Namespace: ns,
					Settings:  map[string]string{"clusterAppRef": "ollamav2"},
				},
			},
		).Build()

	r := &CallerReconciler{Client: c, GatewayNS: "app-gateway"}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("legacy caller egress %q must be GCed on opt-in reconcile", name)
		}
	}

	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-gateway", Name: security.AppGatewayInClusterCallerIngressNPName}, &networkingv1.NetworkPolicy{}); err != nil {
		t.Fatalf("gateway ingress NP must still be present: %v", err)
	}
}

// TestCallerReconciler_optOutGCsLegacyEgress validates the upgrade-path cleanup:
// any pre-v1.0 caller egress NPs still in the namespace are GCed when the app
// opts out (or is deleted).
func TestCallerReconciler_optOutGCsLegacyEgress(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name:        ns,
				Annotations: map[string]string{LinkerdInjectAnnotation: LinkerdInjectEnabled},
			}},
			security.NewCallerMeshEgressNP(ns),
			security.NewCallerToAppGatewayEgressNP(ns, "app-gateway"),
			security.NewCallerDNSEgressNP(ns),
			security.NewCallerMiddlewareEgressNP(ns, "user-system-alice"),
		).Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	for _, name := range []string{
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
	} {
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &networkingv1.NetworkPolicy{}); err == nil {
			t.Fatalf("legacy NP %q should be GCed on opt-out", name)
		}
	}
}
