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

func TestConfigurePodHTTPMeshForPodsWithoutApplicationSidecars(t *testing.T) {
	pod := &corev1.Pod{Spec: corev1.PodSpec{Containers: []corev1.Container{{
		Name: LinkerdProxyContainer, Env: []corev1.EnvVar{{Name: LinkerdProxyInboundPolicyEnv, Value: InboundPolicyAllUnauthenticated}},
	}}}}
	changed, err := ConfigurePodHTTPMesh(pod, "shared", 4, "", true, false)
	if err != nil || !changed || pod.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled ||
		!ProxyHasInboundPolicy(pod, InboundPolicyClusterAuthenticated) {
		t.Fatalf("pod did not adopt strict HTTP mesh posture: changed=%v err=%v", changed, err)
	}
}

func TestConfigurePodHTTPMeshRejectsInterceptionBypasses(t *testing.T) {
	for _, mutation := range []func(*corev1.Pod){
		func(p *corev1.Pod) { p.Spec.HostNetwork = true },
		func(p *corev1.Pod) { p.Spec.Containers[0].Ports[0].HostPort = 8080 },
		func(p *corev1.Pod) { p.Annotations[LinkerdSkipInboundPortsAnnotation] = "8080" },
		func(p *corev1.Pod) { p.Annotations[LinkerdSkipOutboundPortsAnnotation] = "80" },
		func(p *corev1.Pod) { p.Annotations[LinkerdSkipSubnetsAnnotation] = "10.0.0.0/8" },
	} {
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Ports: []corev1.ContainerPort{{ContainerPort: 8080}}}}}}
		mutation(pod)
		if changed, err := ConfigurePodHTTPMesh(pod, "app-alice", 4, "", true, false); changed || err == nil {
			t.Fatalf("unexpected admission posture: changed=%v err=%v", changed, err)
		}
	}
}
