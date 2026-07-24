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
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

// applySharedLinkerdMeshNetworkPolicy creates/updates the NP that lets the
// Linkerd control plane reach meshed proxies in a shared workload namespace.
// Soft-skips when the target NS is not labelled shared or the Service has no selector.
func applySharedLinkerdMeshNetworkPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry, svc *corev1.Service) error {
	if svc == nil || len(svc.Spec.Selector) == 0 {
		klog.V(2).Infof("mesh-xport: skip shared mesh NP for srr=%s/%s: empty service selector",
			srr.Namespace, srr.Name)
		return nil
	}
	npNS := networkPolicyNamespace(srr, svc)
	shared, err := isSharedNamespaceByName(ctx, c, npNS)
	if err != nil {
		klog.Errorf("mesh-xport: check shared ns %s failed: %v", npNS, err)
		return err
	}
	if !shared {
		klog.V(2).Infof("mesh-xport: skip shared mesh NP for srr=%s/%s: ns %q not shared",
			srr.Namespace, srr.Name, npNS)
		return nil
	}
	// Empty PodSelector: one NP per shared NS admits os-mesh to all pods.
	// Do not attach SRR OwnerReferences (Controller=true): multiple gateway-mode
	// SRRs share this singleton; dual controllers are rejected by the API.
	desired := security.NewSharedLinkerdControlPlaneIngressNetworkPolicy(npNS, nil)
	if desired.Labels == nil {
		desired.Labels = map[string]string{}
	}
	desired.Labels[ManagedByLabel] = ManagedByValue
	desired.Labels[RouteControlComponentLabel] = RouteControlComponentValue
	delete(desired.Labels, InstanceLabel)

	current := &networkingv1.NetworkPolicy{}
	err = c.Get(ctx, types.NamespacedName{Namespace: npNS, Name: security.SharedLinkerdMeshIngressNPName}, current)
	switch {
	case apierrors.IsNotFound(err):
		if err := c.Create(ctx, desired); err != nil {
			klog.Errorf("mesh-xport: create mesh NP %s/%s failed: %v", npNS, desired.Name, err)
			return err
		}
		klog.Infof("mesh-xport: created mesh NP %s/%s", npNS, desired.Name)
		return nil
	case err != nil:
		klog.Errorf("mesh-xport: get mesh NP %s/%s failed: %v", npNS, security.SharedLinkerdMeshIngressNPName, err)
		return err
	}
	current.Spec = desired.Spec
	if current.Labels == nil {
		current.Labels = map[string]string{}
	}
	current.Labels[ManagedByLabel] = ManagedByValue
	current.Labels[RouteControlComponentLabel] = RouteControlComponentValue
	delete(current.Labels, InstanceLabel)
	current.SetOwnerReferences(nil)
	if err := c.Update(ctx, current); err != nil {
		klog.Errorf("mesh-xport: update mesh NP %s/%s failed: %v", npNS, current.Name, err)
		return err
	}
	return nil
}

func deleteSharedLinkerdMeshNetworkPolicy(ctx context.Context, c client.Client, srr *srrv1alpha1.SharedRouteRegistry) error {
	for _, ns := range networkPolicyNamespacesToClean(srr) {
		stillNeeded, err := hasOtherGatewayModeSRRForUpstream(ctx, c, srr, ns)
		if err != nil {
			return err
		}
		if stillNeeded {
			continue
		}
		obj := &networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      security.SharedLinkerdMeshIngressNPName,
				Namespace: ns,
			},
		}
		if err := c.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			klog.Errorf("mesh-xport: delete mesh NP %s/%s failed: %v", ns, obj.Name, err)
			return err
		}
	}
	return nil
}

// ensureSharedNamespaceLinkerdInject toggles linkerd.io/inject on a shared
// workload namespace when its SRR enters or leaves gateway mode.
func ensureSharedNamespaceLinkerdInject(ctx context.Context, c client.Client, namespace string, enable bool) error {
	if namespace == "" {
		return nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("mesh-xport: skip inject on ns %q: not found", namespace)
			return nil
		}
		klog.Errorf("mesh-xport: get ns %s failed: %v", namespace, err)
		return err
	}
	if !isSharedWorkloadNamespace(&ns) {
		klog.V(2).Infof("mesh-xport: skip inject on ns %q: missing %s=true",
			namespace, security.NamespaceSharedLabel)
		return nil
	}
	desired := mesh.LinkerdInjectEnabled
	if !enable {
		desired = mesh.LinkerdInjectDisabled
	}
	if ns.Annotations[mesh.LinkerdInjectAnnotation] == desired {
		return nil
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Annotations[mesh.LinkerdInjectAnnotation] = desired
	if err := c.Update(ctx, &ns); err != nil {
		klog.Errorf("mesh-xport: update inject on ns %s failed: %v", namespace, err)
		return err
	}
	klog.Infof("mesh-xport: ns=%s linkerd.io/inject=%s", namespace, desired)
	return nil
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
