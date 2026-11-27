package callerjwt

import (
	"context"
	"maps"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"
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

func TestJWKSPoliciesAllowOnlyMeshedAppGatewayOnBackendPort(t *testing.T) {
	data, err := os.ReadFile("../../../../app-gateway/.olares/config/cluster/deploy/app-gateway-jwks-egress.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var egress networkingv1.NetworkPolicy
	if err := yaml.UnmarshalStrict(data, &egress); err != nil {
		t.Fatal(err)
	}
	ingress := desiredJWKSIngressNP(8444)
	if len(egress.Spec.PolicyTypes) != 1 || egress.Spec.PolicyTypes[0] != networkingv1.PolicyTypeEgress ||
		len(egress.Spec.Egress) != 1 || len(egress.Spec.Egress[0].To) != 1 ||
		len(egress.Spec.Egress[0].Ports) != 1 || len(ingress.Spec.Ingress[0].From) != 1 {
		t.Fatalf("unexpected JWKS policy shape: %+v", egress.Spec)
	}
	dst, src := egress.Spec.Egress[0].To[0], ingress.Spec.Ingress[0].From[0]
	port := egress.Spec.Egress[0].Ports[0]
	if egress.Namespace != "os-gateway" || port.Port == nil || *port.Port != intstr.FromInt(8444) ||
		port.Protocol == nil || *port.Protocol != corev1.ProtocolTCP ||
		dst.NamespaceSelector == nil || !maps.Equal(dst.NamespaceSelector.MatchLabels, map[string]string{corev1.LabelMetadataName: "os-framework"}) ||
		dst.PodSelector == nil || !maps.Equal(dst.PodSelector.MatchLabels, ingress.Spec.PodSelector.MatchLabels) ||
		src.PodSelector == nil || !maps.Equal(src.PodSelector.MatchLabels, egress.Spec.PodSelector.MatchLabels) {
		t.Fatal("JWKS policies disagree on the TLS endpoint or caller boundary")
	}
	selector, err := metav1.LabelSelectorAsSelector(src.PodSelector)
	if err != nil {
		t.Fatal(err)
	}
	gateway := labels.Set{
		"gateway.envoyproxy.io/owning-gateway-name":      "app-gateway",
		"gateway.envoyproxy.io/owning-gateway-namespace": "os-gateway",
		"app.kubernetes.io/component":                    "proxy",
		"app.kubernetes.io/managed-by":                   "envoy-gateway",
		"app.kubernetes.io/name":                         "envoy",
		"linkerd.io/control-plane-ns":                    "os-mesh",
	}
	if !selector.Matches(gateway) {
		t.Fatal("app-gateway data plane is excluded")
	}
	for key := range gateway {
		other := maps.Clone(gateway)
		delete(other, key)
		if selector.Matches(other) {
			t.Errorf("allowed a source missing %s", key)
		}
	}
}
