package gateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
)

// routeModePlatformRequeue paces retries while platform domain / app-gateway
// readiness is not yet available for auto gateway opt-in.
const routeModePlatformRequeue = 15 * time.Second

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
	return r.reconcileApp(ctx, app)
}

func (r *SharedRouteProducerReconciler) reconcileApp(ctx context.Context, app *appv1alpha1.Application) (reconcile.Result, error) {
	if app == nil || app.Spec.Namespace == "" || app.Spec.Name == "" {
		return reconcile.Result{}, nil
	}
	// Zone / Gateway may arrive after the Application. Requeue instead of
	// returning success with no route-mode write (false-green empty reconcile).
	// Still delete SRRs when not opted in so stale gateway state is not left behind.
	if wait, delay := routeModeWaitForPlatform(ctx, r.Client, app); wait {
		klog.V(2).Infof("shared-route-producer: requeue app=%s/%s: platform not ready for route-mode",
			app.Namespace, app.Name)
		if !IsOptedIn(app) {
			if err := DeleteAllForApp(ctx, r.Client, app); err != nil {
				klog.Errorf("shared-route-producer: delete SRRs while waiting for platform app=%s/%s: %v",
					app.Namespace, app.Name, err)
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{RequeueAfter: delay}, nil
	}
	// Convergence path for the route-mode automation: shared apps created
	// before app-gateway was ready, or whose annotation was removed, get the
	// gateway annotation on the next Application event. The patch triggers
	// another reconcile that then passes the opt-in check below.
	if err := EnsureRouteModeAnnotation(ctx, r.Client, app); err != nil {
		klog.Warningf("ensure gateway route-mode for app %s err=%v", app.Spec.Name, err)
		return reconcile.Result{}, err
	}
	if !IsOptedIn(app) {
		return reconcile.Result{}, DeleteAllForApp(ctx, r.Client, app)
	}

	// Remove the legacy "shared-<appName>" SRR, then write one per entrance.
	if err := Delete(ctx, r.Client, app); err != nil {
		return reconcile.Result{}, fmt.Errorf("remove legacy SRR: %w", err)
	}

	platformDomain := cluster.GetPlatformDomain(ctx)
	if platformDomain == "" {
		klog.Warningf("shared-route-producer: platformDomain empty, requeue app=%s/%s", app.Namespace, app.Name)
		return reconcile.Result{RequeueAfter: routeModePlatformRequeue}, nil
	}
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	if appid == "" {
		return reconcile.Result{}, fmt.Errorf("invalid appid in app.spec.appid for app %q", app.Spec.Name)
	}
	desired := make(map[string]struct{}, len(app.Spec.SharedEntrances)+len(app.Spec.Entrances))

	for i := range app.Spec.SharedEntrances {
		entrance := app.Spec.SharedEntrances[i]
		if entrance.Name == "" {
			klog.Warningf("SRR skip: app=%s entrance#%d has empty name", app.Spec.Name, i)
			continue
		}
		if entrance.Host == "" {
			return reconcile.Result{}, fmt.Errorf("shared entrance %q on app %s has empty host", entrance.Name, app.Spec.Name)
		}
		svc, err := ResolveSharedEntranceService(ctx, r.Client, app, entrance.Host)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("resolve backing service for entrance %q: %w", entrance.Name, err)
		}
		spec, err := BuildSpecForEntrance(app, entrance, i, len(app.Spec.SharedEntrances), svc, platformDomain,
			srrv1alpha1.EntranceClassShared)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("build SRR spec for entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntrance(appid, entrance.Name)
		if err := CheckLogicalPatternUniqueness(ctx, r.Client, spec.HostPatterns[0], app.Spec.Namespace, name); err != nil {
			return reconcile.Result{}, fmt.Errorf("uniqueness check for entrance %q: %w", entrance.Name, err)
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec); err != nil {
			return reconcile.Result{}, err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)
	}

	for i := range app.Spec.Entrances {
		entrance := app.Spec.Entrances[i]
		if entrance.Name == "" {
			klog.Warningf("SRR skip: app=%s entrance#%d has empty name", app.Spec.Name, i)
			continue
		}
		if entrance.Host == "" {
			klog.Warningf("SRR skip: app=%s entrance#%d has empty host", app.Spec.Name, i)
			continue
		}
		svc, err := resolveApplicationEntranceService(ctx, r.Client, app, entrance.Host)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("resolve backing service for application entrance %q: %w", entrance.Name, err)
		}
		spec, err := BuildSpecForEntrance(app, entrance, i, len(app.Spec.Entrances), svc, platformDomain,
			srrv1alpha1.EntranceClassApplication)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("build SRR spec for application entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntranceApp(appid, entrance.Name)
		if err := CheckLogicalPatternUniqueness(ctx, r.Client, spec.HostPatterns[0], app.Spec.Namespace, name); err != nil {
			return reconcile.Result{}, fmt.Errorf("uniqueness check for application entrance %q: %w", entrance.Name, err)
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec); err != nil {
			return reconcile.Result{}, err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("application SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)
	}

	if err := PruneEntranceSRRs(ctx, r.Client, app, desired); err != nil {
		return reconcile.Result{}, fmt.Errorf("prune stale SRRs: %w", err)
	}
	return reconcile.Result{}, nil
}

// routeModeWaitForPlatform reports whether auto gateway opt-in should wait for
// platform domain / app-gateway readiness instead of returning success with no write.
func routeModeWaitForPlatform(ctx context.Context, c client.Client, app *appv1alpha1.Application) (bool, time.Duration) {
	if app == nil || IsOptedIn(app) {
		return false, 0
	}
	if _, ok := explicitRouteModeAnnotation(app); ok {
		return false, 0
	}
	if s := settingsRouteMode(app); s == AnnotationRouteModeDirect {
		return false, 0
	}
	if appcfg.IsGatewaySharedApp(app) {
		if !cluster.GetInClusterGatewayEnabled(ctx) {
			return false, 0
		}
		if cluster.GetPlatformDomain(ctx) == "" || (c != nil && !appGatewayReady(ctx, c)) {
			return true, routeModePlatformRequeue
		}
		return false, 0
	}
	if ordinaryAppNeedsGateway(app) && !ordinaryAppGatewayPlatformReady(ctx, c) {
		return true, routeModePlatformRequeue
	}
	if ordinaryAppGatewayEligible(ctx) && c != nil && !appGatewayReady(ctx, c) {
		return true, routeModePlatformRequeue
	}
	return false, 0
}

func resolveApplicationEntranceService(ctx context.Context, c client.Client,
	app *appv1alpha1.Application, serviceName string) (*corev1.Service, error) {
	if app == nil || app.Spec.Namespace == "" {
		return nil, fmt.Errorf("application or spec.namespace is empty")
	}
	if serviceName == "" {
		return nil, fmt.Errorf("application entrance service name is empty")
	}
	svc := &corev1.Service{}
	if err := c.Get(ctx, client.ObjectKey{Namespace: app.Spec.Namespace, Name: serviceName}, svc); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("backing service %q not found in %s", serviceName, app.Spec.Namespace)
		}
		return nil, fmt.Errorf("get backing service %s/%s: %w", app.Spec.Namespace, serviceName, err)
	}
	return svc, nil
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
