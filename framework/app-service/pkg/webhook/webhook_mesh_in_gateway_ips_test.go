package webhook

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLookupMeshInGatewayIPsUsesOsGateway(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: meshinagent.GatewayDataServiceName, Namespace: security.AppGatewayNamespace},
			Spec:       corev1.ServiceSpec{ClusterIP: "10.233.5.10"},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: meshinagent.GatewayDataServiceName, Namespace: "app-gateway"},
			Spec:       corev1.ServiceSpec{ClusterIP: "10.233.9.9"},
		},
	)}
	got := wh.lookupMeshInGatewayIPs(context.Background())
	if got != "10.233.5.10" {
		t.Fatalf("got %q, want os-gateway ClusterIP", got)
	}
}

func TestLookupMeshInGatewayIPsIgnoresLegacyNamespace(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset(
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: meshinagent.GatewayDataServiceName, Namespace: "app-gateway"},
			Spec:       corev1.ServiceSpec{ClusterIP: "10.233.9.9"},
		},
	)}
	got := wh.lookupMeshInGatewayIPs(context.Background())
	if got != "" {
		t.Fatalf("got %q, want empty when os-gateway Service is absent", got)
	}
}

func TestLookupMeshInGatewayIPsEmptyWhenMissing(t *testing.T) {
	wh := &Webhook{kubeClient: fake.NewSimpleClientset()}
	if got := wh.lookupMeshInGatewayIPs(context.Background()); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
