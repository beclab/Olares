package routecontrol

import (
	"context"
	"strings"

	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

// CallerReconciler materialises caller-side Linkerd namespace injection for
// workloads that opt into cluster-internal Shared access, plus the singleton
// gateway-side ingress NetworkPolicy that admits caller traffic.
//
// requirement (archdoc Shared集群内访问v1 §3, NP-minimal v1.0):
//   - Caller namespaces MUST NOT receive managed egress NetworkPolicies; user
//     application NS template (NPAppSpace) is Ingress-only, so the default
//     egress allow-all is preserved.
//   - Gateway namespace receives ONE ingress NP (#1) admitting all opted-in
//     caller namespaces on any port. This is reconciled idempotently here.
//   - Legacy managed caller egress NPs (caller-dns / caller-middleware /
//     caller-mesh / caller-to-app-gateway) are retained in the cleanup name
//     list so existing clusters GC them on first reconcile after upgrade.
type CallerReconciler struct {
	Client    client.Client
	GatewayNS string
}

func (r *CallerReconciler) gatewayNS() string {
	if r != nil && r.GatewayNS != "" {
		return r.GatewayNS
	}
	return defaultGatewayNS
}

// Reconcile applies or removes caller-side mesh injection and ensures the
// gateway-side singleton ingress NP. NP-minimal v1.0: no managed egress is
// written into caller namespaces (default Allow on app-np Ingress-only).
func (r *CallerReconciler) Reconcile(ctx context.Context, ns string) error {
	if r == nil || r.Client == nil || ns == "" {
		return nil
	}
	if !isMeshMandatoryCallerNamespace(ns) {
		return nil
	}
	var nsObj corev1.Namespace
	if err := r.Client.Get(ctx, types.NamespacedName{Name: ns}, &nsObj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if nsObj.Annotations[AnnotationLinkerdInject] == LinkerdInjectDisabled {
		return r.cleanupCallerResources(ctx, ns)
	}
	optedIn, err := r.namespaceOptedIntoGateway(ctx, ns)
	if err != nil {
		return err
	}
	if !optedIn {
		return r.cleanupCallerResources(ctx, ns)
	}
	// Upgrade-time GC: clusters that ran pre-v1.0 still carry the 4 legacy
	// caller egress NPs in opted-in namespaces; cleanupCallerResources only
	// fires on opt-out, so we must GC them on every opt-in reconcile until
	// the cleanup names are removed from the codebase entirely.
	if err := r.gcLegacyCallerEgress(ctx, ns); err != nil {
		return err
	}
	if err := ensureCallerNamespaceLinkerdInject(ctx, r.Client, ns, true); err != nil {
		return err
	}
	return r.applyNetworkPolicy(ctx, security.NewAppGatewayInClusterCallerIngressNP(r.gatewayNS()))
}

// gcLegacyCallerEgress deletes the 4 pre-v1.0 managed caller egress NPs from
// an opted-in caller namespace. Idempotent (NotFound is success).
func (r *CallerReconciler) gcLegacyCallerEgress(ctx context.Context, ns string) error {
	for _, name := range []string{
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
	} {
		obj := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		}
		if err := r.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (r *CallerReconciler) namespaceOptedIntoGateway(ctx context.Context, ns string) (bool, error) {
	var list appv1alpha1.ApplicationList
	if err := r.Client.List(ctx, &list); err != nil {
		return false, err
	}
	for i := range list.Items {
		app := &list.Items[i]
		if app.Spec.Namespace != ns {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(app.Annotations[gateway.AnnotationInCluster]), gateway.InClusterGateway) {
			continue
		}
		if strings.TrimSpace(app.Spec.Settings["clusterAppRef"]) != "" {
			return true, nil
		}
	}
	return false, nil
}

func (r *CallerReconciler) applyNetworkPolicy(ctx context.Context, desired *networkingv1.NetworkPolicy) error {
	if desired == nil {
		return nil
	}
	current := &networkingv1.NetworkPolicy{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: desired.Namespace,
		Name:      desired.Name,
	}, current)
	switch {
	case apierrors.IsNotFound(err):
		return r.Client.Create(ctx, desired)
	case err != nil:
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	for k, v := range desired.Labels {
		current.Labels[k] = v
	}
	return r.Client.Update(ctx, current)
}

// cleanupCallerResources GCs legacy managed caller egress NPs (NP-minimal v1.0
// drops them) and disables linkerd inject on opt-out. The cleanup name list is
// retained so existing clusters upgrade cleanly without manual kubectl delete.
func (r *CallerReconciler) cleanupCallerResources(ctx context.Context, ns string) error {
	for _, name := range []string{
		security.CallerDNSEgressNPName,
		security.CallerMiddlewareEgressNPName,
		security.CallerMeshEgressNPName,
		security.CallerToAppGatewayEgressNPName,
	} {
		obj := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		}
		if err := r.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return ensureCallerNamespaceLinkerdInject(ctx, r.Client, ns, false)
}

// isMeshMandatoryCallerNamespace reports whether ns is a user/third-party workload
// namespace eligible for caller mesh (excludes platform / L4 / mesh system NS).
func isMeshMandatoryCallerNamespace(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}
	excluded := []string{
		"kube-",
		"linkerd",
		"app-gateway",
		"os-framework",
		"os-platform",
		"os-network",
	}
	for _, p := range excluded {
		if name == p || strings.HasPrefix(name, p) {
			return false
		}
	}
	if strings.HasPrefix(name, "user-space-") || strings.HasPrefix(name, "user-system-") {
		return true
	}
	return strings.Contains(name, "-")
}

// ensureCallerNamespaceLinkerdInject toggles linkerd.io/inject on caller namespaces only.
func ensureCallerNamespaceLinkerdInject(ctx context.Context, c client.Client, namespace string, enable bool) error {
	if namespace == "" || !isMeshMandatoryCallerNamespace(namespace) {
		return nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("skip caller linkerd inject on ns %q: not found", namespace)
			return nil
		}
		return err
	}
	if ns.Annotations[AnnotationLinkerdInject] == LinkerdInjectDisabled {
		return nil
	}
	desired := LinkerdInjectEnabled
	if !enable {
		desired = LinkerdInjectDisabled
	}
	if ns.Annotations[LinkerdInjectAnnotation] == desired {
		return nil
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Annotations[LinkerdInjectAnnotation] = desired
	return c.Update(ctx, &ns)
}
