package mesh

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestHardenLinkerdProxyAdminProbesConvertsHTTPGet(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app"},
				{
					Name: LinkerdProxyContainerName,
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/live",
								Port: intstr.FromInt(constants.LinkerdAdminPort),
							},
						},
						InitialDelaySeconds: 10,
						TimeoutSeconds:      1,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/ready",
								Port: intstr.FromString(LinkerdAdminPortName),
							},
						},
						InitialDelaySeconds: 2,
						TimeoutSeconds:      1,
					},
				},
			},
		},
	}
	if !HardenLinkerdProxyAdminProbes(pod) {
		t.Fatal("expected probes to change")
	}
	c := pod.Spec.Containers[1]
	if c.LivenessProbe.HTTPGet != nil || c.LivenessProbe.Exec == nil {
		t.Fatalf("liveness not converted: %#v", c.LivenessProbe)
	}
	if c.ReadinessProbe.HTTPGet != nil || c.ReadinessProbe.Exec == nil {
		t.Fatalf("readiness not converted: %#v", c.ReadinessProbe)
	}
	want := []string{LinkerdAwaitBinary, "--timeout=1s", "--port=4191"}
	for i, s := range want {
		if c.LivenessProbe.Exec.Command[i] != s {
			t.Fatalf("liveness cmd[%d]=%q want %q", i, c.LivenessProbe.Exec.Command[i], s)
		}
	}
	if c.LivenessProbe.TimeoutSeconds < 2 {
		t.Fatalf("liveness TimeoutSeconds=%d want >=2", c.LivenessProbe.TimeoutSeconds)
	}
	if c.LivenessProbe.InitialDelaySeconds != 10 {
		t.Fatalf("liveness InitialDelaySeconds changed: %d", c.LivenessProbe.InitialDelaySeconds)
	}
	// idempotent
	if HardenLinkerdProxyAdminProbes(pod) {
		t.Fatal("second harden should be no-op")
	}
}

func TestHardenLinkerdProxyAdminProbesSkipsNonAdmin(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: LinkerdProxyContainerName,
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/live",
							Port: intstr.FromInt(8080),
						},
					},
				},
			}},
		},
	}
	if HardenLinkerdProxyAdminProbes(pod) {
		t.Fatal("non-admin probe must not change")
	}
	if pod.Spec.Containers[0].LivenessProbe.HTTPGet == nil {
		t.Fatal("httpGet cleared unexpectedly")
	}
}

func TestHardenLinkerdProxyAdminProbesNoProxy(t *testing.T) {
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app"}},
		},
	}
	if HardenLinkerdProxyAdminProbes(pod) {
		t.Fatal("no linkerd-proxy: expect false")
	}
	if HardenLinkerdProxyAdminProbes(nil) {
		t.Fatal("nil pod: expect false")
	}
}
