package controllers

import (
	"context"
	"fmt"
	"slices"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var kubernetesAPIKey = client.ObjectKey{Namespace: security.KubernetesAPINamespace, Name: security.KubernetesAPIName}

func trafficGroupNetworkPolicyRequests(_ context.Context, obj client.Object) []reconcile.Request {
	key := client.ObjectKeyFromObject(obj)
	if pod, ok := obj.(*corev1.Pod); ok {
		if security.IsL4HostNetworkPod(pod) {
			return []reconcile.Request{{NamespacedName: client.ObjectKey{Name: "os-gateway"}}}
		}
		return nil
	}
	var namespaces []string
	if key == kubernetesAPIKey {
		namespaces = security.KubernetesAPIPolicyNamespaces()
	} else if _, ok := obj.(*corev1.Service); ok {
		namespaces = security.PlatformTrafficNamespaces(key)
	}
	requests := make([]reconcile.Request, 0, len(namespaces))
	for _, namespace := range namespaces {
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKey{Name: namespace}})
	}
	return requests
}

func watchTrafficGroupNetworkPolicySources(c controller.Controller, sourceCache cache.Cache) error {
	for _, obj := range []client.Object{&corev1.Service{}, &corev1.Endpoints{}, &corev1.Pod{}} {
		if err := c.Watch(source.Kind(sourceCache, obj, handler.EnqueueRequestsFromMapFunc(trafficGroupNetworkPolicyRequests))); err != nil {
			return err
		}
	}
	return nil
}

func (r *SecurityReconciler) reconcileTrafficGroupDependencies(ctx context.Context, namespace string) error {
	if slices.Contains(security.KubernetesAPIPolicyNamespaces(), namespace) {
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
	}
	services := make(map[client.ObjectKey]*corev1.Service)
	for _, key := range security.PlatformTrafficServices(namespace) {
		var service corev1.Service
		if err := r.Get(ctx, key, &service); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("read platform Service %s: %w", key, err)
		}
		services[key] = &service
	}
	policies := security.BuildPlatformTrafficPolicies(namespace, services)
	if namespace == "os-gateway" {
		var l4Pods corev1.PodList
		if err := r.List(ctx, &l4Pods, client.InNamespace("os-network"),
			client.MatchingLabels{"app": "l4-bfl-proxy", "tier": "bfl"}); err != nil {
			return fmt.Errorf("list L4 hostNetwork Pods: %w", err)
		}
		policies = append(policies, security.L4HostNetworkIngressPolicy(l4Pods.Items))
	}
	for _, policy := range policies {
		if err := r.upsertNetworkPolicy(ctx, &policy); err != nil {
			return fmt.Errorf("reconcile NetworkPolicy %s/%s: %w", namespace, policy.Name, err)
		}
	}
	return nil
}
