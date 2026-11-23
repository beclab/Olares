package security

import (
	"cmp"
	"net/netip"
	"slices"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KubernetesAPINamespace  = "default"
	KubernetesAPIName       = "kubernetes"
	coreDNSIngressName      = "olares-traffic-group-coredns-ingress"
	kubernetesAPIEgressName = "olares-traffic-group-kube-api-egress"
)

func BuildTrafficGroupDependencyPolicies(namespace string, service *corev1.Service, endpoints *corev1.Endpoints) []netv1.NetworkPolicy {
	var selector map[string]string
	var policies []netv1.NetworkPolicy
	switch namespace {
	case metav1.NamespaceSystem:
		selector = map[string]string{"k8s-app": "kube-dns"}
		policies = append(policies, netv1.NetworkPolicy{
			ObjectMeta: npMeta(coreDNSIngressName, namespace),
			Spec: netv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: selector},
				PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
				Ingress: []netv1.NetworkPolicyIngressRule{{
					From:  []netv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{}}},
					Ports: []netv1.NetworkPolicyPort{tcpPort(53), udpPort(53)},
				}},
			},
		})
	case "os-framework":
		selector = map[string]string{"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true"}
	default:
		return nil
	}
	return append(policies, netv1.NetworkPolicy{
		ObjectMeta: npMeta(kubernetesAPIEgressName, namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: selector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress:      kubernetesAPIEgress(service, endpoints),
		},
	})
}

// kubernetesAPIEgress allows only the discovered HTTPS addresses and their own
// ports, on both sides of Service DNAT. Missing or invalid data grants nothing.
func kubernetesAPIEgress(service *corev1.Service, endpoints *corev1.Endpoints) []netv1.NetworkPolicyEgressRule {
	if service == nil || service.Namespace != KubernetesAPINamespace || service.Name != KubernetesAPIName ||
		!service.DeletionTimestamp.IsZero() || service.Spec.Type != corev1.ServiceTypeClusterIP ||
		len(service.Spec.Selector) != 0 {
		return nil
	}

	type destination struct {
		cidr string
		port int32
	}
	destinations := make(map[destination]struct{})
	add := func(address string, port int32) {
		ip, err := netip.ParseAddr(address)
		if err != nil || ip.Zone() != "" || !ip.IsGlobalUnicast() || port < 1 || port > 65535 {
			return
		}
		ip = ip.Unmap()
		destinations[destination{netip.PrefixFrom(ip, ip.BitLen()).String(), port}] = struct{}{}
	}
	addresses := service.Spec.ClusterIPs
	if len(addresses) == 0 {
		addresses = []string{service.Spec.ClusterIP}
	}
	for _, port := range service.Spec.Ports {
		if port.Name == "https" && port.Protocol == corev1.ProtocolTCP {
			for _, address := range addresses {
				add(address, port.Port)
			}
		}
	}
	if len(destinations) == 0 {
		return nil
	}
	if endpoints != nil && endpoints.Name == service.Name && endpoints.Namespace == service.Namespace &&
		endpoints.DeletionTimestamp.IsZero() {
		for _, subset := range endpoints.Subsets {
			for _, port := range subset.Ports {
				if port.Name == "https" && port.Protocol == corev1.ProtocolTCP {
					for _, address := range subset.Addresses {
						add(address.IP, port.Port)
					}
				}
			}
		}
	}

	ordered := make([]destination, 0, len(destinations))
	for target := range destinations {
		ordered = append(ordered, target)
	}
	slices.SortFunc(ordered, func(a, b destination) int {
		if byAddress := cmp.Compare(a.cidr, b.cidr); byAddress != 0 {
			return byAddress
		}
		return cmp.Compare(a.port, b.port)
	})
	rules := make([]netv1.NetworkPolicyEgressRule, 0, len(ordered))
	for _, target := range ordered {
		rules = append(rules, netv1.NetworkPolicyEgressRule{
			To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: target.cidr}}},
			Ports: []netv1.NetworkPolicyPort{tcpPort(target.port)},
		})
	}
	return rules
}
