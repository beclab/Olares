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

func TestShouldInjectHTTPMeshRespectsDisable(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
		LinkerdInjectAnnotation: LinkerdInjectDisabled,
	}}}
	if ShouldInjectHTTPMesh(true, "os-gateway", 4, pod) {
		t.Fatal("explicit disable must win")
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
	if pod.Annotations[LinkerdOpaquePortsAnnotation] != "3306" {
		t.Fatal("opaque")
	}
}
