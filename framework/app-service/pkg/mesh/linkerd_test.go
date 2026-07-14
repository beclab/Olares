package mesh

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func linkerdReadyClient() *fake.Clientset {
	objs := make([]runtime.Object, 0, 4)
	for _, name := range linkerdControlPlaneDeployments {
		objs = append(objs, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: linkerdNamespace},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
		})
	}
	objs = append(objs, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: linkerdPKIGuardianDeploy, Namespace: linkerdNamespace},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
	})
	return fake.NewSimpleClientset(objs...)
}

func TestIsLinkerdLayer1ReadyFalseWithoutDeployments(t *testing.T) {
	if IsLinkerdLayer1Ready(context.Background(), fake.NewSimpleClientset()) {
		t.Fatal("expected Linkerd not ready without control plane deployments")
	}
}

func TestIsLinkerdLayer1ReadyTrueWhenControlPlaneReady(t *testing.T) {
	if !IsLinkerdLayer1Ready(context.Background(), linkerdReadyClient()) {
		t.Fatal("expected Linkerd ready when control plane deployments are available")
	}
}

func TestShouldSkipEnvoySidecarUsesLinkerdReady(t *testing.T) {
	if !ShouldSkipEnvoySidecar(context.Background(), linkerdReadyClient()) {
		t.Fatal("expected envoy sidecar skip when Linkerd Layer 1 is ready")
	}
}

func TestShouldSkipInboundEntranceSidecarRequiresExtAuth(t *testing.T) {
	if ShouldSkipInboundEntranceSidecar(context.Background(), linkerdReadyClient(), "demo-user", "app-demo-web") {
		t.Fatal("must not skip entrance sidecar without extAuth SecurityPolicy")
	}
}

func TestEntranceExtAuthPolicyName(t *testing.T) {
	if got := EntranceExtAuthPolicyName("app-demo-web"); got != "app-demo-web-entrance-ext-auth" {
		t.Fatalf("policy name = %q", got)
	}
}

func TestEvaluateSkipOes(t *testing.T) {
	cases := []struct {
		name                                 string
		linkerd, extAuth, provider, egress bool
		want                                 bool
	}{
		{"all ready no provider", true, true, false, false, true},
		{"provider needs egress", true, true, true, false, false},
		{"provider with egress", true, true, true, true, true},
		{"no linkerd", false, true, false, false, false},
		{"no extAuth", true, false, false, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateSkipOes(tc.linkerd, tc.extAuth, tc.provider, tc.egress)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}
