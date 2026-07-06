package webhook

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldInjectEnvoySidecarSkipsWhenLinkerdMeshEnabled(t *testing.T) {
	mesh.PrimeLinkerdMeshEnabledForTest(true)
	t.Cleanup(mesh.ResetLinkerdMeshEnabledForTest)

	wh := &Webhook{}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "entrance"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}

	if wh.shouldInjectEnvoySidecar(context.Background(), true, appCfg, pod) {
		t.Fatal("olares-envoy-sidecar must be skipped when Linkerd mesh is enabled")
	}
}

func TestShouldInjectEnvoySidecarKeepsEntranceBeforeLinkerdCutover(t *testing.T) {
	mesh.ResetLinkerdMeshEnabledForTest()
	t.Cleanup(mesh.ResetLinkerdMeshEnabledForTest)

	wh := &Webhook{}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "entrance"}}
	appCfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "user"}

	if !wh.shouldInjectEnvoySidecar(context.Background(), true, appCfg, pod) {
		t.Fatal("expected entrance envoy injection before Linkerd steady-state cutover")
	}
}
