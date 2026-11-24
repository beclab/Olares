package routecontrol

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// SharedRouteReconciler turns each SharedRouteRegistry into an HTTPRoute (plus
// NetworkPolicy + ReferenceGrant) on the app-gateway, and records the outcome
// on SRR.status. The generated HTTPRoute is owned by the SRR so deleting the
// SRR garbage-collects the route and re-creates it on drift.
type SharedRouteReconciler struct {
	Client client.Client
}

// Reconcile applies the route objects for one SRR.
func (r *SharedRouteReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	srr := &srrv1alpha1.SharedRouteRegistry{}
	if err := r.Client.Get(ctx, req.NamespacedName, srr); err != nil {
		if apierrors.IsNotFound(err) {
			// Owner references cannot cover transport policies in another namespace.
			return reconcile.Result{}, pruneMeshTransport(ctx, r.Client, req.NamespacedName, nil)
		}
		return reconcile.Result{}, err
	}

	res, rerr := ReconcileSharedRoute(ctx, r.Client, GatewayRef{}, srr)
	if rerr != nil {
		klog.Warningf("reconcile shared route %s/%s failed: %v", srr.Namespace, srr.Name, rerr)
		res = ReconcileResult{
			Status:  metav1.ConditionFalse,
			Reason:  ReasonRouteApplyFailed,
			Message: rerr.Error(),
		}
	}
	var retry time.Duration
	if rerr == nil && res.Reason == ReasonAwaitingDataPlane {
		retry = 5 * time.Second
	}
	if rerr == nil && res.HTTPRouteName != "" {
		if err := servingReadiness(ctx, r.Client, GatewayRef{}, srr, res.HTTPRouteName); err != nil {
			res.Status, res.Reason, res.Message = metav1.ConditionFalse, ReasonAwaitingDataPlane, err.Error()
			retry = 5 * time.Second
		} else {
			res.Message = "Entry route and current mesh endpoints are ready"
			if isSystemDesktopRoute(srr) {
				res.Reason = ReasonDesktopEntryAuthorizationReady
			}
			// Endpoint changes need not modify the owned Route.
			retry = 30 * time.Second
		}
	}
	if statusErr := UpdateSRRStatus(ctx, r.Client, srr, res); statusErr != nil {
		klog.Warningf("update SRR status %s/%s failed: %v", srr.Namespace, srr.Name, statusErr)
	}
	return reconcile.Result{RequeueAfter: retry}, rerr
}

// SetupWithManager registers the reconciler against SharedRouteRegistry and the
// HTTPRoutes it owns (drift self-heal).
func (r *SharedRouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	ownedRoute := &unstructured.Unstructured{}
	ownedRoute.SetGroupVersionKind(schema.GroupVersionKind{
		Group: "gateway.networking.k8s.io", Version: "v1", Kind: "HTTPRoute",
	})
	return ctrl.NewControllerManagedBy(mgr).
		Named("shared-route").
		For(&srrv1alpha1.SharedRouteRegistry{}).
		Owns(ownedRoute).
		Complete(r)
}
