package controllers

import (
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOverlayMACLifecycleWaitsForPodsBeforeReleasingFinalizer(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add application scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add pod scheme: %v", err)
	}
	deletedAt := metav1.Now()
	app := &appv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "app-space-jellyfin",
			DeletionTimestamp: &deletedAt,
			Finalizers:        []string{overlayMACFinalizer},
		},
		Spec: appv1alpha1.ApplicationSpec{
			Name:      "jellyfin",
			Namespace: "app-space",
			Owner:     "",
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "jellyfin-pod",
			Namespace: "app-space",
			Labels: map[string]string{
				constants.ApplicationNameLabel:        "jellyfin",
				constants.ApplicationOwnerLabel:       "",
				constants.ApplicationMacvlanInitLabel: "true",
			},
			Finalizers: []string{"test/pod-finalizer"},
		},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(app, pod).Build()
	reconciler := &OverlayMACLifecycleReconciler{Client: kubeClient, Scheme: scheme}
	req := ctrlRequest("app-space-jellyfin")

	result, err := reconciler.Reconcile(t.Context(), req)
	if err != nil {
		t.Fatalf("reconcile with active pod: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected finalizer reconciliation to wait for active pod")
	}

	if err := kubeClient.Delete(t.Context(), pod); err != nil {
		t.Fatalf("mark pod terminating: %v", err)
	}
	result, err = reconciler.Reconcile(t.Context(), req)
	if err != nil {
		t.Fatalf("reconcile with terminating pod: %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected finalizer reconciliation to wait for terminating pod")
	}

	terminating := &corev1.Pod{}
	if err := kubeClient.Get(t.Context(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, terminating); err != nil {
		t.Fatalf("get terminating pod: %v", err)
	}
	terminating.Finalizers = nil
	if err := kubeClient.Update(t.Context(), terminating); err != nil {
		t.Fatalf("remove pod finalizer: %v", err)
	}
	if _, err := reconciler.Reconcile(t.Context(), req); err != nil {
		t.Fatalf("reconcile after pod deletion: %v", err)
	}
	updated := &appv1alpha1.Application{}
	if err := kubeClient.Get(t.Context(), types.NamespacedName{Name: app.Name}, updated); err != nil {
		if apierrors.IsNotFound(err) {
			return
		}
		t.Fatalf("get application: %v", err)
	}
	if containsFinalizer(updated.Finalizers, overlayMACFinalizer) {
		t.Fatalf("overlay MAC finalizer was not removed: %v", updated.Finalizers)
	}
}

func ctrlRequest(name string) ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: name}}
}
