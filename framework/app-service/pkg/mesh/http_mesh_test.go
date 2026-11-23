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

func TestNativeLinkerdProxyAdmissionAndReadiness(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways
	proxy := corev1.Container{
		Name: LinkerdProxyContainer, RestartPolicy: &always,
		Env: []corev1.EnvVar{{Name: LinkerdProxyInboundPolicyEnv, Value: InboundPolicyAllUnauthenticated}},
	}
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{InitContainers: []corev1.Container{{Name: "linkerd-init"}, proxy}},
		Status: corev1.PodStatus{
			ContainerStatuses:     []corev1.ContainerStatus{{Name: LinkerdProxyContainer, Ready: true}},
			InitContainerStatuses: []corev1.ContainerStatus{{Name: LinkerdProxyContainer}},
		},
	}
	if LinkerdProxyReady(pod) {
		t.Fatal("readiness from the regular container list cannot satisfy a native proxy")
	}
	if changed, err := ConfigurePodHTTPMesh(pod, "user-space-alice", 4, "", true, false); err != nil || !changed {
		t.Fatalf("configure native proxy: changed=%v err=%v", changed, err)
	}
	pod.Status.InitContainerStatuses[0].Ready = true
	if !LinkerdProxyReady(pod) || !ProxyHasInboundPolicy(pod, InboundPolicyClusterAuthenticated) {
		t.Fatal("native proxy should be ready and use the enforced inbound policy")
	}
	pod.Spec.InitContainers[1].Env = []corev1.EnvVar{{Name: "UNRELATED_ENV", Value: InboundPolicyClusterAuthenticated}}
	if ProxyHasInboundPolicy(pod, InboundPolicyClusterAuthenticated) {
		t.Fatal("an unrelated environment variable cannot satisfy the inbound policy gate")
	}
	pod.Spec.InitContainers[1].RestartPolicy = nil
	if LinkerdProxyReady(pod) || ProxyHasInboundPolicy(pod, InboundPolicyClusterAuthenticated) {
		t.Fatal("a one-shot init container is not a native sidecar")
	}
}

func TestHTTPMeshInitRejectsNativeProxyBypass(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want bool
	}{
		{"built-in probe and RDP exclusions", []string{"--inbound-ports-to-ignore", "4190,4191,3389"}, true},
		{"empty outbound exclusions", []string{"--outbound-ports-to-ignore="}, true},
		{"inbound HTTP exclusion", []string{"--inbound-ports-to-ignore=4191,9091"}, false},
		{"outbound RDP exclusion", []string{"--outbound-ports-to-ignore", "3389"}, false},
		{"subnet exclusion", []string{"--subnets-to-ignore", "10.0.0.0/8"}, false},
		{"malformed flag", []string{"--inbound-ports-to-ignore"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			always := corev1.ContainerRestartPolicyAlways
			pod := &corev1.Pod{Spec: corev1.PodSpec{InitContainers: []corev1.Container{
				{Name: "linkerd-init", Args: tc.args},
				{Name: LinkerdProxyContainer, RestartPolicy: &always},
				{Name: "check-auth"},
			}}}
			if got := HasHTTPMeshInit(pod); got != tc.want {
				t.Fatalf("interception check = %v, want %v", got, tc.want)
			}
			pod.Spec.InitContainers = pod.Spec.InitContainers[1:]
			if HasHTTPMeshInit(pod) {
				t.Fatal("native proxy is insufficient when the interception init container is missing")
			}
		})
	}
}

func TestParseHTTPMeshWaveBounds(t *testing.T) {
	for input, expected := range map[string]int{
		"": 0, "invalid": 0, "-1": 0, "2": 2, " 4 ": 4, "5": 4,
	} {
		if actual := ParseHTTPMeshWave(input); actual != expected {
			t.Errorf("wave %q = %d, want %d", input, actual, expected)
		}
	}
}

func TestConfigureHTTPMeshRespectsRolloutEligibility(t *testing.T) {
	for _, tc := range []struct {
		namespace string
		wave      int
		eligible  bool
		legacy    bool
		expected  bool
	}{
		{"os-platform", 1, true, false, false},
		{"os-platform", 2, true, false, true},
		{"user-system-alice", 2, true, false, false},
		{"user-system-alice", 3, true, false, true},
		{"user-system-alice", 4, false, true, false},
		{"user-system-alice", 0, true, true, true},
	} {
		pod := &corev1.Pod{}
		changed, err := ConfigurePodHTTPMesh(pod, tc.namespace, tc.wave, "", tc.eligible, tc.legacy)
		if err != nil || changed != tc.expected || (pod.Annotations[LinkerdInjectAnnotation] == LinkerdInjectEnabled) != tc.expected {
			t.Errorf("%s wave %d: changed=%v, want %v, annotations=%v, err=%v", tc.namespace, tc.wave, changed, tc.expected, pod.Annotations, err)
		}
	}
}

func TestExplicitHTTPMeshEnrollmentCannotBeDowngradedByRollout(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
			LinkerdServerTLSModeAnnotation: MTLSModeRequired,
		}},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name: LinkerdProxyContainer,
			Env:  []corev1.EnvVar{{Name: LinkerdProxyInboundPolicyEnv, Value: InboundPolicyClusterAuthenticated}},
		}}},
	}
	if changed, err := ConfigurePodHTTPMesh(pod, "user-space-alice", 3, "", true, false); !changed || err != nil {
		t.Fatalf("configure enrolled pod: changed=%v err=%v", changed, err)
	}
	if pod.Annotations[LinkerdServerTLSModeAnnotation] != MTLSModeRequired ||
		pod.Annotations[LinkerdDefaultInboundPolicyAnnotation] != InboundPolicyClusterAuthenticated ||
		!ProxyHasInboundPolicy(pod, InboundPolicyClusterAuthenticated) {
		t.Fatal("application rollout downgraded a required workload")
	}
}
