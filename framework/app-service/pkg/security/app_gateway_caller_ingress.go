package security

import (
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppGatewayInClusterCallerIngressNPName is the managed ingress NetworkPolicy in the
// app-gateway namespace that admits traffic from all in-cluster caller workload
// namespaces toward the gateway data plane.
//
// requirement: per the NP-minimal scheme,
// callers in user-space-*, user-system-* and <app>-<user> namespaces must reach
// the Envoy Gateway data plane on every port (HTTP 80/443 + linkerd 4143 +
// future ports) without enumerating each caller namespace.
//
// behavior:
//   - PodSelector empty: covers every pod in app-gateway namespace (EG data plane,
//     ext_authz, sidecars, future workloads); avoids re-reconcile when ext_authz
//     is split into a separate Deployment.
//   - NamespaceSelector matches bytetrade.io/ns-owner Exists: static, no Watch on
//     Application required; gateway is still PEP, ext_authz still enforces host-user.
//   - Ports omitted: any TCP port is allowed from caller namespaces (includes 80,
//     443, 4143 mesh inbound).
const AppGatewayInClusterCallerIngressNPName = "app-gateway-incluster-caller-ingress-np"

// NewAppGatewayInClusterCallerIngressNP builds the single ingress NP that admits
// all opted-in caller namespaces into the gateway namespace.
//
// gatewayNS is the namespace where the NP is written (defaults to "app-gateway"
// when empty).
func NewAppGatewayInClusterCallerIngressNP(gatewayNS string) *netv1.NetworkPolicy {
	if gatewayNS == "" {
		gatewayNS = "app-gateway"
	}
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppGatewayInClusterCallerIngressNPName,
			Namespace: gatewayNS,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": callerNPManagedBy,
				"app.kubernetes.io/component":  callerNPComponentLabel,
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{
				{
					From: []netv1.NetworkPolicyPeer{
						{
							NamespaceSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      NamespaceOwnerLabel,
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
