package security

import (
	"reflect"
	"slices"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSharedEntranceOSNamespacesFrozen(t *testing.T) {
	want := []string{
		"os-framework",
		"os-platform",
		"os-network",
		"os-gateway",
		"os-mesh",
		"os-protected",
		"os-gpu",
	}
	if !reflect.DeepEqual(SharedEntranceOSNamespaces, want) {
		t.Fatalf("SharedEntranceOSNamespaces=%v, want %v", SharedEntranceOSNamespaces, want)
	}
	if slices.Contains(SharedEntranceOSNamespaces, "os-mesh-viz") {
		t.Fatal("os-mesh-viz must not be in the os-* allow-list")
	}
}

func TestExcludeSharedEntrancePodsNil(t *testing.T) {
	ExcludeSharedEntrancePods(nil)
}

func TestExcludeSharedEntrancePodsIdempotent(t *testing.T) {
	sel := &metav1.LabelSelector{}
	ExcludeSharedEntrancePods(sel)
	ExcludeSharedEntrancePods(sel)
	if !hasSharedEntranceNotIn(sel) {
		t.Fatal("expected NotIn shared-entrance after exclude")
	}
	if n := countSharedEntranceNotIn(sel); n != 1 {
		t.Fatalf("expected one NotIn expression, got %d", n)
	}
}

func TestAppendNodeTunnelIngressRuleIdempotent(t *testing.T) {
	np := NPSharedEntrance.DeepCopy()
	peers := []netv1.NetworkPolicyPeer{
		{IPBlock: &netv1.IPBlock{CIDR: "10.233.98.1/32"}},
		{IPBlock: &netv1.IPBlock{CIDR: "10.233.87.0/32"}},
		{IPBlock: &netv1.IPBlock{CIDR: "10.233.110.0/32"}},
	}

	AppendNodeTunnelIngressRule(np, peers)
	AppendNodeTunnelIngressRule(np, peers)

	if len(np.Spec.Ingress) != 2 {
		t.Fatalf("ingress rules=%d, want 2", len(np.Spec.Ingress))
	}
	if !reflect.DeepEqual(np.Spec.Ingress[1].From, peers) {
		t.Fatalf("node tunnel peers=%v, want %v", np.Spec.Ingress[1].From, peers)
	}
}

func TestAppendNodeTunnelIngressRuleNilAndEmpty(t *testing.T) {
	AppendNodeTunnelIngressRule(nil, []netv1.NetworkPolicyPeer{
		{IPBlock: &netv1.IPBlock{CIDR: "10.233.98.1/32"}},
	})

	np := NPSharedEntrance.DeepCopy()
	AppendNodeTunnelIngressRule(np, nil)
	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("ingress rules=%d, want 1", len(np.Spec.Ingress))
	}
}

func TestNPAppSpaceTemplateUnchanged(t *testing.T) {
	if hasSharedEntranceNotIn(&NPAppSpace.Spec.PodSelector) {
		t.Fatal("NPAppSpace package template must stay empty (regular app NS)")
	}
	if hasSharedEntranceNotIn(&NPSharedSpace.Spec.PodSelector) {
		t.Fatal("NPSharedSpace package template must stay empty")
	}
	_ = SharedNamespacePolicies()
	if hasSharedEntranceNotIn(&NPAppSpace.Spec.PodSelector) {
		t.Fatal("SharedNamespacePolicies mutated NPAppSpace template")
	}
	if hasSharedEntranceNotIn(&NPSharedSpace.Spec.PodSelector) {
		t.Fatal("SharedNamespacePolicies mutated NPSharedSpace template")
	}
}

func TestNPSharedEntranceAllowList(t *testing.T) {
	np := NPSharedEntrance
	if got := np.Spec.PodSelector.MatchLabels[constants.AppSharedEntrancesLabel]; got != "true" {
		t.Fatalf("podSelector shared-entrance=%q, want true", got)
	}
	if len(np.Spec.PolicyTypes) != 1 || np.Spec.PolicyTypes[0] != netv1.PolicyTypeIngress {
		t.Fatalf("policyTypes=%v, want [Ingress]", np.Spec.PolicyTypes)
	}
	if len(np.Spec.Ingress) != 1 {
		t.Fatalf("ingress rules=%d, want 1", len(np.Spec.Ingress))
	}
	from := np.Spec.Ingress[0].From
	if len(from) != 3 {
		t.Fatalf("from peers=%d, want 3", len(from))
	}
	for i, peer := range from {
		if peer.IPBlock != nil {
			t.Fatalf("peer%d must not set IPBlock, got %+v", i, peer.IPBlock)
		}
	}

	if from[0].PodSelector == nil || from[0].NamespaceSelector != nil {
		t.Fatalf("peer0 should be same-NS empty podSelector, got %+v", from[0])
	}
	if len(from[0].PodSelector.MatchLabels) != 0 || len(from[0].PodSelector.MatchExpressions) != 0 {
		t.Fatalf("peer0 podSelector must be empty, got %+v", from[0].PodSelector)
	}

	nsSel := from[1].NamespaceSelector
	if nsSel == nil || from[1].PodSelector != nil {
		t.Fatalf("peer1 should be os-* namespaceSelector only, got %+v", from[1])
	}
	if len(nsSel.MatchExpressions) != 1 {
		t.Fatalf("peer1 matchExpressions=%d, want 1", len(nsSel.MatchExpressions))
	}
	expr := nsSel.MatchExpressions[0]
	if expr.Key != "kubernetes.io/metadata.name" || expr.Operator != metav1.LabelSelectorOpIn {
		t.Fatalf("peer1 expression key/op = %s %s", expr.Key, expr.Operator)
	}
	if !reflect.DeepEqual(expr.Values, SharedEntranceOSNamespaces) {
		t.Fatalf("peer1 In values=%v, want %v", expr.Values, SharedEntranceOSNamespaces)
	}
	if slices.Contains(expr.Values, "os-mesh-viz") {
		t.Fatal("peer1 must not include os-mesh-viz")
	}

	if from[2].NamespaceSelector == nil || from[2].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != L4ProxyNamespace {
		t.Fatalf("peer2 ns should be %s, got %+v", L4ProxyNamespace, from[2])
	}
	if from[2].PodSelector == nil || from[2].PodSelector.MatchLabels["app"] != L4ProxyAppLabel {
		t.Fatalf("peer2 pod should be app=%s, got %+v", L4ProxyAppLabel, from[2].PodSelector)
	}
	if isEmptyIngressRule(np.Spec.Ingress[0]) {
		t.Fatal("shared-entrance-np must not use empty {} allow-all")
	}
}

func TestSharedNamespacePoliciesExcludeEntrance(t *testing.T) {
	nps := SharedNamespacePolicies()
	if len(nps) != 4 {
		t.Fatalf("policies=%d, want 4", len(nps))
	}
	appNP := nps.Main()
	if appNP == nil {
		t.Fatal("missing app-np main")
	}
	if !hasSharedEntranceNotIn(&appNP.Spec.PodSelector) {
		t.Fatal("app-np copy must exclude shared-entrance")
	}

	var sharedNP, providerNP, entranceNP *netv1.NetworkPolicy
	for _, np := range nps.Additional() {
		switch np.Name {
		case "shared-np":
			sharedNP = np
		case "system-provider-np":
			providerNP = np
		case "shared-entrance-np":
			entranceNP = np
		}
	}
	if sharedNP == nil || !hasSharedEntranceNotIn(&sharedNP.Spec.PodSelector) {
		t.Fatal("shared-np copy must exclude shared-entrance")
	}
	if providerNP == nil || hasSharedEntranceNotIn(&providerNP.Spec.PodSelector) {
		t.Fatal("system-provider-np must not get the entrance exclude")
	}
	if entranceNP == nil {
		t.Fatal("missing shared-entrance-np")
	}
	if isEmptyIngressRule(entranceNP.Spec.Ingress[0]) {
		t.Fatal("shared-entrance-np must keep the tightened From list")
	}
	for i, peer := range entranceNP.Spec.Ingress[0].From {
		if peer.IPBlock != nil {
			t.Fatalf("entrance peer%d must not set IPBlock, got %+v", i, peer.IPBlock)
		}
	}
}

func countSharedEntranceNotIn(sel *metav1.LabelSelector) int {
	n := 0
	for _, expr := range sel.MatchExpressions {
		if expr.Key == constants.AppSharedEntrancesLabel && expr.Operator == metav1.LabelSelectorOpNotIn {
			n++
		}
	}
	return n
}

func isEmptyIngressRule(rule netv1.NetworkPolicyIngressRule) bool {
	return len(rule.From) == 0 && len(rule.Ports) == 0
}
