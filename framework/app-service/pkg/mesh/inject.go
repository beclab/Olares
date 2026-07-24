package mesh

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/security"
)

const (
	// LinkerdInjectAnnotation is the upstream Linkerd inject annotation key.
	LinkerdInjectAnnotation = "linkerd.io/inject"
	// LinkerdInjectEnabled / LinkerdInjectDisabled are canonical values.
	LinkerdInjectEnabled  = "enabled"
	LinkerdInjectDisabled = "disabled"
)

// ShouldInjectLinkerdProxy reports whether linkerd-proxy should be injected.
// ARCH S6: same result as mesh-in (injectMeshIn).
func ShouldInjectLinkerdProxy(injectMeshIn bool) bool {
	return injectMeshIn
}

// AnnotatePodForLinkerdInject sets linkerd.io/inject=enabled on the pod.
func AnnotatePodForLinkerdInject(pod *corev1.Pod, enable bool) {
	if pod == nil || !enable {
		return
	}
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations[LinkerdInjectAnnotation] = LinkerdInjectEnabled
}

// EnsureCallerNamespaceMeshAccess labels the caller namespace so static
// app-gateway-mesh-np (os-mesh) admits its proxies to the control plane.
// When enable is false, the in-cluster-caller label is removed.
func EnsureCallerNamespaceMeshAccess(ctx context.Context, c client.Client, namespace string, enable bool) error {
	if c == nil || namespace == "" {
		return nil
	}
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: namespace}, &ns); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(2).Infof("mesh-xport: skip caller ns %q: not found", namespace)
			return nil
		}
		klog.Errorf("mesh-xport: get caller ns %s failed: %v", namespace, err)
		return err
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	cur := ns.Labels[security.NamespaceInClusterCallerLabel]
	if enable {
		if cur == "true" {
			return nil
		}
		ns.Labels[security.NamespaceInClusterCallerLabel] = "true"
	} else {
		if cur == "" {
			return nil
		}
		delete(ns.Labels, security.NamespaceInClusterCallerLabel)
	}
	if err := c.Update(ctx, &ns); err != nil {
		klog.Errorf("mesh-xport: update caller ns %s label failed: %v", namespace, err)
		return err
	}
	klog.Infof("mesh-xport: caller ns=%s in-cluster-caller=%v", namespace, enable)
	return nil
}
