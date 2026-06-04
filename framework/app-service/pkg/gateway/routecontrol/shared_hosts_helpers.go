package routecontrol

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	sharedHostsManagedByValue         = "app-service"
	sharedHostsContentHashAnnotation  = "gateway.olares.io/d2-shared-hosts-content-hash"
	sharedHostsReconciledAtAnnotation = "gateway.olares.io/d2-shared-hosts-reconciled-at"
	sharedHostsFileHeader             = "# managed by app-service SharedHostsReconciler (WI-N6); do not edit by hand\n" +
		"# format: lowercase host per line; ignored: empty lines, '#' comments\n"
	// reconcile result label values.
	rResUpdated          = "updated"
	rResSkipped          = "skipped"
	rResDeleted          = "deleted"
	rResNotFound         = "not_found_wait_placeholder"
	rResListFailed       = "list_failed"
	rResGetFailed        = "get_failed"
	rResUpdateFailed     = "update_failed"
	rResUpdateConflict   = "update_conflict_retried"
	rResSkippedUnmanaged = "skipped_unmanaged"
	// drop reason label values.
	rDropOwnerUnresolved  = "owner_unresolved"
	rDropMultiRef         = "multi_ref_unsupported"
	rDropNonPlatformHost  = "non_platform_host"
	rDropV2PatternGuarded = "v2_pattern_guarded"
	rDropMultiWildcard    = "multi_wildcard"
	rDropInvalidChars     = "invalid_chars"
	rDropEmptyPatterns    = "empty_patterns"
	// GC reason label values.
	rGCNSOptOut  = "ns_opt_out"
	rGCNSDeleted = "ns_deleted"
	rGCViewer    = "viewer_gone"
)

var (
	sharedHostsReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_d2_shared_hosts_reconcile_total",
		Help: "Count of olares-d2-shared-hosts ConfigMap reconcile outcomes by result.",
	}, []string{"result"})
	sharedHostsDropTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_d2_shared_hosts_drop_total",
		Help: "Count of host derivation drops by reason.",
	}, []string{"reason"})
	sharedHostsGCTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_d2_shared_hosts_gc_total",
		Help: "Count of shared-hosts ConfigMap GC events by reason.",
	}, []string{"reason"})
	sharedHostsTargetCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_d2_shared_hosts_target_count",
		Help: "Current per-namespace (viewer) target count in the demand index.",
	}, []string{"caller_ns"})
	sharedHostsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_d2_shared_hosts_count",
		Help: "Current per-namespace host row count in shared-hosts.txt.",
	}, []string{"caller_ns"})
	sharedHostsContentHashAgeSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_d2_shared_hosts_content_hash_age_seconds",
		Help: "Seconds since the shared-hosts content hash last changed.",
	}, []string{"caller_ns"})

	sharedHostsHashStateMu sync.Mutex
	sharedHostsHashState   = map[string]hashSnapshot{}
)

func init() {
	prometheus.MustRegister(
		sharedHostsReconcileTotal, sharedHostsDropTotal, sharedHostsGCTotal,
		sharedHostsTargetCount, sharedHostsCount, sharedHostsContentHashAgeSeconds,
	)
}

// SharedHostsTarget is one (callerNS, viewer) -> hosts demand entry.
type SharedHostsTarget struct {
	CallerNamespace string
	Viewer          string
	Hosts           []string
}

func isSharedHostsConfigMap(obj client.Object) bool {
	return obj != nil && obj.GetName() == constants.D2SharedHostsVolumeName
}
func isGatewayModeSRR(obj client.Object) bool {
	srr, ok := obj.(*srrv1alpha1.SharedRouteRegistry)
	if !ok || srr == nil {
		return false
	}
	return srr.Spec.RouteMode == srrv1alpha1.RouteModeGateway || srr.Spec.RouteMode == ""
}
func isClusterScopedOrCallerApp(obj client.Object) bool {
	app, ok := obj.(*appv1alpha1.Application)
	if !ok || app == nil {
		return false
	}
	if appcfg.IsSharedServerApp(app) {
		return true
	}
	return strings.TrimSpace(app.Spec.Settings["clusterAppRef"]) != ""
}
func inClusterCallerNamespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return hasInClusterCallerLabel(e.Object) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return hasInClusterCallerLabel(e.ObjectOld) || hasInClusterCallerLabel(e.ObjectNew) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return hasInClusterCallerLabel(e.Object) },
		GenericFunc: func(e event.GenericEvent) bool { return hasInClusterCallerLabel(e.Object) },
	}
}
func hasInClusterCallerLabel(obj client.Object) bool {
	if obj == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(obj.GetLabels()[security.NamespaceInClusterCallerLabel]), "true")
}
func namespaceOptedIntoSharedHosts(ns *corev1.Namespace) bool {
	if ns == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(ns.Labels[security.NamespaceInClusterCallerLabel]), "true")
}
func sharedHostsManagedByUs(cm *corev1.ConfigMap) bool {
	if cm == nil {
		return false
	}
	v, ok := cm.Labels[constants.D2SharedHostsManagedByLabel]
	if !ok {
		// Adopt webhook-created placeholder (label not yet set); next Update writes it.
		return true
	}
	return strings.EqualFold(strings.TrimSpace(v), sharedHostsManagedByValue)
}

func enumerateHostsForViewer(viewer string, srrs []srrv1alpha1.SharedRouteRegistry, platformDomain string) []string {
	seen := map[string]struct{}{}
	for i := range srrs {
		patterns := srrs[i].Spec.HostPatterns
		if len(patterns) == 0 {
			sharedHostsDropTotal.WithLabelValues(rDropEmptyPatterns).Inc()
			continue
		}
		for _, pattern := range patterns {
			h, reason := materializeHost(pattern, viewer, platformDomain)
			if h == "" {
				if reason != "" {
					sharedHostsDropTotal.WithLabelValues(reason).Inc()
				}
				continue
			}
			seen[h] = struct{}{}
		}
	}
	return sortedKeys(seen)
}

func materializeHost(pattern, viewer, platformDomain string) (string, string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", rDropEmptyPatterns
	}
	domLower := strings.ToLower(strings.TrimSpace(platformDomain))
	viewerLower := strings.ToLower(strings.TrimSpace(viewer))
	if viewerLower == "" {
		return "", rDropOwnerUnresolved
	}
	if lp, ok := ParseLogicalPattern(pattern); ok {
		if lp.PlatformDomain != domLower {
			return "", rDropNonPlatformHost
		}
		h := lp.Hash8 + "." + viewerLower + "." + lp.PlatformDomain
		if matchesV2GuardGo(h, lp.PlatformDomain) {
			return "", rDropV2PatternGuarded
		}
		return h, ""
	}
	p := strings.ToLower(pattern)
	if strings.Contains(p, "*") {
		return "", rDropMultiWildcard
	}
	if !validDNSChars(p) {
		return "", rDropInvalidChars
	}
	if !isPlatformHostGo(p, domLower) {
		return "", rDropNonPlatformHost
	}
	if matchesV2GuardGo(p, domLower) {
		return "", rDropV2PatternGuarded
	}
	return p, ""
}

func isPlatformHostGo(host, platformDomain string) bool {
	if platformDomain == "" || host == "" {
		return false
	}
	return strings.HasSuffix(host, "."+platformDomain) && host != "."+platformDomain
}
func matchesV2GuardGo(host, platformDomain string) bool {
	if !isPlatformHostGo(host, platformDomain) {
		return false
	}
	rest := strings.TrimSuffix(host, "."+platformDomain)
	parts := strings.Split(rest, ".")
	return len(parts) == 2 && parts[1] == "shared"
}
func validDNSChars(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '.' || r == '-':
		default:
			return false
		}
	}
	return true
}

func buildSharedHostsConfigMapData(targets []SharedHostsTarget) map[string]string {
	data := map[string]string{}
	perViewer := map[string]map[string]struct{}{}
	all := map[string]struct{}{}
	for _, t := range targets {
		viewer := strings.ToLower(strings.TrimSpace(t.Viewer))
		if viewer == "" || viewer == constants.D2SharedHostsFileName {
			continue
		}
		if _, ok := perViewer[viewer]; !ok {
			perViewer[viewer] = map[string]struct{}{}
		}
		for _, h := range t.Hosts {
			all[h] = struct{}{}
			perViewer[viewer][h] = struct{}{}
		}
	}
	data[constants.D2SharedHostsFileName] = sharedHostsFileText(sortedKeys(all))
	for viewer, set := range perViewer {
		data[viewer] = sharedHostsFileText(sortedKeys(set))
	}
	return data
}
func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func sharedHostsFileText(hosts []string) string {
	var b strings.Builder
	b.WriteString(sharedHostsFileHeader)
	for _, h := range hosts {
		b.WriteString(h)
		b.WriteByte('\n')
	}
	return b.String()
}
func sharedHostsContentHash(data map[string]string) string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write([]byte(data[k]))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
func countSharedHostsRows(targets []SharedHostsTarget) int {
	all := map[string]struct{}{}
	for _, t := range targets {
		for _, h := range t.Hosts {
			all[h] = struct{}{}
		}
	}
	return len(all)
}

func buildNamespaceOwnerIndex(apps []appv1alpha1.Application) map[string]string {
	idx := map[string]string{}
	for i := range apps {
		app := apps[i]
		if !appcfg.IsSharedServerApp(&app) {
			continue
		}
		ns := strings.TrimSpace(app.Spec.Namespace)
		owner := strings.ToLower(strings.TrimSpace(app.Spec.Owner))
		if ns == "" || owner == "" {
			continue
		}
		idx[ns] = owner
	}
	return idx
}
func groupSRRByOwner(srrs []srrv1alpha1.SharedRouteRegistry, nsOwnerIdx map[string]string) map[string][]srrv1alpha1.SharedRouteRegistry {
	out := map[string][]srrv1alpha1.SharedRouteRegistry{}
	for i := range srrs {
		srr := srrs[i]
		if srr.Spec.RouteMode != srrv1alpha1.RouteModeGateway && srr.Spec.RouteMode != "" {
			continue
		}
		viewer := ""
		if v, ok := viewerFromUserSpaceNamespace(srr.Namespace); ok {
			viewer = v
		} else if owner, ok := nsOwnerIdx[srr.Namespace]; ok {
			viewer = owner
		}
		viewer = strings.ToLower(strings.TrimSpace(viewer))
		if viewer == "" {
			continue
		}
		out[viewer] = append(out[viewer], srr)
	}
	return out
}

func hashCallerNS(ns string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(ns)))
	return hex.EncodeToString(sum[:8])
}
func updateSharedHostsHashAge(callerNS, hash string, changed bool) {
	callerNS = strings.TrimSpace(callerNS)
	if callerNS == "" {
		return
	}
	now := time.Now()
	sharedHostsHashStateMu.Lock()
	state, ok := sharedHostsHashState[callerNS]
	if !ok || changed || state.hash != hash {
		state = hashSnapshot{hash: hash, at: now}
		sharedHostsHashState[callerNS] = state
	}
	age := now.Sub(state.at).Seconds()
	sharedHostsHashStateMu.Unlock()
	if age < 0 {
		age = 0
	}
	sharedHostsContentHashAgeSeconds.WithLabelValues(hashCallerNS(callerNS)).Set(age)
}
func clearSharedHostsHashState(callerNS string) {
	callerNS = strings.TrimSpace(callerNS)
	if callerNS == "" {
		return
	}
	sharedHostsHashStateMu.Lock()
	delete(sharedHostsHashState, callerNS)
	sharedHostsHashStateMu.Unlock()
	sharedHostsContentHashAgeSeconds.DeleteLabelValues(hashCallerNS(callerNS))
}
