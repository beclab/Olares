package mesh

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

const (
	// LinkerdInjectAnnotation is the upstream Linkerd inject annotation key.
	LinkerdInjectAnnotation = "linkerd.io/inject"
	// LinkerdInjectEnabled / LinkerdInjectDisabled are canonical values.
	LinkerdInjectEnabled  = "enabled"
	LinkerdInjectDisabled = "disabled"
	// CallerLinkerdInjectManagedAnnotation marks template inject owned by app-service.
	CallerLinkerdInjectManagedAnnotation = "gateway.olares.io/caller-linkerd-inject"
	// CallerLinkerdInjectManagedValue is the managed marker value.
	CallerLinkerdInjectManagedValue = "managed"
	annotDecide                     = "gateway.olares.io/shared-caller-decide"
)

// ShouldInjectLinkerdProxy reports whether linkerd-proxy should be injected.
// Same result as mesh-in (injectMeshIn).
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

// NamespaceHasDecideTrueCaller reports whether any Application with Spec.Namespace
// equal to namespace declares Shared caller (via declares predicate or decide annotation).
func NamespaceHasDecideTrueCaller(ctx context.Context, c client.Client, namespace string, declares func(map[string]string) bool) (bool, error) {
	if c == nil || namespace == "" || declares == nil {
		return false, nil
	}
	var list appv1alpha1.ApplicationList
	if err := c.List(ctx, &list); err != nil {
		klog.Errorf("mesh-xport: list applications for ns OR %s failed: %v", namespace, err)
		return false, err
	}
	for i := range list.Items {
		app := &list.Items[i]
		if app.Spec.Namespace != namespace {
			continue
		}
		if declares(app.Spec.Settings) {
			return true, nil
		}
		if app.Annotations != nil {
			v := app.Annotations[annotDecide]
			if v == "true" || v == "True" {
				return true, nil
			}
		}
	}
	return false, nil
}

// EnsureCallerNamespaceMeshAccess labels the caller namespace so static
// app-gateway-mesh-np (os-mesh) admits its proxies to the control plane.
// It does not set namespace-scoped linkerd.io/inject; workload templates own inject.
// When enable is true, any stale NS inject annotation is deleted.
// When enable is false, the in-cluster-caller label and inject annotation are cleared.
// Shared workload namespaces are skipped: their inject annotation is owned by
// SharedRouteRegistry reconcile (ensureSharedNamespaceLinkerdInject).
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
	if ns.Labels[security.NamespaceSharedLabel] == "true" {
		klog.V(2).Infof("mesh-xport: skip caller mesh access on shared ns %q", namespace)
		return nil
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	curLabel := ns.Labels[security.NamespaceInClusterCallerLabel]
	curInject := ns.Annotations[LinkerdInjectAnnotation]
	changed := false
	if enable {
		if curLabel != "true" {
			ns.Labels[security.NamespaceInClusterCallerLabel] = "true"
			changed = true
		}
		if curInject != "" {
			delete(ns.Annotations, LinkerdInjectAnnotation)
			changed = true
		}
	} else {
		if curLabel != "" {
			delete(ns.Labels, security.NamespaceInClusterCallerLabel)
			changed = true
		}
		if curInject != "" {
			delete(ns.Annotations, LinkerdInjectAnnotation)
			changed = true
		}
	}
	if !changed {
		if enable {
			return EnsureAppGatewayMeshNetworkPolicies(ctx, c)
		}
		return nil
	}
	if err := c.Update(ctx, &ns); err != nil {
		klog.Errorf("mesh-xport: update caller ns %s label/inject failed: %v", namespace, err)
		return err
	}
	klog.Infof("mesh-xport: caller ns=%s in-cluster-caller=%v (ns inject not written)",
		namespace, enable)
	if enable {
		if err := EnsureAppGatewayMeshNetworkPolicies(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

// EnsureCallerWorkloadLinkerdInject sets or clears linkerd inject on Application
// owned Deployment/StatefulSet pod templates. Only clears when the managed marker
// is present.
func EnsureCallerWorkloadLinkerdInject(ctx context.Context, c client.Client, appNamespace, appName string, enable bool) error {
	if c == nil || appNamespace == "" || appName == "" {
		return nil
	}
	sel := client.MatchingLabels{constants.ApplicationNameLabel: appName}
	var deps appsv1.DeploymentList
	if err := c.List(ctx, &deps, client.InNamespace(appNamespace), sel); err != nil {
		klog.Errorf("mesh-xport: list deploys ns=%s app=%s failed: %v", appNamespace, appName, err)
		return err
	}
	for i := range deps.Items {
		dep := &deps.Items[i]
		if !mutatePodTemplateAnnotations(&dep.Spec.Template.ObjectMeta, enable) {
			continue
		}
		if err := c.Update(ctx, dep); err != nil {
			klog.Errorf("mesh-xport: update deploy %s/%s inject failed: %v", dep.Namespace, dep.Name, err)
			return err
		}
		klog.Infof("mesh-xport: deploy %s/%s template linkerd inject enable=%v", dep.Namespace, dep.Name, enable)
	}
	var stss appsv1.StatefulSetList
	if err := c.List(ctx, &stss, client.InNamespace(appNamespace), sel); err != nil {
		klog.Errorf("mesh-xport: list sts ns=%s app=%s failed: %v", appNamespace, appName, err)
		return err
	}
	for i := range stss.Items {
		sts := &stss.Items[i]
		if !mutatePodTemplateAnnotations(&sts.Spec.Template.ObjectMeta, enable) {
			continue
		}
		if err := c.Update(ctx, sts); err != nil {
			klog.Errorf("mesh-xport: update sts %s/%s inject failed: %v", sts.Namespace, sts.Name, err)
			return err
		}
		klog.Infof("mesh-xport: sts %s/%s template linkerd inject enable=%v", sts.Namespace, sts.Name, enable)
	}
	return nil
}

func mutatePodTemplateAnnotations(meta *metav1.ObjectMeta, enable bool) bool {
	if meta == nil {
		return false
	}
	if meta.Annotations == nil {
		meta.Annotations = map[string]string{}
	}
	ann := meta.Annotations
	if enable {
		changed := false
		if ann[LinkerdInjectAnnotation] != LinkerdInjectEnabled {
			ann[LinkerdInjectAnnotation] = LinkerdInjectEnabled
			changed = true
		}
		if ann[CallerLinkerdInjectManagedAnnotation] != CallerLinkerdInjectManagedValue {
			ann[CallerLinkerdInjectManagedAnnotation] = CallerLinkerdInjectManagedValue
			changed = true
		}
		return changed
	}
	if ann[CallerLinkerdInjectManagedAnnotation] != CallerLinkerdInjectManagedValue {
		return false
	}
	delete(ann, LinkerdInjectAnnotation)
	delete(ann, CallerLinkerdInjectManagedAnnotation)
	return true
}
