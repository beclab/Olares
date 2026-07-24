package security

import (
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SharedLinkerdMeshIngressNPName allows the Linkerd control plane to reach
	// meshed proxies in shared workload namespaces (identity/policy callbacks).
	SharedLinkerdMeshIngressNPName = "shared-linkerd-mesh-ingress-np"
)

// SharedLinkerdMeshIngressPeerNamespaces are platform namespaces whose pods must
// reach linkerd-proxy in shared workload namespaces. Order: control plane first.
// RP installs Linkerd into os-mesh (not the upstream default "linkerd").
var SharedLinkerdMeshIngressPeerNamespaces = []string{"os-mesh"}

// NewSharedLinkerdControlPlaneIngressNetworkPolicy allows Linkerd control-plane
// pods to reach meshed proxies in a shared workload namespace.
// podSelector nil/empty selects all pods (required when multiple SRRs share one NP).
func NewSharedLinkerdControlPlaneIngressNetworkPolicy(namespace string, podSelector map[string]string) *netv1.NetworkPolicy {
	from := make([]netv1.NetworkPolicyPeer, 0, len(SharedLinkerdMeshIngressPeerNamespaces))
	for _, peer := range SharedLinkerdMeshIngressPeerNamespaces {
		from = append(from, netv1.NetworkPolicyPeer{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": peer,
				},
			},
		})
	}
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SharedLinkerdMeshIngressNPName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-mesh",
				"app.kubernetes.io/managed-by": "app-service",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: podSelector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress:     []netv1.NetworkPolicyIngressRule{{From: from}},
		},
	}
}
