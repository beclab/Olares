package gateway

import (
	"context"
	"fmt"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

// SharedRouteProducerReconciler turns shared gateway-mode Applications into
// SharedRouteRegistry objects (one per sharedEntrance). It is the SRR producer;
// the separate SRR controller turns SRRs into HTTPRoutes.
type SharedRouteProducerReconciler struct {
	Client client.Client
}

// Reconcile writes the per-entrance SRRs for one Application, or removes them
// when the app is not a gateway shared app / has opted out / is gone.
func (r *SharedRouteProducerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	app := &appv1alpha1.Application{}
	if err := r.Client.Get(ctx, req.NamespacedName, app); err != nil {
		// Application gone: owner-ref GC removes the SRRs, nothing to do.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	return reconcile.Result{}, r.reconcileApp(ctx, app)
}

func (r *SharedRouteProducerReconciler) reconcileApp(ctx context.Context, app *appv1alpha1.Application) error {
	if app == nil || app.Spec.Namespace == "" || app.Spec.Name == "" {
		return nil
	}
	if !appcfg.IsGatewaySharedApp(app) || !IsOptedIn(app) {
		return DeleteAllForApp(ctx, r.Client, app)
	}

	// Remove the legacy "shared-<appName>" SRR, then write one per entrance.
	if err := Delete(ctx, r.Client, app); err != nil {
		return fmt.Errorf("remove legacy SRR: %w", err)
	}

	platformDomain := cluster.GetPlatformDomain(ctx)
	if platformDomain == "" {
		return fmt.Errorf("platformDomain is empty")
	}
	appid := EntranceAppID(app)
	desired := make(map[string]struct{}, len(app.Spec.SharedEntrances))

	for i := range app.Spec.SharedEntrances {
		entrance := app.Spec.SharedEntrances[i]
		if entrance.Name == "" {
			klog.Warningf("SRR skip: app=%s entrance#%d has empty name", app.Spec.Name, i)
			continue
		}
		if entrance.Host == "" {
			return fmt.Errorf("shared entrance %q on app %s has empty host", entrance.Name, app.Spec.Name)
		}
		svc, err := ResolveSharedEntranceService(ctx, r.Client, app, entrance.Host)
		if err != nil {
			return fmt.Errorf("resolve backing service for entrance %q: %w", entrance.Name, err)
		}
		spec, err := BuildSpecForEntrance(app, entrance, i, svc, platformDomain)
		if err != nil {
			return fmt.Errorf("build SRR spec for entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntrance(appid, entrance.Name)
		if err := CheckLogicalPatternUniqueness(ctx, r.Client, spec.HostPatterns[0], app.Spec.Namespace, name); err != nil {
			return fmt.Errorf("uniqueness check for entrance %q: %w", entrance.Name, err)
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec); err != nil {
			return err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)
	}

	if err := PruneEntranceSRRs(ctx, r.Client, app, desired); err != nil {
		return fmt.Errorf("prune stale SRRs: %w", err)
	}
	return nil
}

// SetupWithManager registers the producer against Applications.
func (r *SharedRouteProducerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("shared-route-producer").
		For(&appv1alpha1.Application{}).
		Complete(r)
}
