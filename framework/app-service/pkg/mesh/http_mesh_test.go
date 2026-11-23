package mesh

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceWaveEligible(t *testing.T) {
	if NamespaceWaveEligible("wordpress-alice", 0) {
		t.Fatal("wave 0 off")
	}
	if !NamespaceWaveEligible("os-gateway", 1) {
		t.Fatal("wave 1 gateway")
	}
	if NamespaceWaveEligible("wordpress-alice", 3) {
		t.Fatal("wave 3 excludes app")
	}
	if !NamespaceWaveEligible("wordpress-alice", 4) {
		t.Fatal("wave 4 all apps")
	}
}

func TestDesiredInboundPolicy(t *testing.T) {
	if DesiredInboundPolicy(0) != "" {
		t.Fatal("wave 0 no policy")
	}
	if DesiredInboundPolicy(2) != InboundPolicyAllUnauthenticated {
		t.Fatalf("wave 2: %s", DesiredInboundPolicy(2))
	}
	if DesiredInboundPolicy(4) != InboundPolicyClusterAuthenticated {
		t.Fatalf("wave 4: %s", DesiredInboundPolicy(4))
	}
}

func TestShouldInjectHTTPMeshIgnoresDisableInSteadyState(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		LinkerdInjectAnnotation: LinkerdInjectDisabled,
	}}}
	// Product steady state: HTTP mTLS is mandatory; opaque-ports cover non-HTTP.
	if !ShouldInjectHTTPMesh(false, "os-gateway", 4, pod) {
		t.Fatal("wave-4 must inject even when annotation says disabled")
	}
}

func TestAnnotatePodForHTTPMesh(t *testing.T) {
	pod := &corev1.Pod{}
	AnnotatePodForHTTPMesh(pod, true, 4, "3306")
	if pod.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatal("inject")
	}
	if pod.Annotations[LinkerdServerTLSModeAnnotation] != MTLSModeRequired {
		t.Fatalf("mode=%s", pod.Annotations[LinkerdServerTLSModeAnnotation])
	}
	if pod.Annotations[LinkerdDefaultInboundPolicyAnnotation] != InboundPolicyClusterAuthenticated {
		t.Fatalf("inbound=%s", pod.Annotations[LinkerdDefaultInboundPolicyAnnotation])
	}
	if pod.Annotations[LinkerdOpaquePortsAnnotation] != "3306" {
		t.Fatal("opaque")
	}

	early := &corev1.Pod{}
	AnnotatePodForHTTPMesh(early, true, 1, "")
	if early.Annotations[LinkerdDefaultInboundPolicyAnnotation] != InboundPolicyAllUnauthenticated {
		t.Fatalf("wave1 inbound=%s", early.Annotations[LinkerdDefaultInboundPolicyAnnotation])
	}
}
