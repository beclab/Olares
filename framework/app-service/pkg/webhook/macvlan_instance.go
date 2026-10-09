package webhook

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// overlayWorkloadIdentity validates the controller chain against live objects.
// The key survives controller recreation within the same Application lifetime.
func (wh *Webhook) overlayWorkloadIdentity(ctx context.Context, pod *corev1.Pod, appName string, checkOverlap bool) (string, error) {
	if wh.kubeClient == nil {
		return "", fmt.Errorf("kubernetes client is required for overlay instance identity")
	}
	owner := metav1.GetControllerOf(pod)
	if owner == nil || owner.UID == "" || owner.APIVersion != "apps/v1" {
		return "", fmt.Errorf("fixed MAC requires a StatefulSet or single-replica Recreate Deployment controller")
	}
	base := pod.Namespace + "/" + appName
	var key string
	var template corev1.PodTemplateSpec
	switch owner.Kind {
	case "StatefulSet":
		sts, err := wh.kubeClient.AppsV1().StatefulSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if sts.UID != owner.UID || !sts.DeletionTimestamp.IsZero() {
			return "", fmt.Errorf("StatefulSet owner UID mismatch or deleting")
		}
		suffix := strings.TrimPrefix(pod.Name, sts.Name+"-")
		ordinal, err := strconv.Atoi(suffix)
		if err != nil || ordinal < 0 || pod.Name != sts.Name+"-"+strconv.Itoa(ordinal) {
			return "", fmt.Errorf("invalid StatefulSet pod name %q", pod.Name)
		}
		if label := pod.Labels["apps.kubernetes.io/pod-index"]; label != "" && label != suffix {
			return "", fmt.Errorf("pod-index does not match StatefulSet pod name")
		}
		key, template = base+"/StatefulSet/"+sts.Name+"/"+suffix, sts.Spec.Template
	case "ReplicaSet":
		rs, err := wh.kubeClient.AppsV1().ReplicaSets(pod.Namespace).Get(ctx, owner.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		deploymentOwner := metav1.GetControllerOf(rs)
		if rs.UID != owner.UID || deploymentOwner == nil || deploymentOwner.Kind != "Deployment" || deploymentOwner.APIVersion != "apps/v1" {
			return "", fmt.Errorf("ReplicaSet has no verified Deployment owner")
		}
		deployment, err := wh.kubeClient.AppsV1().Deployments(pod.Namespace).Get(ctx, deploymentOwner.Name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if deployment.UID != deploymentOwner.UID || !deployment.DeletionTimestamp.IsZero() {
			return "", fmt.Errorf("Deployment owner UID mismatch or deleting")
		}
		if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas > 1 {
			return "", fmt.Errorf("Deployment %s requires one replica for fixed MAC", deployment.Name)
		}
		if deployment.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
			return "", fmt.Errorf("Deployment %s requires Recreate strategy before enabling fixed MAC", deployment.Name)
		}
		key, template = base+"/Deployment/"+deployment.Name+"/singleton", rs.Spec.Template
	default:
		return "", fmt.Errorf("fixed MAC does not support controller %s", owner.Kind)
	}
	if template.Labels[constants.ApplicationNameLabel] != appName || template.Labels[constants.ApplicationOwnerLabel] != pod.Labels[constants.ApplicationOwnerLabel] {
		return "", fmt.Errorf("pod application labels do not match controller template")
	}
	if checkOverlap {
		pods, err := wh.kubeClient.CoreV1().Pods(pod.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		for i := range pods.Items {
			other := &pods.Items[i]
			if pod.UID != "" && pod.UID == other.UID {
				continue
			}
			otherOwner := metav1.GetControllerOf(other)
			if otherOwner == nil {
				continue
			}
			// Do not ignore terminating Pods: their CNI links may still be alive.
			same := otherOwner.UID == owner.UID
			if !same && owner.Kind == "ReplicaSet" && otherOwner.Kind == "ReplicaSet" {
				rs, err := wh.kubeClient.AppsV1().ReplicaSets(pod.Namespace).Get(ctx, otherOwner.Name, metav1.GetOptions{})
				if err != nil {
					return "", fmt.Errorf("cannot verify existing pod owner: %w", err)
				}
				d := metav1.GetControllerOf(rs)
				same = d != nil && strings.HasSuffix(key, "/Deployment/"+d.Name+"/singleton")
			}
			if same && (owner.Kind != "StatefulSet" || other.Name == pod.Name) {
				return "", fmt.Errorf("fixed MAC instance %s still has pod %s (including termination)", key, other.Name)
			}
		}
	}
	return key, nil
}
