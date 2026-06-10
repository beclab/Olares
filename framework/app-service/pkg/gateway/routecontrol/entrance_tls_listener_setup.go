package routecontrol

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SetupWithManager registers EntranceTLSListenerReconciler against per-viewer
// shared-entrance-tls-* Secrets and the managed Gateway. The Gateway watch lets
// the reconciler re-converge after a helm upgrade reverts the certRef (TC-12).
func (r *EntranceTLSListenerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	gateway := newGatewayUnstructured()
	return builder.ControllerManagedBy(mgr).
		Named("entrance-tls-listener").
		For(&corev1.Secret{}, builder.WithPredicates(predicate.NewPredicateFuncs(isEntranceTLSSecret))).
		Watches(gateway, handler.EnqueueRequestsFromMapFunc(enqueueManagedGateway),
			builder.WithPredicates(predicate.NewPredicateFuncs(isManagedGateway))).
		Complete(r)
}

// isEntranceTLSSecret matches primary shared-entrance-tls-<viewer> Secrets in
// the gateway namespace (replicas in caller namespaces are excluded).
func isEntranceTLSSecret(obj client.Object) bool {
	if obj == nil || obj.GetNamespace() != defaultGatewayNS {
		return false
	}
	if obj.GetLabels()[labelTLSReplica] == "true" {
		return false
	}
	return strings.HasPrefix(obj.GetName(), entranceTLSSecretPrefix)
}

// isManagedGateway matches the single app-gateway Gateway this reconciler owns.
func isManagedGateway(obj client.Object) bool {
	if obj == nil {
		return false
	}
	return obj.GetNamespace() == defaultGatewayNS && obj.GetName() == defaultGatewayName
}

// enqueueManagedGateway funnels every trigger to the single managed Gateway; the
// reconciler ignores the request and rebuilds from the full Secret set.
func enqueueManagedGateway(_ context.Context, _ client.Object) []reconcile.Request {
	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName},
	}}
}
