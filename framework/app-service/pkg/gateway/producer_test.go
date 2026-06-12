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
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
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

func TestSharedRouteProducerReconcilerBuildsApplicationSRRs(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	defer cluster.ResetPlatformDomainForTest()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationRouteMode: AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "demo-user",
			Appid:     "demo1234",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "web-svc", Port: 8080},
				{Name: "api", Host: "api-svc", Port: 9090},
			},
		},
	}
	webSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	apiSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "api-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 9090, Protocol: corev1.ProtocolTCP}}},
	}

	c := fake.NewClientBuilder().WithScheme(producerTestScheme(t)).WithObjects(app, webSvc, apiSvc).Build()
	r := &SharedRouteProducerReconciler{Client: c}
	if err := r.reconcileApp(context.Background(), app); err != nil {
		t.Fatalf("reconcileApp: %v", err)
	}

	for i, entrance := range app.Spec.Entrances {
		name := ResourceNameForEntranceApp(app.Spec.Appid, entrance.Name)
		got := &srrv1alpha1.SharedRouteRegistry{}
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: app.Spec.Namespace, Name: name}, got); err != nil {
			t.Fatalf("get application SRR %q: %v", name, err)
		}
		if got.Spec.EntranceClass != srrv1alpha1.EntranceClassApplication {
			t.Fatalf("%s entranceClass = %q, want %q", name, got.Spec.EntranceClass, srrv1alpha1.EntranceClassApplication)
		}
		wantHost := appv1alpha1.EntranceID(app.Spec.Appid, i, len(app.Spec.Entrances)) + ".*.olares.com"
		if len(got.Spec.HostPatterns) != 1 || got.Spec.HostPatterns[0] != wantHost {
			t.Fatalf("%s hostPatterns = %v, want %q", name, got.Spec.HostPatterns, wantHost)
		}
	}
}

func TestSharedRouteProducerReconcilerSingleApplicationEntranceUsesBareAppID(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	defer cluster.ResetPlatformDomainForTest()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationRouteMode: AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "demo-user",
			Appid:     "demo1234",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "web-svc", Port: 8080},
			},
		},
	}
	webSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "web-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}

	c := fake.NewClientBuilder().WithScheme(producerTestScheme(t)).WithObjects(app, webSvc).Build()
	r := &SharedRouteProducerReconciler{Client: c}
	if err := r.reconcileApp(context.Background(), app); err != nil {
		t.Fatalf("reconcileApp: %v", err)
	}

	name := ResourceNameForEntranceApp(app.Spec.Appid, app.Spec.Entrances[0].Name)
	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: app.Spec.Namespace, Name: name}, got); err != nil {
		t.Fatalf("get application SRR: %v", err)
	}
	wantHost := app.Spec.Appid + ".*.olares.com"
	if len(got.Spec.HostPatterns) != 1 || got.Spec.HostPatterns[0] != wantHost {
		t.Fatalf("single entrance hostPatterns = %v, want %q", got.Spec.HostPatterns, wantHost)
	}
}

func TestSharedRouteProducerReconcilerNotOptedInDeletesOwnedSRRs(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	defer cluster.ResetPlatformDomainForTest()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "demo-user",
			Appid:     "demo1234",
			Entrances: []appv1alpha1.Entrance{
				{Name: "web", Host: "web-svc", Port: 8080},
			},
		},
	}
	targetAppSRR := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-demo1234-web",
			Namespace: "demo-user",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "demo",
			},
		},
	}
	targetSharedSRR := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-demo1234-web",
			Namespace: "demo-user",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "demo",
			},
		},
	}
	otherAppSRR := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-demo9999-web",
			Namespace: "demo-user",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "other",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(producerTestScheme(t)).
		WithObjects(app, targetAppSRR, targetSharedSRR, otherAppSRR).Build()
	r := &SharedRouteProducerReconciler{Client: c}
	if err := r.reconcileApp(context.Background(), app); err != nil {
		t.Fatalf("reconcileApp: %v", err)
	}

	for _, name := range []string{"app-demo1234-web", "shared-demo1234-web"} {
		err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: name}, &srrv1alpha1.SharedRouteRegistry{})
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected %s to be deleted, get err=%v", name, err)
		}
	}

	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: "app-demo9999-web"},
		&srrv1alpha1.SharedRouteRegistry{}); err != nil {
		t.Fatalf("unrelated SRR must be kept: %v", err)
	}
}

func TestSharedRouteProducerReconcilerPrunesSharedAndApplicationSeparately(t *testing.T) {
	cluster.PrimePlatformDomainForTest("olares.com")
	defer cluster.ResetPlatformDomainForTest()

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationRouteMode: AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "demo",
			Namespace: "demo-user",
			Appid:     "demo1234",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "shared-web", Host: "shared-svc", Port: 8080},
			},
			Entrances: []appv1alpha1.Entrance{
				{Name: "app-web", Host: "app-svc", Port: 8081},
			},
		},
	}
	sharedSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "shared-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}
	appSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "app-svc", Namespace: "demo-user"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8081, Protocol: corev1.ProtocolTCP}}},
	}
	staleShared := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shared-demo1234-old",
			Namespace: "demo-user",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "demo",
			},
		},
	}
	staleApp := &srrv1alpha1.SharedRouteRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-demo1234-old",
			Namespace: "demo-user",
			Labels: map[string]string{
				"app.kubernetes.io/instance": "demo",
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(producerTestScheme(t)).
		WithObjects(app, sharedSvc, appSvc, staleShared, staleApp).Build()
	r := &SharedRouteProducerReconciler{Client: c}
	if err := r.reconcileApp(context.Background(), app); err != nil {
		t.Fatalf("reconcileApp: %v", err)
	}

	sharedName := ResourceNameForEntrance(app.Spec.Appid, app.Spec.SharedEntrances[0].Name)
	appName := ResourceNameForEntranceApp(app.Spec.Appid, app.Spec.Entrances[0].Name)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: sharedName},
		&srrv1alpha1.SharedRouteRegistry{}); err != nil {
		t.Fatalf("shared SRR should exist: %v", err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: appName},
		&srrv1alpha1.SharedRouteRegistry{}); err != nil {
		t.Fatalf("application SRR should exist: %v", err)
	}

	for _, name := range []string{"shared-demo1234-old", "app-demo1234-old"} {
		err := c.Get(context.Background(), types.NamespacedName{Namespace: "demo-user", Name: name},
			&srrv1alpha1.SharedRouteRegistry{})
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected stale SRR %s to be pruned, get err=%v", name, err)
		}
	}
}
