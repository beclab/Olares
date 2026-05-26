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

func TestCallerReconciler_optInWritesNPAndInject(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
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
	for _, name := range []string{security.CallerMeshEgressNPName, security.CallerToAppGatewayEgressNPName} {
		var np networkingv1.NetworkPolicy
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: name}, &np); err != nil {
			t.Fatalf("get %s: %v", name, err)
		}
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
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "os-network", Name: security.CallerMeshEgressNPName}, &np)
	if err == nil {
		t.Fatal("os-network must not get caller NP")
	}
}

func TestCallerReconciler_optOutCleansUp(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appv1alpha1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	ns := "user-space-alice"
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: ns,
				Annotations: map[string]string{LinkerdInjectAnnotation: LinkerdInjectEnabled},
			}},
			security.NewCallerMeshEgressNP(ns),
			security.NewCallerToAppGatewayEgressNP(ns, "app-gateway"),
			&appv1alpha1.Application{
				ObjectMeta: metav1.ObjectMeta{Name: "litellm"},
				Spec: appv1alpha1.ApplicationSpec{
					Name:      "litellm",
					Namespace: ns,
					Settings:  map[string]string{"clusterAppRef": "ollamav2"},
				},
			},
		).Build()
	r := &CallerReconciler{Client: c}
	if err := r.Reconcile(context.Background(), ns); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: ns, Name: security.CallerMeshEgressNPName}, &networkingv1.NetworkPolicy{})
	if err == nil {
		t.Fatal("mesh NP should be deleted on opt-out")
	}
}
