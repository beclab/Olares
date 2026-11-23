package routecontrol

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
)

const userSpaceNSPrefix = "user-space-"

// SystemPlaneRouteReconciler projects auth/desktop/wizard/profile SRRs from
// zone-ssl-config so AGW owns system-plane HTTP after L4 edge TLS.
type SystemPlaneRouteReconciler struct {
	Client client.Client
}

// Reconcile upserts system-plane SharedRouteRegistry objects for one userspace.
func (r *SystemPlaneRouteReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	if req.Name != zoneSSLConfigMapName {
		return reconcile.Result{}, nil
	}
	if !strings.HasPrefix(req.Namespace, userSpaceNSPrefix) {
		return reconcile.Result{}, nil
	}
	cm := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, gateway.DeleteSystemPlaneRoutes(ctx, r.Client, req.Namespace)
		}
		return reconcile.Result{}, err
	}
	zone := strings.TrimSpace(cm.Data["zone"])
	if zone == "" {
		klog.V(4).Infof("system-plane: zone-ssl-config %s/%s missing zone", cm.Namespace, cm.Name)
		return reconcile.Result{}, gateway.DeleteSystemPlaneRoutes(ctx, r.Client, req.Namespace)
	}
	user := strings.TrimPrefix(cm.Namespace, userSpaceNSPrefix)
	if user == "" {
		return reconcile.Result{}, nil
	}
	ephemeral := strings.EqualFold(strings.TrimSpace(cm.Data["ephemeral"]), "true")
	if err := gateway.EnsureSystemPlaneRoutes(ctx, r.Client, user, zone, cm.Namespace, ephemeral); err != nil {
		klog.Errorf("system-plane: ensure routes user=%s ns=%s: %v", user, cm.Namespace, err)
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// SetupWithManager watches zone-ssl-config ConfigMaps in user-space namespaces.
func (r *SystemPlaneRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	onlyZoneSSL := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == zoneSSLConfigMapName &&
			strings.HasPrefix(o.GetNamespace(), userSpaceNSPrefix)
	})
	return ctrl.NewControllerManagedBy(mgr).
		Named("system-plane-routes").
		For(&corev1.ConfigMap{}, builder.WithPredicates(onlyZoneSSL)).
		Complete(r)
}
