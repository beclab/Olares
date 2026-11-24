package security

import (
	"maps"
	"net/netip"
	"slices"
	"strings"

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
	platformAuthelia    = platformWorkload{"os-framework", map[string]string{"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true"}}
	platformVault       = platformWorkload{"os-framework", map[string]string{"app": "vault-server", platformMeshControlPlaneLabel: "os-mesh"}}
	bflPods             = map[string]string{"tier": "bfl"}
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
	{"lldap-nats", workload("os-platform", "app", "lldap"), types.NamespacedName{Namespace: "os-platform", Name: "nats"}, natsPods, 4222, 4222, "nats"},
	{"lldap-citus", workload("os-platform", "app", "lldap"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"authelia-citus", workload("os-framework", "app", "authelia-backend"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"infisical-citus", workload("os-framework", "app", "infisical"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"infisical-kvrocks", workload("os-framework", "app", "infisical"), types.NamespacedName{Namespace: "os-platform", Name: "redis-cluster-proxy"}, kvrocksPods, 6379, 6666, ""},
	{"files-kvrocks", workload("os-framework", "app", "files"), types.NamespacedName{Namespace: "os-platform", Name: "redis-cluster-proxy"}, kvrocksPods, 6379, 6666, ""},
	{"vault-citus", workload("os-framework", "app", "vault-server"), types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"}, citusPods, 5432, 5432, ""},
	{"vault-nats", workload("os-framework", "app", "vault-server"), types.NamespacedName{Namespace: "os-platform", Name: "nats"}, natsPods, 4222, 4222, "nats"},
	{"vault-authelia", platformVault, types.NamespacedName{Namespace: "os-framework", Name: "authelia-backend"}, map[string]string{"app": "authelia-backend"}, 9091, 9091, ""},
	{"authelia-ldap", workload("os-framework", "app", "authelia-backend"), types.NamespacedName{Namespace: "os-platform", Name: "lldap-service"}, map[string]string{"app": "lldap"}, 3890, 3890, ""},
	// Token cache initialization still uses the LLDAP HTTP API during mesh migration.
	{"authelia-lldap-token", platformAuthelia, types.NamespacedName{Namespace: "os-platform", Name: "lldap-service"}, map[string]string{"app": "lldap"}, 17170, 17170, ""},
	{"entry-gateway-xds", platformWorkload{"os-gateway", gatewayDataPlanePods("app-gateway")}, types.NamespacedName{Namespace: "os-gateway", Name: "envoy-gateway"}, envoyControllerPods, 18000, 18000, ""},
	{"internal-gateway-xds", platformWorkload{"os-gateway", gatewayDataPlanePods("app-access-internal")}, types.NamespacedName{Namespace: "os-gateway", Name: "envoy-gateway"}, envoyControllerPods, 18000, 18000, ""},
	{"l4-entry", platformWorkload{"os-network", map[string]string{"app": "l4-bfl-proxy", "tier": "bfl"}}, types.NamespacedName{Namespace: "os-gateway", Name: "app-gateway-l4-mtls"}, gatewayDataPlanePods("app-gateway"), 8443, 8443, ""},
	{"l4-authz", platformWorkload{"os-network", map[string]string{"app": "l4-bfl-proxy", "tier": "bfl"}}, types.NamespacedName{Namespace: "os-gateway", Name: "app-gateway-l4-mtls"}, gatewayDataPlanePods("app-gateway"), 9444, 9444, ""},
}

const (
	platformMeshControlPlaneLabel = "linkerd.io/control-plane-ns"
	autheliaBFLFlowName           = "authelia-bfl"
)

func autheliaBFLFlow(space string) platformServiceFlow {
	return platformServiceFlow{autheliaBFLFlowName, platformAuthelia, types.NamespacedName{Namespace: space, Name: "bfl"}, bflPods, 80, 8080, ""}
}

func autheliaBFLMeshFlow(space string) platformServiceFlow {
	flow := autheliaBFLFlow(space)
	flow.name += "-mesh"
	flow.from.selector = maps.Clone(flow.from.selector)
	flow.from.selector[platformMeshControlPlaneLabel] = "os-mesh"
	return flow
}

func vaultBFLFlow(space string) platformServiceFlow {
	return platformServiceFlow{"vault-bfl", platformVault, types.NamespacedName{Namespace: space, Name: "bfl"}, bflPods, 80, 8080, ""}
}

func platformTrafficFlows(namespace string) []platformServiceFlow {
	var tenant string
	for _, prefix := range []string{"user-space-", "user-system-"} {
		if suffix, ok := strings.CutPrefix(namespace, prefix); ok && suffix != "" {
			tenant = suffix
			break
		}
	}
	if tenant == "" {
		return platformServiceFlows
	}
	space, system := "user-space-"+tenant, "user-system-"+tenant
	frontend := func(app string) platformWorkload {
		return platformWorkload{space, map[string]string{"app": app, platformMeshControlPlaneLabel: "os-mesh"}}
	}
	auth := types.NamespacedName{Namespace: system, Name: "authelia-backend"}
	bfl := types.NamespacedName{Namespace: space, Name: "bfl"}
	systemProxy := map[string]string{"app": "systemserver"}
	return []platformServiceFlow{
		{"authelia-frontend-auth", frontend("authelia"), auth, systemProxy, 9091, 9091, "auth-proxy"},
		{"desktop-auth", frontend("olares-app"), auth, systemProxy, 9091, 9091, "auth-proxy"},
		{"wizard-auth", frontend("wizard"), auth, systemProxy, 9091, 9091, "auth-proxy"},
		{"desktop-logout", frontend("olares-app"), types.NamespacedName{Namespace: space, Name: "authelia-svc"}, map[string]string{"app": "authelia"}, 80, 80, ""},
		{"login-bfl", frontend("authelia"), bfl, bflPods, 80, 8080, ""},
		{"desktop-bfl", frontend("olares-app"), bfl, bflPods, 80, 8080, ""},
		{"wizard-bfl", frontend("wizard"), bfl, bflPods, 80, 8080, ""},
		{"tapr-images-bfl", platformWorkload{system, map[string]string{
			"app": "tapr-images", platformMeshControlPlaneLabel: "os-mesh",
		}}, bfl, bflPods, 80, 8080, ""},
		{"bfl-app-service", platformWorkload{space, map[string]string{
			"tier": "bfl", platformMeshControlPlaneLabel: "os-mesh",
		}}, types.NamespacedName{Namespace: system, Name: "app-service"}, systemProxy, 28080, 28080, ""},
		{"desktop-app-service", frontend("olares-app"), types.NamespacedName{Namespace: system, Name: "app-service"}, systemProxy, 28080, 28080, ""},
		{"desktop-nats", frontend("olares-app"), types.NamespacedName{Namespace: system, Name: "nats"}, systemProxy, 4222, 4222, ""},
		{"bfl-cert-refresh", platformWorkload{space, map[string]string{
			"app.kubernetes.io/name": "bfl-cert-refresh", platformMeshControlPlaneLabel: "os-mesh",
		}}, bfl, bflPods, 80, 8080, ""},
		autheliaBFLFlow(space),
		autheliaBFLMeshFlow(space),
		vaultBFLFlow(space),
	}
}

func (flow platformServiceFlow) destination() (map[string]string, []netv1.NetworkPolicyPort) {
	ports := []netv1.NetworkPolicyPort{tcpPort(flow.targetPort)}
	if flow.from.selector[platformMeshControlPlaneLabel] == "" {
		return flow.to, ports
	}
	selector := maps.Clone(flow.to)
	selector[platformMeshControlPlaneLabel] = flow.from.selector[platformMeshControlPlaneLabel]
	return selector, append(ports, tcpPort(4143))
}

func PlatformTrafficServices(namespace string) []types.NamespacedName {
	var keys []types.NamespacedName
	for _, flow := range platformTrafficFlows(namespace) {
		if (flow.from.namespace == namespace || flow.service.Namespace == namespace) && !slices.Contains(keys, flow.service) {
			keys = append(keys, flow.service)
		}
	}
	return keys
}

// ActivePlatformTrafficServices lists the known upstream dependencies of the
// running workload snapshot being evaluated for first egress isolation.
func ActivePlatformTrafficServices(namespace string, pods []corev1.Pod) []types.NamespacedName {
	var keys []types.NamespacedName
	for _, flow := range platformTrafficFlows(namespace) {
		if flow.from.namespace == namespace && !slices.Contains(keys, flow.service) &&
			slices.ContainsFunc(pods, func(pod corev1.Pod) bool {
				for k, v := range flow.from.selector {
					if pod.Labels[k] != v {
						return false
					}
				}
				return true
			}) {
			keys = append(keys, flow.service)
		}
	}
	if IsUserSystemNamespaces(namespace) && slices.ContainsFunc(pods, func(pod corev1.Pod) bool {
		return pod.Labels["app"] == "systemserver" && pod.Labels["traffic.olares.io/platform-mesh"] == "true"
	}) {
		keys = append(keys, types.NamespacedName{Namespace: "os-platform", Name: "citus-headless"})
	}
	return keys
}

func PlatformTrafficNamespaces(service types.NamespacedName) []string {
	var namespaces []string
	for _, flow := range platformTrafficFlows(service.Namespace) {
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
	if namespace == platformAuthelia.namespace {
		// Platform callers use BFL in multiple tenants. Rebuild the source
		// policies from the Service inventory so stale grants are withdrawn.
		var spaces []string
		for key := range services {
			if key.Name == "bfl" && IsUserSpaceNamespaces(key.Namespace) {
				spaces = append(spaces, key.Namespace)
			}
		}
		slices.Sort(spaces)
		for _, buildFlow := range []func(string) platformServiceFlow{autheliaBFLFlow, autheliaBFLMeshFlow, vaultBFLFlow} {
			out := buildFlow("").egressPolicy(nil)
			for _, space := range spaces {
				flow := buildFlow(space)
				out.Spec.Egress = append(out.Spec.Egress, flow.egressRules(services[flow.service])...)
			}
			policies = append(policies, out)
		}
	}
	for _, flow := range platformTrafficFlows(namespace) {
		service := services[flow.service]
		if namespace == flow.from.namespace {
			policies = append(policies, flow.egressPolicy(service))
		}
		if namespace == flow.service.Namespace {
			destination, ports := flow.destination()
			policy := netv1.NetworkPolicy{
				ObjectMeta: npMeta("olares-platform-"+flow.name+"-ingress", namespace),
				Spec: netv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{MatchLabels: destination},
					PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
				},
			}
			if flow.matchesService(service) {
				policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
					From:  []netv1.NetworkPolicyPeer{platformPeer(flow.from.namespace, flow.from.selector)},
					Ports: ports,
				}}
			}
			policies = append(policies, policy)
		}
		if flow.name == "desktop-bfl" && namespace == flow.from.namespace {
			policies = append(policies, unmeshedDesktopBFLPolicies(flow, service)...)
		}
	}
	if tenant, ok := strings.CutPrefix(namespace, "user-system-"); ok && tenant != "" {
		policies = append(policies, systemServerCitusEgress(namespace))
	}
	return policies
}

// Desktop still uses the BFL API on installations where BFL has not enrolled
// in Linkerd. Only allow the verified API target port on those BFL Pods.
func unmeshedDesktopBFLPolicies(flow platformServiceFlow, service *corev1.Service) []netv1.NetworkPolicy {
	name, namespace := "olares-platform-desktop-bfl-unmeshed", flow.from.namespace
	source := metav1.LabelSelector{MatchLabels: flow.from.selector}
	target := metav1.LabelSelector{
		MatchLabels: flow.to,
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key: platformMeshControlPlaneLabel, Operator: metav1.LabelSelectorOpDoesNotExist,
		}},
	}
	egress := netv1.NetworkPolicy{
		ObjectMeta: npMeta(name+"-egress", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: source,
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
		},
	}
	ingress := netv1.NetworkPolicy{
		ObjectMeta: npMeta(name+"-ingress", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: target,
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
		},
	}
	if flow.matchesService(service) {
		targetPeer := platformPeer(namespace, flow.to)
		targetPeer.PodSelector = &target
		ports := []netv1.NetworkPolicyPort{tcpPort(flow.targetPort)}
		egress.Spec.Egress = []netv1.NetworkPolicyEgressRule{{To: []netv1.NetworkPolicyPeer{targetPeer}, Ports: ports}}
		ingress.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
			From: []netv1.NetworkPolicyPeer{platformPeer(namespace, flow.from.selector)}, Ports: ports,
		}}
	}
	return []netv1.NetworkPolicy{egress, ingress}
}

// SystemServerCitusEgressRule permits native PostgreSQL traffic to the coordinator.
func SystemServerCitusEgressRule() netv1.NetworkPolicyEgressRule {
	// The system-server Envoy connects to the coordinator's headless DNS
	// record, so select the coordinator Pod without using a Service VIP.
	coordinator := maps.Clone(citusPods)
	coordinator["statefulset.kubernetes.io/pod-name"] = "citus-0"
	return netv1.NetworkPolicyEgressRule{
		To:    []netv1.NetworkPolicyPeer{platformPeer("os-platform", coordinator)},
		Ports: []netv1.NetworkPolicyPort{tcpPort(5432)},
	}
}

func systemServerCitusEgress(namespace string) netv1.NetworkPolicy {
	systemServer := map[string]string{"app": "systemserver", "traffic.olares.io/platform-mesh": "true"}
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-platform-systemserver-citus-egress", namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: systemServer},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress:      []netv1.NetworkPolicyEgressRule{SystemServerCitusEgressRule()},
		},
	}
}

func (flow platformServiceFlow) egressPolicy(service *corev1.Service) netv1.NetworkPolicy {
	return netv1.NetworkPolicy{
		ObjectMeta: npMeta("olares-platform-"+flow.name+"-egress", flow.from.namespace),
		Spec: netv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{MatchLabels: flow.from.selector},
			PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
			Egress:      flow.egressRules(service),
		},
	}
}

func (flow platformServiceFlow) egressRules(service *corev1.Service) []netv1.NetworkPolicyEgressRule {
	if !flow.matchesService(service) {
		return nil
	}
	destination, ports := flow.destination()
	rules := []netv1.NetworkPolicyEgressRule{{
		To:    []netv1.NetworkPolicyPeer{platformPeer(flow.service.Namespace, destination)},
		Ports: ports,
	}}
	for _, addr := range serviceClusterIPs(service) {
		if cidr := exactIPCIDR(addr); cidr != "" {
			rules = append(rules, netv1.NetworkPolicyEgressRule{
				To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: cidr}}},
				Ports: []netv1.NetworkPolicyPort{tcpPort(flow.servicePort)},
			})
		}
	}
	return rules
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
