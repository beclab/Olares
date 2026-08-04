package callerjwt

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testSchemeWithNetworking(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := testScheme(t)
	if err := networkingv1.AddToScheme(s); err != nil {
		t.Fatalf("add networking scheme: %v", err)
	}
	return s
}

func TestReconcileJWKSIngressNP_AC_NP_1(t *testing.T) {
	scheme := testSchemeWithNetworking(t)
	ring, err := NewKeyRingForTest(false)
	if err != nil {
		t.Fatalf("NewKeyRingForTest: %v", err)
	}
	issuer, err := NewIssuer(ring)
	if err != nil {
		t.Fatalf("NewIssuer: %v", err)
	}

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	r := &IssuerReconciler{Client: c, Scheme: scheme, issuer: issuer}
	if err := r.reconcileJWKSSurface(context.Background()); err != nil {
		t.Fatalf("reconcileJWKSSurface: %v", err)
	}

	svc := &corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: JWKSServiceNamespace,
		Name:      JWKSServiceName,
	}, svc); err != nil {
		t.Fatalf("get JWKS service: %v", err)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: JWKSServiceNamespace,
		Name:      JWKSIngressNPName,
	}, np); err != nil {
		t.Fatalf("get JWKS NetworkPolicy: %v", err)
	}
	if np.Labels[managedByLabel] != managedByValue {
		t.Fatalf("managed-by = %q, want %q", np.Labels[managedByLabel], managedByValue)
	}
	if np.Labels[managedByComponentLabel] != JWKSIngressNPComponentValue {
		t.Fatalf("component = %q, want %q", np.Labels[managedByComponentLabel], JWKSIngressNPComponentValue)
	}
	if got := np.Spec.PodSelector.MatchLabels[jwksAppServiceSelectorKey]; got != jwksAppServiceSelectorValue {
		t.Fatalf("podSelector = %q, want %q", got, jwksAppServiceSelectorValue)
	}
	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("ingress rules = %d, want 1", len(np.Spec.Ingress))
	}
	rule := np.Spec.Ingress[0]
	if len(rule.From) != 1 || rule.From[0].NamespaceSelector == nil {
		t.Fatalf("from = %#v", rule.From)
	}
	if got := rule.From[0].NamespaceSelector.MatchLabels[corev1.LabelMetadataName]; got != JWKSIngressNPFromNamespace {
		t.Fatalf("from ns = %q, want %q", got, JWKSIngressNPFromNamespace)
	}
	if len(rule.Ports) != 1 || rule.Ports[0].Port == nil {
		t.Fatalf("ports = %#v", rule.Ports)
	}
	wantPort := intstr.FromInt(8444)
	if *rule.Ports[0].Port != wantPort {
		t.Fatalf("port = %v, want %v", *rule.Ports[0].Port, wantPort)
	}

	// Idempotent update path.
	if err := r.reconcileJWKSIngressNP(context.Background()); err != nil {
		t.Fatalf("reconcileJWKSIngressNP second pass: %v", err)
	}
}

func TestDesiredJWKSIngressNPShape(t *testing.T) {
	np := desiredJWKSIngressNP(8444)
	if np.Name != JWKSIngressNPName || np.Namespace != JWKSServiceNamespace {
		t.Fatalf("meta = %s/%s", np.Namespace, np.Name)
	}
	if len(np.Spec.PolicyTypes) != 1 || np.Spec.PolicyTypes[0] != networkingv1.PolicyTypeIngress {
		t.Fatalf("policyTypes = %#v", np.Spec.PolicyTypes)
	}
}
