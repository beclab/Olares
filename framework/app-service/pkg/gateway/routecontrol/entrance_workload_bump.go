package routecontrol

import (
	"context"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const annotationD2ReplicaRevision = "app.bytetrade.io/d2-replica-rev"

func bumpSharedEntranceWorkloadForReplica(ctx context.Context, c client.Client, target ReplicaTarget) error {
	if c == nil || !isSharedServerNamespace(target.CallerNamespace) {
		return nil
	}

	sourceRV, err := sourceSecretResourceVersion(ctx, c, target.CertViewer)
	if err != nil {
		return err
	}
	if sourceRV == "" {
		return nil
	}

	var pods corev1.PodList
	if err := c.List(ctx, &pods,
		client.InNamespace(target.CallerNamespace),
		client.MatchingLabels{constants.AppSharedEntrancesLabel: "true"},
	); err != nil {
		return err
	}

	for i := range pods.Items {
		pod := &pods.Items[i]
		if hasD2SidecarContainer(pod) {
			continue
		}
		bumped, err := bumpPodOwnerWorkloadTemplate(ctx, c, pod, sourceRV)
		if err != nil {
			return err
		}
		if bumped {
			return nil
		}
	}
	return nil
}

func sourceSecretResourceVersion(ctx context.Context, c client.Client, viewer string) (string, error) {
	if strings.TrimSpace(viewer) == "" {
		return "", nil
	}
	src := &corev1.Secret{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: defaultGatewayNS,
		Name:      entranceTLSSecretName(viewer),
	}, src)
	if apierrors.IsNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(src.ResourceVersion), nil
}

func bumpPodOwnerWorkloadTemplate(ctx context.Context, c client.Client, pod *corev1.Pod, sourceRV string) (bool, error) {
	if pod == nil || strings.TrimSpace(sourceRV) == "" {
		return false, nil
	}

	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef == nil {
		return false, nil
	}
	switch controllerRef.Kind {
	case "Deployment":
		return bumpDeploymentTemplate(ctx, c, pod.Namespace, controllerRef.Name, sourceRV)
	case "StatefulSet":
		return bumpStatefulSetTemplate(ctx, c, pod.Namespace, controllerRef.Name, sourceRV)
	case "ReplicaSet":
		rs := &appsv1.ReplicaSet{}
		if err := c.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: controllerRef.Name}, rs); err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		rsOwnerRef := metav1.GetControllerOf(rs)
		if rsOwnerRef == nil {
			return false, nil
		}
		switch rsOwnerRef.Kind {
		case "Deployment":
			return bumpDeploymentTemplate(ctx, c, pod.Namespace, rsOwnerRef.Name, sourceRV)
		case "StatefulSet":
			return bumpStatefulSetTemplate(ctx, c, pod.Namespace, rsOwnerRef.Name, sourceRV)
		default:
			return false, nil
		}
	default:
		return false, nil
	}
}

func bumpDeploymentTemplate(ctx context.Context, c client.Client, namespace, name, sourceRV string) (bool, error) {
	deploy := &appsv1.Deployment{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, deploy); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	before := deploy.DeepCopy()
	if deploy.Spec.Template.Annotations[annotationD2ReplicaRevision] == sourceRV {
		return false, nil
	}
	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}
	deploy.Spec.Template.Annotations[annotationD2ReplicaRevision] = sourceRV
	if err := c.Patch(ctx, deploy, client.MergeFrom(before)); err != nil {
		return false, err
	}
	klog.Infof("Bumped shared entrance deployment %s/%s annotation %s=%s for d2 convergence",
		namespace, name, annotationD2ReplicaRevision, sourceRV)
	return true, nil
}

func bumpStatefulSetTemplate(ctx context.Context, c client.Client, namespace, name, sourceRV string) (bool, error) {
	sts := &appsv1.StatefulSet{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, sts); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	before := sts.DeepCopy()
	if sts.Spec.Template.Annotations[annotationD2ReplicaRevision] == sourceRV {
		return false, nil
	}
	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}
	sts.Spec.Template.Annotations[annotationD2ReplicaRevision] = sourceRV
	if err := c.Patch(ctx, sts, client.MergeFrom(before)); err != nil {
		return false, err
	}
	klog.Infof("Bumped shared entrance statefulset %s/%s annotation %s=%s for d2 convergence",
		namespace, name, annotationD2ReplicaRevision, sourceRV)
	return true, nil
}

func hasD2SidecarContainer(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	for i := range pod.Spec.Containers {
		if strings.TrimSpace(pod.Spec.Containers[i].Name) == constants.D2SidecarContainerName {
			return true
		}
	}
	return false
}

func isSharedServerNamespace(namespace string) bool {
	ns := strings.TrimSpace(namespace)
	return ns != "" && strings.HasSuffix(ns, "-shared") && !strings.HasPrefix(ns, userSpaceNamespacePrefix)
}
