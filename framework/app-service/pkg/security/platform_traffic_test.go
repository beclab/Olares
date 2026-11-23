package security

import (
	"reflect"
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPlatformTrafficUsesPairedServiceAndTargetPorts(t *testing.T) {
	key := types.NamespacedName{Namespace: "os-platform", Name: "redis-cluster-proxy"}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.23",
			Selector: map[string]string{"app": "kvrocks", "app.kubernetes.io/name": "kvrocks"},
			Ports:    []corev1.ServicePort{{Port: 6379, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(6666)}},
		},
	}
	services := map[types.NamespacedName]*corev1.Service{key: svc}
	lookup := func(namespace, name string) netv1.NetworkPolicy {
		t.Helper()
		for _, policy := range BuildPlatformTrafficPolicies(namespace, services) {
			if policy.Name == name {
				return policy
			}
		}
		t.Fatalf("missing %s/%s", namespace, name)
		return netv1.NetworkPolicy{}
	}
	out := lookup("os-framework", "olares-platform-infisical-kvrocks-egress")
	in := lookup("os-platform", "olares-platform-infisical-kvrocks-ingress")
	if len(out.Spec.Egress) != 2 || !reflect.DeepEqual(out.Spec.PodSelector.MatchLabels, map[string]string{"app": "infisical"}) ||
		!reflect.DeepEqual(out.Spec.Egress[0].To, []netv1.NetworkPolicyPeer{platformPeer("os-platform", svc.Spec.Selector)}) ||
		!reflect.DeepEqual(out.Spec.Egress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(6666)}) ||
		!reflect.DeepEqual(out.Spec.Egress[1].To, []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.96.0.23/32"}}}) ||
		!reflect.DeepEqual(out.Spec.Egress[1].Ports, []netv1.NetworkPolicyPort{tcpPort(6379)}) {
		t.Fatalf("unexpected Redis egress: %+v", out.Spec)
	}
	if len(in.Spec.Ingress) != 1 || !reflect.DeepEqual(in.Spec.PodSelector.MatchLabels, svc.Spec.Selector) ||
		!reflect.DeepEqual(in.Spec.Ingress[0].From, []netv1.NetworkPolicyPeer{platformPeer("os-framework", map[string]string{"app": "infisical"})}) ||
		!reflect.DeepEqual(in.Spec.Ingress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(6666)}) {
		t.Fatalf("unexpected Redis ingress: %+v", in.Spec)
	}
	if !IsTrafficGroupManagedNP(&out) || !IsTrafficGroupManagedNP(&in) {
		t.Fatal("legacy sweep would prune platform policies")
	}
	if got := PlatformTrafficNamespaces(key); !reflect.DeepEqual(got, []string{"os-framework", "os-platform"}) {
		t.Fatalf("service changes enqueue %v", got)
	}
	wrongPort := svc.DeepCopy()
	wrongPort.Spec.Ports[0].TargetPort = intstr.FromInt(17170)
	wrongSelector := svc.DeepCopy()
	wrongSelector.Spec.Selector = map[string]string{"app": "lldap"}
	external := svc.DeepCopy()
	external.Spec.Type = corev1.ServiceTypeExternalName
	for _, changed := range []*corev1.Service{nil, wrongPort, wrongSelector, external} {
		services[key] = changed
		if len(lookup("os-framework", out.Name).Spec.Egress) != 0 || len(lookup("os-platform", in.Name).Spec.Ingress) != 0 {
			t.Fatal("missing or changed Service generated a grant")
		}
	}
}

func TestPlatformTrafficInventoryOnlyAddsVerifiedPorts(t *testing.T) {
	for _, flow := range platformServiceFlows {
		if slices.Contains([]int32{8080, 17170, 17171, 18001, 18002, 8222}, flow.servicePort) ||
			slices.Contains([]int32{8080, 17170, 17171, 18001, 18002, 8222}, flow.targetPort) {
			t.Fatalf("%s grants an unapproved port", flow.name)
		}
	}
	for _, policy := range BuildPlatformTrafficPolicies("user-space-test", nil) {
		if len(policy.Spec.Ingress) != 0 || len(policy.Spec.Egress) != 0 {
			t.Fatalf("tenant received a grant without a matching Service: %+v", policy)
		}
	}
	ingress := NPAllowFromL4("os-gateway").Spec
	if !reflect.DeepEqual(ingress.PodSelector.MatchLabels, gatewayDataPlanePods("app-gateway")) ||
		!reflect.DeepEqual(ingress.Ingress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(8443), tcpPort(9444)}) {
		t.Fatalf("unexpected L4 listener scope: %+v", ingress)
	}
}

func TestL4HostIngressTracksOnlySelectedNodeAddresses(t *testing.T) {
	l4 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-network", Labels: map[string]string{"app": "l4-bfl-proxy", "tier": "bfl"}},
		Spec:       corev1.PodSpec{HostNetwork: true, NodeName: "edge", ServiceAccountName: "os-network-internal"},
		Status:     corev1.PodStatus{HostIP: "192.168.1.2", HostIPs: []corev1.HostIP{{IP: "fd00::2"}}},
	}
	unowned := l4.DeepCopy()
	unowned.Spec.ServiceAccountName = "default"
	unowned.Status.HostIP = "192.168.1.3"
	deleting := l4.DeepCopy()
	now := metav1.Now()
	deleting.DeletionTimestamp = &now
	deleting.Status.HostIP = "192.168.1.4"
	policy := L4HostNetworkIngressPolicy([]corev1.Pod{l4, *unowned, *deleting})
	if !IsTrafficGroupManagedNP(&policy) || len(policy.Spec.Ingress) != 1 {
		t.Fatalf("missing managed L4 host ingress: %+v", policy)
	}
	want := []netv1.NetworkPolicyPeer{
		{IPBlock: &netv1.IPBlock{CIDR: "192.168.1.2/32"}},
		{IPBlock: &netv1.IPBlock{CIDR: "fd00::2/128"}},
	}
	if rule := policy.Spec.Ingress[0]; !reflect.DeepEqual(rule.From, want) ||
		!reflect.DeepEqual(rule.Ports, []netv1.NetworkPolicyPort{tcpPort(8443), tcpPort(9444)}) {
		t.Fatalf("unexpected L4 host scope: %+v", rule)
	}
	if got := L4HostNetworkIngressPolicy([]corev1.Pod{*deleting}); len(got.Spec.Ingress) != 0 {
		t.Fatalf("old host remained allowed: %+v", got.Spec)
	}
}

func TestLLDAPNATSPreservesOnlyTheNativeServicePort(t *testing.T) {
	key := types.NamespacedName{Namespace: "os-platform", Name: "nats"}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, Selector: natsPods,
			Ports: []corev1.ServicePort{
				{Name: "nats", Port: 4222, TargetPort: intstr.FromString("nats"), Protocol: corev1.ProtocolTCP},
				{Name: "http-monitor", Port: 8222, TargetPort: intstr.FromInt(8222), Protocol: corev1.ProtocolTCP},
			},
		},
	}
	var inbound, outbound bool
	for _, policy := range BuildPlatformTrafficPolicies("os-platform", map[types.NamespacedName]*corev1.Service{key: service}) {
		switch policy.Name {
		case "olares-platform-lldap-nats-ingress":
			inbound = len(policy.Spec.Ingress) == 1 &&
				reflect.DeepEqual(policy.Spec.Ingress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(4222)}) &&
				reflect.DeepEqual(policy.Spec.Ingress[0].From, []netv1.NetworkPolicyPeer{platformPeer("os-platform", map[string]string{"app": "lldap"})})
		case "olares-platform-lldap-nats-egress":
			outbound = len(policy.Spec.Egress) == 1 &&
				reflect.DeepEqual(policy.Spec.PodSelector.MatchLabels, map[string]string{"app": "lldap"}) &&
				reflect.DeepEqual(policy.Spec.Egress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(4222)})
		}
	}
	if !inbound || !outbound {
		t.Fatalf("missing scoped LLDAP NATS rules: ingress=%v egress=%v", inbound, outbound)
	}
}

func TestAutheliaFrontendTransportStaysWithinTenant(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "user-system-alice", Name: "authelia-backend"},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.12",
			Selector: map[string]string{"app": "systemserver"},
			Ports:    []corev1.ServicePort{{Port: 9091, TargetPort: intstr.FromInt(9091), Protocol: corev1.ProtocolTCP}},
		},
	}
	key := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	services := map[types.NamespacedName]*corev1.Service{key: service}
	if got := PlatformTrafficNamespaces(key); !reflect.DeepEqual(got, []string{"user-space-alice", "user-system-alice"}) {
		t.Fatalf("Service change must enqueue both tenant namespaces: %v", got)
	}
	egress := BuildPlatformTrafficPolicies("user-space-alice", services)[0]
	ingress := BuildPlatformTrafficPolicies("user-system-alice", services)[0]
	if len(egress.Spec.Egress) != 2 || len(ingress.Spec.Ingress) != 1 || !IsTrafficGroupManagedNP(&egress) || !IsTrafficGroupManagedNP(&ingress) {
		t.Fatalf("expected two persistent transport policies: %+v %+v", egress, ingress)
	}
	backend := map[string]string{"app": "systemserver", platformMeshControlPlaneLabel: "os-mesh"}
	frontend := map[string]string{"app": "authelia", platformMeshControlPlaneLabel: "os-mesh"}
	ports := []netv1.NetworkPolicyPort{tcpPort(9091), tcpPort(4143)}
	if !reflect.DeepEqual(ingress.Spec.PodSelector.MatchLabels, backend) ||
		!reflect.DeepEqual(egress.Spec.PodSelector.MatchLabels, frontend) ||
		!reflect.DeepEqual(ingress.Spec.Ingress[0].From, []netv1.NetworkPolicyPeer{platformPeer("user-space-alice", frontend)}) ||
		!reflect.DeepEqual(egress.Spec.Egress[0].To, []netv1.NetworkPolicyPeer{platformPeer("user-system-alice", backend)}) ||
		!reflect.DeepEqual(ingress.Spec.Ingress[0].Ports, ports) ||
		!reflect.DeepEqual(egress.Spec.Egress[0].Ports, ports) ||
		!reflect.DeepEqual(egress.Spec.Egress[1].Ports, []netv1.NetworkPolicyPort{tcpPort(9091)}) {
		t.Fatalf("transport selected an unexpected workload or port: %+v %+v", egress.Spec, ingress.Spec)
	}
	if got := BuildPlatformTrafficPolicies("user-space-bob", services)[0]; len(got.Spec.Egress) != 0 {
		t.Fatalf("another tenant's Service activated egress: %+v", got.Spec)
	}
	service.Spec.Ports[0].TargetPort = intstr.FromInt(8080)
	if got := BuildPlatformTrafficPolicies("user-system-alice", services)[0]; len(got.Spec.Ingress) != 0 {
		t.Fatalf("retargeting the Service must withdraw ingress: %+v", got.Spec)
	}
}
