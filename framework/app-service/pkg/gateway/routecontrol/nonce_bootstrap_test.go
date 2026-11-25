package routecontrol

import (
	"reflect"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSystemServerBootstrapAPIEgressBeforeReadiness(t *testing.T) {
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "kubernetes"}, Spec: corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP, ClusterIP: "10.96.0.1", Ports: []corev1.ServicePort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 443}},
	}}
	ep := &corev1.Endpoints{ObjectMeta: svc.ObjectMeta, Subsets: []corev1.EndpointSubset{{
		Addresses: []corev1.EndpointAddress{{IP: "192.0.2.10"}}, Ports: []corev1.EndpointPort{{Name: "https", Protocol: corev1.ProtocolTCP, Port: 6443}},
	}}}
	c := fake.NewClientBuilder().WithScheme(testSchemeWithSecurityPolicy(t)).WithObjects(svc, ep).Build()
	source := meshTransportPeer{"user-system-alice", map[string]string{"app": "systemserver", "linkerd.io/control-plane-ns": "os-mesh"}}
	policies, err := authTransportPair(t.Context(), c, nil, "nonce-test", source, meshTransportBackend{})
	if err != nil {
		t.Fatal(err)
	}
	rules := policies[0].Spec.Egress
	dependencies := security.BuildTrafficGroupDependencyPolicies("user-system-alice", svc, ep)
	if len(dependencies) != 1 || !reflect.DeepEqual(rules[len(rules)-len(dependencies[0].Spec.Egress):], dependencies[0].Spec.Egress) {
		t.Fatal("system-server discovery lost its API egress before readiness")
	}
	seen := map[int32]bool{}
	for _, rule := range rules {
		for _, peer := range rule.To {
			if peer.IPBlock != nil && peer.IPBlock.CIDR != "10.96.0.1/32" && peer.IPBlock.CIDR != "192.0.2.10/32" {
				t.Fatal("API egress widened")
			}
			if peer.NamespaceSelector != nil && peer.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] == "os-mesh" {
				if peer.PodSelector == nil || len(peer.PodSelector.MatchLabels) != 1 {
					t.Fatal("mesh bootstrap lacks component restriction")
				}
				for _, port := range rule.Ports {
					if port.Port.Type == intstr.Int {
						seen[port.Port.IntVal] = true
					}
				}
			}
		}
	}
	for _, port := range []int32{8080, 8086, 8090} {
		if !seen[port] {
			t.Fatalf("mesh bootstrap port %d missing", port)
		}
	}
}
