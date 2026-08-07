package mesh

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvaluateL4ProxyReady(t *testing.T) {
	cases := []struct {
		name                  string
		verify, cookie, probe bool
		want                  bool
	}{
		{"all true", true, true, true, true},
		{"no verify", false, true, true, false},
		{"no cookie", true, false, true, false},
		{"no probe", true, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateL4ProxyReady(tc.verify, tc.cookie, tc.probe)
			if got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestIsL4ProxyDeploymentReady(t *testing.T) {
	if IsL4ProxyDeploymentReady(context.Background(), fake.NewSimpleClientset()) {
		t.Fatal("missing deployment must not be ready")
	}
	kube := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: L4ProxyDeploymentName, Namespace: L4ProxyNamespace},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
	})
	if !IsL4ProxyDeploymentReady(context.Background(), kube) {
		t.Fatal("expected l4 Deployment Ready")
	}
}

func TestIsL4EdgePEPReadyViaL4Deployment(t *testing.T) {
	kube := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: L4ProxyDeploymentName, Namespace: L4ProxyNamespace},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: 1},
	})
	if !IsL4EdgePEPReady(context.Background(), kube) {
		t.Fatal("live l4 Deployment Ready must unlock edge PEP without SteadyGate")
	}
}
