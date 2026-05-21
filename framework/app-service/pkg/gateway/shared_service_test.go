package gateway

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResolveSharedEntranceService_v2SharedSubchartNS(t *testing.T) {
	app := &appv1alpha1.Application{
		Spec: appv1alpha1.ApplicationSpec{
			Namespace: "ollamav2-brucedai",
			Name:      "ollamav2",
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "sharedentrances-ollama", Namespace: "ollamaserver-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80, Protocol: corev1.ProtocolTCP}},
		},
	}
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "ollamaserver-shared",
			Labels: map[string]string{NamespaceSharedLabel: "true"},
		},
	}
	c := fake.NewClientBuilder().WithObjects(svc, ns).Build()

	got, err := ResolveSharedEntranceService(context.Background(), c, app, "sharedentrances-ollama")
	if err != nil {
		t.Fatalf("ResolveSharedEntranceService: %v", err)
	}
	if got.Namespace != "ollamaserver-shared" || got.Name != "sharedentrances-ollama" {
		t.Fatalf("got %s/%s", got.Namespace, got.Name)
	}
}
