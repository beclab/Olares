package mesh

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

func TestShouldInjectLinkerdProxy(t *testing.T) {
	if ShouldInjectLinkerdProxy(false) {
		t.Fatal("false mesh-in must not inject linkerd")
	}
	if !ShouldInjectLinkerdProxy(true) {
		t.Fatal("true mesh-in must inject linkerd (ARCH S6)")
	}
}

func TestAnnotatePodForLinkerdInject(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "p"}}
	AnnotatePodForLinkerdInject(pod, true)
	if got := pod.Annotations[LinkerdInjectAnnotation]; got != LinkerdInjectEnabled {
		t.Fatalf("inject = %q, want %q", got, LinkerdInjectEnabled)
	}
	AnnotatePodForLinkerdInject(nil, true) // must not panic
}

func TestEnsureCallerNamespaceMeshAccess(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "caller-ns",
			Annotations: map[string]string{
				LinkerdInjectAnnotation: LinkerdInjectEnabled,
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "caller-ns", true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	got := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "caller-ns"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Labels[security.NamespaceInClusterCallerLabel] != "true" {
		t.Fatalf("label = %#v", got.Labels)
	}
	if _, ok := got.Annotations[LinkerdInjectAnnotation]; ok {
		t.Fatalf("enable must clear NS inject, got %#v", got.Annotations)
	}
	meshNP := &networkingv1.NetworkPolicy{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: security.MeshControlPlaneNamespace, Name: security.AppGatewayMeshNPName,
	}, meshNP); err != nil {
		t.Fatalf("expected app-gateway-mesh-np on enable: %v", err)
	}

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "caller-ns", false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	_ = c.Get(context.Background(), types.NamespacedName{Name: "caller-ns"}, got)
	if _, ok := got.Labels[security.NamespaceInClusterCallerLabel]; ok {
		t.Fatalf("label should be removed: %#v", got.Labels)
	}
}

func TestEnsureCallerNamespaceMeshAccess_SharedNSSetsCallerLabelPreservesInject(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ollama-shared",
			Labels: map[string]string{
				security.NamespaceSharedLabel: "true",
			},
			Annotations: map[string]string{
				LinkerdInjectAnnotation: LinkerdInjectEnabled,
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ns).Build()

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "ollama-shared", true); err != nil {
		t.Fatalf("shared ns enable: %v", err)
	}
	got := &corev1.Namespace{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "ollama-shared"}, got); err != nil {
		t.Fatal(err)
	}
	if got.Labels[security.NamespaceInClusterCallerLabel] != "true" {
		t.Fatalf("shared caller ns must get in-cluster-caller, got %#v", got.Labels)
	}
	if got.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("shared inject must be preserved, got %#v", got.Annotations)
	}

	if err := EnsureCallerNamespaceMeshAccess(context.Background(), c, "ollama-shared", false); err != nil {
		t.Fatalf("shared ns disable: %v", err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Name: "ollama-shared"}, got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Labels[security.NamespaceInClusterCallerLabel]; ok {
		t.Fatalf("in-cluster-caller must be cleared on disable, got %#v", got.Labels)
	}
	if got.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("shared inject must still be preserved after disable, got %#v", got.Annotations)
	}
}

func TestEnsureCallerWorkloadLinkerdInject(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "app-ns",
			Labels:    map[string]string{constants.ApplicationNameLabel: "web"},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep).Build()

	if err := EnsureCallerWorkloadLinkerdInject(context.Background(), c, "app-ns", "web", true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	got := &appsv1.Deployment{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "app-ns", Name: "web"}, got); err != nil {
		t.Fatal(err)
	}
	ann := got.Spec.Template.Annotations
	if ann[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatalf("inject = %#v", ann)
	}
	if ann[CallerLinkerdInjectManagedAnnotation] != CallerLinkerdInjectManagedValue {
		t.Fatalf("managed = %#v", ann)
	}

	if err := EnsureCallerWorkloadLinkerdInject(context.Background(), c, "app-ns", "web", false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: "app-ns", Name: "web"}, got)
	if _, ok := got.Spec.Template.Annotations[LinkerdInjectAnnotation]; ok {
		t.Fatalf("inject should be cleared: %#v", got.Spec.Template.Annotations)
	}
}

func TestEnsureCallerWorkloadLinkerdInject_SkipsUnmanaged(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "app-ns",
			Labels:    map[string]string{constants.ApplicationNameLabel: "web"},
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{LinkerdInjectAnnotation: LinkerdInjectEnabled},
				},
			},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dep).Build()
	if err := EnsureCallerWorkloadLinkerdInject(context.Background(), c, "app-ns", "web", false); err != nil {
		t.Fatalf("disable unmanaged: %v", err)
	}
	got := &appsv1.Deployment{}
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: "app-ns", Name: "web"}, got)
	if got.Spec.Template.Annotations[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
		t.Fatal("unmanaged inject must be preserved")
	}
}

func TestNamespaceHasDecideTrueCaller(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appv1alpha1.AddToScheme(scheme)
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{Name: "user-space-u-web"},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "web",
			Namespace: "app-ns",
			Settings:  map[string]string{annotDecide: "true"},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app).Build()
	declares := func(s map[string]string) bool {
		return s != nil && s[annotDecide] == "true"
	}
	ok, err := NamespaceHasDecideTrueCaller(context.Background(), c, "app-ns", declares)
	if err != nil || !ok {
		t.Fatalf("want true, got %v %v", ok, err)
	}
	ok, err = NamespaceHasDecideTrueCaller(context.Background(), c, "other", declares)
	if err != nil || ok {
		t.Fatalf("want false for other ns, got %v %v", ok, err)
	}
}
