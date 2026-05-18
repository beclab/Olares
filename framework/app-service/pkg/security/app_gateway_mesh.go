package security

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

const (
	// AppGatewayMeshNPName is the supplemental NetworkPolicy managed by app-service for mesh.
	AppGatewayMeshNPName = "app-gateway-mesh-np"
)

// LinkerdMeshIngressPortsFromAppGateway are TCP ports on linkerd control-plane pods that
// app-gateway data-plane proxies must reach (identity, destination, policy, webhooks).
var LinkerdMeshIngressPortsFromAppGateway = []int32{8080, 8086, 8090, 9443, 443}

// NewAppGatewayMeshNetworkPolicy builds the supplemental ingress policy for app-gateway / Linkerd.
// It is applied in addition to others-np (union of rules); others-np is unchanged.
func NewAppGatewayMeshNetworkPolicy(ns, peerNS string) *netv1.NetworkPolicy {
	ingress := netv1.NetworkPolicyIngressRule{
		From: []netv1.NetworkPolicyPeer{
			{
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": peerNS,
					},
				},
			},
		},
	}
	if ns == "linkerd" {
		tcp := (*corev1.Protocol)(pointer.String(string(corev1.ProtocolTCP)))
		for _, port := range LinkerdMeshIngressPortsFromAppGateway {
			ingress.Ports = append(ingress.Ports, netv1.NetworkPolicyPort{
				Protocol: tcp,
				Port:     intstrPtr(port),
			})
		}
	}

	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppGatewayMeshNPName,
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-mesh",
				"app.kubernetes.io/managed-by": "app-service",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress:     []netv1.NetworkPolicyIngressRule{ingress},
		},
	}
}

func intstrPtr(port int32) *intstr.IntOrString {
	v := intstr.FromInt32(port)
	return &v
}
