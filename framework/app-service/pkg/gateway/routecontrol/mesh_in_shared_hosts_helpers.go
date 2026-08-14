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
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	sharedHostsManagedByValue         = "app-service"
	sharedHostsContentHashAnnotation  = "gateway.olares.io/mesh-in-shared-hosts-content-hash"
	sharedHostsReconciledAtAnnotation = "gateway.olares.io/mesh-in-shared-hosts-reconciled-at"
	sharedHostsFileHeader             = "# managed by app-service MeshInSharedHostsReconciler (WI-OC-MESH-IN-CT1-02); do not edit by hand\n" +
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
	// rResPlatformDomainUnavailable means host materialization was skipped and
	// requeued because the platform domain did not resolve.
	rResPlatformDomainUnavailable = "platform_domain_unavailable"
	// drop reason label values.
	rDropOwnerUnresolved = "owner_unresolved"
	rDropMultiRef        = "multi_ref_unsupported"
	rDropNonPlatformHost = "non_platform_host"
	rDropCrossViewerHost = "cross_viewer_host"
	rDropMultiWildcard   = "multi_wildcard"
	rDropInvalidChars    = "invalid_chars"
	rDropEmptyPatterns   = "empty_patterns"
	rDropSharedAuthOnly  = "shared_auth_only"
	// rDropRefUnresolved means a caller named a shared dep that publishes no
	// gateway SRR, so there is nothing to materialize for it.
	rDropRefUnresolved = "shared_ref_unresolved"
	// GC reason label values.
	rGCNSOptOut  = "ns_opt_out"
	rGCNSDeleted = "ns_deleted"
	rGCViewer    = "viewer_gone"
)

const (
	// sharedHostsPlatformDomainRequeue paces retries while the platform domain
	// is unresolvable, so a caller installed during that window converges
	// without operator action.
	sharedHostsPlatformDomainRequeue = 15 * time.Second
	// platformDomainWarnInterval throttles the unavailability warning: the
	// condition fans out over every opt-in namespace on each retry.
	platformDomainWarnInterval = time.Minute
	// emptyAllowlistWarnInterval throttles the empty-allowlist warning per
	// caller namespace; reconciles fan out on every SRR or Application event.
	emptyAllowlistWarnInterval = time.Minute
)

var (
	sharedHostsReconcileTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_mesh_in_shared_hosts_reconcile_total",
		Help: "Count of olares-mesh-in-shared-hosts ConfigMap reconcile outcomes by result.",
	}, []string{"result"})
	sharedHostsDropTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_mesh_in_shared_hosts_drop_total",
		Help: "Count of host derivation drops by reason.",
	}, []string{"reason"})
	sharedHostsGCTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_mesh_in_shared_hosts_gc_total",
		Help: "Count of shared-hosts ConfigMap GC events by reason.",
	}, []string{"reason"})
	sharedHostsTargetCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_mesh_in_shared_hosts_target_count",
		Help: "Current per-namespace (viewer) target count in the demand index.",
	}, []string{"caller_ns"})
	sharedHostsCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_mesh_in_shared_hosts_count",
		Help: "Current per-namespace host row count in shared-hosts.txt.",
	}, []string{"caller_ns"})
	sharedHostsContentHashAgeSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "olares_mesh_in_shared_hosts_content_hash_age_seconds",
		Help: "Seconds since the shared-hosts content hash last changed.",
	}, []string{"caller_ns"})
	sharedHostsPlatformDomainReady = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "olares_mesh_in_shared_hosts_platform_domain_ready",
		Help: "1 when the platform domain resolves for host materialization, 0 while it is unavailable.",
	})
	sharedHostsEmptyAllowlistTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "olares_mesh_in_shared_hosts_empty_allowlist_total",
		Help: "Count of reconciles where a declared caller viewer resolved zero auth hosts.",
	}, []string{"caller_ns"})

	sharedHostsHashStateMu sync.Mutex
	sharedHostsHashState   = map[string]meshInHashSnapshot{}

	platformDomainWarnMu   sync.Mutex
	platformDomainWarnedAt time.Time

	emptyAllowlistWarnMu   sync.Mutex
	emptyAllowlistWarnedAt = map[string]time.Time{}
)

// warnPlatformDomainUnavailable logs at default verbosity, throttled to one line
// per platformDomainWarnInterval. The previous klog.V(2) call sites made this
// failure invisible on a default-verbosity deployment.
func warnPlatformDomainUnavailable(callerNS string) {
	platformDomainWarnMu.Lock()
	defer platformDomainWarnMu.Unlock()
	if !platformDomainWarnedAt.IsZero() && time.Since(platformDomainWarnedAt) < platformDomainWarnInterval {
		return
	}
	platformDomainWarnedAt = time.Now()
	klog.Warningf("shared_hosts: platform domain unavailable, requeue in %s and keep existing allowlist ns=%s",
		sharedHostsPlatformDomainRequeue, hashCallerNS(callerNS))
}

// noteEmptySharedHostsTargets reports caller viewers that opted in but resolved
// no auth host. mesh-in then falls back to passthrough without injecting the
// platform JWT and the callee answers "Jwt is missing", which used to be silent
// on the control plane side. Returns how many targets were reported.
func noteEmptySharedHostsTargets(callerNS string, targets []SharedHostsTarget) int {
	empty := 0
	for _, t := range targets {
		if len(t.Hosts) > 0 {
			continue
		}
		empty++
		sharedHostsEmptyAllowlistTotal.WithLabelValues(hashCallerNS(callerNS)).Inc()
		warnEmptySharedHostsAllowlist(callerNS, t.Viewer)
	}
	return empty
}

func warnEmptySharedHostsAllowlist(callerNS, viewer string) {
	emptyAllowlistWarnMu.Lock()
	defer emptyAllowlistWarnMu.Unlock()
	if at, ok := emptyAllowlistWarnedAt[callerNS]; ok && time.Since(at) < emptyAllowlistWarnInterval {
		return
	}
	emptyAllowlistWarnedAt[callerNS] = time.Now()
	klog.Warningf("shared_hosts: caller declared shared access but allowlist is empty ns=%s viewer=%s",
		hashCallerNS(callerNS), viewer)
}

type meshInHashSnapshot struct {
	hash string
	at   time.Time
}

func init() {
	prometheus.MustRegister(
		sharedHostsReconcileTotal, sharedHostsDropTotal, sharedHostsGCTotal,
		sharedHostsTargetCount, sharedHostsCount, sharedHostsContentHashAgeSeconds,
		sharedHostsPlatformDomainReady, sharedHostsEmptyAllowlistTotal,
	)
}

// SharedHostsTarget is one (callerNS, viewer) → auth/tls host demand entry.
type SharedHostsTarget struct {
	CallerNamespace string
	Viewer          string
	Hosts           []string // auth-hosts (shared-hosts.txt)
	TLSHosts        []string // tls-hosts (tls-hosts.txt); application only
}

func isSharedHostsConfigMap(obj client.Object) bool {
	return obj != nil && obj.GetName() == constants.MeshInSharedHostsCMName
}

// isCustomDomainTLSSecret matches os-gateway CustomDomainTLS Secrets that
// listMaterializedCustomTLSDomains uses to admit exact FQDNs into tls-hosts.
func isCustomDomainTLSSecret(obj client.Object) bool {
	if obj == nil {
		return false
	}
	if obj.GetNamespace() != defaultGatewayNS {
		return false
	}
	if strings.HasPrefix(obj.GetName(), customDomainTLSPrefix) {
		return true
	}
	labels := obj.GetLabels()
	return labels != nil && strings.TrimSpace(labels[labelTLSCustomDomain]) != ""
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
	if len(callerSharedAppRefs(app)) > 0 {
		return true
	}
	return meshinagent.DeclaresSharedCaller(app.Spec.Settings)
}

// callerSharedAppRefs collects Shared callee refs from clusterAppRef / sharedAppDeps
// and Decide edges (annotation or settings). Empty when the app has no Shared edges.
func callerSharedAppRefs(app *appv1alpha1.Application) []string {
	if app == nil {
		return nil
	}
	const (
		settingSharedAppDeps = "sharedAppDeps"
		annotDecideEdges     = "gateway.olares.io/shared-caller-edges"
	)
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		for _, p := range gateway.SplitClusterAppRefs(raw) {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	if app.Spec.Settings != nil {
		add(app.Spec.Settings["clusterAppRef"])
		add(app.Spec.Settings[settingSharedAppDeps])
		add(app.Spec.Settings[annotDecideEdges])
	}
	if app.Annotations != nil {
		add(app.Annotations[annotDecideEdges])
	}
	sort.Strings(out)
	return out
}
func inClusterCallerNamespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool { return hasInClusterCallerLabel(e.Object) },
		UpdateFunc: func(e event.UpdateEvent) bool {
			return hasInClusterCallerLabel(e.ObjectOld) || hasInClusterCallerLabel(e.ObjectNew)
		},
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

func isSharedEntrancePod(obj client.Object) bool {
	if obj == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(obj.GetLabels()[constants.AppSharedEntrancesLabel]), "true")
}
func sharedHostsManagedByUs(cm *corev1.ConfigMap) bool {
	if cm == nil {
		return false
	}
	v, ok := cm.Labels[constants.MeshInSharedHostsManagedByLabel]
	if !ok {
		// Adopt webhook-created placeholder (label not yet set); next Update writes it.
		return true
	}
	return strings.EqualFold(strings.TrimSpace(v), sharedHostsManagedByValue)
}

func enumerateHostsForViewer(viewer string, srrs []srrv1alpha1.SharedRouteRegistry, platformDomain string) (auth, tls []string) {
	authSeen := map[string]struct{}{}
	tlsSeen := map[string]struct{}{}
	domLower := strings.ToLower(strings.TrimSpace(platformDomain))
	for i := range srrs {
		patterns := srrs[i].Spec.HostPatterns
		if len(patterns) == 0 {
			sharedHostsDropTotal.WithLabelValues(rDropEmptyPatterns).Inc()
			continue
		}
		owners := map[string]string{}
		if srrs[i].Annotations != nil {
			owners = gateway.ParseExactHostOwnersJSON(srrs[i].Annotations[gateway.AnnotationExactHostOwners])
		}
		// Empty EntranceClass matches HTTPRoute parent selection: treat as shared.
		allowTLSApp := srrs[i].Spec.EntranceClass == srrv1alpha1.EntranceClassApplication
		for _, pattern := range patterns {
			h, reason := materializeHost(pattern, viewer, platformDomain, owners)
			if h == "" {
				if reason != "" {
					sharedHostsDropTotal.WithLabelValues(reason).Inc()
				}
				continue
			}
			authSeen[h] = struct{}{}
			exactCustom := isExactCustomHost(pattern, domLower)
			if allowTLSApp || exactCustom {
				tlsSeen[h] = struct{}{}
			} else {
				sharedHostsDropTotal.WithLabelValues(rDropSharedAuthOnly).Inc()
			}
		}
	}
	return sortedKeys(authSeen), sortedKeys(tlsSeen)
}

func isExactCustomHost(pattern, platformDomain string) bool {
	p := strings.ToLower(strings.TrimSpace(pattern))
	if p == "" || strings.Contains(p, "*") {
		return false
	}
	return !isPlatformHostGo(p, platformDomain)
}

func materializeHost(pattern, viewer, platformDomain string, owners map[string]string) (string, string) {
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
		return lp.Prefix + "." + viewerLower + "." + lp.PlatformDomain, ""
	}
	p := strings.ToLower(pattern)
	if strings.Contains(p, "*") {
		return "", rDropMultiWildcard
	}
	if !validDNSChars(p) {
		return "", rDropInvalidChars
	}
	// Owned exact hosts (third_party or per-owner third_level) must match viewer
	// even when the FQDN is under platformDomain (e.g. api.alice.olares.com).
	if owner, ok := owners[p]; ok {
		if owner == viewerLower {
			return p, ""
		}
		return "", rDropCrossViewerHost
	}
	// Unowned platform hosts (e.g. hash8.shared.<platform>) stay cluster-shared.
	if isPlatformHostGo(p, domLower) {
		return p, ""
	}
	return "", rDropNonPlatformHost
}

func isPlatformHostGo(host, platformDomain string) bool {
	if platformDomain == "" || host == "" {
		return false
	}
	return strings.HasSuffix(host, "."+platformDomain) && host != "."+platformDomain
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

func buildSharedHostsConfigMapData(targets []SharedHostsTarget, platformDomain string) map[string]string {
	data := map[string]string{}
	perViewer := map[string]map[string]struct{}{}
	allAuth := map[string]struct{}{}
	allTLS := map[string]struct{}{}
	customTLS := map[string]struct{}{}
	dom := strings.ToLower(strings.TrimSpace(platformDomain))
	for _, t := range targets {
		viewer := strings.ToLower(strings.TrimSpace(t.Viewer))
		if viewer == "" || viewer == constants.MeshInSharedHostsFileName {
			continue
		}
		if _, ok := perViewer[viewer]; !ok {
			perViewer[viewer] = map[string]struct{}{}
		}
		for _, h := range t.Hosts {
			allAuth[h] = struct{}{}
			perViewer[viewer][h] = struct{}{}
		}
		for _, h := range t.TLSHosts {
			h = strings.ToLower(strings.TrimSpace(h))
			if h == "" {
				continue
			}
			allTLS[h] = struct{}{}
			// Exact third-party FQDNs must never present the viewer platform cert.
			if dom != "" && !isPlatformHostGo(h, dom) {
				customTLS[h] = struct{}{}
			}
		}
	}
	data[constants.MeshInSharedHostsFileName] = sharedHostsFileText(sortedKeys(allAuth))
	data[constants.MeshInTLSHostsFileName] = sharedHostsFileText(sortedKeys(allTLS))
	data[constants.MeshInCustomTLSHostsFileName] = sharedHostsFileText(sortedKeys(customTLS))
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

// sharedGatewaySRRIndex is the routable menu offered to in-cluster callers: the
// gateway-mode SRRs published by shared server Applications. It is deliberately
// keyed by callee, never by the installer of the shared app -- the installer is
// only one of the users allowed to call it, so indexing by owner left every
// other user with an empty allowlist (DEFECT-SH-N6-OWNER-01).
type sharedGatewaySRRIndex struct {
	// all backs eligibility callers, which declare no named callee and may
	// reach any installed shared app.
	all []srrv1alpha1.SharedRouteRegistry
	// byAppRef narrows the menu for callers with named shared deps. The key is
	// Application spec.name, matching gateway.BuildClusterAppOwnerIndex refs.
	byAppRef map[string][]srrv1alpha1.SharedRouteRegistry
}

func buildSharedGatewaySRRIndex(srrs []srrv1alpha1.SharedRouteRegistry, apps []appv1alpha1.Application) sharedGatewaySRRIndex {
	refsByNS := map[string][]string{}
	for i := range apps {
		app := apps[i]
		if !appcfg.IsSharedServerApp(&app) {
			continue
		}
		ns := strings.TrimSpace(app.Spec.Namespace)
		name := strings.TrimSpace(app.Spec.Name)
		if ns == "" || name == "" {
			continue
		}
		refsByNS[ns] = append(refsByNS[ns], name)
	}
	idx := sharedGatewaySRRIndex{byAppRef: map[string][]srrv1alpha1.SharedRouteRegistry{}}
	for i := range srrs {
		srr := srrs[i]
		if srr.Spec.RouteMode != srrv1alpha1.RouteModeGateway && srr.Spec.RouteMode != "" {
			continue
		}
		// A gateway SRR outside a shared server namespace is a private per-user
		// route; keeping it out bounds the allowlist to shared traffic.
		refs, ok := refsByNS[strings.TrimSpace(srr.Namespace)]
		if !ok {
			continue
		}
		idx.all = append(idx.all, srr)
		for _, ref := range refs {
			idx.byAppRef[ref] = append(idx.byAppRef[ref], srr)
		}
	}
	return idx
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
		state = meshInHashSnapshot{hash: hash, at: now}
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
