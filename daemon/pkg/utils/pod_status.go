package utils

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

// GetPodStatus returns a kubectl-like status string for a pod
// because the pod.Status.Phase field is unreliable
func GetPodStatus(pod *corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		if pod.Status.Reason == "NodeLost" {
			return "Unknown"
		}
		return "Terminating"
	}

	for i, container := range pod.Status.InitContainerStatuses {
		if container.State.Terminated != nil && container.State.Terminated.ExitCode == 0 {
			continue
		}

		if container.State.Terminated != nil {
			if container.State.Terminated.Reason != "" {
				return fmt.Sprintf("Init:%s", container.State.Terminated.Reason)
			}
			if container.State.Terminated.Signal != 0 {
				return fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
			}
			return fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
		}

		if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
			return fmt.Sprintf("Init:%s", container.State.Waiting.Reason)
		}

		return fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
	}

	hasRunning := false
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
			return container.State.Waiting.Reason
		}

		if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
			return container.State.Terminated.Reason
		}

		if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
			if container.State.Terminated.Signal != 0 {
				return fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
			}
			return fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
		}

		if container.State.Running != nil && container.Ready {
			hasRunning = true
		}
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse && condition.Reason != "" {
			return condition.Reason
		}
	}

	if pod.Status.Phase == corev1.PodRunning && hasRunning {
		return "Running"
	}

	if pod.Status.Phase != "" {
		return string(pod.Status.Phase)
	}

	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}

	return "Unknown"
}

// IsPodReady checks if a pod is fully ready (all containers ready)
func IsPodReady(pod *corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}

	return false
}
