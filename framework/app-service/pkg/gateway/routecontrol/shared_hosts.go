package routecontrol

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SharedHostsReconciler keeps olares-d2-shared-hosts ConfigMap in each
// opted-in caller namespace aligned with the per-(caller,viewer) host allow
// set derived from SRR HostPatterns and Application clusterAppRef wiring.
//
// requirement: caller NS opt-in must receive the host allow set so d2 nginx
// njs decideOffload can offload v3 Shared traffic; webhook
// ensureD2SharedHostsPlaceholder writes only an empty placeholder.
// behavior: NS-keyed reconcile with content-hash idempotent Update; NotFound
// waits for the webhook placeholder; GC on NS opt-out, NS delete, or viewer
// gone; managed-by label rules out third-party same-name ConfigMaps.
type SharedHostsReconciler struct {
	Client         client.Client
	platformDomain string
}

// Reconcile is the controller-runtime entry. req.Namespace is the caller NS;
// req.Name is fixed to D2SharedHostsVolumeName for dedupe + audit.
func (r *SharedHostsReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil || req.Namespace == "" {
		return reconcile.Result{}, nil
	}
	var ns corev1.Namespace
	if err := r.Client.Get(ctx, types.NamespacedName{Name: req.Namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			sharedHostsGCTotal.WithLabelValues(rGCNSDeleted).Inc()
			clearSharedHostsHashState(req.Namespace)
			return reconcile.Result{}, nil
		}
		sharedHostsReconcileTotal.WithLabelValues(rResListFailed).Inc()
		return reconcile.Result{}, err
	}
	if !namespaceOptedIntoSharedHosts(&ns) {
		if err := r.gcSharedHostsConfigMap(ctx, req.Namespace, rGCNSOptOut); err != nil {
			sharedHostsReconcileTotal.WithLabelValues(rResUpdateFailed).Inc()
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}
	demand, err := BuildSharedHostsDemand(ctx, r.Client, r.platformDomain)
	if err != nil {
		sharedHostsReconcileTotal.WithLabelValues(rResListFailed).Inc()
		return reconcile.Result{}, err
	}
	var nsTargets []SharedHostsTarget
	for _, t := range demand {
		if t.CallerNamespace == req.Namespace {
			nsTargets = append(nsTargets, t)
		}
	}
	sharedHostsTargetCount.WithLabelValues(hashCallerNS(req.Namespace)).Set(float64(len(nsTargets)))
	return reconcile.Result{}, r.ReconcileNamespace(ctx, req.Namespace, nsTargets)
}

// ReconcileNamespace upserts the per-NS olares-d2-shared-hosts ConfigMap.
// fail-safe: List/Get/Update failures leave the existing ConfigMap intact.
func (r *SharedHostsReconciler) ReconcileNamespace(ctx context.Context, callerNS string, targets []SharedHostsTarget) error {
	if r == nil || r.Client == nil || callerNS == "" {
		return nil
	}
	cm := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: callerNS, Name: constants.D2SharedHostsVolumeName,
	}, cm)
	if apierrors.IsNotFound(err) {
		sharedHostsReconcileTotal.WithLabelValues(rResNotFound).Inc()
		klog.V(4).Infof("shared_hosts: configmap not found; awaiting webhook placeholder ns=%s", hashCallerNS(callerNS))
		return nil
	}
	if err != nil {
		sharedHostsReconcileTotal.WithLabelValues(rResGetFailed).Inc()
		return err
	}
	if !sharedHostsManagedByUs(cm) {
		sharedHostsReconcileTotal.WithLabelValues(rResSkippedUnmanaged).Inc()
		klog.Warningf("shared_hosts: configmap not managed by app-service ns=%s name=%s",
			hashCallerNS(callerNS), cm.Name)
		return nil
	}
	desiredData := buildSharedHostsConfigMapData(targets)
	desiredHash := sharedHostsContentHash(desiredData)
	desiredHostsCount := countSharedHostsRows(targets)
	if cm.Annotations != nil && cm.Annotations[sharedHostsContentHashAnnotation] == desiredHash {
		sharedHostsReconcileTotal.WithLabelValues(rResSkipped).Inc()
		updateSharedHostsHashAge(callerNS, desiredHash, false)
		sharedHostsCount.WithLabelValues(hashCallerNS(callerNS)).Set(float64(desiredHostsCount))
		return nil
	}
	for key := range cm.Data {
		if key == constants.D2SharedHostsFileName {
			continue
		}
		if _, ok := desiredData[key]; !ok {
			sharedHostsGCTotal.WithLabelValues(rGCViewer).Inc()
		}
	}
	cm.Data = desiredData
	if cm.Labels == nil {
		cm.Labels = map[string]string{}
	}
	cm.Labels[constants.D2SharedHostsManagedByLabel] = sharedHostsManagedByValue
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[sharedHostsContentHashAnnotation] = desiredHash
	cm.Annotations[sharedHostsReconciledAtAnnotation] = time.Now().UTC().Format(time.RFC3339)
	if err := r.Client.Update(ctx, cm); err != nil {
		if apierrors.IsConflict(err) {
			sharedHostsReconcileTotal.WithLabelValues(rResUpdateConflict).Inc()
		} else {
			sharedHostsReconcileTotal.WithLabelValues(rResUpdateFailed).Inc()
		}
		return err
	}
	sharedHostsReconcileTotal.WithLabelValues(rResUpdated).Inc()
	updateSharedHostsHashAge(callerNS, desiredHash, true)
	sharedHostsCount.WithLabelValues(hashCallerNS(callerNS)).Set(float64(desiredHostsCount))
	return nil
}

func (r *SharedHostsReconciler) gcSharedHostsConfigMap(ctx context.Context, callerNS, reason string) error {
	cm := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: callerNS, Name: constants.D2SharedHostsVolumeName,
	}, cm)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !sharedHostsManagedByUs(cm) {
		return nil
	}
	if err := r.Client.Delete(ctx, cm); err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	sharedHostsGCTotal.WithLabelValues(reason).Inc()
	sharedHostsReconcileTotal.WithLabelValues(rResDeleted).Inc()
	clearSharedHostsHashState(callerNS)
	sharedHostsTargetCount.DeleteLabelValues(hashCallerNS(callerNS))
	sharedHostsCount.DeleteLabelValues(hashCallerNS(callerNS))
	return nil
}

// BuildSharedHostsDemand walks SRR + Application + opt-in Namespace lists and
// produces the desired per-(callerNS, viewer) host allow sets. See WI-N6 §2.3.
func BuildSharedHostsDemand(ctx context.Context, c client.Client, platformDomain string) ([]SharedHostsTarget, error) {
	if c == nil {
		return nil, nil
	}
	var srrList srrv1alpha1.SharedRouteRegistryList
	if err := c.List(ctx, &srrList); err != nil {
		return nil, err
	}
	var appList appv1alpha1.ApplicationList
	if err := c.List(ctx, &appList); err != nil {
		return nil, err
	}
	var nsList corev1.NamespaceList
	if err := c.List(ctx, &nsList, client.MatchingLabels{security.NamespaceInClusterCallerLabel: "true"}); err != nil {
		return nil, err
	}
	ownerIdx := gateway.BuildClusterAppOwnerIndex(appList.Items)
	nsOwnerIdx := buildNamespaceOwnerIndex(appList.Items)
	srrByOwner := groupSRRByOwner(srrList.Items, nsOwnerIdx)
	appsByNS := map[string][]appv1alpha1.Application{}
	for i := range appList.Items {
		app := appList.Items[i]
		ns := strings.TrimSpace(app.Spec.Namespace)
		if ns == "" {
			ns = strings.TrimSpace(app.Namespace)
		}
		if ns == "" {
			continue
		}
		appsByNS[ns] = append(appsByNS[ns], app)
	}
	type key struct{ ns, viewer string }
	hostsByKey := map[key]map[string]struct{}{}
	for i := range nsList.Items {
		callerNS := nsList.Items[i].Name
		for _, app := range appsByNS[callerNS] {
			refs := gateway.SplitClusterAppRefs(app.Spec.Settings["clusterAppRef"])
			if len(refs) == 0 {
				continue
			}
			if len(refs) > 1 {
				sharedHostsDropTotal.WithLabelValues(rDropMultiRef).Inc()
			}
			owners := gateway.SplitClusterAppRefs(gateway.ResolveClusterAppOwner(ownerIdx, refs[0]))
			if len(owners) == 0 {
				sharedHostsDropTotal.WithLabelValues(rDropOwnerUnresolved).Inc()
				continue
			}
			for _, owner := range owners {
				viewer := strings.ToLower(strings.TrimSpace(owner))
				if viewer == "" {
					continue
				}
				k := key{ns: callerNS, viewer: viewer}
				if hostsByKey[k] == nil {
					hostsByKey[k] = map[string]struct{}{}
				}
				for _, h := range enumerateHostsForViewer(viewer, srrByOwner[viewer], platformDomain) {
					hostsByKey[k][h] = struct{}{}
				}
			}
		}
	}
	out := make([]SharedHostsTarget, 0, len(hostsByKey))
	for k, hosts := range hostsByKey {
		list := make([]string, 0, len(hosts))
		for h := range hosts {
			list = append(list, h)
		}
		sort.Strings(list)
		out = append(out, SharedHostsTarget{CallerNamespace: k.ns, Viewer: k.viewer, Hosts: list})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CallerNamespace == out[j].CallerNamespace {
			return out[i].Viewer < out[j].Viewer
		}
		return out[i].CallerNamespace < out[j].CallerNamespace
	})
	return out, nil
}

// SetupWithManager registers the reconciler on the shared manager. The shared
// manager is already configured with LeaderElection (main.go), so this
// reconciler auto-follows the elected leader (OQ-N6-2).
func (r *SharedHostsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if snap, err := cluster.GetSnapshot(context.Background()); err == nil {
		r.platformDomain = snap.PlatformDomain
	}
	return builder.ControllerManagedBy(mgr).
		Named("shared-hosts").
		For(&corev1.ConfigMap{}, builder.WithPredicates(predicate.NewPredicateFuncs(isSharedHostsConfigMap))).
		Watches(&srrv1alpha1.SharedRouteRegistry{}, handler.EnqueueRequestsFromMapFunc(r.fanOutOnSRR),
			builder.WithPredicates(predicate.NewPredicateFuncs(isGatewayModeSRR))).
		Watches(&appv1alpha1.Application{}, handler.EnqueueRequestsFromMapFunc(r.fanOutOnApplication),
			builder.WithPredicates(predicate.NewPredicateFuncs(isClusterScopedOrCallerApp))).
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.requeueNamespace),
			builder.WithPredicates(inClusterCallerNamespacePredicate())).
		Watches(&corev1.Pod{}, handler.EnqueueRequestsFromMapFunc(r.fanOutOnPod),
			builder.WithPredicates(predicate.NewPredicateFuncs(isSharedEntrancePod))).
		Complete(r)
}

func (r *SharedHostsReconciler) fanOutOnSRR(ctx context.Context, _ client.Object) []reconcile.Request {
	return r.fanOutOptInNamespaces(ctx)
}
func (r *SharedHostsReconciler) fanOutOnPod(ctx context.Context, _ client.Object) []reconcile.Request {
	return r.fanOutOptInNamespaces(ctx)
}
func (r *SharedHostsReconciler) fanOutOnApplication(ctx context.Context, obj client.Object) []reconcile.Request {
	if r == nil || r.Client == nil || obj == nil {
		return nil
	}
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app == nil {
		return nil
	}
	if appcfg.IsSharedServerApp(app) {
		return r.fanOutOptInNamespaces(ctx)
	}
	ns := strings.TrimSpace(app.Spec.Namespace)
	if ns == "" {
		ns = strings.TrimSpace(app.Namespace)
	}
	if ns == "" {
		return nil
	}
	return []reconcile.Request{requestForNS(ns)}
}
func (r *SharedHostsReconciler) requeueNamespace(_ context.Context, obj client.Object) []reconcile.Request {
	if obj == nil {
		return nil
	}
	return []reconcile.Request{requestForNS(obj.GetName())}
}
func (r *SharedHostsReconciler) fanOutOptInNamespaces(ctx context.Context) []reconcile.Request {
	if r == nil || r.Client == nil {
		return nil
	}
	var nsList corev1.NamespaceList
	if err := r.Client.List(ctx, &nsList, client.MatchingLabels{security.NamespaceInClusterCallerLabel: "true"}); err != nil {
		return nil
	}
	out := make([]reconcile.Request, 0, len(nsList.Items))
	for i := range nsList.Items {
		out = append(out, requestForNS(nsList.Items[i].Name))
	}
	return out
}
func requestForNS(ns string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{
		Namespace: ns, Name: constants.D2SharedHostsVolumeName,
	}}
}
