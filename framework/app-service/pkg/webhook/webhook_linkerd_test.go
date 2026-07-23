package webhook

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func linkerdReadyKube() *fake.Clientset {
	ns := mesh.LinkerdNamespace
	objs := []runtime.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "linkerd-destination", Namespace: ns}, Status: appsv1.DeploymentStatus{ReadyReplicas: 1}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "linkerd-identity", Namespace: ns}, Status: appsv1.DeploymentStatus{ReadyReplicas: 1}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "linkerd-proxy-injector", Namespace: ns}, Status: appsv1.DeploymentStatus{ReadyReplicas: 1}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "linkerd-pki-guardian", Namespace: ns}, Status: appsv1.DeploymentStatus{ReadyReplicas: 1}},
	}
	return fake.NewSimpleClientset(objs...)
}

func TestShouldInjectEnvoySidecarSkipsWhenLinkerdReady(t *testing.T) {
	wh := &Webhook{kubeClient: linkerdReadyKube()}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "worker"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}
	if wh.shouldInjectEnvoySidecar(context.Background(), false, appCfg, pod) {
		t.Fatal("olares-envoy-sidecar must be skipped when Linkerd Layer 1 is ready")
	}
}

func TestShouldInjectEnvoySidecarKeepsEntranceBeforeLinkerdCutover(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset()}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "entrance"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}
	if !wh.shouldInjectEnvoySidecar(context.Background(), true, appCfg, pod) {
		t.Fatal("expected entrance envoy injection before Linkerd steady-state cutover")
	}
}

func TestShouldSkipInboundEntranceSidecarWithoutExtAuth(t *testing.T) {
	if mesh.ShouldSkipInboundEntranceSidecar(context.Background(), linkerdReadyKube(), "demo-user", "app-demo-web") {
		t.Fatal("entrance sidecar must not skip without extAuth policy")
	}
}
