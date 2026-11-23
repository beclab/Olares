package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
			srrv1alpha1.EntranceClassShared)
		if err != nil {
			return fmt.Errorf("build SRR spec for entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntrance(appid, entrance.Name)
		for _, hp := range spec.HostPatterns {
			if err := CheckLogicalPatternUniqueness(ctx, r.Client, hp, app.Spec.Namespace, name); err != nil {
				return fmt.Errorf("uniqueness check for entrance %q pattern %q: %w", entrance.Name, hp, err)
			}
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec); err != nil {
			return err
		}
		desired[name] = struct{}{}
		klog.V(1).Infof("SRR reconciled app=%s/%s entrance=%s name=%s hostPatterns=%v upstream=%s/%s:%d",
			app.Spec.Namespace, app.Spec.Name, entrance.Name, name, spec.HostPatterns,
			spec.Upstream.ServiceNamespace, spec.Upstream.ServiceName, spec.Upstream.Port)
	}

	var zone string
	if len(app.Spec.Entrances) != 0 {
		if app.Spec.Owner == "" {
			return PruneEntranceSRRs(ctx, r.Client, app, desired)
		}
		var registeredZone corev1.ConfigMap
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: "user-space-" + app.Spec.Owner, Name: "zone-ssl-config"}, &registeredZone); err != nil {
			if apierrors.IsNotFound(err) {
				return PruneEntranceSRRs(ctx, r.Client, app, desired)
			}
			return fmt.Errorf("resolve registered zone for application %s: %w", app.Name, err)
		}
		zone = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(registeredZone.Data["zone"]), "."))
		if normalized, err := NormalizeHostOrLogicalPattern(zone); err != nil || normalized != zone || strings.Contains(zone, "*") {
			return PruneEntranceSRRs(ctx, r.Client, app, desired)
		}
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
			srrv1alpha1.EntranceClassApplication)
		if err != nil {
			return fmt.Errorf("build SRR spec for application entrance %q: %w", entrance.Name, err)
		}
		spec.HostPatterns, err = bindApplicationAuthorities(app, entrance.Name, zone, spec.HostPatterns)
		if err != nil {
			return fmt.Errorf("bind authorities for entrance %q: %w", entrance.Name, err)
		}
		name := ResourceNameForEntranceApp(appid, entrance.Name)
		for _, hp := range spec.HostPatterns {
			if err := CheckLogicalPatternUniqueness(ctx, r.Client, hp, app.Spec.Namespace, name); err != nil {
				return fmt.Errorf("uniqueness check for application entrance %q pattern %q: %w", entrance.Name, hp, err)
			}
		}
		if _, err := ReconcileForEntrance(ctx, r.Client, app, entrance, spec); err != nil {
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

// bindApplicationAuthorities materializes the tenant's registered zone and
// per-entrance custom domains. A zone label never derives the tenant identity.
func bindApplicationAuthorities(app *appv1alpha1.Application, entrance, zone string, logical []string) ([]string, error) {
	var custom map[string]map[string]string
	if raw := app.EffectiveSettings(app.Spec.Owner)["customDomain"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &custom); err != nil {
			return nil, err
		}
	}
	primary := strings.SplitN(zone, ".", 2)[0]
	localZone := primary + ".olares.local"
	seen := make(map[string]bool)
	var hosts []string
	add := func(raw string) error {
		host, err := NormalizeHostOrLogicalPattern(raw)
		if err != nil {
			return err
		}
		if strings.Contains(host, "*") {
			return fmt.Errorf("application entrance requires an exact authority")
		}
		if !seen[host] {
			hosts = append(hosts, host)
			seen[host] = true
		}
		return nil
	}
	addPrefix := func(prefix string) error {
		if err := add(prefix + "." + zone); err != nil {
			return err
		}
		return add(prefix + "." + localZone)
	}
	for _, pattern := range logical {
		label, _, _ := strings.Cut(pattern, ".*.")
		if err := addPrefix(label); err != nil {
			return nil, err
		}
	}
	if prefix := custom[entrance]["third_level_domain"]; prefix != "" {
		if err := addPrefix(prefix); err != nil {
			return nil, err
		}
	}
	if host := custom[entrance]["third_party_domain"]; host != "" {
		if err := add(host); err != nil {
			return nil, err
		}
	}
	return hosts, nil
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
	const ownerIndex = "gateway.olares.io/application-owner"
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appv1alpha1.Application{}, ownerIndex, func(obj client.Object) []string {
		return []string{obj.(*appv1alpha1.Application).Spec.Owner}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named("shared-route-producer").
		For(&appv1alpha1.Application{}).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
			var apps appv1alpha1.ApplicationList
			if err := r.Client.List(ctx, &apps, client.MatchingFields{ownerIndex: strings.TrimPrefix(obj.GetNamespace(), "user-space-")}); err != nil {
				klog.Warningf("list applications affected by registered zone change: %v", err)
				return nil
			}
			var requests []reconcile.Request
			for i := range apps.Items {
				requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&apps.Items[i])})
			}
			return requests
		}), builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetName() == "zone-ssl-config" && strings.HasPrefix(obj.GetNamespace(), "user-space-")
		}))).
		Complete(r)
}
