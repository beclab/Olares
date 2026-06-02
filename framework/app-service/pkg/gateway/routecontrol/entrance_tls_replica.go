package routecontrol

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelTLSReplica = "gateway.olares.io/tls-replica"
	nsOwnerLabel    = "bytetrade.io/ns-owner"
)

var (
	replicaSyncTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "olares_d2_replica_sync_total",
			Help: "Count of entrance TLS replica sync operations by result.",
		},
		[]string{"result"},
	)
	replicaErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "olares_d2_replica_errors_total",
			Help: "Count of entrance TLS replica errors by reason.",
		},
		[]string{"reason"},
	)
	replicaContentHashAgeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "olares_d2_replica_content_hash_age_seconds",
			Help: "Age in seconds since the replica content hash last changed.",
		},
		[]string{"viewer", "caller_ns"},
	)

	replicaHashStateMu sync.Mutex
	replicaHashState   = map[string]hashSnapshot{}

	testBuildClusterAppOwnerIndexHook func()
	testResolveClusterAppOwnerHook    func()
)

type hashSnapshot struct {
	hash string
	at   time.Time
}

func init() {
	prometheus.MustRegister(replicaSyncTotal, replicaErrorsTotal, replicaContentHashAgeSeconds)
}

// ReplicaTarget identifies one desired TLS secret copy.
type ReplicaTarget struct {
	CallerNamespace string // user-space-<C>
	CertViewer      string // TLS secret suffix <viewer>
}

// ClusterAppOwnerIndex maps cluster-scoped Application Spec.Name to owner viewer.
type ClusterAppOwnerIndex map[string]string

// BuildDemandIndex builds the desired caller-namespace replica demand set.
func BuildDemandIndex(ctx context.Context, c client.Client, platformDomain string) ([]ReplicaTarget, error) {
	if c == nil {
		return nil, nil
	}
	_ = platformDomain

	var podList corev1.PodList
	if err := c.List(ctx, &podList, client.MatchingLabels{
		constants.AppSharedEntrancesLabel: "true",
	}); err != nil {
		recordReplicaError("index_failed")
		return nil, err
	}

	var appList appv1alpha1.ApplicationList
	if err := c.List(ctx, &appList); err != nil {
		recordReplicaError("index_failed")
		return nil, err
	}
	var nsList corev1.NamespaceList
	if err := c.List(ctx, &nsList); err != nil {
		recordReplicaError("index_failed")
		return nil, err
	}

	ownerIdx := buildClusterAppOwnerIndex(appList.Items)
	nsOwnerIdx := make(map[string]string, len(nsList.Items))
	for i := range nsList.Items {
		ns := nsList.Items[i]
		owner := strings.ToLower(strings.TrimSpace(ns.Labels[nsOwnerLabel]))
		if owner == "" {
			continue
		}
		nsOwnerIdx[strings.TrimSpace(ns.Name)] = owner
	}
	appsByNS := make(map[string][]appv1alpha1.Application)
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

	demandSet := make(map[string]ReplicaTarget)
	for i := range podList.Items {
		pod := podList.Items[i]
		callerNS := strings.TrimSpace(pod.Namespace)
		viewer, ok := viewerFromUserSpaceNamespace(callerNS)
		if !ok {
			owner, hasOwner := nsOwnerIdx[callerNS]
			if hasOwner {
				addReplicaTarget(demandSet, ReplicaTarget{
					CallerNamespace: callerNS,
					CertViewer:      owner,
				})
			}
			continue
		}

		addReplicaTarget(demandSet, ReplicaTarget{
			CallerNamespace: callerNS,
			CertViewer:      viewer,
		})

		for _, app := range appsByNS[callerNS] {
			refs := splitClusterAppRefs(app.Spec.Settings["clusterAppRef"])
			for _, ref := range refs {
				owners := splitClusterAppRefs(resolveClusterAppOwner(ownerIdx, ref))
				if len(owners) == 0 {
					recordReplicaError("app_ref_unresolved")
					continue
				}
				for _, owner := range owners {
					addReplicaTarget(demandSet, ReplicaTarget{
						CallerNamespace: callerNS,
						CertViewer:      owner,
					})
				}
			}
		}
	}

	out := make([]ReplicaTarget, 0, len(demandSet))
	for _, target := range demandSet {
		out = append(out, target)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CallerNamespace == out[j].CallerNamespace {
			return out[i].CertViewer < out[j].CertViewer
		}
		return out[i].CallerNamespace < out[j].CallerNamespace
	})
	return out, nil
}

func buildClusterAppOwnerIndex(apps []appv1alpha1.Application) ClusterAppOwnerIndex {
	if testBuildClusterAppOwnerIndexHook != nil {
		testBuildClusterAppOwnerIndexHook()
	}
	idx := make(ClusterAppOwnerIndex, len(apps))
	for i := range apps {
		app := apps[i]
		if strings.TrimSpace(app.Spec.Settings["clusterScoped"]) != "true" {
			continue
		}
		name := strings.TrimSpace(app.Spec.Name)
		owner := strings.TrimSpace(app.Spec.Owner)
		if name == "" || owner == "" {
			continue
		}
		if prev, ok := idx[name]; ok && prev != owner {
			klog.Warningf("replica.app_ref_multi_owner app=%s owner_old=%s owner_new=%s", name, prev, owner)
			owners := splitClusterAppRefs(prev)
			already := false
			for _, o := range owners {
				if o == owner {
					already = true
					break
				}
			}
			if !already {
				owners = append(owners, owner)
				idx[name] = strings.Join(owners, ",")
			}
			continue
		}
		idx[name] = owner
	}
	return idx
}

func resolveClusterAppOwner(idx ClusterAppOwnerIndex, appRef string) string {
	if testResolveClusterAppOwnerHook != nil {
		testResolveClusterAppOwnerHook()
	}
	if len(idx) == 0 {
		return ""
	}
	return strings.TrimSpace(idx[strings.TrimSpace(appRef)])
}

// ReconcileReplica reconciles one replica target from app-gateway source Secret.
func ReconcileReplica(ctx context.Context, c client.Client, target ReplicaTarget) (bool, error) {
	if c == nil || target.CallerNamespace == "" || target.CertViewer == "" {
		return false, nil
	}

	src := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      entranceTLSSecretName(target.CertViewer),
	}, src)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			recordReplicaError("source_get_failed")
			return false, err
		}

		dst := &corev1.Secret{}
		err = c.Get(ctx, types.NamespacedName{
			Namespace: target.CallerNamespace,
			Name:      entranceTLSSecretName(target.CertViewer),
		}, dst)
		if apierrors.IsNotFound(err) {
			replicaSyncTotal.WithLabelValues("noop").Inc()
			return false, nil
		}
		if err != nil {
			recordReplicaError("replica_get_failed")
			return false, err
		}
		if err := c.Delete(ctx, dst); err != nil && !apierrors.IsNotFound(err) {
			recordReplicaError("replica_delete_failed")
			return false, err
		}
		replicaSyncTotal.WithLabelValues("deleted").Inc()
		return true, nil
	}

	callerNS := &corev1.Namespace{}
	if err := c.Get(ctx, types.NamespacedName{Name: target.CallerNamespace}, callerNS); err != nil {
		recordReplicaError("namespace_get_failed")
		return false, err
	}

	dst := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{
		Namespace: target.CallerNamespace,
		Name:      entranceTLSSecretName(target.CertViewer),
	}, dst)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			recordReplicaError("replica_get_failed")
			return false, err
		}
		dst = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      entranceTLSSecretName(target.CertViewer),
				Namespace: target.CallerNamespace,
			},
		}
	}

	wrote, err := applyReplicaPatch(ctx, c, src, dst, callerNS)
	if err != nil {
		recordReplicaError("replica_apply_failed")
		return false, err
	}
	if wrote {
		if dst.ResourceVersion == "" {
			replicaSyncTotal.WithLabelValues("created").Inc()
		} else {
			replicaSyncTotal.WithLabelValues("updated").Inc()
		}
		return true, nil
	}
	replicaSyncTotal.WithLabelValues("noop").Inc()
	return false, nil
}

// SyncReplicasForViewer reconciles all targets that consume one cert viewer.
func SyncReplicasForViewer(ctx context.Context, c client.Client, certViewer string, index []ReplicaTarget) error {
	desiredNamespaces := map[string]struct{}{}
	var errs []error
	for _, target := range index {
		if target.CertViewer != certViewer {
			continue
		}
		desiredNamespaces[target.CallerNamespace] = struct{}{}
		changed, err := ReconcileReplica(ctx, c, target)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !changed {
			continue
		}
		if err := bumpSharedEntranceWorkloadForReplica(ctx, c, target); err != nil {
			errs = append(errs, err)
		}
	}
	var list corev1.SecretList
	if err := c.List(ctx, &list, client.MatchingLabels{
		labelTLSReplica: "true",
		labelTLSViewer:  certViewer,
	}); err != nil {
		errs = append(errs, err)
	} else {
		for i := range list.Items {
			sec := &list.Items[i]
			if _, ok := desiredNamespaces[sec.Namespace]; ok {
				continue
			}
			if err := c.Delete(ctx, sec); err != nil && !apierrors.IsNotFound(err) {
				errs = append(errs, err)
				continue
			}
			replicaSyncTotal.WithLabelValues("deleted").Inc()
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("sync replicas for viewer %q failed on %d target(s): %v", certViewer, len(errs), errs)
}

// applyReplicaPatch is the shared replica writer for reconcile and fan-out paths.
func applyReplicaPatch(ctx context.Context, c client.Client, src, dst *corev1.Secret, callerNS *corev1.Namespace) (bool, error) {
	if c == nil || src == nil || dst == nil || callerNS == nil {
		return false, nil
	}

	cert := append([]byte(nil), src.Data[corev1.TLSCertKey]...)
	key := append([]byte(nil), src.Data[corev1.TLSPrivateKeyKey]...)
	if len(cert) == 0 || len(key) == 0 {
		err := fmt.Errorf("D2_REPLICA_BAD_SCHEMA source %s/%s missing tls.crt/tls.key", src.Namespace, src.Name)
		klog.Error(err)
		return false, err
	}
	hash := strings.TrimSpace(src.Annotations[annotationTLSContentHash])
	if hash == "" {
		hash = tlsMaterialHash(string(cert), string(key))
	}

	sameHash := dst.Annotations != nil && dst.Annotations[annotationTLSContentHash] == hash
	if sameHash {
		updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, false)
		return false, nil
	}

	desiredLabels := copyStringMap(src.Labels)
	desiredLabels[labelTLSReplica] = "true"
	desiredAnnotations := map[string]string{
		annotationTLSContentHash: hash,
	}
	f := false
	ownerRef := []metav1.OwnerReference{
		{
			APIVersion:         "v1",
			Kind:               "Namespace",
			Name:               callerNS.Name,
			UID:                callerNS.UID,
			Controller:         &f,
			BlockOwnerDeletion: &f,
		},
	}
	desiredData := map[string][]byte{
		corev1.TLSCertKey:       cert,
		corev1.TLSPrivateKeyKey: key,
	}

	dst.Type = corev1.SecretTypeTLS
	dst.Data = desiredData
	dst.StringData = nil
	dst.Labels = desiredLabels
	dst.Annotations = desiredAnnotations
	dst.OwnerReferences = ownerRef

	if err := validateReplicaSecretSchema(dst); err != nil {
		klog.Error(err)
		return false, err
	}

	if dst.ResourceVersion == "" {
		if err := c.Create(ctx, dst); err != nil {
			return false, err
		}
		updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, true)
		return true, nil
	}
	if err := c.Update(ctx, dst); err != nil {
		return false, err
	}
	updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, true)
	return true, nil
}

// sweepOrphanReplicas removes tls-replica Secrets that are not in current demand.
func sweepOrphanReplicas(ctx context.Context, c client.Client, currentDemand []ReplicaTarget) error {
	if c == nil {
		return nil
	}
	demandSet := make(map[string]struct{}, len(currentDemand))
	for _, target := range currentDemand {
		if target.CallerNamespace == "" || target.CertViewer == "" {
			continue
		}
		demandSet[replicaTargetKey(target.CallerNamespace, target.CertViewer)] = struct{}{}
	}

	var list corev1.SecretList
	if err := c.List(ctx, &list, client.MatchingLabels{labelTLSReplica: "true"}); err != nil {
		recordReplicaError("gc_list_failed")
		return err
	}

	deleted := 0
	for i := range list.Items {
		sec := &list.Items[i]
		viewer := strings.TrimSpace(sec.Labels[labelTLSViewer])
		if viewer == "" {
			derived, ok := viewerFromSecretName(sec.Name)
			if !ok {
				continue
			}
			viewer = derived
		}
		if _, keep := demandSet[replicaTargetKey(sec.Namespace, viewer)]; keep {
			continue
		}
		if err := c.Delete(ctx, sec); err != nil && !apierrors.IsNotFound(err) {
			recordReplicaError("gc_delete_failed")
			return err
		}
		deleted++
	}
	if deleted > 0 {
		replicaSyncTotal.WithLabelValues("gc_periodic").Add(float64(deleted))
	}
	return nil
}

func splitClusterAppRefs(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	refs := make([]string, 0, len(parts))
	for _, part := range parts {
		ref := strings.TrimSpace(part)
		if ref == "" {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

func addReplicaTarget(set map[string]ReplicaTarget, target ReplicaTarget) {
	if set == nil {
		return
	}
	if target.CallerNamespace == "" || target.CertViewer == "" {
		return
	}
	set[replicaTargetKey(target.CallerNamespace, target.CertViewer)] = target
}

func replicaTargetKey(callerNS, viewer string) string {
	return callerNS + "|" + viewer
}

func viewerFromSecretName(name string) (string, bool) {
	if !strings.HasPrefix(name, entranceTLSSecretPrefix) {
		return "", false
	}
	viewer := strings.TrimSpace(strings.TrimPrefix(name, entranceTLSSecretPrefix))
	if viewer == "" {
		return "", false
	}
	return viewer, true
}

func validateReplicaSecretSchema(sec *corev1.Secret) error {
	if sec == nil {
		return nil
	}
	if sec.Type != corev1.SecretTypeTLS {
		return fmt.Errorf("D2_REPLICA_BAD_SCHEMA secret %s/%s type=%q", sec.Namespace, sec.Name, sec.Type)
	}
	if len(sec.Data) != 2 {
		return fmt.Errorf("D2_REPLICA_BAD_SCHEMA secret %s/%s data_keys=%d", sec.Namespace, sec.Name, len(sec.Data))
	}
	if _, ok := sec.Data[corev1.TLSCertKey]; !ok {
		return fmt.Errorf("D2_REPLICA_BAD_SCHEMA secret %s/%s missing %s", sec.Namespace, sec.Name, corev1.TLSCertKey)
	}
	if _, ok := sec.Data[corev1.TLSPrivateKeyKey]; !ok {
		return fmt.Errorf("D2_REPLICA_BAD_SCHEMA secret %s/%s missing %s", sec.Namespace, sec.Name, corev1.TLSPrivateKeyKey)
	}
	return nil
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func recordReplicaError(reason string) {
	if strings.TrimSpace(reason) == "" {
		reason = "unknown"
	}
	replicaErrorsTotal.WithLabelValues(reason).Inc()
}

func updateReplicaHashAge(viewer, callerNS, hash string, changed bool) {
	viewer = strings.TrimSpace(viewer)
	callerNS = strings.TrimSpace(callerNS)
	if viewer == "" || callerNS == "" {
		return
	}
	key := replicaTargetKey(callerNS, viewer)
	now := time.Now()

	replicaHashStateMu.Lock()
	state, ok := replicaHashState[key]
	if !ok || changed || state.hash != hash {
		state = hashSnapshot{hash: hash, at: now}
		replicaHashState[key] = state
	}
	age := now.Sub(state.at).Seconds()
	replicaHashStateMu.Unlock()

	if age < 0 {
		age = 0
	}
	replicaContentHashAgeSeconds.WithLabelValues(viewer, callerNS).Set(age)
}
