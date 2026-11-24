package meshinagent

import (
	"context"
	"errors"
	"slices"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/mesh"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	nativeSidecarMigrationAnnotation = "gateway.olares.io/native-sidecar-migration"
	nativeSidecarOptOut              = "config.linkerd.io/proxy-enable-native-sidecar"
	nativeSidecarMigrationRevision   = "v2"
)

// NativeInjectorReady waits for the injector deployment to observe its new
// template and serve from the native-sidecar configuration.
func NativeInjectorReady(ctx context.Context, kube kubernetes.Interface, dep *appsv1.Deployment) (bool, error) {
	if dep == nil || dep.Status.ObservedGeneration < dep.Generation ||
		dep.Status.UpdatedReplicas == 0 || dep.Status.AvailableReplicas < dep.Status.UpdatedReplicas ||
		dep.Status.UpdatedReplicas != dep.Status.Replicas {
		return false, nil
	}
	if !HasNativeLinkerd(&corev1.Pod{Spec: dep.Spec.Template.Spec}) {
		return false, nil
	}
	cm, err := kube.CoreV1().ConfigMaps(dep.Namespace).Get(ctx, "linkerd-config", metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	var values struct {
		Proxy struct {
			NativeSidecar bool `json:"nativeSidecar"`
		} `json:"proxy"`
	}
	err = yaml.Unmarshal([]byte(cm.Data["values"]), &values)
	return values.Proxy.NativeSidecar, err
}

func HasNativeLinkerd(pod *corev1.Pod) bool {
	proxy := mesh.LinkerdProxy(pod)
	return proxy != nil && proxy.RestartPolicy != nil && *proxy.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

func needsNativeSidecars(pod *corev1.Pod, installedApp bool) bool {
	if pod.DeletionTimestamp != nil || pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return false
	}
	legacyLinkerd := mesh.LinkerdProxy(pod) != nil && !HasNativeLinkerd(pod) &&
		(installedApp || pod.Annotations[mesh.LinkerdInjectAnnotation] == mesh.LinkerdInjectEnabled ||
			pod.Labels["linkerd.io/control-plane-ns"] == mesh.LinkerdNamespace) &&
		pod.Annotations[nativeSidecarOptOut] != "false"
	legacyMeshIn := false
	for _, container := range pod.Spec.Containers {
		if container.Name == ContainerName {
			legacyMeshIn = true
			break
		}
	}
	nativeMeshIn := HasNativeMeshIn(pod)
	return legacyMeshIn || legacyLinkerd || (installedApp && !nativeMeshIn) ||
		(nativeMeshIn && needsMeshInBootstrapUpdate(pod))
}

func needsMeshInBootstrapUpdate(pod *corev1.Pod) bool {
	var bootstrapReady bool
	for _, container := range pod.Spec.InitContainers {
		switch container.Name {
		case InitContainerName:
			want := []string{"/bin/sh", "-c", IptablesInitScript()}
			bootstrapReady = container.Image == meshInAgentImage() && slices.Equal(container.Command, want)
		case ContainerName:
			if container.Image != meshInAgentImage() {
				return true
			}
		}
	}
	return !bootstrapReady
}

// Queue only controllers that support rolling replacement. Jobs and standalone
// Pods pick up injection on their next creation without being interrupted.
func (w *RolloutWorker) EnqueueNativeSidecarMigration(ctx context.Context) error {
	var namespaces corev1.NamespaceList
	var managers appv1alpha1.ApplicationManagerList
	if err := w.Client.List(ctx, &namespaces); err != nil {
		return err
	}
	if err := w.Client.List(ctx, &managers); err != nil {
		return err
	}
	var pods corev1.PodList
	if err := w.Client.List(ctx, &pods); err != nil {
		return err
	}
	installedNamespaces := make(map[string]bool)
	var scanErr error
	for i := range namespaces.Items {
		ns := &namespaces.Items[i]
		for j := range managers.Items {
			app := &managers.Items[j]
			if !InstalledAppOwnsNamespace(ns, app) {
				continue
			}
			if err := ensureInstalledNamespaceMeshLabels(ctx, w.Client, ns, app); err != nil {
				scanErr = errors.Join(scanErr, err)
				break
			}
			installedNamespaces[ns.Name] = true
			break
		}
	}
	queued := make(map[WorkloadRef]bool)
	for i := range pods.Items {
		pod := &pods.Items[i]
		// Pre-injected Linkerd workloads are replaced by their deployment manifests.
		if pod.Namespace == mesh.LinkerdNamespace || !needsNativeSidecars(pod, installedNamespaces[pod.Namespace]) {
			continue
		}
		ref, err := nativeRolloutOwner(ctx, w.Client, pod)
		if err != nil {
			scanErr = errors.Join(scanErr, err)
			continue
		}
		if ref.Name != "" && !queued[ref] {
			queued[ref] = true
			w.Enqueue(RolloutJob{
				Key: "native-sidecars/" + ref.Key(), Reason: ReasonNativeSidecars,
				Workloads: []WorkloadRef{ref},
			})
		}
	}
	return scanErr
}

// Older primary namespaces can predate the installer ownership labels.
func ensureInstalledNamespaceMeshLabels(ctx context.Context, c client.Client, ns *corev1.Namespace, app *appv1alpha1.ApplicationManager) error {
	if ns.Labels[constants.ApplicationNameLabel] != "" && ns.Labels[constants.ApplicationInstallUserLabel] != "" {
		return nil
	}
	updated := ns.DeepCopy()
	if updated.Labels == nil {
		updated.Labels = make(map[string]string)
	}
	for key, value := range map[string]string{
		constants.ApplicationNameLabel:        app.Spec.AppName,
		constants.ApplicationInstallUserLabel: app.Spec.AppOwner,
	} {
		if updated.Labels[key] == "" {
			updated.Labels[key] = value
		}
	}
	return c.Patch(ctx, updated, client.MergeFromWithOptions(ns, client.MergeFromWithOptimisticLock{}))
}

func nativeRolloutOwner(ctx context.Context, c client.Client, pod *corev1.Pod) (WorkloadRef, error) {
	owner := metav1.GetControllerOf(pod)
	if owner == nil || owner.APIVersion != "apps/v1" {
		return WorkloadRef{}, nil
	}
	if owner.Kind == "ReplicaSet" {
		var rs appsv1.ReplicaSet
		if err := c.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: owner.Name}, &rs); err != nil {
			return WorkloadRef{}, client.IgnoreNotFound(err)
		}
		if rs.UID != owner.UID {
			return WorkloadRef{}, nil
		}
		owner = metav1.GetControllerOf(&rs)
		if owner == nil || owner.Kind != "Deployment" || owner.APIVersion != "apps/v1" {
			return WorkloadRef{}, nil
		}
	}
	ref := WorkloadRef{Namespace: pod.Namespace, Kind: owner.Kind, Name: owner.Name, UID: owner.UID}
	obj, template := nativeRolloutTemplate(ref.Kind)
	if obj == nil {
		return WorkloadRef{}, nil
	}
	if err := c.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, obj); err != nil {
		return WorkloadRef{}, client.IgnoreNotFound(err)
	}
	if obj.GetUID() != owner.UID || !supportsNativeRollingUpdate(obj) ||
		template.Annotations[nativeSidecarMigrationAnnotation] == nativeSidecarMigrationRevision {
		return WorkloadRef{}, nil
	}
	return ref, nil
}

func nativeRolloutTemplate(kind string) (client.Object, *corev1.PodTemplateSpec) {
	switch kind {
	case "Deployment":
		obj := &appsv1.Deployment{}
		return obj, &obj.Spec.Template
	case "StatefulSet":
		obj := &appsv1.StatefulSet{}
		return obj, &obj.Spec.Template
	case "DaemonSet":
		obj := &appsv1.DaemonSet{}
		return obj, &obj.Spec.Template
	default:
		return nil, nil
	}
}

func supportsNativeRollingUpdate(obj client.Object) bool {
	switch obj := obj.(type) {
	case *appsv1.Deployment:
		return !obj.Spec.Paused && obj.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType
	case *appsv1.StatefulSet:
		rolling := obj.Spec.UpdateStrategy.RollingUpdate
		return obj.Spec.UpdateStrategy.Type != appsv1.OnDeleteStatefulSetStrategyType &&
			(rolling == nil || rolling.Partition == nil || *rolling.Partition == 0)
	case *appsv1.DaemonSet:
		return obj.Spec.UpdateStrategy.Type != appsv1.OnDeleteDaemonSetStrategyType
	default:
		return false
	}
}

func bumpNativeSidecars(ctx context.Context, c client.Client, ref WorkloadRef) error {
	obj, template := nativeRolloutTemplate(ref.Kind)
	if err := c.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, obj); err != nil {
		return err
	}
	if obj.GetUID() != ref.UID || !supportsNativeRollingUpdate(obj) ||
		template.Annotations[nativeSidecarMigrationAnnotation] == nativeSidecarMigrationRevision {
		return nil
	}
	base := obj.DeepCopyObject().(client.Object)
	if template.Annotations == nil {
		template.Annotations = map[string]string{}
	}
	template.Annotations[nativeSidecarMigrationAnnotation] = nativeSidecarMigrationRevision
	if dep, ok := obj.(*appsv1.Deployment); ok {
		// Respect workloads explicitly configured to roll without a surge.
		rolling := dep.Spec.Strategy.RollingUpdate
		if rolling == nil || rolling.MaxSurge == nil ||
			(rolling.MaxSurge.String() != "0" && rolling.MaxSurge.String() != "0%") {
			ensureSurgeFirstStrategy(dep)
		}
	}
	return c.Patch(ctx, obj, client.MergeFromWithOptions(base, client.MergeFromWithOptimisticLock{}))
}
