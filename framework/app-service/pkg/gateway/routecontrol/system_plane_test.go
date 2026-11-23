package routecontrol

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

func TestSystemPlaneRouteReconcilerCreatesSRRs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-alice"},
		Data:       map[string]string{"zone": "alice.olares.com"},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	r := &SystemPlaneRouteReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	for _, name := range []string{"sys-auth", "sys-desktop", "sys-wizard", "sys-profile"} {
		srr := &srrv1alpha1.SharedRouteRegistry{}
		if err := c.Get(context.Background(), types.NamespacedName{Namespace: "user-space-alice", Name: name}, srr); err != nil {
			t.Fatalf("get %s: %v", name, err)
		}
		if srr.Spec.EntranceClass != srrv1alpha1.EntranceClassApplication {
			t.Fatalf("%s class=%s", name, srr.Spec.EntranceClass)
		}
		if srr.Annotations[gateway.AnnotationExtAuthUser] != "alice" {
			t.Fatalf("%s missing extauth-user", name)
		}
	}

	auth := &srrv1alpha1.SharedRouteRegistry{}
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: "user-space-alice", Name: "sys-auth"}, auth)
	if auth.Annotations[gateway.AnnotationAuthMode] != gateway.AuthModeNone {
		t.Fatalf("auth mode=%s", auth.Annotations[gateway.AnnotationAuthMode])
	}
	if auth.Annotations[gateway.AnnotationStampBFLUser] != "alice" {
		t.Fatal("auth must stamp X-BFL-USER")
	}
	desktop := &srrv1alpha1.SharedRouteRegistry{}
	_ = c.Get(context.Background(), types.NamespacedName{Namespace: "user-space-alice", Name: "sys-desktop"}, desktop)
	if desktop.Annotations[gateway.AnnotationAuthMode] != gateway.AuthModeExtAuth {
		t.Fatalf("desktop mode=%s", desktop.Annotations[gateway.AnnotationAuthMode])
	}
	if err := c.Delete(context.Background(), cm); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name},
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: cm.Namespace, Name: "sys-desktop"}, desktop); !apierrors.IsNotFound(err) {
		t.Fatalf("deleted user zone retained Desktop route: %v", err)
	}

}

func TestSystemPlaneRouteReconcilerEphemeralWizardOnly(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = srrv1alpha1.AddToScheme(scheme)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: zoneSSLConfigMapName, Namespace: "user-space-guest1"},
		Data: map[string]string{
			"zone":      "alice.olares.com",
			"ephemeral": "true",
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	r := &SystemPlaneRouteReconciler{Client: c}
	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Namespace: cm.Namespace, Name: cm.Name},
	}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	wizard := &srrv1alpha1.SharedRouteRegistry{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "user-space-guest1", Name: "sys-wizard"}, wizard); err != nil {
		t.Fatalf("wizard: %v", err)
	}
	if len(wizard.Spec.HostPatterns) == 0 || wizard.Spec.HostPatterns[0] != "wizard-guest1.alice.olares.com" {
		t.Fatalf("wizard hosts=%v", wizard.Spec.HostPatterns)
	}
	for _, name := range []string{"sys-auth", "sys-desktop", "sys-profile"} {
		srr := &srrv1alpha1.SharedRouteRegistry{}
		err := c.Get(context.Background(), types.NamespacedName{Namespace: "user-space-guest1", Name: name}, srr)
		if err == nil {
			t.Fatalf("ephemeral must not create %s", name)
		}
	}
}
