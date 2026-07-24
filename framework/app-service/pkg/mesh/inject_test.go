package mesh

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

func TestShouldInjectLinkerdProxy(t *testing.T) {
	if ShouldInjectLinkerdProxy(false) {
		t.Fatal("false mesh-in must not inject linkerd")
	}
	if !ShouldInjectLinkerdProxy(true) {
		t.Fatal("true mesh-in must inject linkerd (ARCH S6)")
	}
}

func TestAnnotatePodForLinkerdInject(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}
	AnnotatePodForLinkerdInject(pod, true)
	if got := pod.Annotations[LinkerdInjectAnnotation]; got != LinkerdInjectEnabled {
		t.Fatalf("inject = %q, want %q", got, LinkerdInjectEnabled)
	}
	AnnotatePodForLinkerdInject(nil, true) // must not panic
}

func TestEnsureCallerNamespaceMeshAccess(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "caller-ns"}}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "caller-ns", true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	got := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "caller-ns"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Labels[security.NamespaceInClusterCallerLabel] != "true" {
		t.Fatalf("label = %#v", got.Labels)
	}
	if got.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("inject = %#v", got.Annotations)
	}

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "caller-ns", false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	_ = c.Get(context.Background(), types.NamespacedName{Name: "caller-ns"}, got)
	if _, ok := got.Labels[security.NamespaceInClusterCallerLabel]; ok {
		t.Fatalf("label should be removed: %#v", got.Labels)
	}
	if _, ok := got.Annotations[LinkerdInjectAnnotation]; ok {
		t.Fatalf("inject annotation should be removed: %#v", got.Annotations)
	}
}
