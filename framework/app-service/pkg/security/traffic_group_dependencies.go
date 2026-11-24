package security

import (
	"cmp"
	"maps"
	"slices"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	KubernetesAPINamespace  = "default"
	KubernetesAPIName       = "kubernetes"
	CoreDNSServiceName      = "coredns"
	coreDNSIngressName      = "olares-traffic-group-coredns-ingress"
	kubernetesAPIEgressName = "olares-traffic-group-kube-api-egress"
)

var platformAPIClients = []struct {
	name string
	pods platformWorkload
}{
	{"app-service", workload("os-framework", "tier", "app-service")},
	{"backup", workload("os-framework", "tier", "backup")},
	{"envoy-controller", platformWorkload{"os-gateway", envoyControllerPods}},
	{"linkerd-identity", workload("os-mesh", "linkerd.io/control-plane-component", "identity")},
	{"linkerd-destination", workload("os-mesh", "linkerd.io/control-plane-component", "destination")},
	{"linkerd-injector", workload("os-mesh", "linkerd.io/control-plane-component", "proxy-injector")},
	{"linkerd-pki-guardian", workload("os-mesh", "app.kubernetes.io/component", "linkerd-pki-guardian")},
	{"middleware-operator", workload("os-platform", "app", "tapr-middleware")},
	{"hami-scheduler", workload("kube-system", "app.kubernetes.io/component", "hami-scheduler")},
	{"opa", workload("os-platform", "app", "opa")},
	{"otel", platformWorkload{"os-platform", map[string]string{"app.kubernetes.io/name": "opentelemetry-operator", "app.kubernetes.io/component": "controller-manager"}}},
}

func KubernetesAPIPolicyNamespaces() []string {
	namespaces := []string{metav1.NamespaceSystem, "os-framework"}
	for _, client := range platformAPIClients {
		if !slices.Contains(namespaces, client.pods.namespace) {
			namespaces = append(namespaces, client.pods.namespace)
		}
	}
	return namespaces
}

func UsesKubernetesAPIPolicy(namespace string) bool {
	return IsUserSystemNamespaces(namespace) || IsUserSpaceNamespaces(namespace) ||
		slices.Contains(KubernetesAPIPolicyNamespaces(), namespace)
}

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
		selector = platformAuthelia.selector
	}
	rules := kubernetesAPIEgress(service, endpoints)
	if IsUserSystemNamespaces(namespace) {
		policies = append(policies, apiEgressPolicy(namespace, kubernetesAPIEgressName+"-systemserver",
			map[string]string{"app": "systemserver"}, rules))
	}
	if IsUserSpaceNamespaces(namespace) {
		policies = append(policies, apiEgressPolicy(namespace, kubernetesAPIEgressName+"-bfl",
			map[string]string{"tier": "bfl"}, rules))
	}
	if selector != nil {
		policies = append(policies, apiEgressPolicy(namespace, kubernetesAPIEgressName, selector, rules))
	}
	for _, client := range platformAPIClients {
		if client.pods.namespace == namespace {
			policies = append(policies, apiEgressPolicy(namespace, kubernetesAPIEgressName+"-"+client.name, client.pods.selector, rules))
		}
	}
	return append(policies, platformWebhookPolicies(namespace, service, rules)...)
}

// CoreDNSEgressRule allows only DNS to CoreDNS Pods and its verified Service IPs.
func CoreDNSEgressRule(service *corev1.Service) netv1.NetworkPolicyEgressRule {
	selector := map[string]string{"k8s-app": "kube-dns"}
	rule := netv1.NetworkPolicyEgressRule{
		To:    []netv1.NetworkPolicyPeer{platformPeer(metav1.NamespaceSystem, selector)},
		Ports: []netv1.NetworkPolicyPort{tcpPort(53), udpPort(53)},
	}
	if service == nil || service.Namespace != metav1.NamespaceSystem || service.Name != CoreDNSServiceName ||
		!service.DeletionTimestamp.IsZero() || service.Spec.Type != corev1.ServiceTypeClusterIP ||
		!maps.Equal(service.Spec.Selector, selector) {
		return rule
	}
	var tcp, udp bool
	for _, port := range service.Spec.Ports {
		if port.Port != 53 || (port.TargetPort != intstr.FromInt32(53) && port.TargetPort != intstr.FromInt32(0)) {
			continue
		}
		switch port.Protocol {
		case corev1.ProtocolTCP:
			tcp = true
		case corev1.ProtocolUDP:
			udp = true
		}
	}
	if tcp && udp {
		for _, ip := range serviceClusterIPs(service) {
			if cidr := exactIPCIDR(ip); cidr != "" {
				rule.To = append(rule.To, netv1.NetworkPolicyPeer{IPBlock: &netv1.IPBlock{CIDR: cidr}})
			}
		}
	}
	return rule
}

func apiEgressPolicy(namespace, name string, selector map[string]string, rules []netv1.NetworkPolicyEgressRule) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta(name, namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: selector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress:      rules,
		},
	}
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
		if cidr := exactIPCIDR(address); cidr != "" && port >= 1 && port <= 65535 {
			destinations[destination{cidr, port}] = struct{}{}
		}
	}
	for _, port := range service.Spec.Ports {
		if port.Name == "https" && port.Protocol == corev1.ProtocolTCP {
			for _, address := range serviceClusterIPs(service) {
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

var platformWebhooks = []struct {
	name string
	pods platformWorkload
	port int32
}{
	{"app-service", workload("os-framework", "tier", "app-service"), 8433},
	{"search3-validation", workload("os-framework", "app", "search3-validation"), 8443},
	{"opa", workload("os-platform", "app", "opa"), 8443},
	{"otel", platformWorkload{"os-platform", map[string]string{"app.kubernetes.io/name": "opentelemetry-operator", "app.kubernetes.io/component": "controller-manager"}}, 9443},
	{"envoy-topology", platformWorkload{"os-gateway", envoyControllerPods}, 9443},
	{"hami-scheduler", workload("kube-system", "app.kubernetes.io/component", "hami-scheduler"), 443},
	{"linkerd-injector", workload("os-mesh", "linkerd.io/control-plane-component", "proxy-injector"), 8443},
	{"linkerd-sp-validator", workload("os-mesh", "linkerd.io/control-plane-component", "destination"), 8443},
	{"linkerd-policy-validator", workload("os-mesh", "linkerd.io/control-plane-component", "destination"), 9443},
}

// A ClusterIP is a destination, not an API server source. Only the discovered
// control-plane endpoints are eligible to call these platform webhook ports.
func platformWebhookPolicies(namespace string, service *corev1.Service, apiRules []netv1.NetworkPolicyEgressRule) []netv1.NetworkPolicy {
	var peers []netv1.NetworkPolicyPeer
	if service != nil {
		serviceIPs := make(map[string]struct{})
		for _, ip := range serviceClusterIPs(service) {
			serviceIPs[exactIPCIDR(ip)] = struct{}{}
		}
		for _, rule := range apiRules {
			peer := rule.To[0]
			if _, virtual := serviceIPs[peer.IPBlock.CIDR]; !virtual && !slices.ContainsFunc(peers, func(existing netv1.NetworkPolicyPeer) bool {
				return existing.IPBlock.CIDR == peer.IPBlock.CIDR
			}) {
				peers = append(peers, peer)
			}
		}
	}
	var policies []netv1.NetworkPolicy
	for _, hook := range platformWebhooks {
		if hook.pods.namespace != namespace {
			continue
		}
		policy := netv1.NetworkPolicy{
			ObjectMeta: npMeta("olares-platform-webhook-"+hook.name, namespace),
			Spec: netv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: hook.pods.selector},
				PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
			},
		}
		if len(peers) > 0 {
			policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{From: peers, Ports: []netv1.NetworkPolicyPort{tcpPort(hook.port)}}}
		}
		policies = append(policies, policy)
	}
	return policies
}
