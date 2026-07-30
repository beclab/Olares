package meshinagent

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// linkerdProxyContainerName is the upstream Linkerd sidecar container name.
const linkerdProxyContainerName = "linkerd-proxy"

// ListAppWorkloads returns Deployments/StatefulSets owned by the Application
// (label applications.app.bytetrade.io/name=<appName> in app namespace).
func ListAppWorkloads(ctx context.Context, c client.Client, appNamespace, appName string) ([]WorkloadRef, error) {
	if c == nil || appNamespace == "" || appName == "" {
		return nil, nil
	}
	sel := labels.SelectorFromSet(labels.Set{constants.ApplicationNameLabel: appName})
	var out []WorkloadRef
	var deps appsv1.DeploymentList
	if err := c.List(ctx, &deps, client.InNamespace(appNamespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		klog.Errorf("mesh-in-rollout: list deploys ns=%s app=%s failed: %v", appNamespace, appName, err)
		return nil, err
	}
	for i := range deps.Items {
		out = append(out, WorkloadRef{Kind: "Deployment", Namespace: deps.Items[i].Namespace, Name: deps.Items[i].Name})
	}
	var stss appsv1.StatefulSetList
	if err := c.List(ctx, &stss, client.InNamespace(appNamespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		klog.Errorf("mesh-in-rollout: list sts ns=%s app=%s failed: %v", appNamespace, appName, err)
		return nil, err
	}
	for i := range stss.Items {
		out = append(out, WorkloadRef{Kind: "StatefulSet", Namespace: stss.Items[i].Namespace, Name: stss.Items[i].Name})
	}
	return out, nil
}

// ListSharedInjectWorkloads lists workloads in Shared namespaces with linkerd.io/inject=enabled.
func ListSharedInjectWorkloads(ctx context.Context, c client.Client) ([]WorkloadRef, error) {
	if c == nil {
		return nil, nil
	}
	var nss corev1.NamespaceList
	if err := c.List(ctx, &nss, client.MatchingLabels{security.NamespaceSharedLabel: "true"}); err != nil {
		klog.Errorf("mesh-in-rollout: list shared namespaces failed: %v", err)
		return nil, err
	}
	var out []WorkloadRef
	for i := range nss.Items {
		ns := &nss.Items[i]
		if ns.Annotations[mesh.LinkerdInjectAnnotation] != mesh.LinkerdInjectEnabled {
			continue
		}
		refs, err := listAllWorkloadsInNamespace(ctx, c, ns.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, refs...)
	}
	return out, nil
}

// ListEGDataplaneWorkloads lists os-gateway Deployments whose pod template asks for Linkerd inject.
func ListEGDataplaneWorkloads(ctx context.Context, c client.Client) ([]WorkloadRef, error) {
	if c == nil {
		return nil, nil
	}
	var deps appsv1.DeploymentList
	if err := c.List(ctx, &deps, client.InNamespace(EGDataplaneNamespace)); err != nil {
		klog.Errorf("mesh-in-rollout: list eg deploys failed: %v", err)
		return nil, err
	}
	var out []WorkloadRef
	for i := range deps.Items {
		dep := &deps.Items[i]
		ann := dep.Spec.Template.Annotations
		if ann == nil || ann[mesh.LinkerdInjectAnnotation] != mesh.LinkerdInjectEnabled {
			continue
		}
		out = append(out, WorkloadRef{Kind: "Deployment", Namespace: dep.Namespace, Name: dep.Name})
	}
	return out, nil
}

// ListDecideTrueApplications returns Applications with decide=true (settings or annotation).
func ListDecideTrueApplications(ctx context.Context, c client.Client) ([]appv1alpha1.Application, error) {
	if c == nil {
		return nil, nil
	}
	var list appv1alpha1.ApplicationList
	if err := c.List(ctx, &list); err != nil {
		klog.Errorf("mesh-in-rollout: list applications failed: %v", err)
		return nil, err
	}
	var out []appv1alpha1.Application
	for i := range list.Items {
		app := list.Items[i]
		if DeclaresSharedCaller(app.Spec.Settings) {
			out = append(out, app)
			continue
		}
		if app.Annotations != nil &&
			(app.Annotations[AnnotDecide] == "true" || app.Annotations[AnnotDecide] == "True") {
			out = append(out, app)
		}
	}
	return out, nil
}

// InjectSpec is the sidecar set a workload is expected to run after the rollout.
// Shared callees and the EG dataplane take linkerd-proxy only; Shared callers add
// the mesh-in agent.
type InjectSpec struct {
	MeshIn  bool
	Linkerd bool
}

// SharedCallerInject is the Shared caller expectation (both sidecars).
func SharedCallerInject(inject bool) InjectSpec {
	if !inject {
		return InjectSpec{}
	}
	return InjectSpec{MeshIn: true, Linkerd: true}
}

// LinkerdOnlyInject is the Shared callee / EG dataplane expectation.
func LinkerdOnlyInject() InjectSpec {
	return InjectSpec{Linkerd: true}
}

// FilterWorkloadsNeedingInject drops workloads whose running pods already match
// want, so an unchanged sidecar set never costs users a restart. Lookup failures
// drop the workload as well: a later reconcile retries, which is preferable to
// restarting a healthy workload on a transient API error.
func FilterWorkloadsNeedingInject(ctx context.Context, c client.Client, refs []WorkloadRef, want InjectSpec) []WorkloadRef {
	if c == nil {
		return nil
	}
	var out []WorkloadRef
	for _, ref := range refs {
		need, err := workloadNeedsInjectChange(ctx, c, ref, want)
		if err != nil {
			klog.Warningf("mesh-in-rollout: skip %s: inject state unknown: %v", ref.Key(), err)
			continue
		}
		if !need {
			klog.V(4).Infof("mesh-in-rollout: skip %s: sidecars already match", ref.Key())
			continue
		}
		out = append(out, ref)
	}
	return out
}

func workloadNeedsInjectChange(ctx context.Context, c client.Client, ref WorkloadRef, want InjectSpec) (bool, error) {
	selector, err := workloadPodSelector(ctx, c, ref)
	if err != nil {
		return false, err
	}
	if selector == nil {
		return false, nil
	}
	var pods corev1.PodList
	if err := c.List(ctx, &pods, client.InNamespace(ref.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return false, err
	}
	if len(pods.Items) == 0 {
		return false, nil
	}
	for i := range pods.Items {
		if pods.Items[i].DeletionTimestamp != nil {
			continue
		}
		if podNeedsInjectChange(&pods.Items[i], want) {
			return true, nil
		}
	}
	return false, nil
}

func podNeedsInjectChange(pod *corev1.Pod, want InjectSpec) bool {
	has := InjectSpec{}
	for _, ctr := range pod.Spec.Containers {
		switch ctr.Name {
		case ContainerName:
			has.MeshIn = true
		case linkerdProxyContainerName:
			has.Linkerd = true
		}
	}
	if !want.MeshIn && !want.Linkerd {
		return has.MeshIn || has.Linkerd
	}
	return (want.MeshIn && !has.MeshIn) || (want.Linkerd && !has.Linkerd)
}

func workloadPodSelector(ctx context.Context, c client.Client, ref WorkloadRef) (labels.Selector, error) {
	key := types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
	switch ref.Kind {
	case "Deployment", "":
		var dep appsv1.Deployment
		if err := c.Get(ctx, key, &dep); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(dep.Spec.Selector)
	case "StatefulSet":
		var sts appsv1.StatefulSet
		if err := c.Get(ctx, key, &sts); err != nil {
			if apierrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	default:
		return nil, nil
	}
}

func listAllWorkloadsInNamespace(ctx context.Context, c client.Client, ns string) ([]WorkloadRef, error) {
	var out []WorkloadRef
	var deps appsv1.DeploymentList
	if err := c.List(ctx, &deps, client.InNamespace(ns)); err != nil {
		klog.Errorf("mesh-in-rollout: list deploys ns=%s failed: %v", ns, err)
		return nil, err
	}
	for i := range deps.Items {
		out = append(out, WorkloadRef{Kind: "Deployment", Namespace: ns, Name: deps.Items[i].Name})
	}
	var stss appsv1.StatefulSetList
	if err := c.List(ctx, &stss, client.InNamespace(ns)); err != nil {
		klog.Errorf("mesh-in-rollout: list sts ns=%s failed: %v", ns, err)
		return nil, err
	}
	for i := range stss.Items {
		out = append(out, WorkloadRef{Kind: "StatefulSet", Namespace: ns, Name: stss.Items[i].Name})
	}
	return out, nil
}
