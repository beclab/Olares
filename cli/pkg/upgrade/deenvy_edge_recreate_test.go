package upgrade

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRecreatePodsWithBusinessOES(t *testing.T) {
	kube := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "app-1", Namespace: "user-space"},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "olares-envoy-sidecar", Image: "beclab/envoy:v1"}},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "l4", Namespace: "os-network", Labels: map[string]string{"app": "l4-bfl-proxy"}},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "envoy", Image: "beclab/envoy:v1"}},
			},
		},
	)
	n, err := recreatePodsWithBusinessOES(context.Background(), kube)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("deleted=%d want 1", n)
	}
	if _, err := kube.CoreV1().Pods("user-space").Get(context.Background(), "app-1", metav1.GetOptions{}); err == nil {
		t.Fatal("business oes pod must be deleted")
	}
	if _, err := kube.CoreV1().Pods("os-network").Get(context.Background(), "l4", metav1.GetOptions{}); err != nil {
		t.Fatal("platform l4 pod must remain")
	}
}

func TestWaitL4ProxyReady(t *testing.T) {
	kube := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: l4ProxyDeploymentName, Namespace: l4ProxyNamespace},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
		},
	)
	if !waitL4ProxyReady(context.Background(), kube, time.Second) {
		t.Fatal("expected Ready")
	}
	empty := fake.NewSimpleClientset()
	if waitL4ProxyReady(context.Background(), empty, 50*time.Millisecond) {
		t.Fatal("missing deployment must time out")
	}
}

func TestDeenvyEdgeUpgradeTasksUpgradeOnly(t *testing.T) {
	tasks := deenvyEdgeUpgradeTasks()
	if len(tasks) != 1 || tasks[0].GetName() != "DeenvyEdgeBestEffortRecreate" {
		t.Fatalf("got %#v", tasks)
	}
}
