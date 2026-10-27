package security

import (
	"reflect"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SharedEntranceNetworkPolicyName identifies the ingress policy for Shared
	// entrance workloads.
	SharedEntranceNetworkPolicyName = "shared-entrance-np"
	// L4ProxyNamespace hosts the single l4-bfl-proxy replica.
	L4ProxyNamespace = "os-network"
	// L4ProxyAppLabel is the app= label on l4-bfl-proxy pods.
	L4ProxyAppLabel = "l4-bfl-proxy"
)

// SharedEntranceOSNamespaces is the frozen allow-list of platform namespaces
// whose pods may reach Shared entrance ClusterIPs. kubernetes namespaceSelector
// cannot match a name prefix, so new os-* namespaces must be added here.
var SharedEntranceOSNamespaces = []string{
	"os-framework",
	"os-platform",
	"os-network",
	"os-gateway",
	"os-mesh",
	"os-protected",
	"os-gpu",
}

// ExcludeSharedEntrancePods appends a NotIn expression so the selector
// no longer matches pods labeled shared-entrance=true. Used only on the
// Shared-namespace copies of app-np / shared-np; the package-level
// NPAppSpace / NPSharedSpace templates stay unchanged.
func ExcludeSharedEntrancePods(sel *metav1.LabelSelector) {
	if sel == nil {
		return
	}
	if hasSharedEntranceNotIn(sel) {
		return
	}
	sel.MatchExpressions = append(sel.MatchExpressions, metav1.LabelSelectorRequirement{
		Key:      constants.AppSharedEntrancesLabel,
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   []string{"true"},
	})
}

func hasSharedEntranceNotIn(sel *metav1.LabelSelector) bool {
	for _, expr := range sel.MatchExpressions {
		if expr.Key == constants.AppSharedEntrancesLabel && expr.Operator == metav1.LabelSelectorOpNotIn {
			return true
		}
	}
	return false
}

// AppendNodeTunnelIngressRule adds the current node tunnel peers as one
// ingress rule. The caller supplies the peers so generation remains separate
// from policy mutation and can be tested without a Kubernetes client.
func AppendNodeTunnelIngressRule(np *netv1.NetworkPolicy, peers []netv1.NetworkPolicyPeer) {
	if np == nil || len(peers) == 0 {
		return
	}
	for _, rule := range np.Spec.Ingress {
		if reflect.DeepEqual(rule.From, peers) {
			return
		}
	}
	np.Spec.Ingress = append(np.Spec.Ingress, netv1.NetworkPolicyIngressRule{
		From: peers,
	})
}

// SharedNamespacePolicies returns the four NetworkPolicies emitted into a
// Shared app namespace. app-np and shared-np exclude entrance pods so the
// remaining OR policies cannot reopen ClusterIP bypass.
func SharedNamespacePolicies() NetworkPolicies {
	appSpace := NPAppSpace.DeepCopy()
	ExcludeSharedEntrancePods(&appSpace.Spec.PodSelector)

	sharedSpace := NPSharedSpace.DeepCopy()
	sharedSpace.Name = "shared-np"
	ExcludeSharedEntrancePods(&sharedSpace.Spec.PodSelector)

	return NetworkPolicies{
		appSpace,
		sharedSpace,
		NPSystemProvider.DeepCopy(),
		NPSharedEntrance.DeepCopy(),
	}
}
