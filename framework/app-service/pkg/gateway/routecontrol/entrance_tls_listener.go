package routecontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
)

const (
	// entranceTLSListenerFieldManager is the dedicated Server-Side Apply field
	// manager that co-manages the Gateway .spec.listeners by listMapKey=name.
	// The app-gateway chart owns the listener list except the certificateRefs
	// of the https listener and the per-viewer https-<section> listeners.
	entranceTLSListenerFieldManager = "app-service-entrance-tls-listener"

	// gatewayHTTPSSection is the chart listener reused by the first/sticky viewer.
	gatewayHTTPSSection = "https"

	// gatewayHTTPSListenerPrefix names the per-viewer listeners appended for the
	// second and subsequent viewers (https-<section>).
	gatewayHTTPSListenerPrefix = "https-"

	gatewayListenerPort = int64(443)
	gatewayHTTPSProto   = "HTTPS"
	gatewayTLSTerminate = "Terminate"

	gatewayAPIGroup   = "gateway.networking.k8s.io"
	gatewayAPIVersion = "v1"
	gatewayAPIKind    = "Gateway"

	shortHashLen = 8
)

// dns1123SectionRE matches a value already usable as a Gateway SectionName /
// DNS-1123 subdomain (lowercase alphanumeric, '-' and '.', alphanumeric ends).
var dns1123SectionRE = regexp.MustCompile(`^[a-z0-9]([-a-z0-9.]*[a-z0-9])?$`)

// EntranceTLSListenerReconciler binds per-viewer shared-entrance-tls-<viewer>
// Secrets to the Lite Gateway as per-hostname HTTPS listeners.
//
// requirement: replace the placeholder app-gateway-tls certRef with real
// per-viewer certificates so the Lite :443 listener can be PROGRAMMED.
// behavior: every viewer gets its own filter chain (deterministic, non-draining)
// keyed by SNI hostname *.<viewer>.<platformDomain>; only runs on meshProfile=lite.
type EntranceTLSListenerReconciler struct {
	Client client.Client
}

// Reconcile rebuilds the managed Gateway listeners from the current set of
// per-viewer entrance TLS Secrets. The request is ignored: a single Gateway is
// always recomputed from the full Secret set.
func (r *EntranceTLSListenerReconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}

	snap, _ := cluster.GetSnapshot(ctx)
	if !strings.EqualFold(snap.MeshProfile, cluster.MeshProfileLite) {
		// Gate: full / absent meshProfile is a no-op (D4).
		return reconcile.Result{}, nil
	}
	domain := strings.TrimSpace(snap.PlatformDomain)
	if domain == "" {
		// fail-closed: never write a truncated hostname (R-M3).
		klog.Warning("entrance-tls-listener: platformDomain empty, requeue without writing listeners")
		return reconcile.Result{Requeue: true}, nil
	}

	viewers, err := r.listViewers(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	current := newGatewayUnstructured()
	err = r.Client.Get(ctx, types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}, current)
	if apierrors.IsNotFound(err) {
		// Gateway not yet created by the chart; the Gateway watch re-triggers on create.
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	managed := buildDesiredListeners(parseCurrentListeners(current), viewers, domain)
	if err := r.applyListeners(ctx, managed); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// listViewers returns the sorted, de-duplicated viewer set backed by primary
// shared-entrance-tls-<viewer> Secrets in the gateway namespace (replicas in
// caller namespaces are excluded).
func (r *EntranceTLSListenerReconciler) listViewers(ctx context.Context) ([]string, error) {
	var list corev1.SecretList
	if err := r.Client.List(ctx, &list, client.InNamespace(defaultGatewayNS)); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var viewers []string
	for i := range list.Items {
		s := &list.Items[i]
		if s.Labels[labelTLSReplica] == "true" {
			continue
		}
		viewer, ok := viewerFromEntranceTLSSecretName(s.Name)
		if !ok {
			continue
		}
		if _, dup := seen[viewer]; dup {
			continue
		}
		seen[viewer] = struct{}{}
		viewers = append(viewers, viewer)
	}
	sort.Strings(viewers)
	return viewers, nil
}

// managedListener is one Gateway listener owned by this reconciler.
type managedListener struct {
	Name     string
	Hostname string
	// Viewer is the viewer whose Secret backs this listener; "" means the
	// listener keeps an empty certRef (tolerated not-PROGRAMMED, R-M2).
	Viewer string
}

// currentListener is the minimal projection of an existing Gateway listener
// needed to decide ownership: its name and the viewer its certRef points at.
type currentListener struct {
	name       string
	certViewer string
}

// buildDesiredListeners computes the managed listener set (the https listener
// plus per-viewer https-<section> listeners) from the current Gateway state and
// the live viewer set. It is deterministic and non-draining:
//   - the viewer currently bound to the https listener stays bound (sticky);
//   - viewers that already own a dedicated listener keep it;
//   - the https listener is reused by the first unassigned viewer when free,
//     else emptied (never restoring the app-gateway-tls placeholder).
func buildDesiredListeners(current []currentListener, viewers []string, domain string) []managedListener {
	viewerSet := make(map[string]struct{}, len(viewers))
	for _, v := range viewers {
		viewerSet[v] = struct{}{}
	}

	httpsOwner := ""
	dedicated := map[string]struct{}{}
	for _, c := range current {
		if c.certViewer == "" {
			continue
		}
		if _, live := viewerSet[c.certViewer]; !live {
			continue
		}
		switch {
		case c.name == gatewayHTTPSSection:
			httpsOwner = c.certViewer
		case strings.HasPrefix(c.name, gatewayHTTPSListenerPrefix):
			dedicated[c.certViewer] = struct{}{}
		}
	}

	assigned := map[string]struct{}{}
	if httpsOwner != "" {
		assigned[httpsOwner] = struct{}{}
	}
	for v := range dedicated {
		assigned[v] = struct{}{}
	}

	var unassigned []string
	for _, v := range viewers {
		if _, ok := assigned[v]; !ok {
			unassigned = append(unassigned, v)
		}
	}
	sort.Strings(unassigned)

	result := make([]managedListener, 0, len(viewers)+1)

	// The https listener: keep the sticky owner, else adopt the first
	// unassigned viewer, else empty the certRef (R-M2).
	switch {
	case httpsOwner != "":
		result = append(result, managedListener{
			Name:     gatewayHTTPSSection,
			Hostname: wildcardHostname(httpsOwner, domain),
			Viewer:   httpsOwner,
		})
	case len(unassigned) > 0:
		owner := unassigned[0]
		unassigned = unassigned[1:]
		result = append(result, managedListener{
			Name:     gatewayHTTPSSection,
			Hostname: wildcardHostname(owner, domain),
			Viewer:   owner,
		})
	default:
		result = append(result, managedListener{Name: gatewayHTTPSSection})
	}

	// Dedicated per-viewer listeners: existing ones are re-affirmed and newly
	// unassigned viewers are appended. Section names are a pure function of the
	// viewer so a listener never gets renamed when the viewer set changes.
	perViewer := make([]string, 0, len(viewers))
	for v := range dedicated {
		perViewer = append(perViewer, v)
	}
	perViewer = append(perViewer, unassigned...)
	sort.Strings(perViewer)
	for _, v := range perViewer {
		result = append(result, managedListener{
			Name:     gatewayHTTPSListenerPrefix + listenerSectionForViewer(v),
			Hostname: wildcardHostname(v, domain),
			Viewer:   v,
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

// applyListeners writes the managed listeners via Server-Side Apply under the
// dedicated field manager, co-managing .spec.listeners by listMapKey=name. The
// chart-owned http listener is never included and stays untouched.
func (r *EntranceTLSListenerReconciler) applyListeners(ctx context.Context, managed []managedListener) error {
	listeners := make([]any, 0, len(managed))
	for _, m := range managed {
		listeners = append(listeners, m.toMap())
	}
	obj := newGatewayUnstructured()
	obj.SetName(defaultGatewayName)
	obj.SetNamespace(defaultGatewayNS)
	if err := unstructured.SetNestedSlice(obj.Object, listeners, "spec", "listeners"); err != nil {
		return err
	}
	return r.Client.Patch(ctx, obj, client.Apply,
		client.FieldOwner(entranceTLSListenerFieldManager), client.ForceOwnership)
}

func (m managedListener) toMap() map[string]any {
	certRefs := []any{}
	if m.Viewer != "" {
		certRefs = append(certRefs, map[string]any{
			"kind":  "Secret",
			"group": "",
			"name":  entranceTLSSecretName(m.Viewer),
		})
	}
	listener := map[string]any{
		"name":     m.Name,
		"protocol": gatewayHTTPSProto,
		"port":     gatewayListenerPort,
		"allowedRoutes": map[string]any{
			"namespaces": map[string]any{"from": "All"},
		},
		"tls": map[string]any{
			"mode":            gatewayTLSTerminate,
			"certificateRefs": certRefs,
		},
	}
	if m.Hostname != "" {
		listener["hostname"] = m.Hostname
	}
	return listener
}

// parseCurrentListeners projects the existing Gateway .spec.listeners into the
// minimal name/certViewer form used by buildDesiredListeners.
func parseCurrentListeners(u *unstructured.Unstructured) []currentListener {
	if u == nil {
		return nil
	}
	raw, found, err := unstructured.NestedSlice(u.Object, "spec", "listeners")
	if err != nil || !found {
		return nil
	}
	out := make([]currentListener, 0, len(raw))
	for _, item := range raw {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := entry["name"].(string)
		if name == "" {
			continue
		}
		out = append(out, currentListener{
			name:       name,
			certViewer: certViewerFromListener(entry),
		})
	}
	return out
}

// certViewerFromListener returns the viewer behind the first certificateRef of
// a listener, or "" when the ref is the placeholder / unset.
func certViewerFromListener(entry map[string]any) string {
	tls, ok := entry["tls"].(map[string]any)
	if !ok {
		return ""
	}
	refs, ok := tls["certificateRefs"].([]any)
	if !ok || len(refs) == 0 {
		return ""
	}
	ref, ok := refs[0].(map[string]any)
	if !ok {
		return ""
	}
	name, _ := ref["name"].(string)
	viewer, ok := viewerFromEntranceTLSSecretName(name)
	if !ok {
		return ""
	}
	return viewer
}

// viewerFromEntranceTLSSecretName extracts <viewer> from
// shared-entrance-tls-<viewer>.
func viewerFromEntranceTLSSecretName(name string) (string, bool) {
	if !strings.HasPrefix(name, entranceTLSSecretPrefix) {
		return "", false
	}
	viewer := strings.TrimSpace(strings.TrimPrefix(name, entranceTLSSecretPrefix))
	if viewer == "" {
		return "", false
	}
	return viewer, true
}

// listenerSectionForViewer derives a stable, DNS-1123 safe SectionName suffix
// for a viewer. The result depends only on the viewer string (never on the
// viewer set) so listeners are never renamed, and a short content hash is added
// whenever sanitisation would otherwise risk a collision (R-m2).
func listenerSectionForViewer(viewer string) string {
	v := strings.ToLower(strings.TrimSpace(viewer))
	if dns1123SectionRE.MatchString(v) {
		return v
	}
	sanitized := sanitizeDNS1123(v)
	hash := shortHash(viewer)
	if sanitized == "" {
		return hash
	}
	return sanitized + "-" + hash
}

func sanitizeDNS1123(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := b.String()
	out = strings.Trim(out, "-.")
	return out
}

func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:shortHashLen]
}

func wildcardHostname(viewer, domain string) string {
	return "*." + viewer + "." + domain
}

func newGatewayUnstructured() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gatewayGVK())
	return u
}

func gatewayGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: gatewayAPIGroup, Version: gatewayAPIVersion, Kind: gatewayAPIKind}
}
