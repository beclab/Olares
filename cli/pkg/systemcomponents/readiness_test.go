package systemcomponents

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ptr[T any](v T) *T { return &v }

// natsPod rebuilds os-platform/nats-0 as it looked while enable-juicefs was
// waiting on it: cm-sidecar is a sidecar that had already restarted once and is
// running and ready, and the main containers it gates are serving.
func natsPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "os-platform", Name: "nats-0"},
		Spec: corev1.PodSpec{
			InitContainers: []corev1.Container{{
				Name:          "cm-sidecar",
				RestartPolicy: ptr(corev1.ContainerRestartPolicyAlways),
			}},
			Containers: []corev1.Container{
				{Name: "nats"},
				{Name: "reloader"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodInitialized, Status: corev1.ConditionTrue},
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
				{Type: corev1.ContainersReady, Status: corev1.ConditionTrue},
			},
			InitContainerStatuses: []corev1.ContainerStatus{{
				Name:         "cm-sidecar",
				State:        corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
				Ready:        true,
				Started:      ptr(true),
				RestartCount: 1,
			}},
			ContainerStatuses: []corev1.ContainerStatus{
				{Name: "nats", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}, Ready: true},
				{Name: "reloader", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}, Ready: true},
			},
		},
	}
}

// A running sidecar never terminates, so treating it like a plain init
// container made this pod permanently unready and hung every node scoped wait
// (enable-juicefs, change-ip, start, staged upgrade) until it ran out of
// retries.
func TestAssertPodReadyAcceptsARunningSidecar(t *testing.T) {
	if err := AssertPodReady(natsPod()); err != nil {
		t.Fatalf("pod with a ready sidecar reported not ready: %v", err)
	}
}

func TestAssertPodReadyRejectsAnUnhealthySidecar(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*corev1.Pod)
		want   string
	}{
		{
			name: "crash looping",
			mutate: func(p *corev1.Pod) {
				p.Status.InitContainerStatuses[0].State = corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
				}
				p.Status.InitContainerStatuses[0].Ready = false
				p.Status.InitContainerStatuses[0].Started = ptr(false)
			},
			want: "sidecar cm-sidecar in pod os-platform/nats-0 is waiting (reason=CrashLoopBackOff",
		},
		{
			name: "gave up and exited",
			mutate: func(p *corev1.Pod) {
				p.Status.InitContainerStatuses[0].State = corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
				}
				p.Status.InitContainerStatuses[0].Ready = false
			},
			want: "sidecar cm-sidecar in pod os-platform/nats-0 terminated (exitCode=1",
		},
		{
			name: "running but failing its probe",
			mutate: func(p *corev1.Pod) {
				p.Status.InitContainerStatuses[0].Ready = false
			},
			want: "sidecar cm-sidecar in pod os-platform/nats-0 is not ready",
		},
		{
			name: "not started yet",
			mutate: func(p *corev1.Pod) {
				p.Status.InitContainerStatuses[0].Started = ptr(false)
			},
			want: "sidecar cm-sidecar in pod os-platform/nats-0 has not started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := natsPod()
			tt.mutate(pod)
			err := AssertPodReady(pod)
			if err == nil {
				t.Fatalf("unhealthy sidecar reported ready")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("got %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

// A plain init container still has to run to completion, so the sidecar
// handling must not weaken the ordinary case.
func TestAssertPodReadyStillWaitsForAPlainInitContainer(t *testing.T) {
	pod := natsPod()
	pod.Spec.InitContainers[0].RestartPolicy = nil
	pod.Status.InitContainerStatuses[0].RestartCount = 0

	err := AssertPodReady(pod)
	if err == nil {
		t.Fatal("running init container reported ready")
	}
	if want := "init container cm-sidecar is still running"; !strings.Contains(err.Error(), want) {
		t.Errorf("got %q, want it to contain %q", err, want)
	}
}
