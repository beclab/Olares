package controllers

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var kubernetesAPIKey = client.ObjectKey{Namespace: security.KubernetesAPINamespace, Name: security.KubernetesAPIName}

func kubernetesAPINetworkPolicyRequests(_ context.Context, obj client.Object) []reconcile.Request {
	if client.ObjectKeyFromObject(obj) != kubernetesAPIKey {
		return nil
	}
	return []reconcile.Request{
		{NamespacedName: client.ObjectKey{Name: metav1.NamespaceSystem}},
		{NamespacedName: client.ObjectKey{Name: "os-framework"}},
	}
}

func watchKubernetesAPINetworkPolicySources(c controller.Controller, sourceCache cache.Cache) error {
	for _, obj := range []client.Object{&corev1.Service{}, &corev1.Endpoints{}} {
		if err := c.Watch(source.Kind(sourceCache, obj, handler.EnqueueRequestsFromMapFunc(kubernetesAPINetworkPolicyRequests))); err != nil {
			return err
		}
	}
	return nil
}

func (r *SecurityReconciler) reconcileTrafficGroupDependencies(ctx context.Context, namespace string) error {
	if namespace != metav1.NamespaceSystem && namespace != "os-framework" {
		return nil
	}
	var service corev1.Service
	if err := r.Get(ctx, kubernetesAPIKey, &service); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read Kubernetes API Service: %w", err)
	}
	var endpoints corev1.Endpoints
	if err := r.Get(ctx, kubernetesAPIKey, &endpoints); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("read Kubernetes API Endpoints: %w", err)
	}
	for _, policy := range security.BuildTrafficGroupDependencyPolicies(namespace, &service, &endpoints) {
		if err := r.upsertNetworkPolicy(ctx, &policy); err != nil {
			return fmt.Errorf("reconcile NetworkPolicy %s/%s: %w", namespace, policy.Name, err)
		}
	}
	return nil
}
