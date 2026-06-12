package gateway

import (
	"context"
	"testing"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func producerTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add application scheme: %v", err)
	}
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("add srr scheme: %v", err)
	}
	return s
}

func TestSharedRouteProducerReconcilerSetsEntranceClassShared(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	defer cluster.ResetPlatformDomainForTest()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Labels: map[string]string{
				constants.AppApiVersionLabel: constants.AppVersionV3,
				constants.AppSharedLabel:     constants.AppSharedTrue,
			},
			Annotations: map[string]string{
				AnnotationRouteMode: AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "demo-shared",
			Appid:     "demo1234",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "demo-svc", Port: 8080},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-svc", Namespace: "demo-shared"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}

	c := fake.NewClientBuilder().WithScheme(producerTestScheme(t)).WithObjects(app, svc).Build()
	r := &SharedRouteProducerReconciler{Client: c}
	if err := r.reconcileApp(context.Background(), app); err != nil {
		t.Fatalf("reconcileApp: %v", err)
	}

	name := ResourceNameForEntrance(app.Spec.Appid, app.Spec.SharedEntrances[0].Name)
	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: app.Spec.Namespace, Name: name}, got); err != nil {
		t.Fatalf("get SRR: %v", err)
	}
	if got.Spec.EntranceClass != srrv1alpha1.EntranceClassShared {
		t.Fatalf("srr.spec.entranceClass = %q, want %q", got.Spec.EntranceClass, srrv1alpha1.EntranceClassShared)
	}
}
