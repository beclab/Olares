package security

import "testing"

func TestInferTrafficGroup(t *testing.T) {
	cases := map[string]string{
		"os-mesh":             TrafficGroupMesh,
		"os-gateway":          TrafficGroupGateway,
		"os-framework":        TrafficGroupOSFramework,
		"os-platform":         TrafficGroupOSSystem,
		"user-space-alice":    TrafficGroupUserSpace,
		"user-system-alice":   TrafficGroupUserSpace,
		"wordpress-alice":     TrafficGroupApp,
	}
	for ns, want := range cases {
		if got := InferTrafficGroup(ns); got != want {
			t.Fatalf("%s: got %s want %s", ns, got, want)
		}
	}
}

func TestNPGroupDefaultDenyKeepsMeshAndDNS(t *testing.T) {
	np := NPGroupDefaultDeny("wordpress-alice")
	if len(np.Spec.PolicyTypes) != 2 {
		t.Fatalf("policyTypes=%v", np.Spec.PolicyTypes)
	}
	if len(np.Spec.Ingress) != 0 {
		t.Fatal("ingress must be empty deny")
	}
	if len(np.Spec.Egress) < 2 {
		t.Fatal("expected mesh and DNS egress")
	}
	same := NPAllowSameTrafficGroup("wordpress-alice", TrafficGroupApp)
	if same.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[TrafficGroupLabel] != TrafficGroupApp {
		t.Fatal("same-group selector mismatch")
	}
}
