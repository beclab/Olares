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

func TestShouldInjectEnvoySidecarKeepsOutboundWhenLinkerdReady(t *testing.T) {
	wh := &Webhook{kubeClient: linkerdReadyKube()}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "worker"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}
	// Outbound-only path (injectPolicy=false): R1 keeps oes; Shared-caller skip
	// is gated separately by ShouldSkipOesForSharedCaller in CreatePatch.
	if !wh.shouldInjectEnvoySidecar(context.Background(), false, appCfg, pod) {
		t.Fatal("outbound oes must remain when Linkerd ready without Shared-caller gate")
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

func TestShouldInjectEnvoySidecarEntranceKeptWithoutMeshIn(t *testing.T) {
	wh := &Webhook{kubeClient: linkerdReadyKube()}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "entrance"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}
	if !wh.shouldInjectEnvoySidecar(context.Background(), true, appCfg, pod) {
		t.Fatal("entrance without mesh-in must still inject oes when injectPolicy is true")
	}
	// Shared-caller gate alone decides skip; without mesh-in it must not fire.
	if mesh.ShouldSkipOesForSharedCaller(context.Background(), linkerdReadyKube(), false, false, false) {
		t.Fatal("Shared-caller skip must not fire without mesh-in")
	}
}

func TestShouldSkipOesForSharedCallerEntranceWithMeshIn(t *testing.T) {
	kube := linkerdReadyKube()
	if !mesh.ShouldSkipOesForSharedCaller(context.Background(), kube, true, false, false) {
		t.Fatal("entrance Shared caller with mesh-in and no provider must skip oes")
	}
	if mesh.ShouldSkipOesForSharedCaller(context.Background(), kube, true, true, false) {
		t.Fatal("Shared caller with provider and no mesh-out must keep oes")
	}
	if !mesh.ShouldSkipOesForSharedCaller(context.Background(), kube, true, true, true) {
		t.Fatal("Shared caller with provider and mesh-out must skip oes")
	}
}
