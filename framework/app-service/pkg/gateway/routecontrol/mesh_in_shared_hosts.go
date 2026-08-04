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
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MeshInSharedHostsReconciler keeps olares-mesh-in-shared-hosts ConfigMap in each
// opted-in caller namespace aligned with the per-(caller,viewer) host allow
// set derived from SRR HostPatterns and Application clusterAppRef wiring.
//
// requirement: caller NS opt-in must receive the host allow set so d2 nginx
// njs decideOffload can offload v3 Shared traffic; webhook
// ensureD2SharedHostsPlaceholder writes only an empty placeholder.
// behavior: NS-keyed reconcile with content-hash idempotent Update; NotFound
// waits for the webhook placeholder; GC on NS opt-out, NS delete, or viewer
// gone; managed-by label rules out third-party same-name ConfigMaps.
type MeshInSharedHostsReconciler struct {
	Client client.Client
	// platformDomain pins the domain used for host materialization. Empty means
	// resolve it per reconcile via cluster.GetPlatformDomain; tests set it to
	// avoid the live lookup.
	platformDomain string
}

// Reconcile is the controller-runtime entry. req.Namespace is the caller NS;
// req.Name is fixed to D2SharedHostsVolumeName for dedupe + audit.
func (r *MeshInSharedHostsReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
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
	platformDomain := r.resolvePlatformDomain(ctx)
	if platformDomain == "" {
		// Materializing with an empty domain drops every SRR pattern as
		// non-platform, which would blank an already correct allowlist and send
		// mesh-in back to SNI passthrough. Retry instead of writing that.
		sharedHostsReconcileTotal.WithLabelValues(rResPlatformDomainUnavailable).Inc()
		sharedHostsPlatformDomainReady.Set(0)
		warnPlatformDomainUnavailable(req.Namespace)
		return reconcile.Result{RequeueAfter: sharedHostsPlatformDomainRequeue}, nil
	}
	sharedHostsPlatformDomainReady.Set(1)
	demand, err := BuildSharedHostsDemand(ctx, r.Client, platformDomain)
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

// resolvePlatformDomain returns the domain used to materialize viewer hosts.
// It is resolved per reconcile rather than captured once at manager setup: the
// lookup depends on the cluster owner User carrying bytetrade.io/zone, which is
// not guaranteed to exist yet while app-service starts during a fresh install.
// A one-shot capture that lost that race pinned the reconciler to "" for the
// whole process lifetime. cluster.GetPlatformDomain caches for a short TTL, and
// this reads without mutating shared state, so per-reconcile is cheap and safe.
func (r *MeshInSharedHostsReconciler) resolvePlatformDomain(ctx context.Context) string {
	if d := strings.ToLower(strings.TrimSpace(r.platformDomain)); d != "" {
		return d
	}
	return strings.ToLower(strings.TrimSpace(cluster.GetPlatformDomain(ctx)))
}

// ReconcileNamespace upserts the per-NS olares-mesh-in-shared-hosts ConfigMap.
// fail-safe: List/Get/Update failures leave the existing ConfigMap intact.
func (r *MeshInSharedHostsReconciler) ReconcileNamespace(ctx context.Context, callerNS string, targets []SharedHostsTarget) error {
	if r == nil || r.Client == nil || callerNS == "" {
		return nil
	}
	cm := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: callerNS, Name: constants.MeshInSharedHostsCMName,
	}, cm)
	if apierrors.IsNotFound(err) {
		desiredData := buildSharedHostsConfigMapData(targets)
		desiredHash := sharedHostsContentHash(desiredData)
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.MeshInSharedHostsCMName,
				Namespace: callerNS,
				Labels: map[string]string{
					constants.MeshInSharedHostsManagedByLabel: sharedHostsManagedByValue,
				},
				Annotations: map[string]string{
					sharedHostsContentHashAnnotation:  desiredHash,
					sharedHostsReconciledAtAnnotation: time.Now().UTC().Format(time.RFC3339),
				},
			},
			Data: desiredData,
		}
		if err := r.Client.Create(ctx, cm); err != nil {
			sharedHostsReconcileTotal.WithLabelValues(rResUpdateFailed).Inc()
			return err
		}
		sharedHostsReconcileTotal.WithLabelValues(rResUpdated).Inc()
		updateSharedHostsHashAge(callerNS, desiredHash, true)
		sharedHostsCount.WithLabelValues(hashCallerNS(callerNS)).Set(float64(countSharedHostsRows(targets)))
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
		if key == constants.MeshInSharedHostsFileName || key == constants.MeshInTLSHostsFileName {
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
	cm.Labels[constants.MeshInSharedHostsManagedByLabel] = sharedHostsManagedByValue
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

func (r *MeshInSharedHostsReconciler) gcSharedHostsConfigMap(ctx context.Context, callerNS, reason string) error {
	cm := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: callerNS, Name: constants.MeshInSharedHostsCMName,
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
	authByKey := map[key]map[string]struct{}{}
	tlsByKey := map[key]map[string]struct{}{}
	addViewerHosts := func(callerNS, viewer string) {
		viewer = strings.ToLower(strings.TrimSpace(viewer))
		if viewer == "" {
			return
		}
		k := key{ns: callerNS, viewer: viewer}
		if authByKey[k] == nil {
			authByKey[k] = map[string]struct{}{}
		}
		if tlsByKey[k] == nil {
			tlsByKey[k] = map[string]struct{}{}
		}
		auth, tlsHosts := enumerateHostsForViewer(viewer, srrByOwner[viewer], platformDomain)
		for _, h := range auth {
			authByKey[k][h] = struct{}{}
		}
		for _, h := range tlsHosts {
			tlsByKey[k][h] = struct{}{}
		}
	}
	for i := range nsList.Items {
		callerNS := nsList.Items[i].Name
		for _, app := range appsByNS[callerNS] {
			refs := callerSharedAppRefs(&app)
			if len(refs) == 0 {
				// No named callee refs: still project viewer hosts when the app is a caller.
				if !meshinagent.DeclaresSharedCaller(app.Spec.Settings) {
					continue
				}
				addViewerHosts(callerNS, app.Spec.Owner)
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
				addViewerHosts(callerNS, owner)
			}
		}
	}
	out := make([]SharedHostsTarget, 0, len(authByKey))
	for k, hosts := range authByKey {
		list := make([]string, 0, len(hosts))
		for h := range hosts {
			list = append(list, h)
		}
		sort.Strings(list)
		tlsList := make([]string, 0, len(tlsByKey[k]))
		for h := range tlsByKey[k] {
			tlsList = append(tlsList, h)
		}
		sort.Strings(tlsList)
		out = append(out, SharedHostsTarget{
			CallerNamespace: k.ns,
			Viewer:          k.viewer,
			Hosts:           list,
			TLSHosts:        tlsList,
		})
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
func (r *MeshInSharedHostsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	return builder.ControllerManagedBy(mgr).
		Named("mesh-in-shared-hosts").
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

func (r *MeshInSharedHostsReconciler) fanOutOnSRR(ctx context.Context, _ client.Object) []reconcile.Request {
	return r.fanOutOptInNamespaces(ctx)
}
func (r *MeshInSharedHostsReconciler) fanOutOnPod(ctx context.Context, _ client.Object) []reconcile.Request {
	return r.fanOutOptInNamespaces(ctx)
}
func (r *MeshInSharedHostsReconciler) fanOutOnApplication(ctx context.Context, obj client.Object) []reconcile.Request {
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
func (r *MeshInSharedHostsReconciler) requeueNamespace(_ context.Context, obj client.Object) []reconcile.Request {
	if obj == nil {
		return nil
	}
	return []reconcile.Request{requestForNS(obj.GetName())}
}
func (r *MeshInSharedHostsReconciler) fanOutOptInNamespaces(ctx context.Context) []reconcile.Request {
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
		Namespace: ns, Name: constants.MeshInSharedHostsCMName,
	}}
}
