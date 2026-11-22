package security

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Traffic group label for namespace isolation (L3/L4). Cross-group HTTP must
// traverse mesh-authorized paths once policy compilation is enabled.
const (
	TrafficGroupLabel = "traffic.olares.io/group"

	TrafficGroupL4          = "l4"
	TrafficGroupOSSystem    = "os-system"
	TrafficGroupOSFramework = "os-framework"
	TrafficGroupUserSpace   = "user-space"
	TrafficGroupApp         = "app"
	TrafficGroupMesh        = "mesh"
	TrafficGroupGateway     = "gateway"
)

// InferTrafficGroup maps legacy namespace naming onto traffic groups.
func InferTrafficGroup(ns string) string {
	switch {
	case ns == "os-mesh" || ns == "linkerd":
		return TrafficGroupMesh
	case ns == "os-gateway":
		return TrafficGroupGateway
	case ns == "os-network" || ns == "os-platform":
		return TrafficGroupOSSystem
	case strings.HasPrefix(ns, "os-"):
		return TrafficGroupOSFramework
	case strings.HasPrefix(ns, "user-space-"), strings.HasPrefix(ns, "user-system-"):
		return TrafficGroupUserSpace
	default:
		return TrafficGroupApp
	}
}

// LabelNamespaceTrafficGroup sets traffic.olares.io/group without removing other labels.
func LabelNamespaceTrafficGroup(ns *corev1.Namespace) {
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if _, ok := ns.Labels[TrafficGroupLabel]; ok {
		return
	}
	ns.Labels[TrafficGroupLabel] = InferTrafficGroup(ns.Name)
}

func tcpPort(p int32) netv1.NetworkPolicyPort {
	proto := corev1.ProtocolTCP
	port := intstr.FromInt(int(p))
	return netv1.NetworkPolicyPort{Protocol: &proto, Port: &port}
}

func udpPort(p int32) netv1.NetworkPolicyPort {
	proto := corev1.ProtocolUDP
	port := intstr.FromInt(int(p))
	return netv1.NetworkPolicyPort{Protocol: &proto, Port: &port}
}

// NPGroupDefaultDeny returns ingress+egress deny-all for pods in a namespace,
// retaining DNS and Linkerd control-plane egress so sidecars can start.
func NPGroupDefaultDeny(namespace string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "olares-traffic-group-default-deny",
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "app-service",
				"app.kubernetes.io/name":       "traffic-group",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress, netv1.PolicyTypeEgress},
			Ingress:     []netv1.NetworkPolicyIngressRule{},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To: []netv1.NetworkPolicyPeer{{
						NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
							TrafficGroupLabel: TrafficGroupMesh,
						}},
					}},
					Ports: []netv1.NetworkPolicyPort{
						tcpPort(8080),
						tcpPort(8086),
						tcpPort(8090),
					},
				},
				{
					Ports: []netv1.NetworkPolicyPort{tcpPort(53), udpPort(53)},
				},
			},
		},
	}
}

// NPAllowSameTrafficGroup permits ingress from namespaces sharing the same group label.
func NPAllowSameTrafficGroup(namespace, group string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "olares-traffic-group-same",
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "app-service",
				"app.kubernetes.io/name":       "traffic-group",
			},
		},
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{{
					NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
						TrafficGroupLabel: group,
					}},
				}},
			}},
		},
	}
}
