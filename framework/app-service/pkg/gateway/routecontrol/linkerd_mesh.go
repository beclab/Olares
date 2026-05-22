package routecontrol

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	srrv1alpha1 "github.com/beclab/Olares/framework/app-service/pkg/gateway/v1alpha1"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

const (
	// AnnotationLinkerdInject is the user-facing opt-out annotation for shared workload
	// namespaces. Set on the Namespace as "disabled" to prevent app-service from
	// flipping linkerd.io/inject when entering gateway mode. The controller never
	// writes this key — it is reserved for operators.
	AnnotationLinkerdInject = "gateway.olares.io/linkerd-inject"

	// LinkerdInjectAnnotation is the upstream Linkerd namespace-level inject annotation.
	LinkerdInjectAnnotation = "linkerd.io/inject"

	// LinkerdInjectEnabled / LinkerdInjectDisabled are the canonical values for
	// LinkerdInjectAnnotation.
	LinkerdInjectEnabled  = "enabled"
	LinkerdInjectDisabled = "disabled"
)

// applySharedLinkerdMeshNetworkPolicy creates / updates the NetworkPolicy that lets
// linkerd control plane (and optionally linkerd-viz) reach the meshed proxies in a
// shared workload namespace.
//
// The helper soft-skips (returns nil + V(2) log) when prerequisites for mesh
// injection aren't met: target namespace not labelled bytetrade.io/ns-shared=true,
// or backend Service has no Pod selector (headless / ExternalName). Those cases
// occur for v3 same-namespace shared apps which run inside the Application
// namespace and don't take part in the dedicated shared mesh.
func applySharedLinkerdMeshNetworkPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) error {
	if svc == nil || len(svc.Spec.Selector) == 0 {
		klog.V(2).Infof("skip shared linkerd mesh NP for srr=%s/%s: service selector empty",
			srr.Namespace, srr.Name)
		return nil
	}
	npNS := networkPolicyNamespace(srr, svc)
	shared, err := isSharedNamespaceByName(ctx, c, npNS)
	if err != nil {
		return err
	}
	if !shared {
		klog.V(2).Infof("skip shared linkerd mesh NP for srr=%s/%s: ns %q is not a shared workload namespace",
			srr.Namespace, srr.Name, npNS)
		return nil
	}
	desired := security.NewSharedLinkerdControlPlaneIngressNetworkPolicy(npNS, svc.Spec.Selector)
	setOwnerSRR(desired, srr)
	if desired.Labels == nil {
		desired.Labels = map[string]string{}
	}
	desired.Labels[ManagedByLabel] = ManagedByValue
	desired.Labels[InstanceLabel] = srr.Name

	current := &networkingv1.NetworkPolicy{}
	err = c.Get(ctx, types.NamespacedName{Namespace: npNS, Name: security.SharedLinkerdMeshIngressNPName}, current)
	switch {
	case apierrors.IsNotFound(err):
		return c.Create(ctx, desired)
	case err != nil:
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	current.Labels[InstanceLabel] = srr.Name
	setOwnerSRR(current, srr)
	return c.Update(ctx, current)
}

func deleteSharedLinkerdMeshNetworkPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	for _, ns := range networkPolicyNamespacesToClean(srr) {
		obj := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      security.SharedLinkerdMeshIngressNPName,
				Namespace: ns,
			},
		}
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// ensureSharedNamespaceLinkerdInject toggles linkerd.io/inject on a shared workload
// namespace when its SRR enters or leaves gateway mode.
//
// Rules:
//   - Only namespaces labelled security.NamespaceSharedLabel="true" are mutated; any
//     other namespace (including missing namespaces) is silently skipped with a
//     V(2) log so v3 same-namespace shared apps and partially-populated test
//     fixtures do not break the reconcile loop.
//   - The user opt-out annotation (AnnotationLinkerdInject="disabled") wins on both
//     enable and disable paths and is never overwritten by the controller.
//   - If the upstream linkerd.io/inject annotation already matches the desired value,
//     the namespace is left untouched (no Update, no ResourceVersion churn).
func ensureSharedNamespaceLinkerdInject(ctx context.Context, c client.Client, namespace string, enable bool) error {
	if namespace == "" {
		return nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("skip linkerd inject on ns %q: not found", namespace)
			return nil
		}
		return err
	}
	if !isSharedWorkloadNamespace(&ns) {
		klog.V(2).Infof("skip linkerd inject on ns %q: missing label %s=true",
			namespace, security.NamespaceSharedLabel)
		return nil
	}
	if ns.Annotations[AnnotationLinkerdInject] == LinkerdInjectDisabled {
		// Operator opt-out: do not flip linkerd.io/inject in either direction.
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

func isSharedWorkloadNamespace(ns *corev1.Namespace) bool {
	if ns == nil {
		return false
	}
	return ns.Labels[security.NamespaceSharedLabel] == "true"
}

func isSharedNamespaceByName(ctx context.Context, c client.Client, name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return isSharedWorkloadNamespace(&ns), nil
}
