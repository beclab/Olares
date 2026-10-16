package podhealth

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

// isSidecar reports whether an init container is one Kubernetes runs alongside
// the main containers for the life of the pod, rather than to completion.
func isSidecar(c corev1.Container) bool {
	return c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways
}

// assertSidecarReady reports why a sidecar is not serving, or nil when it is
// running and ready. A restart is not held against it: unlike a plain init
// container it is expected to be restarted, and only its current state decides.
func assertSidecarReady(s corev1.ContainerStatus, podKey string) error {
	if w := s.State.Waiting; w != nil {
		return fmt.Errorf(
			"sidecar %s in pod %s is waiting (reason=%s, message=%s)",
			s.Name, podKey, w.Reason, w.Message,
		)
	}
	if t := s.State.Terminated; t != nil {
		return fmt.Errorf(
			"sidecar %s in pod %s terminated (exitCode=%d, reason=%s, message=%s)",
			s.Name, podKey, t.ExitCode, t.Reason, t.Message,
		)
	}
	if s.Started == nil || !*s.Started {
		return fmt.Errorf("sidecar %s in pod %s has not started", s.Name, podKey)
	}
	if !s.Ready {
		return fmt.Errorf("sidecar %s in pod %s is not ready", s.Name, podKey)
	}
	return nil
}

// AssertPodReady reports why a pod is not serving, or nil when it is running
// with every container ready.
//
// Finished pods are ignored: a Succeeded or Failed pod is an execution record
// rather than a running replica, and its owner creates a replacement.
func AssertPodReady(pod *corev1.Pod) error {
	if pod == nil {
		return fmt.Errorf("pod is nil")
	}

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
			// A sidecar keeps running for as long as the pod does, so waiting
			// for it to terminate would never end. Kubelet does not start the
			// main containers until it is ready, so asserting that much is both
			// all that is available and all that is needed.
			if isSidecar(ic) {
				if err := assertSidecarReady(s, podKey); err != nil {
					return err
				}
				continue
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
				cStatus.Name, podKey, t.ExitCode, t.Reason, t.Message,
			)
		}

		if cStatus.State.Running == nil {
			if w := cStatus.State.Waiting; w != nil {
				return fmt.Errorf(
					"container %s in pod %s is waiting (reason=%s, message=%s)",
					cStatus.Name, podKey, w.Reason, w.Message,
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
