package routecontrol

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
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
	labelTLSReplica          = "gateway.olares.io/tls-replica"
	labelReplicaDemandSource = "gateway.olares.io/demand-source"
	nsOwnerLabel             = "bytetrade.io/ns-owner"
	replicaDemandSourceServer = "server"
	replicaDemandSourceCaller = "caller"
	entranceTLSSecretPrefix   = sharedEntranceTLSPrefix
)

var (
	replicaSyncTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "olares_mesh_in_replica_sync_total",
			Help: "Count of entrance TLS replica sync operations by result.",
		},
		[]string{"result", "demand_source"},
	)
	replicaErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "olares_mesh_in_replica_errors_total",
			Help: "Count of entrance TLS replica errors by reason.",
		},
		[]string{"reason", "demand_source"},
	)
	replicaContentHashAgeSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "olares_mesh_in_replica_content_hash_age_seconds",
			Help: "Age in seconds since the replica content hash last changed.",
		},
		[]string{"viewer", "caller_ns", "demand_source"},
	)

	replicaHashStateMu sync.Mutex
	replicaHashState   = map[string]meshInHashSnapshot{}
)

func init() {
	prometheus.MustRegister(replicaSyncTotal, replicaErrorsTotal, replicaContentHashAgeSeconds)
}

// ReplicaTarget identifies one desired TLS secret copy.
type ReplicaTarget struct {
	CallerNamespace string // user-space-<C>
	CertViewer      string // TLS secret suffix <viewer>
	DemandSource    string // server | caller
}

// BuildDemandIndex builds the desired mesh-in TLS replica demand set.
// Destinations are opted-in caller namespaces (gateway.olares.io/in-cluster-caller),
// not Shared server namespaces. CertViewer is the caller NS owner (CT-1 per-viewer
// public cert), matching webhook CertsVolumeForViewer(OwnerName).
func BuildDemandIndex(ctx context.Context, c client.Client, platformDomain string) ([]ReplicaTarget, error) {
	if c == nil {
		return nil, nil
	}
	_ = platformDomain

	var nsList corev1.NamespaceList
	if err := c.List(ctx, &nsList, client.MatchingLabels{
		security.NamespaceInClusterCallerLabel: "true",
	}); err != nil {
		recordReplicaError("index_failed", replicaDemandSourceCaller)
		return nil, err
	}

	var appList appv1alpha1.ApplicationList
	if err := c.List(ctx, &appList); err != nil {
		recordReplicaError("index_failed", replicaDemandSourceCaller)
		return nil, err
	}

	demandSet := make(map[string]ReplicaTarget)
	for i := range nsList.Items {
		ns := &nsList.Items[i]
		callerNS := strings.TrimSpace(ns.Name)
		if callerNS == "" {
			continue
		}
		viewer := certViewerForCallerNamespace(ns, appList.Items)
		if viewer == "" {
			recordReplicaError("caller_viewer_unresolved", replicaDemandSourceCaller)
			continue
		}
		addReplicaTarget(demandSet, ReplicaTarget{
			CallerNamespace: callerNS,
			CertViewer:      viewer,
			DemandSource:    replicaDemandSourceCaller,
		})
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

// certViewerForCallerNamespace picks the CT-1 TLS replica suffix for a caller NS:
// bytetrade.io/ns-owner, else Application Spec.Owner in that NS, else user-space-*.
func certViewerForCallerNamespace(ns *corev1.Namespace, apps []appv1alpha1.Application) string {
	if ns == nil {
		return ""
	}
	if owner := strings.ToLower(strings.TrimSpace(ns.Labels[nsOwnerLabel])); owner != "" {
		return owner
	}
	callerNS := strings.TrimSpace(ns.Name)
	for i := range apps {
		app := &apps[i]
		appNS := strings.TrimSpace(app.Spec.Namespace)
		if appNS == "" {
			appNS = strings.TrimSpace(app.Namespace)
		}
		if appNS != callerNS {
			continue
		}
		if owner := strings.ToLower(strings.TrimSpace(app.Spec.Owner)); owner != "" {
			return owner
		}
	}
	if viewer, ok := viewerFromUserSpaceNS(callerNS); ok {
		return viewer
	}
	return ""
}

// ReconcileReplica reconciles one replica target from app-gateway source Secret.
func ReconcileReplica(ctx context.Context, c client.Client, target ReplicaTarget) (bool, error) {
	if c == nil || target.CallerNamespace == "" || target.CertViewer == "" {
		return false, nil
	}

	src := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      sharedEntranceTLSName(target.CertViewer),
	}, src)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			recordReplicaError("source_get_failed", target.DemandSource)
			return false, err
		}

		dst := &corev1.Secret{}
		err = c.Get(ctx, types.NamespacedName{
			Namespace: target.CallerNamespace,
			Name:      meshInTLSSecretName(target.CertViewer),
		}, dst)
		if apierrors.IsNotFound(err) {
			replicaSyncTotal.WithLabelValues("noop", metricDemandSource(target.DemandSource)).Inc()
			return false, nil
		}
		if err != nil {
			recordReplicaError("replica_get_failed", target.DemandSource)
			return false, err
		}
		if err := c.Delete(ctx, dst); err != nil && !apierrors.IsNotFound(err) {
			recordReplicaError("replica_delete_failed", target.DemandSource)
			return false, err
		}
		replicaSyncTotal.WithLabelValues("deleted", metricDemandSource(target.DemandSource)).Inc()
		return true, nil
	}

	callerNS := &corev1.Namespace{}
	if err := c.Get(ctx, types.NamespacedName{Name: target.CallerNamespace}, callerNS); err != nil {
		recordReplicaError("namespace_get_failed", target.DemandSource)
		return false, err
	}

	dst := &corev1.Secret{}
	err = c.Get(ctx, types.NamespacedName{
		Namespace: target.CallerNamespace,
		Name:      meshInTLSSecretName(target.CertViewer),
	}, dst)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			recordReplicaError("replica_get_failed", target.DemandSource)
			return false, err
		}
		dst = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      meshInTLSSecretName(target.CertViewer),
				Namespace: target.CallerNamespace,
			},
		}
	}

	wrote, err := applyReplicaPatch(ctx, c, src, dst, callerNS, target.DemandSource)
	if err != nil {
		recordReplicaError("replica_apply_failed", target.DemandSource)
		return false, err
	}
	if wrote {
		if dst.ResourceVersion == "" {
			replicaSyncTotal.WithLabelValues("created", metricDemandSource(target.DemandSource)).Inc()
		} else {
			replicaSyncTotal.WithLabelValues("updated", metricDemandSource(target.DemandSource)).Inc()
		}
		return true, nil
	}
	replicaSyncTotal.WithLabelValues("noop", metricDemandSource(target.DemandSource)).Inc()
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
		_ = changed
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
			replicaSyncTotal.WithLabelValues("deleted", metricDemandSource(sec.Labels[labelReplicaDemandSource])).Inc()
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("sync replicas for viewer %q failed on %d target(s): %v", certViewer, len(errs), errs)
}

// applyReplicaPatch is the shared replica writer for reconcile and fan-out paths.
func applyReplicaPatch(ctx context.Context, c client.Client, src, dst *corev1.Secret, callerNS *corev1.Namespace, demandSource string) (bool, error) {
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
		updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, demandSource, false)
		return false, nil
	}

	desiredLabels := copyStringMap(src.Labels)
	desiredLabels[labelTLSReplica] = "true"
	desiredLabels[labelReplicaDemandSource] = metricDemandSource(demandSource)
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

	if dst.Name == "" {
		dst.Name = meshInTLSSecretName(strings.TrimSpace(src.Labels[labelTLSViewer]))
		if dst.Name == constants.MeshInTLSSecretNamePrefix {
			dst.Name = meshInTLSSecretName(strings.TrimPrefix(src.Name, sharedEntranceTLSPrefix))
		}
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
		updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, demandSource, true)
		return true, nil
	}
	if err := c.Update(ctx, dst); err != nil {
		return false, err
	}
	updateReplicaHashAge(dst.Labels[labelTLSViewer], dst.Namespace, hash, demandSource, true)
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
		recordReplicaError("gc_list_failed", replicaDemandSourceServer)
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
			recordReplicaError("gc_delete_failed", metricDemandSource(sec.Labels[labelReplicaDemandSource]))
			return err
		}
		deleted++
	}
	if deleted > 0 {
		replicaSyncTotal.WithLabelValues("gc_periodic", replicaDemandSourceServer).Add(float64(deleted))
	}
	return nil
}

func addReplicaTarget(set map[string]ReplicaTarget, target ReplicaTarget) {
	if set == nil {
		return
	}
	if target.CallerNamespace == "" || target.CertViewer == "" {
		return
	}
	target.DemandSource = metricDemandSource(target.DemandSource)
	key := replicaTargetKey(target.CallerNamespace, target.CertViewer)
	if existing, ok := set[key]; ok {
		if existing.DemandSource == replicaDemandSourceServer {
			return
		}
	}
	set[key] = target
}

func replicaTargetKey(callerNS, viewer string) string {
	return callerNS + "|" + viewer
}

func viewerFromSecretName(name string) (string, bool) {
	name = strings.TrimSpace(name)
	for _, prefix := range []string{constants.MeshInTLSSecretNamePrefix, entranceTLSSecretPrefix} {
		if strings.HasPrefix(name, prefix) {
			viewer := strings.TrimSpace(strings.TrimPrefix(name, prefix))
			if viewer != "" {
				return viewer, true
			}
		}
	}
	return "", false
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

func recordReplicaError(reason, demandSource string) {
	if strings.TrimSpace(reason) == "" {
		reason = "unknown"
	}
	replicaErrorsTotal.WithLabelValues(reason, metricDemandSource(demandSource)).Inc()
}

func updateReplicaHashAge(viewer, callerNS, hash, demandSource string, changed bool) {
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
		state = meshInHashSnapshot{hash: hash, at: now}
		replicaHashState[key] = state
	}
	age := now.Sub(state.at).Seconds()
	replicaHashStateMu.Unlock()

	if age < 0 {
		age = 0
	}
	replicaContentHashAgeSeconds.WithLabelValues(viewer, callerNS, metricDemandSource(demandSource)).Set(age)
}

func metricDemandSource(source string) string {
	switch strings.ToLower(strings.TrimSpace(source)) {
	case replicaDemandSourceCaller:
		return replicaDemandSourceCaller
	case replicaDemandSourceServer:
		return replicaDemandSourceServer
	default:
		return replicaDemandSourceServer
	}
}

func meshInTLSSecretName(viewer string) string {
	return constants.MeshInTLSSecretNamePrefix + strings.ToLower(strings.TrimSpace(viewer))
}
