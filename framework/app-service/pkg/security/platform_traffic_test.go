package security

import (
	"maps"
	"reflect"
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGatewayBootstrapScopesDiscoveryAndValidatesControllerVIP(t *testing.T) {
	controller := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "os-gateway", Name: "envoy-gateway"},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.24",
			Selector: maps.Clone(envoyControllerPods),
			Ports:    []corev1.ServicePort{{Port: 18000, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(18000)}}}}
	for _, mode := range []string{"valid", "missing", "foreign-selector", "wrong-port", "deleting"} {
		t.Run(mode, func(t *testing.T) {
			svc := controller.DeepCopy()
			switch mode {
			case "missing":
				svc = nil
			case "foreign-selector":
				svc.Spec.Selector = map[string]string{"app": "other"}
			case "wrong-port":
				svc.Spec.Ports[0].TargetPort = intstr.FromInt(18001)
			case "deleting":
				now := metav1.Now()
				svc.DeletionTimestamp = &now
			}
			rules := GatewayBootstrapEgressRules(nil, svc)
			wantCount := 3
			if mode == "valid" {
				wantCount = 4
			}
			if len(rules) != wantCount {
				t.Fatalf("unexpected bootstrap grants: %+v", rules)
			}
			dns := rules[0]
			if len(dns.To) != 1 || dns.To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "kube-system" ||
				!reflect.DeepEqual(dns.To[0].PodSelector.MatchLabels, map[string]string{"k8s-app": "kube-dns"}) ||
				!reflect.DeepEqual(dns.Ports, []netv1.NetworkPolicyPort{tcpPort(53), udpPort(53)}) {
				t.Fatalf("DNS grant escaped CoreDNS: %+v", dns)
			}
			mesh := rules[1]
			if len(mesh.To) != 1 || mesh.To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "os-mesh" ||
				!reflect.DeepEqual(mesh.Ports, []netv1.NetworkPolicyPort{tcpPort(8080), tcpPort(8086), tcpPort(8090)}) {
				t.Fatalf("unexpected mesh control-plane grant: %+v", mesh)
			}
			xds := rules[2]
			if len(xds.To) != 1 || xds.To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "os-gateway" ||
				!reflect.DeepEqual(xds.To[0].PodSelector.MatchLabels, envoyControllerPods) ||
				!reflect.DeepEqual(xds.Ports, []netv1.NetworkPolicyPort{tcpPort(18000)}) {
				t.Fatalf("xDS grant escaped its native controller: %+v", xds)
			}
			if mode == "valid" && (rules[3].To[0].IPBlock.CIDR != "10.96.0.24/32" ||
				!reflect.DeepEqual(rules[3].Ports, []netv1.NetworkPolicyPort{tcpPort(18000)})) {
				t.Fatalf("controller VIP grant widened: %+v", rules[3])
			}
		})
	}
}

func TestDesktopReadinessTrafficIsScopedToMeshedAppServiceAndEntry(t *testing.T) {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "os-gateway", Name: "app-gateway-l4-mtls"},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.24", Selector: gatewayDataPlanePods("app-gateway"),
			Ports: []corev1.ServicePort{{Port: 8443, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(8443)}, {Port: 9444, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(9444)}}}}
	for _, port := range []int32{8443, 9444} {
		flowName := "entry"
		if port == 9444 {
			flowName = "authz"
		}
		for _, ns := range []string{"os-framework", "os-gateway"} {
			policies := BuildPlatformTrafficPolicies(ns, map[types.NamespacedName]*corev1.Service{{Namespace: svc.Namespace, Name: svc.Name}: svc})
			found := false
			for _, p := range policies {
				if p.Name != "olares-platform-app-service-"+flowName+"-readiness-egress" && p.Name != "olares-platform-app-service-"+flowName+"-readiness-ingress" {
					continue
				}
				found = true
				if !IsTrafficGroupManagedNP(&p) {
					t.Fatal("readiness grant is not protected from the legacy sweep")
				}
				if ns == "os-framework" {
					if !reflect.DeepEqual(p.Spec.PodSelector.MatchLabels, map[string]string{"tier": "app-service", "linkerd.io/control-plane-ns": "os-mesh"}) ||
						len(p.Spec.Egress) != 2 || !reflect.DeepEqual(p.Spec.Egress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(port), tcpPort(4143)}) ||
						p.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "os-gateway" ||
						p.Spec.Egress[0].To[0].PodSelector.MatchLabels["gateway.envoyproxy.io/owning-gateway-name"] != "app-gateway" {
						t.Fatalf("readiness source or destination widened: %+v", p.Spec)
					}
				} else if len(p.Spec.Ingress) != 1 || p.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "os-framework" ||
					!reflect.DeepEqual(p.Spec.Ingress[0].From[0].PodSelector.MatchLabels, map[string]string{"tier": "app-service", "linkerd.io/control-plane-ns": "os-mesh"}) {
					t.Fatalf("readiness ingress trusts another caller: %+v", p.Spec)
				}
			}
			if !found {
				t.Fatal("missing readiness peer policy", ns, port)
			}
		}
	}
}

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

func TestSystemServerCitusEgressScopesCoordinatorTraffic(t *testing.T) {
	lookup := func(namespace, name string) *netv1.NetworkPolicy {
		for _, policy := range BuildPlatformTrafficPolicies(namespace, nil) {
			if policy.Name == name {
				return &policy
			}
		}
		return nil
	}
	out := lookup("user-system-alice", "olares-platform-systemserver-citus-egress")
	if !IsTrafficGroupManagedNP(out) {
		t.Fatal("the legacy sweep must retain the coordinator egress policy")
	}
	source := map[string]string{"app": "systemserver", "traffic.olares.io/platform-mesh": "true"}
	target := map[string]string{"app": "citus", "app.kubernetes.io/name": "citus", "statefulset.kubernetes.io/pod-name": "citus-0"}
	if !reflect.DeepEqual(out.Spec, netv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{MatchLabels: source},
		PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeEgress},
		Egress: []netv1.NetworkPolicyEgressRule{{
			To:    []netv1.NetworkPolicyPeer{platformPeer("os-platform", target)},
			Ports: []netv1.NetworkPolicyPort{tcpPort(5432)},
		}},
	}) {
		t.Fatalf("unexpected system-server egress: %+v", out.Spec)
	}
	for _, ns := range []string{"user-space-alice", "user-system-", "os-framework", "os-platform"} {
		if lookup(ns, out.Name) != nil {
			t.Fatalf("unexpected system-server rule in %s", ns)
		}
	}
	if lookup("os-platform", "olares-platform-systemserver-citus-ingress") != nil {
		t.Fatal("the existing Citus ingress policy must remain authoritative")
	}
}

func TestPlatformTrafficInventoryOnlyAddsVerifiedPorts(t *testing.T) {
	for _, flow := range platformServiceFlows {
		blocked := []int32{8080, 17171, 18001, 18002, 8222}
		if flow.name != "authelia-lldap-token" {
			blocked = append(blocked, 17170)
		}
		if slices.Contains(blocked, flow.servicePort) || slices.Contains(blocked, flow.targetPort) {
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

func TestAutheliaLLDAPTokenTrafficRequiresMatchingService(t *testing.T) {
	key := types.NamespacedName{Namespace: "os-platform", Name: "lldap-service"}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.233.7.245",
			Selector: map[string]string{"app": "lldap"},
			Ports: []corev1.ServicePort{
				{Port: 3890, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(3890)},
				{Port: 17170, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(17170)},
				{Port: 17171, Protocol: corev1.ProtocolTCP, TargetPort: intstr.FromInt(17171)},
			},
		},
	}
	policy := func(namespace, name string, service *corev1.Service) netv1.NetworkPolicy {
		t.Helper()
		for _, p := range BuildPlatformTrafficPolicies(namespace, map[types.NamespacedName]*corev1.Service{key: service}) {
			if p.Name == "olares-platform-"+name {
				if !IsTrafficGroupManagedNP(&p) {
					t.Fatalf("legacy sweep would prune %s/%s", namespace, p.Name)
				}
				return p
			}
		}
		t.Fatalf("policy %s/%s is missing", namespace, name)
		return netv1.NetworkPolicy{}
	}
	source := map[string]string{"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true"}
	out := policy("os-framework", "authelia-lldap-token-egress", svc)
	in := policy(key.Namespace, "authelia-lldap-token-ingress", svc)
	wantPort := []netv1.NetworkPolicyPort{tcpPort(17170)}
	wantEgress := []netv1.NetworkPolicyEgressRule{
		{To: []netv1.NetworkPolicyPeer{platformPeer(key.Namespace, svc.Spec.Selector)}, Ports: wantPort},
		{To: []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.233.7.245/32"}}}, Ports: wantPort},
	}
	wantIngress := []netv1.NetworkPolicyIngressRule{{
		From: []netv1.NetworkPolicyPeer{platformPeer("os-framework", source)}, Ports: wantPort,
	}}
	if !reflect.DeepEqual(out.Spec.PodSelector.MatchLabels, source) || !reflect.DeepEqual(out.Spec.Egress, wantEgress) ||
		!reflect.DeepEqual(in.Spec.PodSelector.MatchLabels, svc.Spec.Selector) || !reflect.DeepEqual(in.Spec.Ingress, wantIngress) {
		t.Fatalf("unexpected LLDAP token API grants: egress=%+v ingress=%+v", out.Spec, in.Spec)
	}
	svc.Spec.Ports = slices.Delete(svc.Spec.Ports, 1, 2)
	if len(policy("os-framework", "authelia-lldap-token-egress", svc).Spec.Egress) != 0 ||
		len(policy(key.Namespace, "authelia-lldap-token-ingress", svc).Spec.Ingress) != 0 ||
		len(policy("os-framework", "authelia-ldap-egress", svc).Spec.Egress) != 2 {
		t.Fatal("removing the token API port must revoke only the token API grants")
	}
}

func TestSystemServerHeadscaleManagementEgress(t *testing.T) {
	key := types.NamespacedName{Namespace: "os-network", Name: "headscale-server-svc"}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.23", Selector: map[string]string{"app": "headscale"},
			Ports: []corev1.ServicePort{
				{Port: 8000, TargetPort: intstr.FromString("management"), Protocol: corev1.ProtocolTCP},
				{Port: 8001, TargetPort: intstr.FromString("policy"), Protocol: corev1.ProtocolTCP},
				{Port: 8080, TargetPort: intstr.FromString("http"), Protocol: corev1.ProtocolTCP},
			},
		},
	}
	if !slices.Contains(SystemServerDependencyServices(), key) {
		t.Fatal("Headscale Service is missing from the route dependency inventory")
	}
	want := []netv1.NetworkPolicyEgressRule{
		{To: []netv1.NetworkPolicyPeer{platformPeer("os-network", map[string]string{"app": "headscale"})}, Ports: []netv1.NetworkPolicyPort{tcpPort(8000)}},
		{To: []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.96.0.23/32"}}}, Ports: []netv1.NetworkPolicyPort{tcpPort(8000)}},
	}
	if got := SystemServerDependencyEgressRules(map[types.NamespacedName]*corev1.Service{key: service}); !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected Headscale management egress: %+v", got)
	}
	for _, bad := range []func(*corev1.Service){
		func(s *corev1.Service) { s.Spec.Ports[0].TargetPort = intstr.FromString("policy") },
		func(s *corev1.Service) { s.Spec.Ports = s.Spec.Ports[1:] },
		func(s *corev1.Service) { s.Spec.Selector["app"] = "unrelated" },
	} {
		mutated := service.DeepCopy()
		bad(mutated)
		if got := SystemServerDependencyEgressRules(map[types.NamespacedName]*corev1.Service{key: mutated}); len(got) != 0 {
			t.Fatalf("unverified Service granted egress: %+v", got)
		}
	}
}

func TestSystemServerDesktopDependenciesRequireVerifiedServices(t *testing.T) {
	lldapKey := types.NamespacedName{Namespace: "os-platform", Name: "lldap-service"}
	marketKey := types.NamespacedName{Namespace: "os-framework", Name: "appstore-svc"}
	monitorKey := types.NamespacedName{Namespace: "os-framework", Name: "monitoring-server"}
	filesKey := types.NamespacedName{Namespace: "os-framework", Name: "files-service"}
	for _, tc := range []struct {
		key         types.NamespacedName
		app         string
		servicePort int32
		targetPort  int32
		wantRules   int
	}{
		{lldapKey, "lldap", 17170, 17170, 2},
		{marketKey, "appstore", 81, 8080, 1},
		{monitorKey, "monitoring-server", 80, 8000, 2},
		{filesKey, "files", 80, 80, 2},
	} {
		t.Run(tc.key.Name, func(t *testing.T) {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Namespace: tc.key.Namespace, Name: tc.key.Name},
				Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.10",
					Selector: map[string]string{"app": tc.app}, Ports: []corev1.ServicePort{{
						Port: tc.servicePort, TargetPort: intstr.FromInt32(tc.targetPort), Protocol: corev1.ProtocolTCP,
					}},
				},
			}
			if !slices.Contains(SystemServerDependencyServices(), tc.key) {
				t.Fatal("route reconciliation cannot discover the dependency Service")
			}
			services := map[types.NamespacedName]*corev1.Service{tc.key: service}
			if got := len(SystemServerDependencyEgressRules(services)); got != tc.wantRules {
				t.Fatalf("verified Service yielded %d rules, want %d", got, tc.wantRules)
			}
			service.Spec.Ports[0].TargetPort = intstr.FromInt32(tc.targetPort + 1)
			if got := SystemServerDependencyEgressRules(services); len(got) != 0 {
				t.Fatalf("retargeted Service retained egress: %+v", got)
			}
			delete(services, tc.key)
			if got := SystemServerDependencyEgressRules(services); len(got) != 0 {
				t.Fatalf("absent Service retained egress: %+v", got)
			}
		})
	}
	tenant := "user-system-alice"
	active := []corev1.Pod{{ObjectMeta: metav1.ObjectMeta{Namespace: tenant, Labels: map[string]string{
		"app": "systemserver", "traffic.olares.io/platform-mesh": "true",
	}}}}
	if !slices.Contains(ActivePlatformTrafficServices(tenant, active), lldapKey) || slices.Contains(ActivePlatformTrafficServices(tenant, nil), lldapKey) {
		t.Fatal("LLDAP token verification is a readiness dependency of the active system-server")
	}
	for _, dep := range ActivePlatformTrafficServices(tenant, active) {
		if dep == monitorKey || dep == filesKey {
			t.Fatalf("optional provider %s must not block tenant namespace isolation", dep)
		}
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

func TestSystemFrontendGrantsUseOnlyVerifiedServicePorts(t *testing.T) {
	space, system := "user-space-alice", "user-system-alice"
	bflKey := types.NamespacedName{Namespace: space, Name: "bfl"}
	natsKey := types.NamespacedName{Namespace: system, Name: "nats"}
	redisKey := types.NamespacedName{Namespace: system, Name: "redis-cluster-proxy"}
	services := map[types.NamespacedName]*corev1.Service{
		bflKey: {
			ObjectMeta: metav1.ObjectMeta{Namespace: space, Name: bflKey.Name},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"tier": "bfl"}, ClusterIP: "10.96.0.13",
				Ports: []corev1.ServicePort{
					{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP},
					{Port: 443, TargetPort: intstr.FromInt(443), Protocol: corev1.ProtocolTCP},
				}},
		},
		natsKey: {
			ObjectMeta: metav1.ObjectMeta{Namespace: system, Name: natsKey.Name},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "systemserver"},
				Ports:    []corev1.ServicePort{{Port: 4222, TargetPort: intstr.FromInt(4222), Protocol: corev1.ProtocolTCP}}},
		},
		redisKey: {
			ObjectMeta: metav1.ObjectMeta{Namespace: system, Name: redisKey.Name},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{"app": "systemserver"}, ClusterIP: "10.96.0.20",
				Ports: []corev1.ServicePort{{Port: 6379, TargetPort: intstr.FromInt(6379), Protocol: corev1.ProtocolTCP}}},
		},
	}
	policies := map[string]netv1.NetworkPolicy{}
	for _, p := range BuildPlatformTrafficPolicies(space, services) {
		policies[p.Name] = p
	}
	for _, tc := range []struct {
		name, app, namespace string
		servicePort, target  int32
	}{
		{"login-bfl", "authelia", space, 80, 8080},
		{"desktop-bfl", "olares-app", space, 80, 8080},
		{"desktop-nats", "olares-app", system, 0, 4222},
		{"desktop-redis", "olares-app", system, 6379, 6379},
	} {
		p := policies["olares-platform-"+tc.name+"-egress"]
		if p.Spec.PodSelector.MatchLabels["app"] != tc.app || len(p.Spec.Egress) == 0 ||
			p.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != tc.namespace ||
			!reflect.DeepEqual(p.Spec.Egress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(tc.target), tcpPort(4143)}) {
			t.Fatalf("unexpected transport for %s: %+v", tc.name, p.Spec)
		}
		if tc.servicePort != 0 &&
			(len(p.Spec.Egress) != 2 || !reflect.DeepEqual(p.Spec.Egress[1].Ports, []netv1.NetworkPolicyPort{tcpPort(tc.servicePort)})) {
			t.Fatalf("unexpected virtual Service ports for %s: %+v", tc.name, p.Spec)
		}
	}
}

func TestDesktopBFLCompatibilityOnlyAllowsSameTenantUnmeshedAPI(t *testing.T) {
	space := "user-space-alice"
	key := types.NamespacedName{Namespace: space, Name: "bfl"}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.13", Selector: map[string]string{"tier": "bfl"},
			Ports: []corev1.ServicePort{
				{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP},
				{Port: 443, TargetPort: intstr.FromInt(443), Protocol: corev1.ProtocolTCP},
			},
		},
	}
	services := map[types.NamespacedName]*corev1.Service{key: svc}
	lookup := func(namespace string) (out, in *netv1.NetworkPolicy) {
		for _, policy := range BuildPlatformTrafficPolicies(namespace, services) {
			switch policy.Name {
			case "olares-platform-desktop-bfl-unmeshed-egress":
				out = &policy
			case "olares-platform-desktop-bfl-unmeshed-ingress":
				in = &policy
			}
		}
		return out, in
	}
	out, in := lookup(space)
	if !IsTrafficGroupManagedNP(out) || !IsTrafficGroupManagedNP(in) ||
		len(out.Spec.Egress) != 1 || len(in.Spec.Ingress) != 1 {
		t.Fatalf("expected a managed egress and ingress rule: %+v %+v", out, in)
	}
	desktop := map[string]string{"app": "olares-app", "linkerd.io/control-plane-ns": "os-mesh"}
	if !reflect.DeepEqual(out.Spec.PodSelector.MatchLabels, desktop) ||
		!reflect.DeepEqual(in.Spec.Ingress[0].From, []netv1.NetworkPolicyPeer{platformPeer(space, desktop)}) ||
		!reflect.DeepEqual(in.Spec.Ingress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(8080)}) {
		t.Fatalf("unexpected Desktop source or BFL ingress port: %+v %+v", out.Spec, in.Spec)
	}
	to := out.Spec.Egress[0].To
	if len(to) != 1 || to[0].IPBlock != nil ||
		!reflect.DeepEqual(to[0].NamespaceSelector, platformPeer(space, nil).NamespaceSelector) ||
		!reflect.DeepEqual(to[0].PodSelector, &in.Spec.PodSelector) ||
		!reflect.DeepEqual(out.Spec.Egress[0].Ports, []netv1.NetworkPolicyPort{tcpPort(8080)}) {
		t.Fatalf("unexpected BFL egress destination: %+v", out.Spec)
	}
	target, err := metav1.LabelSelectorAsSelector(&in.Spec.PodSelector)
	if err != nil || !target.Matches(labels.Set{"tier": "bfl"}) ||
		target.Matches(labels.Set{"tier": "other"}) ||
		target.Matches(labels.Set{"tier": "bfl", "linkerd.io/control-plane-ns": "os-mesh"}) ||
		target.Matches(labels.Set{"tier": "bfl", "linkerd.io/control-plane-ns": "another-mesh"}) {
		t.Fatalf("the compatibility policy must select only unmeshed BFL Pods: %v %v", target, err)
	}
	if otherOut, otherIn := lookup("user-space-bob"); len(otherOut.Spec.Egress) != 0 || len(otherIn.Spec.Ingress) != 0 {
		t.Fatal("a tenant received compatibility grants from another tenant's Service")
	}
	svc.Spec.Ports[0].TargetPort = intstr.FromInt(443)
	out, in = lookup(space)
	if len(out.Spec.Egress) != 0 || len(in.Spec.Ingress) != 0 {
		t.Fatal("retargeting the BFL API must withdraw both grants")
	}
	delete(services, key)
	out, in = lookup(space)
	if len(out.Spec.Egress) != 0 || len(in.Spec.Ingress) != 0 {
		t.Fatal("deleting the Service must withdraw both grants")
	}
}

func TestBFLMeshTransportForRegisteredCallers(t *testing.T) {
	key := types.NamespacedName{Namespace: "user-space-alice", Name: "bfl"}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.13",
			Selector: map[string]string{"tier": "bfl"},
			Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(8080), Protocol: corev1.ProtocolTCP}},
		},
	}
	services := map[types.NamespacedName]*corev1.Service{key: svc}
	lookup := func(namespace, name string) netv1.NetworkPolicy {
		t.Helper()
		for _, policy := range BuildPlatformTrafficPolicies(namespace, services) {
			if policy.Name == "olares-platform-"+name {
				if !IsTrafficGroupManagedNP(&policy) {
					t.Fatalf("legacy sweep would prune %s", policy.Name)
				}
				return policy
			}
		}
		t.Fatalf("missing policy %s/%s", namespace, name)
		return netv1.NetworkPolicy{}
	}
	ports := []netv1.NetworkPolicyPort{tcpPort(8080), tcpPort(4143)}
	backend := map[string]string{"tier": "bfl", "linkerd.io/control-plane-ns": "os-mesh"}
	for _, tc := range []struct {
		name, namespace string
		selector        map[string]string
	}{
		{"authelia-bfl-mesh", "os-framework", map[string]string{
			"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"bfl-cert-refresh", key.Namespace, map[string]string{
			"app.kubernetes.io/name": "bfl-cert-refresh", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"desktop-bfl", key.Namespace, map[string]string{
			"app": "olares-app", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"login-bfl", key.Namespace, map[string]string{
			"app": "authelia", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"wizard-bfl", key.Namespace, map[string]string{
			"app": "wizard", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"tapr-images-bfl", "user-system-alice", map[string]string{
			"app": "tapr-images", "linkerd.io/control-plane-ns": "os-mesh",
		}},
		{"vault-bfl", "os-framework", map[string]string{
			"app": "vault-server", "linkerd.io/control-plane-ns": "os-mesh",
		}},
	} {
		out, in := lookup(tc.namespace, tc.name+"-egress"), lookup(key.Namespace, tc.name+"-ingress")
		if len(out.Spec.Egress) != 2 || len(in.Spec.Ingress) != 1 ||
			!reflect.DeepEqual(out.Spec.PodSelector.MatchLabels, tc.selector) ||
			!reflect.DeepEqual(in.Spec.PodSelector.MatchLabels, backend) ||
			!reflect.DeepEqual(out.Spec.Egress[0].To, []netv1.NetworkPolicyPeer{platformPeer(key.Namespace, backend)}) ||
			!reflect.DeepEqual(in.Spec.Ingress[0].From, []netv1.NetworkPolicyPeer{platformPeer(tc.namespace, tc.selector)}) ||
			!reflect.DeepEqual(out.Spec.Egress[0].Ports, ports) || !reflect.DeepEqual(in.Spec.Ingress[0].Ports, ports) {
			t.Fatalf("unexpected mTLS transport for %s: %+v %+v", tc.name, out.Spec, in.Spec)
		}
	}
	delete(services, key)
	if len(lookup("os-framework", "authelia-bfl-mesh-egress").Spec.Egress) != 0 ||
		len(lookup("os-framework", "vault-bfl-egress").Spec.Egress) != 0 ||
		len(lookup(key.Namespace, "bfl-cert-refresh-egress").Spec.Egress) != 0 ||
		len(lookup(key.Namespace, "authelia-bfl-mesh-ingress").Spec.Ingress) != 0 ||
		len(lookup(key.Namespace, "bfl-cert-refresh-ingress").Spec.Ingress) != 0 {
		t.Fatal("deleting the tenant BFL Service must withdraw mesh data-plane grants")
	}
}
