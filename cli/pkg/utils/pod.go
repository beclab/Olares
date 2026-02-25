package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func AssertPodReady(pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("pod is nil")
	}

	// simply ignore finished pod
	// it can be seen as just an execution record, not a running pod
	// and the deployment will create a new replica
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return nil
	}

	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	if pod.DeletionTimestamp != nil {
		return fmt.Errorf("pod %s is terminating", podKey)
	}
	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod %s is not running (phase=%s)", podKey, pod.Status.Phase)
	}

	if len(pod.Spec.InitContainers) > 0 {
		initStatusByName := make(map[string]corev1.ContainerStatus, len(pod.Status.InitContainerStatuses))
		for i := range pod.Status.InitContainerStatuses {
			s := pod.Status.InitContainerStatuses[i]
			initStatusByName[s.Name] = s
		}
		for _, ic := range pod.Spec.InitContainers {
			s, ok := initStatusByName[ic.Name]
			if !ok {
				return fmt.Errorf("pod %s has not started init container %s yet", podKey, ic.Name)
			}
			if t := s.State.Terminated; t != nil {
				if t.ExitCode != 0 {
					return fmt.Errorf(
						"init container %s in pod %s terminated (exitCode=%d, reason=%s, message=%s)",
						s.Name, podKey, t.ExitCode, t.Reason, t.Message,
					)
				}
				continue
			}
			if w := s.State.Waiting; w != nil {
				return fmt.Errorf(
					"init container %s in pod %s is waiting (reason=%s, message=%s)",
					s.Name, podKey, w.Reason, w.Message,
				)
			}
			return fmt.Errorf("pod %s init container %s is still running", podKey, s.Name)
		}
	}

	readyCondFound := false
	for i := range pod.Status.Conditions {
		cond := pod.Status.Conditions[i]
		if cond.Type != corev1.PodReady {
			continue
		}
		readyCondFound = true
		if cond.Status != corev1.ConditionTrue {
			if cond.Reason != "" || cond.Message != "" {
				return fmt.Errorf("pod %s is not ready (reason=%s, message=%s)", podKey, cond.Reason, cond.Message)
			}
			return fmt.Errorf("pod %s is not ready", podKey)
		}
		break
	}
	if !readyCondFound {
		return fmt.Errorf("pod %s is not ready (missing Ready condition)", podKey)
	}

	statusByName := make(map[string]corev1.ContainerStatus, len(pod.Status.ContainerStatuses))
	for i := range pod.Status.ContainerStatuses {
		s := pod.Status.ContainerStatuses[i]
		statusByName[s.Name] = s
	}

	for _, c := range pod.Spec.Containers {
		cStatus, ok := statusByName[c.Name]
		if !ok {
			return fmt.Errorf("pod %s has not started container %s yet", podKey, c.Name)
		}

		if t := cStatus.State.Terminated; t != nil {
			return fmt.Errorf(
				"container %s in pod %s terminated (exitCode=%d, reason=%s, message=%s)",
				cStatus.Name,
				podKey,
				t.ExitCode,
				t.Reason,
				t.Message,
			)
		}

		if cStatus.State.Running == nil {
			if w := cStatus.State.Waiting; w != nil {
				return fmt.Errorf(
					"container %s in pod %s is waiting (reason=%s, message=%s)",
					cStatus.Name,
					podKey,
					w.Reason,
					w.Message,
				)
			}
			return fmt.Errorf("container %s in pod %s is not running", cStatus.Name, podKey)
		}
		if !cStatus.Ready {
			return fmt.Errorf("container %s in pod %s is not ready", cStatus.Name, podKey)
		}
	}
	return nil
}
