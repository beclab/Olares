package security

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Traffic group label for namespace isolation (L3/L4). Cross-group HTTP must
// traverse the gateway (ViaAppGateway) or an explicitly compiled peer allow.
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
	case ns == "os-network":
		// Edge L4 (shared l4-bfl-proxy) lives in os-network.
		return TrafficGroupL4
	case ns == "os-platform":
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

// TrafficGroupNPEnabled reports whether operators opted into applying NP bundles.
func TrafficGroupNPEnabled(raw string) bool {
	v := strings.TrimSpace(raw)
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes")
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

func npMeta(name, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "app-service",
			"app.kubernetes.io/name":       "traffic-group",
		},
	}
}

// IsTrafficGroupManagedNP identifies the exact policies reconciled by the
// traffic-group writer, including policies created by earlier versions.
func IsTrafficGroupManagedNP(np *netv1.NetworkPolicy) bool {
	if np == nil || np.Labels["app.kubernetes.io/managed-by"] != "app-service" ||
		np.Labels["app.kubernetes.io/name"] != "traffic-group" {
		return false
	}
	for _, policy := range BuildTrafficGroupIsolationBundle(np.Namespace) {
		if policy.Name == np.Name {
			return true
		}
	}
	for _, policy := range BuildTrafficGroupDependencyPolicies(np.Namespace, nil, nil) {
		if policy.Name == np.Name {
			return true
		}
	}
	return false
}

func nsPeer(group string) netv1.NetworkPolicyPeer {
	return netv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{
			TrafficGroupLabel: group,
		}},
	}
}

// NPGroupDefaultDeny returns ingress+egress deny-all baseline, retaining DNS and
// Linkerd control-plane egress so sidecars can start. Callers MUST also apply
// peer allow policies (same-group / gateway / L4) before relying on this object.
func NPGroupDefaultDeny(namespace string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-traffic-group-default-deny", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress, netv1.PolicyTypeEgress},
			Ingress:     []netv1.NetworkPolicyIngressRule{},
			Egress: []netv1.NetworkPolicyEgressRule{
				{
					To:    []netv1.NetworkPolicyPeer{nsPeer(TrafficGroupMesh)},
					Ports: []netv1.NetworkPolicyPort{tcpPort(8080), tcpPort(8086), tcpPort(8090)},
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
		ObjectMeta: npMeta("olares-traffic-group-same", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{nsPeer(group)},
			}},
		},
	}
}

// NPAllowFromGateway permits ingress from the gateway traffic group (ViaAppGateway / N-S).
func NPAllowFromGateway(namespace string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-traffic-group-from-gateway", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{nsPeer(TrafficGroupGateway)},
			}},
		},
	}
}

// NPAllowFromL4 permits ingress from the L4 edge group into gateway (north-south).
func NPAllowFromL4(namespace string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-traffic-group-from-l4", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{nsPeer(TrafficGroupL4)},
			}},
		},
	}
}

// NPAllowFromMesh permits ingress from the mesh control-plane group (probes / tap as needed).
func NPAllowFromMesh(namespace string) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-traffic-group-from-mesh", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			Ingress: []netv1.NetworkPolicyIngressRule{{
				From: []netv1.NetworkPolicyPeer{nsPeer(TrafficGroupMesh)},
			}},
		},
	}
}

// BuildTrafficGroupIsolationBundle returns the safe NP set for a namespace.
// Cross-group east-west is denied at L3 except via gateway (and L4→gateway for N-S).
// Mesh identity authorization remains a separate layer (Server / AuthorizationPolicy).
func BuildTrafficGroupIsolationBundle(namespace string) []netv1.NetworkPolicy {
	group := InferTrafficGroup(namespace)
	out := []netv1.NetworkPolicy{
		NPGroupDefaultDeny(namespace),
		NPAllowSameTrafficGroup(namespace, group),
		NPAllowFromMesh(namespace),
	}
	switch group {
	case TrafficGroupGateway:
		out = append(out, NPAllowFromL4(namespace))
	case TrafficGroupApp, TrafficGroupUserSpace, TrafficGroupOSFramework, TrafficGroupOSSystem:
		out = append(out, NPAllowFromGateway(namespace))
	}
	return out
}

// BundleHasInboundAllow reports whether the bundle includes at least one ingress allow
// besides the empty default-deny (guards against shipping deny-only policies).
func BundleHasInboundAllow(policies []netv1.NetworkPolicy) bool {
	for _, p := range policies {
		for _, rule := range p.Spec.Ingress {
			if len(rule.From) > 0 || len(rule.Ports) > 0 {
				return true
			}
		}
	}
	return false
}
