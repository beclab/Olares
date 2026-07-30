package mesh

import (
	"context"
	"errors"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func linkerdControlPlaneObjects() []runtime.Object {
	objs := make([]runtime.Object, 0, len(linkerdControlPlaneDeployments)+1)
	for _, name := range linkerdControlPlaneDeployments {
		objs = append(objs, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: linkerdNamespace},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
		})
	}
	return objs
}

func linkerdReadyClient() *fake.Clientset {
	objs := append(linkerdControlPlaneObjects(), &appsv1.Deployment{
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

func TestIsLinkerdLayer1ReadyFalseWhenPKIGuardianAbsent(t *testing.T) {
	client := fake.NewSimpleClientset(linkerdControlPlaneObjects()...)
	if IsLinkerdLayer1Ready(context.Background(), client) {
		t.Fatal("expected Linkerd not ready when linkerd-pki-guardian is missing")
	}
	if ShouldSkipOesForSharedCaller(context.Background(), client, true, false, false) {
		t.Fatal("must not skip Shared-caller oes without a ready PKI guardian")
	}
}

func TestIsLinkerdLayer1ReadyFalseWhenPKIGuardianHasNoReadyReplica(t *testing.T) {
	objs := append(linkerdControlPlaneObjects(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: linkerdPKIGuardianDeploy, Namespace: linkerdNamespace},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 0},
	})
	if IsLinkerdLayer1Ready(context.Background(), fake.NewSimpleClientset(objs...)) {
		t.Fatal("expected Linkerd not ready when linkerd-pki-guardian has no ready replica")
	}
}

func TestIsLinkerdLayer1ReadyFalseWhenPKIGuardianLookupFails(t *testing.T) {
	client := linkerdReadyClient()
	client.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.(k8stesting.GetAction).GetName() == linkerdPKIGuardianDeploy {
			return true, nil, errors.New("apiserver unavailable")
		}
		return false, nil, nil
	})
	if IsLinkerdLayer1Ready(context.Background(), client) {
		t.Fatal("expected Linkerd not ready when the guardian lookup fails")
	}
}

func TestShouldSkipEnvoySidecarNeverBlanketsOnLinkerdReady(t *testing.T) {
	// R1: Linkerd ready alone must not retire outbound oes (ADR-DEENVY-SCOPE-SHARED).
	if ShouldSkipEnvoySidecar(context.Background(), linkerdReadyClient()) {
		t.Fatal("ShouldSkipEnvoySidecar must stay false until L2-c blanket retire")
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

func TestEvaluateSkipOesForSharedCaller(t *testing.T) {
	cases := []struct {
		name                                           string
		meshIn, linkerd, provider, meshOut bool
		want                                           bool
	}{
		{"mesh-in no provider", true, true, false, false, true},
		{"mesh-in provider needs mesh-out", true, true, true, false, false},
		{"mesh-in provider with mesh-out", true, true, true, true, true},
		{"no mesh-in", false, true, false, false, false},
		{"mesh-in no linkerd", true, false, false, false, false},
		{"entrance-like mesh-in still skips", true, true, false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateSkipOesForSharedCaller(tc.meshIn, tc.linkerd, tc.provider, tc.meshOut)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestShouldSkipOesForSharedCallerUsesLinkerdReady(t *testing.T) {
	if !ShouldSkipOesForSharedCaller(context.Background(), linkerdReadyClient(), true, false, false) {
		t.Fatal("expected Shared-caller oes skip when mesh-in and Linkerd ready")
	}
	if ShouldSkipOesForSharedCaller(context.Background(), fake.NewSimpleClientset(), true, false, false) {
		t.Fatal("must not skip Shared-caller oes when Linkerd is not ready")
	}
}
