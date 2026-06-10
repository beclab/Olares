package routecontrol

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/beclab/Olares/framework/app-service/pkg/cluster"
	"github.com/beclab/Olares/framework/app-service/pkg/gateway"
	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

var callerSRRListRetryExhaustedTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "caller_srr_list_retry_exhausted_total",
	Help: "CallerReconciler SRR list retries exhausted; fail-safe callee branch taken.",
})

func init() {
	metrics.Registry.MustRegister(callerSRRListRetryExhaustedTotal)
}

// CallerReconciler materialises caller-side Linkerd namespace injection for
// workloads that opt into cluster-internal Shared access, plus the singleton
// gateway-side ingress NetworkPolicy that admits caller traffic.
//
// requirement (NP-minimal scheme):
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
	optedIn, err := gateway.NamespaceOptedIntoGateway(ctx, r.Client, ns)
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
	if callerMeshEnabled(ctx) {
		if err := ensureCallerNamespaceLinkerdInject(ctx, r.Client, ns, true); err != nil {
			return err
		}
		if err := ensureCallerNamespaceLinkerdSkipPorts(ctx, r.Client, ns); err != nil {
			return err
		}
		if err := ensureCallerNamespaceInClusterLabel(ctx, r.Client, ns, true); err != nil {
			return err
		}
	}
	if err := r.applyNetworkPolicy(ctx, security.NewAppGatewayInClusterCallerIngressNP(r.gatewayNS())); err != nil {
		return err
	}
	// Self-heal routecontrol-managed shared ingress NPs in upstream namespaces.
	// This backfills only app-gateway-shared-ingress-np objects and does not
	// touch template-managed namespace policies.
	return BackfillSharedIngressNetworkPolicies(ctx, r.Client, GatewayRef{
		GatewayNamespace: r.gatewayNS(),
	})
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
	if callerMeshEnabled(ctx) {
		if err := ensureCallerNamespaceLinkerdInject(ctx, r.Client, ns, false); err != nil {
			return err
		}
		if err := removeCallerNamespaceLinkerdSkipPorts(ctx, r.Client, ns); err != nil {
			return err
		}
		return ensureCallerNamespaceInClusterLabel(ctx, r.Client, ns, false)
	}
	return nil
}

func callerMeshEnabled(ctx context.Context) bool {
	snap, err := cluster.GetSnapshot(ctx)
	if err != nil {
		return true
	}
	return snap.MeshLinkerdEnabled()
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

// ensureCallerNamespaceInClusterLabel toggles gateway.olares.io/in-cluster-caller on
// the namespace so app-gateway-mesh-np admits meshed caller proxies to linkerd-identity.
func ensureCallerNamespaceInClusterLabel(ctx context.Context, c client.Client, namespace string, enable bool) error {
	if namespace == "" || !isMeshMandatoryCallerNamespace(namespace) {
		return nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	desired := "true"
	if !enable {
		delete(ns.Labels, security.NamespaceInClusterCallerLabel)
		return c.Update(ctx, &ns)
	}
	if ns.Labels[security.NamespaceInClusterCallerLabel] == desired {
		return nil
	}
	ns.Labels[security.NamespaceInClusterCallerLabel] = desired
	return c.Update(ctx, &ns)
}

func ensureCallerNamespaceLinkerdSkipPorts(ctx context.Context, c client.Client, namespace string) error {
	if namespace == "" || !isMeshMandatoryCallerNamespace(namespace) {
		return nil
	}

	// Keep caller strong paths mesh-hijacked:
	// - HTTPS strong identity on :8081
	// - HTTP strong path on :8082
	skipOutbound, err := ComputeSkipOutboundPorts(
		MeshHijackServicePorts(DefaultInClusterStrongIdentityServicePort),
	)
	if err != nil {
		return fmt.Errorf("compute skip-outbound for namespace %q: %w", namespace, err)
	}

	isCallee, err := isCalleeNamespace(ctx, c, namespace)
	if err != nil {
		return err
	}
	skipInbound := PureCallerInboundSkipPorts
	if isCallee {
		skipInbound = OlaresEnvoyInboundSkipPorts
	}

	return patchNSAnnotations(ctx, c, namespace, map[string]string{
		LinkerdSkipInboundPortsAnnotation:  skipInbound,
		LinkerdSkipOutboundPortsAnnotation: skipOutbound,
	})
}

func removeCallerNamespaceLinkerdSkipPorts(ctx context.Context, c client.Client, namespace string) error {
	if namespace == "" || !isMeshMandatoryCallerNamespace(namespace) {
		return nil
	}

	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if len(ns.Annotations) == 0 {
		return nil
	}

	_, hasInbound := ns.Annotations[LinkerdSkipInboundPortsAnnotation]
	_, hasOutbound := ns.Annotations[LinkerdSkipOutboundPortsAnnotation]
	if !hasInbound && !hasOutbound {
		return nil
	}

	delete(ns.Annotations, LinkerdSkipInboundPortsAnnotation)
	delete(ns.Annotations, LinkerdSkipOutboundPortsAnnotation)
	return c.Update(ctx, &ns)
}

func isCalleeNamespace(ctx context.Context, c client.Client, namespace string) (bool, error) {
	var (
		lastErr error
		list    srrv1alpha1.SharedRouteRegistryList
	)
	backoff := wait.Backoff{
		Steps:    3,
		Duration: 50 * time.Millisecond,
		Factor:   2.0,
	}
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if listErr := c.List(ctx, &list, client.InNamespace(namespace)); listErr != nil {
			lastErr = listErr
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Warningf("CALLER_SRR_LIST_FAILED namespace=%q retry_exhausted err=%v", namespace, lastErr)
		callerSRRListRetryExhaustedTotal.Inc()
		return true, nil
	}

	for i := range list.Items {
		if srrHasReadyTrueCondition(&list.Items[i]) {
			return true, nil
		}
	}
	return false, nil
}

func srrHasReadyTrueCondition(srr *srrv1alpha1.SharedRouteRegistry) bool {
	if srr == nil {
		return false
	}
	for i := range srr.Status.Conditions {
		cond := &srr.Status.Conditions[i]
		if cond.Type == ConditionReady && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}

func patchNSAnnotations(ctx context.Context, c client.Client, namespace string, desired map[string]string) error {
	if namespace == "" || len(desired) == 0 {
		return nil
	}

	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}

	changed := false
	for k, want := range desired {
		if got := ns.Annotations[k]; got != want {
			ns.Annotations[k] = want
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return c.Update(ctx, &ns)
}
