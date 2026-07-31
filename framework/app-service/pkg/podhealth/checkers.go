package podhealth

import (
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// DefaultCheckers is the canonical detection chain, ordered cheapest/most
// actionable first. To add a new detection, implement Checker and append it
// here (and, if it needs data beyond Pods, extend Input).
func DefaultCheckers() []Checker {
	return []Checker{
		unschedulableChecker{},
		containerWaitingChecker{},
		crashLoopChecker{},
		permanentMountChecker{},
	}
}

// hardWaitingReasons maps a container waiting reason that never self-heals to
// the Signal reason code we surface. These are terminal misconfigurations
// (unpullable/bad image reference, unresolvable container config) that will not
// recover without a spec change.
var hardWaitingReasons = map[string]string{
	"ImagePullBackOff":           ReasonImagePull,
	"ErrImagePull":               ReasonImagePull,
	"InvalidImageName":           ReasonImagePull,
	"CreateContainerConfigError": ReasonContainerConfig,
	"CreateContainerError":       ReasonContainerConfig,
	"RunContainerError":          ReasonContainerConfig,
}

// unschedulableChecker flags pods the scheduler has rejected (PodScheduled
// False / Unschedulable), e.g. no node satisfies the resource request or
// affinity/taint constraints.
type unschedulableChecker struct{}

func (unschedulableChecker) Name() string { return "unschedulable" }

func (unschedulableChecker) Check(in Input) (Signal, bool) {
	for i := range in.Pods {
		pod := &in.Pods[i]
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodScheduled && cond.Status == corev1.ConditionFalse &&
				cond.Reason == corev1.PodReasonUnschedulable {
				return Signal{
					Reason:  ReasonUnschedulable,
					Message: fmt.Sprintf("pod %s unschedulable: %s", pod.Name, cond.Message),
					Grace:   HardErrorGrace,
				}, true
			}
		}
	}
	return Signal{}, false
}

// containerWaitingChecker flags containers stuck in a hard waiting reason (see
// hardWaitingReasons), covering both init and regular containers.
type containerWaitingChecker struct{}

func (containerWaitingChecker) Name() string { return "container-waiting" }

func (containerWaitingChecker) Check(in Input) (Signal, bool) {
	for i := range in.Pods {
		pod := &in.Pods[i]
		for _, cs := range allContainerStatuses(pod) {
			waiting := cs.State.Waiting
			if waiting == nil {
				continue
			}
			if code, ok := hardWaitingReasons[waiting.Reason]; ok {
				return Signal{
					Reason:  code,
					Message: fmt.Sprintf("pod %s container %s: %s (%s)", pod.Name, cs.Name, waiting.Reason, waiting.Message),
					Grace:   HardErrorGrace,
				}, true
			}
		}
	}
	return Signal{}, false
}

// crashLoopChecker flags containers in CrashLoopBackOff that have restarted at
// least CrashLoopRestartThreshold times, which distinguishes a persistently
// failing dependency from a slow first start.
type crashLoopChecker struct{}

func (crashLoopChecker) Name() string { return "crashloop" }

func (crashLoopChecker) Check(in Input) (Signal, bool) {
	for i := range in.Pods {
		pod := &in.Pods[i]
		for _, cs := range allContainerStatuses(pod) {
			waiting := cs.State.Waiting
			if waiting == nil {
				continue
			}
			if waiting.Reason == "CrashLoopBackOff" && cs.RestartCount >= in.CrashLoopRestartThreshold {
				return Signal{
					Reason: ReasonCrashLoop,
					Message: fmt.Sprintf("pod %s container %s: CrashLoopBackOff after %d restarts (%s)",
						pod.Name, cs.Name, cs.RestartCount, waiting.Message),
					Grace: HardErrorGrace,
				}, true
			}
		}
	}
	return Signal{}, false
}

// permanentMountChecker flags a deterministic, non-self-healing FailedMount
// (missing Secret/ConfigMap volume, bad subPath, hostPath type check). These
// never surface as a container waiting reason, so they are invisible to the
// pod-status checkers and must be read from events. It only scans pods that have
// not started any container yet (see podBlockedBeforeStart) and is a no-op when
// Input.FetchEvents is nil.
type permanentMountChecker struct{}

func (permanentMountChecker) Name() string { return "permanent-mount" }

func (permanentMountChecker) Check(in Input) (Signal, bool) {
	if in.FetchEvents == nil {
		return Signal{}, false
	}
	var graced *Signal
	for i := range in.Pods {
		pod := in.Pods[i]
		if !podBlockedBeforeStart(pod) {
			continue
		}
		events, err := in.FetchEvents(pod)
		if err != nil {
			klog.Warningf("podhealth: fetch events for pod %s failed, skip mount-failure check: %v", pod.Name, err)
			continue
		}
		klog.V(4).Infof("podhealth: scanning %d events of pod %s for mount failures", len(events), pod.Name)
		reason, kind, ok := scanPermanentMountFailure(events, in.Now, in.MountEventRecency)
		if !ok {
			continue
		}
		klog.V(4).Infof("podhealth: pod %s permanent mount failure (kind=%d): %s", pod.Name, kind, reason)
		// Immediate failures win outright (no create-race window); keep the
		// first graced hit as a fallback in case no immediate one shows up.
		if kind == mountImmediate {
			return Signal{Reason: ReasonPermanentMount, Message: reason, Grace: 0}, true
		}
		if graced == nil {
			sig := Signal{Reason: ReasonPermanentMount, Message: reason, Grace: MountFailureGrace}
			graced = &sig
		}
	}
	if graced != nil {
		return *graced, true
	}
	return Signal{}, false
}

// allContainerStatuses returns init + regular container statuses so callers
// treat a failing init container the same as a failing app container.
func allContainerStatuses(pod *corev1.Pod) []corev1.ContainerStatus {
	statuses := append([]corev1.ContainerStatus{}, pod.Status.InitContainerStatuses...)
	return append(statuses, pod.Status.ContainerStatuses...)
}

// podBlockedBeforeStart reports whether the pod is still Pending with no
// container having started, which is the only window in which a FailedMount can
// be the reason it is stuck. It gates the (relatively expensive) per-pod event
// scan to pods that could plausibly be hitting a mount failure.
//
// It deliberately does not key off a specific waiting reason: kubelet reports
// "ContainerCreating" for a plain pod but "PodInitializing" for a pod with init
// containers, and a pod whose volumes never mount stays in the latter forever
// (volume manager waits for every pod volume before starting any container,
// init included). Matching only ContainerCreating therefore missed every app
// with an init container.
//
// A running or terminated container proves the volumes did mount, so those pods
// are skipped.
func podBlockedBeforeStart(pod corev1.Pod) bool {
	if pod.Status.Phase != corev1.PodPending {
		return false
	}
	for _, cs := range allContainerStatuses(&pod) {
		if cs.State.Running != nil || cs.State.Terminated != nil {
			return false
		}
	}
	return true
}

// mountKind classifies a permanent FailedMount by whether it can transiently
// occur during normal startup (and thus warrants a grace window) or is
// deterministic with no such race (and can fail immediately).
type mountKind int

const (
	// mountNone means the message is not a recognized permanent failure (or is
	// a transient one such as attach timeout / Multi-Attach).
	mountNone mountKind = iota
	// mountImmediate is a deterministic failure with no create-race window:
	// hostPath type check failure or a missing subPath. These never self-heal,
	// so they can fail without waiting.
	mountImmediate
	// mountGraced can transiently occur while helm is still creating the
	// referenced Secret/ConfigMap concurrently with the workload, so it needs a
	// short grace before being treated as permanent.
	mountGraced
)

// classifyMountFailure classifies a FailedMount message. It intentionally
// returns mountNone for transient failures such as attach timeouts ("timed out
// waiting for the condition") or Multi-Attach errors, which routinely resolve
// once an old pod terminates or the volume detaches.
func classifyMountFailure(msg string) mountKind {
	switch {
	// hostPath type check, e.g. "/dev/ttyUSB0 is not a character device".
	case strings.Contains(msg, "hostPath type check failed"):
		return mountImmediate
	// subPath does not exist on the volume.
	case strings.Contains(msg, "failed to prepare subPath"):
		return mountImmediate
	case strings.Contains(msg, "subPath") && strings.Contains(msg, "no such file or directory"):
		return mountImmediate
	// Referenced Secret/ConfigMap used as a volume does not exist. This may be
	// a create race with helm, so it is graced rather than immediate.
	case (strings.Contains(msg, "secret \"") || strings.Contains(msg, "configmap \"")) &&
		strings.Contains(msg, "not found"):
		return mountGraced
	}
	return mountNone
}

// scanPermanentMountFailure scans a pod's events and reports whether it has a
// recent, permanent FailedMount, returning a human-readable reason and its kind
// (mountImmediate vs mountGraced). Events older than recency are ignored so an
// already-recovered mount does not cause a false positive. When both kinds are
// present the immediate one wins. Pure function over the event slice.
func scanPermanentMountFailure(events []corev1.Event, now time.Time, recency time.Duration) (string, mountKind, bool) {
	var gracedReason string
	var gracedFound bool
	for i := range events {
		e := &events[i]
		if e.Reason != "FailedMount" {
			continue
		}
		ts := e.LastTimestamp.Time
		if ts.IsZero() {
			// series / newer events populate EventTime instead of LastTimestamp.
			ts = e.EventTime.Time
		}
		if !ts.IsZero() && now.Sub(ts) > recency {
			continue
		}
		switch classifyMountFailure(e.Message) {
		case mountImmediate:
			return fmt.Sprintf("pod %s FailedMount: %s", e.InvolvedObject.Name, e.Message), mountImmediate, true
		case mountGraced:
			if !gracedFound {
				gracedReason = fmt.Sprintf("pod %s FailedMount: %s", e.InvolvedObject.Name, e.Message)
				gracedFound = true
			}
		}
	}
	if gracedFound {
		return gracedReason, mountGraced, true
	}
	return "", mountNone, false
}
