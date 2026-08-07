package upgrade

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluateEdgeAcceptSuitePassed(t *testing.T) {
	ok := map[string]bool{
		conditionL4ProxyReady: true,
		"ZeroOesInventory":    true,
		"NoEntranceEGExtAuth": true,
	}
	if !evaluateEdgeAcceptSuitePassed(ok) {
		t.Fatal("expected pass without LinkerdReady")
	}
	bad := map[string]bool{
		conditionL4ProxyReady: true,
		"ZeroOesInventory":    true,
		"NoEntranceEGExtAuth": false,
	}
	if evaluateEdgeAcceptSuitePassed(bad) {
		t.Fatal("expected fail without M-EG")
	}
	if evaluateEdgeAcceptSuitePassed(nil) {
		t.Fatal("nil must fail")
	}
}

func TestIsEntranceEGPEPObject(t *testing.T) {
	if !isEntranceEGPEPObject("app-web-entrance-ext-auth", nil) {
		t.Fatal("ext-auth suffix")
	}
	if !isEntranceEGPEPObject("x", map[string]string{authKindLabel: authKindEntranceCookie}) {
		t.Fatal("cookie auth-kind")
	}
	if isEntranceEGPEPObject("shared-demo-jwt-authn", nil) {
		t.Fatal("jwt-authn must not match by suffix alone without auth-kind")
	}
}

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
	_, err = kube.CoreV1().Pods("user-space").Get(context.Background(), "app-1", metav1.GetOptions{})
	if err == nil {
		t.Fatal("business oes pod must be deleted")
	}
	if _, err := kube.CoreV1().Pods("os-network").Get(context.Background(), "l4", metav1.GetOptions{}); err != nil {
		t.Fatal("platform l4 pod must remain")
	}
}
