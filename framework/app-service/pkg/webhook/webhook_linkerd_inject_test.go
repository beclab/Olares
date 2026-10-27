package webhook

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
)

func TestAnnotatePodForLinkerdInjectWithMeshIn(t *testing.T) {
	if !mesh.ShouldInjectLinkerdProxy(true) {
		t.Fatal("ARCH S6: mesh-in must imply linkerd inject")
	}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "caller", Namespace: "app-ns"}}
	mesh.AnnotatePodForLinkerdInject(pod, mesh.ShouldInjectLinkerdProxy(true))
	if got := pod.Annotations[mesh.LinkerdInjectAnnotation]; got != mesh.LinkerdInjectEnabled {
		t.Fatalf("inject = %q, want enabled", got)
	}
}

func TestAnnotatePodForLinkerdInjectSkippedWithoutMeshIn(t *testing.T) {
	if mesh.ShouldInjectLinkerdProxy(false) {
		t.Fatal("no mesh-in must not inject linkerd")
	}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "plain"}}
	mesh.AnnotatePodForLinkerdInject(pod, mesh.ShouldInjectLinkerdProxy(false))
	if _, ok := pod.Annotations[mesh.LinkerdInjectAnnotation]; ok {
		t.Fatalf("unexpected inject annotation: %#v", pod.Annotations)
	}
}
