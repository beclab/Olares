package routecontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	entranceTLSListenerFieldManager = "app-service-entrance-tls-listener"
	entranceTLSRequeueDelay         = 30 * time.Second
	customDomainTLSPrefix           = "shared-entrance-tls-custom-"
	labelTLSCustomDomain            = "gateway.olares.io/tls-custom-domain"
)

var gatewayGVK = schema.GroupVersionKind{Group: "gateway.networking.k8s.io", Version: "v1", Kind: "Gateway"}

// EntranceTLSListenerReconciler maintains per-viewer HTTPS listeners on the
// app-gateway Gateway HTTPS listeners for the lite gateway line.
type EntranceTLSListenerReconciler struct {
	Client client.Client
}

// Reconcile patches Gateway listeners from shared-entrance-tls-* Secrets.
func (r *EntranceTLSListenerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r == nil || r.Client == nil {
		return reconcile.Result{}, nil
	}
	platformDomain := cluster.GetPlatformDomain(ctx)
	if platformDomain == "" {
		klog.V(2).Info("entrance-tls-listener: platform domain empty, requeue")
		return reconcile.Result{RequeueAfter: entranceTLSRequeueDelay}, nil
	}
	viewers, err := listViewerTLSSecrets(ctx, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	custom, err := listCustomDomainTLSSecrets(ctx, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, applyGatewayHTTPSListeners(ctx, r.Client, viewers, custom, platformDomain)
}

type viewerTLS struct {
	Viewer     string
	SecretName string
}

type customDomainTLS struct {
	Domain     string
	SecretName string
}

func listViewerTLSSecrets(ctx context.Context, c client.Client) ([]viewerTLS, error) {
	var secList corev1.SecretList
	if err := c.List(ctx, &secList, client.InNamespace(defaultGatewayNS), client.MatchingLabels{
		ManagedByLabel: ManagedByValue,
	}); err != nil {
		return nil, err
	}
	var out []viewerTLS
	for _, sec := range secList.Items {
		if !strings.HasPrefix(sec.Name, sharedEntranceTLSPrefix) {
			continue
		}
		if strings.HasPrefix(sec.Name, customDomainTLSPrefix) || sec.Labels[labelTLSCustomDomain] != "" {
			continue
		}
		viewer := sec.Labels[labelTLSViewer]
		if viewer == "" {
			viewer = strings.TrimPrefix(sec.Name, sharedEntranceTLSPrefix)
		}
		if viewer == "" || !dns1123Label.MatchString(viewer) {
			continue
		}
		out = append(out, viewerTLS{Viewer: viewer, SecretName: sec.Name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Viewer < out[j].Viewer })
	return out, nil
}

func listCustomDomainTLSSecrets(ctx context.Context, c client.Client) ([]customDomainTLS, error) {
	var secList corev1.SecretList
	if err := c.List(ctx, &secList, client.InNamespace(defaultGatewayNS), client.MatchingLabels{
		ManagedByLabel: ManagedByValue,
	}); err != nil {
		return nil, err
	}
	var out []customDomainTLS
	for _, sec := range secList.Items {
		if !strings.HasPrefix(sec.Name, customDomainTLSPrefix) {
			continue
		}
		domain := sec.Labels[labelTLSCustomDomain]
		if domain == "" {
			domain = sec.Annotations[annotationTLSHostname]
		}
		if domain == "" {
			continue
		}
		out = append(out, customDomainTLS{Domain: domain, SecretName: sec.Name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Domain < out[j].Domain })
	return out, nil
}

func applyGatewayHTTPSListeners(ctx context.Context, c client.Client, viewers []viewerTLS, custom []customDomainTLS, platformDomain string) error {
	gw := &unstructured.Unstructured{}
	gw.SetGroupVersionKind(gatewayGVK)
	key := types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}
	if err := c.Get(ctx, key, gw); err != nil {
		return client.IgnoreNotFound(err)
	}

	current, _, _ := unstructured.NestedSlice(gw.Object, "spec", "listeners")
	preserved := preserveNonHTTPSListeners(current)
	desired := append(preserved, buildHTTPSListeners(viewers, platformDomain)...)
	desired = append(desired, buildCustomDomainHTTPSListeners(custom)...)

	if listenerSlicesEqual(current, desired) {
		return nil
	}
	patch := &unstructured.Unstructured{}
	patch.SetGroupVersionKind(gatewayGVK)
	patch.SetNamespace(key.Namespace)
	patch.SetName(key.Name)
	if err := unstructured.SetNestedSlice(patch.Object, desired, "spec", "listeners"); err != nil {
		return err
	}
	return c.Patch(ctx, patch, client.Apply, client.FieldOwner(entranceTLSListenerFieldManager), client.ForceOwnership)
}

func preserveNonHTTPSListeners(listeners []interface{}) []interface{} {
	var out []interface{}
	for _, raw := range listeners {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _, _ := unstructured.NestedString(m, "name")
		protocol, _, _ := unstructured.NestedString(m, "protocol")
		port, _, _ := unstructured.NestedInt64(m, "port")
		if protocol == "HTTPS" && port == 443 && (name == defaultGatewaySectN || strings.HasPrefix(name, "https-") || strings.HasPrefix(name, "https-custom-")) {
			continue
		}
		out = append(out, raw)
	}
	return out
}

func buildHTTPSListeners(viewers []viewerTLS, platformDomain string) []interface{} {
	if len(viewers) == 0 {
		return []interface{}{emptyHTTPSListener()}
	}
	used := map[string]struct{}{defaultGatewaySectN: {}}
	var out []interface{}
	for i, v := range viewers {
		listenerName := defaultGatewaySectN
		if i > 0 {
			listenerName = uniqueHTTPSListenerName(v.Viewer, used)
		}
		used[listenerName] = struct{}{}
		out = append(out, httpsListenerObject(listenerName, v.Viewer, platformDomain, v.SecretName))
	}
	return out
}

func emptyHTTPSListener() interface{} {
	return httpsListenerObject(defaultGatewaySectN, "", "", "")
}

func httpsListenerObject(name, viewer, platformDomain, secretName string) map[string]interface{} {
	obj := map[string]interface{}{
		"name":     name,
		"protocol": "HTTPS",
		"port":     int64(443),
		"allowedRoutes": map[string]interface{}{
			"namespaces": map[string]interface{}{"from": "All"},
		},
	}
	tls := map[string]interface{}{
		"mode":            "Terminate",
		"certificateRefs": []interface{}{},
	}
	if viewer != "" && platformDomain != "" && secretName != "" {
		obj["hostname"] = fmt.Sprintf("*.%s.%s", viewer, platformDomain)
		tls["certificateRefs"] = []interface{}{
			map[string]interface{}{
				"group": "",
				"kind":  "Secret",
				"name":  secretName,
			},
		}
	}
	obj["tls"] = tls
	return obj
}

func buildCustomDomainHTTPSListeners(custom []customDomainTLS) []interface{} {
	used := map[string]struct{}{}
	var out []interface{}
	for _, cd := range custom {
		name := uniqueCustomHTTPSListenerName(cd.Domain, used)
		used[name] = struct{}{}
		out = append(out, customDomainListenerObject(name, cd.Domain, cd.SecretName))
	}
	return out
}

func customDomainListenerObject(name, domain, secretName string) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"protocol": "HTTPS",
		"port":     int64(443),
		"hostname": domain,
		"allowedRoutes": map[string]interface{}{
			"namespaces": map[string]interface{}{"from": "All"},
		},
		"tls": map[string]interface{}{
			"mode": "Terminate",
			"certificateRefs": []interface{}{
				map[string]interface{}{
					"group": "",
					"kind":  "Secret",
					"name":  secretName,
				},
			},
		},
	}
}

func httpsListenerName(viewer string) string {
	return uniqueHTTPSListenerName(viewer, nil)
}

func uniqueHTTPSListenerName(viewer string, used map[string]struct{}) string {
	base := sanitizeListenerSectionName(viewer)
	candidate := "https-" + base
	if used == nil || !listenerNameTaken(candidate, used) {
		return candidate
	}
	sum := sha256.Sum256([]byte(viewer))
	suffix := hex.EncodeToString(sum[:3])
	return "https-" + base + "-" + suffix
}

func uniqueCustomHTTPSListenerName(domain string, used map[string]struct{}) string {
	base := sanitizeListenerSectionName(domain)
	candidate := "https-custom-" + base
	if !listenerNameTaken(candidate, used) {
		return candidate
	}
	sum := sha256.Sum256([]byte(domain))
	suffix := hex.EncodeToString(sum[:3])
	return "https-custom-" + base + "-" + suffix
}

func listenerNameTaken(name string, used map[string]struct{}) bool {
	if used == nil {
		return false
	}
	_, ok := used[name]
	return ok
}

func sanitizeListenerSectionName(viewer string) string {
	lower := strings.ToLower(viewer)
	var b strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "viewer"
	}
	if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(out) {
		return "viewer"
	}
	return out
}

func listenerCertRef(listener map[string]interface{}) string {
	refs, _, _ := unstructured.NestedSlice(listener, "tls", "certificateRefs")
	if len(refs) == 0 {
		return ""
	}
	m, ok := refs[0].(map[string]interface{})
	if !ok {
		return ""
	}
	name, _, _ := unstructured.NestedString(m, "name")
	return name
}

func listenerSlicesEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		am, aok := a[i].(map[string]interface{})
		bm, bok := b[i].(map[string]interface{})
		if !aok || !bok {
			return false
		}
		if am["name"] != bm["name"] {
			return false
		}
		ah, _, _ := unstructured.NestedString(am, "hostname")
		bh, _, _ := unstructured.NestedString(bm, "hostname")
		if ah != bh {
			return false
		}
		if listenerCertRef(am) != listenerCertRef(bm) {
			return false
		}
	}
	return true
}

// SetupWithManager registers watches for entrance TLS Secrets and the Gateway.
func (r *EntranceTLSListenerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r == nil {
		return nil
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	onlyEntranceTLS := predicate.NewPredicateFuncs(func(o client.Object) bool {
		if o.GetNamespace() != defaultGatewayNS {
			return false
		}
		return strings.HasPrefix(o.GetName(), sharedEntranceTLSPrefix) || strings.HasPrefix(o.GetName(), customDomainTLSPrefix)
	})
	onlyAppGateway := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetNamespace() == defaultGatewayNS && o.GetName() == defaultGatewayName
	})
	gatewayWatch := &unstructured.Unstructured{}
	gatewayWatch.SetGroupVersionKind(gatewayGVK)
	gatewayNN := types.NamespacedName{Namespace: defaultGatewayNS, Name: defaultGatewayName}
	return ctrl.NewControllerManagedBy(mgr).
		Named("entrance-tls-listener").
		For(&corev1.Secret{}, builder.WithPredicates(onlyEntranceTLS)).
		Watches(
			gatewayWatch,
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, _ client.Object) []reconcile.Request {
				return []reconcile.Request{{NamespacedName: gatewayNN}}
			}),
			builder.WithPredicates(onlyAppGateway),
		).
		Complete(r)
}
