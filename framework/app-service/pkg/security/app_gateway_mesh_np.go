package security

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// AppGatewayMeshNPName admits mesh peers to Linkerd CP (os-mesh) and
	// admits os-mesh into os-gateway. Chart twin:
	// framework/app-gateway/.olares/config/cluster/deploy/app-gateway-mesh-np-*.yaml
	AppGatewayMeshNPName = "app-gateway-mesh-np"

	// MeshControlPlaneNamespace is where Linkerd identity/destination run.
	MeshControlPlaneNamespace = "os-mesh"
	// AppGatewayNamespace hosts app-gateway-data / EG.
	AppGatewayNamespace = "os-gateway"
)

// NewAppGatewayMeshNPOsMesh allows caller/shared/gateway peers to reach Linkerd
// control-plane ports in os-mesh (identity 8080, policy, etc.).
func NewAppGatewayMeshNPOsMesh() *netv1.NetworkPolicy {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppGatewayMeshNPName,
			Namespace: MeshControlPlaneNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-mesh",
				"app.kubernetes.io/managed-by": "app-service",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{
					{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": AppGatewayNamespace,
					}}},
					{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": "os-mesh-viz",
					}}},
					{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						NamespaceSharedLabel: "true",
					}}},
					{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						NamespaceInClusterCallerLabel: "true",
					}}},
				},
				Ports: appGatewayMeshNPPorts(),
			}},
		},
	}
}

// NewAppGatewayMeshNPOsGateway allows os-mesh proxies to reach os-gateway.
func NewAppGatewayMeshNPOsGateway() *netv1.NetworkPolicy {
	return &netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AppGatewayMeshNPName,
			Namespace: AppGatewayNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "app-gateway",
				"app.kubernetes.io/component":  "linkerd-mesh",
				"app.kubernetes.io/managed-by": "app-service",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{
					{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"kubernetes.io/metadata.name": MeshControlPlaneNamespace,
					}}},
				},
			}},
		},
	}
}

func appGatewayMeshNPPorts() []netv1.NetworkPolicyPort {
	tcp := corev1.ProtocolTCP
	ports := []int32{8080, 8086, 8090, 9443, 443}
	out := make([]netv1.NetworkPolicyPort, 0, len(ports))
	for _, p := range ports {
		port := intstr.FromInt32(p)
		out = append(out, netv1.NetworkPolicyPort{Protocol: &tcp, Port: &port})
	}
	return out
}
