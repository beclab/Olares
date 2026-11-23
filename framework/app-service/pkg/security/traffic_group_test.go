package security

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func TestInferTrafficGroup(t *testing.T) {
	cases := map[string]string{
		"os-mesh":           TrafficGroupMesh,
		"os-gateway":        TrafficGroupGateway,
		"os-network":        TrafficGroupL4,
		"os-platform":       TrafficGroupOSSystem,
		"os-framework":      TrafficGroupOSFramework,
		"user-space-alice":  TrafficGroupUserSpace,
		"user-system-alice": TrafficGroupUserSpace,
		"wordpress-alice":   TrafficGroupApp,
	}
	for ns, want := range cases {
		if got := InferTrafficGroup(ns); got != want {
			t.Fatalf("%s: got %s want %s", ns, got, want)
		}
	}
}

func TestLabelNamespaceTrafficGroupIdempotent(t *testing.T) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "os-gateway"}}
	LabelNamespaceTrafficGroup(ns)
	LabelNamespaceTrafficGroup(ns)
	if ns.Labels[TrafficGroupLabel] != TrafficGroupGateway {
		t.Fatalf("label=%v", ns.Labels)
	}
}

func TestTrafficGroupNPEnabled(t *testing.T) {
	if TrafficGroupNPEnabled("") || TrafficGroupNPEnabled("false") {
		t.Fatal("default off")
	}
	if !TrafficGroupNPEnabled("true") || !TrafficGroupNPEnabled("1") {
		t.Fatal("opt-in")
	}
}

func TestBuildTrafficGroupIsolationBundleSafe(t *testing.T) {
	for _, ns := range []string{"os-gateway", "wordpress-alice", "user-space-bob"} {
		bundle := BuildTrafficGroupIsolationBundle(ns)
		if len(bundle) < 3 {
			t.Fatalf("%s: short bundle %d", ns, len(bundle))
		}
		if !BundleHasInboundAllow(bundle) {
			t.Fatalf("%s: bundle must include inbound allow (AC-W3)", ns)
		}
		var hasDeny, hasSame bool
		for _, p := range bundle {
			switch p.Name {
			case "olares-traffic-group-default-deny":
				hasDeny = true
				if len(p.Spec.Ingress) != 0 {
					t.Fatalf("%s: default-deny must keep empty ingress list", ns)
				}
			case "olares-traffic-group-same":
				hasSame = true
			}
		}
		if !hasDeny || !hasSame {
			t.Fatalf("%s: missing deny/same policies", ns)
		}
	}
	gw := BuildTrafficGroupIsolationBundle("os-gateway")
	var fromL4 bool
	for _, p := range gw {
		if p.Name == "olares-traffic-group-from-l4" {
			fromL4 = true
		}
	}
	if !fromL4 {
		t.Fatal("gateway bundle must allow L4 ingress")
	}
	app := BuildTrafficGroupIsolationBundle("wordpress-alice")
	var fromGW bool
	for _, p := range app {
		if p.Name == "olares-traffic-group-from-gateway" {
			fromGW = true
		}
	}
	if !fromGW {
		t.Fatal("app bundle must allow gateway ingress")
	}
}

func TestSameTrafficGroupKeepsTenantNamespacesSeparate(t *testing.T) {
	for _, target := range []string{"user-space-alice", "user-system-alice"} {
		policy := NPAllowSameTrafficGroup(target, TrafficGroupUserSpace)
		selector, err := metav1.LabelSelectorAsSelector(policy.Spec.Ingress[0].From[0].NamespaceSelector)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			namespace, group string
			allowed          bool
		}{
			{"user-space-alice", TrafficGroupUserSpace, true},
			{"user-system-alice", TrafficGroupUserSpace, true},
			{"user-space-bob", TrafficGroupUserSpace, false},
			{"user-system-bob", TrafficGroupUserSpace, false},
			{"user-space-alice", TrafficGroupApp, false},
		} {
			if matched := selector.Matches(labels.Set{
				"kubernetes.io/metadata.name": tc.namespace,
				TrafficGroupLabel:             tc.group,
			}); matched != tc.allowed {
				t.Fatalf("%s receiving from %s (%s): matched=%v, want %v", target, tc.namespace, tc.group, matched, tc.allowed)
			}
		}
	}
}
