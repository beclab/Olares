package security

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func TestTrafficGroupDependencyPolicies(t *testing.T) {
	meta := metav1.ObjectMeta{Namespace: "default", Name: "kubernetes"}
	service := &corev1.Service{
		ObjectMeta: meta,
		Spec: corev1.ServiceSpec{
			Type:       corev1.ServiceTypeClusterIP,
			ClusterIPs: []string{"10.233.0.1", "fd00::1", "invalid"},
			Ports: []corev1.ServicePort{
				{Name: "https", Protocol: corev1.ProtocolTCP, Port: 443},
				{Name: "other", Protocol: corev1.ProtocolTCP, Port: 8080},
			},
		},
	}
	endpoints := &corev1.Endpoints{
		ObjectMeta: meta,
		Subsets: []corev1.EndpointSubset{{
			Addresses:         []corev1.EndpointAddress{{IP: "192.168.128.102"}, {IP: "192.168.128.102"}, {IP: "0.0.0.0"}},
			NotReadyAddresses: []corev1.EndpointAddress{{IP: "192.168.128.103"}},
			Ports:             []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 6443}},
		}},
	}
	dns := BuildTrafficGroupDependencyPolicies("kube-system", service, endpoints)
	auth := BuildTrafficGroupDependencyPolicies("os-framework", service, endpoints)
	if len(dns) != 4 || len(auth) != 5 || len(BuildTrafficGroupDependencyPolicies("wordpress-test", service, endpoints)) != 0 {
		t.Fatal("unexpected dependency policy scope")
	}
	dnsSelector := map[string]string{"k8s-app": "kube-dns"}
	if spec := dns[0].Spec; !reflect.DeepEqual(spec.PodSelector.MatchLabels, dnsSelector) ||
		!reflect.DeepEqual(spec.PolicyTypes, []netv1.PolicyType{netv1.PolicyTypeIngress}) ||
		len(spec.Ingress) != 1 || len(spec.Ingress[0].From) != 1 ||
		!reflect.DeepEqual(spec.Ingress[0].From[0], netv1.NetworkPolicyPeer{NamespaceSelector: &metav1.LabelSelector{}}) ||
		!reflect.DeepEqual(spec.Ingress[0].Ports, []netv1.NetworkPolicyPort{
			{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt(53))},
			{Protocol: ptr.To(corev1.ProtocolUDP), Port: ptr.To(intstr.FromInt(53))},
		}) {
		t.Fatalf("unexpected CoreDNS ingress: %+v", spec)
	}
	want := map[string]int32{"10.233.0.1/32": 443, "192.168.128.102/32": 6443, "fd00::1/128": 443}
	for _, item := range []struct {
		policy   netv1.NetworkPolicy
		selector map[string]string
	}{
		{dns[1], dnsSelector},
		{auth[0], map[string]string{"app": "authelia-backend", "traffic.olares.io/platform-mesh": "true"}},
	} {
		spec := item.policy.Spec
		got := make(map[string]int32)
		for _, rule := range spec.Egress {
			if len(rule.To) != 1 || rule.To[0].IPBlock == nil || len(rule.Ports) != 1 ||
				rule.Ports[0].Port == nil || rule.Ports[0].Protocol == nil || *rule.Ports[0].Protocol != corev1.ProtocolTCP {
				t.Fatalf("egress is not scoped to one address and TCP port: %+v", rule)
			}
			got[rule.To[0].IPBlock.CIDR] = int32(rule.Ports[0].Port.IntValue())
		}
		if !reflect.DeepEqual(spec.PodSelector.MatchLabels, item.selector) ||
			!reflect.DeepEqual(spec.PolicyTypes, []netv1.PolicyType{netv1.PolicyTypeEgress}) ||
			len(spec.Egress) != len(want) || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: unexpected API destinations: %v", item.policy.Namespace, got)
		}
	}

	for _, policy := range auth {
		if policy.Name != "olares-platform-webhook-app-service" {
			continue
		}
		rules := policy.Spec.Ingress
		if len(rules) != 1 || !reflect.DeepEqual(rules[0].Ports, []netv1.NetworkPolicyPort{tcpPort(8433)}) ||
			!reflect.DeepEqual(rules[0].From, []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "192.168.128.102/32"}}}) {
			t.Fatalf("webhook must use real API sources and the Pod target port: %+v", rules)
		}
	}

	// Do not trust endpoint data from an unrelated namespace, or retain stale
	// endpoint destinations while the canonical Endpoints object is absent.
	endpoints.Namespace = "user-space-test"
	service.Spec.ClusterIP = "10.233.0.1"
	service.Spec.ClusterIPs = nil
	wantEgress := []netv1.NetworkPolicyEgressRule{{
		To:    []netv1.NetworkPolicyPeer{{IPBlock: &netv1.IPBlock{CIDR: "10.233.0.1/32"}}},
		Ports: []netv1.NetworkPolicyPort{{Protocol: ptr.To(corev1.ProtocolTCP), Port: ptr.To(intstr.FromInt(443))}},
	}}
	for _, source := range []*corev1.Endpoints{endpoints, nil} {
		if got := BuildTrafficGroupDependencyPolicies("os-framework", service, source)[0].Spec.Egress; !reflect.DeepEqual(got, wantEgress) {
			t.Fatalf("unexpected destinations without canonical endpoints: %+v", got)
		}
	}

	// Endpoints without a valid canonical Service grant nothing.
	service.Spec.Selector = map[string]string{"app": "unrelated"}
	if got := BuildTrafficGroupDependencyPolicies("os-framework", service, endpoints)[0].Spec.Egress; len(got) != 0 {
		t.Fatalf("unexpected egress for invalid API Service: %+v", got)
	}
}

func TestTrafficGroupPolicyOwnership(t *testing.T) {
	if IsTrafficGroupManagedNP(nil) {
		t.Fatal("nil policy claimed")
	}
	for _, namespace := range []string{"kube-system", "os-framework", "os-gateway", "user-space-test"} {
		policies := append(BuildTrafficGroupIsolationBundle(namespace), BuildTrafficGroupDependencyPolicies(namespace, nil, nil)...)
		for _, policy := range policies {
			if !IsTrafficGroupManagedNP(&policy) {
				t.Fatalf("candidate policy %s/%s not protected", namespace, policy.Name)
			}
			foreign := policy.DeepCopy()
			delete(foreign.Labels, "app.kubernetes.io/managed-by")
			if IsTrafficGroupManagedNP(foreign) {
				t.Fatalf("unowned policy %s/%s claimed", namespace, policy.Name)
			}
		}
	}
	for _, meta := range []metav1.ObjectMeta{
		npMeta(coreDNSIngressName, "user-space-test"),
		npMeta(kubernetesAPIEgressName, "os-gateway"),
		npMeta("olares-traffic-group-retired", "os-framework"),
	} {
		if IsTrafficGroupManagedNP(&netv1.NetworkPolicy{ObjectMeta: meta}) {
			t.Fatalf("unexpected policy %s/%s protected", meta.Namespace, meta.Name)
		}
	}
}

func TestTenantControlPlaneAllowsOnlySystemWorkloads(t *testing.T) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "kubernetes"},
		Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.1",
			Ports: []corev1.ServicePort{{Name: "https", Port: 443, Protocol: corev1.ProtocolTCP}}},
	}
	for _, tc := range []struct{ namespace, key, value string }{
		{"user-system-alice", "app", "systemserver"},
		{"user-space-alice", "tier", "bfl"},
	} {
		policies := BuildTrafficGroupDependencyPolicies(tc.namespace, service, nil)
		if len(policies) != 1 || !UsesKubernetesAPIPolicy(tc.namespace) ||
			!reflect.DeepEqual(policies[0].Spec.PodSelector.MatchLabels, map[string]string{tc.key: tc.value}) ||
			!reflect.DeepEqual(policies[0].Spec.Egress, kubernetesAPIEgress(service, nil)) ||
			!IsTrafficGroupManagedNP(&policies[0]) {
			t.Fatalf("unexpected tenant Kubernetes API permit in %s: %+v", tc.namespace, policies)
		}
	}
	if UsesKubernetesAPIPolicy("third-party-alice") || len(BuildTrafficGroupDependencyPolicies("third-party-alice", service, nil)) != 0 {
		t.Fatal("a third-party namespace received the tenant API policy")
	}
}
