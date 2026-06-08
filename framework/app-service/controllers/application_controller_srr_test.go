package controllers

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/routecontrol"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcileSharedRouteRegistry_WritesRouteObjects(t *testing.T) {
	t.Setenv("OLARES_PLATFORM_DOMAIN", "olares.com")

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("srr scheme: %v", err)
	}
	httpRouteGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
	httpRouteListGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRouteList"}
	scheme.AddKnownTypeWithName(httpRouteGVK, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(httpRouteListGVK, &unstructured.UnstructuredList{})

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama",
			UID:  types.UID("app-uid"),
			Labels: map[string]string{
				constants.AppApiVersionLabel: constants.AppVersionV3,
			},
			Annotations: map[string]string{
				gateway.AnnotationRouteMode: gateway.AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			Appid:     "a5be2268",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "ollama"},
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP},
			},
		},
	}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).
		WithObjects(app, svc).
		Build()

	r := &ApplicationReconciler{Client: c, Scheme: scheme}
	if err := r.reconcileSharedRouteRegistry(context.Background(), app); err != nil {
		t.Fatalf("reconcileSharedRouteRegistry: %v", err)
	}

	// Per-entrance SRR names derive from GetAppID(Spec.Name); the divergent
	// Spec.Appid above must be ignored, so the object lands here and survives prune.
	srrName := gateway.ResourceNameForEntrance(gateway.EntranceAppID(app), "api")
	srr := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: srrName}, srr); err != nil {
		t.Fatalf("get SRR: %v", err)
	}
	if len(srr.Spec.HostPatterns) != 1 || !gateway.IsLogicalHostPattern(srr.Spec.HostPatterns[0]) {
		t.Fatalf("unexpected hostPatterns: %v", srr.Spec.HostPatterns)
	}

	hr := &unstructured.Unstructured{}
	hr.SetGroupVersionKind(httpRouteGVK)
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: srrName}, hr); err != nil {
		t.Fatalf("get HTTPRoute: %v", err)
	}
	spec := hr.Object["spec"].(map[string]any)
	hosts := spec["hostnames"].([]any)
	if len(hosts) != 1 || hosts[0].(string) != "*.olares.com" {
		t.Fatalf("HTTPRoute hostnames: %v", hosts)
	}

	np := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: routecontrol.NetworkPolicyName}, np); err != nil {
		t.Fatalf("get NetworkPolicy: %v", err)
	}

	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: srrName}, got); err != nil {
		t.Fatalf("get SRR status: %v", err)
	}
	if got.Status.HTTPRouteName != srrName {
		t.Fatalf("status.httpRouteName=%q", got.Status.HTTPRouteName)
	}
	if len(got.Status.Conditions) == 0 || got.Status.Conditions[0].Reason != routecontrol.ReasonReconciled {
		t.Fatalf("status conditions: %+v", got.Status.Conditions)
	}
}

func TestReconcileSharedRouteRegistry_AppidIgnoredForNaming(t *testing.T) {
	t.Setenv("OLARES_PLATFORM_DOMAIN", "olares.com")

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("srr scheme: %v", err)
	}
	httpRouteGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
	scheme.AddKnownTypeWithName(httpRouteGVK, &unstructured.Unstructured{})

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama",
			UID:  types.UID("app-uid"),
			Labels: map[string]string{
				constants.AppApiVersionLabel: constants.AppVersionV3,
			},
			Annotations: map[string]string{
				gateway.AnnotationRouteMode: gateway.AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollama",
			Namespace: "ollama-shared",
			Appid:     "deadbeef",
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "api", Host: "ollama", Port: 11434},
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "ollama", Namespace: "ollama-shared"},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 11434, Protocol: corev1.ProtocolTCP},
			},
		},
	}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).
		WithObjects(app, svc).
		Build()

	r := &ApplicationReconciler{Client: c, Scheme: scheme}
	if err := r.reconcileSharedRouteRegistry(context.Background(), app); err != nil {
		t.Fatalf("reconcileSharedRouteRegistry: %v", err)
	}

	nameDerived := gateway.ResourceNameForEntrance(gateway.EntranceAppID(app), "api")
	specAppidName := gateway.ResourceNameForEntrance(app.Spec.Appid, "api")
	if nameDerived == specAppidName {
		t.Fatalf("test setup invalid: Spec.Appid must diverge from name-derived appid")
	}

	got := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: nameDerived}, got); err != nil {
		t.Fatalf("SRR must exist under name-derived appid and survive prune: %v", err)
	}
	if len(got.Status.Conditions) == 0 || got.Status.Conditions[0].Reason != routecontrol.ReasonReconciled {
		t.Fatalf("status conditions: %+v", got.Status.Conditions)
	}

	stray := &srrv1alpha1.SharedRouteRegistry{}
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollama-shared", Name: specAppidName}, stray)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("Spec.Appid-derived SRR must not exist: err=%v", err)
	}
}

func TestReconcileSharedRouteRegistry_v2ClusterScopedSharedNS(t *testing.T) {
	t.Setenv("OLARES_PLATFORM_DOMAIN", "olares.com")

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("clientgo scheme: %v", err)
	}
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("app scheme: %v", err)
	}
	if err := srrv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("srr scheme: %v", err)
	}
	httpRouteGVK := schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute"}
	scheme.AddKnownTypeWithName(httpRouteGVK, &unstructured.Unstructured{})

	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollamav2-alice-ollamav2",
			UID:  types.UID("app-uid-v2"),
			Annotations: map[string]string{
				gateway.AnnotationRouteMode: gateway.AnnotationRouteModeGateway,
			},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "ollamav2",
			Namespace: "ollamav2-alice",
			Appid:     "a5be2268",
			Settings:  map[string]string{"clusterScoped": "true"},
			SharedEntrances: []appv1alpha1.Entrance{
				{Name: "ollamav2", Host: "sharedentrances-ollama", Port: 80},
			},
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
			Labels: map[string]string{gateway.NamespaceSharedLabel: "true"},
		},
	}
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&srrv1alpha1.SharedRouteRegistry{}).
		WithObjects(app, svc, ns).
		Build()

	r := &ApplicationReconciler{Client: c, Scheme: scheme}
	if err := r.reconcileSharedRouteRegistry(context.Background(), app); err != nil {
		t.Fatalf("reconcileSharedRouteRegistry: %v", err)
	}

	srrName := gateway.ResourceNameForEntrance("a5be2268", "ollamav2")
	srr := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "ollamav2-alice", Name: srrName}, srr); err != nil {
		t.Fatalf("get SRR: %v", err)
	}
	if srr.Spec.Upstream.ServiceNamespace != "ollamaserver-shared" {
		t.Fatalf("upstream NS: got %q", srr.Spec.Upstream.ServiceNamespace)
	}

	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamaserver-shared", Name: routecontrol.NetworkPolicyName,
	}, &networkingv1.NetworkPolicy{}); err != nil {
		t.Fatalf("get NetworkPolicy in upstream NS: %v", err)
	}
	err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "ollamav2-alice", Name: routecontrol.NetworkPolicyName,
	}, &networkingv1.NetworkPolicy{})
	if !apierrors.IsNotFound(err) {
		t.Fatalf("NP should not be in SRR namespace: err=%v", err)
	}
}
