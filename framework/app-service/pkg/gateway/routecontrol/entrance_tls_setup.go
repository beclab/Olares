package routecontrol

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SetupWithManager registers EntranceTLSReconciler against zone-ssl-config
// ConfigMaps in user-space-* namespaces.
func (r *EntranceTLSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	return builder.ControllerManagedBy(mgr).
		Named("entrance-tls").
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc:  func(e event.CreateEvent) bool { return entranceTLSConfigMap(e.Object) },
			UpdateFunc:  func(e event.UpdateEvent) bool { return entranceTLSConfigMap(e.ObjectNew) },
			DeleteFunc:  func(e event.DeleteEvent) bool { return entranceTLSConfigMap(e.Object) },
			GenericFunc: func(e event.GenericEvent) bool { return entranceTLSConfigMap(e.Object) },
		}).
		Complete(reconcile.Func(func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
			if req.Name != zoneSSLConfigMapName || !strings.HasPrefix(req.Namespace, userSpaceNamespacePrefix) {
				return reconcile.Result{}, nil
			}
			var cm corev1.ConfigMap
			if err := r.Client.Get(ctx, types.NamespacedName{Namespace: req.Namespace, Name: req.Name}, &cm); err != nil {
				if apierrors.IsNotFound(err) {
					viewer, ok := viewerFromUserSpaceNamespace(req.Namespace)
					if ok {
						return reconcile.Result{}, deleteEntranceTLSSecret(ctx, r.Client, viewer)
					}
					return reconcile.Result{}, nil
				}
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, r.ReconcileConfigMap(ctx, &cm)
		}))
}

func entranceTLSConfigMap(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	return obj.GetName() == zoneSSLConfigMapName && strings.HasPrefix(obj.GetNamespace(), userSpaceNamespacePrefix)
}
