package security

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewAppGatewayInClusterCallerIngressNP_Defaults(t *testing.T) {
	np := NewAppGatewayInClusterCallerIngressNP("")
	if np.Namespace != "app-gateway" {
		t.Fatalf("default namespace = %q, want app-gateway", np.Namespace)
	}
	if np.Name != AppGatewayInClusterCallerIngressNPName {
		t.Fatalf("name = %q, want %q", np.Name, AppGatewayInClusterCallerIngressNPName)
	}
}

// TestNewAppGatewayInClusterCallerIngressNP_MinimalFields locks in the
// NP-minimal v1.0 invariants for the single gateway ingress NP:
//   - empty PodSelector (every pod in app-gateway NS)
//   - namespaceSelector matches bytetrade.io/ns-owner Exists (any opted-in caller)
//   - no Ports (any TCP port: 80, 443, 4143, future)
//   - managed-by: app-service label so IsManagedNetworkPolicy keeps it.
func TestNewAppGatewayInClusterCallerIngressNP_MinimalFields(t *testing.T) {
	np := NewAppGatewayInClusterCallerIngressNP("app-gateway")

	if len(np.Spec.PodSelector.MatchLabels) != 0 || len(np.Spec.PodSelector.MatchExpressions) != 0 {
		t.Fatalf("PodSelector must be empty, got %#v", np.Spec.PodSelector)
	}

	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("expected 1 ingress rule, got %d", len(np.Spec.Ingress))
	}
	rule := np.Spec.Ingress[0]
	if len(rule.Ports) != 0 {
		t.Fatalf("Ports must be empty (any TCP), got %#v", rule.Ports)
	}
	if len(rule.From) != 1 {
		t.Fatalf("expected 1 From peer, got %d", len(rule.From))
	}
	ns := rule.From[0].NamespaceSelector
	if ns == nil || len(ns.MatchExpressions) != 1 {
		t.Fatalf("expected single MatchExpressions on namespaceSelector, got %#v", ns)
	}
	expr := ns.MatchExpressions[0]
	if expr.Key != NamespaceOwnerLabel || expr.Operator != metav1.LabelSelectorOpExists {
		t.Fatalf("expected ns-owner Exists, got key=%q op=%q", expr.Key, expr.Operator)
	}

	if np.Labels["app.kubernetes.io/managed-by"] != callerNPManagedBy {
		t.Fatalf("managed-by label missing, got %#v", np.Labels)
	}
}
