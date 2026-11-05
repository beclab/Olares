package resources

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

func TestCheckHostNetwork_RejectsTrue(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithPodSpec("web", corev1.PodSpec{HostNetwork: true, HostPID: true}),
	}
	err := CheckHostNetwork(list)
	if err == nil {
		t.Fatal("expected hostNetwork and hostPID to be rejected")
	}
	msg := err.Error()
	if !strings.Contains(msg, "hostNetwork") || !strings.Contains(msg, "hostPID") {
		t.Fatalf("error should name both fields, got: %v", err)
	}
	if !strings.Contains(msg, "web") {
		t.Fatalf("error should name the workload, got: %v", err)
	}
}

func TestCheckHostNetwork_AllowsFalseAndUnset(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithPodSpec("web", corev1.PodSpec{}),
		deploymentWithPodSpec("api", corev1.PodSpec{HostNetwork: false, HostPID: false}),
	}
	if err := CheckHostNetwork(list); err != nil {
		t.Fatalf("false and unset must pass: %v", err)
	}
}
