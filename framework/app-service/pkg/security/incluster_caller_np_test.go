package security

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNewCallerToAppGatewayEgressNP(t *testing.T) {
	np := NewCallerToAppGatewayEgressNP("user-space-alice", "app-gateway")
	if np.Name != CallerToAppGatewayEgressNPName {
		t.Fatalf("name = %q", np.Name)
	}
	if np.Namespace != "user-space-alice" {
		t.Fatalf("namespace = %q", np.Namespace)
	}
	if np.Labels["app.kubernetes.io/managed-by"] != "app-service" {
		t.Fatalf("managed-by label missing")
	}
	if np.Labels["app.kubernetes.io/component"] != "route-control" {
		t.Fatalf("component label = %q", np.Labels["app.kubernetes.io/component"])
	}
	if len(np.Spec.Egress) != 1 || len(np.Spec.Egress[0].Ports) != 2 {
		t.Fatalf("egress ports = %#v", np.Spec.Egress)
	}
	wantPorts := map[int32]bool{80: true, 443: true}
	for _, p := range np.Spec.Egress[0].Ports {
		if p.Protocol == nil || *p.Protocol != corev1.ProtocolTCP {
			t.Fatalf("protocol = %v", p.Protocol)
		}
		wantPorts[p.Port.IntVal] = false
	}
	for port, missing := range wantPorts {
		if missing {
			t.Fatalf("missing port %d", port)
		}
	}
	peer := np.Spec.Egress[0].To[0]
	if peer.NamespaceSelector.MatchLabels["kubernetes.io/metadata.name"] != "app-gateway" {
		t.Fatalf("gateway ns selector = %#v", peer.NamespaceSelector)
	}
	if peer.PodSelector.MatchLabels[egOwningGatewayLabel] != defaultOwningGatewayName {
		t.Fatalf("eg owning label = %#v", peer.PodSelector.MatchLabels)
	}
}

func TestNewCallerMeshEgressNP(t *testing.T) {
	np := NewCallerMeshEgressNP("user-space-alice")
	if np.Name != CallerMeshEgressNPName {
		t.Fatalf("name = %q", np.Name)
	}
	if len(np.Spec.Egress) != 1 {
		t.Fatalf("egress rules = %d", len(np.Spec.Egress))
	}
	if len(np.Spec.Egress[0].To) != 2 {
		t.Fatalf("peers = %d", len(np.Spec.Egress[0].To))
	}
	wantPorts := map[int32]bool{8080: true, 8086: true, 8090: true}
	for _, p := range np.Spec.Egress[0].Ports {
		wantPorts[p.Port.IntVal] = false
	}
	for port, missing := range wantPorts {
		if missing {
			t.Fatalf("missing mesh port %d", port)
		}
	}
}
