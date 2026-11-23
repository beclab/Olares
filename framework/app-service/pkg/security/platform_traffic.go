package security

import (
	"maps"
	"net/netip"
	"slices"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type platformWorkload struct {
	namespace string
	selector  map[string]string
}

type platformServiceFlow struct {
	name        string
	from        platformWorkload
	service     types.NamespacedName
	to          map[string]string
	servicePort int32
	targetPort  int32
	targetName  string
}

var (
	natsPods            = map[string]string{"app.kubernetes.io/name": "nats", "app.kubernetes.io/instance": "nats", "app.kubernetes.io/component": "nats"}
	citusPods           = map[string]string{"app": "citus", "app.kubernetes.io/name": "citus"}
	kvrocksPods         = map[string]string{"app": "kvrocks", "app.kubernetes.io/name": "kvrocks"}
	envoyControllerPods = map[string]string{"control-plane": "envoy-gateway", "app.kubernetes.io/name": "gateway-helm", "app.kubernetes.io/instance": "app-gateway-eg"}
)

func gatewayDataPlanePods(gateway string) map[string]string {
	return map[string]string{
		"gateway.envoyproxy.io/owning-gateway-name":      gateway,
		"gateway.envoyproxy.io/owning-gateway-namespace": "os-gateway",
		"app.kubernetes.io/component":                    "proxy",
		"app.kubernetes.io/managed-by":                   "envoy-gateway",
		"app.kubernetes.io/name":                         "envoy",
	}
}

func workload(namespace, key, value string) platformWorkload {
	return platformWorkload{namespace, map[string]string{key: value}}
}

// This inventory describes existing platform call sites. It does not infer
// permissions from an open port or from labels supplied by tenant workloads.
var platformServiceFlows = []platformServiceFlow{
	{"app-service-nats", workload("os-framework", "tier", "app-service"), types.NamespacedName{Namespace: "os-platform", Name: "nats"}, natsPods, 4222, 4222, "nats"},
	{"backup-nats", workload("os-framework", "tier", "backup"), types.NamespacedName{Namespace: "os-platform", Name: "nats"}, natsPods, 4222, 4222, "nats"},
	{"authelia-nats", workload("os-framework", "app", "authelia-backend"), types.NamespacedName{Namespace: "os-platform", Name: "nats"}, natsPods, 4222, 4222, "nats"},
	{"lldap-citus", workload("os-platform", "app", "lldap"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"authelia-citus", workload("os-framework", "app", "authelia-backend"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"infisical-citus", workload("os-framework", "app", "infisical"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"infisical-kvrocks", workload("os-framework", "app", "infisical"), types.NamespacedName{Namespace: "os-platform", Name: "redis-cluster-proxy"}, kvrocksPods, 6379, 6666, ""},
	{"files-kvrocks", workload("os-framework", "app", "files"), types.NamespacedName{Namespace: "os-platform", Name: "redis-cluster-proxy"}, kvrocksPods, 6379, 6666, ""},
	{"authelia-ldap", workload("os-framework", "app", "authelia-backend"), types.NamespacedName{Namespace: "os-platform", Name: "lldap-service"}, map[string]string{"app": "lldap"}, 3890, 3890, ""},
	{"entry-gateway-xds", platformWorkload{"os-gateway", gatewayDataPlanePods("app-gateway")}, types.NamespacedName{Namespace: "os-gateway", Name: "envoy-gateway"}, envoyControllerPods, 18000, 18000, ""},
	{"internal-gateway-xds", platformWorkload{"os-gateway", gatewayDataPlanePods("app-access-internal")}, types.NamespacedName{Namespace: "os-gateway", Name: "envoy-gateway"}, envoyControllerPods, 18000, 18000, ""},
	{"l4-entry", platformWorkload{"os-network", map[string]string{"app": "l4-bfl-proxy", "tier": "bfl"}}, types.NamespacedName{Namespace: "os-gateway", Name: "app-gateway-l4-mtls"}, gatewayDataPlanePods("app-gateway"), 8443, 8443, ""},
	{"l4-authz", platformWorkload{"os-network", map[string]string{"app": "l4-bfl-proxy", "tier": "bfl"}}, types.NamespacedName{Namespace: "os-gateway", Name: "app-gateway-l4-mtls"}, gatewayDataPlanePods("app-gateway"), 9444, 9444, ""},
}

func PlatformTrafficServices(namespace string) []types.NamespacedName {
	var keys []types.NamespacedName
	for _, flow := range platformServiceFlows {
		if (flow.from.namespace == namespace || flow.service.Namespace == namespace) && !slices.Contains(keys, flow.service) {
			keys = append(keys, flow.service)
		}
	}
	return keys
}

func PlatformTrafficNamespaces(service types.NamespacedName) []string {
	var namespaces []string
	for _, flow := range platformServiceFlows {
		if flow.service != service {
			continue
		}
		for _, namespace := range []string{flow.from.namespace, flow.service.Namespace} {
			if !slices.Contains(namespaces, namespace) {
				namespaces = append(namespaces, namespace)
			}
		}
	}
	return namespaces
}

func BuildPlatformTrafficPolicies(namespace string, services map[types.NamespacedName]*corev1.Service) []netv1.NetworkPolicy {
	var policies []netv1.NetworkPolicy
	for _, flow := range platformServiceFlows {
		service := services[flow.service]
		usable := flow.matchesService(service)
		if namespace == flow.from.namespace {
			policy := netv1.NetworkPolicy{
				ObjectMeta: npMeta("olares-platform-"+flow.name+"-egress", namespace),
				Spec: netv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{MatchLabels: flow.from.selector},
					PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
				},
			}
			if usable {
				policy.Spec.Egress = []netv1.NetworkPolicyEgressRule{{
					To:    []netv1.NetworkPolicyPeer{platformPeer(flow.service.Namespace, service.Spec.Selector)},
					Ports: []netv1.NetworkPolicyPort{tcpPort(flow.targetPort)},
				}}
				for _, addr := range serviceClusterIPs(service) {
					if cidr := exactIPCIDR(addr); cidr != "" {
						policy.Spec.Egress = append(policy.Spec.Egress, netv1.NetworkPolicyEgressRule{
							To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: cidr}}},
							Ports: []netv1.NetworkPolicyPort{tcpPort(flow.servicePort)},
						})
					}
				}
			}
			policies = append(policies, policy)
		}
		if namespace == flow.service.Namespace {
			policy := netv1.NetworkPolicy{
				ObjectMeta: npMeta("olares-platform-"+flow.name+"-ingress", namespace),
				Spec: netv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{MatchLabels: flow.to},
					PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
				},
			}
			if usable {
				policy.Spec.PodSelector.MatchLabels = service.Spec.Selector
				policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
					From:  []netv1.NetworkPolicyPeer{platformPeer(flow.from.namespace, flow.from.selector)},
					Ports: []netv1.NetworkPolicyPort{tcpPort(flow.targetPort)},
				}}
			}
			policies = append(policies, policy)
		}
	}
	return policies
}

func (flow platformServiceFlow) matchesService(service *corev1.Service) bool {
	if service == nil || service.Name != flow.service.Name || service.Namespace != flow.service.Namespace ||
		!service.DeletionTimestamp.IsZero() || service.Spec.Type != corev1.ServiceTypeClusterIP ||
		!maps.Equal(service.Spec.Selector, flow.to) {
		return false
	}
	for _, port := range service.Spec.Ports {
		if port.Port != flow.servicePort || port.Protocol != corev1.ProtocolTCP {
			continue
		}
		switch port.TargetPort.Type {
		case intstr.Int:
			if port.TargetPort.IntVal == flow.targetPort ||
				(port.TargetPort.IntVal == 0 && flow.targetPort == flow.servicePort) {
				return true
			}
		case intstr.String:
			if flow.targetName != "" && port.TargetPort.StrVal == flow.targetName {
				return true
			}
		}
	}
	return false
}

func platformPeer(namespace string, selector map[string]string) netv1.NetworkPolicyPeer {
	return netv1.NetworkPolicyPeer{
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/metadata.name": namespace}},
		PodSelector:       &metav1.LabelSelector{MatchLabels: selector},
	}
}

func serviceClusterIPs(service *corev1.Service) []string {
	if len(service.Spec.ClusterIPs) > 0 {
		return service.Spec.ClusterIPs
	}
	return []string{service.Spec.ClusterIP}
}

func exactIPCIDR(address string) string {
	ip, err := netip.ParseAddr(address)
	if err != nil || ip.Zone() != "" || !ip.IsGlobalUnicast() {
		return ""
	}
	ip = ip.Unmap()
	return netip.PrefixFrom(ip, ip.BitLen()).String()
}

const l4HostIngressName = "olares-platform-l4-host-ingress"

func IsL4HostNetworkPod(pod *corev1.Pod) bool {
	return pod != nil && pod.Namespace == "os-network" && pod.Spec.HostNetwork &&
		pod.Spec.ServiceAccountName == "os-network-internal" &&
		pod.Labels["app"] == "l4-bfl-proxy" && pod.Labels["tier"] == "bfl"
}

// A hostNetwork Pod can present its node IP to the CNI. On this path the
// listener's Entry client certificate is the workload authentication boundary.
func L4HostNetworkIngressPolicy(pods []corev1.Pod) netv1.NetworkPolicy {
	policy := netv1.NetworkPolicy{
		ObjectMeta: npMeta(l4HostIngressName, "os-gateway"),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: gatewayDataPlanePods("app-gateway")},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
		},
	}
	cidrs := make(map[string]struct{})
	for _, pod := range pods {
		if !IsL4HostNetworkPod(&pod) || !pod.DeletionTimestamp.IsZero() || pod.Spec.NodeName == "" {
			continue
		}
		for _, address := range append(slices.Clone(pod.Status.HostIPs), corev1.HostIP{IP: pod.Status.HostIP}) {
			if cidr := exactIPCIDR(address.IP); cidr != "" {
				cidrs[cidr] = struct{}{}
			}
		}
	}
	var peers []netv1.NetworkPolicyPeer
	for _, cidr := range slices.Sorted(maps.Keys(cidrs)) {
		peers = append(peers, netv1.NetworkPolicyPeer{IPBlock: &netv1.IPBlock{CIDR: cidr}})
	}
	if len(peers) > 0 {
		policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
			From: peers, Ports: []netv1.NetworkPolicyPort{tcpPort(8443), tcpPort(9444)},
		}}
	}
	return policy
}
