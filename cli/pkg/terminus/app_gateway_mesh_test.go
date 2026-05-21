package terminus

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEgDataPlaneMeshReady_skipsTerminating(t *testing.T) {
	now := metav1.Now()
	c := fake.NewClientBuilder().WithScheme(meshTestScheme()).
		WithObjects(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "terminating", Namespace: "app-gateway",
					Labels: map[string]string{egDataPlaneGatewayLabel: "app-gateway"},
					DeletionTimestamp: &now,
					Finalizers:        []string{"kubernetes"},
				},
				Status: corev1.PodStatus{Phase: corev1.PodRunning},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ready", Namespace: "app-gateway",
					Labels: map[string]string{egDataPlaneGatewayLabel: "app-gateway"},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "linkerd-proxy", Ready: true},
						{Name: "envoy", Ready: true},
					},
				},
			},
		).Build()
	ok, reason, err := egDataPlaneMeshReady(context.Background(), c, "app-gateway", "app-gateway")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected ready, reason=%q", reason)
	}
}

func TestEgDataPlaneMeshReady_waitsForActivePod(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(meshTestScheme()).
		WithObjects(
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "starting", Namespace: "app-gateway",
					Labels: map[string]string{egDataPlaneGatewayLabel: "app-gateway"},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Name: "linkerd-proxy", Ready: false},
						{Name: "envoy", Ready: false},
					},
				},
			},
		).Build()
	ok, reason, err := egDataPlaneMeshReady(context.Background(), c, "app-gateway", "app-gateway")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected not ready")
	}
	if reason != "starting not ready" {
		t.Fatalf("reason: got %q", reason)
	}
}

func meshTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	return s
}
