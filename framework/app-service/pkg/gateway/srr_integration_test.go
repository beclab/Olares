package gateway

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func newReconcileFixture(t *testing.T) (client.Client, *appv1alpha1.Application, *corev1.Service) {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add app.bytetrade.io scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add gateway scheme: %v", err)
	}

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama-shared-ollama",
			UID:  types.UID("ollama-uid"),
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434, URL: "abc.shared.example.com"},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()
	return c, app, svc
}

func TestReconcile_CreateThenUpdate(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	ctx := context.Background()
	spec, err := BuildSpec(app, svc)
	if err != nil {
		t.Fatalf("BuildSpec: %v", err)
	}
	created, err := Reconcile(ctx, c, app, spec)
	if err != nil {
		t.Fatalf("Reconcile create: %v", err)
	}
	if created.Name != "shared-ollama" || created.Namespace != "ollama-shared" {
		t.Fatalf("unexpected SRR identity: %+v", created.ObjectMeta)
	}
	if len(created.OwnerReferences) != 1 || created.OwnerReferences[0].UID != app.UID {
		t.Fatalf("missing/incorrect ownerReference: %+v", created.OwnerReferences)
	}

	app.Spec.SharedEntrances[0].URL = "DEF.shared.example.com"
	spec2, err := BuildSpec(app, svc)
	if err != nil {
		t.Fatalf("BuildSpec 2: %v", err)
	}
	if _, err := Reconcile(ctx, c, app, spec2); err != nil {
		t.Fatalf("Reconcile update: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama"}, got); err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if len(got.Spec.HostPatterns) != 1 || got.Spec.HostPatterns[0] != "def.shared.example.com" {
		t.Fatalf("hostPatterns not updated: %v", got.Spec.HostPatterns)
	}
}

func TestDelete_IdempotentAndCleansUp(t *testing.T) {
	c, app, svc := newReconcileFixture(t)
	ctx := context.Background()
	spec, _ := BuildSpec(app, svc)
	if _, err := Reconcile(ctx, c, app, spec); err != nil {
		t.Fatalf("seed Reconcile: %v", err)
	}

	if err := Delete(ctx, c, app); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	got := &srrv1alpha1.SharedRouteRegistry{}
	err := c.Get(ctx, types.NamespacedName{Namespace: "ollama-shared", Name: "shared-ollama"}, got)
	if err == nil {
		t.Fatalf("SRR still present after Delete")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected error after Delete: %v", err)
	}

	if err := Delete(ctx, c, app); err != nil {
		t.Fatalf("second Delete should be idempotent: %v", err)
	}
}
