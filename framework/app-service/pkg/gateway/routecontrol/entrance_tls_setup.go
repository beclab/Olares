package routecontrol

import (
	"context"
	"strings"
	"time"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	if snap, err := cluster.GetSnapshot(context.Background()); err == nil {
		r.platformDomain = snap.PlatformDomain
	}
	err := builder.ControllerManagedBy(mgr).
		Named("entrance-tls").
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc:  func(e event.CreateEvent) bool { return entranceTLSConfigMap(e.Object) },
			UpdateFunc:  func(e event.UpdateEvent) bool { return entranceTLSConfigMap(e.ObjectNew) },
			DeleteFunc:  func(e event.DeleteEvent) bool { return entranceTLSConfigMap(e.Object) },
			GenericFunc: func(e event.GenericEvent) bool { return entranceTLSConfigMap(e.Object) },
		}).
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.mapPodToReplica),
			builder.WithPredicates(predicate.NewPredicateFuncs(isSharedEntrancePod))).
		Watches(&appv1alpha1.Application{}, handler.EnqueueRequestsFromMapFunc(r.mapApplicationToReplica),
			builder.WithPredicates(predicate.NewPredicateFuncs(isClusterScopedApplication))).
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
			if err := r.ReconcileConfigMap(ctx, &cm); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, r.refreshDemandSnapshot(ctx)
		}))
	if err != nil {
		return err
	}
	return mgr.Add(&orphanSweepRunnable{
		client:   mgr.GetClient(),
		interval: 15 * time.Minute,
		demand:   r.demandSnapshotFn(),
		refresh:  r.refreshDemandSnapshot,
	})
}

func entranceTLSConfigMap(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	return obj.GetName() == zoneSSLConfigMapName && strings.HasPrefix(obj.GetNamespace(), userSpaceNamespacePrefix)
}

func isSharedEntrancePod(obj client.Object) bool {
	if obj == nil || !strings.HasPrefix(obj.GetNamespace(), userSpaceNamespacePrefix) {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(obj.GetLabels()[constants.AppSharedEntrancesLabel]), "true")
}

func isClusterScopedApplication(obj client.Object) bool {
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app == nil {
		return false
	}
	return strings.TrimSpace(app.Spec.Settings["clusterScoped"]) == "true"
}

func (r *EntranceTLSReconciler) mapPodToReplica(ctx context.Context, obj client.Object) []reconcile.Request {
	if obj == nil {
		return nil
	}
	ns := strings.TrimSpace(obj.GetNamespace())
	if ns == "" || !strings.HasPrefix(ns, userSpaceNamespacePrefix) {
		return nil
	}
	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{Namespace: ns, Name: zoneSSLConfigMapName},
	}}
}

func (r *EntranceTLSReconciler) mapApplicationToReplica(ctx context.Context, obj client.Object) []reconcile.Request {
	if obj == nil || r == nil || r.Client == nil {
		return nil
	}
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app == nil {
		return nil
	}
	var reqs []reconcile.Request
	if ns := strings.TrimSpace(app.Spec.Namespace); strings.HasPrefix(ns, userSpaceNamespacePrefix) {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: ns, Name: zoneSSLConfigMapName},
		})
	}
	var cms corev1.ConfigMapList
	if err := r.Client.List(ctx, &cms); err != nil {
		return reqs
	}
	seen := map[string]struct{}{}
	for _, req := range reqs {
		seen[req.Namespace] = struct{}{}
	}
	for i := range cms.Items {
		cm := cms.Items[i]
		if cm.Name != zoneSSLConfigMapName || !strings.HasPrefix(cm.Namespace, userSpaceNamespacePrefix) {
			continue
		}
		if _, ok := seen[cm.Namespace]; ok {
			continue
		}
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: cm.Namespace, Name: zoneSSLConfigMapName},
		})
		seen[cm.Namespace] = struct{}{}
	}
	return reqs
}

type orphanSweepRunnable struct {
	client   client.Client
	interval time.Duration
	demand   func() []ReplicaTarget
	refresh  func(context.Context) error
}

func (s *orphanSweepRunnable) NeedLeaderElection() bool { return true }

func (s *orphanSweepRunnable) Start(ctx context.Context) error {
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if s.refresh != nil {
				_ = s.refresh(ctx)
			}
			demand := []ReplicaTarget(nil)
			if s.demand != nil {
				demand = s.demand()
			}
			_ = sweepOrphanReplicas(ctx, s.client, demand)
		}
	}
}

var _ manager.LeaderElectionRunnable = (*orphanSweepRunnable)(nil)
