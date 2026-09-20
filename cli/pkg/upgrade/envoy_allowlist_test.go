package upgrade

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsBusinessOESContainer(t *testing.T) {
	biz := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "user-space-a", Labels: map[string]string{"app": "demo"}}}
	if !isBusinessOESContainer(biz, corev1.Container{Name: "olares-envoy-sidecar", Image: "beclab/envoy:v1"}) {
		t.Fatal("business oes must be flagged")
	}
	l4 := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "l4-bfl-proxy-x", Namespace: "os-network", Labels: map[string]string{"app": "l4-bfl-proxy"}}}
	if isBusinessOESContainer(l4, corev1.Container{Name: "envoy", Image: "beclab/envoy:v1"}) {
		t.Fatal("l4 platform Envoy must be allow-listed")
	}
	ss := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "system-server", Namespace: "user-system-a", Labels: map[string]string{"app": "systemserver"}}}
	if isBusinessOESContainer(ss, corev1.Container{Name: "proxy", Image: "beclab/envoy:v1"}) {
		t.Fatal("system-server proxy must be allow-listed")
	}
}
