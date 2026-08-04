package podhealth

import corev1 "k8s.io/api/core/v1"

// PodContainersStarted reports whether every regular container in the pod has
// started (or has already completed). It mirrors kubelet's per-container
// Started flag and additionally treats a container Terminated with reason
// "Completed" as started, so job-style pods that run to completion count as up.
//
// It is the single source of truth for the "is this pod up?" predicate shared
// by the install/upgrade wait (HelmOps.checkIfStartup, which requires every
// pod to be up) and the resume wait (appstate.isStartUp, which needs any pod
// to be up). The nil-check on Started guards pods whose kubelet has not yet
// populated container statuses.
func PodContainersStarted(pod corev1.Pod) bool {
	total := len(pod.Spec.Containers)
	started := 0
	for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
		cs := pod.Status.ContainerStatuses[i]
		if cs.Started != nil && *cs.Started {
			started++
			continue
		}
		if cs.State.Terminated != nil && cs.State.Terminated.Reason == "Completed" {
			started++
		}
	}
	return started == total
}

// AllPodsStarted reports whether every pod in the slice has all its containers
// started (see PodContainersStarted). It is the shared "is the workload up?"
// tally for both the install/upgrade wait (HelmOps.checkIfStartup) and the
// resume wait (appstate.isStartUp). An empty slice returns true, so callers
// that treat "no pods yet" as not-started must check len(pods) themselves.
func AllPodsStarted(pods []corev1.Pod) bool {
	for _, pod := range pods {
		if !PodContainersStarted(pod) {
			return false
		}
	}
	return true
}
