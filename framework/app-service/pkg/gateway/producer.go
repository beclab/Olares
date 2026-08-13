package gateway

import (
	"context"
	"fmt"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
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
	// Convergence path for the route-mode automation: shared apps created
	// before app-gateway was ready, or whose annotation was removed, get the
	// gateway annotation on the next Application event. The patch triggers
	// another reconcile that then passes the opt-in check below.
	if err := EnsureRouteModeAnnotation(ctx, r.Client, app); err != nil {
		klog.Warningf("ensure gateway route-mode for app %s err=%v", app.Spec.Name, err)
	}
	if !IsOptedIn(app) {
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
	appid := strings.ToLower(strings.TrimSpace(app.Spec.Appid))
	if appid == "" {
		return fmt.Errorf("invalid appid in app.spec.appid for app %q", app.Spec.Name)
	}
	materializer := NewConfigMapCertMaterializer(ctx, r.Client)
	desired := make(map[string]struct{}, len(app.Spec.SharedEntrances)+len(app.Spec.Entrances))

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
		spec, err := BuildSpecForEntrance(app, entrance, i, len(app.Spec.SharedEntrances), svc, platformDomain,
			srrv1alpha1.EntranceClassShared, materializer)
		if err != nil {
			return fmt.Errorf("build SRR spec for entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntrance(appid, entrance.Name)
		for _, hp := range spec.HostPatterns {
			if err := CheckLogicalPatternUniqueness(ctx, r.Client, hp, app.Spec.Namespace, name); err != nil {
				return fmt.Errorf("uniqueness check for entrance %q pattern %q: %w", entrance.Name, hp, err)
			}
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec, platformDomain, materializer); err != nil {
			return err
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
			return fmt.Errorf("resolve backing service for application entrance %q: %w", entrance.Name, err)
		}
		spec, err := BuildSpecForEntrance(app, entrance, i, len(app.Spec.Entrances), svc, platformDomain,
			srrv1alpha1.EntranceClassApplication, materializer)
		if err != nil {
			return fmt.Errorf("build SRR spec for application entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntranceApp(appid, entrance.Name)
		for _, hp := range spec.HostPatterns {
			if err := CheckLogicalPatternUniqueness(ctx, r.Client, hp, app.Spec.Namespace, name); err != nil {
				return fmt.Errorf("uniqueness check for application entrance %q pattern %q: %w", entrance.Name, hp, err)
			}
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec, platformDomain, materializer); err != nil {
			return err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("application SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)
	}

	if err := PruneEntranceSRRs(ctx, r.Client, app, desired); err != nil {
		return fmt.Errorf("prune stale SRRs: %w", err)
	}
	return nil
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

// SetupWithManager registers the producer against Applications and custom-domain
// cert ConfigMaps (cert late-arrival / rotation must re-run Eligible).
func (r *SharedRouteProducerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&appv1alpha1.Application{},
		indexApplicationThirdPartyDomain,
		indexApplicationByThirdPartyDomain,
	); err != nil {
		return fmt.Errorf("index Applications by third_party_domain: %w", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("shared-route-producer").
		For(&appv1alpha1.Application{}).
		Watches(&corev1.ConfigMap{}, r.customDomainCertEventHandler(),
			builder.WithPredicates(customDomainCertConfigMapPredicate())).
		Complete(r)
}

func indexApplicationByThirdPartyDomain(obj client.Object) []string {
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app == nil {
		return nil
	}
	return CollectApplicationThirdPartyDomains(app)
}

func (r *SharedRouteProducerReconciler) customDomainCertEventHandler() handler.EventHandler {
	enqueue := func(ctx context.Context, obj client.Object, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
		for _, req := range r.fanOutAppsOnCustomDomainCert(ctx, obj) {
			q.Add(req)
		}
	}
	return handler.Funcs{
		CreateFunc: func(ctx context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			enqueue(ctx, e.Object, q)
		},
		UpdateFunc: func(ctx context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			// Zone corrections must refresh both old and new referring apps.
			enqueue(ctx, e.ObjectOld, q)
			enqueue(ctx, e.ObjectNew, q)
		},
		DeleteFunc: func(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			enqueue(ctx, e.Object, q)
		},
		GenericFunc: func(ctx context.Context, e event.GenericEvent, q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
			enqueue(ctx, e.Object, q)
		},
	}
}

// fanOutAppsOnCustomDomainCert enqueues only Applications that declare the CM
// zone as third_party_domain. Empty/unknown zone → no enqueue (fail closed,
// never broadcast to all apps).
func (r *SharedRouteProducerReconciler) fanOutAppsOnCustomDomainCert(ctx context.Context, obj client.Object) []reconcile.Request {
	if r == nil || r.Client == nil || obj == nil {
		return nil
	}
	zone := customDomainCertZone(obj)
	if zone == "" {
		klog.V(4).Infof("gateway-tpd: custom-domain-cert %s/%s missing zone; skip fan-out",
			obj.GetNamespace(), obj.GetName())
		return nil
	}
	list := &appv1alpha1.ApplicationList{}
	if err := r.Client.List(ctx, list, client.MatchingFields{indexApplicationThirdPartyDomain: zone}); err != nil {
		klog.Errorf("gateway-tpd: list Applications by third_party_domain=%s failed: %v", zone, err)
		return nil
	}
	if len(list.Items) == 0 {
		return nil
	}
	out := make([]reconcile.Request, 0, len(list.Items))
	seen := make(map[types.NamespacedName]struct{}, len(list.Items))
	for i := range list.Items {
		key := client.ObjectKeyFromObject(&list.Items[i])
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, reconcile.Request{NamespacedName: key})
	}
	klog.V(4).Infof("gateway-tpd: custom-domain-cert zone=%s enqueue %d Application(s)", zone, len(out))
	return out
}

func customDomainCertZone(obj client.Object) string {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok || cm == nil || cm.Data == nil {
		return ""
	}
	zone, err := NormalizeHostPattern(cm.Data["zone"])
	if err != nil {
		return ""
	}
	return zone
}

func customDomainCertConfigMapPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isCustomDomainCertConfigMap(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Cert rotation or zone fix — either generation of the labeled CM.
			return isCustomDomainCertConfigMap(e.ObjectNew) || isCustomDomainCertConfigMap(e.ObjectOld)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isCustomDomainCertConfigMap(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isCustomDomainCertConfigMap(e.Object)
		},
	}
}

func isCustomDomainCertConfigMap(obj client.Object) bool {
	if obj == nil {
		return false
	}
	return obj.GetLabels()[customDomainCertLabel] == customDomainCertLabelValue
}
